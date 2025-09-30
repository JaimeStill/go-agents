# Development Summary: Composable Capabilities Implementation

## Starting Point

The system had a ModelFormat-centric architecture where each model was assigned a single format (e.g., "openai-standard") that bundled all protocol capabilities together. This created several issues:

1. **Option Validation Conflicts**: Model-level options were being passed to all protocols, causing validation failures when protocol-specific required options (like `tools` for tools protocol) were missing at model initialization time
2. **Inflexible Protocol Support**: Models could not selectively support different protocols with different formats
3. **Configuration Rigidity**: All protocols inherited the same format and base options from the model level
4. **Validation Timing**: Options were validated at model creation time when request-specific parameters weren't yet available, making required options meaningless

The implementation guide `_context/composable-capabilities.md` was created to address these architectural issues by transforming to a protocol-centric capability composition model.

## Implementation Decisions

### Core Architecture Changes

**Composable Capabilities Model**: Replaced single model format with per-protocol capability configuration. Each protocol (chat, vision, tools, embeddings) can now be configured independently with its own format and options.

**Configuration Structure**: Updated `ModelConfig` to use a `capabilities` map instead of single `format` and `options` fields:
```go
type ModelConfig struct {
    Name         string                       `json:"name"`
    Capabilities map[string]CapabilityConfig  `json:"capabilities"`
}
```

**ProtocolHandler Pattern**: Introduced `ProtocolHandler` type to manage stateful protocol configuration (capability + options) while keeping capabilities themselves stateless:
```go
type ProtocolHandler struct {
    capability capabilities.Capability
    options    map[string]any
}
```

### Configuration Enhancements

**Human-Readable Duration Type**: Implemented custom `Duration` type supporting both string format ("24s", "1s") and numeric nanoseconds, with proper JSON unmarshaling:
```go
type Duration time.Duration

func (d *Duration) UnmarshalJSON(data []byte) error {
    // Try string format first (e.g., "24s")
    // Fall back to numeric nanoseconds
}
```

**Validation Timing**: Moved option validation from model initialization to transport layer after option merge. This ensures validation occurs when all options (configuration + request-specific) are available:
- Removed validation in `models.New()`
- Added validation in `transport.ExecuteProtocol()` after `MergeRequestOptions()`

### Protocol-Specific Response Types

**Tools Protocol Isolation**: Created dedicated `ToolsResponse` type to handle tool calls without cluttering base `Message` type:
```go
type ToolsResponse struct {
    Choices []struct {
        Message struct {
            Role      string     `json:"role"`
            Content   string     `json:"content"`
            ToolCalls []ToolCall `json:"tool_calls,omitempty"`
        } `json:"message"`
    } `json:"choices"`
}
```

**Streaming Removal**: Removed streaming support from tools protocol as tool calls are discrete, structured responses that must be complete before execution. Removed:
- `StreamingCapability` embedding from `ToolsCapability`
- `CreateStreamingRequest()` method
- `ToolsStream()` from Agent interface
- `executeToolsStream()` from prompt-agent
- `stream` option from capability registration

### Option Management

**Variadic Options Pattern**: Added per-request option overrides to Agent methods using idiomatic Go variadic parameters:
```go
Chat(ctx context.Context, prompt string, opts ...map[string]any) (*protocols.ChatResponse, error)
```

**Option Merging Strategy**: Established clear separation between persistent configuration options and transient request options:
- Configuration options: Set via `UpdateProtocolOptions()`, persist across requests
- Request options: Passed per-call, merged with configuration options, non-persistent
- `MergeRequestOptions()` returns new map, never mutates handler state

### Type Safety Improvements

**FunctionDefinition Type**: Changed `setToolDefinitions()` return type from `[]map[string]any` to `[]capabilities.FunctionDefinition` to ensure type assertions succeed:
```go
func setToolDefinitions(tools []Tool) []capabilities.FunctionDefinition {
    defs := make([]capabilities.FunctionDefinition, len(tools))
    for i, tool := range tools {
        defs[i] = capabilities.FunctionDefinition{
            Type: "function",
            Function: map[string]any{
                "name":        tool.Name,
                "description": tool.Description,
                "parameters":  tool.Parameters,
            },
        }
    }
    return defs
}
```

## Final Architecture State

### Package Structure

**pkg/config**:
- `ModelConfig` with capabilities map
- `CapabilityConfig` with format and options
- `Duration` type with custom JSON unmarshaling
- No domain validation (deferred to consuming packages)

