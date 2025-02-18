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
	openAIEndpoint = "https://api.openai.com/v1/chat/completions"
	modelGPT4      = "gpt-4-turbo-preview"
	timeout        = 30 * time.Second
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
	Destination    string    `json:"destination"`
	DepartureDate  time.Time `json:"departure_date"`
	ReturnDate     time.Time `json:"return_date"`
	PreferredPrice float64   `json:"preferred_price"`
	Requirements   []string  `json:"requirements"`
	Error          string    `json:"error,omitempty"`
}

func NewInferenceEngine(apiKey string) (*InferenceEngine, error) {
	if apiKey == "" {
		return nil, errors.New("OpenAI API key is required")
	}

	return &InferenceEngine{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

type openAIRequest struct {
	Model    string      `json:"model"`
	Messages []openAIMsg `json:"messages"`
}

type openAIMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *InferenceEngine) ProcessTravelRequest(ctx context.Context, req TravelRequest) (*TravelResponse, error) {
	// Construct the system prompt
	systemPrompt := `You are an AI travel assistant. Analyze the travel request and extract key information.
Output must be valid JSON with the following structure:
{
    "destination": "city name",
    "departure_date": "YYYY-MM-DD",
    "return_date": "YYYY-MM-DD",
    "preferred_price": number,
    "requirements": ["requirement1", "requirement2"]
}`

	// Construct user prompt with query and deadline
	userPrompt := fmt.Sprintf("Travel request: %s\nBooking deadline: %s",
		req.Query,
		req.Deadline.Format(time.RFC3339))

	// Prepare OpenAI request
	openAIReq := openAIRequest{
		Model: modelGPT4,
		Messages: []openAIMsg{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	// Marshal request body
	reqBody, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", openAIEndpoint, bytes.NewBuffer(reqBody))
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
	var openAIResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for OpenAI error
	if openAIResp.Error != nil {
		return nil, fmt.Errorf("OpenAI error: %s", openAIResp.Error.Message)
	}

	// Check if we have any choices
	if len(openAIResp.Choices) == 0 {
		return nil, errors.New("no response from OpenAI")
	}

	// Parse the AI response into our TravelResponse struct
	var travelResp TravelResponse
	if err := json.Unmarshal([]byte(openAIResp.Choices[0].Message.Content), &travelResp); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &travelResp, nil
}
