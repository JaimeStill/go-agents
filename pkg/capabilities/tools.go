package capabilities

import (
	"encoding/json"
	"fmt"

	"github.com/JaimeStill/go-agents/pkg/protocols"
)

type FunctionDefinition struct {
	Type     string         `json:"type"`
	Function map[string]any `json:"function"`
}

type ToolsCapability struct {
	*StandardStreamingCapability
}

func NewToolsCapability(name string, options []CapabilityOption) *ToolsCapability {
	return &ToolsCapability{
		StandardStreamingCapability: NewStandardStreamingCapability(
			name,
			protocols.Tools,
			options,
		),
	}
}

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

func (c *ToolsCapability) CreateStreamingRequest(req *CapabilityRequest, model string) (*protocols.Request, error) {
	options, err := c.ProcessOptions(req.Options)
	if err != nil {
		return nil, err
	}

	tools, ok := options["tools"].([]FunctionDefinition)
	if !ok || len(tools) == 0 {
		return nil, fmt.Errorf("tools must be a non-empty array of FunctionDefinition")
	}

	options["model"] = model
	options["stream"] = true

	return &protocols.Request{
		Messages: req.Messages,
		Options:  options,
	}, nil
}

func (c *ToolsCapability) ParseResponse(data []byte) (any, error) {
	var response protocols.ChatResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
