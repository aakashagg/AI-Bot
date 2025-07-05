package ai

import "fmt"

const (
	smartPromptBackground = `
Hey! You are a bot to help engineers with their query through AI in Slack.
You can also make Jira tickets but that's still in development

`
)

type SmartPrompt struct {
	OriginalPrompt      string
	KnowledgeBaseAnswer string
	ConversationHistory string
	User                string
}

func NewSmartPrompt(originalPrompt, knowledgeBaseAnswer, history string) *SmartPrompt {
	return &SmartPrompt{
		OriginalPrompt:      originalPrompt,
		KnowledgeBaseAnswer: knowledgeBaseAnswer,
		ConversationHistory: history,
	}
}

func (receiver SmartPrompt) GenerateStringPrompt() string {
	return fmt.Sprintf("%s\nConversation History:\n```\n%s\n```\nThe question from `%s` is: %s\nThe answer from knowledge base is: %s\nPlease provide your own answer to the question with all of your knowledge (including Kubernetes knowledge and YAML examples and team knowledge) using answer from the knowledge base as fact and do it with a hint of sarcasm. Please put the response between two USER_RESPONSE so it can be sent directly back (not USER_RESPONSE:). Also only use ``` blocks for examples of YAML code.", smartPromptBackground, receiver.ConversationHistory, receiver.User, receiver.OriginalPrompt, receiver.KnowledgeBaseAnswer)
}
