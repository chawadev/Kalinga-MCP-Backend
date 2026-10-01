package ai

import (
	"context"
	"fmt"

	"github.com/chawadev/kalinga-backend/internal/config"
	genai "google.golang.org/genai"
)

type Service struct {
	client *genai.Client
	model  string
}

func NewService(cfg *config.Config) (*Service, error) {
	if cfg.GeminiAPIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is not configured")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.GeminiAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &Service{
		client: client,
		model:  "gemini-2.5-flash",
	}, nil
}

func (s *Service) Close() error {
	// The new SDK doesn't have a Close method
	return nil
}

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type FunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

type GenerationRequest struct {
	Message    string
	Tools      []Tool
	System     string
	History    []map[string]interface{}
	ToolResult *ToolResult
}

type ToolResult struct {
	Name     string
	Response interface{}
}

type GenerationResponse struct {
	Text          string
	FunctionCalls []FunctionCall
}

func (s *Service) GenerateContent(ctx context.Context, req *GenerationRequest) (*GenerationResponse, error) {
	// Build prompt with system instruction if provided
	prompt := req.Message
	if req.System != "" {
		prompt = req.System + "\n\n" + req.Message
	}

	// Generate content using the new SDK
	result, err := s.client.Models.GenerateContent(ctx, s.model, genai.Text(prompt), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	response := &GenerationResponse{}

	// Extract text from response
	if result != nil {
		response.Text = result.Text()
	}

	return response, nil
}

