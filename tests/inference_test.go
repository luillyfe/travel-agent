package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"travel-agent/internal/service/ai"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInferenceEngine(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "valid api key",
			apiKey:  "test-key",
			wantErr: false,
		},
		{
			name:    "empty api key",
			apiKey:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := ai.NewInferenceEngine(tt.apiKey)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, engine)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, engine)
			}
		})
	}
}

func TestInferenceEngine_ExtractParameters(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		// Decode request body
		var req ai.AIProviderRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Verify request structure
		assert.Equal(t, "mistral-large-latest", req.Model)
		assert.Len(t, req.Messages, 2)
		assert.Equal(t, "json_object", req.ResponseFormat.Type)

		// Return mock response
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
						Role: "assistant",
						Content: `{
                            "departure_city": "New York",
                            "destination": "Paris",
                            "departure_date": "2025-03-15T00:00:00Z",
                            "return_date": "2025-03-22T00:00:00Z",
                            "preferences": {
                                "budget_range": {
                                    "min": 1000,
                                    "max": 2000
                                },
                                "travel_class": "economy",
                                "activities": ["sightseeing"],
                                "dietary_restrictions": []
                            }
                        }`,
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create inference engine with mock server
	engine, err := ai.NewInferenceEngine("test-key")
	require.NoError(t, err)

	// Override endpoint for testing
	ai.AIProviderEndpoint = server.URL

	// Create test request
	strategy := &ai.TravelExtractionStrategy{}
	decoder := &ai.TravelDecodingStrategy{}
	request := ai.ExtractionRequest{
		Query:    "I want to travel from New York to Paris",
		Deadline: time.Now().Add(24 * time.Hour),
	}

	// Test extraction
	params, err := engine.ExtractParameters(context.Background(), strategy, request, decoder)
	require.NoError(t, err)
	assert.NotNil(t, params)
	assert.Equal(t, "New York", params.DepartureCity)
	assert.Equal(t, "Paris", params.Destination)
}

func TestInferenceEngine_ExtractParameters_Errors(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(http.Handler) http.Handler
		wantErr   string
	}{
		{
			name: "ai provider error",
			setupMock: func(h http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := ai.AIProviderResponse{
						Error: &struct {
							StatusCode int    `json:"status_code"`
							Type       string `json:"type"`
							Message    string `json:"message"`
						}{
							StatusCode: 400,
							Message:    "invalid request",
						},
					}
					json.NewEncoder(w).Encode(response)
				})
			},
			wantErr: "AIProvider error: invalid request",
		},
		{
			name: "no choices in response",
			setupMock: func(h http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					json.NewEncoder(w).Encode(ai.AIProviderResponse{
						Choices: []struct {
							Index   int `json:"index"`
							Message struct {
								Role    string `json:"role"`
								Content string `json:"content"`
							} `json:"message"`
							FinishReason string `json:"finish_reason"`
						}{},
					})
				})
			},
			wantErr: "no response from AIProvider",
		},
		{
			name: "invalid json in response",
			setupMock: func(h http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					json.NewEncoder(w).Encode(ai.AIProviderResponse{
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
									Content: "invalid json",
								},
							},
						},
					})
				})
			},
			wantErr: "failed to parse travel parameters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.setupMock(nil))
			defer server.Close()

			engine, _ := ai.NewInferenceEngine("test-key")
			ai.AIProviderEndpoint = server.URL

			strategy := &ai.TravelExtractionStrategy{}
			decoder := &ai.TravelDecodingStrategy{}
			request := ai.ExtractionRequest{
				Query:    "test query",
				Deadline: time.Now().Add(24 * time.Hour),
			}

			_, err := engine.ExtractParameters(context.Background(), strategy, request, decoder)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestInferenceEngine_Context(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	engine, _ := ai.NewInferenceEngine("test-key")
	ai.AIProviderEndpoint = server.URL

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	strategy := &ai.TravelExtractionStrategy{}
	decoder := &ai.TravelDecodingStrategy{}
	request := ai.ExtractionRequest{
		Query:    "test query",
		Deadline: time.Now().Add(24 * time.Hour),
	}

	_, err := engine.ExtractParameters(ctx, strategy, request, decoder)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}
