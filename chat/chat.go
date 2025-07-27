package chat

// A simple program demonstrating the text area component from the Bubbles
// component library.

import (
	"fmt"
	"strings"
	"os"
	"time"
	"database/sql"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/curator4/io-tui/ai"
	"github.com/curator4/io-tui/api"
	"github.com/curator4/io-tui/db"
	"github.com/curator4/io-tui/types"
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
	viewMode int
)

const (
	chatMode viewMode = iota
	listMode
)


type AIResponseMsg struct {
	message types.Message
}

type AIIntroductionMsg struct {
	message types.Message
}

type AIStreamStartMsg struct {
	textChan <-chan string
	errChan  <-chan error
}

type AIStreamChunkMsg struct {
	chunk    string
	textChan <-chan string
	errChan  <-chan error
}

type AIStreamCompleteMsg struct {
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
	// database reference
	database *sql.DB

	// a struct that has an API interface (handles requests).
	// convoluted way to set it up, i know...
	// but i was considering further stuff like api indepedent tools
	// actually, thinking again, i dont think this was necisarry
	// i used "core" cuz i dont like the term "manager"
	aicore ai.Core

	// to keep track of active session (ai config & conversation)
	ai db.AI
	conversation db.Conversation

	// cache of displayMessages.
	// To prevent having to query the database for chatlog on every update
	// defined in types becuz i use it in api package too
	messages	[]types.Message

	viewport    viewport.Model
	textarea    textarea.Model
	list        list.Model
	ascii		string
	width       int
	height      int
	statusPanel	statusPanel
	viewMode    viewMode
	err         error
}

