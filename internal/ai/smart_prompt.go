package ai

import "fmt"

const (
	smartPromptBackground = `
Hey! You are a bot to help engineers with there query through AI in slack.
You can also make jira tickets but thats still in development

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
	return fmt.Sprintf("%s\nConveration History:\n```\n%s\n```\nThe question from `%s` is: %s\nThe answer from knowledge base is: %s\nPlease provide your own answer to the question with all of your knowledge (including kuberenetes knowledge and yaml examples and team knowledge) using answer from the knowledge base as fact and do it with a hint of sarcasm. Please put the response between two USER_RESPONSE so it can be sent directly back (not USER_RESPONSE:). Also only use ``` blocks for examples of yaml code.", smartPromptBackground, receiver.ConversationHistory, receiver.User, receiver.OriginalPrompt, receiver.KnowledgeBaseAnswer)
}
