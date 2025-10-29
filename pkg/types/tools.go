package types

import (
	"encoding/json"
	"fmt"

	"maps"
)

// ToolDefinition represents a provider-agnostic tool (function) definition.
// Providers transform this generic format to their specific API format
// (OpenAI, Anthropic, Google, etc.).
type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"` // JSON Schema
}

// ToolsRequest represents a tools (function calling) protocol request.
// Separates tool definitions (protocol input data) from model configuration options.
type ToolsRequest struct {
	Messages []Message
	Tools    []ToolDefinition
	Options  map[string]any
}

// GetProtocol returns the Tools protocol identifier.
func (r *ToolsRequest) GetProtocol() Protocol {
	return Tools
}

// GetHeaders returns the HTTP headers for a tools request.
func (r *ToolsRequest) GetHeaders() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

// Marshal converts the tools request to JSON using OpenAI's tool format.
// Tools are wrapped in {"type": "function", "function": {...}} structure
// which is the standard format for OpenAI, Azure, and Ollama providers.
//
// Output structure:
//
//	{
//	  "messages": [...],
//	  "tools": [
//	    {"type": "function", "function": {"name": "...", "description": "...", "parameters": {...}}}
//	  ],
//	  "temperature": 0.7,
//	  "max_tokens": 4096
//	}
//
// Design Decision: We use OpenAI's format as the default since it's used by
// the majority of providers (OpenAI, Azure, Ollama). Future providers that
// require different formats (Anthropic, Google) can transform from this
// standard representation in their PrepareRequest implementation.
func (r *ToolsRequest) Marshal() ([]byte, error) {
	combined := make(map[string]any)
	combined["messages"] = r.Messages

	// Transform tools to OpenAI format: {"type": "function", "function": {...}}
	openAITools := make([]map[string]any, len(r.Tools))
	for i, tool := range r.Tools {
		openAITools[i] = map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.Parameters,
			},
		}
	}
	combined["tools"] = openAITools

	maps.Copy(combined, r.Options)
	return json.Marshal(combined)
}

// ParseToolsResponse parses a tools response from JSON.
// Returns the parsed ToolsResponse or an error if parsing fails.
func ParseToolsResponse(body []byte) (*ToolsResponse, error) {
	var response ToolsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse tools response: %w", err)
	}
	return &response, nil
}

// ParseToolsStreamChunk parses a streaming tools chunk from JSON.
// Tools protocol uses the same streaming format as chat.
// Returns the parsed StreamingChunk or an error if parsing fails.
func ParseToolsStreamChunk(data []byte) (*StreamingChunk, error) {
	var chunk StreamingChunk
	if err := json.Unmarshal(data, &chunk); err != nil {
		return nil, fmt.Errorf("failed to parse tools streaming chunk: %w", err)
	}
	return &chunk, nil
}