func InitialModel(database *sql.DB) Model {



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

	s := spinner.New()
	s.Spinner = spinner.Dot
	statusPanel := statusPanel{
		spinner: s,
		status: AtEase,
	}
	
	activeAI, err := db.GetActiveAI(database)
	if err != nil {
		fmt.Printf("no initial ai %w", err)
		os.Exit(1)
	}
	asciiPath, _:= db.GetAIAsciiPath(database, activeAI.ID)
	asciiContent := loadAscii(asciiPath)


	return Model{
		database:	 database,
		ai:			 activeAI,
		conversation: db.Conversation{}, // Empty struct instead of nil
		aicore:		 ai.NewCore(),
		viewport:    vp,
		textarea:    ta,
		list:        list.New([]list.Item{}, createListDelegate(), 0, 0),
		ascii:		 asciiContent,
		width:       80,
		height:      24,
		statusPanel: statusPanel,
		viewMode:    chatMode,
		err:         nil,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.statusPanel.spinner.Tick)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case AIResponseMsg:
		// Save to database and add to display cache
		if err := db.SaveMessage(m.database, m.conversation.ID, "assistant", msg.message.Content); err != nil {
			// Add error message to chat if save fails
			errorMsg := types.Message{
				Role:    "system",
				Content: fmt.Sprintf("⚠️ Failed to save assistant message: %v", err),
			}
			m.messages = append(m.messages, errorMsg)
		}
		m.messages = append(m.messages, msg.message)
		m.statusPanel.status = AtEase
		if m.viewport.Height > 0 {
			m.viewport.SetContent(m.formatMessages())
			m.viewport.GotoBottom()
		}
		return m, nil

	case AIIntroductionMsg:
		// Add to display without saving to database
		m.messages = append(m.messages, msg.message)
		m.statusPanel.status = AtEase
		if m.viewport.Height > 0 {
			m.viewport.SetContent(m.formatMessages())
			m.viewport.GotoBottom()
		}
		return m, nil

	case AIStreamStartMsg:
		// Add empty bot message immediately
		m.messages = append(m.messages, types.Message{Role: "assistant", Content: ""})
		m.statusPanel.status = Processing
		if m.viewport.Height > 0 {
			m.viewport.SetContent(m.formatMessages())
			m.viewport.GotoBottom()
		}
		// Start reading first chunk
		return m, m.readNextChunk(msg.textChan, msg.errChan)

	case AIStreamChunkMsg:
		// Append chunk to last bot message
		if len(m.messages) > 0 && m.messages[len(m.messages)-1].Role == "assistant" {
			m.messages[len(m.messages)-1].Content += msg.chunk
		}
		// Update viewport and continue reading next chunk
		if m.viewport.Height > 0 {
			m.viewport.SetContent(m.formatMessages())
			m.viewport.GotoBottom()
		}
		return m, m.readNextChunk(msg.textChan, msg.errChan)

	case AIStreamCompleteMsg:
		// Save the complete streamed message to database
		if len(m.messages) > 0 && m.messages[len(m.messages)-1].Role == "assistant" {
			lastMessage := m.messages[len(m.messages)-1]
			if err := db.SaveMessage(m.database, m.conversation.ID, "assistant", lastMessage.Content); err != nil {
				// Add error message to chat if save fails
				errorMsg := types.Message{
					Role:    "system",
					Content: fmt.Sprintf("⚠️ Failed to save streamed message: %v", err),
				}
				m.messages = append(m.messages, errorMsg)
				if m.viewport.Height > 0 {
					m.viewport.SetContent(m.formatMessages())
					m.viewport.GotoBottom()
				}
			}
		}
		m.statusPanel.status = AtEase
		return m, nil
		
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
		// Handle list mode separately
		if m.viewMode == listMode {
			switch msg.Type {
			case tea.KeyEsc:
				// If filtering, let list handle Esc to exit filter first
				if m.list.FilterState() == list.Filtering {
					var listCmd tea.Cmd
					m.list, listCmd = m.list.Update(msg)
					return m, listCmd
				}
				// Otherwise exit list mode back to chat
				m.viewMode = chatMode
				return m, nil
			case tea.KeyRunes:
				// Handle specific key presses
				if len(msg.Runes) > 0 && msg.Runes[0] == 'q' {
					// q should close list, not quit program
					m.viewMode = chatMode
					return m, nil
				}
				// Let other runes (j, k, etc.) pass through to list
				var listCmd tea.Cmd
				m.list, listCmd = m.list.Update(msg)
				return m, listCmd
			case tea.KeyEnter:
				// If filtering, let list handle Enter to select filtered item
				if m.list.FilterState() == list.Filtering {
					var listCmd tea.Cmd
					m.list, listCmd = m.list.Update(msg)
					return m, listCmd
				}
				// Handle selection based on list title
				if selectedItem := m.list.SelectedItem(); selectedItem != nil {
					if aiItem, ok := selectedItem.(aiItem); ok {
						if strings.Contains(m.list.Title, "Select AI") {
							// This is a selector - switch AI
							return m.setAI(aiItem.ai.Name)
						}
						// This is just a view list - do nothing on Enter
						return m, nil
					}
					if apiItem, ok := selectedItem.(apiItem); ok {
						if strings.Contains(m.list.Title, "Select API") {
							// This is a selector - switch API
							return m.setAPI(apiItem.name)
						}
						// This is a view list - navigate to models for this API
						return m.listModels(apiItem.name)
					}
					if modelItem, ok := selectedItem.(modelItem); ok {
						if strings.Contains(m.list.Title, "Select") && strings.Contains(m.list.Title, "Model") {
							// This is a selector - switch model
							return m.setModel(modelItem.name)
						}
						// This is just a view list - do nothing on Enter
						return m, nil
					}
					if conversationItem, ok := selectedItem.(conversationItem); ok {
						// Resume this conversation
						return m.resumeConversation(conversationItem.conversation.ID)
					}
				}
			default:
				// Let list handle navigation
				var listCmd tea.Cmd
				m.list, listCmd = m.list.Update(msg)
				return m, listCmd
			}
			return m, nil
		}
		
		// Chat mode key handling
		switch msg.Type {
		case tea.KeyUp, tea.KeyDown:
			// Arrow keys only go to textarea for navigation
			m.textarea, tiCmd = m.textarea.Update(msg)

		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit

		case tea.KeyEnter:
			userInput := m.textarea.Value()
			
			// Don't send empty messages
			if strings.TrimSpace(userInput) == "" {
				return m, nil
			}

			// Handle slash commands
			if strings.HasPrefix(userInput, "/") {
				return m.handleSlashCommand(userInput)
			}

			// create conversation if none is active
			if m.conversation.ID == 0 {
				conv, _ := db.CreateConversation(m.database, userInput, m.ai.ID)
				m.conversation = conv
			}

			userMessage := types.Message{
				Content: userInput,
				Role: "user",
			}
			// Save to database
			if err := db.SaveMessage(m.database, m.conversation.ID, "user", userInput); err != nil {
				// Add error message to chat if save fails
				errorMsg := types.Message{
					Role:    "system",
					Content: fmt.Sprintf("⚠️ Failed to save user message: %v", err),
				}
				m.messages = append(m.messages, errorMsg)
			}
			m.messages = append(m.messages, userMessage)
			m.statusPanel.status = Processing

			// Update viewport content safely
			if m.viewport.Height > 0 {
				m.viewport.SetContent(m.formatMessages())
				m.viewport.GotoBottom()
			}
			m.textarea.Reset()

			return m, m.callAI(userInput)

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
		
		// Handle spinner updates
		var spinnerCmd tea.Cmd
		m.statusPanel.spinner, spinnerCmd = m.statusPanel.spinner.Update(msg)
		
		return m, tea.Batch(tiCmd, vpCmd, spinnerCmd)
	}

	// Handle spinner updates for all cases
	var spinnerCmd tea.Cmd  
	m.statusPanel.spinner, spinnerCmd = m.statusPanel.spinner.Update(msg)
	
	return m, tea.Batch(tiCmd, vpCmd, spinnerCmd)
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
	
	// Conditional main content based on view mode
	var mainContent string
	if m.viewMode == listMode {
		// Set list size to viewport dimensions  
		m.list.SetSize(m.viewport.Width, m.viewport.Height)
		// Force left alignment for list content
		listStyle := lipgloss.NewStyle().
			Align(lipgloss.Left).
			Width(m.viewport.Width)
		mainContent = listStyle.Render(m.list.View())
	} else {
		// Normal chat viewport
		mainContent = m.viewport.View()
	}
	
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		topPanel,
		horizontalSeparator(contentWidth),
		mainContent,
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
		// Add separator when speaker changes (but not for system messages)
		shouldAddSeparator := i > 0 && msg.Role != lastRole && 
			(lastRole == "user" || lastRole == "assistant") && 
			(msg.Role == "user" || msg.Role == "assistant")
		
		if shouldAddSeparator {
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
		switch msg.Role {
		case "user":
			styledMessage = userStyle.Render(msg.Content)
		case "assistant":
			styledMessage = botStyle.Render(msg.Content)
		case "system":
			systemStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#fbbf24")).
				Align(lipgloss.Left).
				Width(m.viewport.Width)
			styledMessage = "\n" + systemStyle.Render(msg.Content) + "\n"
		}
		content.WriteString(styledMessage + "\n")
		lastRole = msg.Role
	}
	return content.String()
}

