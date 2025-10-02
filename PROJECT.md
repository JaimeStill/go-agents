# Project Roadmap

## Project Scope

**go-agents** is a primitive Agent interface library for Go, providing foundational abstractions for building LLM-powered applications. The library focuses on delivering clean, composable interfaces for interacting with language models across multiple providers.

### What This Library Provides

- **Protocol Abstractions**: Standardized interfaces for core LLM interactions (Chat, Vision, Tools, Embeddings)
- **Capability Format System**: Extensible format registration supporting provider-specific API structures
- **Provider Integration**: Unified interface for multiple LLM providers (OpenAI, Azure, Ollama, etc.)
- **Transport Layer**: HTTP client orchestration with connection pooling, retries, and streaming support
- **Configuration Management**: Type-safe configuration with human-readable durations and validation

### What This Library Does NOT Provide

This library is intentionally scoped as a **primitive interface layer**. The following capabilities are outside the project scope and should be implemented as supplemental packages:

- **Tool Execution**: Runtime execution of tool functions, security policies, and result handling
- **Context Management**: Token counting, context window optimization, and conversation memory
- **Multi-Agent Orchestration**: Agent coordination, workflow graphs, and state machines
- **Retrieval Systems**: Vector stores, RAG implementations, and document processing pipelines
- **Fine-Tuning Infrastructure**: Model training, evaluation, and deployment tooling

## Design Philosophy

### Core Principles

1. **Minimal Abstractions**: Provide only the essential primitives needed for LLM interactions
2. **Provider Agnostic**: Support any LLM provider through clean interface boundaries
3. **Format Extensibility**: Enable new API formats without modifying core library code
4. **Configuration-Driven**: Compose capabilities through declarative configuration
5. **Type Safety**: Leverage Go's type system for compile-time safety and clear contracts

### Architectural Patterns

- **Contract Interface Pattern**: Lower-level packages define minimal interfaces that higher-level packages implement
- **Capability Composition**: Models compose capabilities from registered formats via configuration
- **Provider-Format Separation**: Providers route protocols to endpoints; capabilities handle request/response formatting
- **Dependency Inversion**: Layers interconnect through interfaces, enabling testing and flexibility

## Current Status

### Protocols (Complete)

All core protocols are fully implemented and operational:

- ✅ **Chat**: Text-based completions with streaming support
- ✅ **Vision**: Visual content analysis (images, PDFs, documents) with structured content
- ✅ **Tools**: Function calling with structured tool definitions (definitions only, execution is supplemental)
- ✅ **Embeddings**: Vector embedding generation for text

### Capability Formats (Complete)

The following capability formats are registered and functional:

- ✅ **openai-chat**: Standard OpenAI chat completions (temperature, top_p, frequency_penalty, etc.)
- ✅ **openai-vision**: Vision with structured content (images, PDFs, documents)
- ✅ **openai-tools**: Function calling (tool definitions and call detection)
- ✅ **openai-embeddings**: Text embedding generation
- ✅ **openai-reasoning**: Reasoning models (o1, o3, etc.) with restricted parameters

### Provider Integrations (Complete)

- ✅ **Ollama**: Local model hosting with OpenAI-compatible endpoints
- ✅ **Azure AI Foundry**: Azure OpenAI with API key and Entra ID authentication

### Infrastructure (Complete)

- ✅ Composable capabilities architecture with protocol-specific configuration
- ✅ Thread-safe capability format registry
- ✅ HTTP transport layer with connection pooling, retries, and streaming
- ✅ Human-readable configuration (duration strings, clear option structures)
- ✅ Protocol-specific response types (ChatResponse, ToolsResponse, EmbeddingsResponse)
- ✅ Command-line testing utility (tools/prompt-agent)

## MVP Completion

### Remaining Work

To reach production-ready MVP status, the following tasks remain:

#### 1. Testing Infrastructure

**Unit Tests** (High Priority)
- [ ] `pkg/protocols`: Message, Request, Response structures and helpers
- [ ] `pkg/capabilities`: Each capability format (chat, vision, tools, embeddings, reasoning)
- [ ] `pkg/models`: Model interface and ProtocolHandler
- [ ] `pkg/providers`: Base provider, Ollama, Azure implementations
- [ ] `pkg/transport`: Client interface and HTTP orchestration
- [ ] `pkg/agent`: Agent interface and protocol methods
- [ ] `pkg/config`: Configuration loading, merging, and validation

