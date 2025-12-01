# Architecture

This document describes the current architecture and implementation patterns of the Go Agents library.

## Package Structure

```
pkg/
├── config/              # Configuration management and loading
│   ├── agent.go         # Agent configuration structure (flat: client, provider, model as peers)
│   ├── client.go        # Client and retry configuration (HTTP settings only)
│   ├── duration.go      # Custom Duration type with human-readable strings
│   ├── model.go         # Model configuration with protocol options
│   ├── options.go       # Option extraction and validation utilities
│   └── provider.go      # Provider configuration structures
├── protocol/            # Protocol types and message structures
│   ├── protocol.go      # Protocol constants and type definitions
│   └── message.go       # Message structures
├── response/            # Response parsing and types
│   ├── chat.go          # Chat protocol response types
│   ├── embeddings.go    # Embeddings protocol response types
│   ├── streaming.go     # Streaming chunk types
│   └── tools.go         # Tools protocol response types
├── model/               # Model runtime type
│   └── model.go         # Model type bridging config to runtime
├── providers/           # Provider implementations for different LLM services
│   ├── provider.go      # Provider interface definition
│   ├── base.go          # BaseProvider with common functionality
│   ├── registry.go      # Provider registry and initialization
│   ├── azure.go         # Azure AI Foundry provider implementation
│   └── ollama.go        # Ollama provider implementation
├── request/             # Request interface and protocol-specific request types
│   ├── interface.go     # Request interface definition
│   ├── chat.go          # ChatRequest implementation
│   ├── vision.go        # VisionRequest implementation
│   ├── tools.go         # ToolsRequest implementation
│   └── embeddings.go    # EmbeddingsRequest implementation
├── client/              # Client layer orchestrating requests across providers
│   ├── client.go        # Client interface and implementation
│   └── retry.go         # Exponential backoff retry logic with jitter
├── agent/               # High-level agent orchestration
│   ├── agent.go         # Agent interface and implementation
│   └── tools.go         # Tool definition types
└── mock/                # Mock implementations for testing
    ├── doc.go           # Package documentation
    ├── agent.go         # MockAgent implementation
    ├── client.go        # MockClient implementation
    ├── provider.go      # MockProvider implementation
    └── helpers.go       # Convenience constructors
```

## Core Components

### Protocol System

The protocol system defines the communication contracts for different LLM interactions. Protocols are the primary abstraction for capability identification.

```go
type Protocol string

const (
    Chat       Protocol = "chat"        // Text completion interactions
    Vision     Protocol = "vision"      // Image analysis with text
    Tools      Protocol = "tools"       // Function calling capabilities
    Embeddings Protocol = "embeddings"  // Vector embedding generation
)
```

**Protocol Methods**:
```go
// IsValid checks if protocol string is recognized
func (p Protocol) IsValid() bool

// SupportsStreaming indicates if protocol supports streaming responses
func (p Protocol) SupportsStreaming() bool
```

**Message Structure**: Supports both simple text and structured content for protocols like vision:
```go
type Message struct {
    Role    string `json:"role"`
    Content any    `json:"content"`  // string for text, []map[string]any for structured
}
```

**Protocol-Specific Request Types**: Each protocol has its own request type in `pkg/request` implementing the Request interface:
```go
// Request interface in pkg/request
type Request interface {
    Protocol() protocol.Protocol
    Headers() map[string]string
    Marshal() ([]byte, error)
    Provider() providers.Provider
    Model() *model.Model
}

// ChatRequest encapsulates chat protocol requests
type ChatRequest struct {
    messages []protocol.Message
    options  map[string]any
    provider providers.Provider
    model    *model.Model
}

// VisionRequest encapsulates vision protocol requests
type VisionRequest struct {
    messages      []protocol.Message
    images        []string
    visionOptions map[string]any
    options       map[string]any
    provider      providers.Provider
    model         *model.Model
}

// ToolsRequest encapsulates tools protocol requests
type ToolsRequest struct {
    Messages []Message
    Tools    []ToolDefinition        // Provider-agnostic tool definitions
    Options  map[string]any
}

// Embeddings protocol request - no messages, just input text
type EmbeddingsRequest struct {
    Input   any                      // string or []string for batch
    Options map[string]any
}
```