func (m Model) getAIResponse() tea.Cmd {
	return func() tea.Msg {
		// Filter out system messages for API calls
		var apiMessages []types.Message
		for _, msg := range m.messages {
			if msg.Role != "system" {
				apiMessages = append(apiMessages, msg)
			}
		}
		
		response, err := m.aicore.API.GetResponse(apiMessages, m.ai.SystemPrompt)
		if err != nil {
			return AIResponseMsg{
				message: types.Message{
					Role:    "assistant",
					Content: fmt.Sprintf("Error: %v", err),
				},
			}
		}
		return AIResponseMsg{
			message: types.Message{
				Role:    "assistant",
				Content: response,
			},
		}
	}
}

func (m Model) getAIStreamingResponse(streamingAPI api.StreamingAPI) tea.Cmd {
	return func() tea.Msg {
		// Filter out system messages for API calls
		var apiMessages []types.Message
		for _, msg := range m.messages {
			if msg.Role != "system" {
				apiMessages = append(apiMessages, msg)
			}
		}
		
		// Start streaming
		textChan, errChan := streamingAPI.GetStreamingResponse(apiMessages, m.ai.SystemPrompt)
		
		return AIStreamStartMsg{
			textChan: textChan,
			errChan:  errChan,
		}
	}
}

func (m Model) readNextChunk(textChan <-chan string, errChan <-chan error) tea.Cmd {
	return func() tea.Msg {
		select {
		case chunk, ok := <-textChan:
			if !ok {
				return AIStreamCompleteMsg{}
			}
			return AIStreamChunkMsg{
				chunk:    chunk,
				textChan: textChan,
				errChan:  errChan,
			}
		case err := <-errChan:
			if err != nil {
				return AIResponseMsg{
					message: types.Message{Role: "assistant", Content: fmt.Sprintf("Error: %v", err)},
				}
			}
			return AIStreamCompleteMsg{}
		}
	}
}

