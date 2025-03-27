package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
	"travel-agent/internal/models"
	"travel-agent/pkg/utils"
)

var AIProviderEndpoint = "https://api.mistral.ai/v1/chat/completions"

const (
	model   = "mistral-large-latest"
	timeout = 30 * time.Second
)

type InferenceEngine[T models.TravelOutput, R models.TravelInput] struct {
	apiKey     string
	httpClient *http.Client
}

// ResponseFormat the format that the response must adhere to
type ResponseFormat struct {
	Type string `json:"type"`
}

func NewInferenceEngine[T models.TravelOutput, R models.TravelInput](apiKey string) (*InferenceEngine[T, R], error) {
	if apiKey == "" {
		return nil, errors.New("AIProvider API key is required")
	}

	return &InferenceEngine[T, R]{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

type AIProviderRequest struct {
	Model          string          `json:"model"`
	Messages       []AIProviderMsg `json:"messages"`
	ResponseFormat ResponseFormat  `json:"response_format"`
}

type AIProviderMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// TODO: Remove vendor specific AIProviderResponse struct
// Mistral API response (vendor specific)
type AIProviderResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		StatusCode int    `json:"status_code"`
		Type       string `json:"type"`
		Message    string `json:"message"`
	} `json:"error"`
}

type PromptStrategy[R any] interface {
	GetSystemPrompt() string
	GetUserPrompt(req R) string
}

type DecodingStrategy[T any] interface {
	DecodeResponse(content string) (*T, error)
}

func (p *InferenceEngine[T, R]) ProcessRequest(
	ctx context.Context,
	promptStrategy PromptStrategy[R],
	request R,
	decodingStrategy DecodingStrategy[T],
) (*T, error) {
	// Get prompts
	systemPrompt := promptStrategy.GetSystemPrompt()
	userPrompt := promptStrategy.GetUserPrompt(request)

	// Prepare request
	aiReq := AIProviderRequest{
		Model: model,
		Messages: []AIProviderMsg{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		ResponseFormat: ResponseFormat{
			Type: "json_object",
		},
	}

	// Make request
	resp, err := p.makeRequest(ctx, aiReq)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log the error but don't override any existing error return
			fmt.Printf("error closing response body: %v\n", err)
		}
	}()

	// Log response if enabled
	if err := utils.LogResponseWithoutConsuming(resp); err != nil {
		fmt.Printf("failed to log response: %v\n", err)
	}

	// Parse response
	var aiResp AIProviderResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for errors
	if aiResp.Error != nil {
		return nil, fmt.Errorf("AI provider error: %s", aiResp.Error.Message)
	}

	// Ensure we have a response
	if len(aiResp.Choices) == 0 {
		return nil, errors.New("no response from AI provider")
	}

	// Decode the response
	return decodingStrategy.DecodeResponse(aiResp.Choices[0].Message.Content)
}

// Helper method for making HTTP requests
func (p *InferenceEngine[T, R]) makeRequest(ctx context.Context, AIProviderReq AIProviderRequest) (*http.Response, error) {
	reqBody, err := json.Marshal(AIProviderReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", AIProviderEndpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	return p.httpClient.Do(httpReq)
}
