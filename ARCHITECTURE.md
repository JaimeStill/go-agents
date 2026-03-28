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
├── response/            # Provider-neutral response types
│   ├── doc.go           # Package documentation
│   ├── response.go      # Unified Response with content blocks
│   ├── content.go       # ContentBlock interface, TextBlock, ToolUseBlock
│   ├── streaming.go     # StreamingResponse with content blocks
│   ├── embeddings.go    # EmbeddingsResponse for vector embeddings
│   └── usage.go         # TokenUsage with input/output/total tokens
├── format/              # Format interface, data types, and implementations
│   ├── format.go        # Format interface definition
│   ├── data.go          # Protocol data types (ChatData, VisionData, etc.)
│   ├── openai.go        # OpenAI format implementation
│   ├── converse.go      # AWS Converse format implementation
│   └── registry.go      # Format registry and factory system
├── streaming/           # Streaming transport interface and implementations
│   ├── streaming.go     # StreamReader interface and StreamLine type
│   ├── sse.go           # Server-Sent Events (SSE) reader
│   └── eventstream.go   # AWS EventStream binary framing reader
├── identities/          # Managed identity and credential sources
│   ├── azure.go         # Azure managed identity via azidentity SDK
│   └── aws.go           # AWS credential sourcing and SigV4 request signing
├── model/               # Model runtime type
│   └── model.go         # Model type bridging config to runtime
├── providers/           # Provider implementations (transport and auth only)
│   ├── provider.go      # Provider interface definition
│   ├── base.go          # BaseProvider with common functionality
│   ├── registry.go      # Provider registry and initialization
│   ├── ollama.go        # Ollama provider implementation
│   ├── azure.go         # Azure AI Foundry provider implementation
│   └── bedrock.go       # AWS Bedrock provider implementation
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
│   └── agent.go         # Agent interface, implementation, and Tool type
└── mock/                # Mock implementations for testing
    ├── doc.go           # Package documentation
    ├── agent.go         # MockAgent implementation
    ├── client.go        # MockClient implementation
    ├── provider.go      # MockProvider implementation
    ├── format.go        # MockFormat implementation
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

### Response System

The response system provides provider-neutral types returned by all LLM protocol requests. A unified `Response` type uses typed content blocks rather than protocol-specific response structures.

**Content Blocks**:
```go
type ContentBlock interface {
    blockType() string
}

type TextBlock struct {
    Text string
}

type ToolUseBlock struct {
    ID    string
    Name  string
    Input map[string]any
}
```

**Unified Response**: A single `Response` type handles Chat, Vision, and Tools protocols:
```go
type Response struct {
    Role       string
    Content    []ContentBlock
    StopReason string
    Usage      *TokenUsage
}

// Text extracts all text content from the response.
func (r *Response) Text() string

// ToolCalls extracts all tool use blocks from the response.
func (r *Response) ToolCalls() []ToolUseBlock
```

**Streaming Response**: Partial responses delivered as streaming chunks using the same content block model:
```go
type StreamingResponse struct {
    Content    []ContentBlock
    StopReason string
    Usage      *TokenUsage
    Error      error
}

// Text extracts text content from the streaming chunk.
func (r *StreamingResponse) Text() string
```

**Embeddings Response**: Separate type for vector embedding results:
```go
type EmbeddingsResponse struct {
    Embeddings [][]float64
    Model      string
    Usage      *TokenUsage
}
```

**Token Usage**: Normalized across providers:
```go
type TokenUsage struct {
    InputTokens  int `json:"input_tokens"`
    OutputTokens int `json:"output_tokens"`
    TotalTokens  int `json:"total_tokens"`
}
```

### Format System

The format system decouples wire-format serialization from providers. Providers handle transport and authentication; formats handle marshaling requests and parsing responses into provider-neutral types.

**Format Interface**:
```go
type Format interface {
    Name() string
    Marshal(p protocol.Protocol, data any) ([]byte, error)
    Parse(p protocol.Protocol, body []byte) (any, error)
    ParseStreamChunk(p protocol.Protocol, data []byte) (*response.StreamingResponse, error)
}
```

**Format Data Types**: Protocol-specific data structures passed to `Marshal`:
```go
type ChatData struct {
    Model    string
    Messages []protocol.Message
    Options  map[string]any
}

type VisionData struct {
    Model         string
    Messages      []protocol.Message
    Images        []Image
    VisionOptions map[string]any
    Options       map[string]any
}

type ToolsData struct {
    Model    string
    Messages []protocol.Message
    Tools    []ToolDefinition
    Options  map[string]any
}

type EmbeddingsData struct {
    Model   string
    Input   any
    Options map[string]any
}
```

