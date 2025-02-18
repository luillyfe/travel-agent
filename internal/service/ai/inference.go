package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
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

type TravelRequest struct {
	Query    string
	Deadline time.Time
}

type TravelResponse struct {
	Destination    string   `json:"destination"`
	DepartureDate  string   `json:"departure_date"`
	ReturnDate     string   `json:"return_date"`
	PreferredPrice float64  `json:"preferred_price"`
	Requirements   []string `json:"requirements"`
	Error          string   `json:"error,omitempty"`
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
	Model    string          `json:"model"`
	Messages []AIProviderMsg `json:"messages"`
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

func (p *InferenceEngine) ProcessTravelRequest(ctx context.Context, req TravelRequest) (*TravelResponse, error) {
	// Construct the system prompt
	systemPrompt := `You are an AI travel assistant. Analyze the travel request and extract key information.
Output must be a single valid JSON string with no markdown formatting, no code blocks, and no backticks.
The JSON must have this exact structure:
{
    "destination": "city name",
    "departure_date": "YYYY-MM-DDTHH:MM:SSZ",
    "return_date": "YYYY-MM-DDTHH:MM:SSZ",
    "preferred_price": number,
    "requirements": ["requirement1", "requirement2"]
}
Do not include any explanation or additional text, only return the JSON object.`

	// Construct user prompt with query and deadline
	userPrompt := fmt.Sprintf("Travel request: %s\nBooking deadline: %s",
		req.Query,
		req.Deadline.Format(time.RFC3339))

	// Prepare AIProvider request
	AIProviderReq := AIProviderRequest{
		Model: model,
		Messages: []AIProviderMsg{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	// Marshal request body
	reqBody, err := json.Marshal(AIProviderReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", AIProviderEndpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	// Make request
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var AIProviderResp AIProviderResponse
	if err := json.NewDecoder(resp.Body).Decode(&AIProviderResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for AIProvider error
	if AIProviderResp.Error != nil {
		return nil, fmt.Errorf("AIProvider error: %s", AIProviderResp.Error.Message)
	}

	// Check if we have any choices
	if len(AIProviderResp.Choices) == 0 {
		return nil, errors.New("no response from AIProvider")
	}

	// Parse the AI response into our TravelResponse struct
	var travelResp TravelResponse
	if err := json.Unmarshal([]byte(AIProviderResp.Choices[0].Message.Content), &travelResp); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &travelResp, nil
}
