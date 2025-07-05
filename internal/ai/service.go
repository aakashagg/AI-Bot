package ai

import (
	mytypes "ai-bot/internal/types"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime/types"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"log"
	"os"
	"strings"
)

const (
	defaultRegion     = "us-east-1"
	knowledgeModelArn = "arn:aws:bedrock:us-east-1::foundation-model/anthropic.claude-v2"
	smartModelID      = "meta.llama2-70b-chat-v1" //https://docs.aws.amazon.com/bedrock/latest/userguide/model-ids-arns.html//https://docs.aws.amazon.com/bedrock/latest/userguide/model-ids-arns.html
)

var knowledgeBaseID = os.Getenv("KNOWLEDGE_BASE_ID")

type Service struct {
	BedrockAgentRuntime *bedrockagentruntime.Client
	BedrockRuntime      *bedrockruntime.Client
}

func NewService() (*Service, error) {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = defaultRegion
	}

	awsProfile := os.Getenv("AWS_PROFILE")

	var err error
	var cfg aws.Config
	if awsProfile == "" {
		cfg, err = config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	} else {
		cfg, err = config.LoadDefaultConfig(context.Background(), config.WithRegion(region), config.WithAssumeRoleCredentialOptions(func(options *stscreds.AssumeRoleOptions) {
			options.TokenProvider = stscreds.StdinTokenProvider
		}))
	}

	if err != nil {
		return nil, err
	}

	bar := bedrockagentruntime.NewFromConfig(cfg)
	br := bedrockruntime.NewFromConfig(cfg)
	return &Service{
		BedrockAgentRuntime: bar,
		BedrockRuntime:      br,
	}, nil
}

type Generation struct {
	SessionId string
	Text      string
}

func (s *Service) GenerateFromKnowledge(thread mytypes.Thread, history []string, user, prompt string) Generation {

	input := &bedrockagentruntime.RetrieveAndGenerateInput{
		Input: &types.RetrieveAndGenerateInput{
			Text: aws.String(prompt),
		},
		RetrieveAndGenerateConfiguration: &types.RetrieveAndGenerateConfiguration{
			Type: types.RetrieveAndGenerateTypeKnowledgeBase,
			KnowledgeBaseConfiguration: &types.KnowledgeBaseRetrieveAndGenerateConfiguration{
				KnowledgeBaseId:         aws.String(knowledgeBaseID),
				ModelArn:                aws.String(knowledgeModelArn),
				GenerationConfiguration: &types.GenerationConfiguration{},
				RetrievalConfiguration: &types.KnowledgeBaseRetrievalConfiguration{
					VectorSearchConfiguration: &types.KnowledgeBaseVectorSearchConfiguration{
						NumberOfResults:    aws.Int32(10),
						OverrideSearchType: types.SearchTypeHybrid,
					},
				},
			},
		},
	}

	if thread.BedrockAgentSessionId != "" {
		input.SessionId = aws.String(thread.BedrockAgentSessionId)
	}

	output, err := s.BedrockAgentRuntime.RetrieveAndGenerate(context.Background(), input)

	if err != nil {
		log.Println("failed to invoke model: ", err)
	}

	//sessionID := output.SessionId For tracking

	smartPrompt := &SmartPrompt{
		OriginalPrompt:      prompt,
		KnowledgeBaseAnswer: *output.Output.Text,
		ConversationHistory: strings.Join(history, "\n"),
		User:                user,
	}

	payload := Request{
		Prompt:      smartPrompt.GenerateStringPrompt(),
		Temperature: 0.25,
		TopP:        1,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("raw request", string(payloadBytes))

	out, err := s.BedrockRuntime.InvokeModel(context.Background(), &bedrockruntime.InvokeModelInput{
		Body:        payloadBytes,
		ModelId:     aws.String(smartModelID),
		ContentType: aws.String("application/json"),
	})

	if err != nil {
		log.Fatal("failed to invoke model: ", err)
	}

	log.Println("raw response ", string(out.Body))

	var resp Response

	err = json.Unmarshal(out.Body, &resp)

	if err != nil {
		log.Fatal("failed to unmarshal", err)
	}

	fmt.Println("response from LLM\n", resp.Generation)

	pieces := strings.Split(resp.Generation, "USER_RESPONSE")

	text := resp.Generation
	if len(pieces) >= 2 {
		text = pieces[1]
	}

	return Generation{
		SessionId: *output.SessionId,
		Text:      text,
	}
}

type Request struct {
	Prompt string `json:"prompt"`
	//MaxTokensToSample int      `json:"max_tokens_to_sample"`
	Temperature   float64  `json:"temperature,omitempty"`
	TopP          float64  `json:"top_p,omitempty"`
	TopK          int      `json:"top_k,omitempty"`
	StopSequences []string `json:"stop_sequences,omitempty"`
}

type Response struct {
	Generation string `json:"generation"`
}