**Provider-Neutral Image Type**: Images are represented in their most natural form, avoiding encode/decode round-trips:
```go
type Image struct {
    Data   []byte  // Raw image bytes
    Format string  // Image format (e.g., "png", "jpeg")
    URL    string  // Image URL (alternative to Data+Format)
}
```

Each format handles final encoding: OpenAI builds base64 data URIs at marshal time, Converse uses raw bytes directly.

**Tool Definition**:
```go
type ToolDefinition struct {
    Name        string         `json:"name"`
    Description string         `json:"description"`
    Parameters  map[string]any `json:"parameters"`
}
```

**Format Implementations**:
- **OpenAI** (`openai.go`): Marshals to OpenAI-compatible JSON. Wraps tools in `{"type": "function", "function": {...}}` format. Parses `choices[0].message` responses. Handles SSE stream chunks with `choices[0].delta.content`.
- **Converse** (`converse.go`): Marshals to AWS Bedrock Converse API format. Separates system messages into a top-level `system` field. Wraps tools in `toolSpec` format. Parses `output.message.content` with content blocks. Handles EventStream chunks with `contentBlockDelta`. Does not support embeddings.

**Format Registry**: Parallel to the provider registry, formats are registered by name and created on demand:
```go
func Register(name string, factory Factory)
func Create(name string) (Format, error)
func ListFormats() []string

func init() {
    Register("openai", OpenAIFactory)
    Register("converse", ConverseFactory)
}
```

### Streaming System

The streaming system abstracts transport-level stream reading from format-level chunk parsing. Providers expose a `StreamReader` that converts raw HTTP response bodies into a channel of data lines.

**StreamReader Interface**:
```go
type StreamLine struct {
    Data []byte
    Done bool
    Err  error
}

type StreamReader interface {
    ReadStream(ctx context.Context, reader io.Reader) <-chan StreamLine
}
```

**SSE Reader** (`sse.go`): Reads Server-Sent Events (`text/event-stream`). Parses `data: ` prefixed lines, treats `data: [DONE]` as stream termination. Used by Ollama and Azure providers.

**EventStream Reader** (`eventstream.go`): Reads AWS binary EventStream frames (`application/vnd.amazon.eventstream`). Decodes frames using `aws-sdk-go-v2/aws/protocol/eventstream`, extracts `:event-type` headers, and wraps payloads as JSON envelopes keyed by event type. Used by the Bedrock provider.

**Integration**: The client calls `provider.Stream().ReadStream(ctx, resp.Body)` to get a `<-chan StreamLine`, then passes each line's data through `format.ParseStreamChunk()` to produce typed `StreamingResponse` values.

### Provider System

Providers handle transport and authentication for LLM services. They do not perform request marshaling or response parsing -- that is the format layer's responsibility.

**Provider Interface**:
```go
type Provider interface {
    Name() string
    BaseURL() string

    Endpoint(p protocol.Protocol) (string, error)
    Stream() streaming.StreamReader

    SetHeaders(ctx context.Context, req *http.Request) error

    PrepareRequest(ctx context.Context, p protocol.Protocol, body []byte, headers map[string]string) (*Request, error)
    PrepareStreamRequest(ctx context.Context, p protocol.Protocol, body []byte, headers map[string]string) (*Request, error)
}
```

**Provider Responsibilities**:
- **Endpoint Mapping**: Map protocols to provider-specific API endpoints
- **Stream Reader**: Expose the appropriate `StreamReader` for the provider's streaming transport
- **Authentication**: Handle provider-specific authentication (API keys, bearer tokens, managed identity, SigV4 signing)
- **Request Preparation**: Construct URLs and headers from pre-marshaled request bodies

**Request Structure**:
```go
type Request struct {
    URL     string
    Headers map[string]string
    Body    []byte
}
```

**Implemented Providers**:
- **Ollama**: OpenAI-compatible endpoints via `/v1/*`. SSE streaming. Optional bearer token or API key authentication.
- **Azure**: Azure AI Foundry with deployment-based routing and API version query parameters. SSE streaming. Supports API key, bearer token, and managed identity authentication.
- **Bedrock**: AWS Bedrock Converse API with model ID in URL path. EventStream binary framing. SigV4 request signing via `identities.AWSCredentialSource`. Supports default credential chain, static credentials, and named profiles.

