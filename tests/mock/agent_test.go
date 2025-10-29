package mock_test

import (
	"context"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/mock"
	"github.com/JaimeStill/go-agents/pkg/types"
)

func TestNewMockAgent(t *testing.T) {
	agent := mock.NewMockAgent(
		mock.WithID("test-id"),
	)

	if agent == nil {
		t.Fatal("NewMockAgent returned nil")
	}

	if agent.ID() != "test-id" {
		t.Errorf("got ID %q, want %q", agent.ID(), "test-id")
	}
}

func TestMockAgent_Chat(t *testing.T) {
	expectedResponse := &types.ChatResponse{
		Model: "test-model",
		Choices: []struct {
			Index        int            `json:"index"`
			Message      types.Message  `json:"message"`
			Delta        *struct {
				Role    string `json:"role,omitempty"`
				Content string `json:"content,omitempty"`
			} `json:"delta,omitempty"`
			FinishReason string `json:"finish_reason,omitempty"`
		}{
			{
				Index:   0,
				Message: types.NewMessage("assistant", "Hello"),
			},
		},
	}

	agent := mock.NewMockAgent(
		mock.WithID("test-id"),
		mock.WithChatResponse(expectedResponse, nil),
	)

	response, err := agent.Chat(context.Background(), "test")

	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if response != expectedResponse {
		t.Error("returned different response than configured")
	}
}

func TestMockAgent_Vision(t *testing.T) {
	expectedResponse := &types.ChatResponse{
		Model: "test-model",
		Choices: []struct {
			Index        int            `json:"index"`
			Message      types.Message  `json:"message"`
			Delta        *struct {
				Role    string `json:"role,omitempty"`
				Content string `json:"content,omitempty"`
			} `json:"delta,omitempty"`
			FinishReason string `json:"finish_reason,omitempty"`
		}{
			{
				Index:   0,
				Message: types.NewMessage("assistant", "I see an image"),
			},
		},
	}

	agent := mock.NewMockAgent(
		mock.WithID("test-id"),
		mock.WithVisionResponse(expectedResponse, nil),
	)

	response, err := agent.Vision(context.Background(), "test", []string{"image.png"})

	if err != nil {
		t.Fatalf("Vision failed: %v", err)
	}

	if response != expectedResponse {
		t.Error("returned different response than configured")
	}
}

func TestMockAgent_Tools(t *testing.T) {
	expectedResponse := &types.ToolsResponse{
		Model: "test-model",
		Choices: []struct {
			Index   int `json:"index"`
			Message struct {
				Role      string           `json:"role"`
				Content   string           `json:"content"`
				ToolCalls []types.ToolCall `json:"tool_calls,omitempty"`
			} `json:"message"`
			FinishReason string `json:"finish_reason,omitempty"`
		}{
			{
				Index: 0,
				Message: struct {
					Role      string           `json:"role"`
					Content   string           `json:"content"`
					ToolCalls []types.ToolCall `json:"tool_calls,omitempty"`
				}{
					Role:    "assistant",
					Content: "",
					ToolCalls: []types.ToolCall{
						{
							ID:   "call_123",
							Type: "function",
							Function: types.ToolCallFunction{
								Name:      "test_func",
								Arguments: `{}`,
							},
						},
					},
				},
			},
		},
	}

	agent := mock.NewMockAgent(
		mock.WithID("test-id"),
		mock.WithToolsResponse(expectedResponse, nil),
	)

	response, err := agent.Tools(context.Background(), "test", nil)

	if err != nil {
		t.Fatalf("Tools failed: %v", err)
	}

	if response != expectedResponse {
		t.Error("returned different response than configured")
	}
}

func TestMockAgent_Embed(t *testing.T) {
	expectedResponse := &types.EmbeddingsResponse{
		Object: "list",
		Model:  "test-model",
		Data: []struct {
			Embedding []float64 `json:"embedding"`
			Index     int       `json:"index"`
			Object    string    `json:"object"`
		}{
			{
				Embedding: []float64{0.1, 0.2, 0.3},
				Index:     0,
				Object:    "embedding",
			},
		},
	}

	agent := mock.NewMockAgent(
		mock.WithID("test-id"),
		mock.WithEmbeddingsResponse(expectedResponse, nil),
	)

	response, err := agent.Embed(context.Background(), "test")

	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if response != expectedResponse {
		t.Error("returned different response than configured")
	}
}

func TestNewSimpleChatAgent(t *testing.T) {
	agent := mock.NewSimpleChatAgent("test-id", "Hello, world!")

	response, err := agent.Chat(context.Background(), "test")

	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if response.Content() != "Hello, world!" {
		t.Errorf("got content %q, want %q", response.Content(), "Hello, world!")
	}
}

func TestNewStreamingChatAgent(t *testing.T) {
	agent := mock.NewStreamingChatAgent("test-id", []string{"Hello", ", ", "world!"})

	stream, err := agent.ChatStream(context.Background(), "test")

	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}

	var content string
	for chunk := range stream {
		if chunk.Error != nil {
			t.Fatalf("Stream error: %v", chunk.Error)
		}
		content += chunk.Content()
	}

	if content != "Hello, world!" {
		t.Errorf("got content %q, want %q", content, "Hello, world!")
	}
}
