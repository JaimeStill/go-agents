# Composable Capabilities Architecture

## Problem Context

The current ModelFormat architecture creates several critical issues that prevent flexible capability composition and proper protocol handling:

### Core Issues

1. **Option Validation Conflicts**: Model-level options (like `top_p`) are passed to all protocols, causing failures when protocols don't support certain options
   ```
   Error: unsupported option: top_p
   ```
   This occurs because the tools protocol doesn't recognize `top_p`, even though it's a valid chat option.

2. **Format Proliferation**: Creating separate model formats for different capability combinations leads to an explosion of similar formats
   - `openai-standard` (chat + vision + tools + embeddings)
   - `openai-chat` (chat only)
   - `openai-tools` (chat + tools)
   - Each combination requires a separate format definition

3. **Rigid Bundling**: Cannot selectively enable/disable capabilities or mix capability implementations from different "formats"
   - Models like `llama3.2:3b` support chat and tools but not vision or embeddings
   - Currently requires creating new format combinations for each capability subset

4. **Implicit Protocol Support**: No explicit declaration of which protocols a model actually supports
   - Protocols fail at runtime rather than configuration time
   - No clear way to query what capabilities are available

### Root Cause

The current architecture bundles capabilities at the ModelFormat level, creating tight coupling between:
- Model identity
- Capability implementations
- Protocol support
- Option validation

This prevents the flexibility needed for real-world model capabilities and provider differences.

## Architecture Approach

Transform from ModelFormat-centric bundling to composable capability configuration, where:

1. **Capability Format Registration**: Each capability implementation is registered as a named format
2. **Stateless Capabilities**: Capabilities remain pure protocol behavior implementations
3. **Stateful Protocol Handlers**: ProtocolHandlers manage capability + options state
4. **Explicit Protocol Fields**: Models have explicit fields for each protocol's handler
5. **Configuration-Driven Composition**: Models compose capabilities by specifying format names in configuration
6. **Dynamic Option Updates**: Protocol options can be updated on live models for long-lived agents

### Architectural Transformation

**Current Architecture**:
```
ModelFormat → [Chat, Vision, Tools, Embeddings] → Mixed Options
```

**New Architecture**:
```
Configuration → Capability Format Names → Registry → Capability Instances
            ↓
Model → ProtocolHandlers → Stateless Capabilities + Stateful Options
      → Isolated Options per Protocol
```

### Configuration Structure

```json
{
  "model": {
    "name": "llama3.2:3b",
    "capabilities": {
      "chat": {
        "format": "openai-chat",
        "options": {
          "max_tokens": 4096,
          "temperature": 0.7,
          "top_p": 0.95
        }
      },
      "tools": {
        "format": "openai-tools",
        "options": {
          "max_tokens": 4096,
          "temperature": 0.7,
          "tool_choice": "auto"
        }
      }
    }
  }
}
```

Benefits:
- `top_p` only passed to chat protocol (supports it)
- `tool_choice` only passed to tools protocol (needs it)
- Vision and embeddings are nil (not configured)
- Each protocol has isolated, validated options
- Clear which capability format implements each protocol
- Options can be updated dynamically for long-lived agents

## Implementation Strategy

### Design Principles

This refactor maintains the existing capability infrastructure while enabling flexible composition:

1. **Preserve Working Code**: The existing `pkg/capabilities/` package works well - we keep the Capability interface intact
2. **Registry Pattern**: Add capability format registration without changing capability implementations
3. **Stateless Capabilities**: Capabilities remain pure protocol behavior definitions
4. **Stateful Handlers**: ProtocolHandlers manage capability state and options
5. **Explicit Protocol Fields**: Model has dedicated fields for each protocol's handler
6. **Configuration-Driven**: Format selection happens in configuration
7. **Clean Removal**: Remove ModelFormat layer entirely
8. **Bottom-Up Refactoring**: Follow package dependency hierarchy from low to high level

### Package Dependency Order

Refactoring proceeds from lowest-level to highest-level packages:

1. `pkg/capabilities` - Add registry for capability formats
2. `pkg/config` - Configuration structures for capability composition
3. `pkg/models` - Transform from format-based to protocol handler composition
4. `pkg/providers` - Minor updates for capability selection
5. `pkg/transport` - Update option handling for protocol-specific options
6. `pkg/agent` - No changes needed

## Step-by-Step Implementation

### Step 1: Create Capability Registry

Add a registry for named capability formats without changing the existing Capability interface.

