package models_test

import (
	"testing"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/models"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

func TestNewProtocolHandler(t *testing.T) {
	cap := capabilities.NewChatCapability("test-chat", []capabilities.CapabilityOption{
		{Option: "temperature", Required: false, DefaultValue: 0.7},
	})

	options := map[string]any{
		"temperature": 0.8,
		"max_tokens":  2000,
	}

	handler := models.NewProtocolHandler(cap, options)

	if handler == nil {
		t.Fatal("NewProtocolHandler returned nil")
	}

	if handler.Capability() == nil {
		t.Error("handler capability is nil")
	}

	if len(handler.Options()) != 2 {
		t.Errorf("got %d options, want 2", len(handler.Options()))
	}
}

func TestProtocolHandler_Capability(t *testing.T) {
	cap := capabilities.NewChatCapability("test-chat", nil)
	handler := models.NewProtocolHandler(cap, nil)

	result := handler.Capability()

	if result == nil {
		t.Fatal("Capability() returned nil")
	}

	if result.Name() != "test-chat" {
		t.Errorf("got capability name %q, want %q", result.Name(), "test-chat")
	}

	if result.Protocol() != protocols.Chat {
		t.Errorf("got protocol %q, want %q", result.Protocol(), protocols.Chat)
	}
}

func TestProtocolHandler_Options(t *testing.T) {
	cap := capabilities.NewChatCapability("test-chat", nil)

	options := map[string]any{
		"temperature": 0.7,
		"max_tokens":  4096,
	}

	handler := models.NewProtocolHandler(cap, options)
	result := handler.Options()

	if len(result) != 2 {
		t.Fatalf("got %d options, want 2", len(result))
	}

	if temp, exists := result["temperature"]; !exists {
		t.Error("temperature option missing")
	} else if temp != 0.7 {
		t.Errorf("got temperature %v, want 0.7", temp)
	}

	if tokens, exists := result["max_tokens"]; !exists {
		t.Error("max_tokens option missing")
	} else if tokens != 4096 {
		t.Errorf("got max_tokens %v, want 4096", tokens)
	}
}

func TestProtocolHandler_Options_ConstructorClones(t *testing.T) {
	cap := capabilities.NewChatCapability("test-chat", nil)

	options := map[string]any{
		"temperature": 0.7,
	}

	handler := models.NewProtocolHandler(cap, options)

	// Modify the original options map
	options["temperature"] = 0.9
	options["new_option"] = "value"

	// Verify that handler's options were cloned and are unchanged
	result := handler.Options()

	if result["temperature"] != 0.7 {
		t.Errorf("handler options were affected by original: got temperature %v, want 0.7", result["temperature"])
	}

	if _, exists := result["new_option"]; exists {
		t.Error("new_option should not exist in handler options")
	}
}

func TestProtocolHandler_UpdateOptions(t *testing.T) {
	cap := capabilities.NewChatCapability("test-chat", nil)

	initialOptions := map[string]any{
		"temperature": 0.7,
		"max_tokens":  4096,
	}

	handler := models.NewProtocolHandler(cap, initialOptions)

	updateOptions := map[string]any{
		"temperature": 0.9,
		"top_p":       0.95,
	}

	handler.UpdateOptions(updateOptions)

	result := handler.Options()

	if temp, exists := result["temperature"]; !exists {
		t.Error("temperature option missing")
	} else if temp != 0.9 {
		t.Errorf("got temperature %v, want 0.9", temp)
	}

	if tokens, exists := result["max_tokens"]; !exists {
		t.Error("max_tokens should still exist")
	} else if tokens != 4096 {
		t.Errorf("got max_tokens %v, want 4096", tokens)
	}

	if topP, exists := result["top_p"]; !exists {
		t.Error("top_p should be added")
	} else if topP != 0.95 {
		t.Errorf("got top_p %v, want 0.95", topP)
	}
}

func TestProtocolHandler_MergeOptions(t *testing.T) {
	cap := capabilities.NewChatCapability("test-chat", nil)

	handlerOptions := map[string]any{
		"temperature": 0.7,
		"max_tokens":  4096,
	}

	handler := models.NewProtocolHandler(cap, handlerOptions)

	requestOptions := map[string]any{
		"temperature": 0.9,
		"top_p":       0.95,
	}

	merged := handler.MergeOptions(requestOptions)

	// Check merged result
	if temp, exists := merged["temperature"]; !exists {
		t.Error("temperature option missing from merged")
	} else if temp != 0.9 {
		t.Errorf("got merged temperature %v, want 0.9 (request should override)", temp)
	}

	if tokens, exists := merged["max_tokens"]; !exists {
		t.Error("max_tokens should be in merged")
	} else if tokens != 4096 {
		t.Errorf("got merged max_tokens %v, want 4096", tokens)
	}

	if topP, exists := merged["top_p"]; !exists {
		t.Error("top_p should be in merged")
	} else if topP != 0.95 {
		t.Errorf("got merged top_p %v, want 0.95", topP)
	}

	// Verify handler options are unchanged
	handlerOpts := handler.Options()
	if temp, exists := handlerOpts["temperature"]; !exists {
		t.Error("temperature should still exist in handler")
	} else if temp != 0.7 {
		t.Errorf("handler temperature was modified: got %v, want 0.7", temp)
	}

	if _, exists := handlerOpts["top_p"]; exists {
		t.Error("top_p should not be added to handler options")
	}

	// Verify request options are unchanged
	if temp, exists := requestOptions["temperature"]; !exists {
		t.Error("temperature should still exist in request")
	} else if temp != 0.9 {
		t.Errorf("request temperature was modified: got %v, want 0.9", temp)
	}

	if _, exists := requestOptions["max_tokens"]; exists {
		t.Error("max_tokens should not be added to request options")
	}
}

func TestProtocolHandler_MergeOptions_EmptyRequest(t *testing.T) {
	cap := capabilities.NewChatCapability("test-chat", nil)

	handlerOptions := map[string]any{
		"temperature": 0.7,
	}

	handler := models.NewProtocolHandler(cap, handlerOptions)

	merged := handler.MergeOptions(map[string]any{})

	if len(merged) != 1 {
		t.Errorf("got %d merged options, want 1", len(merged))
	}

	if temp, exists := merged["temperature"]; !exists {
		t.Error("temperature should be in merged result")
	} else if temp != 0.7 {
		t.Errorf("got temperature %v, want 0.7", temp)
	}
}

func TestProtocolHandler_MergeOptions_EmptyHandlerOptions(t *testing.T) {
	cap := capabilities.NewChatCapability("test-chat", nil)
	handler := models.NewProtocolHandler(cap, map[string]any{})

	requestOptions := map[string]any{
		"temperature": 0.9,
	}

	merged := handler.MergeOptions(requestOptions)

	if len(merged) != 1 {
		t.Errorf("got %d merged options, want 1", len(merged))
	}

	if temp, exists := merged["temperature"]; !exists {
		t.Error("temperature should be in merged result")
	} else if temp != 0.9 {
		t.Errorf("got temperature %v, want 0.9", temp)
	}
}
