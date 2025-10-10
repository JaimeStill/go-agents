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

### Completion Status

The MVP is **complete** and production-ready. All core functionality, testing infrastructure, and documentation have been implemented and validated.

#### 1. Testing Infrastructure ✅ **Complete**

**Unit Tests** (Complete)
- ✅ `pkg/config`: Configuration loading, merging, and validation (95.1% coverage)
- ✅ `pkg/protocols`: Message, Request, Response structures and helpers (100.0% coverage)
- ✅ `pkg/capabilities`: Each capability format - chat, vision, tools, embeddings, reasoning (83.1% coverage)
- ✅ `pkg/models`: Model interface and ProtocolHandler (88.6% coverage)
- ✅ `pkg/providers`: Base provider, Ollama, Azure implementations (81.9% coverage)
- ✅ `pkg/transport`: Client interface and HTTP orchestration (83.0% coverage)
- ✅ `pkg/agent`: Agent interface and protocol methods (88.3% coverage)

**Test Coverage Achieved**
- ✅ Overall coverage: **89.3%** (exceeds 80% minimum requirement)
- ✅ Critical paths have excellent coverage (protocols at 100%)
- ✅ 20 test files with comprehensive test cases
- ✅ Black-box testing approach using `package_test` suffix

**Integration Validation**
- ✅ Manual validation strategy established using `tools/prompt-agent`
- ✅ README examples serve as integration validation cases
- ✅ All protocols tested (chat, vision, tools, embeddings)
- ✅ Both providers tested (Ollama, Azure)

#### 2. Code Documentation ✅ **Complete**

**Package Documentation** (Complete)
- ✅ Package-level godoc comments added to all `pkg/*` packages
- ✅ All exported types, interfaces, and functions documented
- ✅ Usage examples included in godoc comments
- ✅ Design decisions and architectural patterns documented

**Documentation Quality**
- ✅ Follows idiomatic Go documentation conventions
- ✅ Complete sentences starting with declared names
- ✅ Code examples for non-trivial usage patterns
- ✅ All exported functions, types, and constants documented

**Documentation Verification**
- ✅ All packages validated with `go doc` commands
- ✅ Documentation is clear, comprehensive, and accurate

### Next Phase: Pre-Release Publishing and Supplemental Package Development

With MVP completion achieved, the project is ready for initial publication and supplemental package development. The go-agents library will be published as a pre-release (v0.1.0), enabling development of three independent supplemental packages:
- `go-agents-orchestration`: Agent coordination and workflow management
- `go-agents-document-context`: Document processing and text extraction
- `go-agents-services`: HTTP service primitives for agent-backed APIs

## Publishing and Versioning

### Version Numbering Strategy