#### 1.1 Create Registry (`pkg/capabilities/registry.go`)

```go
package capabilities

import (
	"fmt"
	"sync"
)

// CapabilityFactory creates instances of capabilities
type CapabilityFactory func() Capability

// capabilityRegistry manages registered capability factories
type capabilityRegistry struct {
	mu         sync.RWMutex
	factories  map[string]CapabilityFactory
}

var registry = &capabilityRegistry{
	factories: make(map[string]CapabilityFactory),
}

// RegisterFormat registers a capability factory with a format name
func RegisterFormat(name string, factory CapabilityFactory) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.factories[name] = factory
}

// GetFormat retrieves a capability instance by format name
func GetFormat(name string) (Capability, error) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	factory, exists := registry.factories[name]
	if !exists {
		return nil, fmt.Errorf("capability format '%s' not registered", name)
	}

	return factory(), nil
}

// ListFormats returns all registered capability format names
func ListFormats() []string {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	names := make([]string, 0, len(registry.factories))
	for name := range registry.factories {
		names = append(names, name)
	}
	return names
}
```

#### 1.2 Register Existing Capability Formats (`pkg/capabilities/init.go`)

```go
package capabilities

func init() {
	// Register OpenAI Chat format
	RegisterFormat("openai-chat", func() Capability {
		return NewChatCapability("openai-chat", []CapabilityOption{
			{Option: "max_tokens", Required: false, DefaultValue: 4096},
			{Option: "temperature", Required: false, DefaultValue: 0.7},
			{Option: "top_p", Required: false, DefaultValue: nil},
			{Option: "frequency_penalty", Required: false, DefaultValue: nil},
			{Option: "presence_penalty", Required: false, DefaultValue: nil},
			{Option: "stop", Required: false, DefaultValue: nil},
			{Option: "stream", Required: false, DefaultValue: false},
		})
	})

	// Register OpenAI Vision format
	RegisterFormat("openai-vision", func() Capability {
		return NewVisionCapability("openai-vision", []CapabilityOption{
			{Option: "images", Required: true, DefaultValue: nil},
			{Option: "max_tokens", Required: false, DefaultValue: 4096},
			{Option: "temperature", Required: false, DefaultValue: 0.7},
			{Option: "detail", Required: false, DefaultValue: "auto"},
			{Option: "stream", Required: false, DefaultValue: false},
		})
	})

	// Register OpenAI Tools format
	RegisterFormat("openai-tools", func() Capability {
		return NewToolsCapability("openai-tools", []CapabilityOption{
			{Option: "tools", Required: true, DefaultValue: nil},
			{Option: "tool_choice", Required: false, DefaultValue: "auto"},
			{Option: "max_tokens", Required: false, DefaultValue: 4096},
			{Option: "temperature", Required: false, DefaultValue: 0.7},
			{Option: "stream", Required: false, DefaultValue: false},
		})
	})

	// Register OpenAI Embeddings format
	RegisterFormat("openai-embeddings", func() Capability {
		return NewEmbeddingsCapability("openai-embeddings", []CapabilityOption{
			{Option: "input", Required: true, DefaultValue: nil},
			{Option: "dimensions", Required: false, DefaultValue: nil},
			{Option: "encoding_format", Required: false, DefaultValue: "float"},
		})
	})

	// Register OpenAI Reasoning format (no temperature/top_p)
	RegisterFormat("openai-reasoning", func() Capability {
		return NewChatCapability("openai-reasoning", []CapabilityOption{
			{Option: "max_completion_tokens", Required: true, DefaultValue: nil},
			{Option: "stream", Required: false, DefaultValue: false},
		})
	})
}
```

### Step 2: Update Configuration Package

Define configuration structures to support capability composition.

#### 2.1 Update Model Configuration (`pkg/config/model.go`)

```go
package config

// CapabilityConfig represents configuration for a single capability
type CapabilityConfig struct {
	Format  string         `json:"format"`
	Options map[string]any `json:"options,omitempty"`
}

// ModelCapabilities maps protocol names to their capability configurations
type ModelCapabilities map[string]CapabilityConfig

// ModelConfig represents model configuration
type ModelConfig struct {
	Name         string            `json:"name,omitempty"`
	Capabilities ModelCapabilities `json:"capabilities,omitempty"`
}

func DefaultModelConfig() *ModelConfig {
	return &ModelConfig{
		Capabilities: make(ModelCapabilities),
	}
}

func (c *ModelConfig) Merge(source *ModelConfig) {
	if source.Name != "" {
		c.Name = source.Name
	}

	// Merge capabilities
	if source.Capabilities != nil {
		if c.Capabilities == nil {
			c.Capabilities = make(ModelCapabilities)
		}
		for protocol, capConfig := range source.Capabilities {
			c.Capabilities[protocol] = capConfig
		}
	}
}
```

