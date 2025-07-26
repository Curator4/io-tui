package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Curator4/io-tui/chat"
)

func main() {
	p := tea.NewProgram(chat.InitialModel(), tea.WithAltScreen())

	if err := p.Start(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
