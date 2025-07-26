package api

import (
    "context"
    "google.golang.org/genai"
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

func (g *GeminiAPI) GetResponse(context string) (string, error) {
    ctx := context.Background()
    result, err := g.client.Models.GenerateContent(
        ctx,
        "gemini-2.5-flash",
        genai.Text(context),
        nil,
    )
    if err != nil {
        return "", err
    }
    
    return result.Text(), nil
}

