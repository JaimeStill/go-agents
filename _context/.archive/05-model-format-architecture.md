# Model Format Architecture Context

## Executive Summary

This document captures the comprehensive architectural design decisions and context for implementing a format-centric compositional capability system in the Agentic Toolkit. The architecture replaces complex capability configuration with a clean compositional approach where ModelFormats define model families and their supported capabilities through reusable CapabilityFormat implementations.

## Problem Context and Evolution

### Original Architecture Issues

The initial approach attempted to separate model capabilities from message formats, creating several problems:

1. **Configuration Complexity**: Separate capability objects with endpoints and parameters
2. **Duplicate Logic**: Capability detection via `SyncCapabilities()` method on providers
3. **Interface Pollution**: Client interface assumed all models support chat completion
4. **Endpoint Management**: Complex routing between capabilities and endpoints
5. **Provider Burden**: Providers needed to implement capability discovery and validation

### Critical Architectural Insight

**The breakthrough realization**: Format should define capabilities holistically. Instead of separating model capabilities from message formats, the format itself should be the authoritative source for what a model family can do.

This insight eliminated:
- Separate capability configuration objects
- `SyncCapabilities()` provider methods
- Complex capability-to-endpoint mapping
- Duplicate configuration management

## Final Architecture Design

### Core Principles

1. **Format-Centric Capabilities**: Message formats are the authoritative source for model capabilities
2. **Compositional Reuse**: CapabilityFormat implementations are shared across multiple ModelFormats
3. **Layered Interfaces**: Clean separation between base capability concerns and streaming extensions
4. **Provider Delegation**: Providers handle transport; formats handle capability-specific logic
5. **Configuration Simplicity**: Single format field defines entire capability set

### Architecture Components

#### 1. CapabilityFormat Interface (Base Layer)
```go
type CapabilityFormat interface {
    Type() Capability
    Endpoint() string

    CreateRequest(data interface{}, config *config.ModelConfig) (FormatRequest, error)
    ParseResponse(data []byte) (interface{}, error)

    SupportsStreaming() (StreamingCapabilityFormat, bool)
}
```

**Key Design Decision**: The `SupportsStreaming()` method provides a type-safe bridge to streaming capabilities without polluting the base interface with streaming-specific methods.

#### 2. StreamingCapabilityFormat Interface (Extension Layer)
```go
type StreamingCapabilityFormat interface {
    CapabilityFormat  // Embed base interface

    CreateStreamingRequest(data interface{}, config *config.ModelConfig) (FormatRequest, error)
    ParseStreamingChunk(data []byte) (*StreamingChunk, error)
    IsStreamComplete(data string) bool
}
```

**Key Design Decision**: Streaming support is provided through interface extension, not base interface pollution. Only capabilities that support streaming implement this extended interface.

#### 3. ModelFormat Structure (Composition Layer)
```go
type ModelFormat struct {
    Name         string                           `json:"name"`
    Capabilities map[Capability]CapabilityFormat  `json:"capabilities,omitempty"`
}
```

**Key Design Decision**: ModelFormat serves as a named composition that maps capability types to their specific implementations. This enables reuse of CapabilityFormat implementations across different ModelFormats.

### Capability Types

```go
type Capability string

const (
    CapabilityCompletion Capability = "completion"
    CapabilityReasoning  Capability = "reasoning"
    CapabilityEmbeddings Capability = "embeddings"
    CapabilityVision     Capability = "vision"
    CapabilityTools      Capability = "tools"
    CapabilityAudio      Capability = "audio"
    CapabilityRealtime   Capability = "realtime"
)
```

## Implementation Strategy

### Two-Phase Approach

**Preparation Phase**: Remove legacy capability infrastructure without breaking existing functionality
- Delete capability configuration files
- Simplify ModelConfig to Format + Name + Options
- Rename Endpoint → URL throughout codebase
- Remove legacy client capability methods

**Feature Phase**: Implement compositional format system
- Create layered capability interfaces
- Implement CapabilityFormat implementations (GPT completion, embeddings, reasoning, vision)
- Create ModelFormat compositions (openai-standard, openai-multimodal, etc.)
- Update provider implementations to delegate to formats
- Update agent layer for capability access

### Package Structure