**Provider Registry**:
```go
func Register(name string, factory Factory)
func Create(c *config.ProviderConfig) (Provider, error)
func ListProviders() []string

func init() {
    Register("ollama", NewOllama)
    Register("azure", NewAzure)
    Register("bedrock", NewBedrock)
}
```

### Model System

Models store the model name and protocol-specific options from configuration:

```go
type Model struct {
    Name    string
    Options map[protocol.Protocol]map[string]any
}
```

**Design Philosophy**: Options are merged at the agent layer, combining model's configured options with runtime overrides. The `Options` map stores protocol-specific configurations from JSON files, which serve as defaults that can be overridden at runtime.

**Model Creation**: Convert configuration to runtime model:
```go
func New(cfg *config.ModelConfig) *Model {
    model := &Model{
        Name:    cfg.Name,
        Options: make(map[protocol.Protocol]map[string]any),
    }

    for protocolName, options := range cfg.Capabilities {
        p := protocol.Protocol(protocolName)
        model.Options[p] = options
    }

    return model
}
```

### Request System

Requests encapsulate protocol-specific input data, carry references to provider, format, and model, and handle marshaling through the format layer.

**Request Interface**:
```go
type Request interface {
    Protocol() protocol.Protocol
    Headers() map[string]string
    Marshal() ([]byte, error)
    Provider() providers.Provider
    Format() format.Format
    Model() *model.Model
}
```

**Request Constructors**: Each protocol has a constructor that takes provider, format, model, and protocol-specific data:
```go
func NewChat(p providers.Provider, f format.Format, m *model.Model, messages []protocol.Message, opts map[string]any) *ChatRequest
func NewVision(p providers.Provider, f format.Format, m *model.Model, messages []protocol.Message, images []format.Image, visionOpts map[string]any, opts map[string]any) *VisionRequest
func NewTools(p providers.Provider, f format.Format, m *model.Model, messages []protocol.Message, tools []format.ToolDefinition, opts map[string]any) *ToolsRequest
func NewEmbeddings(p providers.Provider, f format.Format, m *model.Model, input any, opts map[string]any) *EmbeddingsRequest
```

**Marshal Delegation**: Requests delegate marshaling to the format layer:
```go
func (r *ChatRequest) Marshal() ([]byte, error) {
    return r.fmt.Marshal(protocol.Chat, &format.ChatData{
        Model:    r.model.Name,
        Messages: r.messages,
        Options:  r.options,
    })
}
```

### Client Layer

The client layer orchestrates HTTP execution with retry logic and health tracking. Provider and model come from requests, enabling flexible request composition.

```go
type Client interface {
    HTTPClient() *http.Client
    Execute(ctx context.Context, req request.Request) (any, error)
    ExecuteStream(ctx context.Context, req request.Request) (<-chan *response.StreamingResponse, error)
    IsHealthy() bool
}
```

**Standard Request Flow**:
1. Client calls `req.Marshal()` to get the JSON body (format layer)
2. Client calls `provider.PrepareRequest()` to construct URL and headers
3. Client calls `provider.SetHeaders()` for authentication
4. HTTP request executes with retry logic
5. Client calls `req.Format().Parse()` to parse the response body
6. Client returns the parsed response

**Streaming Request Flow**:
1. Client verifies `protocol.SupportsStreaming()`
2. Client calls `req.Marshal()` and `provider.PrepareStreamRequest()`
3. HTTP request executes (no retry for streaming)
4. Client calls `provider.Stream().ReadStream()` to get `<-chan StreamLine`
5. Client passes each line through `format.ParseStreamChunk()`
6. Client sends typed `StreamingResponse` values on the output channel

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
- Jitter: randomize delay by +/-25% to prevent thundering herd
- Retryable errors: HTTP 429, 502, 503, 504, network errors, DNS errors
- Non-retryable: context cancellation, context deadline, HTTP 4xx (except 429)

### Agent System

Agents provide high-level orchestration with protocol-specific methods:

