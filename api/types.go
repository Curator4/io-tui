package api

import "github.com/curator4/io-tui/types"

// FunctionCall represents a function call from the AI
type FunctionCall struct {
    Name string
    Args map[string]interface{}
}

// ResponseWithFunctions represents a response that may contain both text and function calls
type ResponseWithFunctions struct {
    Text          string
    FunctionCalls []FunctionCall
}

type AIAPI interface {
	GetResponse(messages []types.Message, systemPrompt string) (string, error)
}

type StreamingAPI interface {
	AIAPI
	GetStreamingResponse(messages []types.Message, systemPrompt string) (<-chan string, <-chan error)
}

type EnhancedStreamingAPI interface {
	StreamingAPI
	GetEnhancedStreamingResponse(messages []types.Message, systemPrompt string) (<-chan string, <-chan []FunctionCall, <-chan error)
}

type FunctionAPI interface {
	AIAPI
	GetResponseWithFunctions(messages []types.Message, systemPrompt string) (*ResponseWithFunctions, error)
}
