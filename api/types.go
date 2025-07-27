package api

import "github.com/curator4/io-tui/types"

type AIAPI interface {
	GetResponse(messages []types.Message) (string, error)
}

type StreamingAPI interface {
	AIAPI
	GetStreamingResponse(messages []types.Message) (<-chan string, <-chan error)
}
