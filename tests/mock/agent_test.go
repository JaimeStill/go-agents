package mock_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/agent"
	"github.com/JaimeStill/go-agents/pkg/mock"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

func TestNewMockAgent_DefaultConfiguration(t *testing.T) {
	m := mock.NewMockAgent()

	if m.ID() == "" {
		t.Error("expected non-empty ID")
	}

	if m.Client() == nil {
		t.Error("expected default client")
	}

	if m.Provider() == nil {
		t.Error("expected default provider")
	}

	if m.Model() == nil {
		t.Error("expected default model")
	}
}

func TestMockAgent_WithID(t *testing.T) {
	customID := "custom-agent-id"
	m := mock.NewMockAgent(mock.WithID(customID))

	if m.ID() != customID {
		t.Errorf("expected ID %q, got %q", customID, m.ID())
	}
}

func TestMockAgent_Chat(t *testing.T) {
	expectedResponse := &protocols.ChatResponse{
		Model: "test-model",
	}
	expectedResponse.Choices = append(expectedResponse.Choices, struct {
		Index   int              `json:"index"`
		Message protocols.Message `json:"message"`
		Delta   *struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta,omitempty"`
		FinishReason string `json:"finish_reason,omitempty"`
	}{
		Index:   0,
		Message: protocols.NewMessage("assistant", "test response"),
	})

	m := mock.NewMockAgent(mock.WithChatResponse(expectedResponse, nil))

	ctx := context.Background()
	response, err := m.Chat(ctx, "test prompt")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if response != expectedResponse {
		t.Error("response does not match expected")
	}
}

func TestMockAgent_Chat_WithError(t *testing.T) {
	expectedError := errors.New("chat error")
	m := mock.NewMockAgent(mock.WithChatResponse(nil, expectedError))

	ctx := context.Background()
	response, err := m.Chat(ctx, "test prompt")

	if err != expectedError {
		t.Errorf("expected error %v, got %v", expectedError, err)
	}

	if response != nil {
		t.Error("expected nil response")
	}
}

func TestMockAgent_ChatStream(t *testing.T) {
	chunks := []protocols.StreamingChunk{
		{Model: "test-model"},
		{Model: "test-model"},
	}

	m := mock.NewMockAgent(mock.WithStreamChunks(chunks, nil))

	ctx := context.Background()
	ch, err := m.ChatStream(ctx, "test prompt")

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

func TestMockAgent_ChatStream_WithError(t *testing.T) {
	expectedError := errors.New("stream error")
	m := mock.NewMockAgent(mock.WithStreamChunks(nil, expectedError))

	ctx := context.Background()
	ch, err := m.ChatStream(ctx, "test prompt")

	if err != expectedError {
		t.Errorf("expected error %v, got %v", expectedError, err)
	}

	if ch != nil {
		t.Error("expected nil channel")
	}
}

func TestMockAgent_Vision(t *testing.T) {
	expectedResponse := &protocols.ChatResponse{
		Model: "vision-model",
	}

	m := mock.NewMockAgent(mock.WithVisionResponse(expectedResponse, nil))

	ctx := context.Background()
	response, err := m.Vision(ctx, "describe image", []string{"image.jpg"})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if response != expectedResponse {
		t.Error("response does not match expected")
	}
}

func TestMockAgent_Tools(t *testing.T) {
	expectedResponse := &protocols.ToolsResponse{
		Model: "tools-model",
	}

	m := mock.NewMockAgent(mock.WithToolsResponse(expectedResponse, nil))

	ctx := context.Background()
	tools := []agent.Tool{}
	response, err := m.Tools(ctx, "test prompt", tools)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if response != expectedResponse {
		t.Error("response does not match expected")
	}
}

func TestMockAgent_Embed(t *testing.T) {
	expectedResponse := &protocols.EmbeddingsResponse{
		Model: "embeddings-model",
	}

	m := mock.NewMockAgent(mock.WithEmbeddingsResponse(expectedResponse, nil))

	ctx := context.Background()
	response, err := m.Embed(ctx, "test input")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if response != expectedResponse {
		t.Error("response does not match expected")
	}
}

func TestMockAgent_ImplementsInterface(t *testing.T) {
	var _ agent.Agent = (*mock.MockAgent)(nil)
}
