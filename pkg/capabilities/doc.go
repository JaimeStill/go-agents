// Package capabilities provides a flexible capability system for LLM interactions.
// It defines protocols (chat, vision, tools, embeddings) and manages format-specific
// implementations through a registry pattern.
//
// # Capability System
//
// The capability system separates protocol definitions from format implementations.
// Protocols define what operation to perform (chat, vision, etc.), while formats
// define how to structure requests and parse responses for specific providers.
//
// # Core Interfaces
//
// Capability defines the base interface for all capability implementations:
//
//	type Capability interface {
//	    Name() string
//	    Protocol() protocols.Protocol
//	    Options() []CapabilityOption
//	    ValidateOptions(options map[string]any) error
//	    ProcessOptions(options map[string]any) (map[string]any, error)
//	    CreateRequest(req *CapabilityRequest, model string) (*protocols.Request, error)
//	    ParseResponse(data []byte) (any, error)
//	    SupportsStreaming() bool
//	}
//
// StreamingCapability extends Capability with streaming support:
//
//	type StreamingCapability interface {
//	    Capability
//	    CreateStreamingRequest(req *CapabilityRequest, model string) (*protocols.Request, error)
//	    ParseStreamingChunk(data []byte) (*protocols.StreamingChunk, error)
//	    IsStreamComplete(data string) bool
//	}
//
// # Base Implementations
//
// StandardCapability provides a base implementation for non-streaming capabilities:
//   - Options validation and processing
//   - Default value handling
//   - Required option checking
//
// StandardStreamingCapability extends StandardCapability with streaming support:
//   - SSE (Server-Sent Events) format parsing
//   - Stream completion detection
//   - Chunk parsing
//
// # Concrete Capabilities
//
// The package includes implementations for all major protocols:
//
// ChatCapability - Text-based conversation:
//
//	cap := capabilities.NewChatCapability("openai-chat", []CapabilityOption{
//	    {Option: "temperature", Required: false, DefaultValue: 0.7},
//	    {Option: "max_tokens", Required: false, DefaultValue: 4096},
//	})
//
// VisionCapability - Image understanding with multimodal inputs:
//
//	cap := capabilities.NewVisionCapability("openai-vision", []CapabilityOption{
//	    {Option: "images", Required: true, DefaultValue: nil},
//	    {Option: "detail", Required: false, DefaultValue: "auto"},
//	})
//
// ToolsCapability - Function calling:
//
//	cap := capabilities.NewToolsCapability("openai-tools", []CapabilityOption{
//	    {Option: "tools", Required: true, DefaultValue: nil},
//	    {Option: "tool_choice", Required: false, DefaultValue: "auto"},
//	})
//
// EmbeddingsCapability - Text vectorization:
//
//	cap := capabilities.NewEmbeddingsCapability("openai-embeddings", []CapabilityOption{
//	    {Option: "input", Required: true, DefaultValue: nil},
//	    {Option: "dimensions", Required: false, DefaultValue: nil},
//	})
//
// # Registry Pattern
//
// Capabilities are registered by format name using a factory pattern:
//
//	capabilities.RegisterFormat("openai-chat", func() Capability {
//	    return NewChatCapability("openai-chat", options)
//	})
//
// Retrieve capabilities from the registry:
//
//	cap, err := capabilities.GetFormat("openai-chat")
//	if err != nil {
//	    // handle error
//	}
//
// List all registered formats:
//
//	formats := capabilities.ListFormats()
//
// # Usage Example
//
// Complete workflow for using a capability:
//
//	// Get capability from registry
//	cap, err := capabilities.GetFormat("openai-chat")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create a capability request
//	req := &capabilities.CapabilityRequest{
//	    Protocol: protocols.Chat,
//	    Messages: []protocols.Message{
//	        protocols.NewMessage("user", "What is Go?"),
//	    },
//	    Options: map[string]any{
//	        "temperature": 0.8,
//	        "max_tokens":  2000,
//	    },
//	}
//
//	// Create protocol request
//	protocolReq, err := cap.CreateRequest(req, "gpt-4")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Send request and parse response
//	responseData := // ... get response from API
//	response, err := cap.ParseResponse(responseData)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Thread Safety
//
// The capability registry is thread-safe and supports concurrent registration
// and retrieval operations using read-write mutex synchronization.
package capabilities
