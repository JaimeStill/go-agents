package capabilities_test

import (
	"testing"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

func TestNewChatCapability(t *testing.T) {
	options := []capabilities.CapabilityOption{
		{Option: "temperature", Required: false, DefaultValue: 0.7},
		{Option: "max_tokens", Required: false, DefaultValue: 4096},
	}

	cap := capabilities.NewChatCapability("openai-chat", options)

	if cap.Name() != "openai-chat" {
		t.Errorf("got name %q, want %q", cap.Name(), "openai-chat")
	}

	if cap.Protocol() != protocols.Chat {
		t.Errorf("got protocol %q, want %q", cap.Protocol(), protocols.Chat)
	}

	if !cap.SupportsStreaming() {
		t.Error("ChatCapability should support streaming")
	}
}

func TestChatCapability_CreateRequest(t *testing.T) {
	options := []capabilities.CapabilityOption{
		{Option: "temperature", Required: false, DefaultValue: 0.7},
		{Option: "max_tokens", Required: false, DefaultValue: 4096},
	}

	cap := capabilities.NewChatCapability("openai-chat", options)

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Chat,
		Messages: []protocols.Message{
			protocols.NewMessage("user", "Hello"),
		},
		Options: map[string]any{
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

	if temp, exists := protocolReq.Options["temperature"]; !exists {
		t.Error("temperature option missing")
	} else if temp != 0.8 {
		t.Errorf("got temperature %v, want 0.8", temp)
	}

	if model, exists := protocolReq.Options["model"]; !exists {
		t.Error("model option missing")
	} else if model != "gpt-4" {
		t.Errorf("got model %q, want %q", model, "gpt-4")
	}

	if maxTokens, exists := protocolReq.Options["max_tokens"]; !exists {
		t.Error("max_tokens default missing")
	} else if maxTokens != 4096 {
		t.Errorf("got max_tokens %v, want 4096", maxTokens)
	}
}

func TestChatCapability_CreateStreamingRequest(t *testing.T) {
	options := []capabilities.CapabilityOption{
		{Option: "temperature", Required: false, DefaultValue: 0.7},
	}

	cap := capabilities.NewChatCapability("openai-chat", options)

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Chat,
		Messages: []protocols.Message{
			protocols.NewMessage("user", "Hello"),
		},
		Options: map[string]any{
			"temperature": 0.8,
		},
	}

	protocolReq, err := cap.CreateStreamingRequest(req, "gpt-4")
	if err != nil {
		t.Fatalf("CreateStreamingRequest failed: %v", err)
	}

	if stream, exists := protocolReq.Options["stream"]; !exists {
		t.Error("stream option missing")
	} else if stream != true {
		t.Errorf("got stream %v, want true", stream)
	}

	if model, exists := protocolReq.Options["model"]; !exists {
		t.Error("model option missing")
	} else if model != "gpt-4" {
		t.Errorf("got model %q, want %q", model, "gpt-4")
	}
}

func TestChatCapability_ParseResponse(t *testing.T) {
	cap := capabilities.NewChatCapability("openai-chat", nil)

	responseData := []byte(`{
		"id": "chatcmpl-123",
		"model": "gpt-4",
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": "Hello, how can I help you?"
			},
			"finish_reason": "stop"
		}],
		"usage": {
			"prompt_tokens": 10,
			"completion_tokens": 20,
			"total_tokens": 30
		}
	}`)

	result, err := cap.ParseResponse(responseData)
	if err != nil {
		t.Fatalf("ParseResponse failed: %v", err)
	}

	response, ok := result.(*protocols.ChatResponse)
	if !ok {
		t.Fatal("result is not a ChatResponse")
	}

	if response.ID != "chatcmpl-123" {
		t.Errorf("got ID %q, want %q", response.ID, "chatcmpl-123")
	}

	if response.Model != "gpt-4" {
		t.Errorf("got model %q, want %q", response.Model, "gpt-4")
	}

	if content := response.Content(); content != "Hello, how can I help you?" {
		t.Errorf("got content %q, want %q", content, "Hello, how can I help you?")
	}
}

func TestChatCapability_ParseResponse_InvalidJSON(t *testing.T) {
	cap := capabilities.NewChatCapability("openai-chat", nil)

	responseData := []byte(`{invalid json}`)

	_, err := cap.ParseResponse(responseData)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestChatCapability_CreateRequest_ValidationError(t *testing.T) {
	options := []capabilities.CapabilityOption{
		{Option: "temperature", Required: true, DefaultValue: nil},
	}

	cap := capabilities.NewChatCapability("openai-chat", options)

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Chat,
		Messages: []protocols.Message{
			protocols.NewMessage("user", "Hello"),
		},
		Options: map[string]any{
			// Missing required temperature
		},
	}

	_, err := cap.CreateRequest(req, "gpt-4")
	if err == nil {
		t.Error("expected validation error, got nil")
	}
}