### Step 3: Update Model Layer

Transform the model layer to use ProtocolHandlers for stateful capability management.

#### 3.1 Define ProtocolHandler (`pkg/models/handler.go`)

```go
package models

import (
	"github.com/JaimeStill/go-agents/pkg/capabilities"
)

// ProtocolHandler manages a capability instance with its configured options
type ProtocolHandler struct {
	capability capabilities.Capability
	options    map[string]any
}

// NewProtocolHandler creates a new protocol handler
func NewProtocolHandler(capability capabilities.Capability, options map[string]any) *ProtocolHandler {
	return &ProtocolHandler{
		capability: capability,
		options:    copyOptions(options),
	}
}

// Capability returns the underlying capability
func (h *ProtocolHandler) Capability() capabilities.Capability {
	return h.capability
}

// Options returns the configured options for this protocol
func (h *ProtocolHandler) Options() map[string]any {
	return h.options
}

// UpdateOptions updates the configured options (for long-lived agents)
func (h *ProtocolHandler) UpdateOptions(newOptions map[string]any) {
	h.options = mergeOptions(h.options, newOptions)
}

// MergeRequestOptions merges configured options with request-time options
func (h *ProtocolHandler) MergeRequestOptions(requestOptions map[string]any) map[string]any {
	return mergeOptions(h.options, requestOptions)
}

// Helper functions
func copyOptions(options map[string]any) map[string]any {
	if options == nil {
		return make(map[string]any)
	}
	copied := make(map[string]any)
	for k, v := range options {
		copied[k] = v
	}
	return copied
}

func mergeOptions(base, override map[string]any) map[string]any {
	merged := make(map[string]any)
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range override {
		merged[k] = v
	}
	return merged
}
```

#### 3.2 Update Model Interface and Implementation (`pkg/models/model.go`)

```go
package models

import (
	"fmt"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

type Model interface {
	Name() string

	// Protocol support checking
	SupportsProtocol(p protocols.Protocol) bool

	// Get capability for a protocol
	GetCapability(p protocols.Protocol) (capabilities.Capability, error)

	// Get protocol-specific options
	GetProtocolOptions(p protocols.Protocol) map[string]any

	// Update protocol options (for long-lived agents)
	UpdateProtocolOptions(p protocols.Protocol, options map[string]any) error

	// Merge request options with configured options
	MergeRequestOptions(p protocols.Protocol, requestOptions map[string]any) map[string]any
}

type model struct {
	name string

	// Explicit fields for each protocol's handler
	chat       *ProtocolHandler
	vision     *ProtocolHandler
	tools      *ProtocolHandler
	embeddings *ProtocolHandler
}

func New(cfg *config.ModelConfig) (Model, error) {
	m := &model{
		name: cfg.Name,
	}

	// Initialize protocol handlers from configuration
	for protocolName, capConfig := range cfg.Capabilities {
		// Get capability format from registry
		capability, err := capabilities.GetFormat(capConfig.Format)
		if err != nil {
			return nil, fmt.Errorf("failed to get capability format '%s' for protocol %s: %w",
				capConfig.Format, protocolName, err)
		}

		// Validate options for this capability
		if err := capability.ValidateOptions(capConfig.Options); err != nil {
			return nil, fmt.Errorf("invalid options for %s protocol: %w", protocolName, err)
		}

		// Create protocol handler
		handler := NewProtocolHandler(capability, capConfig.Options)

		// Assign to appropriate protocol field
		switch protocols.Protocol(protocolName) {
		case protocols.Chat:
			m.chat = handler
		case protocols.Vision:
			m.vision = handler
		case protocols.Tools:
			m.tools = handler
		case protocols.Embeddings:
			m.embeddings = handler
		default:
			return nil, fmt.Errorf("unknown protocol: %s", protocolName)
		}
	}

	return m, nil
}

func (m *model) Name() string {
	return m.name
}

func (m *model) SupportsProtocol(p protocols.Protocol) bool {
	return m.getHandler(p) != nil
}

func (m *model) GetCapability(p protocols.Protocol) (capabilities.Capability, error) {
	handler := m.getHandler(p)
	if handler == nil {
		return nil, fmt.Errorf("protocol %s not supported by model %s", p, m.name)
	}
	return handler.Capability(), nil
}

func (m *model) GetProtocolOptions(p protocols.Protocol) map[string]any {
	handler := m.getHandler(p)
	if handler == nil {
		return make(map[string]any)
	}
	return handler.Options()
}

func (m *model) UpdateProtocolOptions(p protocols.Protocol, options map[string]any) error {
	handler := m.getHandler(p)
	if handler == nil {
		return fmt.Errorf("protocol %s not supported by model %s", p, m.name)
	}

	// Validate new options against capability
	if err := handler.Capability().ValidateOptions(options); err != nil {
		return fmt.Errorf("invalid options for %s protocol: %w", p, err)
	}

	handler.UpdateOptions(options)
	return nil
}

func (m *model) MergeRequestOptions(p protocols.Protocol, requestOptions map[string]any) map[string]any {
	handler := m.getHandler(p)
	if handler == nil {
		return requestOptions
	}
	return handler.MergeRequestOptions(requestOptions)
}

// Helper method to get protocol handler
func (m *model) getHandler(p protocols.Protocol) *ProtocolHandler {
	switch p {
	case protocols.Chat:
		return m.chat
	case protocols.Vision:
		return m.vision
	case protocols.Tools:
		return m.tools
	case protocols.Embeddings:
		return m.embeddings
	default:
		return nil
	}
}
```

