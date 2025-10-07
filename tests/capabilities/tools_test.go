package capabilities_test

import (
	"testing"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

func TestNewToolsCapability(t *testing.T) {
	options := []capabilities.CapabilityOption{
		{Option: "tools", Required: true, DefaultValue: nil},
		{Option: "tool_choice", Required: false, DefaultValue: "auto"},
	}

	cap := capabilities.NewToolsCapability("openai-tools", options)

	if cap.Name() != "openai-tools" {
		t.Errorf("got name %q, want %q", cap.Name(), "openai-tools")
	}

	if cap.Protocol() != protocols.Tools {
		t.Errorf("got protocol %q, want %q", cap.Protocol(), protocols.Tools)
	}

	if cap.SupportsStreaming() {
		t.Error("ToolsCapability should not support streaming")
	}
}

func TestToolsCapability_CreateRequest(t *testing.T) {
	options := []capabilities.CapabilityOption{
		{Option: "tools", Required: true, DefaultValue: nil},
		{Option: "tool_choice", Required: false, DefaultValue: "auto"},
		{Option: "temperature", Required: false, DefaultValue: 0.7},
	}

	cap := capabilities.NewToolsCapability("openai-tools", options)

	tools := []capabilities.FunctionDefinition{
		{
			Type: "function",
			Function: map[string]any{
				"name":        "get_weather",
				"description": "Get the current weather",
				"parameters": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The city and state",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Tools,
		Messages: []protocols.Message{
			protocols.NewMessage("user", "What's the weather in Boston?"),
		},
		Options: map[string]any{
			"tools":       tools,
			"tool_choice": "auto",
			"temperature": 0.8,
		},
	}

	protocolReq, err := cap.CreateRequest(req, "gpt-4")
	if err != nil {
		t.Fatalf("CreateRequest failed: %v", err)
	}

	if len(protocolReq.Messages) != 1 {
		t.Errorf("got %d messages, want 1", len(protocolReq.Messages))
	}

	if model, exists := protocolReq.Options["model"]; !exists {
		t.Error("model option missing")
	} else if model != "gpt-4" {
		t.Errorf("got model %q, want %q", model, "gpt-4")
	}

	if _, exists := protocolReq.Options["tools"]; !exists {
		t.Error("tools option missing")
	}

	if toolChoice, exists := protocolReq.Options["tool_choice"]; !exists {
		t.Error("tool_choice option missing")
	} else if toolChoice != "auto" {
		t.Errorf("got tool_choice %v, want %q", toolChoice, "auto")
	}
}

func TestToolsCapability_CreateRequest_EmptyTools(t *testing.T) {
	options := []capabilities.CapabilityOption{
		{Option: "tools", Required: true, DefaultValue: nil},
	}

	cap := capabilities.NewToolsCapability("openai-tools", options)

	tests := []struct {
		name    string
		options map[string]any
	}{
		{
			name:    "missing tools",
			options: map[string]any{},
		},
		{
			name:    "empty array",
			options: map[string]any{"tools": []capabilities.FunctionDefinition{}},
		},
		{
			name:    "wrong type",
			options: map[string]any{"tools": "not-an-array"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &capabilities.CapabilityRequest{
				Protocol: protocols.Tools,
				Messages: []protocols.Message{
					protocols.NewMessage("user", "Test"),
				},
				Options: tt.options,
			}

			_, err := cap.CreateRequest(req, "gpt-4")
			if err == nil {
				t.Error("expected error for invalid tools, got nil")
			}
		})
	}
}

func TestToolsCapability_ParseResponse(t *testing.T) {
	cap := capabilities.NewToolsCapability("openai-tools", nil)

	responseData := []byte(`{
		"id": "chatcmpl-123",
		"model": "gpt-4",
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": "",
				"tool_calls": [{
					"id": "call_abc123",
					"type": "function",
					"function": {
						"name": "get_weather",
						"arguments": "{\"location\": \"Boston, MA\"}"
					}
				}]
			},
			"finish_reason": "tool_calls"
		}],
		"usage": {
			"prompt_tokens": 15,
			"completion_tokens": 25,
			"total_tokens": 40
		}
	}`)

	result, err := cap.ParseResponse(responseData)
	if err != nil {
		t.Fatalf("ParseResponse failed: %v", err)
	}

	response, ok := result.(*protocols.ToolsResponse)
	if !ok {
		t.Fatal("result is not a ToolsResponse")
	}

	if response.ID != "chatcmpl-123" {
		t.Errorf("got ID %q, want %q", response.ID, "chatcmpl-123")
	}

	if response.Model != "gpt-4" {
		t.Errorf("got model %q, want %q", response.Model, "gpt-4")
	}

	if len(response.Choices) != 1 {
		t.Fatalf("got %d choices, want 1", len(response.Choices))
	}

	toolCalls := response.Choices[0].Message.ToolCalls
	if len(toolCalls) != 1 {
		t.Fatalf("got %d tool calls, want 1", len(toolCalls))
	}

	if toolCalls[0].ID != "call_abc123" {
		t.Errorf("got tool call ID %q, want %q", toolCalls[0].ID, "call_abc123")
	}

	if toolCalls[0].Function.Name != "get_weather" {
		t.Errorf("got function name %q, want %q", toolCalls[0].Function.Name, "get_weather")
	}
}

func TestToolsCapability_ParseResponse_InvalidJSON(t *testing.T) {
	cap := capabilities.NewToolsCapability("openai-tools", nil)

	responseData := []byte(`{invalid json}`)

	_, err := cap.ParseResponse(responseData)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}
