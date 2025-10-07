// Package protocols defines the protocol system for LLM interactions.
// It provides request/response structures and protocol constants for chat,
// vision, tools, and embeddings operations.
//
// The package supports four protocols:
//   - Chat: Standard text-based conversation
//   - Vision: Image understanding with multimodal inputs
//   - Tools: Function calling and tool execution
//   - Embeddings: Text vectorization for semantic search
//
// Request Structure:
//
// All protocols use the Request type which combines messages with protocol-specific
// options. When marshaled to JSON, the options are merged at the root level:
//
//	{
//	  "messages": [{"role": "user", "content": "Hello"}],
//	  "temperature": 0.7,
//	  "max_tokens": 4096
//	}
//
// Response Types:
//
// Each protocol has specific response structures:
//   - ChatResponse: Non-streaming text responses with Content() method
//   - StreamingChunk: Streaming responses with delta content
//   - ToolsResponse: Function calling responses with tool calls
//   - EmbeddingsResponse: Vector embeddings with usage tracking
//
// Helper Functions:
//
// ExtractOption provides type-safe option extraction with defaults:
//
//	temperature := protocols.ExtractOption(options, "temperature", 0.7)
//
// IsValid validates protocol strings, and ValidProtocols returns all supported protocols.
//
// Example Usage:
//
//	// Create a chat request
//	req := &protocols.Request{
//	    Messages: []protocols.Message{
//	        protocols.NewMessage("user", "What is Go?"),
//	    },
//	    Options: map[string]any{
//	        "temperature": 0.7,
//	        "max_tokens":  4096,
//	    },
//	}
//
//	// Marshal to JSON
//	data, _ := req.Marshal()
//
//	// Extract content from response
//	var response protocols.ChatResponse
//	json.Unmarshal(responseData, &response)
//	content := response.Content()
package protocols