#### 3.3 Remove ModelFormat Files

Delete the following files as they are no longer needed:
- `pkg/models/format.go`
- `pkg/models/registry.go`
- `pkg/models/openai.go` (capability format registrations moved to pkg/capabilities/init.go)

### Step 4: Update Transport Layer

Modify transport to use protocol-specific options from the new model structure.

#### 4.1 Update Transport Client (`pkg/transport/client.go`)

Update the `ExecuteProtocol` method to use the new model interface:

```go
func (c *client) ExecuteProtocol(ctx context.Context, req *capabilities.CapabilityRequest) (any, error) {
	// Get capability for this protocol
	capability, err := c.model.GetCapability(req.Protocol)
	if err != nil {
		return nil, fmt.Errorf("protocol %s not supported: %w", req.Protocol, err)
	}

	// Merge model's configured options with request options
	mergedOptions := c.model.MergeRequestOptions(req.Protocol, req.Options)

	// Create request using capability
	protocolRequest, err := capability.CreateRequest(&capabilities.CapabilityRequest{
		Protocol: req.Protocol,
		Messages: req.Messages,
		Options:  mergedOptions,
	}, c.model)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Continue with existing execution logic...
	providerRequest, err := c.provider.PrepareRequest(ctx, req.Protocol, protocolRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	resp, err := c.httpClient.Do(providerRequest.ToHTTPRequest())
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	return c.provider.ProcessResponse(resp, capability)
}
```

Update `ExecuteProtocolStream` similarly:

```go
func (c *client) ExecuteProtocolStream(ctx context.Context, req *capabilities.CapabilityRequest) (<-chan protocols.StreamingChunk, error) {
	// Get capability for this protocol
	capability, err := c.model.GetCapability(req.Protocol)
	if err != nil {
		return nil, fmt.Errorf("protocol %s not supported: %w", req.Protocol, err)
	}

	// Check if capability supports streaming
	streamingCapability, ok := capability.(capabilities.StreamingCapability)
	if !ok {
		return nil, fmt.Errorf("protocol %s does not support streaming", req.Protocol)
	}

	// Merge model's configured options with request options
	mergedOptions := c.model.MergeRequestOptions(req.Protocol, req.Options)

	// Create streaming request using capability
	protocolRequest, err := streamingCapability.CreateStreamingRequest(&capabilities.CapabilityRequest{
		Protocol: req.Protocol,
		Messages: req.Messages,
		Options:  mergedOptions,
	}, c.model)
	if err != nil {
		return nil, fmt.Errorf("failed to create streaming request: %w", err)
	}

	// Continue with existing streaming logic...
	providerRequest, err := c.provider.PrepareStreamRequest(ctx, req.Protocol, protocolRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare stream request: %w", err)
	}

	resp, err := c.httpClient.Do(providerRequest.ToHTTPRequest())
	if err != nil {
		return nil, fmt.Errorf("stream request failed: %w", err)
	}

	streamOutput, err := c.provider.ProcessStreamResponse(ctx, resp, streamingCapability)
	if err != nil {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to process stream: %w", err)
	}

	// Convert provider stream to protocol stream
	output := make(chan protocols.StreamingChunk)
	go func() {
		defer close(output)
		for chunk := range streamOutput {
			if streamChunk, ok := chunk.(*protocols.StreamingChunk); ok {
				select {
				case output <- *streamChunk:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return output, nil
}
```

