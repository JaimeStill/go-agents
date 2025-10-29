package mock_test

import (
	"context"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/mock"
	"github.com/JaimeStill/go-agents/pkg/types"
)

func TestNewMockClient(t *testing.T) {
	client := mock.NewMockClient()

	if client == nil {
		t.Fatal("NewMockClient returned nil")
	}
}

func TestMockClient_ExecuteProtocol(t *testing.T) {
	expectedResponse := &types.ChatResponse{
		Model: "test-model",
	}

	client := mock.NewMockClient(
		mock.WithExecuteResponse(expectedResponse, nil),
	)

	chatRequest := &types.ChatRequest{
		Messages: []types.Message{
			types.NewMessage("user", "Hello"),
		},
		Options: map[string]any{"model": "test-model"},
	}

	result, err := client.ExecuteProtocol(context.Background(), chatRequest)

	if err != nil {
		t.Fatalf("ExecuteProtocol failed: %v", err)
	}

	if result != expectedResponse {
		t.Error("returned different response than configured")
	}
}

func TestMockClient_ExecuteProtocolStream(t *testing.T) {
	// Create properly typed StreamingChunk
	chunk := &types.StreamingChunk{
		Model: "test-model",
	}
	chunk.Choices = make([]struct {
		Index        int     `json:"index"`
		Delta        struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	}, 1)
	chunk.Choices[0].Delta.Content = "Hello"

	chunks := []*types.StreamingChunk{chunk}

	client := mock.NewMockClient(
		mock.WithStreamResponse(chunks, nil),
	)

	chatRequest := &types.ChatRequest{
		Messages: []types.Message{
			types.NewMessage("user", "Hello"),
		},
		Options: map[string]any{"model": "test-model", "stream": true},
	}

	stream, err := client.ExecuteProtocolStream(context.Background(), chatRequest)

	if err != nil {
		t.Fatalf("ExecuteProtocolStream failed: %v", err)
	}

	count := 0
	for chunk := range stream {
		if chunk.Error != nil {
			t.Fatalf("Stream error: %v", chunk.Error)
		}
		count++
	}

	if count != len(chunks) {
		t.Errorf("got %d chunks, want %d", count, len(chunks))
	}
}

func TestMockClient_IsHealthy(t *testing.T) {
	tests := []struct {
		name     string
		healthy  bool
		expected bool
	}{
		{
			name:     "healthy",
			healthy:  true,
			expected: true,
		},
		{
			name:     "unhealthy",
			healthy:  false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := mock.NewMockClient(
				mock.WithHealthy(tt.healthy),
			)

			if client.IsHealthy() != tt.expected {
				t.Errorf("got IsHealthy() = %v, want %v", client.IsHealthy(), tt.expected)
			}
		})
	}
}