func (m Model) callAI(userInput string) tea.Cmd {
	if streamingAPI, ok := m.aicore.API.(api.StreamingAPI); ok {
		return m.getAIStreamingResponse(streamingAPI)
	} else {
		return m.getAIResponse()
	}
}

func (m Model) makeInfoPanel() string {
	currentTime := time.Now().Format("15:04:05")

	// Get conversation name or default
	conversationName := "none"
	if m.conversation.ID != 0 {
		conversationName = m.conversation.Name
	}

	// Status styling (hardcoded to online for now)
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
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
		fmt.Sprintf("%s %s", labelStyle.Render("ai:"), valueStyle.Render(m.ai.Name)),
		"",
		fmt.Sprintf("%s %s", labelStyle.Render("api:"), valueStyle.Render(m.ai.API)),
		"",
		fmt.Sprintf("%s %s", labelStyle.Render("model:"), valueStyle.Render(m.ai.Model)),
		"",
		fmt.Sprintf("%s %s", labelStyle.Render("conv:"), valueStyle.Render(conversationName)),
		"",
		fmt.Sprintf("%s %s", labelStyle.Render("status:"), statusStyle.Render("online")),
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

func loadAscii(path string) string {
	artBytes, err := os.ReadFile(path)
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

func createListDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	
	// Left-align and style the list items
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#1e40af")).
		Align(lipgloss.Left).
		Padding(0, 1).
		MarginLeft(0)
	
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e5e7eb")).
		Background(lipgloss.Color("#1e40af")).
		Align(lipgloss.Left).
		Padding(0, 1).
		MarginLeft(0)
	
	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#22d3ee")).
		Align(lipgloss.Left).
		Padding(0, 1).
		MarginLeft(0)
		
	delegate.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6b7280")).
		Align(lipgloss.Left).
		Padding(0, 1).
		MarginLeft(0)
	
	return delegate
}

func (m Model) handleSlashCommand(command string) (tea.Model, tea.Cmd) {
	m.textarea.Reset()
	
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return m, nil
	}
	
	cmd := strings.TrimPrefix(parts[0], "/")
	
	switch cmd {
	case "list":
		if len(parts) < 2 {
			return m.showError("Usage: /list [ai(s)|conversations|api(s)|model(s)]")
		}
		
		switch parts[1] {
		case "ai", "ais":
			return m.listAIs()
		case "conversations":
			return m.listConversations()
		case "api", "apis":
			return m.listAPIs()
		case "model", "models":
			if len(parts) < 3 {
				return m.showError("Usage: /list model(s) [gemini]")
			}
			return m.listModels(parts[2])
		default:
			return m.showError("Unknown list type: " + parts[1])
		}
		
	case "set":
		if len(parts) < 2 {
			return m.showError("Usage: /set [ai|api|model|prompt]")
		}
		switch parts[1] {
		case "ai":
			// Interactive selector only: /set ai
			return m.openAISelector()
		case "api":
			// Interactive selector only: /set api
			return m.openAPISelector()
		case "model":
			// Interactive selector only: /set model
			return m.openModelSelector()
		case "prompt":
			if len(parts) < 3 {
				return m.showError("Usage: /set prompt <your prompt text here>")
			}
			// Join all parts after "prompt" to get the full prompt text
			promptText := strings.Join(parts[2:], " ")
			return m.setPrompt(promptText)
		default:
			return m.showError("Unknown set type: " + parts[1])
		}
		
	case "resume":
		return m.listConversations()
		
	case "clear":
		return m.clearConversation()
		
	case "show":
		if len(parts) < 2 {
			return m.showError("Usage: /show [prompt]")
		}
		switch parts[1] {
		case "prompt":
			return m.showPrompt()
		default:
			return m.showError("Unknown show type: " + parts[1])
		}
		
	case "commands", "help":
		return m.showCommands()
		
	case "rename":
		if len(parts) < 2 {
			return m.showError("Usage: /rename <new conversation name>")
		}
		// Join all parts after "rename" to get the full name
		newName := strings.Join(parts[1:], " ")
		return m.renameConversation(newName)
		
	default:
		return m.showError("Unknown command: " + command)
	}
}
