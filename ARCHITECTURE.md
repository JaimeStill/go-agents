# Architecture

This document describes the current architecture and implementation patterns of the Go Agents library.

## Package Structure

```
pkg/
├── config/              # Configuration management and loading
│   ├── agent.go         # Agent configuration structure
│   ├── model.go         # Model configuration with format selection
│   ├── provider.go      # Provider configuration structures
│   ├── transport.go     # Transport layer configuration
│   └── options.go       # Configuration option utilities
├── protocols/           # Protocol definitions and message structures
│   └── protocol.go      # Protocol types, Message, Request, Response structures
├── capabilities/        # Protocol-specific capability implementations
│   ├── capability.go    # Core capability interface and standard implementations
│   ├── chat.go          # Chat protocol capability
│   ├── vision.go        # Vision protocol capability with structured content
│   ├── tools.go         # Tools protocol capability
│   └── embeddings.go    # Embeddings protocol capability
├── models/              # Model abstraction and format management
│   ├── model.go         # Model interface and core implementation
│   ├── format.go        # ModelFormat definition and methods
│   ├── openai.go        # OpenAI format implementations (Standard, Chat, Reasoning)
│   └── registry.go      # Thread-safe model format registry
├── providers/           # Provider implementations for different LLM services
│   ├── provider.go      # Provider interface definition
│   ├── base.go          # BaseProvider with common functionality
│   ├── registry.go      # Provider registry and initialization
│   ├── azure.go         # Azure AI Foundry provider implementation
│   └── ollama.go        # Ollama provider implementation
├── transport/           # Transport layer orchestrating requests across providers
│   └── client.go        # Transport client interface and implementation
└── agent/               # High-level agent orchestration
    ├── agent.go         # Agent interface and implementation
    └── errors.go        # Agent-specific error definitions
```

## Core Components

### Protocol System

The protocol system defines the communication contracts for different LLM interactions:

```go
type Protocol string

const (
    Chat       Protocol = "chat"        // Text completion interactions
    Vision     Protocol = "vision"      // Image analysis with text
    Tools      Protocol = "tools"       // Function calling capabilities
    Embeddings Protocol = "embeddings"  // Vector embedding generation
)
```

**Message Structure**: Supports both simple text and structured content for protocols like vision:
```go
type Message struct {
    Role    string `json:"role"`
    Content any    `json:"content"`  // string for text, []map[string]any for structured
}
```

**Request/Response Flow**: Unified request structure with protocol-specific options:
```go
type Request struct {
    Messages []Message
    Options  map[string]any
}
```

### Capability System

Capabilities implement protocol-specific behavior and validation for different API formats:

```go
type Capability interface {
    Name() string
    Protocol() protocols.Protocol
    Options() []CapabilityOption
    CreateRequest(req *CapabilityRequest, model ModelInfo) (*protocols.Request, error)
    CreateStreamingRequest(req *CapabilityRequest, model ModelInfo) (*protocols.Request, error)
    ParseResponse(data []byte) (any, error)
}
```

**Capability Formats**: Different providers may use different API formats for the same protocol:
- **OpenAI Chat Format**: Standard OpenAI chat completions API structure
- **Anthropic Chat Format**: Claude-specific API structure (future)
- **Ollama Native Format**: Native Ollama API structure (future)
- **Custom Formats**: Provider-specific implementations

**Option Management**: Each capability defines supported options with validation:
```go
type CapabilityOption struct {
    Option       string `json:"option"`
    Required     bool   `json:"required"`
    DefaultValue any    `json:"default_value"`
}
```

### Model System

Models represent LLM configurations with format-specific capability mappings:

```go
type Model interface {
    Name() string
    Format() *ModelFormat
    Options() map[string]any
}
```

**Model Formats**: Define capability combinations and format requirements:
- **OpenAI Standard**: Uses OpenAI-format capabilities for all protocols
- **OpenAI Chat**: Text completion using OpenAI chat format
- **OpenAI Reasoning**: Reasoning models with OpenAI format but restricted parameters

