package models_test

import (
	"testing"

	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/models"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

func TestNew_Success(t *testing.T) {
	cfg := &config.ModelConfig{
		Name: "test-model",
		Capabilities: map[string]config.CapabilityConfig{
			"chat": {
				Format: "openai-chat",
				Options: map[string]any{
					"temperature": 0.7,
				},
			},
		},
	}

	model, err := models.New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if model == nil {
		t.Fatal("New() returned nil model")
	}

	if model.Name() != "test-model" {
		t.Errorf("got name %q, want %q", model.Name(), "test-model")
	}
}

func TestNew_InvalidProtocol(t *testing.T) {
	cfg := &config.ModelConfig{
		Name: "test-model",
		Capabilities: map[string]config.CapabilityConfig{
			"invalid-protocol": {
				Format: "openai-chat",
			},
		},
	}

	_, err := models.New(cfg)
	if err == nil {
		t.Error("expected error for invalid protocol, got nil")
	}
}

func TestNew_InvalidCapabilityFormat(t *testing.T) {
	cfg := &config.ModelConfig{
		Name: "test-model",
		Capabilities: map[string]config.CapabilityConfig{
			"chat": {
				Format: "non-existent-format",
			},
		},
	}

	_, err := models.New(cfg)
	if err == nil {
		t.Error("expected error for invalid capability format, got nil")
	}
}

func TestModel_Name(t *testing.T) {
	cfg := &config.ModelConfig{
		Name: "my-model",
		Capabilities: map[string]config.CapabilityConfig{
			"chat": {
				Format: "openai-chat",
			},
		},
	}

	model, err := models.New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if model.Name() != "my-model" {
		t.Errorf("got name %q, want %q", model.Name(), "my-model")
	}
}

func TestModel_SupportsProtocol(t *testing.T) {
	cfg := &config.ModelConfig{
		Name: "test-model",
		Capabilities: map[string]config.CapabilityConfig{
			"chat": {
				Format: "openai-chat",
			},
			"vision": {
				Format: "openai-vision",
			},
		},
	}

	model, err := models.New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	tests := []struct {
		name     string
		protocol protocols.Protocol
		expected bool
	}{
		{
			name:     "chat supported",
			protocol: protocols.Chat,
			expected: true,
		},
		{
			name:     "vision supported",
			protocol: protocols.Vision,
			expected: true,
		},
		{
			name:     "tools not supported",
			protocol: protocols.Tools,
			expected: false,
		},
		{
			name:     "embeddings not supported",
			protocol: protocols.Embeddings,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.SupportsProtocol(tt.protocol)
			if result != tt.expected {
				t.Errorf("SupportsProtocol(%s) = %v, want %v", tt.protocol, result, tt.expected)
			}
		})
	}
}

func TestModel_GetCapability_Supported(t *testing.T) {
	cfg := &config.ModelConfig{
		Name: "test-model",
		Capabilities: map[string]config.CapabilityConfig{
			"chat": {
				Format: "openai-chat",
			},
		},
	}

	model, err := models.New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	cap, err := model.GetCapability(protocols.Chat)
	if err != nil {
		t.Fatalf("GetCapability failed: %v", err)
	}

	if cap == nil {
		t.Fatal("GetCapability returned nil capability")
	}

	if cap.Name() != "openai-chat" {
		t.Errorf("got capability name %q, want %q", cap.Name(), "openai-chat")
	}

	if cap.Protocol() != protocols.Chat {
		t.Errorf("got protocol %q, want %q", cap.Protocol(), protocols.Chat)
	}
}

func TestModel_GetCapability_Unsupported(t *testing.T) {
	cfg := &config.ModelConfig{
		Name: "test-model",
		Capabilities: map[string]config.CapabilityConfig{
			"chat": {
				Format: "openai-chat",
			},
		},
	}

	model, err := models.New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	_, err = model.GetCapability(protocols.Tools)
	if err == nil {
		t.Error("expected error for unsupported protocol, got nil")
	}
}

func TestModel_GetProtocolOptions_Supported(t *testing.T) {
	cfg := &config.ModelConfig{
		Name: "test-model",
		Capabilities: map[string]config.CapabilityConfig{
			"chat": {
				Format: "openai-chat",
				Options: map[string]any{
					"temperature": 0.8,
					"max_tokens":  2000,
				},
			},
		},
	}

	model, err := models.New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	options := model.GetProtocolOptions(protocols.Chat)

	if len(options) != 2 {
		t.Fatalf("got %d options, want 2", len(options))
	}

	if temp, exists := options["temperature"]; !exists {
		t.Error("temperature option missing")
	} else if temp != 0.8 {
		t.Errorf("got temperature %v, want 0.8", temp)
	}

	if tokens, exists := options["max_tokens"]; !exists {
		t.Error("max_tokens option missing")
	} else if tokens != 2000 {
		t.Errorf("got max_tokens %v, want 2000", tokens)
	}
}

func TestModel_GetProtocolOptions_Unsupported(t *testing.T) {
	cfg := &config.ModelConfig{
		Name: "test-model",
		Capabilities: map[string]config.CapabilityConfig{
			"chat": {
				Format: "openai-chat",
			},
		},
	}

	model, err := models.New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	options := model.GetProtocolOptions(protocols.Tools)

	if options == nil {
		t.Error("expected empty map, got nil")
	}

	if len(options) != 0 {
		t.Errorf("expected empty map, got %d options", len(options))
	}
}

