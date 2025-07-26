package ai

import (
	"os"
	"fmt"
	"github.com/curator4/io-tui/api"

)

type APIMessage = api.Message

type Core struct {
	API api.AIAPI
}

func NewCore() Core {
	gemini, err := api.NewGeminiAPI()
	if err != nil {
		fmt.Printf("failed to start gemini")
		os.Exit(1)
	}
	return Core{
		API: gemini,
	}
}
