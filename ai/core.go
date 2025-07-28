package ai

import (
	"os"
	"fmt"
	"github.com/curator4/io-tui/api"
)

type Core struct {
	API api.AIAPI
}

func NewCore() Core {
	gemini, err := api.NewGeminiAPI()
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		fmt.Println("\nTo fix this:")
		fmt.Println("1. Set environment variable: export GEMINI_API_KEY=your_key")
		fmt.Println("2. Or add your key to demo_api_key.txt (uncomment the line)")
		os.Exit(1)
	}
	return Core{
		API: gemini,
	}
}