**Integration Tests** (High Priority)
- [ ] End-to-end protocol execution with live providers (Ollama, Azure)
- [ ] Streaming response handling
- [ ] Error handling and retry logic
- [ ] Configuration composition and option merging
- [ ] Provider authentication mechanisms

**Test Coverage Goals**
- Minimum 80% code coverage across all packages
- 100% coverage for critical paths (request/response parsing, validation)

#### 2. Code Documentation

**Package Documentation** (High Priority)
- [ ] Add package-level godoc comments to all `pkg/*` packages
- [ ] Document exported types, interfaces, and functions following Go conventions
- [ ] Include usage examples in godoc comments
- [ ] Document design decisions and architectural patterns

**Inline Documentation** (Medium Priority)
- [ ] Add comments for complex logic and non-obvious implementations
- [ ] Document parameter constraints and edge cases
- [ ] Explain rationale for architectural decisions

**Documentation Standards**
- Follow idiomatic Go documentation conventions
- Use complete sentences starting with the declared name
- Include code examples for non-trivial usage patterns
- Document exported functions, types, and constants

## Future Enhancements

These features are outside the current MVP scope but may be added in future releases:

### Document Protocol

**Use Case**: Document analysis and processing (PDFs, Word documents, etc.)
**Estimated Effort**: 4-6 hours

**Rationale**: Document processing has inconsistent support across providers:
- OpenAI: Native PDF support through vision API (up to 100 pages, 32MB)
- Azure OpenAI: No direct PDF support - requires conversion to images or text extraction
- Ollama: No PDF support - images only
- Mistral: Dedicated `/v1/ocr` endpoint with specialized parameters

**Features**:
- New `Document` protocol constant
- Multiple capability formats for different approaches:
  - `openai-document`: Uses vision endpoint with PDF file inputs
  - `mistral-ocr`: Uses Mistral's OCR endpoint with bounding boxes, annotations
  - `azure-document`: Requires pre-processing (PDF → images conversion)
- Request structure: document URL/base64, page selection, extraction options
- Response structure: Extracted text, structured data, bounding boxes (provider-dependent)
- Agent methods: `Document(ctx, prompt, documents []string, opts) (*DocumentResponse, error)`

**Implementation Considerations**:
- Provider-specific handling due to API differences
- May require file conversion utilities for some providers
- Separate from Vision protocol due to different parameters and response structures

### Image Generation Protocol

**Models**: DALL-E, FLUX, Stable Diffusion
**Estimated Effort**: 2-3 hours

**Features**:
- New `Image` protocol constant
- Image generation capability format (`openai-image-generation`)
- Request structure: prompt, size, quality, style, n
- Response structure: URLs or base64-encoded images
- Agent methods: `Generate(ctx, prompt, opts) (*ImageResponse, error)`

### Audio Protocol

**Models**: GPT-4o Audio, Whisper, TTS models
**Estimated Effort**: 3-4 hours

**Features**:
- New `Audio` protocol constant
- Audio capability format (`openai-audio`)
- Multimodal audio + text message content
- Audio input (speech-to-text) and output (text-to-speech)
- Response structure: audio data + transcription
- Agent methods: `AudioChat(ctx, prompt, audioFiles, opts) (*AudioResponse, error)`

### Realtime Protocol

**Use Case**: WebSocket-based bidirectional streaming
**Estimated Effort**: 8-12 hours

**Features**:
- New `Realtime` protocol constant
- WebSocket transport mechanism (separate from HTTP)
- Event-based message protocol
- Session management and state handling
- Bidirectional streaming with low latency

**Note**: This is a specialized protocol requiring significant architectural additions. Recommend deferring until there's clear demand.

### Additional Provider Support

**Anthropic** (Claude models):
- Anthropic-specific capability formats
- Messages API endpoint mapping
- Provider-specific options (max_tokens_to_sample, etc.)

**Google AI** (Gemini models):
- Gemini capability formats
- Google AI Studio API integration
- Content safety settings

**Hugging Face**:
- Inference API integration
- Model hub browsing

## Publishing and Versioning

Once MVP completion is achieved (testing and documentation complete), the library will be published as a pre-release.

### Pre-Release Strategy

