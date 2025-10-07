// Package models provides model abstraction for LLM interactions.
// It manages protocol-capability mappings and provides a unified interface
// for working with different model configurations.
//
// # Model Interface
//
// The Model interface provides a protocol-agnostic way to work with LLM models:
//
//	type Model interface {
//	    Name() string
//	    SupportsProtocol(p protocols.Protocol) bool
//	    GetCapability(p protocols.Protocol) (capabilities.Capability, error)
//	    GetProtocolOptions(p protocols.Protocol) map[string]any
//	    UpdateProtocolOptions(p protocols.Protocol, options map[string]any) error
//	    MergeRequestOptions(p protocols.Protocol, options map[string]any) map[string]any
//	}
//
// Models are created from configuration that specifies which protocols are supported
// and what capability format to use for each protocol.
//
// # Configuration-Based Creation
//
// Models are constructed from configuration structs that specify:
//   - Model name
//   - Supported protocols and their capability formats
//   - Default options for each protocol
//
// Example configuration:
//
//	cfg := &config.ModelConfig{
//	    Name: "gpt-4",
//	    Capabilities: map[string]config.CapabilityConfig{
//	        "chat": {
//	            Format: "openai-chat",
//	            Options: map[string]any{
//	                "temperature": 0.7,
//	                "max_tokens":  4096,
//	            },
//	        },
//	        "vision": {
//	            Format: "openai-vision",
//	            Options: map[string]any{
//	                "detail": "auto",
//	            },
//	        },
//	    },
//	}
//
//	model, err := models.New(cfg)
//
// # Protocol Handlers
//
// Each protocol supported by a model has an associated ProtocolHandler that:
//   - Manages the capability instance for that protocol
//   - Stores default options
//   - Merges model-level options with request-level options
//
// Handlers are internal to the model implementation and not directly exposed.
//
// # Usage Example
//
// Complete workflow using a model:
//
//	// Create model from configuration
//	cfg := &config.ModelConfig{
//	    Name: "gpt-4",
//	    Capabilities: map[string]config.CapabilityConfig{
//	        "chat": {
//	            Format: "openai-chat",
//	            Options: map[string]any{
//	                "temperature": 0.7,
//	                "max_tokens":  4096,
//	            },
//	        },
//	    },
//	}
//
//	model, err := models.New(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Check protocol support
//	if !model.SupportsProtocol(protocols.Chat) {
//	    log.Fatal("chat protocol not supported")
//	}
//
//	// Get capability for a protocol
//	cap, err := model.GetCapability(protocols.Chat)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get model's default options for a protocol
//	modelOptions := model.GetProtocolOptions(protocols.Chat)
//
//	// Merge with request-specific options
//	requestOptions := map[string]any{
//	    "temperature": 0.9,  // Override model default
//	    "top_p":       0.95, // Add new option
//	}
//	finalOptions := model.MergeRequestOptions(protocols.Chat, requestOptions)
//
//	// Update model's default options for a protocol
//	err = model.UpdateProtocolOptions(protocols.Chat, map[string]any{
//	    "temperature": 0.8,
//	})
//
// # Multi-Protocol Models
//
// Models can support multiple protocols simultaneously:
//
//	cfg := &config.ModelConfig{
//	    Name: "gpt-4",
//	    Capabilities: map[string]config.CapabilityConfig{
//	        "chat":       {Format: "openai-chat"},
//	        "vision":     {Format: "openai-vision"},
//	        "tools":      {Format: "openai-tools"},
//	        "embeddings": {Format: "openai-embeddings"},
//	    },
//	}
//
// Each protocol is independent with its own capability and options.
//
// # Options Management
//
// Options flow through three levels:
//  1. Model default options (from configuration)
//  2. Protocol-specific updates (via UpdateProtocolOptions)
//  3. Request-specific overrides (via MergeRequestOptions)
//
// The MergeRequestOptions method combines model defaults with request options,
// with request options taking precedence for overlapping keys.
package models