**Format Registry**: Thread-safe registration system for extensibility:
```go
func RegisterFormat(name string, format *ModelFormat)
func GetFormat(name string) (*ModelFormat, error)
```

### Provider System

Providers implement LLM service integrations and are format-agnostic:

```go
type Provider interface {
    Name() string
    BaseURL() string
    Model() models.Model
    GetEndpoint(protocol protocols.Protocol) (string, error)
    PrepareRequest(ctx context.Context, protocol protocols.Protocol, request *protocols.Request) (*Request, error)
    ProcessResponse(resp *http.Response, capability capabilities.Capability) (any, error)
}
```

**Provider Responsibilities**:
- **Endpoint Mapping**: Map protocols to provider-specific API endpoints
- **Request Transformation**: Adapt protocol requests to provider API format
- **Authentication**: Handle provider-specific authentication methods
- **Response Processing**: Parse provider responses through capability handlers

**Implemented Providers**:
- **Ollama**: Currently configured for OpenAI-compatible endpoints, but could support native Ollama format
- **Azure**: Azure AI Foundry with OpenAI format support

**Provider Flexibility**: Providers can support any capability format:
```go
// Example: Provider supporting multiple formats
func (p *CustomProvider) PrepareRequest(ctx context.Context, protocol protocols.Protocol, request *protocols.Request) (*Request, error) {
    // The capability has already formatted the request according to its format
    // Provider just needs to route to correct endpoint and handle authentication
    endpoint, err := p.GetEndpoint(protocol)
    if err != nil {
        return nil, err
    }

    // Provider-specific transformations if needed
    return &Request{
        URL:     endpoint,
        Headers: p.getAuthHeaders(),
        Body:    request.Marshal(), // Already formatted by capability
    }, nil
}
```

### Transport Layer

The transport layer orchestrates request routing and execution:

```go
type Client interface {
    Provider() providers.Provider
    Model() models.Model
    ExecuteProtocol(ctx context.Context, req *capabilities.CapabilityRequest) (any, error)
    ExecuteProtocolStream(ctx context.Context, req *capabilities.CapabilityRequest) (<-chan protocols.StreamingChunk, error)
}
```

**Request Flow**:
1. **Protocol Selection**: Determine capability based on requested protocol
2. **Capability Execution**: Use model format to select appropriate capability implementation
3. **Request Preparation**: Transform request through capability-specific logic (formats to API structure)
4. **Provider Routing**: Provider routes to appropriate endpoint and handles authentication
5. **Response Processing**: Parse response through capability-specific handlers

### Agent System

Agents provide high-level orchestration with protocol-specific methods:

```go
type Agent interface {
    Client() transport.Client
    Provider() providers.Provider
    Model() models.Model

    Chat(ctx context.Context, prompt string) (*protocols.ChatResponse, error)
    ChatStream(ctx context.Context, prompt string) (<-chan protocols.StreamingChunk, error)

    Vision(ctx context.Context, prompt string, images []string) (*protocols.ChatResponse, error)
    VisionStream(ctx context.Context, prompt string, images []string) (<-chan protocols.StreamingChunk, error)

    Tools(ctx context.Context, prompt string, tools []Tool) (*protocols.ChatResponse, error)
    ToolsStream(ctx context.Context, prompt string, tools []Tool) (<-chan protocols.StreamingChunk, error)

    Embed(ctx context.Context, input string) (*protocols.EmbeddingsResponse, error)
}
```

## Configuration System

### Agent Configuration

```json
{
  "name": "agent-name",
  "system_prompt": "System instructions for the agent",
  "transport": {
    "provider": {
      "name": "ollama",
      "base_url": "http://localhost:11434",
      "model": {
        "name": "llama3.2:3b",
        "format": "openai-standard",
        "options": {
          "max_tokens": 4096,
          "temperature": 0.7,
          "top_p": 0.95
        }
      }
    },
    "timeout": 60000000000,
    "max_retries": 3,
    "connection_pool_size": 10
  }
}
```

