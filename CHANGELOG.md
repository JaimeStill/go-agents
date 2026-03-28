# Changelog

## [v0.5.0] - 2026-03-28

**Breaking Changes**:
- `Provider` interface: removed `Marshal()`, `ProcessResponse()`, `ProcessStreamResponse()` methods — wire format handling moved to `Format` interface
- `Provider` interface: added `Stream() streaming.StreamReader` method for provider-specific streaming transport
- `Request` interface: added `Format() format.Format` method
- Request constructors (`NewChat`, `NewVision`, `NewTools`, `NewEmbeddings`) now require a `format.Format` parameter
- `Agent.Vision()` and `Agent.VisionStream()` signature changed from `images []string` to `images []format.Image`
- `Agent.Chat()`, `Agent.Vision()`, `Agent.Tools()` now return `*response.Response` instead of `*response.ChatResponse` / `*response.ToolsResponse`
- `Agent.ChatStream()`, `Agent.VisionStream()` now return `<-chan *response.StreamingResponse` instead of `<-chan *response.StreamingChunk`
- `response.ChatResponse`, `response.ToolsResponse`, `response.StreamingChunk` types removed — replaced by unified `response.Response` and `response.StreamingResponse` with content blocks
- `response.EmbeddingsResponse` simplified — `Data []struct{...}` replaced by `Embeddings [][]float64`
- `response.TokenUsage` field names changed: `PromptTokens` → `InputTokens`, `CompletionTokens` → `OutputTokens`
- `DefaultProviderConfig()` no longer sets `BaseURL` — providers auto-construct defaults when `BaseURL` is empty

**Added**:
- `pkg/format` package for wire format abstraction
  - `Format` interface with `Marshal()`, `Parse()`, `ParseStreamChunk()` methods
  - `OpenAIFormat` implementation for OpenAI-compatible providers
  - `ConverseFormat` implementation for AWS Bedrock Converse API
  - `Image` type for provider-neutral image representation with raw bytes and format
  - `ChatData`, `VisionData`, `ToolsData`, `EmbeddingsData` format input types
  - `ToolDefinition` type (moved from `providers` package)
  - Format registry with `Register()`, `Create()`, `ListFormats()`
- `pkg/streaming` package for streaming transport abstraction
  - `StreamReader` interface and `StreamLine` type
  - `SSEReader` for Server-Sent Events (Ollama, Azure)
  - `EventStreamReader` for AWS event stream binary framing (Bedrock)
  - `SSEMedia` and `EventStreamMedia` content type constants
- `pkg/providers/bedrock.go` — AWS Bedrock provider
  - Converse API endpoint routing with model ID in URL path
  - SigV4 request signing via `pkg/identities`
  - Default, static, and profile authentication modes
  - Auto-constructed base URL from region
- `pkg/identities/aws.go` — AWS credential source
  - `AWSAuthType` typed string enum (`AWSAuthDefault`, `AWSAuthStatic`, `AWSAuthProfile`)
  - `AWSCredentialSource` with SigV4 signing via `SignRequest()`
  - Default credential chain, static keys, and named profile support
- `response.Response` — unified response type for Chat, Vision, and Tools protocols
  - `ContentBlock` interface with `TextBlock` and `ToolUseBlock` implementations
  - `Text()` method for concatenated text content
  - `ToolCalls()` method for filtered tool use blocks
- `response.StreamingResponse` — unified streaming chunk type with content blocks
- `AgentConfig.Format` string field (defaults to `"openai"`)
- `MockFormat` type in `pkg/mock` for testing format implementations

**Changed**:
- `BaseProvider` reduced to `Name()` and `BaseURL()` — marshaling removed
- Ollama and Azure providers auto-construct default base URLs when not configured
- `agent.New()` defaults format to `"openai"` when `AgentConfig.Format` is empty
- Mock package helpers (`NewSimpleChatAgent`, `NewToolsAgent`, etc.) use content block response types
- `tools/prompt-agent` updated for new response types and `format.Image`

**Removed**:
- `pkg/providers/data.go` — types moved to `pkg/format/data.go`
- `pkg/providers/sse.go` — moved to `pkg/streaming/sse.go`
- `response.ChatResponse`, `response.ToolsResponse`, `response.StreamingChunk` types
- `response.ToolCall`, `response.ToolCallFunction` types (replaced by `response.ToolUseBlock`)
- `response.Parse()`, `response.ParseStreamChunk()`, and all protocol-specific parse functions — parsing now handled by format implementations
- `providers.ToolDefinition` (moved to `format.ToolDefinition`)

## [v0.4.0] - 2026-03-13

**Breaking Changes**:
- `Provider.SetHeaders` signature changed from `SetHeaders(req *http.Request)` to `SetHeaders(ctx context.Context, req *http.Request) error`

