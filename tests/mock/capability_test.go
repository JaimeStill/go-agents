package mock_test

import (
	"errors"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/mock"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

func TestNewMockCapability_DefaultConfiguration(t *testing.T) {
	c := mock.NewMockCapability()

	if c.Name() != "mock-capability" {
		t.Errorf("expected name 'mock-capability', got %q", c.Name())
	}

	if c.Protocol() != protocols.Chat {
		t.Errorf("expected Chat protocol, got %v", c.Protocol())
	}

	if !c.SupportsStreaming() {
		t.Error("expected streaming support by default")
	}
}

func TestMockCapability_WithCapabilityName(t *testing.T) {
	customName := "custom-capability"
	c := mock.NewMockCapability(mock.WithCapabilityName(customName))

	if c.Name() != customName {
		t.Errorf("expected name %q, got %q", customName, c.Name())
	}
}

func TestMockCapability_WithCapabilityProtocol(t *testing.T) {
	c := mock.NewMockCapability(mock.WithCapabilityProtocol(protocols.Vision))

	if c.Protocol() != protocols.Vision {
		t.Errorf("expected Vision protocol, got %v", c.Protocol())
	}
}

func TestMockCapability_WithSupportsStreaming(t *testing.T) {
	tests := []struct {
		name             string
		supportsStream   bool
		expectedSupports bool
	}{
		{
			name:             "supports streaming",
			supportsStream:   true,
			expectedSupports: true,
		},
		{
			name:             "does not support streaming",
			supportsStream:   false,
			expectedSupports: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := mock.NewMockCapability(mock.WithSupportsStreaming(tt.supportsStream))

			if c.SupportsStreaming() != tt.expectedSupports {
				t.Errorf("expected SupportsStreaming() = %v, got %v", tt.expectedSupports, c.SupportsStreaming())
			}
		})
	}
}

func TestMockCapability_ValidateOptions_Success(t *testing.T) {
	c := mock.NewMockCapability()

	err := c.ValidateOptions(map[string]any{})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMockCapability_ValidateOptions_WithError(t *testing.T) {
	expectedError := errors.New("validation error")
	c := mock.NewMockCapability(mock.WithValidateError(expectedError))

	err := c.ValidateOptions(map[string]any{})

	if err != expectedError {
		t.Errorf("expected error %v, got %v", expectedError, err)
	}
}

func TestMockCapability_ProcessOptions_Success(t *testing.T) {
	processedOptions := map[string]any{
		"temperature": 0.8,
		"max_tokens":  1000,
	}

	c := mock.NewMockCapability(mock.WithProcessedOptions(processedOptions, nil))

	result, err := c.ProcessOptions(map[string]any{})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result["temperature"] != 0.8 {
		t.Error("expected processed temperature option")
	}

	if result["max_tokens"] != 1000 {
		t.Error("expected processed max_tokens option")
	}
}

func TestMockCapability_ProcessOptions_WithError(t *testing.T) {
	expectedError := errors.New("process error")
	c := mock.NewMockCapability(mock.WithProcessedOptions(nil, expectedError))

	result, err := c.ProcessOptions(map[string]any{})

	if err != expectedError {
		t.Errorf("expected error %v, got %v", expectedError, err)
	}

	if result != nil {
		t.Error("expected nil result")
	}
}

func TestMockCapability_CreateRequest_Success(t *testing.T) {
	expectedRequest := &protocols.Request{
		Messages: []protocols.Message{
			protocols.NewMessage("user", "test"),
		},
	}

	c := mock.NewMockCapability(mock.WithCreateRequest(expectedRequest, nil))

	req := &capabilities.CapabilityRequest{
		Messages: []protocols.Message{},
	}

	result, err := c.CreateRequest(req, "test-model")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result != expectedRequest {
		t.Error("request does not match expected")
	}
}

func TestMockCapability_CreateRequest_WithError(t *testing.T) {
	expectedError := errors.New("create request error")
	c := mock.NewMockCapability(mock.WithCreateRequest(nil, expectedError))

	req := &capabilities.CapabilityRequest{
		Messages: []protocols.Message{},
	}

	result, err := c.CreateRequest(req, "test-model")

	if err != expectedError {
		t.Errorf("expected error %v, got %v", expectedError, err)
	}

	if result != nil {
		t.Error("expected nil result")
	}
}

func TestMockCapability_ParseResponse_Success(t *testing.T) {
	expectedResponse := &protocols.ChatResponse{
		Model: "test-model",
	}

	c := mock.NewMockCapability(mock.WithParseResponse(expectedResponse, nil))

	result, err := c.ParseResponse([]byte(`{}`))

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result != expectedResponse {
		t.Error("response does not match expected")
	}
}

func TestMockCapability_ParseResponse_WithError(t *testing.T) {
	expectedError := errors.New("parse error")
	c := mock.NewMockCapability(mock.WithParseResponse(nil, expectedError))

	result, err := c.ParseResponse([]byte(`{}`))

	if err != expectedError {
		t.Errorf("expected error %v, got %v", expectedError, err)
	}

	if result != nil {
		t.Error("expected nil result")
	}
}

func TestMockCapability_CreateStreamingRequest_Success(t *testing.T) {
	expectedRequest := &protocols.Request{
		Messages: []protocols.Message{
			protocols.NewMessage("user", "test"),
		},
	}

	c := mock.NewMockCapability(mock.WithCreateRequest(expectedRequest, nil))

	req := &capabilities.CapabilityRequest{
		Messages: []protocols.Message{},
	}

	result, err := c.CreateStreamingRequest(req, "test-model")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result != expectedRequest {
		t.Error("request does not match expected")
	}
}

func TestMockCapability_ParseStreamingChunk_Success(t *testing.T) {
	expectedChunk := &protocols.StreamingChunk{
		Model: "test-model",
	}

	c := mock.NewMockCapability(mock.WithStreamChunk(expectedChunk, nil))

	result, err := c.ParseStreamingChunk([]byte(`{}`))

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result != expectedChunk {
		t.Error("chunk does not match expected")
	}
}

func TestMockCapability_ParseStreamingChunk_WithError(t *testing.T) {
	expectedError := errors.New("parse chunk error")
	c := mock.NewMockCapability(mock.WithStreamChunk(nil, expectedError))

	result, err := c.ParseStreamingChunk([]byte(`{}`))

	if err != expectedError {
		t.Errorf("expected error %v, got %v", expectedError, err)
	}

	if result != nil {
		t.Error("expected nil result")
	}
}

func TestMockCapability_IsStreamComplete(t *testing.T) {
	tests := []struct {
		name     string
		complete bool
		expected bool
	}{
		{
			name:     "stream complete",
			complete: true,
			expected: true,
		},
		{
			name:     "stream not complete",
			complete: false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := mock.NewMockCapability(mock.WithStreamComplete(tt.complete))

			result := c.IsStreamComplete("")

			if result != tt.expected {
				t.Errorf("expected IsStreamComplete() = %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMockCapability_ImplementsInterfaces(t *testing.T) {
	var _ capabilities.Capability = (*mock.MockCapability)(nil)
	var _ capabilities.StreamingCapability = (*mock.MockCapability)(nil)
}