### Model Format Selection

Model formats determine available capabilities and their API format:

- **openai-standard**: Complete capability set using OpenAI API format
- **openai-chat**: Chat-only using OpenAI format with extended parameters
- **openai-reasoning**: Reasoning models using OpenAI format with restricted parameters

Future formats could include:
- **anthropic-standard**: Claude API format capabilities
- **ollama-native**: Native Ollama API format capabilities

### Provider-Format Compatibility

Providers must be compatible with the capability formats used by models:

**Current Compatibility**:
- **Ollama Provider** + **OpenAI Formats**: Uses OpenAI-compatible endpoints (`/v1/chat/completions`)
- **Azure Provider** + **OpenAI Formats**: Native OpenAI API compatibility

**Future Compatibility Examples**:
- **Ollama Provider** + **Ollama Native Formats**: Uses native Ollama endpoints (`/api/chat`)
- **Anthropic Provider** + **Anthropic Formats**: Uses Claude API endpoints
- **Custom Provider** + **Custom Formats**: Proprietary API integration

## Data Flow

### Standard Request Flow

```
Agent.Chat() → Transport.ExecuteProtocol() → Capability.CreateRequest() (formats to API structure) → Provider.PrepareRequest() (routes + auth) → HTTP Request → Provider.ProcessResponse() → Capability.ParseResponse() → Agent Response
```

### Format-Provider Separation

```
Model Format (openai-standard) → OpenAI Chat Capability → OpenAI API Structure → Ollama Provider → /v1/chat/completions endpoint
Model Format (ollama-native) → Ollama Chat Capability → Ollama API Structure → Ollama Provider → /api/chat endpoint
```

## Design Patterns

### Format-Agnostic Provider Design

Providers don't assume any specific API format - they work with whatever format the capabilities produce:

```go
// Provider doesn't know or care about API format
func (p *GenericProvider) ProcessResponse(resp *http.Response, capability capabilities.Capability) (any, error) {
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    // Capability knows how to parse its specific format
    return capability.ParseResponse(body)
}
```

### Capability-Format Coupling

Capabilities are tightly coupled to specific API formats, enabling:
- **Format Specialization**: Each capability optimized for its target API
- **Provider Independence**: Same capability works with any compatible provider
- **Extension Flexibility**: New formats can be added without changing providers

### Configuration-Driven Compatibility

Compatibility between models and providers is determined by configuration:
- **Model Format**: Determines which capability formats are used
- **Provider Implementation**: Must support the endpoints/authentication required by those formats
- **Runtime Validation**: Mismatches detected at agent creation time

## Current Limitations

### Option Validation Conflicts

Model-level options are passed to all protocols, causing validation failures when protocols don't support certain options.

**Planned Resolution**: Protocol-centric capability composition where each protocol has isolated options.

### Format Proliferation

Different capability combinations require separate model formats.

**Planned Resolution**: Composable capabilities architecture allowing per-protocol format selection.

## Extension Points

### Adding New Providers

1. Implement `Provider` interface with endpoint mapping and authentication
2. Ensure compatibility with desired capability formats
3. Register in provider registry
4. No requirement for specific API format - works with any capability format

### Adding New API Formats

1. Implement capability interfaces for each protocol in the new format
2. Create model formats that use these capabilities
3. Ensure providers support required endpoints and authentication
4. Register formats in format registry

### Cross-Format Provider Support

A single provider can support multiple API formats:

```go
func (p *FlexibleProvider) GetEndpoint(protocol protocols.Protocol) (string, error) {
    // Determine endpoint based on model format being used
    format := p.model.Format().Name()

    switch format {
    case "openai-standard":
        return p.openAIEndpoint(protocol), nil
    case "native-format":
        return p.nativeEndpoint(protocol), nil
    default:
        return "", fmt.Errorf("unsupported format: %s", format)
    }
}
```

The architecture provides complete separation between API formats (handled by capabilities) and service integration (handled by providers), enabling maximum flexibility for supporting diverse LLM services and API standards.
