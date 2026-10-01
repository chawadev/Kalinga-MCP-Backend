package gemini

import (
	"context"
	"fmt"

	genai "google.golang.org/genai"
)

type Client struct {
	client *genai.Client
}

func NewClient(apiKey string) (*Client, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &Client{client: client}, nil
}

func (c *Client) Close() error {
	// The new SDK doesn't have a Close method
	return nil
}

func (c *Client) GenerateContent(ctx context.Context, model string, prompt string) (string, error) {
	result, err := c.client.Models.GenerateContent(ctx, model, genai.Text(prompt), nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if result == nil {
		return "", fmt.Errorf("no response generated")
	}

	return result.Text(), nil
}
