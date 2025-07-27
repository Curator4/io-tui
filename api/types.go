package api

import "github.com/curator4/io-tui/types"

type AIAPI interface {
	GetResponse(messages []types.Message, systemPrompt string) (string, error)
}

type StreamingAPI interface {
	AIAPI
	GetStreamingResponse(messages []types.Message, systemPrompt string) (<-chan string, <-chan error)
}
