# MVP Completion Implementation Guide

## Overview

The go-agents library core implementation is complete and operational. This guide outlines the remaining work to reach production-ready MVP status: comprehensive testing infrastructure and code documentation.

## Codebase Summary

**Total Lines of Code**: ~2,243 lines across 23 Go files
**Test Coverage**: 0% (no test files exist)
**Package Structure**: 7 packages (config, protocols, capabilities, models, providers, transport, agent)

## Implementation Phases

This work is organized into two sequential phases:

1. **Phase 1: Testing Infrastructure** - Unit tests for all packages
2. **Phase 2: Documentation** - Comprehensive godoc and inline documentation

---

## Phase 1: Testing Infrastructure (Unit Tests)

Unit tests should be implemented in **dependency order** (lowest to highest in the package hierarchy) to ensure dependencies are tested before their consumers.

### Package Dependency Order

```
1. pkg/config/       (foundation - no internal dependencies)
2. pkg/protocols/    (depends on: none)
3. pkg/capabilities/ (depends on: protocols)
4. pkg/models/       (depends on: config, protocols, capabilities)
5. pkg/providers/    (depends on: config, protocols, capabilities, models)
6. pkg/transport/    (depends on: config, protocols, capabilities, models, providers)
7. pkg/agent/        (depends on: config, protocols, capabilities, models, providers, transport)
```

---

### 1.1 Package: `pkg/config/`

**Files to Test**: `duration.go`, `options.go`, `model.go`, `provider.go`, `transport.go`, `agent.go`

**Test File Structure**:
- `duration_test.go` - Custom Duration type with human-readable parsing
- `options_test.go` - Option merge and manipulation
- `model_test.go` - Model configuration structure
- `provider_test.go` - Provider configuration structure
- `transport_test.go` - Transport configuration structure
- `agent_test.go` - Agent configuration structure

**Key Test Cases**:

**duration_test.go**:
- `TestDuration_UnmarshalJSON_ParsesStringFormat` - "24s", "1m", "2h"
- `TestDuration_UnmarshalJSON_ParsesNumericNanoseconds` - Raw numbers
- `TestDuration_UnmarshalJSON_InvalidFormat` - Error handling
- `TestDuration_MarshalJSON` - Serialization to string format
- `TestDuration_ToDuration` - Conversion to time.Duration

**options_test.go**:
- Test any option helper functions (if they exist)

**model_test.go**:
- `TestModelConfig_Unmarshal` - JSON unmarshaling
- `TestModelConfig_Capabilities` - Capability map structure
- `TestModelConfig_Validation` - Valid configuration

**provider_test.go**:
- `TestProviderConfig_Unmarshal` - JSON unmarshaling
- `TestProviderConfig_Options` - Provider-specific options

**transport_test.go**:
- `TestTransportConfig_Unmarshal` - JSON unmarshaling
- `TestTransportConfig_Defaults` - Default values
- `TestTransportConfig_ConnectionPooling` - Pool size configuration

**agent_test.go**:
- `TestAgentConfig_Unmarshal` - JSON unmarshaling
- `TestAgentConfig_FullConfiguration` - Complete config with all fields

---

### 1.2 Package: `pkg/protocols/`

**Files to Test**: `protocol.go`

**Test File**: `protocol_test.go`

**Key Test Cases**:

**Protocol Constants**:
- `TestProtocol_IsValid_ValidProtocols` - Chat, Vision, Tools, Embeddings
- `TestProtocol_IsValid_InvalidProtocol` - Unknown protocol strings
- `TestProtocol_ValidProtocols` - Returns all valid protocols
- `TestProtocol_ProtocolStrings` - Comma-separated string representation

**Message**:
- `TestMessage_NewMessage` - Message creation
- `TestMessage_StringContent` - Simple text content
- `TestMessage_StructuredContent` - Complex content (arrays, maps)

**Request**:
- `TestRequest_Marshal` - JSON marshaling with messages and options
- `TestRequest_Marshal_EmptyOptions` - Handles nil options
- `TestRequest_GetHeaders` - Content-Type header

**ChatResponse**:
- `TestChatResponse_Content_StringContent` - Extract string content
- `TestChatResponse_Content_StructuredContent` - Handle complex content
- `TestChatResponse_Content_EmptyChoices` - Returns empty string

**StreamingChunk**:
- `TestStreamingChunk_Content` - Extract delta content
- `TestStreamingChunk_Content_EmptyChoices` - Returns empty string

**ToolsResponse**:
- `TestToolsResponse_Structure` - Validate structure with tool calls

