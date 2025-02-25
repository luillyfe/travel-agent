package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"travel-agent/internal/models"
	"travel-agent/internal/service/ai"
)

// Mock prompt strategy
type MockPromptStrategy struct{}

func (m MockPromptStrategy) GetSystemPrompt() string {
	return "system prompt"
}

func (m MockPromptStrategy) GetUserPrompt(req models.MockTravelRequest) string {
	return "user prompt"
}

// Mock decoding strategy
type MockDecodingStrategy struct{}

func (m MockDecodingStrategy) DecodeResponse(content string) (*models.MockTravelResponse, error) {
	return &models.MockTravelResponse{}, nil
}

func TestNewInferenceEngine(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "Valid API Key",
			apiKey:  "valid-key",
			wantErr: false,
		},
		{
			name:    "Empty API Key",
			apiKey:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := ai.NewInferenceEngine[models.MockTravelResponse, models.MockTravelRequest](tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewInferenceEngine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && engine == nil {
				t.Error("NewInferenceEngine() returned nil engine when no error expected")
			}
		})
	}
}

func TestProcessRequest(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("Invalid Authorization header")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("Invalid Content-Type header")
		}

		// Send mock response
		response := ai.AIProviderResponse{
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: `{}`, // Empty JSON object since models.MockTravelResponse is empty
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Override the endpoint for testing
	originalEndpoint := ai.AIProviderEndpoint
	ai.AIProviderEndpoint = server.URL
	defer func() { ai.AIProviderEndpoint = originalEndpoint }()

	// Create inference engine
	engine, err := ai.NewInferenceEngine[models.MockTravelResponse, models.MockTravelRequest]("test-key")
	if err != nil {
		t.Fatalf("Failed to create inference engine: %v", err)
	}

	// Test ProcessRequest
	ctx := context.Background()
	input := models.MockTravelRequest{}
	promptStrategy := MockPromptStrategy{}
	decodingStrategy := MockDecodingStrategy{}

	output, err := engine.ProcessRequest(ctx, promptStrategy, input, decodingStrategy)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if output == nil {
		t.Error("ProcessRequest returned nil output")
	}
}

func TestProcessRequestTimeout(t *testing.T) {
	// Create a slow server that will trigger timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * 30 * time.Second) // Sleep longer than timeout
	}))
	defer server.Close()

	ai.AIProviderEndpoint = server.URL
	engine, _ := ai.NewInferenceEngine[models.MockTravelResponse, models.MockTravelRequest]("test-key")

	ctx := context.Background()
	input := models.MockTravelRequest{}
	promptStrategy := MockPromptStrategy{}
	decodingStrategy := MockDecodingStrategy{}

	_, err := engine.ProcessRequest(ctx, promptStrategy, input, decodingStrategy)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestProcessRequestError(t *testing.T) {
	// Create server that returns error response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ai.AIProviderResponse{
			Error: &struct {
				StatusCode int    `json:"status_code"`
				Type       string `json:"type"`
				Message    string `json:"message"`
			}{
				StatusCode: 400,
				Type:       "invalid_request",
				Message:    "Test error",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	ai.AIProviderEndpoint = server.URL
	engine, _ := ai.NewInferenceEngine[models.MockTravelResponse, models.MockTravelRequest]("test-key")

	ctx := context.Background()
	input := models.MockTravelRequest{}
	promptStrategy := MockPromptStrategy{}
	decodingStrategy := MockDecodingStrategy{}

	_, err := engine.ProcessRequest(ctx, promptStrategy, input, decodingStrategy)
	if err == nil {
		t.Error("Expected error response, got nil")
	}
}