**Design Rationale**: Protocol-specific request types separate protocol input data (images, tools, input text) from model configuration options (temperature, max_tokens). This enables:
- **Multi-provider support**: Providers can transform protocol data to their specific formats
- **Type safety**: Compile-time checking for protocol-specific fields
- **Clear separation**: Protocol inputs vs. configuration options
- **Extensibility**: Easy to add provider-specific transformations

**Tools Format Decision**: `ToolsRequest.Marshal()` wraps tool definitions in OpenAI's format (`{"type": "function", "function": {...}}`) by default. This decision was made because:
- **Industry Standard**: OpenAI's format is used by the majority of providers (OpenAI, Azure, Ollama)
- **Immediate Compatibility**: Works out-of-the-box with most LLM services
- **Future Extensibility**: Providers requiring different formats (Anthropic, Google) can transform from this standard representation in their `PrepareRequest()` implementation

The library maintains provider-agnostic `ToolDefinition` types in the public API while handling provider-specific formatting internally.

**Protocol-Specific Responses**: Different protocols return specialized response types:
```go
type ChatResponse struct {
    Model   string
    Choices []struct {
        Index        int
        Message      Message
        FinishReason string
    }
    Usage *TokenUsage
}

type ToolsResponse struct {
    Model   string
    Choices []struct {
        Index   int
        Message struct {
            Role      string
            Content   string
            ToolCalls []ToolCall
        }
        FinishReason string
    }
    Usage *TokenUsage
}

type EmbeddingsResponse struct {
    Object string
    Model  string
    Data   []struct {
        Embedding []float64
        Index     int
        Object    string
    }
    Usage *TokenUsage
}
```

**Streaming Support**: Protocols that support streaming use a unified chunk structure:
```go
type StreamingChunk struct {
    ID      string
    Object  string
    Created int64
    Model   string
    Choices []struct {
        Index int
        Delta struct {
            Role    string
            Content string
        }
        FinishReason *string
    }
    Error error
}

// Content extracts the incremental content from the first choice delta
func (c *StreamingChunk) Content() string
```

### Model System

Models store the model name and protocol-specific options from configuration:

```go
type Model struct {
    Name    string
    Options map[Protocol]map[string]any  // Protocol-specific options from config
}
```

**Design Philosophy**: Options are merged at the agent layer, combining model's configured options with runtime overrides. The `Options` map stores protocol-specific configurations from JSON files, which serve as defaults that can be overridden at runtime.

- **Agent layer**: Merges model's configured protocol options with runtime options, adds model name
- **Runtime override**: Call-time options override model defaults
- **Protocol-specific extraction**: Vision protocol extracts `vision_options` from merged options
- **Configuration**: Options from config files provide baseline behavior

**Rationale**: Option merging allows model configurations to define sensible defaults while enabling runtime customization per request. This balances configuration convenience with runtime flexibility.

**Model Creation**: Convert configuration to runtime model:
```go
func FromConfig(cfg *config.ModelConfig) *Model {
    model := &Model{
        Name:    cfg.Name,
        Options: make(map[Protocol]map[string]any),
    }

    // Convert string keys to Protocol constants
    for key, opts := range cfg.Capabilities {
        protocol := Protocol(key)
        if protocol.IsValid() {
            model.Options[protocol] = opts
        }
    }

    return model
}
```

### Configuration Option Merging

Agent methods merge model's configured protocol options with runtime options, providing baseline defaults that can be overridden per request.

**Merging Pattern**: All agent protocol methods follow this standard pattern:

```go
func (a *agent) Chat(ctx context.Context, prompt string, opts ...map[string]any) (*types.ChatResponse, error) {
    messages := a.initMessages(prompt)

    // 1. Start with model's configured protocol options
    options := make(map[string]any)
    if modelOpts := a.client.Model().Options[types.Chat]; modelOpts != nil {
        maps.Copy(options, modelOpts)
    }

    // 2. Merge/override with runtime opts
    if len(opts) > 0 && opts[0] != nil {
        maps.Copy(options, opts[0])
    }

    // 3. Add model name
    options["model"] = a.client.Model().Name

    request := &types.ChatRequest{
        Messages: messages,
        Options:  options,
    }

    result, err := a.client.ExecuteProtocol(ctx, request)
    // ... rest of method
}
```