**EmbeddingsResponse**:
- `TestEmbeddingsResponse_Structure` - Validate embedding data structure

**ExtractOption**:
- `TestExtractOption_ExistsWithCorrectType` - Returns value
- `TestExtractOption_ExistsWithWrongType` - Returns default
- `TestExtractOption_DoesNotExist` - Returns default
- `TestExtractOption_NilOptions` - Returns default

---

### 1.3 Package: `pkg/capabilities/`

**Files to Test**: `capability.go`, `registry.go`, `chat.go`, `vision.go`, `tools.go`, `embeddings.go`

**Test File Structure**:
- `capability_test.go` - Core interfaces and standard implementations
- `registry_test.go` - Format registry operations
- `chat_test.go` - OpenAI chat capability
- `vision_test.go` - OpenAI vision capability
- `tools_test.go` - OpenAI tools capability
- `embeddings_test.go` - OpenAI embeddings capability

**Key Test Cases**:

**capability_test.go**:
- `TestStandardCapability_Name` - Returns capability name
- `TestStandardCapability_Protocol` - Returns associated protocol
- `TestStandardCapability_Options` - Returns option definitions
- `TestStandardCapability_ValidateOptions_ValidOptions` - Accepts valid options
- `TestStandardCapability_ValidateOptions_UnsupportedOption` - Rejects unsupported
- `TestStandardCapability_ValidateOptions_MissingRequired` - Rejects missing required
- `TestStandardCapability_ProcessOptions_AppliesDefaults` - Uses default values
- `TestStandardCapability_ProcessOptions_PreservesProvided` - Keeps provided values
- `TestStandardCapability_SupportsStreaming` - Returns false
- `TestStandardStreamingCapability_SupportsStreaming` - Returns true
- `TestStandardStreamingCapability_IsStreamComplete` - Detects [DONE]
- `TestStandardStreamingCapability_ParseStreamingChunk` - Parses SSE format
- `TestStandardStreamingCapability_ParseStreamingChunk_SkipEmpty` - Skips empty lines

**registry_test.go**:
- `TestRegistry_RegisterFormat` - Registers new format
- `TestRegistry_GetFormat_Exists` - Retrieves registered format
- `TestRegistry_GetFormat_NotExists` - Returns error for unknown format
- `TestRegistry_ThreadSafety` - Concurrent register/get operations
- `TestRegistry_InitialFormats` - Verifies pre-registered formats (openai-chat, etc.)

**chat_test.go**:
- `TestOpenAIChat_Name` - Returns "openai-chat"
- `TestOpenAIChat_Protocol` - Returns Chat protocol
- `TestOpenAIChat_Options` - Defines supported options (temperature, top_p, etc.)
- `TestOpenAIChat_CreateRequest` - Formats request to OpenAI API structure
- `TestOpenAIChat_CreateStreamingRequest` - Adds stream: true
- `TestOpenAIChat_ParseResponse` - Parses ChatResponse
- `TestOpenAIChat_SupportsStreaming` - Returns true

**vision_test.go**:
- `TestOpenAIVision_Name` - Returns "openai-vision"
- `TestOpenAIVision_Protocol` - Returns Vision protocol
- `TestOpenAIVision_Options` - Defines supported options (detail, images)
- `TestOpenAIVision_CreateRequest_LocalImage` - Handles file:// URLs
- `TestOpenAIVision_CreateRequest_WebImage` - Handles http:// URLs
- `TestOpenAIVision_CreateRequest_StructuredContent` - Creates content array
- `TestOpenAIVision_ParseResponse` - Parses ChatResponse
- `TestOpenAIVision_SupportsStreaming` - Returns true

**tools_test.go**:
- `TestOpenAITools_Name` - Returns "openai-tools"
- `TestOpenAITools_Protocol` - Returns Tools protocol
- `TestOpenAITools_Options` - Defines supported options (tools, tool_choice)
- `TestOpenAITools_CreateRequest` - Formats with tool definitions
- `TestOpenAITools_ParseResponse` - Parses ToolsResponse with tool calls
- `TestOpenAITools_SupportsStreaming` - Returns false (non-streaming)

**embeddings_test.go**:
- `TestOpenAIEmbeddings_Name` - Returns "openai-embeddings"
- `TestOpenAIEmbeddings_Protocol` - Returns Embeddings protocol
- `TestOpenAIEmbeddings_Options` - Defines supported options (dimensions, input)
- `TestOpenAIEmbeddings_CreateRequest` - Formats embedding request
- `TestOpenAIEmbeddings_ParseResponse` - Parses EmbeddingsResponse
- `TestOpenAIEmbeddings_SupportsStreaming` - Returns false

