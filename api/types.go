package api

type Message struct {
	Role    string
	Content string
}

type AIAPI interface {
	GetResponse(messages []Message) (string, error)
}

type StreamingAPI interface {
	AIAPI
	GetStreamingResponse(messages []Message) (<-chan string, <-chan error)
}