```go
type Agent interface {
    ID() string
    Client() client.Client
    Provider() providers.Provider
    Format() format.Format
    Model() *model.Model

    Chat(ctx context.Context, prompt string, opts ...map[string]any) (*response.Response, error)
    ChatStream(ctx context.Context, prompt string, opts ...map[string]any) (<-chan *response.StreamingResponse, error)

    Vision(ctx context.Context, prompt string, images []format.Image, opts ...map[string]any) (*response.Response, error)
    VisionStream(ctx context.Context, prompt string, images []format.Image, opts ...map[string]any) (<-chan *response.StreamingResponse, error)

    Tools(ctx context.Context, prompt string, tools []Tool, opts ...map[string]any) (*response.Response, error)

    Embed(ctx context.Context, input string, opts ...map[string]any) (*response.EmbeddingsResponse, error)
}
```

**Agent Responsibilities**:
- **Message Initialization**: Create message arrays with system prompt injection
- **Option Merging**: Merge model's configured protocol options with runtime overrides
- **Request Construction**: Create protocol-specific requests with provider, format, and model
- **Response Type Assertion**: Ensure correct response type for each protocol
- **Streaming Management**: Handle streaming channels for supported protocols

#### Agent Identification

Each agent has a unique identifier assigned at creation time that remains stable throughout its lifetime.

**ID Generation**:
- UUIDv7 format: time-sortable with nanosecond precision
- Auto-generated during agent creation
- Collision-resistant across distributed systems
- Thread-safe for concurrent access

**Orchestration Use Cases**:
- Hub registration and multi-agent coordination
- Message routing to specific agents
- Lifecycle tracking and distributed tracing
- Observability with agent-scoped metrics

## Configuration System

### Agent Configuration

```json
{
  "name": "agent-name",
  "system_prompt": "System instructions for the agent",
  "format": "openai",
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
- `AgentConfig`: Top-level with `name`, `system_prompt`, `format`, `client`, `provider`, and `model`
- `ClientConfig`: HTTP client settings and retry configuration
- `ProviderConfig`: Provider name, base URL, and provider-specific options map
- `ModelConfig`: Model name and protocol-specific capabilities
- `RetryConfig`: Retry behavior configuration

**Format Field**: The `format` field selects the wire format for request marshaling and response parsing. Defaults to `"openai"` when omitted. The format is independent of the provider -- for example, Bedrock uses the `"converse"` format while Ollama and Azure use `"openai"`.

**DefaultProviderConfig**: Sets the provider name to `"ollama"` with an empty `BaseURL`. Providers auto-construct their default base URL when not explicitly configured (e.g., Ollama defaults to `localhost:11434`, Bedrock constructs from region).

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

### Configuration vs Domain Types

**Separation Principle**: Configuration structures use string keys for JSON serialization, while domain types use Protocol constants for type safety.

**Configuration Type** (`config.ModelConfig`):
```go
type ModelConfig struct {
    Name         string
    Capabilities map[string]map[string]any  // String keys for JSON
}
```

**Domain Type** (`model.Model`):
```go
type Model struct {
    Name    string
    Options map[protocol.Protocol]map[string]any  // Protocol constants for type safety
}
```

### Configuration Option Merging

Agent methods merge model's configured protocol options with runtime options, providing baseline defaults that can be overridden per request:

```go
func (a *agent) mergeOptions(
    proto protocol.Protocol,
    opts ...map[string]any,
) map[string]any {
    options := make(map[string]any)
    if modelOpts := a.model.Options[proto]; modelOpts != nil {
        maps.Copy(options, modelOpts)
    }
    if len(opts) > 0 && opts[0] != nil {
        maps.Copy(options, opts[0])
    }
    return options
}
```

**Vision Protocol Special Handling**: After merging, the agent extracts `vision_options` from the merged map, separating image-rendering parameters (e.g., `detail`) from model inference parameters:

```go
var visionOptions map[string]any
if vOpts, exists := options["vision_options"]; exists {
    if vOptsMap, ok := vOpts.(map[string]any); ok {
        visionOptions = vOptsMap
        delete(options, "vision_options")
    }
}
```

**Example Merging Flow**:

```
Configuration: {"vision": {"max_tokens": 4096, "temperature": 0.7, "vision_options": {"detail": "high"}}}
  |
Runtime call: Vision(ctx, prompt, images, map[string]any{"temperature": 0.9})
  |
After merging: {"max_tokens": 4096, "temperature": 0.9, "vision_options": {"detail": "high"}}
  |
After extraction:
  - VisionOptions: {"detail": "high"}
  - Options: {"max_tokens": 4096, "temperature": 0.9}