---

### 1.4 Package: `pkg/models/`

**Files to Test**: `model.go`, `handler.go`

**Test File Structure**:
- `model_test.go` - Model interface and implementation
- `handler_test.go` - ProtocolHandler

**Key Test Cases**:

**model_test.go**:
- `TestModel_New_ValidConfig` - Creates model from config
- `TestModel_New_InvalidProtocol` - Rejects invalid protocol string
- `TestModel_New_UnknownCapabilityFormat` - Rejects unknown format
- `TestModel_Name` - Returns model name
- `TestModel_SupportsProtocol_Supported` - Returns true for configured protocols
- `TestModel_SupportsProtocol_Unsupported` - Returns false for unconfigured
- `TestModel_GetCapability_Supported` - Returns capability instance
- `TestModel_GetCapability_Unsupported` - Returns error
- `TestModel_GetProtocolOptions_Supported` - Returns configured options
- `TestModel_GetProtocolOptions_Unsupported` - Returns empty map
- `TestModel_UpdateProtocolOptions_Valid` - Updates options
- `TestModel_UpdateProtocolOptions_Invalid` - Rejects invalid options
- `TestModel_UpdateProtocolOptions_Unsupported` - Returns error
- `TestModel_MergeRequestOptions` - Merges config + request options
- `TestModel_MergeRequestOptions_RequestOverridesConfig` - Request takes precedence
- `TestModel_MergeRequestOptions_PreservesConfigDefaults` - Keeps config defaults

**handler_test.go**:
- `TestProtocolHandler_New` - Creates handler
- `TestProtocolHandler_Capability` - Returns capability
- `TestProtocolHandler_Options` - Returns options copy
- `TestProtocolHandler_UpdateOptions` - Updates options
- `TestProtocolHandler_MergeOptions` - Merges config + request options
- `TestProtocolHandler_MergeOptions_RequestPrecedence` - Request wins conflicts

---

### 1.5 Package: `pkg/providers/`

**Files to Test**: `provider.go`, `base.go`, `registry.go`, `ollama.go`, `azure.go`

**Test File Structure**:
- `provider_test.go` - Provider interface
- `base_test.go` - BaseProvider functionality
- `registry_test.go` - Provider registry
- `ollama_test.go` - Ollama provider
- `azure_test.go` - Azure provider

**Key Test Cases**:

**base_test.go**:
- `TestBaseProvider_Name` - Returns provider name
- `TestBaseProvider_BaseURL` - Returns base URL
- `TestBaseProvider_Model` - Returns model instance

**registry_test.go**:
- `TestProviderRegistry_Create_Ollama` - Creates Ollama provider
- `TestProviderRegistry_Create_Azure` - Creates Azure provider
- `TestProviderRegistry_Create_Unknown` - Returns error
- `TestProviderRegistry_ThreadSafety` - Concurrent creation

**ollama_test.go**:
- `TestOllama_GetEndpoint_Chat` - Returns /v1/chat/completions
- `TestOllama_GetEndpoint_Vision` - Returns /v1/chat/completions
- `TestOllama_GetEndpoint_Tools` - Returns /v1/chat/completions
- `TestOllama_GetEndpoint_Embeddings` - Returns /v1/embeddings
- `TestOllama_SetHeaders` - Sets Content-Type
- `TestOllama_PrepareRequest` - Marshals request body
- `TestOllama_PrepareStreamRequest` - Adds stream header
- `TestOllama_ProcessResponse_Success` - Parses successful response
- `TestOllama_ProcessResponse_Error` - Handles error responses
- `TestOllama_ProcessStreamResponse` - Returns stream channel

**azure_test.go**:
- `TestAzure_GetEndpoint_Chat` - Returns deployment URL with api-version
- `TestAzure_GetEndpoint_Vision` - Returns deployment URL
- `TestAzure_GetEndpoint_Tools` - Returns deployment URL
- `TestAzure_GetEndpoint_Embeddings` - Returns deployment URL
- `TestAzure_SetHeaders_APIKey` - Sets api-key header
- `TestAzure_SetHeaders_Bearer` - Sets Authorization header
- `TestAzure_PrepareRequest` - Marshals request body
- `TestAzure_PrepareStreamRequest` - Adds stream header
- `TestAzure_ProcessResponse_Success` - Parses successful response
- `TestAzure_ProcessResponse_Error` - Handles error responses
- `TestAzure_Authentication_APIKey` - Tests API key auth
- `TestAzure_Authentication_Bearer` - Tests Bearer token auth