**Vision Protocol Special Handling**: Vision methods extract `vision_options` after merging:

```go
func (a *agent) Vision(ctx context.Context, prompt string, images []string, opts ...map[string]any) (*types.ChatResponse, error) {
    messages := a.initMessages(prompt)

    // 1. Start with model's configured vision options
    options := make(map[string]any)
    if modelOpts := a.client.Model().Options[types.Vision]; modelOpts != nil {
        maps.Copy(options, modelOpts)
    }

    // 2. Merge/override with runtime opts
    if len(opts) > 0 && opts[0] != nil {
        maps.Copy(options, opts[0])
    }

    // 3. Extract vision_options (protocol-specific options)
    var visionOptions map[string]any
    if vOpts, exists := options["vision_options"]; exists {
        if vOptsMap, ok := vOpts.(map[string]any); ok {
            visionOptions = vOptsMap
            delete(options, "vision_options")  // Remove from main options
        }
    }

    // 4. Add model name
    options["model"] = a.client.Model().Name

    request := &types.VisionRequest{
        Messages:      messages,
        Images:        images,
        VisionOptions: visionOptions,  // Separate field for vision-specific options
        Options:       options,         // Model configuration options
    }

    result, err := a.client.ExecuteProtocol(ctx, request)
    // ... rest of method
}
```

**Merging Behavior**:

1. **Base Options**: Starts with model's configured protocol options from `Model.Options[protocol]`
2. **Runtime Override**: Runtime options override matching keys from configuration
3. **Protocol-Specific Extraction**: Vision extracts `vision_options` nested map
4. **Model Name**: Always added to ensure request includes target model

**Example Merging Flow**:

```
Configuration: {"vision": {"max_tokens": 4096, "temperature": 0.7, "vision_options": {"detail": "high"}}}
  ↓
Runtime call: Vision(ctx, prompt, images, map[string]any{"temperature": 0.9})
  ↓
After merging: {"max_tokens": 4096, "temperature": 0.9, "vision_options": {"detail": "high"}}
  ↓
After extraction:
  - VisionOptions: {"detail": "high"}
  - Options: {"max_tokens": 4096, "temperature": 0.9, "model": "gpt-4o"}
```

**Why Vision Needs Special Handling**:

Vision's `detail` parameter controls image rendering behavior (e.g., "low", "high", "auto"), which is protocol-specific rather than a model inference parameter. Other protocols like tools don't need this separation because their parameters (e.g., `tool_choice`) are regular model options that apply to inference behavior.

### Provider System

Providers implement LLM service integrations and handle protocol routing:

```go
type Provider interface {
    Name() string
    BaseURL() string
    Model() *types.Model

    // Request preparation
    GetEndpoint(protocol types.Protocol) (string, error)
    PrepareRequest(ctx context.Context, request types.ProtocolRequest) (*Request, error)
    PrepareStreamRequest(ctx context.Context, request types.ProtocolRequest) (*Request, error)

    // Response processing
    ProcessResponse(resp *http.Response, protocol types.Protocol) (any, error)
    ProcessStreamResponse(ctx context.Context, resp *http.Response, protocol types.Protocol) (<-chan any, error)

    // Authentication
    SetHeaders(req *http.Request)
}
```

**Provider Responsibilities**:
- **Endpoint Mapping**: Map protocols to provider-specific API endpoints
- **Request Transformation**: Format requests for provider API
- **Authentication**: Handle provider-specific authentication methods
- **Response Parsing**: Parse provider responses using protocol-specific parsers

**Implemented Providers**:
- **Ollama**: OpenAI-compatible endpoints via `/v1/*`
- **Azure**: Azure AI Foundry with OpenAI format support

**Request Structure**:
```go
type Request struct {
    URL     string
    Headers map[string]string
    Body    []byte
}
```

### Client Layer

The client layer orchestrates request routing, retry logic, and execution:

```go
type Client interface {
    Provider() providers.Provider
    Model() *types.Model
    HTTPClient() *http.Client
    IsHealthy() bool

    ExecuteProtocol(ctx context.Context, request types.ProtocolRequest) (any, error)
    ExecuteProtocolStream(ctx context.Context, request types.ProtocolRequest) (<-chan *types.StreamingChunk, error)
}
```