### Step 5: Update Configuration Files

Transform existing configuration files to the new capability-based format.

#### 5.1 Update Ollama Configuration (`tools/prompt-agent/config.ollama.json`)

```json
{
  "name": "ollama-agent",
  "system_prompt": "You are an expert software architect specializing in cloud native systems design",
  "transport": {
    "provider": {
      "name": "ollama",
      "base_url": "http://localhost:11434",
      "model": {
        "name": "llama3.2:3b",
        "capabilities": {
          "chat": {
            "format": "openai-chat",
            "options": {
              "max_tokens": 4096,
              "temperature": 0.7,
              "top_p": 0.95
            }
          },
          "tools": {
            "format": "openai-tools",
            "options": {
              "max_tokens": 4096,
              "temperature": 0.7,
              "tool_choice": "auto"
            }
          }
        }
      }
    },
    "timeout": 24000000000,
    "max_retries": 3,
    "retry_backoff_base": 1000000000,
    "connection_pool_size": 10,
    "connection_timeout": 9000000000
  }
}
```

#### 5.2 Update Azure Configuration (`tools/prompt-agent/config.azure.json`)

```json
{
  "name": "azure-key-agent",
  "system_prompt": "You are a paranoid schizophrenic who thinks they are interfacing with a human through a neural network installed on a computer.",
  "transport": {
    "provider": {
      "name": "azure",
      "base_url": "https://go-agents-platform.openai.azure.com/openai",
      "model": {
        "name": "o3-mini",
        "capabilities": {
          "chat": {
            "format": "openai-reasoning",
            "options": {
              "max_completion_tokens": 4096
            }
          }
        }
      },
      "options": {
        "deployment": "o3-mini",
        "api_version": "2025-01-01-preview",
        "auth_type": "api_key"
      }
    },
    "timeout": 12000000000,
    "max_retries": 3,
    "retry_backoff_base": 1000000000,
    "connection_pool_size": 10,
    "connection_timeout": 9000000000
  }
}
```

#### 5.3 Create Helper for Common Configurations

Create helper functions for common capability combinations:

```go
// pkg/models/helpers.go
package models

import "github.com/JaimeStill/go-agents/pkg/config"

// StandardOpenAICapabilities returns a configuration for standard OpenAI models
func StandardOpenAICapabilities(maxTokens int, temperature float64) config.ModelCapabilities {
	return config.ModelCapabilities{
		"chat": config.CapabilityConfig{
			Format: "openai-chat",
			Options: map[string]any{
				"max_tokens":  maxTokens,
				"temperature": temperature,
				"top_p":       0.95,
			},
		},
		"vision": config.CapabilityConfig{
			Format: "openai-vision",
			Options: map[string]any{
				"max_tokens":  maxTokens,
				"temperature": temperature,
				"detail":      "auto",
			},
		},
		"tools": config.CapabilityConfig{
			Format: "openai-tools",
			Options: map[string]any{
				"max_tokens":  maxTokens,
				"temperature": temperature,
				"tool_choice": "auto",
			},
		},
		"embeddings": config.CapabilityConfig{
			Format: "openai-embeddings",
			Options: map[string]any{
				"encoding_format": "float",
			},
		},
	}
}

// ChatOnlyCapabilities returns a configuration for chat-only models
func ChatOnlyCapabilities(format string, maxTokens int, temperature float64) config.ModelCapabilities {
	return config.ModelCapabilities{
		"chat": config.CapabilityConfig{
			Format: format,
			Options: map[string]any{
				"max_tokens":  maxTokens,
				"temperature": temperature,
			},
		},
	}
}

// ReasoningCapabilities returns a configuration for reasoning models
func ReasoningCapabilities(maxTokens int) config.ModelCapabilities {
	return config.ModelCapabilities{
		"chat": config.CapabilityConfig{
			Format: "openai-reasoning",
			Options: map[string]any{
				"max_completion_tokens": maxTokens,
			},
		},
	}
}
```

## Long-Lived Agent Support