**Added**:
- `pkg/identities` package for managed identity token sources
  - `AzureTokenSource` type wrapping `azidentity.ManagedIdentityCredential`
  - `NewAzureTokenSource(scope, clientID)` constructor with default scope
  - `GetToken(ctx)` method for dynamic token acquisition
- Azure provider `managed_identity` auth type for containerized deployments
  - System-assigned identity (default) and user-assigned identity via `client_id` option
  - Configurable token scope via `resource` option (defaults to `https://cognitiveservices.azure.com/.default`)
- `MockProvider.WithSetHeadersError` option for testing authentication failures

**Changed**:
- Azure provider `NewAzure` constructor validates `auth_type` and returns error for unsupported values
- Client layer (`execute`, `executeStream`) now propagates `SetHeaders` errors

## [v0.3.0] - 2025-12-01

**Breaking Changes**:
- Removed `pkg/types` package - split into `pkg/protocol`, `pkg/response`, `pkg/model`, and `pkg/request`
- Flattened `AgentConfig` structure: `Provider` and `Model` are now peer fields with `Client` (not nested under `Client`)
- `ClientConfig` no longer contains `Provider` - HTTP client settings only
- `ProviderConfig` no longer contains `Model` - provider connection settings only
- `Agent.Model()` returns `*model.Model` instead of `*types.Model`

**Added**:
- `pkg/protocol` package for protocol types and message structures
  - `Protocol` type with constants: `Chat`, `Vision`, `Tools`, `Embeddings`
  - `Message` type and `NewMessage()` constructor
  - `IsValid()`, `ValidProtocols()`, `ProtocolStrings()` functions
  - `Protocol.SupportsStreaming()` method
- `pkg/response` package for response parsing and types
  - `ChatResponse` type with `Content()` method
  - `StreamingChunk` type with `Content()` method
  - `EmbeddingsResponse` type
  - `ToolsResponse` type with `ToolCall` and `ToolCallFunction` types
  - `ParseChat()`, `ParseEmbeddings()`, `ParseTools()`, `ParseStream()` functions
- `pkg/model` package for model runtime type
  - `Model` type with `Name` and `Options` fields
  - `New()` function for creating Model from ModelConfig
- `pkg/request` package for request interface and protocol-specific request types
  - `Request` interface with `Protocol()`, `Headers()`, `Marshal()`, `Provider()`, `Model()` methods
  - `ChatRequest` type with `NewChat()` constructor
  - `VisionRequest` type with `NewVision()` constructor
  - `ToolsRequest` type with `NewTools()` constructor
  - `EmbeddingsRequest` type with `NewEmbeddings()` constructor

**Changed**:
- `AgentConfig.Provider` moved from `AgentConfig.Client.Provider` to top-level field
- `AgentConfig.Model` moved from `AgentConfig.Client.Provider.Model` to top-level field
- Mock package types updated to use `pkg/protocol` and `pkg/response`

**Removed**:
- `pkg/types` package (replaced by `pkg/protocol`, `pkg/response`, `pkg/model`, `pkg/request`)
- Nested configuration hierarchy (`Client.Provider.Model`)

## [v0.2.1] - 2025-11-01

**Changed**:
- `types.VisionRequest.ImageOptions` renamed to `VisionOptions` for protocol naming consistency
- Vision protocol configuration format now uses nested `vision_options` map instead of flat `detail` key (breaking configuration change)
- Agent protocol methods now merge model's configured options with runtime options, enabling configuration defaults with runtime overrides

**Fixed**:
- Agent methods now properly apply model's configured protocol options as baseline values

## [v0.2.0] - 2025-10-29

**Breaking Changes**:
- Removed `pkg/capabilities` package - protocol handling now integrated directly into `pkg/types`
- Removed `pkg/models` package - model abstraction replaced by `types.Model` runtime type
- Removed `pkg/protocols` package - merged into `pkg/types` with protocol-specific request types
- Removed `pkg/transport` package - renamed to `pkg/client` with enhanced retry logic
- `Agent.Client()` now returns `client.Client` instead of `transport.Client`
- `Agent.Model()` now returns `*types.Model` instead of `models.Model`
- `Agent.ChatStream()` now returns `<-chan *types.StreamingChunk` instead of `<-chan types.StreamingChunk`
- `Agent.VisionStream()` now returns `<-chan *types.StreamingChunk` instead of `<-chan types.StreamingChunk`
- Configuration field `AgentConfig.Transport` renamed to `AgentConfig.Client`