**pkg/models**:
- `Model` interface with protocol-specific methods
- `ProtocolHandler` for stateful protocol configuration
- Protocol handler composition in model implementation
- Validation removed from initialization

**pkg/capabilities**:
- Stateless capability implementations
- Format registration system unchanged
- Tool-specific types: `FunctionDefinition`, `ToolsCapability`
- Removed streaming support from tools capability

**pkg/protocols**:
- Protocol-specific response types: `ChatResponse`, `ToolsResponse`, `EmbeddingsResponse`
- Tool call types: `ToolCall`, `ToolCallFunction`
- Validation helpers: `IsValid()`, `ProtocolStrings()`

**pkg/transport**:
- Option merging before protocol execution
- Validation after merge in `ExecuteProtocol()` and `ExecuteProtocolStream()`
- Health tracking unchanged

**pkg/agent**:
- Variadic options support on all protocol methods
- Tools method returns `*protocols.ToolsResponse`
- Removed `ToolsStream()` method

### Configuration Examples

**Multi-Protocol Model** (config.ollama.json):
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

**Reasoning Model** (config.azure.json):
```json
{
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
  }
}
```

### Tool Calling Implementation

The tools protocol now properly handles both single and multiple tool calls in a single response:

**Single Tool Call**:
```
Tool Calls:
  - get_weather({"location":"Dallas, TX"})

Tokens: 224 prompt + 19 completion = 243 total
```

**Multiple Tool Calls**:
```
Tool Calls:
  - calculate({"expression":"sqrt(pi)"})
  - get_weather({"location":"Dallas, TX"})

Tokens: 229 prompt + 37 completion = 266 total
```

### Documentation Updates

**README.md**: Updated all configuration examples to use composable capabilities architecture. Added three new sections:
- Tools Protocol (Weather)
- Tools Protocol (Calculate)
- Tools Protocol (Multiple)

All duration values converted to human-readable strings. All capability configurations show proper protocol-specific format assignment.

**Configuration Files**: Updated all prompt-agent configurations:
- `config.ollama.json`: Chat + Tools capabilities
- `config.azure.json` / `config.azure-entra.json`: Reasoning capability only
- `config.gemma.json`: Chat + Vision capabilities (tools removed, unsupported by gemma3:4b)
- `config.embedding.json`: Embeddings capability only

## Current State

### Operational Status

All four core protocols are fully operational with the composable capabilities architecture:

1. **Chat Protocol**: Working with both standard and reasoning formats
2. **Vision Protocol**: Working with multimodal content handling
3. **Tools Protocol**: Working with proper tool call parsing and display (non-streaming only)
4. **Embeddings Protocol**: Working with dimension configuration

### Validation Flow

Options are now validated at the correct time:
1. Configuration loaded with protocol-specific options
2. Model created with protocol handlers (no validation)
3. Request made with optional per-request options
4. Options merged in transport layer
5. **Validation occurs here** after complete options available
6. Protocol executed with validated options

### Testing Coverage

Verified through prompt-agent tool:
- Chat with Ollama (llama3.2:3b)
- Chat with Azure (o3-mini reasoning model)
- Vision with Ollama (gemma3:4b)
- Tools with Ollama (llama3.2:3b) - weather, calculate, multiple calls
- Embeddings with Ollama (embeddinggemma:300m)

### Known Limitations

**Streaming Tools**: Deliberately removed as tool calls require complete structure before execution. Streaming would add complexity for unclear benefit.

**Model Support**: Tool calling support varies by model:
- llama3.2:3b (Ollama): Supports tools
- gemma3:4b (Ollama): Does not support tools
- o3-mini (Azure): Reasoning model, different parameter set

## Architectural Improvements

### Flexibility

Models can now support any combination of protocols with different formats and options. A single model could theoretically support chat with one format and vision with another if needed.

### Maintainability

Protocol-specific concerns are isolated:
- Tool call handling in `ToolsResponse` type
- Embeddings output in `EmbeddingsResponse` type
- Vision content handling in `ChatResponse` type
- No protocol concerns leak into base types

### Configuration Clarity

The capabilities map makes it immediately clear which protocols a model supports and how each is configured. No more implicit format inheritance or shared option conflicts.

### Validation Correctness

Options are validated when complete. Required options make sense again since request-specific parameters are available at validation time.

## Next Steps

The composable capabilities architecture is complete and operational. Future enhancements could include:

1. Additional capability formats for different providers (Anthropic, Google, etc.)
2. More sophisticated option merging strategies if needed
3. Protocol-specific configuration validation rules
4. Extended tool calling features (parallel execution, result feedback loops)

No current blockers exist in the implementation.
