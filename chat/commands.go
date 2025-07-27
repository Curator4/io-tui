package chat

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/curator4/io-tui/api"
	"github.com/curator4/io-tui/db"
	"github.com/curator4/io-tui/types"
)

// List item for AI selection
type aiItem struct {
	ai db.AI
}

func (a aiItem) FilterValue() string { return a.ai.Name }
func (a aiItem) Title() string       { return a.ai.Name }
func (a aiItem) Description() string { 
	status := ""
	if a.ai.IsActive {
		status = " (active)"
	}
	return fmt.Sprintf("%s - %s%s", a.ai.API, a.ai.Model, status)
}

// List item for API selection
type apiItem struct {
	name string
	info api.APIInfo
}

func (a apiItem) FilterValue() string { return a.name }
func (a apiItem) Title() string       { return a.info.Name }
func (a apiItem) Description() string { 
	return fmt.Sprintf("Default: %s (%d models available)", a.info.DefaultModel, len(a.info.Models))
}

// List item for model selection
type modelItem struct {
	name string
	api  string
}

func (m modelItem) FilterValue() string { return m.name }
func (m modelItem) Title() string       { return m.name }
func (m modelItem) Description() string { 
	return fmt.Sprintf("API: %s", m.api)
}

// List item for conversation selection
type conversationItem struct {
	conversation db.Conversation
}

func (c conversationItem) FilterValue() string { return c.conversation.Name }
func (c conversationItem) Title() string       { return c.conversation.Name }
func (c conversationItem) Description() string { 
	status := ""
	if c.conversation.IsActive {
		status = " (active)"
	}
	return fmt.Sprintf("Created: %s%s", c.conversation.Created, status)
}

// List functions - opens interactive lists
func (m Model) listAIs() (tea.Model, tea.Cmd) {
	ais, err := db.ListAIs(m.database)
	if err != nil {
		return m.showError("Error loading AIs: " + err.Error())
	}
	
	var items []list.Item
	for _, ai := range ais {
		items = append(items, aiItem{ai: ai})
	}
	
	m.list.SetItems(items)
	m.list.Title = "Available AIs (Esc to close)"
	
	// Configure for view-only mode
	m.list.SetShowStatusBar(false)
	m.list.SetFilteringEnabled(true)
	m.list.SetShowHelp(true)
	
	m.viewMode = listMode
	
	return m, nil
}

// Opens AI selector for switching
func (m Model) openAISelector() (tea.Model, tea.Cmd) {
	ais, err := db.ListAIs(m.database)
	if err != nil {
		return m.showError("Error loading AIs: " + err.Error())
	}
	
	var items []list.Item
	for _, ai := range ais {
		items = append(items, aiItem{ai: ai})
	}
	
	m.list.SetItems(items)
	m.list.Title = "Select AI (Enter to switch, Esc to cancel)"
	m.viewMode = listMode
	
	return m, nil
}

// Opens API selector for switching
func (m Model) openAPISelector() (tea.Model, tea.Cmd) {
	var items []list.Item
	for name, info := range api.AvailableAPIs {
		items = append(items, apiItem{name: name, info: info})
	}
	
	m.list.SetItems(items)
	m.list.Title = "Select API (Enter to switch, Esc to cancel)"
	m.viewMode = listMode
	
	return m, nil
}

// Opens model selector for current API
func (m Model) openModelSelector() (tea.Model, tea.Cmd) {
	// Get models for current API
	apiInfo, exists := api.AvailableAPIs[m.ai.API]
	if !exists {
		return m.showError("Current API not found in available APIs")
	}
	
	var items []list.Item
	for _, model := range apiInfo.Models {
		items = append(items, modelItem{name: model, api: m.ai.API})
	}
	
	m.list.SetItems(items)
	m.list.Title = fmt.Sprintf("Select %s Model (Enter to switch, Esc to cancel)", apiInfo.Name)
	m.viewMode = listMode
	
	return m, nil
}