```

## Data Flow

### Standard Request Flow

```
Agent.Chat(prompt, options)
  |
Agent merges model defaults + runtime options
  |
request.NewChat(provider, format, model, messages, options)
  |
Client.Execute(request)
  |
request.Marshal()  -->  format.Marshal(protocol, ChatData)  -->  JSON bytes
  |
provider.PrepareRequest(protocol, body, headers)  -->  Request{URL, Headers, Body}
  |
provider.SetHeaders(httpRequest)  -->  authentication applied
  |
HTTP Request with Retry Logic
  |
format.Parse(protocol, responseBody)  -->  *response.Response
  |
Agent returns *response.Response
```

### Streaming Request Flow

```
Agent.ChatStream(prompt, options)
  |
Agent merges options, sets "stream": true
  |
request.NewChat(provider, format, model, messages, options)
  |
Client.ExecuteStream(request)
  |
request.Marshal()  -->  format.Marshal(protocol, ChatData)  -->  JSON bytes
  |
provider.PrepareStreamRequest(protocol, body, headers)  -->  Request with stream headers
  |
HTTP Streaming Request (no retry)
  |
provider.Stream().ReadStream(resp.Body)  -->  <-chan StreamLine
  |
format.ParseStreamChunk(protocol, line.Data)  -->  *response.StreamingResponse
  |
Agent streams chunks to caller via <-chan *response.StreamingResponse
```

## Dependency Hierarchy

Packages are ordered from lowest-level to highest-level. Lower layers must not import higher layers.

```
config
  |
protocol
  |
response
  |
format
  |
streaming
  |
identities
  |
providers
  |
model
  |
request
  |
client
  |
agent
  |
mock
```

## Design Patterns

### Protocol-Centric Architecture

Protocols are the primary abstraction, eliminating the need for a separate capability layer:

**Benefits**:
- Simpler architecture with fewer layers
- Direct protocol-to-format mapping
- Clear protocol support via `Protocol.IsValid()` and `Protocol.SupportsStreaming()`
- Protocol constants prevent typos and enable compile-time validation

### Separation of Format and Provider

The format layer handles all wire-format concerns (marshaling, parsing, stream chunk parsing), while the provider layer handles all transport concerns (endpoint routing, authentication, stream transport). This separation enables:

- **Mix-and-match**: Any format can work with any provider (e.g., Converse format with Bedrock provider)
- **Single responsibility**: Adding a new wire format doesn't require touching provider code
- **Testability**: Formats and providers can be tested independently

### Configuration Lifecycle

Configuration only exists during initialization:

```go
// Load configuration
cfg, _ := config.LoadAgentConfig("config.json")

// Create agent (config transforms to domain types)
agent, _ := agent.New(cfg)

// Runtime uses domain types (model.Model, format.Format, providers.Provider)
// Configuration is no longer referenced
```

### Interface-Based Layer Interconnection

Layers communicate through interfaces, not concrete types:

```go
// Agent depends on client, provider, format interfaces
type Agent interface {
    Client() client.Client
    Provider() providers.Provider
    Format() format.Format
    Model() *model.Model
}

// Client receives requests carrying provider and format
type Client interface {
    Execute(ctx context.Context, req request.Request) (any, error)
}

// Request carries provider and format for execution
type Request interface {
    Provider() providers.Provider
    Format() format.Format
}
```

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

## Extension Points

### Adding New Providers

1. Implement `Provider` interface with endpoint mapping, stream reader, and authentication
2. Register in provider registry
3. Configure in JSON

Example:
```go
type CustomProvider struct {
    *BaseProvider
    stream streaming.StreamReader
}

func (p *CustomProvider) Endpoint(proto protocol.Protocol) (string, error) {
    endpoints := map[protocol.Protocol]string{
        protocol.Chat:       "/v1/chat/completions",
        protocol.Vision:     "/v1/chat/completions",
        protocol.Tools:      "/v1/chat/completions",
        protocol.Embeddings: "/v1/embeddings",
    }
    endpoint, ok := endpoints[proto]
    if !ok {
        return "", fmt.Errorf("protocol %s not supported", proto)
    }
    return p.BaseURL() + endpoint, nil
}

func (p *CustomProvider) Stream() streaming.StreamReader {
    return p.stream
}