---

### 1.6 Package: `pkg/transport/`

**Files to Test**: `client.go`

**Test File**: `client_test.go`

**Key Test Cases**:

- `TestClient_New_ValidConfig` - Creates client from config
- `TestClient_New_InvalidProvider` - Handles provider creation error
- `TestClient_Provider` - Returns provider instance
- `TestClient_Model` - Returns model instance
- `TestClient_HTTPClient` - Returns configured HTTP client
- `TestClient_HTTPClient_Timeout` - Verifies timeout configuration
- `TestClient_HTTPClient_ConnectionPool` - Verifies pool size
- `TestClient_ExecuteProtocol_Success` - Executes non-streaming request
- `TestClient_ExecuteProtocol_UnsupportedCapability` - Returns error
- `TestClient_ExecuteProtocol_InvalidOptions` - Validation error
- `TestClient_ExecuteProtocol_OptionMerge` - Merges config + request options
- `TestClient_ExecuteProtocol_NetworkError` - Sets unhealthy on error
- `TestClient_ExecuteProtocolStream_Success` - Executes streaming request
- `TestClient_ExecuteProtocolStream_NotStreamingCapability` - Returns error
- `TestClient_ExecuteProtocolStream_InvalidOptions` - Validation error
- `TestClient_IsHealthy` - Returns health status
- `TestClient_HealthTracking` - Updates health on success/failure

**Mock Strategy**: Use httptest.Server for mocking provider responses

---

### 1.7 Package: `pkg/agent/`

**Files to Test**: `agent.go`, `errors.go`

**Test File Structure**:
- `agent_test.go` - Agent interface and implementation
- `errors_test.go` - Error types (if custom errors exist)

**Key Test Cases**:

**agent_test.go**:
- `TestAgent_New_ValidConfig` - Creates agent from config
- `TestAgent_New_InvalidTransport` - Handles transport creation error
- `TestAgent_Client` - Returns transport client
- `TestAgent_Provider` - Returns provider
- `TestAgent_Model` - Returns model
- `TestAgent_Chat_Success` - Executes chat request
- `TestAgent_Chat_WithSystemPrompt` - Injects system prompt
- `TestAgent_Chat_WithOptions` - Passes options to transport
- `TestAgent_Chat_Error` - Handles execution error
- `TestAgent_ChatStream_Success` - Executes streaming chat
- `TestAgent_ChatStream_SetsStreamOption` - Adds stream: true
- `TestAgent_Vision_Success` - Executes vision request
- `TestAgent_Vision_LocalImage` - Handles file paths
- `TestAgent_Vision_WebImage` - Handles URLs
- `TestAgent_Vision_MultipleImages` - Handles image array
- `TestAgent_VisionStream_Success` - Executes streaming vision
- `TestAgent_Tools_Success` - Executes tools request
- `TestAgent_Tools_ToolDefinitions` - Formats tool definitions
- `TestAgent_Embed_Success` - Executes embeddings request
- `TestAgent_Embed_WithOptions` - Passes options
- `TestAgent_InitMessages_WithSystemPrompt` - Creates message array
- `TestAgent_InitMessages_WithoutSystemPrompt` - Only user message

**Mock Strategy**: Mock transport.Client interface for isolated testing

---

### Test Coverage Goals

**Minimum Coverage**: 80% across all packages
**Critical Path Coverage**: 100% for:
- Request/response parsing (protocols package)
- Configuration validation (config package)
- Protocol routing (transport package)
- Option merging and validation (models, transport)

