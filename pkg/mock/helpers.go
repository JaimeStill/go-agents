package mock

import (
	"github.com/JaimeStill/go-agents/pkg/response"
)

// NewSimpleChatAgent creates a MockAgent configured for simple chat responses.
// Useful for basic orchestration testing without complex protocol handling.
func NewSimpleChatAgent(id string, content string) *MockAgent {
	return NewMockAgent(
		WithID(id),
		WithChatResponse(&response.Response{
			Role: "assistant",
			Content: []response.ContentBlock{
				response.TextBlock{Text: content},
			},
		}, nil),
	)
}

// NewStreamingChatAgent creates a MockAgent configured for streaming chat.
// Returns chunks sequentially when ChatStream is called.
func NewStreamingChatAgent(id string, chunks []string) *MockAgent {
	streamChunks := make([]response.StreamingResponse, len(chunks))
	for i, content := range chunks {
		streamChunks[i] = response.StreamingResponse{
			Content: []response.ContentBlock{
				response.TextBlock{Text: content},
			},
		}
	}

	return NewMockAgent(
		WithID(id),
		WithStreamChunks(streamChunks, nil),
	)
}

// NewToolsAgent creates a MockAgent configured for tool calling.
// Returns tool calls in the Tools response.
func NewToolsAgent(id string, toolCalls []response.ToolUseBlock) *MockAgent {
	content := make([]response.ContentBlock, len(toolCalls))
	for i, tc := range toolCalls {
		content[i] = tc
	}

	return NewMockAgent(
		WithID(id),
		WithToolsResponse(&response.Response{
			Role:    "assistant",
			Content: content,
		}, nil),
	)
}

// NewEmbeddingsAgent creates a MockAgent configured for embeddings generation.
// Returns the provided embeddings vector.
func NewEmbeddingsAgent(id string, embedding []float64) *MockAgent {
	return NewMockAgent(
		WithID(id),
		WithEmbeddingsResponse(&response.EmbeddingsResponse{
			Embeddings: [][]float64{embedding},
			Model:      "mock-model",
		}, nil),
	)
}

// NewMultiProtocolAgent creates a MockAgent configured for multiple protocols.
// Useful for testing agents that handle different protocol types.
func NewMultiProtocolAgent(id string) *MockAgent {
	chatResponse := &response.Response{
		Role: "assistant",
		Content: []response.ContentBlock{
			response.TextBlock{Text: "Mock chat response"},
		},
	}

	toolsResponse := &response.Response{
		Role:    "assitant",
		Content: []response.ContentBlock{},
	}

	return NewMockAgent(
		WithID(id),
		WithChatResponse(chatResponse, nil),
		WithVisionResponse(chatResponse, nil),
		WithToolsResponse(toolsResponse, nil),
		WithEmbeddingsResponse(&response.EmbeddingsResponse{
			Embeddings: [][]float64{{0.1, 0.2, 0.3}},
			Model:      "mock-model",
		}, nil),
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