func (m Model) listConversations() (tea.Model, tea.Cmd) {
	conversations, err := db.ListConversationsByAI(m.database, m.ai.ID)
	if err != nil {
		return m.showError("Error loading conversations: " + err.Error())
	}
	
	var items []list.Item
	for _, conv := range conversations {
		items = append(items, conversationItem{conversation: conv})
	}
	
	m.list.SetItems(items)
	m.list.Title = "Available Conversations (Enter to resume, Esc to close)"
	
	// Configure for view-only mode
	m.list.SetShowStatusBar(false)
	m.list.SetFilteringEnabled(true)
	m.list.SetShowHelp(true)
	
	m.viewMode = listMode
	
	return m, nil
}

func (m Model) listAPIs() (tea.Model, tea.Cmd) {
	var items []list.Item
	for name, info := range api.AvailableAPIs {
		items = append(items, apiItem{name: name, info: info})
	}
	
	m.list.SetItems(items)
	m.list.Title = "Available APIs (Esc to close)"
	
	// Configure for view-only mode
	m.list.SetShowStatusBar(false)
	m.list.SetFilteringEnabled(true)
	m.list.SetShowHelp(true)
	
	m.viewMode = listMode
	
	return m, nil
}

func (m Model) listModels(apiName string) (tea.Model, tea.Cmd) {
	// Check if API exists
	apiInfo, exists := api.AvailableAPIs[apiName]
	if !exists {
		return m.showError("Unknown API: " + apiName)
	}
	
	var items []list.Item
	for _, model := range apiInfo.Models {
		items = append(items, modelItem{name: model, api: apiName})
	}
	
	m.list.SetItems(items)
	m.list.Title = fmt.Sprintf("%s Models (Esc to close)", apiInfo.Name)
	
	// Configure for view-only mode
	m.list.SetShowStatusBar(false)
	m.list.SetFilteringEnabled(true)
	m.list.SetShowHelp(true)
	
	m.viewMode = listMode
	
	return m, nil
}

// Set functions
func (m Model) setAI(name string) (tea.Model, tea.Cmd) {
	// Check if this AI is already active
	if m.ai.Name == name {
		return m.showError(fmt.Sprintf("AI '%s' is already active! 🎯", name))
	}
	
	newAI, err := db.SetActiveAI(m.database, name)
	if err != nil {
		return m.showError("Error switching to AI '" + name + "': " + err.Error())
	}
	
	// Update model with new AI
	m.ai = newAI
	
	// Load new ASCII art
	m.ascii = loadAscii(newAI.AsciiArtPath)
	
	// Clear active conversation since we switched AIs
	m.conversation = db.Conversation{}
	
	// Clear chat log completely
	m.messages = []types.Message{}
	m.viewport.SetContent(m.formatMessages())
	m.viewport.GotoBottom()
	
	// Return to chat mode
	m.viewMode = chatMode
	
	// Set processing status before getting AI introduction
	m.statusPanel.status = Processing
	
	// Get AI introduction
	return m, m.getAIIntroduction()
}

func (m Model) setAPI(apiName string) (tea.Model, tea.Cmd) {
	// Check if API exists
	apiInfo, exists := api.AvailableAPIs[apiName]
	if !exists {
		return m.showError("Unknown API: " + apiName)
	}
	
	// Update the active AI's API and set to default model
	updatedAI, err := db.UpdateActiveAIAPI(m.database, apiName, apiInfo.DefaultModel)
	if err != nil {
		return m.showError("Error switching to API '" + apiName + "': " + err.Error())
	}
	
	// Update model with new AI info
	m.ai = updatedAI
	
	// Clear active conversation since we switched APIs
	m.conversation = db.Conversation{}
	
	// Clear chat log and add success message
	m.messages = []types.Message{}
	successMsg := types.Message{
		Role:    "system",
		Content: fmt.Sprintf("Switched to API: %s (Model: %s)", apiInfo.Name, apiInfo.DefaultModel),
	}
	m.messages = append(m.messages, successMsg)
	m.viewport.SetContent(m.formatMessages())
	m.viewport.GotoBottom()
	
	// Return to chat mode
	m.viewMode = chatMode
	
	return m, nil
}