The go-agents library follows [Semantic Versioning 2.0.0](https://semver.org/) with a deliberate pre-release phase to validate API design through real-world usage before committing to long-term stability.

**Pre-Release Versions (v0.x.x)**:
- Initial release: `v0.1.0`
- Minor version increments for feature additions and improvements
- Breaking changes allowed between minor versions during pre-release phase
- Version format: `v0.MINOR.PATCH`

**Release Candidates (v1.0.0-rc.x)**:
- Signal approaching API stability
- Final opportunity for breaking changes before v1.0.0
- Thorough validation and community feedback
- Version format: `v1.0.0-rc.1`, `v1.0.0-rc.2`, etc.

**Stable Release (v1.0.0+)**:
- API stability commitment: no breaking changes within major version
- Breaking changes require major version increment (v1.x.x → v2.0.0)
- Backward compatibility guaranteed within major version
- Predictable upgrade path for library consumers

### Pre-Release Philosophy and Approach

**Purpose of Pre-Release Phase**:
- Validate API design through supplemental package development
- Gather feedback from real-world usage before stabilization
- Identify missing capabilities and awkward patterns
- Refine interfaces based on integration experience
- Build confidence in architectural decisions

**Development Through Usage**:
The pre-release phase enables validation through building supplemental packages (`go-agents-orchestration`, `go-agents-document-context`, `go-agents-services`) that depend on the published go-agents library. This approach:
- Exercises the public API from a consumer perspective
- Reveals integration friction and missing abstractions
- Validates that the library delivers on its primitive interface promise
- Demonstrates the componentized platform vision through real packages
- Builds practical experience with Go package lifecycle management

**Breaking Change Policy (v0.x.x)**:
During the pre-release phase, breaking changes are acceptable but managed carefully:
- Breaking changes documented in CHANGELOG.md with upgrade guidance
- Migration examples provided for significant API changes
- Breaking changes batched when possible to minimize disruption
- Clear communication about stability expectations
- Community feedback actively sought before major restructuring

**Duration and Criteria**:
The pre-release phase continues until:
- Supplemental packages validate all core abstractions
- No major architectural concerns identified during usage
- API surface feels natural and complete for primitive interface library
- Documentation comprehensively covers all capabilities
- Test coverage remains above 80% across all changes
- Community feedback incorporated and addressed
- Minimum 2-3 months of real-world usage and iteration

### Go Module Publishing

After merging the PR into main, follow these steps:

1. Ensure Clean Repository State

```sh
# Pull latest main
git checkout main
git pull origin main

# Verify no uncommitted changes
git status
```

2. Tag the Release

```sh
# Create and push the version tag
git tag v0.1.0
git push origin v0.1.0
```

3. Create GitHub Release

  1. Navigate to https://github.com/JaimeStill/go-agents/releases/new
  2. Select tag: `v0.1.0`
  3. Title: `v0.1.0 - [Heading]`
  4. Description: Copy from CHANGELOG.md
  5. Check: "Set as a pre-release"
  6. Click: "Publish release"

4. Verify Publication

```sh
# Verify the module is indexed (may take a few minutes)
go list -m -versions github.com/JaimeStill/go-agents

# Should show: v0.1.0 v0.1.0
```

5. Test Installation

```sh
# In a separate test directory
go get github.com/JaimeStill/go-agents@v0.1.0
```

**Version Discovery**:
- Go module proxy indexes all tagged versions
- Users can browse versions: `go list -m -versions github.com/JaimeStill/go-agents`
- Specific version installation: `go get github.com/JaimeStill/go-agents@v0.2.0`
- Latest pre-release: `go get github.com/JaimeStill/go-agents@latest`

**Module Maintenance**:
- Each version immutable once published
- Breaking changes require new version
- CHANGELOG.md maintained with detailed version history
- GitHub releases created for each version with release notes
- Migration guides provided for breaking changes

### Publishing Checklist

Before publishing any version, ensure:

**Code Quality**:
- [ ] All unit tests passing with 80%+ coverage
- [ ] No critical bugs or known issues
- [ ] Code review completed for all changes since last version
- [ ] Go vet and staticcheck clean
- [ ] All deprecated features documented with migration path

**Documentation**:
- [ ] README updated with current capabilities and examples
- [ ] README examples verified via `tools/prompt-agent`
- [ ] Complete package documentation (godoc)
- [ ] ARCHITECTURE.md reflects current implementation
- [ ] CHANGELOG.md updated with version changes
- [ ] Migration guide provided for breaking changes (if applicable)

**Repository State**:
- [ ] All changes committed to main branch
- [ ] Git repository clean (no uncommitted changes)
- [ ] Git tags follow semantic versioning
- [ ] GitHub release created with detailed release notes

**Communication**:
- [ ] Pre-release status clearly marked in README
- [ ] Breaking change policy documented
- [ ] Expected stability communicated clearly
- [ ] Feedback channels established (GitHub issues)

### Communication Strategy

**Pre-Release Transparency**:
The library is clearly marked as pre-release in all documentation:

```markdown
## Status: Pre-Release (v0.x.x)

**go-agents** is currently in pre-release development. The API may change between minor versions until v1.0.0 is released. Production use is supported, but be prepared for potential breaking changes.

We actively seek feedback on API design, missing capabilities, and integration challenges. Please open issues for any concerns or suggestions.
```

**Version Communication**:
- Each release includes detailed changelog
- Breaking changes highlighted prominently
- Migration examples provided inline
- Deprecation warnings added to code with removal version noted
- Community notified through GitHub releases

**Feedback Channels**:
- GitHub issues for bug reports and feature requests
- GitHub discussions for API design conversations
- Detailed issue templates for structured feedback
- Regular review of integration challenges from supplemental packages

### Promotion to v1.0.0

**Stability Criteria**:
The library graduates to v1.0.0 when:
- All three supplemental packages developed and validated
- No significant API concerns identified during supplemental package development
- Community feedback addressed and incorporated
- API surface complete for primitive interface library scope
- Documentation comprehensive and accurate
- Test coverage maintained above 80%
- Minimum 2-3 months of pre-release usage without major issues
- Confidence in long-term API stability commitment

**Post-1.0.0 Commitment**:
Once v1.0.0 is released:
- API stability guaranteed within major version
- Breaking changes only in major version increments
- Backward compatibility maintained for v1.x.x series
- Deprecation cycle for any API changes (minimum one minor version)
- Predictable upgrade path for consumers

## Supplemental Package Development Roadmap

The supplemental package roadmap establishes three independent, production-ready libraries that extend the go-agents primitive interface library. Each package is developed as a standalone repository depending only on the published go-agents library.

### Overview

**Development Approach**: Build supplemental packages as independent repositories, each depending solely on the published go-agents library. This approach validates the go-agents API design through real-world usage, ensures each package can be used independently or composed together, and builds practical experience with Go package lifecycle management.

**Development Sequence**: Orchestration → Document Processing → Services. The sequence reflects practical development priorities rather than dependency chains—each package depends only on go-agents.

**Integration Demonstration**: Document classification service serves as a reference implementation showing how all packages compose together in a production application.

**Supplemental Packages**:
1. **go-agents-orchestration**: Agent coordination, workflow management, and state machines
2. **go-agents-document-context**: Universal document → text context conversion
3. **go-agents-services**: HTTP service primitives for agent-backed APIs

### Package Organization

Each supplemental package is structured as an independent Go module with its own repository:

```
github.com/JaimeStill/go-agents-orchestration/
├── hub/                    # Multi-hub coordination
│   ├── hub.go              # Hub interface and implementation
│   ├── channel.go          # Message channels
│   └── registry.go         # Agent registration
├── messaging/              # Inter-agent messaging
│   ├── message.go          # Message structures
│   ├── builder.go          # Message builders
│   └── filter.go           # Message filtering
├── state/                  # Workflow state management
│   ├── graph.go            # State graph execution
│   ├── state.go            # State structures
│   └── transition.go       # State transitions
├── patterns/               # Composition patterns
│   ├── chain.go            # Sequential chains
│   ├── parallel.go         # Parallel execution
│   └── router.go           # Conditional routing
└── observability/          # Execution observability
    ├── trace.go            # Execution trace capture
    ├── decision.go         # Decision point logging
    ├── confidence.go       # Confidence scoring utilities
    └── metrics.go          # Performance metrics

github.com/JaimeStill/go-agents-document-context/
├── processor.go            # Core processor interface
├── registry.go             # Format processor registry
├── formats/                # Format-specific extractors
│   ├── pdf.go              # PDF text extraction
│   ├── docx.go             # OpenXML .docx processing
│   ├── xlsx.go             # OpenXML .xlsx processing
│   ├── pptx.go             # OpenXML .pptx processing
│   └── image.go            # OCR-based text extraction
└── context.go              # Context structure and optimization

github.com/JaimeStill/go-agents-services/
├── server.go               # HTTP server primitives
├── handler.go              # Agent-backed handlers
├── lifecycle.go            # Agent lifecycle management
└── middleware.go           # Service middleware
```

### go-agents-orchestration

**Repository**: `github.com/JaimeStill/go-agents-orchestration`
**Dependency**: `github.com/JaimeStill/go-agents@v0.x.x`
**Development Priority**: First (enables agent coordination patterns)

**Purpose**: Provide Go-native agent coordination primitives with LangGraph-inspired state management, multi-hub messaging architecture, and composable workflow patterns.

**Core Capabilities**:

**Hub Architecture**:
- Multi-hub coordination with hierarchical organization
- Agent registration across multiple hubs with context-aware handlers
- Cross-hub message routing and pub/sub patterns
- Hub lifecycle management (initialization, shutdown, cleanup)
- Foundation for recursive composition (hubs containing orchestrator agents managing sub-hubs)
- Port hub implementation from `go-agents-research` repository

**Messaging Primitives**:
- Structured message types for inter-agent communication
- Message builders for constructing complex communications
- Message filtering and routing logic
- Channel-based message delivery using Go concurrency

**State Management**:
- State graph execution with transitions and predicates
- State structure definitions and mutation patterns
- Checkpointing for recovery and rollback
- Cycle detection and loop handling
- Leverage Go concurrency primitives (channels, goroutines, contexts)
- Explore Go-native patterns rather than directly porting Python approaches

**Workflow Patterns**:
- **Sequential chains**: Linear workflows with state accumulation
- **Parallel execution**: Fan-out/fan-in with state merge and result aggregation
- **Conditional routing**: State-based routing decisions with dynamic handler selection
- **Stateful workflows**: Complex state machines with cycles, retries, and checkpoints

**Observability Infrastructure**:
- Execution trace capture across workflow steps
- Decision point logging with reasoning
- Confidence scoring utilities for agent outputs
- Performance metrics (token usage, timing, retries)
- Designed for production debugging and optimization

**Agent Role Abstractions**:
- **Orchestrator**: Supervisory agents driving workflows, registered in multiple hubs
- **Processor**: Functional agents with clear input→output contracts
- **Actor**: Profile-based agents with perspectives (foundation for future expansion)

**Key Architectural Questions**:
- State management: Are Go channels + contexts sufficient, or do we need more sophisticated state machines?
- Hub scalability: Can the hub pattern support recursive composition?
- Observability overhead: How much observability can we add without impacting performance?
- Go concurrency patterns: What unique state management patterns emerge from Go's concurrency model?

### go-agents-document-context

**Repository**: `github.com/JaimeStill/go-agents-document-context`
**Dependency**: `github.com/JaimeStill/go-agents@v0.x.x`
**Development Priority**: Second (provides document processing for integration scenarios)

**Purpose**: Universal document → text context conversion for LLM consumption. Extract text from various document formats using pure Go implementations, optimize context structure for token efficiency, and provide clean interfaces for document processing pipelines.

**Core Capabilities**:

**Document Processing**:
- Core `Processor` interface for format-agnostic document handling
- Format detection and automatic processor selection
- Registry pattern for extensible format support
- Streaming support for large documents
- Pure Go implementations (minimize external dependencies)

**Supported Formats**:
- **PDF**: Text extraction using pure Go libraries (pdfcpu or similar)
- **Office Documents**: OpenXML processing for .docx, .xlsx, .pptx
  - Minimal extraction approach: unzip archive → parse XML → extract text
  - Direct OpenXML standard implementation (no SDK dependencies)
  - Support for structured data extraction from spreadsheets and presentations
- **Images**: OCR-based text extraction (gosseract or similar)
  - Integration with Tesseract for optical character recognition
  - Support for common image formats (PNG, JPEG, TIFF)

**Context Optimization**:
- Context structure design for LLM consumption
- Chunking strategies for long documents
- Metadata extraction (title, author, creation date, page count)
- Token usage optimization through preprocessing
- Format-specific context enhancement (tables, lists, structure preservation)

**Integration Patterns**:
- Standalone usage (document processing without agents)
- Hybrid workflows (Go preprocessing + agent analysis)
- When to extract text vs. when to use vision protocols
- Error boundaries for document processing failures

**Key Architectural Questions**:
- Which formats require special handling beyond basic text extraction?
- What context structure optimizes token usage while preserving meaning?
- How to minimize external dependencies while maintaining quality?
- OpenXML processing: Can we implement minimal document extraction without external SDKs?
- Token optimization: How much Go preprocessing reduces token costs while maintaining quality?

### go-agents-services

**Repository**: `github.com/JaimeStill/go-agents-services`
**Dependency**: `github.com/JaimeStill/go-agents@v0.x.x`
**Development Priority**: Third (enables production HTTP service patterns)

**Purpose**: HTTP service primitives for building agent-backed REST APIs. Provides patterns for agent lifecycle management in service context, containerization templates, and integration patterns for production deployments.

**Core Capabilities**:

**HTTP Server Primitives**:
- Server initialization and configuration
- Graceful shutdown with agent cleanup
- Health check and readiness endpoints
- Request/response handling patterns
- Middleware chains for cross-cutting concerns

**Agent Lifecycle Management**:
- Agent initialization at service startup
- Connection pooling and resource management
- Agent reuse across requests
- Cleanup on service shutdown
- Error recovery and circuit breaking

**Handler Patterns**:
- Simple request → agent → response pattern
- Multi-step workflow with state management
- File upload handling (documents, images)
- Streaming response support
- Error handling and response formatting

**Containerization**:
- `Dockerfile` templates for production deployment
- `docker-compose.yml` for local development
- Multi-stage builds for minimal image size
- Configuration via environment variables + config files
- Health checks and container orchestration support

**Configuration Management**:
- Environment-specific configuration
- Secret management (API keys, credentials)
- Configuration validation at startup
- Runtime configuration updates (where applicable)
- Configuration for different environments (air-gapped, cloud, hybrid)

**Integration Patterns**:
- REST API with agent processing
- Hybrid workflows (Go preprocessing + agent intelligence)
- Document processing services (upload → process → analyze)
- Multi-agent workflow coordination
- Observability integration (logging, metrics, tracing)

**Key Architectural Questions**:
- Agent lifecycle: How to manage agents in long-running services?
- Connection pooling: How to reuse agent connections efficiently?
- Error handling: What error boundaries prevent cascading failures?
- Configuration flexibility: Can services be reconfigured for different environments through config alone?
- Resource management: How to handle resource limits in containerized services?

### Integration Reference Implementation

The document classification service serves as a reference implementation demonstrating how all supplemental packages compose together in a production application.

**Repository**: `github.com/JaimeStill/document-classification-service`
**Dependencies**:
- `github.com/JaimeStill/go-agents@v0.x.x`
- `github.com/JaimeStill/go-agents-orchestration@v0.x.x`
- `github.com/JaimeStill/go-agents-document-context@v0.x.x`
- `github.com/JaimeStill/go-agents-services@v0.x.x`

**Use Case**: Detect DoD security classification markings in documents

**Classification Types**: Unclassified, Confidential, Secret, Top Secret

**Architecture Integration**:
- **Document Processing**: `go-agents-document-context` for text extraction from uploaded documents
- **Sequential Chain**: Parser → Classifier → Validator workflow using `go-agents-orchestration` patterns
- **Conditional Routing**: State-based routing to classification-specific validators using orchestration router pattern
- **Multi-Hub Orchestration**: Audit hub logs all classification decisions using hub coordination primitives
- **HTTP Service**: REST API built with `go-agents-services` for document upload and classification
- **Observability**: Full execution trace, confidence scoring, decision logging via orchestration observability infrastructure

**Service Workflow**:
1. HTTP endpoint receives document upload
2. `go-agents-document-context` extracts text (Go preprocessing)
3. Sequential agent chain processes document using `go-agents-orchestration`:
   - Parser: Extract potential classification markers
   - Classifier: Determine classification level using `go-agents`
   - Validator: Verify classification evidence
4. Conditional routing: Route to classification-specific validator based on state
5. Multi-hub audit: Log classification decision to audit hub
6. HTTP response with classification + confidence + evidence + trace

**Response Structure**:
```json
{
  "classification": "SECRET",
  "confidence": 0.95,
  "evidence": {
    "banner_marking": "SECRET//NOFORN",
    "portion_markings": ["(S)", "(S)"],
    "header_footer": true
  },
  "trace": {
    "extraction_time_ms": 245,
    "parser_tokens": 1523,
    "classifier_tokens": 892,
    "validator_tokens": 634,
    "total_time_ms": 3421,
    "decisions": [...]
  }
}
```

**Containerization**:
```bash
docker compose up
curl -X POST http://localhost:8080/classify -F "file=@classified-doc.pdf"
```

**Demonstrates**:
- Production integration patterns across all packages
- Observability in production workflows
- Confidence scoring and evidence tracking
- Audit trail requirements
- Performance optimization through Go preprocessing
- Composition of independent packages
- Real-world value of the go-agents ecosystem


## Future Enhancements

These features and capabilities are envisioned for development after the supplemental package development roadmap completes:

### Developer Toolkit (Stretch Goal)

**Vision**: Empower developers and technical users to build powerful agents and agentic workflows using the complete go-agents ecosystem.

**Target Audience**: Developers, ML engineers, and technical users seeking production-ready tooling for building agent-powered services. Provides an alternative to notebook-centric development with service-oriented, cloud-native patterns.

**Capabilities**:
- Configuration-driven agent creation and orchestration
- Visual workflow design tools (optional)
- Template library for common patterns
- Testing and debugging utilities
- Performance profiling and optimization tools
- Documentation generator for agentic workflows

**Value Proposition**: Enable technical users to contribute production-ready, cloud-native services using agents without requiring deep Go expertise. Bridge the gap between prototyping and production deployment.

### Additional Protocols

These protocol extensions expand the library's capabilities to additional interaction modes:

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

### Additional Supplemental Packages

These packages extend the go-agents ecosystem beyond the three core libraries developed through the supplemental package development roadmap:

#### go-agents-tools

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

#### go-agents-context

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
