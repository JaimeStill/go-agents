# Architecture

This document describes the current architecture and implementation patterns of the Go Agents library.

## Package Structure

```
pkg/
├── config/              # Configuration management and loading
│   ├── agent.go         # Agent configuration structure
│   ├── duration.go      # Custom Duration type with human-readable strings
│   ├── model.go         # Model configuration with composable capabilities
│   ├── provider.go      # Provider configuration structures
│   └── transport.go     # Transport layer configuration
├── protocols/           # Protocol definitions and message structures
│   └── protocol.go      # Protocol types, Message, Request, Response structures
├── capabilities/        # Protocol-specific capability implementations
│   ├── capability.go    # Core capability interface and standard implementations
│   ├── registry.go      # Thread-safe capability format registry
│   ├── init.go          # Capability format registrations
│   ├── chat.go          # Chat protocol capability
│   ├── vision.go        # Vision protocol capability with structured content
│   ├── tools.go         # Tools protocol capability
│   └── embeddings.go    # Embeddings protocol capability
├── models/              # Model abstraction and protocol handler composition
│   ├── model.go         # Model interface and implementation
│   └── handler.go       # ProtocolHandler for stateful capability management
├── providers/           # Provider implementations for different LLM services
│   ├── provider.go      # Provider interface definition
│   ├── base.go          # BaseProvider with common functionality
│   ├── registry.go      # Provider registry and initialization
│   ├── azure.go         # Azure AI Foundry provider implementation
│   └── ollama.go        # Ollama provider implementation
├── transport/           # Transport layer orchestrating requests across providers
│   └── client.go        # Transport client interface and implementation
└── agent/               # High-level agent orchestration
    └── agent.go         # Agent interface and implementation
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

**Protocol-Specific Responses**: Different protocols return specialized response types:
```go
type ChatResponse struct {
    Choices []struct {
        Message Message
    }
    Usage *TokenUsage
}

type ToolsResponse struct {
    Choices []struct {
        Message struct {
            Role      string
            Content   string
            ToolCalls []ToolCall
        }
    }
    Usage *TokenUsage
}

type EmbeddingsResponse struct {
    Data []struct {
        Embedding []float64
        Index     int
    }
    Usage *TokenUsage
}
```

### Capability System

Capabilities implement protocol-specific behavior and validation for different API formats:

```go
type Capability interface {
    Name() string
    Protocol() protocols.Protocol
    Options() []CapabilityOption
    ValidateOptions(options map[string]any) error
    ProcessOptions(options map[string]any) (map[string]any, error)
    CreateRequest(req *CapabilityRequest, model string) (*protocols.Request, error)
    ParseResponse(data []byte) (any, error)
    SupportsStreaming() bool
}

type StreamingCapability interface {
    Capability
    CreateStreamingRequest(req *CapabilityRequest, model string) (*protocols.Request, error)
    ParseStreamingChunk(data []byte) (*protocols.StreamingChunk, error)
    IsStreamComplete(data string) bool
}
```

**Capability Registry**: Thread-safe registration system for capability formats:
```go
func RegisterFormat(name string, factory CapabilityFactory)
func GetFormat(name string) (Capability, error)
```

**Registered Capability Formats**:
- **openai-chat**: Standard OpenAI chat completions (supports temperature, top_p, etc.)
- **openai-vision**: OpenAI vision with structured content
- **openai-tools**: OpenAI function calling (non-streaming)
- **openai-embeddings**: OpenAI embeddings generation
- **openai-reasoning**: OpenAI reasoning models (restricted parameters, max_completion_tokens only)

**Option Management**: Each capability defines supported options with validation:
```go
type CapabilityOption struct {
    Option       string `json:"option"`
    Required     bool   `json:"required"`
    DefaultValue any    `json:"default_value"`
}
```

### Model System

Models use composable capabilities with protocol-specific handlers:

```go
type Model interface {
    Name() string
    SupportsProtocol(p protocols.Protocol) bool
    GetCapability(p protocols.Protocol) (capabilities.Capability, error)
    GetProtocolOptions(p protocols.Protocol) map[string]any
    UpdateProtocolOptions(p protocols.Protocol, options map[string]any) error
    MergeRequestOptions(p protocols.Protocol, requestOptions map[string]any) map[string]any
}
```

**ProtocolHandler Pattern**: Manages stateful protocol configuration:
```go
type ProtocolHandler struct {
    capability capabilities.Capability  // Stateless behavior
    options    map[string]any            // Stateful configuration
}
```

**Model Implementation**: Explicit protocol fields with handlers:
```go
type model struct {
    name       string
    chat       *ProtocolHandler  // nil if not configured
    vision     *ProtocolHandler  // nil if not configured
    tools      *ProtocolHandler  // nil if not configured
    embeddings *ProtocolHandler  // nil if not configured
}
```

**Option Management**:
- **Configuration Options**: Set at model creation, persisted across requests
- **Request Options**: Passed per-request, merged with configuration options
- **Validation Timing**: Options validated after merge in transport layer

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
    ID() string
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

#### Agent Identification

Each agent has a unique identifier assigned at creation time that remains stable throughout its lifetime.

**ID Generation**:
- UUIDv7 format: time-sortable with nanosecond precision
- Auto-generated during agent creation
- Collision-resistant across distributed systems
- Thread-safe for concurrent access

**ID Guarantees**:
- **Uniqueness**: Each agent receives a globally unique identifier
- **Stability**: ID never changes after creation
- **Thread-Safety**: Safe to call `ID()` from multiple goroutines
- **Map Key Safety**: Safe to use as map keys in registries and routing tables

**Orchestration Use Cases**:
- **Hub Registration**: Register agents in multi-hub coordination systems
- **Message Routing**: Route messages to specific agents using their IDs
- **Lifecycle Tracking**: Track agent creation, activity, and destruction
- **Distributed Tracing**: Correlate agent operations across service boundaries
- **Observability**: Aggregate metrics and logs by agent ID

**Example - Hub Registration**:
```go
// Create multiple agents
agent1, _ := agent.New(config)
agent2, _ := agent.New(config)

