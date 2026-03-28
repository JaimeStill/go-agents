package mock_test

import (
	"context"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/mock"
	"github.com/JaimeStill/go-agents/pkg/response"
)

func TestNewMockClient(t *testing.T) {
	client := mock.NewMockClient()

	if client == nil {
		t.Fatal("NewMockClient returned nil")
	}
}

func TestMockClient_Execute(t *testing.T) {
	expectedResponse := &response.Response{
		Role: "assistant",
		Content: []response.ContentBlock{
			response.TextBlock{Text: "Hello"},
		},
	}

	client := mock.NewMockClient(
		mock.WithExecuteResponse(expectedResponse, nil),
	)

	result, err := client.Execute(context.Background(), nil)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result != expectedResponse {
		t.Error("returned different response than configured")
	}
}

func TestMockClient_ExecuteStream(t *testing.T) {
	chunk := &response.StreamingResponse{
		Content: []response.ContentBlock{
			response.TextBlock{Text: "Hello"},
		},
	}

	chunks := []*response.StreamingResponse{chunk}

	client := mock.NewMockClient(
		mock.WithStreamResponse(chunks, nil),
	)

	stream, err := client.ExecuteStream(context.Background(), nil)

	if err != nil {
		t.Fatalf("ExecuteStream failed: %v", err)
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
