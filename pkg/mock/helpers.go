package mock

import (
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

// NewSimpleChatAgent creates a MockAgent configured for simple chat responses.
// Useful for basic orchestration testing without complex protocol handling.
func NewSimpleChatAgent(id string, response string) *MockAgent {
	chatResponse := &protocols.ChatResponse{
		Model: "mock-model",
	}
	chatResponse.Choices = append(chatResponse.Choices, struct {
		Index   int     `json:"index"`
		Message protocols.Message `json:"message"`
		Delta   *struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta,omitempty"`
		FinishReason string `json:"finish_reason,omitempty"`
	}{
		Index:   0,
		Message: protocols.NewMessage("assistant", response),
	})

	return NewMockAgent(
		WithID(id),
		WithChatResponse(chatResponse, nil),
	)
}

// NewStreamingChatAgent creates a MockAgent configured for streaming chat.
// Returns chunks sequentially when ChatStream is called.
func NewStreamingChatAgent(id string, chunks []string) *MockAgent {
	streamChunks := make([]protocols.StreamingChunk, len(chunks))
	for i, content := range chunks {
		chunk := protocols.StreamingChunk{
			Model: "mock-model",
		}
		chunk.Choices = append(chunk.Choices, struct {
			Index int `json:"index"`
			Delta struct {
				Role    string `json:"role,omitempty"`
				Content string `json:"content,omitempty"`
			} `json:"delta"`
			FinishReason *string `json:"finish_reason"`
		}{
			Index: 0,
			Delta: struct {
				Role    string `json:"role,omitempty"`
				Content string `json:"content,omitempty"`
			}{
				Content: content,
			},
		})
		streamChunks[i] = chunk
	}

	return NewMockAgent(
		WithID(id),
		WithStreamChunks(streamChunks, nil),
	)
}

// NewToolsAgent creates a MockAgent configured for tool calling.
// Returns tool calls in the Tools response.
func NewToolsAgent(id string, toolCalls []protocols.ToolCall) *MockAgent {
	toolsResponse := &protocols.ToolsResponse{
		Model: "mock-model",
	}
	toolsResponse.Choices = append(toolsResponse.Choices, struct {
		Index   int `json:"index"`
		Message struct {
			Role      string     `json:"role"`
			Content   string     `json:"content"`
			ToolCalls []protocols.ToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason,omitempty"`
	}{
		Index: 0,
		Message: struct {
			Role      string     `json:"role"`
			Content   string     `json:"content"`
			ToolCalls []protocols.ToolCall `json:"tool_calls,omitempty"`
		}{
			Role:      "assistant",
			Content:   "",
			ToolCalls: toolCalls,
		},
	})

	return NewMockAgent(
		WithID(id),
		WithToolsResponse(toolsResponse, nil),
	)
}

// NewEmbeddingsAgent creates a MockAgent configured for embeddings generation.
// Returns the provided embeddings vector.
func NewEmbeddingsAgent(id string, embedding []float64) *MockAgent {
	embeddingsResponse := &protocols.EmbeddingsResponse{
		Model: "mock-model",
	}
	embeddingsResponse.Data = append(embeddingsResponse.Data, struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
		Object    string    `json:"object"`
	}{
		Embedding: embedding,
		Index:     0,
		Object:    "embedding",
	})

	return NewMockAgent(
		WithID(id),
		WithEmbeddingsResponse(embeddingsResponse, nil),
	)
}

// NewMultiProtocolAgent creates a MockAgent configured for multiple protocols.
// Useful for testing agents that handle different protocol types.
func NewMultiProtocolAgent(id string) *MockAgent {
	chatResponse := &protocols.ChatResponse{
		Model: "mock-model",
	}
	chatResponse.Choices = append(chatResponse.Choices, struct {
		Index   int     `json:"index"`
		Message protocols.Message `json:"message"`
		Delta   *struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta,omitempty"`
		FinishReason string `json:"finish_reason,omitempty"`
	}{
		Index:   0,
		Message: protocols.NewMessage("assistant", "Mock chat response"),
	})

	toolsResponse := &protocols.ToolsResponse{
		Model: "mock-model",
	}
	toolsResponse.Choices = append(toolsResponse.Choices, struct {
		Index   int `json:"index"`
		Message struct {
			Role      string     `json:"role"`
			Content   string     `json:"content"`
			ToolCalls []protocols.ToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason,omitempty"`
	}{
		Index: 0,
		Message: struct {
			Role      string     `json:"role"`
			Content   string     `json:"content"`
			ToolCalls []protocols.ToolCall `json:"tool_calls,omitempty"`
		}{
			Role:      "assistant",
			Content:   "",
			ToolCalls: []protocols.ToolCall{},
		},
	})

	embeddingsResponse := &protocols.EmbeddingsResponse{
		Model: "mock-model",
	}
	embeddingsResponse.Data = append(embeddingsResponse.Data, struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
		Object    string    `json:"object"`
	}{
		Embedding: []float64{0.1, 0.2, 0.3},
		Index:     0,
		Object:    "embedding",
	})

	return NewMockAgent(
		WithID(id),
		WithChatResponse(chatResponse, nil),
		WithVisionResponse(chatResponse, nil),
		WithToolsResponse(toolsResponse, nil),
		WithEmbeddingsResponse(embeddingsResponse, nil),
	)
}

// NewFailingAgent creates a MockAgent that returns errors for all operations.
// Useful for testing error handling in orchestration scenarios.
func NewFailingAgent(id string, err error) *MockAgent {
	return NewMockAgent(
		WithID(id),
		WithChatResponse(nil, err),
		WithVisionResponse(nil, err),
		WithToolsResponse(nil, err),
		WithEmbeddingsResponse(nil, err),
		WithStreamChunks(nil, err),
	)
}