**Coverage Commands**:
```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out

# View coverage by package
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

---

## Phase 2: Documentation

Documentation should be added **after** testing to ensure accurate descriptions of tested behavior. Follow idiomatic Go conventions.

### 2.1 Package Documentation (High Priority)

Add package-level godoc comments to each package. Place at the top of the primary file in each package.

**pkg/config/config.go** (create if doesn't exist, or use agent.go):
```go
// Package config provides configuration management for the go-agents library.
// It defines structures for agent, model, provider, and transport configuration
// with support for human-readable duration strings and JSON serialization.
//
// Configuration files use hierarchical JSON structure with transport-based
// organization. Example:
//
//	{
//	  "name": "my-agent",
//	  "system_prompt": "You are a helpful assistant",
//	  "transport": {
//	    "provider": {
//	      "name": "ollama",
//	      "base_url": "http://localhost:11434",
//	      "model": {
//	        "name": "llama3.2:3b",
//	        "capabilities": {
//	          "chat": {"format": "openai-chat", "options": {...}}
//	        }
//	      }
//	    },
//	    "timeout": "24s",
//	    "connection_pool_size": 10
//	  }
//	}
//
// Duration values support human-readable strings ("24s", "1m", "2h") or
// numeric nanoseconds for programmatic configuration.
package config
```

**pkg/protocols/protocol.go**:
```go
// Package protocols defines the core protocol types and message structures
// for LLM interactions. It provides constants for Chat, Vision, Tools, and
// Embeddings protocols, along with request/response structures and helper
// functions for working with protocol messages and options.
//
// The protocol system defines communication contracts for different LLM
// interaction patterns:
//
//   - Chat: Text-based completions with conversation context
//   - Vision: Image analysis combined with text prompts
//   - Tools: Function calling with structured tool definitions
//   - Embeddings: Vector embedding generation for text
//
// Each protocol has specialized response types that capture protocol-specific
// data structures while maintaining a consistent request interface.
package protocols
```

**pkg/capabilities/capability.go**:
```go
// Package capabilities provides protocol-specific capability implementations
// for different LLM API formats. It includes a thread-safe registry system
// for registering and retrieving capability formats, along with implementations
// for OpenAI-compatible formats (chat, vision, tools, embeddings, reasoning).
//
// Capabilities implement the interface between generic protocol requests and
// provider-specific API structures. Each capability defines:
//
//   - Supported options with validation rules
//   - Request formatting for the target API
//   - Response parsing from the target API
//   - Streaming support (where applicable)
//
// The capability format registry enables runtime registration of new formats
// and configuration-driven capability selection. Format names follow the
// pattern "provider-protocol" (e.g., "openai-chat", "anthropic-chat").
//
// Example capability registration:
//
//	func init() {
//	    capabilities.RegisterFormat("openai-chat", func() capabilities.Capability {
//	        return NewOpenAIChatCapability()
//	    })
//	}
package capabilities
```

**pkg/models/model.go**:
```go
// Package models provides the Model abstraction for composing capabilities
// across multiple protocols. Models use ProtocolHandler instances to manage
// stateful protocol configuration while delegating behavior to registered
// capability formats.
//
// The Model interface enables:
//
//   - Protocol support queries
//   - Capability retrieval by protocol
//   - Protocol-specific option management
//   - Request option merging (config + runtime)
//
// Models are created from configuration that specifies which protocols the
// model supports and which capability format to use for each protocol.
// This enables flexible composition where different protocols can use
// different API formats.
//
// Example model configuration:
//
//	{
//	  "name": "gpt-4o",
//	  "capabilities": {
//	    "chat": {"format": "openai-chat", "options": {"temperature": 0.7}},
//	    "vision": {"format": "openai-vision", "options": {"detail": "auto"}},
//	    "tools": {"format": "openai-tools", "options": {"tool_choice": "auto"}}
//	  }
//	}
package models
```

**pkg/providers/provider.go**:
```go
// Package providers implements LLM service integrations. Providers are
// format-agnostic and work with any capability format by routing protocols
// to appropriate endpoints and handling service-specific authentication
// and request/response transformations.
//
// Provider responsibilities:
//
//   - Endpoint mapping: Route protocols to provider-specific API URLs
//   - Authentication: Handle API keys, bearer tokens, or custom auth schemes
//   - Request preparation: Add provider-specific headers and URL parameters
//   - Response processing: Delegate parsing to capability implementations
//
// Providers work with pre-formatted requests from capabilities. The provider's
// role is to route the request to the correct endpoint and handle service-level
// concerns, not to understand or modify the request/response format.
//
// Implemented providers:
//
//   - Ollama: Local model hosting with OpenAI-compatible endpoints
//   - Azure: Azure OpenAI with API key and Entra ID authentication
package providers
```

**pkg/transport/client.go**:
```go
// Package transport provides the Client abstraction for orchestrating
// requests across providers. It handles option merging, validation,
// HTTP client configuration, connection pooling, and health tracking.
//
// The transport layer serves as the execution orchestrator:
//
//  1. Capability selection based on protocol
//  2. Option merging (model config + request options)
//  3. Option validation (after merge)
//  4. Request formatting via capability
//  5. Provider routing and execution
//  6. Response parsing via capability
//
// Transport clients maintain health status based on request success/failure
// and configure HTTP clients with connection pooling, timeouts, and retry
// policies.
//
// The client interface provides both non-streaming and streaming execution
// paths with protocol-agnostic request handling.
package transport
```

**pkg/agent/agent.go**:
```go
// Package agent provides high-level orchestration for LLM interactions.
// Agents expose protocol-specific methods (Chat, Vision, Tools, Embed)
// with simplified interfaces that handle message initialization and
// option management.
//
// The Agent interface abstracts away transport and provider complexity,
// offering intuitive methods for each protocol:
//
//   - Chat/ChatStream: Text completions with optional system prompts
//   - Vision/VisionStream: Image analysis with text prompts
//   - Tools: Function calling with tool definitions
//   - Embed: Vector embedding generation
//
// Agents are created from configuration and automatically inject system
// prompts, initialize message structures, and forward requests to the
// transport layer for execution.
//
// Example usage:
//
//	agent, err := agent.New(config)
//	if err != nil {
//	    return err
//	}
//
//	response, err := agent.Chat(ctx, "What is Go?")
//	if err != nil {
//	    return err
//	}
//
//	fmt.Println(response.Content())
package agent
```

---

### 2.2 Validation Strategy

While the library includes comprehensive unit tests with mocks, integration validation is performed manually using real provider interactions rather than automated integration tests.

**Validation Approach**:

The `tools/prompt-agent` command-line utility serves as the primary integration validation tool. This approach eliminates credential management complexity, removes dependencies on live services from the test suite, and provides real-world validation of provider integration.

**How Validation Works**:

1. **README Examples as Tests**: All usage examples in README.md are executable via `tools/prompt-agent`
2. **Manual Testing Workflow**: Developers run README examples against live providers when needed
3. **Real Integration Verification**: If README examples execute successfully, integration works
4. **No Automated Integration Tests**: The test suite contains only unit tests with mocks

**Benefits**:
- ✅ No credential exposure in test files
- ✅ No live service dependencies in CI/CD
- ✅ Real provider validation when needed
- ✅ Executable documentation (examples are actually tested)
- ✅ Developer-controlled testing (credentials only needed for manual validation)

**Validation Commands**:

```bash
# Test Ollama integration (requires local Ollama running)
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.ollama.json \
  -prompt "What is Go?" \
  -stream