**Request Flow**:
1. **Protocol Request**: Client receives protocol-specific request from agent
2. **Request Preparation**: Provider marshals request and formats for its API
3. **Retry Logic**: Exponential backoff with jitter for retryable errors
4. **HTTP Execution**: Send request with provider authentication
5. **Response Parsing**: Parse response using protocol-specific parsers
6. **Health Tracking**: Update client health status based on results

**Note**: Agent methods merge model's configured protocol options with runtime options before creating protocol requests. The client receives fully-formed requests with merged options and routes them to the appropriate provider.

**Retry Configuration**:
```go
type RetryConfig struct {
    MaxRetries        int
    InitialBackoff    Duration
    MaxBackoff        Duration
    BackoffMultiplier float64
    Jitter            bool
}
```

**Retry Logic** (`client/retry.go`):
- Exponential backoff: delay = initialBackoff * (multiplier ^ attempt)
- Jitter: randomize delay by ±25% to prevent thundering herd
- Retryable errors: HTTP 429, 502, 503, 504, network errors, DNS errors
- Non-retryable: context cancellation, context deadline, HTTP 4xx (except 429)

### Agent System

Agents provide high-level orchestration with protocol-specific methods:

```go
type Agent interface {
    ID() string
    Client() client.Client
    Provider() providers.Provider
    Model() *types.Model

    Chat(ctx context.Context, prompt string, opts ...map[string]any) (*types.ChatResponse, error)
    ChatStream(ctx context.Context, prompt string, opts ...map[string]any) (<-chan *types.StreamingChunk, error)

    Vision(ctx context.Context, prompt string, images []string, opts ...map[string]any) (*types.ChatResponse, error)
    VisionStream(ctx context.Context, prompt string, images []string, opts ...map[string]any) (<-chan *types.StreamingChunk, error)

    Tools(ctx context.Context, prompt string, tools []Tool, opts ...map[string]any) (*types.ToolsResponse, error)

    Embed(ctx context.Context, input string, opts ...map[string]any) (*types.EmbeddingsResponse, error)
}
```

**Agent Responsibilities**:
- **Message Initialization**: Create message arrays with system prompt injection
- **Protocol Execution**: Route to client's ExecuteProtocol methods
- **Response Type Assertion**: Ensure correct response type for each protocol
- **Streaming Management**: Handle streaming channels for supported protocols

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

## Configuration System

### Agent Configuration

```json
{
  "name": "agent-name",
  "system_prompt": "System instructions for the agent",
  "client": {
    "timeout": "24s",
    "retry": {
      "max_retries": 3,
      "initial_backoff": "1s",
      "max_backoff": "30s",
      "backoff_multiplier": 2.0,
      "jitter": true
    },
    "connection_pool_size": 10,
    "connection_timeout": "9s"
  },
  "provider": {
    "name": "ollama",
    "base_url": "http://localhost:11434"
  },
  "model": {
    "name": "llama3.2:3b",
    "capabilities": {
      "chat": {
        "max_tokens": 4096,
        "temperature": 0.7,
        "top_p": 0.95
      },
      "tools": {
        "max_tokens": 4096,
        "temperature": 0.7,
        "tool_choice": "auto"
      }
    }
  }
}
```

**Configuration Structure** (Flattened):
- `AgentConfig`: Top-level with `client`, `provider`, and `model` as peers
- `ClientConfig`: HTTP client settings and retry configuration
- `ProviderConfig`: Provider name, base URL, and provider-specific options
- `ModelConfig`: Model name and protocol-specific capabilities
- `RetryConfig`: Retry behavior configuration

**Duration Format**: Supports human-readable strings ("24s", "1m", "2h") or numeric nanoseconds:
```go
type Duration time.Duration

func (d *Duration) UnmarshalJSON(data []byte) error {
    // Try parsing as duration string first ("24s")
    // Fall back to numeric nanoseconds
}
```

### Protocol Configuration

Each protocol is configured independently with its own options:

**Per-Protocol Configuration**:
```json
"capabilities": {
  "chat": {
    "temperature": 0.7,
    "top_p": 0.95,
    "max_tokens": 4096
  },
  "vision": {
    "temperature": 0.7,
    "max_tokens": 4096,
    "vision_options": {
      "detail": "high"
    }
  },
  "tools": {
    "temperature": 0.7,
    "tool_choice": "auto"
  }
}
```

**Benefits**:
- **Option Isolation**: Each protocol has its own options (no conflicts)
- **Selective Support**: Models declare only supported protocols
- **Runtime Merging**: Request options override model defaults
- **Protocol-Specific Options**: Vision uses `vision_options` nested map for protocol-specific parameters like `detail`

### Configuration vs Domain Types

**Separation Principle**: Configuration structures use string keys for JSON serialization, while domain types use Protocol constants for type safety.

**Configuration Type** (`config.ModelConfig`):
```go
type ModelConfig struct {
    Name         string
    Capabilities map[string]map[string]any  // String keys for JSON
}
```

**Domain Type** (`types.Model`):
```go
type Model struct {
    Name    string
    Options map[Protocol]map[string]any  // Protocol constants for type safety
}
```

**Conversion** (`types.FromConfig`):
```go
func FromConfig(cfg *config.ModelConfig) *Model {
    model := &Model{
        Name:    cfg.Name,
        Options: make(map[Protocol]map[string]any),
    }

    for key, opts := range cfg.Capabilities {
        protocol := Protocol(key)
        if protocol.IsValid() {
            model.Options[protocol] = opts
        }
    }

    return model
}
```

## Data Flow

### Standard Request Flow

```
Agent.Chat(prompt, options)
  ↓
Agent creates ChatRequest{Messages, Options}  // Options includes model name
  ↓
Client.ExecuteProtocol(request)
  ↓
Provider.PrepareRequest(request)  // Marshal and format for API
  ↓
HTTP Request with Retry Logic
  ↓
Provider.ProcessResponse(response, request.GetProtocol())
  ↓
types.ParseChatResponse(body)  // Parse protocol-specific response
  ↓
Agent returns *types.ChatResponse
```

### Streaming Request Flow

```
Agent.ChatStream(prompt, options)
  ↓
Agent creates ChatRequest{Messages, Options}
options["stream"] = true  // Add streaming flag
  ↓
Client.ExecuteProtocolStream(request)
  ↓
request.GetProtocol().SupportsStreaming() check
  ↓
Provider.PrepareStreamRequest(request)  // Marshal with streaming headers
  ↓
HTTP Streaming Request (no retry)
  ↓
Provider.ProcessStreamResponse(response, request.GetProtocol())
  ↓
Channel of types.StreamingChunk
  ↓
Agent streams chunks to caller
```

### Option Merging Flow

```
Configuration: {"chat": {"temperature": 0.7, "max_tokens": 4096}}
  ↓
Model created with Options[types.Chat] = {"temperature": 0.7, "max_tokens": 4096}
  ↓
Agent.Chat(prompt, {"temperature": 0.9})
  ↓
Agent merges: config options + runtime override + model name
  = {"temperature": 0.9, "max_tokens": 4096, "model": "llama3.2:3b"}
  ↓
ChatRequest{Messages, Options: {"temperature": 0.9, "max_tokens": 4096, "model": "..."}}
  ↓
Provider.PrepareRequest marshals request with merged options
  ↓
HTTP request body: {"messages": [...], "temperature": 0.9, "max_tokens": 4096, "model": "..."}
```

**Note**: Agent methods merge model's configured protocol options with runtime options. Runtime options override configuration values for matching keys. The agent adds the model name last to ensure it's always included.

## Design Patterns

### Protocol-Centric Architecture

Protocols are the primary abstraction, eliminating the need for a separate capability layer:

**Benefits**:
- Simpler architecture with fewer layers
- Direct protocol-to-parser mapping
- Clear protocol support via `Protocol.IsValid()` and `Protocol.SupportsStreaming()`
- Protocol constants prevent typos and enable compile-time validation

### Configuration Lifecycle

Configuration only exists during initialization:

```go
// Load configuration
cfg, _ := config.LoadAgentConfig("config.json")

// Create agent (config transforms to domain types)
agent, _ := agent.New(cfg)

// Runtime uses domain types (types.Model, types.Protocol)
// Configuration is no longer referenced
```

**Rationale**: Prevents configuration infrastructure from persisting too deeply into package layers, maintaining clear separation between initialization and runtime.

### Interface-Based Layer Interconnection

Layers communicate through interfaces, not concrete types:

```go
// Agent depends on client interface
type Agent interface {
    Client() client.Client  // Returns interface, not concrete type
}

// Client depends on provider interface
type Client interface {
    Provider() providers.Provider  // Returns interface
}
```

**Benefits**:
- Loose coupling between layers
- Testing through mocks
- Multiple implementations possible
- Clear contracts between components

### Retry Pattern

Intelligent retry with exponential backoff and jitter:

```go
func isRetryableError(err error) bool {
    // Context cancellation/deadline: not retryable
    if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
        return false
    }

    // HTTP status codes
    var httpErr *HTTPStatusError
    if errors.As(err, &httpErr) {
        return httpErr.StatusCode == 429 ||  // Rate limit
               httpErr.StatusCode == 502 ||  // Bad gateway
               httpErr.StatusCode == 503 ||  // Service unavailable
               httpErr.StatusCode == 504     // Gateway timeout
    }

    // Network and DNS errors: retryable
    return true
}
```

**Retry Strategy**:
```go
func doWithRetry[T any](
    ctx context.Context,
    cfg config.RetryConfig,
    operation func() (T, error),
) (T, error) {
    backoff := cfg.InitialBackoff.ToDuration()

    for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
        result, err := operation()

        if err == nil || !isRetryableError(err) {
            return result, err
        }

        if attempt < cfg.MaxRetries {
            delay := calculateDelay(backoff, cfg)
            time.Sleep(delay)
            backoff *= time.Duration(cfg.BackoffMultiplier)
        }
    }
}
```

## Extension Points

### Adding New Providers

1. Implement `Provider` interface with endpoint mapping and authentication
2. Implement protocol-specific parsers in `ProcessResponse`
3. Register in provider registry
4. Configure in JSON

Example:
```go
type CustomProvider struct {
    *BaseProvider
}

func (p *CustomProvider) GetEndpoint(protocol types.Protocol) (string, error) {
    endpoints := map[types.Protocol]string{
        types.Chat:       "/v1/chat/completions",
        types.Vision:     "/v1/chat/completions",
        types.Tools:      "/v1/chat/completions",
        types.Embeddings: "/v1/embeddings",
    }
    endpoint, ok := endpoints[protocol]
    if !ok {
        return "", fmt.Errorf("protocol %s not supported", protocol)
    }
    return p.BaseURL() + endpoint, nil
}

func (p *CustomProvider) ProcessResponse(resp *http.Response, protocol types.Protocol) (any, error) {
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    return types.ParseResponse(protocol, body)
}
```

### Adding New Protocols

1. Add protocol constant to `types/protocol.go`
2. Implement request/response types in `types/<protocol>.go`
3. Implement `Parse<Protocol>Response` function
4. If streaming supported, implement `Parse<Protocol>StreamChunk`
5. Update `Protocol.IsValid()` and `Protocol.SupportsStreaming()`
6. Update providers to support new protocol endpoint

Example:
```go
// types/protocol.go
const (
    Chat       Protocol = "chat"
    Vision     Protocol = "vision"
    Tools      Protocol = "tools"
    Embeddings Protocol = "embeddings"
    Audio      Protocol = "audio"  // New protocol
)

// types/audio.go
type AudioResponse struct {
    Model   string
    Audio   []byte
    Format  string
    Usage   *TokenUsage
}

func ParseAudioResponse(data []byte) (*AudioResponse, error) {
    var resp AudioResponse
    if err := json.Unmarshal(data, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

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
│   ├── client_test.go
│   └── agent_test.go
├── protocol/
│   └── protocol_test.go
├── response/
│   └── response_test.go
├── providers/
│   ├── base_test.go
│   ├── ollama_test.go
│   ├── azure_test.go
│   └── registry_test.go
├── client/
│   └── client_test.go
├── agent/
│   └── agent_test.go
├── mock/
│   ├── agent_test.go
│   ├── client_test.go
│   └── provider_test.go
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
func TestProtocol_IsValid(t *testing.T) {
    tests := []struct {
        name     string
        protocol protocol.Protocol
        expected bool
    }{
        {name: "chat", protocol: protocol.Chat, expected: true},
        {name: "invalid", protocol: protocol.Protocol("invalid"), expected: false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := tt.protocol.IsValid(); got != tt.expected {
                t.Errorf("got %v, want %v", got, tt.expected)
            }
        })
    }
}
```