func TestModel_UpdateProtocolOptions_Success(t *testing.T) {
	cfg := &config.ModelConfig{
		Name: "test-model",
		Capabilities: map[string]config.CapabilityConfig{
			"chat": {
				Format: "openai-chat",
				Options: map[string]any{
					"temperature": 0.7,
				},
			},
		},
	}

	model, err := models.New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	newOptions := map[string]any{
		"temperature": 0.9,
		"max_tokens":  4096,
	}

	err = model.UpdateProtocolOptions(protocols.Chat, newOptions)
	if err != nil {
		t.Fatalf("UpdateProtocolOptions failed: %v", err)
	}

	options := model.GetProtocolOptions(protocols.Chat)

	if temp, exists := options["temperature"]; !exists {
		t.Error("temperature should be updated")
	} else if temp != 0.9 {
		t.Errorf("got temperature %v, want 0.9", temp)
	}

	if tokens, exists := options["max_tokens"]; !exists {
		t.Error("max_tokens should be added")
	} else if tokens != 4096 {
		t.Errorf("got max_tokens %v, want 4096", tokens)
	}
}

func TestModel_UpdateProtocolOptions_ValidationError(t *testing.T) {
	cfg := &config.ModelConfig{
		Name: "test-model",
		Capabilities: map[string]config.CapabilityConfig{
			"chat": {
				Format: "openai-chat",
			},
		},
	}

	model, err := models.New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Try to update with an unsupported option
	// Note: openai-chat capability validates against its allowed options
	invalidOptions := map[string]any{
		"unsupported_option": "value",
	}

	err = model.UpdateProtocolOptions(protocols.Chat, invalidOptions)
	if err == nil {
		t.Error("expected validation error, got nil")
	}
}

func TestModel_UpdateProtocolOptions_UnsupportedProtocol(t *testing.T) {
	cfg := &config.ModelConfig{
		Name: "test-model",
		Capabilities: map[string]config.CapabilityConfig{
			"chat": {
				Format: "openai-chat",
			},
		},
	}

	model, err := models.New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	err = model.UpdateProtocolOptions(protocols.Tools, map[string]any{"temperature": 0.8})
	if err == nil {
		t.Error("expected error for unsupported protocol, got nil")
	}
}

func TestModel_MergeRequestOptions_Supported(t *testing.T) {
	cfg := &config.ModelConfig{
		Name: "test-model",
		Capabilities: map[string]config.CapabilityConfig{
			"chat": {
				Format: "openai-chat",
				Options: map[string]any{
					"temperature": 0.7,
					"max_tokens":  4096,
				},
			},
		},
	}

	model, err := models.New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	requestOptions := map[string]any{
		"temperature": 0.9,
		"top_p":       0.95,
	}

	merged := model.MergeRequestOptions(protocols.Chat, requestOptions)

	// Request options should override handler options
	if temp, exists := merged["temperature"]; !exists {
		t.Error("temperature should be in merged")
	} else if temp != 0.9 {
		t.Errorf("got temperature %v, want 0.9 (request should override)", temp)
	}

	// Handler options should be included
	if tokens, exists := merged["max_tokens"]; !exists {
		t.Error("max_tokens should be in merged from handler")
	} else if tokens != 4096 {
		t.Errorf("got max_tokens %v, want 4096", tokens)
	}

	// New options from request should be included
	if topP, exists := merged["top_p"]; !exists {
		t.Error("top_p should be in merged from request")
	} else if topP != 0.95 {
		t.Errorf("got top_p %v, want 0.95", topP)
	}
}

func TestModel_MergeRequestOptions_Unsupported(t *testing.T) {
	cfg := &config.ModelConfig{
		Name: "test-model",
		Capabilities: map[string]config.CapabilityConfig{
			"chat": {
				Format: "openai-chat",
			},
		},
	}

	model, err := models.New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	requestOptions := map[string]any{
		"temperature": 0.9,
	}

	// For unsupported protocol, should return request options unchanged
	merged := model.MergeRequestOptions(protocols.Tools, requestOptions)

	if len(merged) != 1 {
		t.Errorf("got %d merged options, want 1", len(merged))
	}

	if temp, exists := merged["temperature"]; !exists {
		t.Error("temperature should be in result")
	} else if temp != 0.9 {
		t.Errorf("got temperature %v, want 0.9", temp)
	}
}

func TestModel_MultipleProtocols(t *testing.T) {
	cfg := &config.ModelConfig{
		Name: "multi-protocol-model",
		Capabilities: map[string]config.CapabilityConfig{
			"chat": {
				Format: "openai-chat",
				Options: map[string]any{
					"temperature": 0.7,
				},
			},
			"vision": {
				Format: "openai-vision",
				Options: map[string]any{
					"images": []any{"test.jpg"},
					"detail": "high",
				},
			},
			"embeddings": {
				Format: "openai-embeddings",
				Options: map[string]any{
					"input": "test text",
				},
			},
		},
	}

	model, err := models.New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Verify all protocols are supported
	if !model.SupportsProtocol(protocols.Chat) {
		t.Error("chat should be supported")
	}

	if !model.SupportsProtocol(protocols.Vision) {
		t.Error("vision should be supported")
	}

	if !model.SupportsProtocol(protocols.Embeddings) {
		t.Error("embeddings should be supported")
	}

	if model.SupportsProtocol(protocols.Tools) {
		t.Error("tools should not be supported")
	}

	// Verify each protocol has correct options
	chatOpts := model.GetProtocolOptions(protocols.Chat)
	if chatOpts["temperature"] != 0.7 {
		t.Errorf("chat temperature = %v, want 0.7", chatOpts["temperature"])
	}

	visionOpts := model.GetProtocolOptions(protocols.Vision)
	if visionOpts["detail"] != "high" {
		t.Errorf("vision detail = %v, want 'high'", visionOpts["detail"])
	}

	embeddingsOpts := model.GetProtocolOptions(protocols.Embeddings)
	if embeddingsOpts["input"] != "test text" {
		t.Errorf("embeddings input = %v, want 'test text'", embeddingsOpts["input"])
	}
}