# Test Azure integration (requires valid token)
AZURE_TOKEN=$(. scripts/azure/utilities/get-foundry-token.sh)
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.azure-entra.json \
  -token $AZURE_TOKEN \
  -prompt "Describe Kubernetes" \
  -stream

# Test all protocols (chat, vision, tools, embeddings)
# See README.md for comprehensive protocol examples
```

**When to Run Validation**:
- Before releases
- After provider-specific changes
- When adding new capability formats
- To verify configuration changes

---

### 2.3 Exported Type/Function Documentation

Document all exported types, interfaces, functions, and methods following Go conventions:

**Documentation Rules**:
1. Start with the declared name
2. Use complete sentences
3. Include usage examples for non-trivial patterns
4. Document parameter constraints and edge cases
5. Cross-reference related types with [TypeName]

**Example - Interface Documentation**:
```go
// Model represents an LLM model with composable protocol capabilities.
// Models are created from ModelConfig and support querying for protocol
// support, retrieving capabilities, and managing protocol-specific options.
//
// Each protocol is configured independently with its own capability format
// and options, enabling flexible composition where different protocols can
// use different API formats (e.g., OpenAI chat + Anthropic tools).
//
// Options are managed at two levels:
//
//   - Configuration options: Set at model creation, persisted across requests
//   - Request options: Passed per-request, merged with configuration options
//
// Request options take precedence over configuration options during merge.
// Option validation occurs after merge in the transport layer.
//
// Example:
//
//	cfg := &config.ModelConfig{
//	    Name: "gpt-4o",
//	    Capabilities: map[string]config.CapabilityConfig{
//	        "chat": {
//	            Format: "openai-chat",
//	            Options: map[string]any{"temperature": 0.7},
//	        },
//	    },
//	}
//	model, err := models.New(cfg)
//	if err != nil {
//	    return err
//	}
//
//	// Check protocol support
//	if !model.SupportsProtocol(protocols.Chat) {
//	    return errors.New("chat not supported")
//	}
type Model interface {
    // Name returns the model identifier used in API requests.
    Name() string

    // SupportsProtocol reports whether the model supports the given protocol.
    // Returns false if the protocol was not configured in the model's capabilities.
    SupportsProtocol(p protocols.Protocol) bool

    // GetCapability returns the capability implementation for the given protocol.
    // Returns an error if the protocol is not supported by this model.
    GetCapability(p protocols.Protocol) (capabilities.Capability, error)

    // GetProtocolOptions returns the configuration options for the given protocol.
    // Returns an empty map if the protocol is not supported.
    GetProtocolOptions(p protocols.Protocol) map[string]any

    // UpdateProtocolOptions updates the configuration options for the given protocol.
    // Options are validated against the protocol's capability before being stored.
    // Returns an error if the protocol is not supported or options are invalid.
    UpdateProtocolOptions(p protocols.Protocol, options map[string]any) error

    // MergeRequestOptions merges configuration options with request options for
    // the given protocol. Request options take precedence over configuration options.
    // Returns the merged options map.
    MergeRequestOptions(p protocols.Protocol, requestOptions map[string]any) map[string]any
}
```

**Example - Function Documentation**:
```go
// New creates a Model from configuration.
//
// The configuration must specify at least one capability with a valid protocol
// name and registered capability format. Protocol names must match one of the
// constants defined in the protocols package (Chat, Vision, Tools, Embeddings).
//
// Returns an error if:
//   - Any protocol name is invalid
//   - Any capability format is not registered
//   - The configuration is otherwise malformed
//
// Example:
//
//	cfg := &config.ModelConfig{
//	    Name: "llama3.2:3b",
//	    Capabilities: map[string]config.CapabilityConfig{
//	        "chat": {Format: "openai-chat", Options: map[string]any{"temperature": 0.7}},
//	        "tools": {Format: "openai-tools", Options: map[string]any{"tool_choice": "auto"}},
//	    },
//	}
//	model, err := New(cfg)
//	if err != nil {
//	    log.Fatalf("Failed to create model: %v", err)
//	}
func New(cfg *config.ModelConfig) (Model, error)
```

**Example - Struct Field Documentation**:
```go
// CapabilityOption defines a supported option for a capability format.
type CapabilityOption struct {
    // Option is the option name as it appears in the API request.
    Option string

    // Required indicates whether this option must be provided.
    // If true, requests missing this option will fail validation.
    Required bool

    // DefaultValue is used when the option is not provided.
    // Must be the correct type for this option. A nil value means no default.
    DefaultValue any
}
```

---

### 2.4 Inline Documentation (Medium Priority)

Add inline comments for complex logic and non-obvious implementations.

**Focus Areas**:

1. **Option Merging Logic** (models/handler.go):
```go
// MergeOptions combines configuration options with request options.
// Request options take precedence when keys conflict. Both option maps
// are preserved unchanged - a new map is returned with merged values.
func (h *ProtocolHandler) MergeOptions(requestOptions map[string]any) map[string]any {
    result := make(map[string]any)

    // Start with configuration options as the base
    for k, v := range h.options {
        result[k] = v
    }

    // Override with request options (request takes precedence)
    for k, v := range requestOptions {
        result[k] = v
    }

    return result
}
```

2. **Streaming Chunk Parsing** (capabilities/capability.go):
```go
// ParseStreamingChunk parses a Server-Sent Events (SSE) formatted line.
// SSE format: "data: {json}\n\n"
// Special case: "data: [DONE]\n\n" signals stream completion
func (s *StandardStreamingCapability) ParseStreamingChunk(data []byte) (*protocols.StreamingChunk, error) {
    line := string(data)

    // Handle Server-Sent Events format - strip "data: " prefix
    if after, ok := strings.CutPrefix(line, "data: "); ok {
        line = after
    }

    // Skip empty lines or stream completion markers
    if line == "" || strings.Contains(line, "[DONE]") {
        return nil, fmt.Errorf("skip line")
    }

    var chunk protocols.StreamingChunk
    if err := json.Unmarshal([]byte(line), &chunk); err != nil {
        return nil, err
    }
    return &chunk, nil
}
```

3. **Provider Endpoint Mapping** (providers/azure.go):
```go
// GetEndpoint constructs the Azure OpenAI endpoint URL.
// Azure uses deployment-based routing: /openai/deployments/{deployment}/...
// The deployment name is configured separately from the model name to support
// scenarios where deployment name differs from model name.
func (a *AzureProvider) GetEndpoint(protocol protocols.Protocol) (string, error) {
    deployment := protocols.ExtractOption(a.options, "deployment", a.Model().Name())
    apiVersion := protocols.ExtractOption(a.options, "api_version", "2024-08-01-preview")

    var path string
    switch protocol {
    case protocols.Chat, protocols.Vision, protocols.Tools:
        path = fmt.Sprintf("/openai/deployments/%s/chat/completions", deployment)
    case protocols.Embeddings:
        path = fmt.Sprintf("/openai/deployments/%s/embeddings", deployment)
    default:
        return "", fmt.Errorf("unsupported protocol: %s", protocol)
    }

    return fmt.Sprintf("%s%s?api-version=%s", a.BaseURL(), path, apiVersion), nil
}
```

4. **Authentication Header Construction** (providers/azure.go):
```go
// SetHeaders configures authentication headers for Azure OpenAI requests.
// Supports two authentication modes:
//   - api_key: Uses "api-key" header (default)
//   - bearer: Uses "Authorization: Bearer {token}" header (Entra ID)
func (a *AzureProvider) SetHeaders(req *http.Request) {
    authType := protocols.ExtractOption(a.options, "auth_type", "api_key")

    switch authType {
    case "api_key":
        // Azure OpenAI uses "api-key" header (not "Authorization")
        req.Header.Set("api-key", a.Token())
    case "bearer":
        // Entra ID authentication uses standard Bearer token
        req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.Token()))
    }
}
```

5. **HTTP Client Configuration** (transport/client.go):
```go
// HTTPClient returns a configured HTTP client with connection pooling,
// timeouts, and keep-alive settings optimized for LLM API requests.
func (c *client) HTTPClient() *http.Client {
    return &http.Client{
        // Overall request timeout includes connection time + response time
        Timeout: c.config.Timeout.ToDuration(),
        Transport: &http.Transport{
            // Connection pooling: maintain idle connections for reuse
            MaxIdleConns:        c.config.ConnectionPoolSize,
            MaxIdleConnsPerHost: c.config.ConnectionPoolSize,
            // Close idle connections after this duration to prevent stale connections
            IdleConnTimeout:     c.config.ConnectionTimeout.ToDuration(),
        },
    }
}
```

---

### 2.5 Documentation Verification

**Check Documentation Quality**:
```bash
# Generate godoc locally
go install golang.org/x/tools/cmd/godoc@latest
godoc -http=:6060

