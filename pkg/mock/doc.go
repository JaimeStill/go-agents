// Package mock provides mock implementations of core go-agents interfaces for testing.
//
// This package enables testing of code that depends on go-agents without requiring
// real LLM providers or network connections. Each mock is configurable with
// predetermined responses and behavior.
//
// # Mock Implementations
//
// MockAgent: Implements agent.Agent with configurable protocol responses.
//
// MockClient: Implements client.Client for client layer testing.
//
// MockProvider: Implements providers.Provider with endpoint mapping and streaming.
//
// MockFormat: Implements format.Format for format layer testing.
//
// # Usage Example
//
//	mockAgent := mock.NewMockAgent(
//	    mock.WithChatResponse(&response.Response{
//	        Role: "assistant",
//	        Content: []response.ContentBlock{
//	            response.TextBlock{Text: "Test response"},
//	        },
//	    }, nil),
//	)
//
//	resp, err := mockAgent.Chat(context.Background(), "test prompt")
//	fmt.Println(resp.Text()) // "Test response"
//
// # Streaming Support
//
//	mockAgent := mock.NewMockAgent(
//	    mock.WithStreamChunks([]response.StreamingResponse{
//	        {Content: []response.ContentBlock{response.TextBlock{Text: "chunk1"}}},
//	        {Content: []response.ContentBlock{response.TextBlock{Text: "chunk2"}}},
//	    }, nil),
//	)
//
//	chunks, _ := mockAgent.ChatStream(context.Background(), "prompt")
//	for chunk := range chunks {
//	    fmt.Print(chunk.Text())
//	}
//
// # Helper Constructors
//
// Convenience functions for common test scenarios:
//
//	agent := mock.NewSimpleChatAgent("id", "response text")
//	agent := mock.NewStreamingChatAgent("id", []string{"chunk1", "chunk2"})
//	agent := mock.NewToolsAgent("id", []response.ToolUseBlock{...})
//	agent := mock.NewEmbeddingsAgent("id", []float64{0.1, 0.2, 0.3})
//	agent := mock.NewMultiProtocolAgent("id")
//	agent := mock.NewFailingAgent("id", errors.New("test error"))
package mock