func (p *CustomProvider) PrepareRequest(ctx context.Context, proto protocol.Protocol, body []byte, headers map[string]string) (*Request, error) {
    endpoint, err := p.Endpoint(proto)
    if err != nil {
        return nil, err
    }
    return &Request{URL: endpoint, Headers: headers, Body: body}, nil
}
```

### Adding New Formats

1. Implement `Format` interface with `Marshal`, `Parse`, and `ParseStreamChunk`
2. Register in format registry

Example:
```go
type CustomFormat struct{}

func (f *CustomFormat) Name() string { return "custom" }

func (f *CustomFormat) Marshal(proto protocol.Protocol, data any) ([]byte, error) {
    // Transform format data types into provider-specific JSON
}

func (f *CustomFormat) Parse(proto protocol.Protocol, body []byte) (any, error) {
    // Parse provider-specific JSON into *response.Response or *response.EmbeddingsResponse
}

func (f *CustomFormat) ParseStreamChunk(proto protocol.Protocol, data []byte) (*response.StreamingResponse, error) {
    // Parse streaming chunk into *response.StreamingResponse
}

func init() {
    format.Register("custom", func() (format.Format, error) {
        return &CustomFormat{}, nil
    })
}
```

### Adding New Protocols

1. Add protocol constant to `protocol/protocol.go`
2. Add data type to `format/data.go`
3. Add response type to `response/` if needed
4. Update `Protocol.IsValid()` and `Protocol.SupportsStreaming()`
5. Add marshal/parse cases in format implementations
6. Update providers to support new protocol endpoint

### Adding New Streaming Transports

1. Implement `StreamReader` interface
2. Use in provider constructor

Example:
```go
type CustomStreamReader struct{}

func (r *CustomStreamReader) ReadStream(ctx context.Context, reader io.Reader) <-chan StreamLine {
    output := make(chan StreamLine)
    go func() {
        defer close(output)
        // Read frames from reader, send StreamLine values
    }()
    return output
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
├── format/
│   └── ...
├── streaming/
│   └── ...
├── identities/
│   └── ...
├── providers/
│   ├── base_test.go
│   ├── ollama_test.go
│   ├── azure_test.go
│   └── registry_test.go
├── client/
│   └── client_test.go
├── agent/
│   └── agent_test.go
└── mock/
    ├── agent_test.go
    ├── client_test.go
    └── provider_test.go
```

### Black-Box Testing

All tests use black-box testing approach with `package <name>_test`:

```go
package config_test

import (
    "testing"
    "github.com/JaimeStill/go-agents/pkg/config"
)
```

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
func TestClient_Execute(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]any{
            "choices": []map[string]any{{
                "message": map[string]any{
                    "role":    "assistant",
                    "content": "Test response",
                },
            }},
        })
    }))
    defer server.Close()

    // Test with server.URL
}
```

### Coverage Requirements

**Minimum Coverage**: 80% across all packages

**Critical Path Coverage**: 100% for:
- Format marshal/parse (format package)
- Configuration validation (config package)
- Protocol routing (client package)
- Request construction (request package)

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

### Mock Package

**Purpose**: The `pkg/mock` package provides configurable mock implementations of all core interfaces for testing code that depends on go-agents.

**Package Structure**:
```
pkg/mock/
├── doc.go           # Package documentation
├── agent.go         # MockAgent implementation
├── client.go        # MockClient implementation
├── provider.go      # MockProvider implementation
├── format.go        # MockFormat implementation
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
   - Options: `WithBaseURL`, `WithEndpointMapping`, `WithPrepareResponse`, `WithEndpointError`, `WithSetHeadersError`

4. **MockFormat** (`format.go`)
   - Implements: `format.Format`
   - Configurable marshal, parse, and stream chunk responses
   - Options: `WithFormatName`, `WithFormatMarshalResponse`, `WithFormatParseResponse`, `WithFormatStreamChunk`

**Helper Constructors** (`helpers.go`):

```go
// Simple chat agent
agent := mock.NewSimpleChatAgent("id", "response text")

// Streaming chat agent
agent := mock.NewStreamingChatAgent("id", []string{"chunk1", "chunk2"})

// Tools agent
agent := mock.NewToolsAgent("id", []response.ToolUseBlock{...})

// Embeddings agent
agent := mock.NewEmbeddingsAgent("id", []float64{0.1, 0.2, 0.3})

// Multi-protocol agent
agent := mock.NewMultiProtocolAgent("id")

// Failing agent (for error handling tests)
agent := mock.NewFailingAgent("id", errors.New("test error"))
```
