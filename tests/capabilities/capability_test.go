package capabilities_test

import (
	"testing"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

func TestStandardCapability_Name(t *testing.T) {
	cap := capabilities.NewStandardCapability("test-cap", protocols.Chat, nil)
	if cap.Name() != "test-cap" {
		t.Errorf("got name %q, want %q", cap.Name(), "test-cap")
	}
}

func TestStandardCapability_Protocol(t *testing.T) {
	cap := capabilities.NewStandardCapability("test-cap", protocols.Vision, nil)
	if cap.Protocol() != protocols.Vision {
		t.Errorf("got protocol %q, want %q", cap.Protocol(), protocols.Vision)
	}
}

func TestStandardCapability_Options(t *testing.T) {
	options := []capabilities.CapabilityOption{
		{Option: "temperature", Required: false, DefaultValue: 0.7},
		{Option: "max_tokens", Required: true, DefaultValue: nil},
	}

	cap := capabilities.NewStandardCapability("test-cap", protocols.Chat, options)
	result := cap.Options()

	if len(result) != 2 {
		t.Fatalf("got %d options, want 2", len(result))
	}

	if result[0].Option != "temperature" {
		t.Errorf("got option[0].Option %q, want %q", result[0].Option, "temperature")
	}

	if result[1].Option != "max_tokens" {
		t.Errorf("got option[1].Option %q, want %q", result[1].Option, "max_tokens")
	}
}

func TestStandardCapability_SupportsStreaming(t *testing.T) {
	cap := capabilities.NewStandardCapability("test-cap", protocols.Chat, nil)
	if cap.SupportsStreaming() {
		t.Error("StandardCapability should not support streaming")
	}
}

func TestStandardCapability_ValidateOptions_Valid(t *testing.T) {
	options := []capabilities.CapabilityOption{
		{Option: "temperature", Required: false, DefaultValue: 0.7},
		{Option: "max_tokens", Required: false, DefaultValue: 4096},
	}

	cap := capabilities.NewStandardCapability("test-cap", protocols.Chat, options)

	providedOptions := map[string]any{
		"temperature": 0.8,
		"max_tokens":  2000,
	}

	if err := cap.ValidateOptions(providedOptions); err != nil {
		t.Errorf("ValidateOptions failed: %v", err)
	}
}

func TestStandardCapability_ValidateOptions_UnsupportedOption(t *testing.T) {
	options := []capabilities.CapabilityOption{
		{Option: "temperature", Required: false, DefaultValue: 0.7},
	}

	cap := capabilities.NewStandardCapability("test-cap", protocols.Chat, options)

	providedOptions := map[string]any{
		"temperature":    0.8,
		"unsupported_opt": "value",
	}

	if err := cap.ValidateOptions(providedOptions); err == nil {
		t.Error("expected error for unsupported option, got nil")
	}
}

func TestStandardCapability_ValidateOptions_RequiredMissing(t *testing.T) {
	options := []capabilities.CapabilityOption{
		{Option: "temperature", Required: false, DefaultValue: 0.7},
		{Option: "max_tokens", Required: true, DefaultValue: nil},
	}

	cap := capabilities.NewStandardCapability("test-cap", protocols.Chat, options)

	providedOptions := map[string]any{
		"temperature": 0.8,
	}

	if err := cap.ValidateOptions(providedOptions); err == nil {
		t.Error("expected error for missing required option, got nil")
	}
}

func TestStandardCapability_ValidateOptions_RequiredProvided(t *testing.T) {
	options := []capabilities.CapabilityOption{
		{Option: "temperature", Required: false, DefaultValue: 0.7},
		{Option: "max_tokens", Required: true, DefaultValue: nil},
	}

	cap := capabilities.NewStandardCapability("test-cap", protocols.Chat, options)

	providedOptions := map[string]any{
		"temperature": 0.8,
		"max_tokens":  2000,
	}

	if err := cap.ValidateOptions(providedOptions); err != nil {
		t.Errorf("ValidateOptions failed: %v", err)
	}
}

func TestStandardCapability_ProcessOptions_WithDefaults(t *testing.T) {
	options := []capabilities.CapabilityOption{
		{Option: "temperature", Required: false, DefaultValue: 0.7},
		{Option: "max_tokens", Required: false, DefaultValue: 4096},
	}

	cap := capabilities.NewStandardCapability("test-cap", protocols.Chat, options)

	providedOptions := map[string]any{
		"temperature": 0.8,
	}

	result, err := cap.ProcessOptions(providedOptions)
	if err != nil {
		t.Fatalf("ProcessOptions failed: %v", err)
	}

	if temp, ok := result["temperature"].(float64); !ok || temp != 0.8 {
		t.Errorf("got temperature %v, want 0.8", result["temperature"])
	}

	if tokens, ok := result["max_tokens"].(int); !ok || tokens != 4096 {
		t.Errorf("got max_tokens %v, want 4096", result["max_tokens"])
	}
}

func TestStandardCapability_ProcessOptions_WithoutDefaults(t *testing.T) {
	options := []capabilities.CapabilityOption{
		{Option: "temperature", Required: false, DefaultValue: 0.7},
		{Option: "top_p", Required: false, DefaultValue: nil},
	}

	cap := capabilities.NewStandardCapability("test-cap", protocols.Chat, options)

	providedOptions := map[string]any{
		"temperature": 0.8,
	}

	result, err := cap.ProcessOptions(providedOptions)
	if err != nil {
		t.Fatalf("ProcessOptions failed: %v", err)
	}

	if temp, ok := result["temperature"].(float64); !ok || temp != 0.8 {
		t.Errorf("got temperature %v, want 0.8", result["temperature"])
	}

	if _, exists := result["top_p"]; exists {
		t.Error("top_p should not be in result when not provided and no default")
	}
}

func TestStandardCapability_ProcessOptions_RequiredOption(t *testing.T) {
	options := []capabilities.CapabilityOption{
		{Option: "max_tokens", Required: true, DefaultValue: nil},
	}

	cap := capabilities.NewStandardCapability("test-cap", protocols.Chat, options)

	providedOptions := map[string]any{
		"max_tokens": 2000,
	}

	result, err := cap.ProcessOptions(providedOptions)
	if err != nil {
		t.Fatalf("ProcessOptions failed: %v", err)
	}

	if tokens, ok := result["max_tokens"].(int); !ok || tokens != 2000 {
		t.Errorf("got max_tokens %v, want 2000", result["max_tokens"])
	}
}

func TestStandardStreamingCapability_SupportsStreaming(t *testing.T) {
	cap := capabilities.NewStandardStreamingCapability("test-cap", protocols.Chat, nil)
	if !cap.SupportsStreaming() {
		t.Error("StandardStreamingCapability should support streaming")
	}
}

func TestStandardStreamingCapability_IsStreamComplete_WithDone(t *testing.T) {
	cap := capabilities.NewStandardStreamingCapability("test-cap", protocols.Chat, nil)

	tests := []struct {
		name     string
		data     string
		expected bool
	}{
		{
			name:     "contains [DONE]",
			data:     "data: [DONE]",
			expected: true,
		},
		{
			name:     "[DONE] alone",
			data:     "[DONE]",
			expected: true,
		},
		{
			name:     "normal data",
			data:     `{"choices":[{"delta":{"content":"hello"}}]}`,
			expected: false,
		},
		{
			name:     "empty string",
			data:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cap.IsStreamComplete(tt.data)
			if result != tt.expected {
				t.Errorf("IsStreamComplete(%q) = %v, want %v", tt.data, result, tt.expected)
			}
		})
	}
}

