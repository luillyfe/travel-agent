package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
	"travel-agent/pkg/utils"
)

const (
	AIProviderEndpoint = "https://api.mistral.ai/v1/chat/completions"
	model              = "mistral-large-latest"
	timeout            = 30 * time.Second
)

type InferenceEngine struct {
	apiKey     string
	httpClient *http.Client
}

// ResponseFormat the format that the response must adhere to
type ResponseFormat struct {
	Type string `json:"type"`
}

func NewInferenceEngine(apiKey string) (*InferenceEngine, error) {
	if apiKey == "" {
		return nil, errors.New("AIProvider API key is required")
	}

	return &InferenceEngine{
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

type PromptStrategy[T any] interface {
	GetSystemPrompt() string
	GetUserPrompt(req T) string
}

type DecodingStrategy[T any] interface {
	DecodeResponse(content string) (*T, error)
}

func (p *InferenceEngine) ExtractParameters(ctx context.Context, strategy *TravelExtractionStrategy, request ExtractionRequest, decodingStrategy DecodingStrategy[TravelParameters]) (*TravelParameters, error) {
	systemPrompt := strategy.GetSystemPrompt()
	userPrompt := strategy.GetUserPrompt(request)

	// Prepare AIProvider request
	AIProviderReq := AIProviderRequest{
		Model: model,
		Messages: []AIProviderMsg{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		ResponseFormat: ResponseFormat{
			Type: "json_object",
		},
	}

	// Make HTTP request
	resp, err := p.makeRequest(ctx, AIProviderReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Log response details
	utils.LogResponseWithoutConsuming(resp)

	// Parse the AI provider response
	var AIProviderResp AIProviderResponse
	if err := json.NewDecoder(resp.Body).Decode(&AIProviderResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for AI provider errors
	if AIProviderResp.Error != nil {
		return nil, fmt.Errorf("AIProvider error: %s", AIProviderResp.Error.Message)
	}

	// Ensure we have a response
	if len(AIProviderResp.Choices) == 0 {
		return nil, errors.New("no response from AIProvider")
	}

	// Use decoding strategy to parse the response
	return decodingStrategy.DecodeResponse(AIProviderResp.Choices[0].Message.Content)
}

// Helper method for making HTTP requests
func (p *InferenceEngine) makeRequest(ctx context.Context, AIProviderReq AIProviderRequest) (*http.Response, error) {
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
