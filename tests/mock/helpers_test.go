package mock_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/agent"
	"github.com/JaimeStill/go-agents/pkg/mock"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

func TestNewSimpleChatAgent(t *testing.T) {
	id := "test-agent"
	response := "Hello, world!"

	a := mock.NewSimpleChatAgent(id, response)

	if a.ID() != id {
		t.Errorf("expected ID %q, got %q", id, a.ID())
	}

	ctx := context.Background()
	chatResponse, err := a.Chat(ctx, "test prompt")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if chatResponse == nil {
		t.Fatal("expected non-nil chat response")
	}

	if len(chatResponse.Choices) == 0 {
		t.Fatal("expected at least one choice")
	}

	message := chatResponse.Choices[0].Message
	if message.Content != response {
		t.Errorf("expected message content %q, got %q", response, message.Content)
	}
}

func TestNewStreamingChatAgent(t *testing.T) {
	id := "streaming-agent"
	chunks := []string{"Hello", " ", "world", "!"}

	a := mock.NewStreamingChatAgent(id, chunks)

	if a.ID() != id {
		t.Errorf("expected ID %q, got %q", id, a.ID())
	}

	ctx := context.Background()
	ch, err := a.ChatStream(ctx, "test prompt")

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

	// Verify chunk content
	for i, chunk := range received {
		if len(chunk.Choices) == 0 {
			t.Fatalf("chunk %d has no choices", i)
		}

		content := chunk.Choices[0].Delta.Content
		if content != chunks[i] {
			t.Errorf("chunk %d: expected content %q, got %q", i, chunks[i], content)
		}
	}
}

func TestNewToolsAgent(t *testing.T) {
	id := "tools-agent"
	toolCalls := []protocols.ToolCall{
		{
			ID:   "call-1",
			Type: "function",
			Function: protocols.ToolCallFunction{
				Name:      "get_weather",
				Arguments: `{"location": "San Francisco"}`,
			},
		},
	}

	a := mock.NewToolsAgent(id, toolCalls)

	if a.ID() != id {
		t.Errorf("expected ID %q, got %q", id, a.ID())
	}

	ctx := context.Background()
	tools := []agent.Tool{}
	toolsResponse, err := a.Tools(ctx, "test prompt", tools)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if toolsResponse == nil {
		t.Fatal("expected non-nil tools response")
	}

	if len(toolsResponse.Choices) == 0 {
		t.Fatal("expected at least one choice")
	}

	returnedToolCalls := toolsResponse.Choices[0].Message.ToolCalls
	if len(returnedToolCalls) != len(toolCalls) {
		t.Errorf("expected %d tool calls, got %d", len(toolCalls), len(returnedToolCalls))
	}

	if len(returnedToolCalls) > 0 {
		if returnedToolCalls[0].ID != toolCalls[0].ID {
			t.Errorf("expected tool call ID %q, got %q", toolCalls[0].ID, returnedToolCalls[0].ID)
		}
	}
}

func TestNewEmbeddingsAgent(t *testing.T) {
	id := "embeddings-agent"
	embedding := []float64{0.1, 0.2, 0.3, 0.4, 0.5}

	a := mock.NewEmbeddingsAgent(id, embedding)

	if a.ID() != id {
		t.Errorf("expected ID %q, got %q", id, a.ID())
	}

	ctx := context.Background()
	embeddingsResponse, err := a.Embed(ctx, "test input")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if embeddingsResponse == nil {
		t.Fatal("expected non-nil embeddings response")
	}

	if len(embeddingsResponse.Data) == 0 {
		t.Fatal("expected at least one embedding")
	}

	returnedEmbedding := embeddingsResponse.Data[0].Embedding
	if len(returnedEmbedding) != len(embedding) {
		t.Errorf("expected embedding length %d, got %d", len(embedding), len(returnedEmbedding))
	}

	for i, val := range embedding {
		if returnedEmbedding[i] != val {
			t.Errorf("embedding[%d]: expected %f, got %f", i, val, returnedEmbedding[i])
		}
	}
}

func TestNewMultiProtocolAgent(t *testing.T) {
	id := "multi-protocol-agent"

	a := mock.NewMultiProtocolAgent(id)

	if a.ID() != id {
		t.Errorf("expected ID %q, got %q", id, a.ID())
	}

	ctx := context.Background()

	// Test chat capability
	chatResponse, err := a.Chat(ctx, "test prompt")
	if err != nil {
		t.Errorf("unexpected chat error: %v", err)
	}
	if chatResponse == nil {
		t.Error("expected non-nil chat response")
	}

	// Test vision capability
	visionResponse, err := a.Vision(ctx, "describe", []string{"image.jpg"})
	if err != nil {
		t.Errorf("unexpected vision error: %v", err)
	}
	if visionResponse == nil {
		t.Error("expected non-nil vision response")
	}

	// Test tools capability
	toolsResponse, err := a.Tools(ctx, "test prompt", []agent.Tool{})
	if err != nil {
		t.Errorf("unexpected tools error: %v", err)
	}
	if toolsResponse == nil {
		t.Error("expected non-nil tools response")
	}

	// Test embeddings capability
	embeddingsResponse, err := a.Embed(ctx, "test input")
	if err != nil {
		t.Errorf("unexpected embeddings error: %v", err)
	}
	if embeddingsResponse == nil {
		t.Error("expected non-nil embeddings response")
	}
	if len(embeddingsResponse.Data) == 0 {
		t.Error("expected at least one embedding")
	}
}

func TestNewFailingAgent(t *testing.T) {
	id := "failing-agent"
	expectedError := errors.New("test error")

	a := mock.NewFailingAgent(id, expectedError)

	if a.ID() != id {
		t.Errorf("expected ID %q, got %q", id, a.ID())
	}

	ctx := context.Background()

	// Test that all methods return the expected error
	_, chatErr := a.Chat(ctx, "test")
	if chatErr != expectedError {
		t.Errorf("Chat: expected error %v, got %v", expectedError, chatErr)
	}

	_, visionErr := a.Vision(ctx, "test", []string{})
	if visionErr != expectedError {
		t.Errorf("Vision: expected error %v, got %v", expectedError, visionErr)
	}

	_, toolsErr := a.Tools(ctx, "test", []agent.Tool{})
	if toolsErr != expectedError {
		t.Errorf("Tools: expected error %v, got %v", expectedError, toolsErr)
	}

	_, embedErr := a.Embed(ctx, "test")
	if embedErr != expectedError {
		t.Errorf("Embed: expected error %v, got %v", expectedError, embedErr)
	}

	_, streamErr := a.ChatStream(ctx, "test")
	if streamErr != expectedError {
		t.Errorf("ChatStream: expected error %v, got %v", expectedError, streamErr)
	}

	_, visionStreamErr := a.VisionStream(ctx, "test", []string{})
	if visionStreamErr != expectedError {
		t.Errorf("VisionStream: expected error %v, got %v", expectedError, visionStreamErr)
	}
}