# Visit http://localhost:6060/pkg/github.com/JaimeStill/go-agents/
```

**Documentation Checklist**:
- [ ] All packages have package-level documentation
- [ ] All exported types have documentation
- [ ] All exported functions have documentation
- [ ] All exported methods have documentation
- [ ] Complex logic has inline comments
- [ ] Usage examples included for non-trivial APIs
- [ ] Cross-references use [TypeName] syntax
- [ ] Documentation follows "start with declared name" convention
- [ ] No spelling or grammar errors

---

## Implementation Order Summary

**Phase 1: Unit Testing (Estimated: 16-24 hours)**
1. pkg/config/ (2-3 hours)
2. pkg/protocols/ (2-3 hours)
3. pkg/capabilities/ (4-6 hours)
4. pkg/models/ (2-3 hours)
5. pkg/providers/ (3-4 hours)
6. pkg/transport/ (2-3 hours)
7. pkg/agent/ (1-2 hours)

**Phase 2: Documentation (Estimated: 8-12 hours)**
1. Package documentation (2-3 hours)
2. Validation strategy documentation (included in package docs)
3. Exported type/function documentation (4-6 hours)
4. Inline documentation (2-3 hours)

**Total Estimated Effort**: 24-36 hours

---

## Success Criteria

### Phase 1 Complete
- [ ] All packages have corresponding test files
- [ ] Test coverage reaches 80% minimum
- [ ] Critical paths have 100% coverage
- [ ] All tests pass: `go test ./...`

### Phase 2 Complete
- [ ] All packages have godoc comments
- [ ] All exported types/functions documented
- [ ] Complex logic has inline comments
- [ ] Validation strategy documented
- [ ] `go doc` produces readable output for all packages
- [ ] Documentation verified at http://localhost:6060

### MVP Complete
- [ ] Both phases completed
- [ ] Documentation reviewed for accuracy
- [ ] Test coverage report generated and reviewed
- [ ] README examples verified via prompt-agent tool
- [ ] PROJECT.md updated with publishing section
- [ ] Ready for v0.1.0 pre-release tag

---

## Additional Resources

**Go Testing**:
- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Testing Techniques](https://golang.org/doc/tutorial/add-a-test)

**Go Documentation**:
- [Effective Go - Commentary](https://golang.org/doc/effective_go#commentary)
- [Godoc Documentation](https://go.dev/blog/godoc)
- [Go Doc Comments](https://go.dev/doc/comment)

**Testing Libraries**:
- [testify](https://github.com/stretchr/testify) - Assertions and mocking
- [httptest](https://golang.org/pkg/net/http/httptest/) - HTTP testing utilities
- [gomock](https://github.com/golang/mock) - Mock generation
