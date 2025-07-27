package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/curator4/io-tui/chat"
	"github.com/curator4/io-tui/db"
)

func main() {
	database, err := db.Init()
	if err != nil {
		fmt.Printf("could not init database: %w", err)
		os.Exit(1)
	}

	p := tea.NewProgram(chat.InitialModel(database), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if err := p.Start(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