// Register in hub using IDs
hub.Register(agent1.ID(), agent1)
hub.Register(agent2.ID(), agent2)

// Route message to specific agent
hub.SendTo(agent1.ID(), message)
```

**Example - Distributed Tracing**:
```go
agent, _ := agent.New(config)

// Include agent ID in structured logs
log.Info("agent processing request",
    "agent_id", agent.ID(),
    "request_id", requestID)

// Correlate operations across service boundaries
ctx = context.WithValue(ctx, "agent_id", agent.ID())
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
    "timeout": "24s",
    "max_retries": 3,
    "retry_backoff_base": "1s",
    "connection_pool_size": 10,
    "connection_timeout": "9s"
  }
}
```

**Duration Format**: Supports human-readable strings ("24s", "1m", "2h") or numeric nanoseconds:
```go
type Duration time.Duration

func (d *Duration) UnmarshalJSON(data []byte) error {
    // Try parsing as duration string first ("24s")
    // Fall back to numeric nanoseconds
}
```

### Composable Capabilities

Each protocol is configured independently with its own capability format and options:

**Per-Protocol Configuration**:
```json
"capabilities": {
  "chat": {
    "format": "openai-chat",
    "options": {"temperature": 0.7, "top_p": 0.95}
  },
  "tools": {
    "format": "openai-tools",
    "options": {"tool_choice": "auto"}
  }
}
```

**Benefits**:
- **Option Isolation**: Each protocol has its own options (no conflicts)
- **Selective Support**: Models declare only supported protocols
- **Format Flexibility**: Different protocols can use different formats
- **Runtime Updates**: Protocol options can be updated on live agents

### Provider-Capability Compatibility

Providers work with any capability format that matches their API endpoints:

**Current Compatibility**:
- **Ollama Provider**: Supports OpenAI-format capabilities via `/v1/*` endpoints
- **Azure Provider**: Native OpenAI API compatibility

**Future Extensibility**:
- **Ollama Provider**: Could add native format support via `/api/*` endpoints
- **Anthropic Provider**: Would support Anthropic-format capabilities
- **Universal Provider**: Could detect and route to appropriate endpoints based on capability format

## Data Flow

### Standard Request Flow

```
Agent.Chat(prompt, options)
  ↓
Transport.ExecuteProtocol(protocol, messages, options)
  ↓
Model.MergeRequestOptions(protocol, options)  // Merge config + request options
  ↓
Capability.ValidateOptions(merged_options)    // Validate after merge
  ↓
Capability.CreateRequest(messages, merged_options)  // Format to API structure
  ↓
Provider.PrepareRequest(protocol, request)    // Route to endpoint + auth
  ↓
HTTP Request → Response
  ↓
Provider.ProcessResponse(response, capability)
  ↓
Capability.ParseResponse(data)  // Parse API-specific response
  ↓
Agent returns typed response
```

### Capability-Protocol Flow

```
Configuration: {"chat": {"format": "openai-chat", "options": {...}}}
  ↓
Model creates ProtocolHandler(OpenAIChatCapability, options)
  ↓
Request arrives with per-request options
  ↓
Merged options = config options + request options
  ↓
OpenAI Chat Capability formats to OpenAI API structure
  ↓
Ollama Provider routes to /v1/chat/completions endpoint
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

### Configuration-Driven Composition

Model capabilities are composed from registered formats:
- **Capability Configuration**: Each protocol specifies its format and options
- **Protocol Isolation**: Options are validated and managed per-protocol
- **Provider Routing**: Providers route requests based on protocol and endpoint compatibility
- **Runtime Flexibility**: Protocol options can be updated on live models

## Extension Points

### Adding New Providers

1. Implement `Provider` interface with endpoint mapping and authentication
2. Ensure compatibility with desired capability formats
3. Register in provider registry
4. Works with any capability format - just needs compatible endpoints

Example:
```go
type CustomProvider struct {
    *BaseProvider
}

func (p *CustomProvider) GetEndpoint(protocol protocols.Protocol) (string, error) {
    // Map protocols to provider-specific endpoints
    endpoints := map[protocols.Protocol]string{
        protocols.Chat:   "/v1/chat/completions",
        protocols.Vision: "/v1/chat/completions",
        protocols.Tools:  "/v1/chat/completions",
    }
    return p.BaseURL() + endpoints[protocol], nil
}
```

### Adding New Capability Formats

1. Implement `Capability` interface for the new format
2. Define protocol-specific options and validation
3. Register format in capability registry using `init()`
4. Use in model configuration

Example:
```go
func init() {
    RegisterFormat("anthropic-chat", func() Capability {
        return NewAnthropicChatCapability()
    })
}
```

Configuration:
```json
"capabilities": {
  "chat": {
    "format": "anthropic-chat",
    "options": {"max_tokens": 4096}
  }
}
```

### Multi-Protocol Configuration

Models can compose capabilities from different formats:

```json
"model": {
  "name": "multi-capability-model",
  "capabilities": {
    "chat": {
      "format": "openai-chat",
      "options": {"temperature": 0.7}
    },
    "tools": {
      "format": "custom-tools",
      "options": {"execution_mode": "sandbox"}
    },
    "embeddings": {
      "format": "openai-embeddings",
      "options": {"dimensions": 1536}
    }
  }
}
```

The architecture provides complete separation between API formats (handled by capabilities), protocol configuration (handled by models), and service integration (handled by providers), enabling maximum flexibility for supporting diverse LLM services and API standards.

## Testing Strategy

### Test Organization

Tests are organized in a separate `tests/` directory that mirrors the `pkg/` structure:

```
tests/
├── config/
│   ├── duration_test.go
│   ├── options_test.go
│   ├── model_test.go
│   ├── provider_test.go
│   ├── transport_test.go
│   └── agent_test.go
├── protocols/
│   └── protocol_test.go
├── capabilities/
│   └── ...
└── ...
```

**Rationale**: Separating tests from production code keeps the `pkg/` directory clean and focused on implementation.

### Black-Box Testing

All tests use black-box testing approach with `package <name>_test`:

```go
package config_test

import (
    "testing"
    "github.com/JaimeStill/go-agents/pkg/config"
)
```

**Benefits**:
- Tests validate the public API from a consumer perspective
- Cannot access unexported members, ensuring tests reflect real usage
- Encourages well-designed public interfaces
- Internal refactoring doesn't break tests
- Reduces test volume by focusing only on public functionality

### Test Patterns

**Table-Driven Tests**: Used for testing multiple scenarios with different inputs:

```go
func TestDuration_UnmarshalJSON(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected time.Duration
    }{
        {name: "seconds", input: `"24s"`, expected: 24 * time.Second},
        {name: "minutes", input: `"1m"`, expected: 1 * time.Minute},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

**HTTP Mocking**: Use `httptest.Server` for mocking provider responses:

```go
func TestProvider_Request(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // mock response
        json.NewEncoder(w).Encode(mockResponse)
    }))
    defer server.Close()

    // test with server.URL
}
```

### Coverage Requirements

**Minimum Coverage**: 80% across all packages

**Critical Path Coverage**: 100% for:
- Request/response parsing (protocols package)
- Configuration validation (config package)
- Protocol routing (transport package)
- Option merging and validation (models, transport)

**Coverage Commands**:
```bash
# Generate coverage for specific package
go test ./tests/config/... -coverprofile=coverage.out -coverpkg=./pkg/config/...

# Generate coverage for all packages
go test ./tests/... -coverprofile=coverage.out -coverpkg=./pkg/...

# View coverage summary
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

### Integration Validation

**No Automated Integration Tests**: The library does not include automated integration tests that require live LLM providers or credentials.

**Manual Validation**: Integration validation is performed manually using the `tools/prompt-agent` CLI utility:

```bash
# Test Ollama integration
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.ollama.json \
  -prompt "Test prompt"

# Test Azure integration
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.azure.json \
  -token $AZURE_API_KEY \
  -prompt "Test prompt"
```

**Validation Approach**:
- README examples serve as integration test cases
- All examples are executable via `tools/prompt-agent`
- If README examples execute successfully, integration works
- No credential management in test suite
- No live service dependencies in CI/CD

**When to Run Validation**:
- Before releases
- After provider-specific changes
- When adding new capability formats
- To verify configuration changes