**Version Numbering**:
- Pre-release versions: `v0.1.0`, `v0.2.0`, etc.
- Release candidates: `v1.0.0-rc.1`, `v1.0.0-rc.2`, etc.
- Stable release: `v1.0.0`

**Publishing Checklist**:
- [ ] All unit tests passing with 80%+ coverage
- [ ] All integration tests passing
- [ ] Complete package documentation (godoc)
- [ ] README updated with installation and usage instructions
- [ ] ARCHITECTURE.md reflects current implementation
- [ ] CHANGELOG.md created with version history
- [ ] Git tags for semantic versioning
- [ ] GitHub release with release notes

**Versioning Philosophy**:
- Follow [Semantic Versioning 2.0.0](https://semver.org/)
- Pre-1.0 versions (v0.x.x) may introduce breaking changes between minor versions
- Post-1.0 versions guarantee backward compatibility within major versions
- Breaking changes require major version bump (e.g., v1.x.x → v2.0.0)

**Go Module Publishing**:
```bash
# Tag the release
git tag v0.1.0
git push origin v0.1.0

# Go modules will automatically pick up the tagged version
# Users can install with: go get github.com/JaimeStill/go-agents@v0.1.0
```

**Feedback Period**:
- Maintain v0.x.x series for at least 2-3 months
- Gather community feedback on API design
- Make necessary breaking changes before v1.0.0
- Stabilize API surface for v1.0.0 release

**Pre-1.0 Communication**:
- Clearly mark as pre-release in README
- Document expected stability and breaking change policy
- Provide migration guides for breaking changes
- Maintain CHANGELOG with detailed upgrade notes

## Supplemental Package Roadmap

These packages build upon the core `go-agents` library to provide higher-level functionality:

### 1. go-agents-tools

**Purpose**: Tool function registry, execution, and security

**Features**:
- Thread-safe tool registry with function registration
- Built-in tools (calculator, weather, datetime, web search, file operations)
- Tool execution orchestration with multi-turn conversations
- Security policies (allowlist/blocklist, resource limits, timeouts)
- Graceful error handling and partial result support
- Custom tool loading from external packages

**Design Reference**: See `.admin/tools-implementation.md` for detailed implementation guide

**Key Interfaces**:
```go
type ToolRegistry interface {
    Register(name string, fn ToolFunc) error
    Execute(ctx context.Context, name string, args map[string]any) (*ToolResult, error)
    ListTools() []string
}

type ToolFunc func(ctx context.Context, args map[string]any) (any, error)
```

### 2. go-agents-context

**Purpose**: Context window management and conversation memory

**Features**:
- Token counting for different model tokenizers
- Context window optimization (sliding windows, summarization)
- Conversation history management
- Memory strategies (buffer, summary, entity extraction)
- Automatic context pruning based on token limits

**Key Interfaces**:
```go
type ContextManager interface {
    AddMessage(msg protocols.Message) error
    GetMessages() []protocols.Message
    TokenCount() int
    Prune(maxTokens int) error
}

type MemoryStrategy interface {
    Store(ctx context.Context, messages []protocols.Message) error
    Retrieve(ctx context.Context, query string, limit int) ([]protocols.Message, error)
}
```

### 3. go-agents-orchestration

**Purpose**: Multi-agent workflows and coordination patterns

**Features**:

**LangChain-Style Patterns**:
- Sequential chains (LLMChain, TransformChain, etc.)
- Router chains for conditional execution
- Map-reduce patterns for parallel processing

**LangGraph-Style Patterns**:
- State machine-based workflows
- Cyclic graph execution (loops, conditionals)
- Agent collaboration and handoff patterns

**Go Concurrency Patterns**:
- Goroutine-based parallel agent execution
- Channel-based message passing between agents
- Context-based cancellation and timeouts
- Worker pool patterns for batch processing

**Key Interfaces**:
```go
type Workflow interface {
    Execute(ctx context.Context, input any) (any, error)
    ExecuteStream(ctx context.Context, input any) (<-chan WorkflowEvent, error)
}

type StateGraph interface {
    AddNode(name string, fn NodeFunc) error
    AddEdge(from, to string, condition EdgeCondition) error
    Run(ctx context.Context, initialState State) (State, error)
}

type AgentPool interface {
    Submit(ctx context.Context, task Task) (<-chan Result, error)
    Broadcast(ctx context.Context, task Task) ([]Result, error)
}
```
 Custom model deployment support