**HTTP Mocking**: Use `httptest.Server` for mocking provider responses:

```go
func TestClient_Execute_Chat(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        resp := response.ChatResponse{
            Model: "test-model",
        }
        resp.Choices = append(resp.Choices, struct {
            Index        int
            Message      protocol.Message
            FinishReason string
        }{
            Index:   0,
            Message: protocol.NewMessage("assistant", "Test response"),
        })
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(resp)
    }))
    defer server.Close()

    // Test with server.URL
}
```

### Coverage Requirements

**Minimum Coverage**: 80% across all packages

**Critical Path Coverage**: 100% for:
- Request/response parsing (types package)
- Protocol request marshaling (types.*Request.Marshal)
- Configuration validation (config package)
- Protocol routing (client package)

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

# Test streaming
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.ollama.json \
  -prompt "Test prompt" \
  -stream

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
- When adding new protocols
- To verify configuration changes

### Mock Package

**Purpose**: The `pkg/mock` package provides configurable mock implementations of all core interfaces for testing code that depends on go-agents.

**Package Structure**:
```
pkg/mock/
├── doc.go           # Package documentation
├── agent.go         # MockAgent implementation
├── client.go        # MockClient implementation
├── provider.go      # MockProvider implementation
└── helpers.go       # Convenience constructors
```

**Mock Types**:

1. **MockAgent** (`agent.go`)
   - Implements: `agent.Agent`
   - Configurable responses for: Chat, Vision, Tools, Embeddings
   - Streaming support for Chat and Vision
   - Options: `WithID`, `WithChatResponse`, `WithVisionResponse`, `WithToolsResponse`, `WithEmbeddingsResponse`, `WithStreamChunks`

2. **MockClient** (`client.go`)
   - Implements: `client.Client`
   - Configurable protocol execution and streaming
   - Health status management
   - Options: `WithExecuteResponse`, `WithStreamResponse`, `WithHealthy`, `WithHTTPClient`

3. **MockProvider** (`provider.go`)
   - Implements: `providers.Provider`
   - Custom endpoint mapping per protocol
   - Request preparation and response processing
   - Options: `WithBaseURL`, `WithEndpointMapping`, `WithPrepareResponse`, `WithProcessResponse`

**Helper Constructors** (`helpers.go`):

For common testing scenarios without manual configuration:

```go
// Simple chat agent
agent := mock.NewSimpleChatAgent("id", "response text")

// Streaming chat agent
agent := mock.NewStreamingChatAgent("id", []string{"chunk1", "chunk2"})

// Tools agent
agent := mock.NewToolsAgent("id", []response.ToolCall{...})

// Embeddings agent
agent := mock.NewEmbeddingsAgent("id", []float64{0.1, 0.2, 0.3})

// Multi-protocol agent
agent := mock.NewMultiProtocolAgent("id")

// Failing agent (for error handling tests)
agent := mock.NewFailingAgent("id", errors.New("test error"))
```

**Usage Pattern**:

The option pattern allows precise control over mock behavior:

```go
// Configure specific behaviors
mockAgent := mock.NewMockAgent(
    mock.WithID("custom-id"),
    mock.WithChatResponse(&response.ChatResponse{...}, nil),
    mock.WithStreamChunks([]*response.StreamingChunk{...}, nil),
)

// Test error handling
failingAgent := mock.NewMockAgent(
    mock.WithID("failing-agent"),
    mock.WithChatResponse(nil, errors.New("connection failed")),
)
```

**Use Cases**:
- Testing orchestration systems without live LLM calls
- Testing error handling and failure scenarios
- Testing multi-agent coordination
- Testing protocol-specific behavior
- Unit testing supplemental packages that extend go-agents