func (m Model) setModel(modelName string) (tea.Model, tea.Cmd) {
	// Check if model exists in current API
	apiInfo, exists := api.AvailableAPIs[m.ai.API]
	if !exists {
		return m.showError("Current API not found in available APIs")
	}
	
	// Check if model is available for current API
	modelExists := false
	for _, availableModel := range apiInfo.Models {
		if availableModel == modelName {
			modelExists = true
			break
		}
	}
	
	if !modelExists {
		return m.showError(fmt.Sprintf("Model '%s' not available for API '%s'", modelName, m.ai.API))
	}
	
	// Update the active AI's model
	updatedAI, err := db.UpdateActiveAIModel(m.database, modelName)
	if err != nil {
		return m.showError("Error switching to model '" + modelName + "': " + err.Error())
	}
	
	// Update model with new AI info
	m.ai = updatedAI
	
	// Clear active conversation since we switched models
	m.conversation = db.Conversation{}
	
	// Clear chat log and add success message
	m.messages = []types.Message{}
	successMsg := types.Message{
		Role:    "system",
		Content: fmt.Sprintf("Switched to model: %s", modelName),
	}
	m.messages = append(m.messages, successMsg)
	m.viewport.SetContent(m.formatMessages())
	m.viewport.GotoBottom()
	
	// Return to chat mode
	m.viewMode = chatMode
	
	return m, nil
}

func (m Model) resumeConversation(conversationID int) (tea.Model, tea.Cmd) {
	// Set this conversation as active
	conversation, err := db.SetActiveConversation(m.database, conversationID)
	if err != nil {
		return m.showError("Error resuming conversation: " + err.Error())
	}
	
	// Load messages for this conversation
	dbMessages, err := db.LoadMessages(m.database, conversationID)
	if err != nil {
		return m.showError("Error loading conversation messages: " + err.Error())
	}
	
	// Convert db.Message to types.Message
	m.messages = []types.Message{}
	for _, dbMsg := range dbMessages {
		typeMsg := types.Message{
			Role:    dbMsg.Role,
			Content: dbMsg.Content,
		}
		m.messages = append(m.messages, typeMsg)
	}
	
	// Update model with resumed conversation
	m.conversation = conversation
	
	// Add success message to the loaded conversation
	if len(dbMessages) > 0 {
		successMsg := types.Message{
			Role:    "system", 
			Content: fmt.Sprintf("✨ Resumed conversation: %s (%d messages)", conversation.Name, len(dbMessages)),
		}
		m.messages = append(m.messages, successMsg)
	} else {
		successMsg := types.Message{
			Role:    "system",
			Content: fmt.Sprintf("✨ Resumed conversation: %s (empty)", conversation.Name),
		}
		m.messages = append(m.messages, successMsg)
	}
	
	// Update viewport with all messages
	if m.viewport.Height > 0 {
		m.viewport.SetContent(m.formatMessages())
		m.viewport.GotoBottom()
	}
	
	// Return to chat mode
	m.viewMode = chatMode
	
	return m, nil
}

func (m Model) clearConversation() (tea.Model, tea.Cmd) {
	// Clear active conversation in database
	err := db.ClearActiveConversations(m.database)
	if err != nil {
		return m.showError("Error clearing conversation: " + err.Error())
	}
	
	// Clear local conversation and messages
	m.conversation = db.Conversation{}
	m.messages = []types.Message{}
	
	// Update viewport
	if m.viewport.Height > 0 {
		m.viewport.SetContent(m.formatMessages())
		m.viewport.GotoTop()
	}
	
	return m, nil
}

func (m Model) setPrompt(promptText string) (tea.Model, tea.Cmd) {
	// Update the active AI's prompt in database
	updatedAI, err := db.UpdateActiveAIPrompt(m.database, promptText)
	if err != nil {
		return m.showError("Error updating prompt: " + err.Error())
	}
	
	// Update local AI state
	m.ai = updatedAI
	
	// Clear current conversation since we're changing the AI's behavior
	err = db.ClearActiveConversations(m.database)
	if err != nil {
		return m.showError("Error clearing conversation: " + err.Error())
	}
	
	// Clear local conversation and messages
	m.conversation = db.Conversation{}
	m.messages = []types.Message{}
	
	// Add success message to fresh conversation
	successMsg := types.Message{
		Role:    "system",
		Content: fmt.Sprintf("✨ Updated system prompt for %s (conversation cleared)", updatedAI.Name),
	}
	m.messages = append(m.messages, successMsg)
	
	// Update viewport
	if m.viewport.Height > 0 {
		m.viewport.SetContent(m.formatMessages())
		m.viewport.GotoTop()
	}
	
	return m, nil
}

