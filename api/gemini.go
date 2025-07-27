package api

import (
	"fmt"
    "context"
    "google.golang.org/genai"
    "github.com/curator4/io-tui/types"
)

type GeminiAPI struct {
    client *genai.Client
}

func NewGeminiAPI() (*GeminiAPI, error) {
    ctx := context.Background()
    client, err := genai.NewClient(ctx, nil)
    if err != nil {
        return nil, err
    }
    
    return &GeminiAPI{
        client: client,
    }, nil
}

func (g *GeminiAPI) prepareChatSession(messages []types.Message, systemPrompt string) (*genai.Chat, string, error) {
    ctx := context.Background()
    
    if len(messages) == 0 {
        return nil, "", fmt.Errorf("no messages to process")
    }
    
    // Convert ALL messages to genai Content format for history
    var history []*genai.Content
    var lastUserMessage string
    
    // Add all messages except the very last one to history
    for i, msg := range messages {
        var role genai.Role
        if msg.Role == "user" {
            role = genai.RoleUser
        } else if msg.Role == "assistant" {
            role = genai.RoleModel
        }
        
        // Add all messages except the last one to history
        if i < len(messages)-1 {
            history = append(history, genai.NewContentFromText(msg.Content, role))
        } else {
            // The last message should be a user message that we'll send
            if msg.Role == "user" {
                lastUserMessage = msg.Content
            }
        }
    }
    
    // Create config with system instruction if provided
    var config *genai.GenerateContentConfig
    if systemPrompt != "" {
        config = &genai.GenerateContentConfig{
            SystemInstruction: genai.NewContentFromText(systemPrompt, genai.RoleUser),
        }
    }
    
    // Create chat with full conversation history and system instruction
    chat, err := g.client.Chats.Create(ctx, "gemini-2.5-flash", config, history)
    if err != nil {
        return nil, "", err
    }
    
    return chat, lastUserMessage, nil
}

func (g *GeminiAPI) GetResponse(messages []types.Message, systemPrompt string) (string, error) {
    chat, lastUserMessage, err := g.prepareChatSession(messages, systemPrompt)
    if err != nil {
        return "No messages to process", nil
    }
    
    // Send the latest user message
    ctx := context.Background()
    res, err := chat.SendMessage(ctx, genai.Part{Text: lastUserMessage})
    if err != nil {
        return "", err
    }
    
    if len(res.Candidates) > 0 && len(res.Candidates[0].Content.Parts) > 0 {
        return res.Candidates[0].Content.Parts[0].Text, nil
    }
    
    return "No response received", nil
}

func (g *GeminiAPI) GetStreamingResponse(messages []types.Message, systemPrompt string) (<-chan string, <-chan error) {
	textChan := make(chan string)
	errChan := make(chan error, 1)

	go func() {
		defer close(textChan)
		defer close(errChan)

		chat, lastUserMessage, err := g.prepareChatSession(messages, systemPrompt)
		if err != nil {
			errChan <- err
			return
		}

		ctx := context.Background()
		stream := chat.SendMessageStream(ctx, genai.Part{Text: lastUserMessage})

		for chunk := range stream {
			// Add safety checks to prevent segmentation faults
			if chunk == nil {
				continue
			}
			if len(chunk.Candidates) > 0 && 
			   chunk.Candidates[0] != nil && 
			   chunk.Candidates[0].Content != nil &&
			   len(chunk.Candidates[0].Content.Parts) > 0 &&
			   chunk.Candidates[0].Content.Parts[0].Text != "" {
				textChan <- chunk.Candidates[0].Content.Parts[0].Text
			}
		}
	}()

	return textChan, errChan
}
