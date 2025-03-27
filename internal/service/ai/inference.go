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
	"travel-agent/internal/service/ai/tools"
	"travel-agent/pkg/utils"
)

var AIProviderEndpoint = "https://api.mistral.ai/v1/chat/completions"

const (
	model   = "mistral-large-latest"
	timeout = 30 * time.Second
)

type InferenceEngine[T models.TravelOutput, R models.TravelInput] struct {
	apiKey       string
	httpClient   *http.Client
	toolRegistry *tools.ToolRegistry
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
		toolRegistry: tools.NewToolRegistry(),
	}, nil
}

// RegisterTool adds a tool to the inference engine's registry
func (p *InferenceEngine[T, R]) RegisterTool(tool tools.Tool) error {
	if p.toolRegistry == nil {
		p.toolRegistry = tools.NewToolRegistry()
	}
	return p.toolRegistry.RegisterTool(tool)
}

type AIProviderRequest struct {
	Model          string                   `json:"model"`
	Messages       []AIProviderMsg          `json:"messages"`
	ResponseFormat ResponseFormat           `json:"response_format"`
	Tools          []map[string]interface{} `json:"tools,omitempty"`
	ToolChoice     interface{}              `json:"tool_choice,omitempty"`
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
			Role      string `json:"role"`
			Content   string `json:"content"`
			ToolCalls []struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls,omitempty"`
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

	// Add tools if available
	if p.toolRegistry != nil && len(p.toolRegistry.ListTools()) > 0 {
		aiReq.Tools = p.toolRegistry.ListMistralTools()
		// Set tool_choice to "auto" to let the model decide when to use tools
		aiReq.ToolChoice = "auto"
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

	// Check if there are tool calls to process
	if len(aiResp.Choices) > 0 && len(aiResp.Choices[0].Message.ToolCalls) > 0 {
		// Process tool calls
		response, err := p.processToolCalls(ctx, aiResp.Choices[0].Message.ToolCalls)
		if err != nil {
			return nil, fmt.Errorf("failed to process tool calls: %w", err)
		}

		// Decode the response after tool processing
		return decodingStrategy.DecodeResponse(response)
	}

	// Decode the response if no tool calls
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

// processToolCalls executes the tools called by the AI and returns the result
func (p *InferenceEngine[T, R]) processToolCalls(ctx context.Context, toolCalls []struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}) (string, error) {
	if p.toolRegistry == nil {
		return "", fmt.Errorf("tool registry is not initialized")
	}

	// Process each tool call
	var toolResults []map[string]interface{}
	for _, call := range toolCalls {
		// Get the tool from registry
		tool, exists := p.toolRegistry.GetTool(call.Function.Name)
		if !exists {
			return "", fmt.Errorf("tool '%s' not found in registry", call.Function.Name)
		}

		// Parse arguments
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
			return "", fmt.Errorf("failed to parse tool arguments: %w", err)
		}

		// Execute the tool
		result, err := tool.Execute(ctx, args)
		if err != nil {
			return "", fmt.Errorf("failed to execute tool '%s': %w", call.Function.Name, err)
		}

		// Add result to the list
		toolResults = append(toolResults, map[string]interface{}{
			"tool_call_id": call.ID,
			"name":         call.Function.Name,
			"result":       result,
		})
	}

	// Convert results to JSON
	resultsJSON, err := json.Marshal(toolResults)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tool results: %w", err)
	}

	return string(resultsJSON), nil
}
