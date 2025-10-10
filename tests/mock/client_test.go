package mock_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/mock"
	"github.com/JaimeStill/go-agents/pkg/protocols"
	"github.com/JaimeStill/go-agents/pkg/transport"
)

func TestNewMockClient_DefaultConfiguration(t *testing.T) {
	c := mock.NewMockClient()

	if c.Provider() == nil {
		t.Error("expected default provider")
	}

	if c.Model() == nil {
		t.Error("expected default model")
	}

	if !c.IsHealthy() {
		t.Error("expected healthy by default")
	}
}

func TestMockClient_WithHealthy(t *testing.T) {
	tests := []struct {
		name     string
		healthy  bool
		expected bool
	}{
		{
			name:     "healthy client",
			healthy:  true,
			expected: true,
		},
		{
			name:     "unhealthy client",
			healthy:  false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := mock.NewMockClient(mock.WithHealthy(tt.healthy))

			if c.IsHealthy() != tt.expected {
				t.Errorf("expected IsHealthy() = %v, got %v", tt.expected, c.IsHealthy())
			}
		})
	}
}

func TestMockClient_ExecuteProtocol_Success(t *testing.T) {
	expectedResponse := &protocols.ChatResponse{
		Model: "test-model",
	}

	c := mock.NewMockClient(mock.WithExecuteResponse(expectedResponse, nil))

	ctx := context.Background()
	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Chat,
		Messages: []protocols.Message{},
	}

	response, err := c.ExecuteProtocol(ctx, req)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if response != expectedResponse {
		t.Error("response does not match expected")
	}
}

func TestMockClient_ExecuteProtocol_WithError(t *testing.T) {
	expectedError := errors.New("execution error")

	c := mock.NewMockClient(mock.WithExecuteResponse(nil, expectedError))

	ctx := context.Background()
	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Chat,
		Messages: []protocols.Message{},
	}

	response, err := c.ExecuteProtocol(ctx, req)

	if err != expectedError {
		t.Errorf("expected error %v, got %v", expectedError, err)
	}

	if response != nil {
		t.Error("expected nil response")
	}
}

func TestMockClient_ExecuteProtocolStream_Success(t *testing.T) {
	chunks := []protocols.StreamingChunk{
		{Model: "chunk-1"},
		{Model: "chunk-2"},
	}

	c := mock.NewMockClient(mock.WithStreamResponse(chunks, nil))

	ctx := context.Background()
	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Chat,
		Messages: []protocols.Message{},
	}

	ch, err := c.ExecuteProtocolStream(ctx, req)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	received := []protocols.StreamingChunk{}
	for chunk := range ch {
		received = append(received, chunk)
	}

	if len(received) != len(chunks) {
		t.Errorf("expected %d chunks, got %d", len(chunks), len(received))
	}
}

func TestMockClient_ExecuteProtocolStream_WithError(t *testing.T) {
	expectedError := errors.New("stream error")

	c := mock.NewMockClient(mock.WithStreamResponse(nil, expectedError))

	ctx := context.Background()
	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Chat,
		Messages: []protocols.Message{},
	}

	ch, err := c.ExecuteProtocolStream(ctx, req)

	if err != expectedError {
		t.Errorf("expected error %v, got %v", expectedError, err)
	}

	if ch != nil {
		t.Error("expected nil channel")
	}
}

func TestMockClient_ImplementsInterface(t *testing.T) {
	var _ transport.Client = (*mock.MockClient)(nil)
}