func TestStandardStreamingCapability_ParseStreamingChunk_SSEFormat(t *testing.T) {
	cap := capabilities.NewStandardStreamingCapability("test-cap", protocols.Chat, nil)

	data := []byte(`data: {"model":"gpt-4","choices":[{"delta":{"content":"Hello"}}]}`)

	chunk, err := cap.ParseStreamingChunk(data)
	if err != nil {
		t.Fatalf("ParseStreamingChunk failed: %v", err)
	}

	if chunk.Model != "gpt-4" {
		t.Errorf("got model %q, want %q", chunk.Model, "gpt-4")
	}

	if content := chunk.Content(); content != "Hello" {
		t.Errorf("got content %q, want %q", content, "Hello")
	}
}

func TestStandardStreamingCapability_ParseStreamingChunk_PlainJSON(t *testing.T) {
	cap := capabilities.NewStandardStreamingCapability("test-cap", protocols.Chat, nil)

	data := []byte(`{"model":"gpt-4","choices":[{"delta":{"content":"World"}}]}`)

	chunk, err := cap.ParseStreamingChunk(data)
	if err != nil {
		t.Fatalf("ParseStreamingChunk failed: %v", err)
	}

	if content := chunk.Content(); content != "World" {
		t.Errorf("got content %q, want %q", content, "World")
	}
}

func TestStandardStreamingCapability_ParseStreamingChunk_EmptyLine(t *testing.T) {
	cap := capabilities.NewStandardStreamingCapability("test-cap", protocols.Chat, nil)

	data := []byte(``)

	_, err := cap.ParseStreamingChunk(data)
	if err == nil {
		t.Error("expected error for empty line, got nil")
	}
}

func TestStandardStreamingCapability_ParseStreamingChunk_DoneMarker(t *testing.T) {
	cap := capabilities.NewStandardStreamingCapability("test-cap", protocols.Chat, nil)

	data := []byte(`data: [DONE]`)

	_, err := cap.ParseStreamingChunk(data)
	if err == nil {
		t.Error("expected error for [DONE] marker, got nil")
	}
}

func TestStandardStreamingCapability_ParseStreamingChunk_InvalidJSON(t *testing.T) {
	cap := capabilities.NewStandardStreamingCapability("test-cap", protocols.Chat, nil)

	data := []byte(`data: {invalid json}`)

	_, err := cap.ParseStreamingChunk(data)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}
