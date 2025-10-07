package capabilities

import (
	"encoding/json"
	"fmt"

	"github.com/JaimeStill/go-agents/pkg/protocols"
)

// FunctionDefinition describes a function that can be called by the model.
// Contains the function type and its specification including name, description, and parameters.
type FunctionDefinition struct {
	Type     string         `json:"type"`
	Function map[string]any `json:"function"`
}

// ToolsCapability implements the tools (function calling) protocol.
// Allows models to request function executions by providing structured tool call information.
// Does not support streaming as tool responses require complete function definitions.
type ToolsCapability struct {
	*StandardCapability
}

// NewToolsCapability creates a new ToolsCapability with the specified options.
// Required options typically include "tools" (array of FunctionDefinition).
// Optional options include "tool_choice" for controlling which tools to use.
func NewToolsCapability(name string, options []CapabilityOption) *ToolsCapability {
	return &ToolsCapability{
		StandardCapability: NewStandardCapability(
			name,
			protocols.Tools,
			options,
		),
	}
}

// CreateRequest creates a protocol request for function calling.
// Validates that tools are provided as a non-empty array of FunctionDefinition.
func (c *ToolsCapability) CreateRequest(req *CapabilityRequest, model string) (*protocols.Request, error) {
	options, err := c.ProcessOptions(req.Options)
	if err != nil {
		return nil, err
	}

	tools, ok := options["tools"].([]FunctionDefinition)
	if !ok || len(tools) == 0 {
		return nil, fmt.Errorf("tools must be a non-empty array of FunctionDefinition")
	}

	options["model"] = model

	return &protocols.Request{
		Messages: req.Messages,
		Options:  options,
	}, nil
}

// ParseResponse parses a tools response containing function call requests.
// Returns a ToolsResponse with tool calls that need to be executed.
func (c *ToolsCapability) ParseResponse(data []byte) (any, error) {
	var response protocols.ToolsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