The ProtocolHandler architecture enables dynamic option updates for long-lived agents:

### Web Application Example

```go
// Initialize agent from configuration
agent := LoadAgent(config)

// Later, user adjusts temperature via web interface
err := agent.GetModel().UpdateProtocolOptions(protocols.Chat, map[string]any{
    "temperature": 0.9,
})
if err != nil {
    log.Printf("Failed to update options: %v", err)
}

// Next chat request uses new temperature
response, err := agent.Chat(ctx, "Hello!")
```

### Option Merging Priority

1. **Capability defaults** - Defined in capability option definitions
2. **Model configuration** - Set during model initialization
3. **Runtime updates** - Applied via `UpdateProtocolOptions`
4. **Request options** - Passed with individual requests

Each level overrides the previous, providing maximum flexibility.

## Future Extensibility

### Adding New Capability Formats

To add a new capability format (e.g., Anthropic's Claude format):

```go
// 1. Create capability implementation (if needed)
type AnthropicChatCapability struct {
	*StandardStreamingCapability
}

func NewAnthropicChatCapability() *AnthropicChatCapability {
	return &AnthropicChatCapability{
		StandardStreamingCapability: NewStandardStreamingCapability(
			"anthropic-chat",
			protocols.Chat,
			[]CapabilityOption{
				{Option: "max_tokens", Required: true, DefaultValue: nil},
				{Option: "temperature", Required: false, DefaultValue: 1.0},
				// Anthropic-specific options
			},
		),
	}
}

// 2. Register the format
func init() {
	RegisterFormat("anthropic-chat", func() Capability {
		return NewAnthropicChatCapability()
	})
}

// 3. Use in configuration
{
  "model": {
    "name": "claude-3-sonnet",
    "capabilities": {
      "chat": {
        "format": "anthropic-chat",
        "options": {
          "max_tokens": 4096,
          "temperature": 0.7
        }
      }
    }
  }
}
```

### Mixed Format Compositions

Models can mix capability formats from different providers:

```json
{
  "model": {
    "name": "multi-format-model",
    "capabilities": {
      "chat": {
        "format": "openai-chat",
        "options": {"temperature": 0.7}
      },
      "vision": {
        "format": "anthropic-vision",
        "options": {"detail": "high"}
      },
      "tools": {
        "format": "custom-tools",
        "options": {"execution_mode": "sandbox"}
      }
    }
  }
}
```

### Native Provider Formats

Support native provider APIs alongside OpenAI-compatible formats:

```go
// Register Ollama native format
RegisterFormat("ollama-native", func() Capability {
	return NewOllamaNativeCapability()
})

// Use in configuration
{
  "model": {
    "name": "llama3.2:3b",
    "capabilities": {
      "chat": {
        "format": "ollama-native",  // Uses Ollama's native API
        "options": {
          "num_predict": 4096,
          "temperature": 0.7,
          "top_k": 40,
          "top_p": 0.95
        }
      }
    }
  }
}
```

## Summary

This refactored architecture addresses all the core issues through a clean, composable design:

1. **Eliminates Option Conflicts**: Each protocol has isolated options through ProtocolHandlers
2. **Reduces Format Proliferation**: Capability formats are registered once and composed as needed
3. **Explicit Protocol Support**: Model has explicit fields for each protocol's handler (nil if not configured)
4. **Configuration-Driven Composition**: Models are composed by selecting registered formats in configuration
5. **Clean Architecture**: Removes ModelFormat entirely in favor of direct capability composition
6. **Future-Proof Design**: Supports both stateless CLI usage and long-lived web applications
7. **Dynamic Flexibility**: Protocol options can be updated on live models

The implementation follows these principles:
- **Minimal Disruption**: Core Capability interface unchanged
- **Clear Separation**: Capabilities handle behavior, ProtocolHandlers manage state
- **Explicit Structure**: Model has dedicated fields for each protocol handler
- **Registry Pattern**: Capability formats are registered and retrieved by name
- **Clean Break**: No backward compatibility complexity
- **Bottom-Up Refactoring**: Changes flow from low-level to high-level packages

This approach provides maximum flexibility while maintaining a clean, understandable architecture where:
- Configuration explicitly declares which formats implement which protocols
- Models have clear, type-safe protocol support through ProtocolHandlers
- Options are isolated per protocol and can be updated dynamically
- Capabilities remain stateless protocol behavior definitions
- The system is easily extensible with new formats and supports various agent lifecycles