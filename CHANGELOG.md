# Changelog

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
