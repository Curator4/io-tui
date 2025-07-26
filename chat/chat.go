package chat

// A simple program demonstrating the text area component from the Bubbles
// component library.

import (
	"fmt"
	"strings"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const gap = "\n\n\n"

// Color theme from ASCII art
const (

	// ui color
	borderColor    = "#1e40af"  // Darker blue for border
	separatorColor = "#60a5fa"  // Lighter blue for separators
	
	// Info panel colors
	timeColor      = "#e5e7eb"  // Gray/white for time
	labelColor     = "#22d3ee"  // Cyan for labels
	valueColor     = "#950056"  // Deep burgundy for values

	// chat colors
	userColor = "#0061cd"
	botColor = "#ce75b7"
)

type (
	errMsg error
)


type message struct {
	message string
	role string
}

// api state
type apiState int
const (
	online apiState = iota
	offline
)

func (a apiState) String() string {
	switch a {
	case online:
		return "online"
	case offline:
		return "offline"
	}
	return ""
}


type infoPanel struct {
	ai string
	api string
	model string
	conversation string
	apiStatus apiState 
}

// ai state
type statusState int
const (
	AtEase statusState = iota
	Processing
	Typing
	Error
)

type statusPanel struct {
	spinner spinner.Model
	status statusState
	
}

type Model struct {
	viewport    viewport.Model
	textarea    textarea.Model
	messages    []message
	ascii		string
	width       int
	height      int
	infoPanel	infoPanel
	statusPanel	statusPanel
	err         error
}

func InitialModel() Model {



	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = ""
	ta.CharLimit = 2000

	ta.SetWidth(30)
	ta.SetHeight(2)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)

	ta.KeyMap.InsertNewline.SetEnabled(true)

	infoPanel := infoPanel{
		model: "gpt-4",
		api: "openai",
		ai: "Io",
		conversation: "conversation1",
		apiStatus: online,
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	statusPanel := statusPanel{
		spinner: s,
		status: AtEase,
	}

	return Model{
		viewport:    vp,
		textarea:    ta,
		messages:    []message{},
		ascii:		 loadAscii(),
		width:       80,
		height:      24,
		infoPanel: 	 infoPanel,
		statusPanel: statusPanel,
		err:         nil,
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		asciiHeight := lipgloss.Height(m.ascii)
		m.viewport.Width = msg.Width - 2
		m.textarea.SetWidth(msg.Width - 2)
		
		// Calculate height with minimum safety check
		newHeight := msg.Height - m.textarea.Height() - lipgloss.Height(gap) - asciiHeight
		if newHeight < 1 {
			newHeight = 1  // Minimum height of 1
		}
		m.viewport.Height = newHeight
		
		// Only update content and scroll if we have valid dimensions
		if len(m.messages) > 0 && m.viewport.Width > 0 && m.viewport.Height > 0 {
			m.viewport.SetContent(m.formatMessages())
			m.viewport.GotoBottom()
		}
		
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp, tea.KeyDown:
			// Arrow keys only go to textarea for navigation
			m.textarea, tiCmd = m.textarea.Update(msg)
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			userMessage := message{
				message: m.textarea.Value(),
				role: "user",
			}

			// TODO: bot message (TEMP)
			botMessage := message{
				message: m.textarea.Value(),
				role: "bot",
			}

			m.messages = append(m.messages, userMessage)
			m.messages = append(m.messages, botMessage)
			
			// Update viewport content safely
			if m.viewport.Height > 0 {
				m.viewport.SetContent(m.formatMessages())
				m.viewport.GotoBottom()
			}
			m.textarea.Reset()
		default:
			// All other keys go to textarea
			m.textarea, tiCmd = m.textarea.Update(msg)
		}
		
	case tea.MouseMsg:
		if msg.Type == tea.MouseWheelUp || msg.Type == tea.MouseWheelDown {
			// Mouse wheel only goes to viewport
			m.viewport, vpCmd = m.viewport.Update(msg)
		}
		
	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
		
	default:
		// Other messages can go to both
		m.textarea, tiCmd = m.textarea.Update(msg)
		m.viewport, vpCmd = m.viewport.Update(msg)
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m Model) View() string {

	// custom border style for content (needs model)
	contentBorder := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor))

	// i am using this width becuz of issue i had personally
	// (i think cuz of hyprland padding, but speculative)
	// Point being, it might be messed up for you. works on my machine 🤷
	// maybe remove the -2 in normal setup
	contentWidth := m.width -2
	
	// Calculate heights for right panel components
	statusPanelHeight := 3
	separatorHeight := 1
	asciiHeight := lipgloss.Height(m.ascii)
	infoPanelHeight := asciiHeight - statusPanelHeight - separatorHeight
	
	// Create info panel with calculated height
	infoPanelStyle := lipgloss.NewStyle().Height(infoPanelHeight)
	infoContent := m.makeInfoPanel()
	styledInfoPanel := infoPanelStyle.Render(infoContent)
	
	// right panel
	rightPanel := lipgloss.JoinVertical(
		lipgloss.Center,
		styledInfoPanel,
		horizontalSeparator(contentWidth - lipgloss.Width(m.ascii) - 2),
		m.makeStatusPanel(),
	)
	
	// top panel
	topPanel := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.ascii,
		verticalSeparator(lipgloss.Height(m.ascii)),
		rightPanel,
	)
	
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		topPanel,
		horizontalSeparator(contentWidth),
		m.viewport.View(),
		horizontalSeparator(contentWidth),
		m.textarea.View(),
	)
	return contentBorder.Render(content)
}

