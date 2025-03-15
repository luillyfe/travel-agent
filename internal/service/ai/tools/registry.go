package tools

import "fmt"

// ToolRegistry manages available tools for the agent
type ToolRegistry struct {
	tools map[string]Tool
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

func (tr *ToolRegistry) RegisterTool(tool Tool) error {
	// Add tool to registry with validation logic
	if tool == nil {
		return fmt.Errorf("tool cannot be nil")
	}

	name := tool.Name()
	if _, exists := tr.tools[name]; exists {
		return fmt.Errorf("tool '%s' already registered", name)
	}

	// Validate parameter schema structure
	if tool.Parameters() == nil {
		return fmt.Errorf("tool '%s' must provide parameters schema", name)
	}
	if _, ok := tool.Parameters()["type"]; !ok {
		return fmt.Errorf("tool '%s' parameters schema must include 'type' field", name)
	}
	if _, ok := tool.Parameters()["properties"]; !ok {
		return fmt.Errorf("tool '%s' parameters schema must include 'properties' field", name)
	}
	for propName, propSchema := range tool.Parameters()["properties"].(map[string]interface{}) {
		if _, ok := propSchema.(map[string]interface{})["type"]; !ok {
			return fmt.Errorf("tool '%s' parameter '%s' must specify type", name, propName)
		}
	}

	tr.tools[name] = tool
	return nil
}

func (tr *ToolRegistry) GetTool(name string) (Tool, bool) {
	// Retrieve tool by name
	tool, ok := tr.tools[name]
	if ok {
		// Validate parameters schema exists
		if tool.Parameters() == nil {
			return nil, false
		}
	}
	return tool, ok
}

func (tr *ToolRegistry) ListTools() []Tool {
	// Return all registered tools
	tools := make([]Tool, 0, len(tr.tools))
	for _, tool := range tr.tools {
		tools = append(tools, tool)
	}
	return tools
}

func (tr *ToolRegistry) ListMistralTools() []map[string]interface{} {
	var mistralTools []map[string]interface{}
	for _, tool := range tr.tools {
		mistralTools = append(mistralTools, map[string]interface{}{
			"name":         tool.Name(),
			"description":  tool.Description(),
			"parameters":   tool.Parameters(),
			"requirements": tool.Requirements(),
		})
	}
	return mistralTools
}
