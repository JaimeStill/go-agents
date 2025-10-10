// Package mock provides mock implementations of core go-agents interfaces for testing.
//
// This package enables testing of code that depends on go-agents without requiring
// real LLM providers or network connections. Each mock is configurable with
// predetermined responses and behavior.
//
// # Mock Implementations
//
// MockAgent: Implements agent.Agent interface with configurable protocol responses
//
// MockClient: Implements transport.Client interface for transport layer testing
//
// MockProvider: Implements providers.Provider interface with endpoint mapping
//
// MockModel: Implements models.Model interface with protocol support configuration
//
// MockCapability: Implements capabilities.Capability interface for format testing
//
// # Usage Example
//
//	// Create a mock agent with predetermined chat response
//	mockAgent := mock.NewMockAgent(
//	    mock.WithChatResponse(&protocols.ChatResponse{
//	        Choices: []struct{ Message protocols.Message }{
//	            {Message: protocols.NewMessage("assistant", "Test response")},
//	        },
//	    }),
//	)
//
//	// Use in tests
//	response, err := mockAgent.Chat(context.Background(), "test prompt")
//	// response contains the predetermined response
//
// # Streaming Support
//
// Streaming methods return pre-populated channels that can be configured
// with test chunks:
//
//	mockAgent := mock.NewMockAgent(
//	    mock.WithStreamChunks([]protocols.StreamingChunk{
//	        {Content: "chunk1"},
//	        {Content: "chunk2"},
//	    }),
//	)
//
//	chunks, _ := mockAgent.ChatStream(context.Background(), "prompt")
//	for chunk := range chunks {
//	    // Process test chunks
//	}
package mock