func (m Model) formatMessages() string {
	userStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(userColor)).
		Align(lipgloss.Right).
		Width(m.viewport.Width)
	botStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(botColor)).
		Align(lipgloss.Left).
		Width(m.viewport.Width)

	var content strings.Builder
	var lastRole string

	for i, msg := range m.messages {
		// Add separator when speaker changes (but not for first message)
		if i > 0 && msg.role != lastRole {
			// Use the color and alignment of whoever just finished speaking
			var separatorColor string
			var separatorStyle lipgloss.Style
			if lastRole == "user" {
				separatorColor = userColor
				separatorStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color(separatorColor)).
					Align(lipgloss.Right).
					Width(m.viewport.Width)
			} else {
				separatorColor = botColor
				separatorStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color(separatorColor)).
					Align(lipgloss.Left).
					Width(m.viewport.Width)
			}
			
			separator := separatorStyle.Render("───")
			content.WriteString(separator + "\n")
		}

		var styledMessage string
		switch msg.role {
		case "user":
			styledMessage = userStyle.Render(msg.message)
		case "bot":
			styledMessage = botStyle.Render(msg.message)
		}
		content.WriteString(styledMessage + "\n")
		lastRole = msg.role
	}
	return content.String()
}


func (m Model) makeInfoPanel() string {
	currentTime := time.Now().Format("15:04:05")

	// apistatus color
	var statusColor string
	switch m.infoPanel.apiStatus {
	case online:
		statusColor = "10"
	case offline:
		statusColor = "9"
	}
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(statusColor)).
		Bold(true)

	// Combine time styling and centering
	centerTimeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(timeColor)).
		Bold(true).
		Align(lipgloss.Center).
		Width(25)
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		centerTimeStyle.Render(currentTime),
		"",
		"",
		fmt.Sprintf("%s %s", labelStyle.Render("ai:"), valueStyle.Render(m.infoPanel.ai)),
		"",
		fmt.Sprintf("%s %s", labelStyle.Render("api:"), valueStyle.Render(m.infoPanel.api)),
		"",
		fmt.Sprintf("%s %s", labelStyle.Render("model:"), valueStyle.Render(m.infoPanel.model)),
		"",
		fmt.Sprintf("%s %s", labelStyle.Render("conversation:"), valueStyle.Render(m.infoPanel.conversation)),
		"",
		fmt.Sprintf("%s %s", labelStyle.Render("status:"), statusStyle.Render(m.infoPanel.apiStatus.String())),
		"",
	)
}

func (m Model) makeStatusPanel() string {
	var icon, text, color string

	switch m.statusPanel.status {
	case AtEase:
		icon, text, color = "●", "ready", "10"
	case Processing:
		icon, text, color = m.statusPanel.spinner.View(), "processing...", "11"
	case Typing:
		icon, text, color = "✎", "typing..", "12"
	case Error:
		icon, text, color = "✗", "error", "9"
	}

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Bold(true)

	content := statusStyle.Render(fmt.Sprintf("%s %s", icon, text))

	// Fixed height for status panel (like textarea height)
	statusPanelHeight := 3
	panelStyle := lipgloss.NewStyle().
		Height(statusPanelHeight).
		AlignVertical(lipgloss.Center)

	return panelStyle.Render(content)
}


func horizontalSeparator(width int) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(separatorColor)).
		Render(strings.Repeat("─", width))
}

func verticalSeparator(height int) string {
	var lines []string
	for i := 0; i < height; i++ {
		lines = append(lines, "│")
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(separatorColor)).
		Render(strings.Join(lines, "\n"))
}

func loadAscii() string {
	artBytes, err := os.ReadFile("/home/curator/workspace/projects/io-tui/avatar_art.txt")
	if err != nil {
		return "🤖"
	}
	
	// Split into lines and remove the last line
	lines := strings.Split(string(artBytes), "\n")
	if len(lines) > 0 {
		lines = lines[:len(lines)-1]  // Remove last line
	}
	
	return strings.Join(lines, "\n")
}

var labelStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color(labelColor))

var valueStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color(valueColor))