func (m Model) showPrompt() (tea.Model, tea.Cmd) {
	// Display the current AI's system prompt
	promptMsg := types.Message{
		Role:    "system",
		Content: fmt.Sprintf("Current system prompt for %s:\n\n%s", m.ai.Name, m.ai.SystemPrompt),
	}
	m.messages = append(m.messages, promptMsg)
	
	// Update viewport
	if m.viewport.Height > 0 {
		m.viewport.SetContent(m.formatMessages())
		m.viewport.GotoBottom()
	}
	
	return m, nil
}

func (m Model) showCommands() (tea.Model, tea.Cmd) {
	commandsText := `Available Commands:

📋 Listing:
  /list ai(s)              - Show all available AIs
  /list conversations      - Show all conversations
  /list api(s)             - Show all available APIs
  /list model(s) <api>     - Show models for specific API

⚙️  Configuration:
  /set ai                  - Open AI selector (interactive)
  /set api                 - Open API selector (interactive)
  /set model               - Open model selector (interactive)
  /set prompt <text>       - Update AI system prompt (clears conversation)

💬 Conversations:
  /resume                  - List and resume previous conversations
  /clear                   - Clear current conversation
  /rename <name>           - Rename current conversation

🔍 Information:
  /show prompt             - Display current AI system prompt
  /commands, /help         - Show this help message

💡 Tips:
  - Use /clear to clear this help message and start fresh`

	commandsMsg := types.Message{
		Role:    "system",
		Content: commandsText,
	}
	m.messages = append(m.messages, commandsMsg)
	
	// Update viewport
	if m.viewport.Height > 0 {
		m.viewport.SetContent(m.formatMessages())
		m.viewport.GotoBottom()
	}
	
	return m, nil
}

func (m Model) renameConversation(newName string) (tea.Model, tea.Cmd) {
	// Check if there's an active conversation to rename
	if m.conversation.ID == 0 {
		return m.showError("No active conversation to rename. Start chatting to create one!")
	}
	
	// Update conversation name in database
	err := db.RenameConversation(m.database, m.conversation.ID, newName)
	if err != nil {
		return m.showError("Error renaming conversation: " + err.Error())
	}
	
	// Update local conversation name
	m.conversation.Name = newName
	
	// Add success message
	successMsg := types.Message{
		Role:    "system",
		Content: fmt.Sprintf("✨ Renamed conversation to: %s", newName),
	}
	m.messages = append(m.messages, successMsg)
	
	// Update viewport
	if m.viewport.Height > 0 {
		m.viewport.SetContent(m.formatMessages())
		m.viewport.GotoBottom()
	}
	
	return m, nil
}

// Helper functions
func (m Model) showError(msg string) (tea.Model, tea.Cmd) {
	errorMsg := types.Message{
		Role:    "system",
		Content: msg,
	}
	m.messages = append(m.messages, errorMsg)
	m.viewport.SetContent(m.formatMessages())
	m.viewport.GotoBottom()
	
	return m, nil
}

func (m Model) getAIIntroduction() tea.Cmd {
	return func() tea.Msg {
		// Create a simple introduction prompt
		introMessages := []types.Message{
			{
				Role:    "user",
				Content: "Please introduce yourself briefly in a friendly way. Keep it to 1-2 sentences.",
			},
		}
		
		response, err := m.aicore.API.GetResponse(introMessages, m.ai.SystemPrompt)
		if err != nil {
			return AIIntroductionMsg{
				message: types.Message{
					Role:    "assistant",
					Content: fmt.Sprintf("Hello! I'm %s 👋", m.ai.Name),
				},
			}
		}
		return AIIntroductionMsg{
			message: types.Message{
				Role:    "assistant", 
				Content: response,
			},
		}
	}
}