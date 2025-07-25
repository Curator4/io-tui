package chat

// NewModel creates and returns a new chat model with default values
func NewModel() Model {
	return Model{
		Chatlog:   []Message{},
		Image:     "",
		InputText: "",
		IsLoading: false,
		Width:     80,
		Height:    24,
	}
}