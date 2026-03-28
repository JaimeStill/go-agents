// Package agent provides a high-level interface for LLM interactions.
// It wraps the client, provider, and format layers with convenient methods
// for common operations like chat, vision, tools, and embeddings, with automatic
// system prompt injection and option merging.
//
// # Agent Interface
//
// The Agent interface provides protocol-specific methods:
//
//	type Agent interface {
//	    ID() string
//	    Client() client.Client
//	    Provider() providers.Provider
//	    Format() format.Format
//	    Model() *model.Model
//
//	    Chat(ctx, prompt, opts...) (*response.Response, error)
//	    ChatStream(ctx, prompt, opts...) (<-chan *response.StreamingResponse, error)
//	    Vision(ctx, prompt, images, opts...) (*response.Response, error)
//	    VisionStream(ctx, prompt, images, opts...) (<-chan *response.StreamingResponse, error)
//	    Tools(ctx, prompt, tools, opts...) (*response.Response, error)
//	    Embed(ctx, input, opts...) (*response.EmbeddingsResponse, error)
//	}
//
// All protocol methods that return conversational content (Chat, Vision, Tools)
// return a unified [response.Response] with typed content blocks. Use
// [response.Response.Text] for text content and [response.Response.ToolCalls]
// for tool use blocks.
//
// # Creating an Agent
//
// Agents are created from configuration:
//
//	cfg := &config.AgentConfig{
//	    Name:         "my-agent",
//	    SystemPrompt: "You are a helpful assistant.",
//	    Format:       "openai",
//	    Provider: &config.ProviderConfig{
//	        Name: "ollama",
//	    },
//	    Model: &config.ModelConfig{
//	        Name: "llama3.2:3b",
//	        Capabilities: map[string]map[string]any{
//	            "chat": {"temperature": 0.7},
//	        },
//	    },
//	}
//
//	a, err := agent.New(cfg)
//
// The Format field defaults to "openai" when empty. Providers auto-construct
// their default base URL when BaseURL is not specified.
//
// # Chat Protocol
//
//	resp, err := a.Chat(ctx, "What is Go?")
//	fmt.Println(resp.Text())
//
// Streaming:
//
//	chunks, err := a.ChatStream(ctx, "Tell me a story")
//	for chunk := range chunks {
//	    if chunk.Error != nil {
//	        log.Printf("Stream error: %v", chunk.Error)
//	        continue
//	    }
//	    fmt.Print(chunk.Text())
//	}
//
// # Vision Protocol
//
// Images use the provider-neutral [format.Image] type:
//
//	images := []format.Image{
//	    {Data: pngBytes, Format: "png"},
//	    {URL: "https://example.com/image.jpg"},
//	}
//
//	resp, err := a.Vision(ctx, "What do you see?", images)
//	fmt.Println(resp.Text())
//
// # Tools Protocol
//
//	tools := []agent.Tool{
//	    {
//	        Name:        "get_weather",
//	        Description: "Get weather for a location",
//	        Parameters:  map[string]any{"type": "object", "properties": map[string]any{...}},
//	    },
//	}
//
//	resp, err := a.Tools(ctx, "What's the weather in Boston?", tools)
//	for _, tc := range resp.ToolCalls() {
//	    fmt.Printf("Tool: %s, Input: %v\n", tc.Name, tc.Input)
//	}
//
// # Embeddings Protocol
//
//	resp, err := a.Embed(ctx, "The quick brown fox")
//	fmt.Printf("Dimensions: %d\n", len(resp.Embeddings[0]))
package agent