```
pkg/models/formats/
├── types.go                    # Core interfaces and types
├── capabilities/
│   ├── gpt_completion.go       # GPTCompletionFormat (implements StreamingCapabilityFormat)
│   ├── gpt_reasoning.go        # GPTReasoningFormat (implements StreamingCapabilityFormat)
│   ├── gpt_embeddings.go       # GPTEmbeddingsFormat (CapabilityFormat only)
│   ├── gpt_vision.go           # GPTVisionFormat (implements StreamingCapabilityFormat)
│   └── anthropic_completion.go # AnthropicCompletionFormat
├── models/
│   ├── openai.go              # OpenAI ModelFormat compositions
│   ├── anthropic.go           # Anthropic ModelFormat compositions
│   └── google.go              # Google ModelFormat compositions
└── registry.go                # ModelFormat registration and lookup
```

## Key Design Decisions and Rationale

### 1. Format Granularity Strategy

**Decision**: Create model family-specific formats (openai-standard, openai-multimodal, anthropic) rather than generic capability formats.

**Rationale**: Model families have consistent standards for capabilities they support. This approach aligns formats with native model capabilities and allows providers to handle platform-specific transport details.

### 2. Compositional Reuse Pattern

**Decision**: CapabilityFormat implementations are reusable components that can be composed into different ModelFormats.

**Example**:
```go
var gptCompletion = &GPTCompletionFormat{}

// Reused across multiple model formats
OpenAIStandardFormat().Capabilities[CapabilityCompletion] = gptCompletion
OpenAIMultimodalFormat().Capabilities[CapabilityCompletion] = gptCompletion
```

**Rationale**: Eliminates code duplication while enabling flexible composition based on model family capabilities.

### 3. Streaming Interface Layering

**Decision**: Use interface extension rather than boolean flags or optional methods for streaming support.

**Rationale**:
- Clean separation of concerns
- Type-safe streaming capability detection
- No interface pollution for non-streaming capabilities
- Explicit streaming method contracts

### 4. Provider Delegation Pattern

**Decision**: Providers delegate capability-specific logic to CapabilityFormat implementations rather than implementing capability logic directly.

**Client Implementation Pattern**:
```go
func (c *Client) Chat(ctx context.Context, request *ChatRequest) (*ChatResponse, error) {
    capabilityFormat, exists := c.modelFormat.GetCapability(CapabilityCompletion)
    if !exists {
        return nil, ErrCapabilityNotSupported("completion", c.modelFormat.Name)
    }

    formatReq, err := capabilityFormat.CreateRequest(request.Messages, c.config.Model)
    endpoint := c.URL() + capabilityFormat.Endpoint()
    // Execute HTTP request using format-created request
}
```

**Rationale**:
- Providers focus on transport concerns (HTTP, authentication, connection management)
- Formats handle capability-specific request/response logic
- Clean separation enables easier testing and provider implementation

### 5. Configuration Simplification

**Before**:
```json
{
  "model": {
    "format": "openai-standard",
    "capabilities": [
      {"name": "completion", "endpoint": "/chat/completions", "parameters": {...}}
    ]
  }
}
```

**After**:
```json
{
  "model": {
    "format": "openai-multimodal",
    "name": "gpt-4o",
    "options": {
      "max_tokens": 4096,
      "temperature": 0.7
    }
  }
}
```

**Rationale**: Single format field defines entire capability set, eliminating complex capability configuration while maintaining parameter flexibility through options.

### 6. URL Construction Pattern

**Decision**: Client.URL() + CapabilityFormat.Endpoint() = Full Request URL

**Rationale**:
- Clean separation between provider base URL and capability-specific endpoints
- Supports capability-specific endpoint routing (e.g., /embeddings vs /chat/completions)
- Enables provider-specific URL construction (Azure deployment paths)

## Current State vs Target State

### What Exists Now

- ✅ Basic package structure with `pkg/models/formats/` and `pkg/config/`
- ✅ Legacy MessageFormat interface and implementations (openai_standard.go, openai_reasoning.go)
- ✅ Basic Client interface with Chat/ChatStream methods
- ✅ Provider implementations for Ollama and Azure
- ✅ Agent layer with basic chat functionality
- ✅ Configuration loading and merging infrastructure

### What Needs Implementation

- ❌ CapabilityFormat and StreamingCapabilityFormat interfaces
- ❌ ModelFormat structure and composition system
- ❌ Individual CapabilityFormat implementations (GPTCompletionFormat, etc.)
- ❌ ModelFormat compositions (openai-standard, openai-multimodal, etc.)
- ❌ Updated provider implementations using format delegation
- ❌ Enhanced client interface with capability-specific methods
- ❌ Updated agent layer with capability checking
- ❌ Capability-specific request/response types (EmbeddingRequest, VisionMessage, etc.)
- ❌ Updated registry system for ModelFormat registration

## Implementation Requirements

### Critical Integration Points

