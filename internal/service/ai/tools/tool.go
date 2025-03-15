package tools

import "context"

// Tool represents an external capability that can be called by the agent
type Tool interface {
	// Name returns the unique identifier for this tool
	Name() string

	// Description provides information about the tool's purpose and usage
	Description() string

	// Parameters returns the JSON Schema for the tool's parameters
	Parameters() map[string]interface{}

	// Requirements returns any specific requirements for using this tool
	Requirements() map[string]interface{}

	// Execute runs the tool with the provided parameters and returns results
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
}
