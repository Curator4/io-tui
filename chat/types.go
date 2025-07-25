package chat

import (
	"time"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Message struct {
	Role string
	Content string
	Time time.Time
}

type Model struct {
	Chatlog   []Message
	Image     string
	InputText string
	IsLoading bool
	Width, Height int
}

func (m Model) Init() tea.Cmd {
	return tea.SetWindowTitle("Io")
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			if m.InputText != "" && !m.IsLoading {
				userMsg := Message{
					Role: "user",
					Content: m.InputText,
					Time: time.Now(),
				}
				m.Chatlog = append(m.Chatlog, userMsg)

				m.InputText = ""
				m.IsLoading = true

				// TODO: return m, callAPI()
			}
		case "backspace":
			if len(m.InputText) > 0 {
				m.InputText = m.InputText[:len(m.InputText)-1]
			}

		default:
			if len(msg.String()) == 1 {
				m.InputText += msg.String()
			}
		}
	
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}

	return m, nil
}

func (m Model) View() string {
	var output strings.Builder

	// chat history
	for _, msg := range m.Chatlog {
		if msg.Role == "user" {
			output.WriteString("You: " + msg.Content + "\n")
		} else {
			output.WriteString("Io: " + msg.Content + "\n")
		}
	}

	// loading indicator
	if m.IsLoading {
		output.WriteString("Bot is thinking ...\n")
	}

	// input
	output.WriteString("\n> " + m.InputText)

	return output.String()
}