1. **Provider URL Construction**: Each provider needs specific URL building logic
   - Ollama: baseURL + "/v1" + capability.Endpoint()
   - Azure: baseURL + "/openai/deployments/" + deployment + capability.Endpoint() + "?api-version=" + version

2. **Authentication Handling**: Provider-specific auth patterns must be preserved
   - Ollama: Optional Bearer token
   - Azure: api-key header or Bearer token based on auth_type

3. **Error Handling**: Capability-specific errors need proper context
   - ErrCapabilityNotSupported when format doesn't support requested capability
   - ErrStreamingNotSupported when capability doesn't support streaming

4. **Agent Integration**: Agents need access to capability information
   - Capability checking methods (SupportsCompletion(), SupportsVision(), etc.)
   - Capability-specific methods (ChatWithVision(), GenerateEmbeddings(), etc.)

### Success Criteria

- [ ] All legacy capability configuration infrastructure removed
- [ ] ModelFormat compositions registered and accessible via format names
- [ ] Provider clients delegate capability logic to CapabilityFormat implementations
- [ ] Agent layer provides capability checking and capability-specific methods
- [ ] Configuration files work with format-based approach
- [ ] Existing chat functionality continues to work unchanged
- [ ] New capabilities (embeddings, vision) accessible through client interface
- [ ] Streaming support works through layered interface system
- [ ] All tests pass and prompt-agent tool works with all configurations

## Testing Strategy

### Configuration Testing
- Test all configuration files load correctly with new format names
- Verify format resolution works for all registered ModelFormats
- Confirm capability checking works correctly

### Capability Testing
- Test completion capability with both streaming and non-streaming
- Test embeddings capability (non-streaming only)
- Test vision capability with multimodal messages
- Test reasoning capability with reasoning-specific parameters

### Provider Testing
- Verify Ollama provider works with format delegation
- Verify Azure provider works with format delegation
- Test URL construction for different capabilities
- Test authentication handling remains unchanged

### Agent Testing
- Test capability checking methods return correct values
- Test capability-specific agent methods work correctly
- Test error handling for unsupported capabilities

## Migration Notes

### Backwards Compatibility
- No backwards compatibility required - this is pre-release architecture refinement
- Existing chat functionality must continue to work seamlessly
- Configuration file format changes are acceptable

### Cleanup Tasks
- Remove `pkg/models/capability.go`
- Remove legacy format files (openai_standard.go, openai_reasoning.go)
- Update all import statements to remove references to deleted files
- Update configuration field names (Endpoint → URL)

## Future Extensibility

### Adding New Capabilities
1. Define new Capability constant
2. Create CapabilityFormat implementation
3. Add to appropriate ModelFormat compositions
4. Update client interface with new methods
5. Update agent layer with capability support

### Adding New Model Families
1. Create provider-specific CapabilityFormat implementations if needed
2. Create new ModelFormat composition
3. Register in format registry
4. Create configuration examples

### Cross-Provider Capability Reuse
The architecture enables providers to reuse CapabilityFormat implementations:
```go
// Ollama can reuse OpenAI capability formats
RegisterModelFormat("ollama-multimodal", &ModelFormat{
    Name: "ollama-multimodal",
    Capabilities: map[Capability]CapabilityFormat{
        CapabilityCompletion: gptCompletion,  // Reuse OpenAI implementation
        CapabilityEmbeddings: gptEmbeddings,  // Reuse OpenAI implementation
    },
})
```

## Implementation Dependencies

### Critical Path
1. Capability interfaces and types → CapabilityFormat implementations → ModelFormat compositions → Provider updates → Agent updates
2. Must maintain working chat functionality throughout implementation
3. Registry updates must happen after ModelFormat compositions are created

### Parallel Development
- CapabilityFormat implementations can be developed in parallel
- Provider updates can happen independently after interfaces are stable
- Configuration file updates can happen after registry is updated

## Architectural Benefits

### Achieved Through This Design

1. **Simplified Configuration**: Single format field replaces complex capability objects
2. **Code Reuse**: CapabilityFormat implementations shared across model families
3. **Type Safety**: Compile-time capability checking through interface composition
4. **Clean Separation**: Formats handle capability logic, providers handle transport
5. **Extensibility**: New capabilities added through composition, not code changes
6. **Provider Independence**: Providers focus on their platform-specific concerns
7. **Interface Clarity**: Layered interfaces provide clean streaming support
8. **Natural Groupings**: ModelFormats align with actual model family capabilities

This architecture establishes a solid foundation for multi-modal agent capabilities while maintaining clean separation of concerns and enabling future extensibility.