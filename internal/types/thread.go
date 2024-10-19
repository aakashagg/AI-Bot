package types

type Thread struct {
	Timestamp             string `dynamo:"timestamp,hash"`
	BedrockAgentSessionId string `dynamo:"bedrockAgentSessionId"`
}