**Added**:
- `pkg/types` package consolidating protocol types, request/response structures, and model runtime type
  - `Protocol` type with constants: `Chat`, `Vision`, `Tools`, `Embeddings`
  - Protocol-specific request types: `ChatRequest`, `VisionRequest`, `ToolsRequest`, `EmbeddingsRequest`
  - `ProtocolRequest` interface for unified request handling
  - `Model` runtime type with protocol-specific options
  - `NewModel()` function for creating models
  - `FromConfig()` function for converting `ModelConfig` to `Model`
  - `Protocol.SupportsStreaming()` method
  - `IsValid()` function for protocol validation
  - `ValidProtocols()` function returning all supported protocols
  - `ProtocolStrings()` function for display formatting
  - `ParseResponse()` function for protocol-aware response parsing
  - `ParseStreamChunk()` function for protocol-aware streaming chunk parsing
  - `ExtractOption[T]()` generic function for type-safe option extraction
  - `ToolDefinition` type for provider-agnostic tool definitions
  - Protocol-specific parsers: `ParseChatResponse()`, `ParseVisionResponse()`, `ParseToolsResponse()`, `ParseEmbeddingsResponse()`
  - Protocol-specific streaming parsers: `ParseChatStreamChunk()`, `ParseVisionStreamChunk()`, `ParseToolsStreamChunk()`
- `pkg/client` package for HTTP client orchestration with retry logic
  - `Client` interface with `ExecuteProtocol()` and `ExecuteProtocolStream()` methods
  - `New()` function for creating clients from configuration
  - Exponential backoff retry logic with jitter for transient failures
  - Health tracking via `IsHealthy()` method
- `pkg/config` package additions:
  - `ClientConfig` type replacing `TransportConfig`
  - `RetryConfig` type for configuring retry behavior
  - `DefaultClientConfig()` function
  - `DefaultRetryConfig()` function
  - `ClientConfig.Merge()` method
- `pkg/agent` package additions:
  - `ErrorType` type for categorizing agent errors
  - `AgentError` type with detailed error context
  - `NewAgentError()` function
  - `NewAgentInitError()` helper function
  - `NewAgentLLMError()` helper function
  - Error option functions: `WithCode()`, `WithCause()`, `WithName()`, `WithClient()`, `WithID()`
- `pkg/providers` package additions:
  - `PrepareStreamRequest()` method for streaming-specific request preparation
  - `ProcessStreamResponse()` method for streaming response processing

**Changed**:
- `Provider.GetEndpoint()` now accepts `types.Protocol` instead of `protocols.Protocol`
- `Provider.PrepareRequest()` now accepts `types.ProtocolRequest` instead of separate protocol and request parameters
- `Provider.ProcessResponse()` now accepts `types.Protocol` parameter for protocol-aware parsing
- `Provider.Model()` now returns `*types.Model` instead of `models.Model`
- Agent protocol methods now accept variadic options: `Chat()`, `ChatStream()`, `Vision()`, `VisionStream()`, `Tools()`, `Embed()`
- Mock package types updated to use `pkg/types` and `pkg/client` instead of removed packages
  - `MockAgent` uses `client.Client` instead of `transport.Client`
  - Mock helper functions use `types.*` response types
  - `WithClient()` accepts `client.Client`

**Removed**:
- `Capability` interface and all capability-related types
- `CapabilityRequest` type (replaced by protocol-specific request types)
- `CapabilityOption` type
- `StreamingCapability` interface
- `StandardCapability` type
- `StandardStreamingCapability` type
- Capability registry and format registration system
- `Model` interface from `pkg/models`
- `ProtocolHandler` type
- `TransportConfig` type (replaced by `ClientConfig`)
- Model option merging and update methods
- `MockModel` type
- `MockCapability` type

## [v0.1.3] - 2025-10-23

**Changed**:
- Capability format naming: renamed from vendor-centric to specification-based naming
  - `openai-chat` → `chat` (standard OpenAI-compatible chat completions)
  - `openai-vision` → `vision` (standard OpenAI-compatible vision API)
  - `openai-tools` → `tools` (standard OpenAI-compatible function calling)
  - `openai-embeddings` → `embeddings` (standard OpenAI-compatible embeddings)
  - `openai-reasoning` → `o-chat` (OpenAI o-series reasoning models)

**Added**:
- `o-vision` capability format for OpenAI o-series vision reasoning models
  - Supports `max_completion_tokens`, `reasoning_effort`, `images`, `detail` parameters
  - Uses o-series parameter restrictions (no temperature/top_p support)

## [v0.1.2] - 2025-10-10

**Added**:
- `pkg/mock` package providing mock implementations for testing
- `MockAgent`, `MockClient`, `MockProvider`, `MockModel`, `MockCapability` types
- Helper constructors: `NewSimpleChatAgent`, `NewStreamingChatAgent`, `NewToolsAgent`, `NewEmbeddingsAgent`, `NewMultiProtocolAgent`, `NewFailingAgent`

## [v0.1.1] - 2025-10-10

**Added**:
- `ID() string` method to Agent interface returning unique UUIDv7 identifier

## [v0.1.0] - 2025-10-09

Initial pre-release.

**Protocols**:
- Chat
- Vision
- Tools
- Embeddings

**Capability Formats**:
- openai-chat
- openai-vision
- openai-tools
- openai-embeddings
- openai-reasoning

**Providers**:
- Ollama
- Azure AI Foundry
