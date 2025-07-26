package ai

import (
	"os"
	"github.com/curator4/io-tui/api"
)

type Core struct {
	api api.AIAPI
}

func NewCore() *Core {
	gemini, err := api.NewGeminiAPI()
	if err != nil {
		fmt(Printf("failed to start gemini"))
		os.Exit(1)
	}
	return core{
		api: gemini,
	}
}
