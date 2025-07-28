package api

import (
	"fmt"
    "context"
    "os"
    "strings"
    "google.golang.org/genai"
    "github.com/curator4/io-tui/types"
)

type GeminiAPI struct {
    client *genai.Client
}


// defineManifestFunction creates the function schema for manifesting characters
func defineManifestFunction() *genai.FunctionDeclaration {
    return &genai.FunctionDeclaration{
        Name: "manifest_character",
        Description: "Call this function ONLY when the user mentions 'manifest' with a character name AND provides an image URL. If no image URL is provided, do NOT call this function - instead ask the user to provide an image URL.",
        Parameters: &genai.Schema{
            Type: genai.TypeObject,
            Properties: map[string]*genai.Schema{
                "name": {
                    Type: genai.TypeString,
                    Description: "Just the character name (e.g., 'L', 'Sherlock Holmes', 'Tony Stark')",
                },
                "image_url": {
                    Type: genai.TypeString,
                    Description: "The EXACT image URL provided by the user in their message. Must be a direct link to a PNG or JPG file. NEVER use placeholders or make up URLs.",
                },
                "description": {
                    Type: genai.TypeString,
                    Description: "You must provide a detailed description of the character's personality, traits, background, and how they should behave. Include specific details about their speaking style, mannerisms, and key characteristics. This will be used to automatically generate their system prompt.",
                },
            },
            Required: []string{"name", "image_url", "description"},
        },
    }
}

func NewGeminiAPI() (*GeminiAPI, error) {
    ctx := context.Background()
    
    // Get API key from environment, or read from demo file
    apiKey := os.Getenv("GEMINI_API_KEY")
    if apiKey == "" {
        if keyBytes, err := os.ReadFile("demo_api_key.txt"); err == nil {
            content := strings.TrimSpace(string(keyBytes))
            lines := strings.Split(content, "\n")
            for _, line := range lines {
                line = strings.TrimSpace(line)
                if !strings.HasPrefix(line, "#") && line != "" {
                    apiKey = line
                    break
                }
            }
        }
    }
    
    if apiKey == "" {
        return nil, fmt.Errorf("No API key found. Either export GEMINI_API_KEY=your_key or add your key to demo_api_key.txt")
    }
    
    // Check if it's the placeholder demo key
    if strings.Contains(apiKey, "Demo_Key_Replace") {
        return nil, fmt.Errorf("Demo API key is placeholder. Replace with real key in demo_api_key.txt")
    }
    
    // Create client with API key
    client, err := genai.NewClient(ctx, &genai.ClientConfig{
        APIKey: apiKey,
    })
    
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
    
    // Create config with system instruction and function tools
    var config *genai.GenerateContentConfig
    if systemPrompt != "" {
        config = &genai.GenerateContentConfig{
            SystemInstruction: genai.NewContentFromText(systemPrompt, genai.RoleUser),
            Tools: []*genai.Tool{
                {
                    FunctionDeclarations: []*genai.FunctionDeclaration{
                        defineManifestFunction(),
                    },
                },
            },
        }
    } else {
        config = &genai.GenerateContentConfig{
            Tools: []*genai.Tool{
                {
                    FunctionDeclarations: []*genai.FunctionDeclaration{
                        defineManifestFunction(),
                    },
                },
            },
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
    response, err := g.GetResponseWithFunctions(messages, systemPrompt)
    if err != nil {
        return "", err
    }
    return response.Text, nil
}

func (g *GeminiAPI) GetResponseWithFunctions(messages []types.Message, systemPrompt string) (*ResponseWithFunctions, error) {
    ctx := context.Background()
    
    if len(messages) == 0 {
        return &ResponseWithFunctions{Text: "No messages to process"}, nil
    }
    
    // Convert ALL messages to genai Content format
    var contents []*genai.Content
    for _, msg := range messages {
        var role genai.Role
        if msg.Role == "user" {
            role = genai.RoleUser
        } else if msg.Role == "assistant" {
            role = genai.RoleModel
        } else {
            continue // Skip system messages
        }
        contents = append(contents, genai.NewContentFromText(msg.Content, role))
    }
    
    // Create tools config
    tools := []*genai.Tool{
        {
            FunctionDeclarations: []*genai.FunctionDeclaration{
                defineManifestFunction(),
            },
        },
    }
    
    config := &genai.GenerateContentConfig{
        Tools: tools,
    }
    
    if systemPrompt != "" {
        config.SystemInstruction = genai.NewContentFromText(systemPrompt, genai.RoleUser)
    }
    
    // Use models.generate_content with full conversation
    res, err := g.client.Models.GenerateContent(ctx, "gemini-2.5-flash", contents, config)
    if err != nil {
        return nil, err
    }
    
    response := &ResponseWithFunctions{}
    
    if len(res.Candidates) == 0 || res.Candidates[0].Content == nil || len(res.Candidates[0].Content.Parts) == 0 {
        response.Text = "No response received"
        return response, nil
    }
    
    if len(res.Candidates) > 0 && len(res.Candidates[0].Content.Parts) > 0 {
        for _, part := range res.Candidates[0].Content.Parts {
            // Handle text parts
            if part.Text != "" {
                response.Text += part.Text
            }
            
            // Handle function call parts
            if part.FunctionCall != nil {
                functionCall := FunctionCall{
                    Name: part.FunctionCall.Name,
                    Args: make(map[string]interface{}),
                }
                
                // Convert function arguments
                for key, value := range part.FunctionCall.Args {
                    functionCall.Args[key] = value
                }
                
                response.FunctionCalls = append(response.FunctionCalls, functionCall)
            }
        }
    }
    
    if response.Text == "" && len(response.FunctionCalls) == 0 {
        response.Text = "No response received"
    }
    
    return response, nil
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
