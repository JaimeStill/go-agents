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

#### 1. Observability Infrastructure (Prerequisite)

Before implementing comprehensive testing and documentation, foundational observability capabilities must be added to enable execution tracing, decision logging, and confidence scoring in examples and production use.

**Requirements**:
- [ ] Observer interface for protocol execution hooks
- [ ] Request/response metadata capture (request IDs, timing, token usage)
- [ ] Error context enrichment with execution details
- [ ] Optional and composable design (agents function without observers)
- [ ] Foundation for examples-based observability layers

**Design Approach**:
- Observer interface in `pkg/agent/` for protocol lifecycle hooks
- Extension points in lower-level packages as needed
- Minimal performance overhead
- Detailed design to be planned in separate session

**Status**: To be designed and implemented before testing/documentation phases begin.

#### 2. Testing Infrastructure

**Unit Tests** (High Priority)
- [ ] `pkg/protocols`: Message, Request, Response structures and helpers
- [ ] `pkg/capabilities`: Each capability format (chat, vision, tools, embeddings, reasoning)
- [ ] `pkg/models`: Model interface and ProtocolHandler
- [ ] `pkg/providers`: Base provider, Ollama, Azure implementations
- [ ] `pkg/transport`: Client interface and HTTP orchestration
- [ ] `pkg/agent`: Agent interface and protocol methods
- [ ] `pkg/config`: Configuration loading, merging, and validation

**Test Coverage Goals**
- Minimum 80% code coverage across all packages
- 100% coverage for critical paths (request/response parsing, validation)

**Integration Validation**
- Manual validation using `tools/prompt-agent` with live providers
- README examples serve as integration validation
- See `_context/mvp-completion.md` for validation strategy

#### 3. Code Documentation

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

## Examples Roadmap

The examples roadmap establishes production-ready supplemental libraries through examples-driven development. The document classification service serves as the first production use case, driving the creation of three core supplemental packages that extend the go-agents primitive library.

### Overview

**Development Approach**: Build production libraries through standalone, demonstrable examples. Each example validates specific patterns and reveals tooling requirements, preventing premature abstraction while creating reusable infrastructure.

**Driving Use Case**: Document classification service for detecting DoD security classification markings in documents. This real-world requirement guides the examples roadmap and ensures practical utility of developed tooling.

**Deliverables**:
1. **go-agents-document-context**: Universal document → text context conversion
2. **go-agents-orchestration**: Go-native agent coordination with state management
3. **go-agents-services**: HTTP service primitives for agent-backed APIs

### Package Organization

Examples use a structured package layout that mirrors future standalone libraries:

```
examples/
├── pkg/                    # Reusable infrastructure extracted to standalone packages
│   ├── document-context/   # Future: github.com/JaimeStill/go-agents-document-context
│   │   ├── processor.go    # Core processor interface
│   │   ├── registry.go     # Format processor registry
│   │   ├── formats/        # Format-specific extractors
│   │   │   ├── pdf.go      # PDF text extraction
│   │   │   ├── docx.go     # OpenXML .docx processing
│   │   │   ├── xlsx.go     # OpenXML .xlsx processing
│   │   │   ├── pptx.go     # OpenXML .pptx processing
│   │   │   └── image.go    # OCR-based text extraction
│   │   └── context.go      # Context structure and optimization
│   │
│   ├── orchestration/      # Future: github.com/JaimeStill/go-agents-orchestration
│   │   ├── hub/            # Multi-hub coordination
│   │   │   ├── hub.go      # Hub interface and implementation
│   │   │   ├── channel.go  # Message channels
│   │   │   └── registry.go # Agent registration
│   │   ├── messaging/      # Inter-agent messaging
│   │   │   ├── message.go  # Message structures
│   │   │   ├── builder.go  # Message builders
│   │   │   └── filter.go   # Message filtering
│   │   ├── state/          # Workflow state management
│   │   │   ├── graph.go    # State graph execution
│   │   │   ├── state.go    # State structures
│   │   │   └── transition.go # State transitions
│   │   ├── patterns/       # Composition patterns
│   │   │   ├── chain.go    # Sequential chains
│   │   │   ├── parallel.go # Parallel execution
│   │   │   └── router.go   # Conditional routing
│   │   └── observability/  # Execution observability
│   │       ├── trace.go    # Execution trace capture
│   │       ├── decision.go # Decision point logging
│   │       ├── confidence.go # Confidence scoring utilities
│   │       └── metrics.go  # Performance metrics
│   │
│   └── services/           # Future: github.com/JaimeStill/go-agents-services
│       ├── server.go       # HTTP server primitives
│       ├── handler.go      # Agent-backed handlers
│       ├── lifecycle.go    # Agent lifecycle management
│       └── middleware.go   # Service middleware
│
├── 01-document-processor/  # Phase 1: Document processing foundation
├── 02-document-analysis/   # Phase 1: Hybrid Go + agent workflow
├── 03-sequential-chain/    # Phase 2: Agent communication primitives
├── 04-parallel-execution/  # Phase 2: Concurrent agent execution
├── 05-conditional-routing/ # Phase 2: State-driven routing
├── 06-stateful-workflow/   # Phase 2: Complex workflow with state
├── 07-multi-hub-coordination/ # Phase 3: Hierarchical coordination
├── 08-http-agent-endpoint/ # Phase 4: Agent-backed REST API
├── 09-hybrid-workflow-service/ # Phase 4: Go + agent integration
└── 10-document-classification-service/ # Phase 5: Full integration
```

### Phase 1: Document Processing Foundation

**Goal**: Establish reusable document → context conversion capability using pure Go.

**Package Development**: `examples/pkg/document-context/`

#### Example 1: `document-processor/`

**Purpose**: Standalone document → text context conversion tool

**Architecture**:
- Core `Processor` interface in `document-context/processor.go`
- Format-specific implementations in `document-context/formats/`
- Format detection → processor selection → text context output
- Pure Go implementation (no external dependencies where possible)

**Supported Formats**:
- **PDF**: Text extraction using pure Go libraries (pdfcpu or similar)
- **Office Documents**: OpenXML processing for .docx, .xlsx, .pptx
  - Minimal extraction: unzip archive → parse XML → extract text
  - No SDK dependencies, direct OpenXML standard implementation
- **Images**: OCR-based text extraction (gosseract or similar)

**Key Patterns**:
- Registry pattern for format processors
- Clean error handling and validation
- Streaming support for large files

**Learning Objectives**:
- Which formats require special handling?
- What context structure optimizes token usage?
- How to minimize external dependencies?

#### Example 2: `document-analysis/`

**Purpose**: Demonstrate hybrid Go + Agent workflow for document understanding

**Workflow**:
1. **Go**: Document processor extracts text
2. **Go**: Optimize context structure (chunking, metadata extraction)
3. **Agent**: Chat/completion protocol for analysis (not vision!)
4. **Go**: Format and return results

**Key Patterns**:
- When to use Go vs Agent processing
- Context optimization for token efficiency
- Result handling and error boundaries

**Learning Objectives**:
- Optimal context structure for agent consumption
- Token optimization through Go preprocessing
- Error boundary placement in hybrid workflows

### Phase 2: Agent Communication Primitives

**Goal**: Establish Go-native agent coordination patterns with LangGraph-inspired state management.

**Package Development**: `examples/pkg/orchestration/patterns/`, `examples/pkg/orchestration/state/`

**State Management Philosophy**: Leverage Go concurrency primitives (channels, goroutines, contexts) to build state machines and workflow patterns. Explore what's uniquely possible with Go's concurrency model rather than directly porting Python patterns.

#### Example 3: `sequential-chain/`

**Pattern**: Linear workflow with state accumulation

**Agents**: Parser → Enricher → Summarizer

**State Management**:
- State object threaded through agent chain
- Each agent adds to shared state via channels
- Error propagation through the chain

**Go Concurrency**: Channel-based state threading between goroutines

**Learning Objectives**:
- How state flows through goroutine chains
- Error handling in sequential workflows
- Result accumulation patterns

#### Example 4: `parallel-execution/`

**Pattern**: Fan-out/fan-in with state merge

**Agents**: Coordinator + Worker pool

**State Management**:
- Coordinator maintains overall state
- Workers contribute partial results via channels
- State merge using sync primitives

**Go Concurrency**: WaitGroup + channel fan-out/fan-in patterns

**Learning Objectives**:
- Concurrent state updates and synchronization
- Result aggregation from parallel workers
- Timeout and cancellation handling

#### Example 5: `conditional-routing/`

**Pattern**: State-based routing decisions

**Agents**: Router + Specialist handlers

**State Management**:
- Router examines state and selects handler
- State updates reflect routing decisions
- Dynamic handler selection based on state predicates

**Go Concurrency**: Select-based routing, dynamic handler dispatch

**Learning Objectives**:
- State-driven decision making
- Dynamic routing logic
- Handler selection patterns

#### Example 6: `stateful-workflow/`

**Pattern**: Complex workflow with cycles and checkpoints

**Agents**: Multi-step process with loops, retries, and state snapshots

**State Management**:
- Full state machine with transitions
- Checkpointing for recovery
- Cycle detection and loop handling
- Rollback capability

**Go Concurrency**: State machine built on Go concurrency primitives

**Package Development**: `examples/pkg/orchestration/state/` - State graph, transitions, persistence

**Learning Objectives**:
- Complex state machine patterns in Go
- Checkpoint and recovery mechanisms
- Cycle handling and termination

### Phase 3: Multi-Hub Orchestration

**Goal**: Demonstrate hierarchical agent coordination with cross-hub messaging.

**Package Development**: `examples/pkg/orchestration/hub/`, `examples/pkg/orchestration/messaging/`

#### Example 7: `multi-hub-coordination/`

**Pattern**: Hierarchical hub organization with cross-hub messaging

**Hub Architecture**:
- **Global Hub**: System-wide coordination
- **Task Hub**: Workflow management
- **Department Hub**: Team-specific communication

**Agent Roles** (preview of future evolution):
- **Orchestrator**: Supervisor agent registered across all hubs, driving workflows
- **Processors**: Worker agents performing specific tasks in task hub
- **Actors**: Department-specific agents with contextual perspectives

**Key Patterns**:
- Agent registration in multiple hubs with context-aware handlers
- Cross-hub message routing
- Pub/sub within and across hubs
- Hub lifecycle management

**Hierarchical Composition**: Demonstrate how hub pattern supports recursive composition (hubs containing orchestrator agents managing sub-hubs)

**Learning Objectives**:
- Multi-hub coordination patterns
- Cross-hub messaging strategies
- Hierarchical agent organization
- Foundation for recursive composition

**Package Extraction**: Port hub code from research repo (`~/code/go-agents-research/hub/`) to `examples/pkg/orchestration/hub/`

### Phase 4: HTTP Service Integration

**Goal**: Establish patterns for agent-backed HTTP services with containerization.

**Package Development**: `examples/pkg/services/`

**Containerization Requirements**: Each service example includes:
- `Dockerfile` for containerization
- `docker-compose.yml` for local deployment
- `config/` directory with environment-specific configurations
- Configuration via environment variables + config files

#### Example 8: `http-agent-endpoint/`

**Pattern**: REST API with agent processing

**Endpoints**:
- `POST /analyze` - Simple request → agent → response
- `POST /workflow` - Multi-step workflow with state

**Key Patterns**:
- Agent lifecycle in service context (initialization, shutdown)
- Request handling and response formatting
- Connection pooling and resource management
- Configuration-driven agent initialization

**Containerization**:
```bash
docker compose up
curl -X POST http://localhost:8080/analyze -d '{"text": "analyze this"}'
```

**Learning Objectives**:
- Service startup/shutdown patterns
- Agent pooling and reuse
- Error handling in HTTP context
- Configuration management for services

#### Example 9: `hybrid-workflow-service/`

**Pattern**: Go workflow with embedded agent intelligence

**Workflow**:
1. HTTP request with document upload
2. Go: Document processing (extraction)
3. Go: Context optimization
4. Agent: Intelligent analysis
5. Go: Result formatting
6. HTTP response

**Key Patterns**:
- Best-of-both-worlds: Go for efficiency, agents for intelligence
- Minimize token costs through Go preprocessing
- Workflow orchestration across Go and agent processing

**Containerization**:
```bash
docker compose up
curl -X POST http://localhost:8080/process \
  -F "file=@document.pdf" \
  -F "operation=summarize"
```

**Learning Objectives**:
- Go vs Agent decision boundaries
- Token cost optimization
- Production workflow patterns
- Resource management in containerized services

### Phase 5: Full Integration

**Goal**: Demonstrate complete integration of all developed tooling in production service.

#### Example 10: `document-classification-service/`

**Use Case**: Detect DoD security classification markings in documents

**Classification Types**: Unclassified, Confidential, Secret, Top Secret

**Architecture Integration**:
- **Document Processing**: `document-context` package for text extraction
- **Sequential Chain**: Parser → Classifier → Validator workflow
- **Conditional Routing**: Different validators per classification level
- **Multi-Hub Orchestration**: Audit hub logs all classification decisions
- **HTTP Service**: REST API for document upload and classification
- **Observability**: Full execution trace, confidence scoring, decision logging

**Service Workflow**:
1. HTTP endpoint receives document upload
2. Document processor extracts text (Go)
3. Sequential agent chain processes document:
   - Parser: Extract potential classification markers
   - Classifier: Determine classification level
   - Validator: Verify classification evidence
4. Conditional routing: Route to classification-specific validator
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

**Learning Objectives**:
- Production integration patterns
- Observability in production workflows
- Confidence scoring and evidence tracking
- Audit trail requirements
- Performance optimization

### Agent Roles Evolution

While not implemented in initial examples, the architecture supports evolution toward specialized agent roles:

**Orchestrator** (Example 7+): Supervisory agents driving workflows, registered in multiple hubs

**Processor** (Examples 3-6): Functional agents with clear input→output contracts, stateless where possible

**Actor** (Future): Profile-based agents with perspectives, experience accumulation, contextual memory

**Recursive Composition**: Hub architecture designed to enable hubs containing orchestrator agents managing sub-hubs, creating fractal organizational patterns.

### Key Architectural Decisions to Validate

Through examples development, validate:

1. **State Management**: Are Go channels + contexts sufficient, or do we need more sophisticated state machines?
2. **Hub Scalability**: Can the hub pattern support recursive composition (hubs containing orchestrator agents managing sub-hubs)?
3. **Observability Overhead**: How much observability can we add without impacting production performance?
4. **Configuration Flexibility**: Can services be reconfigured for different environments (air-gapped, cloud, hybrid) through config alone?
5. **Token Optimization**: How much Go preprocessing reduces token costs while maintaining quality?
6. **OpenXML Processing**: Can we implement minimal document extraction without external SDKs?
7. **Go Concurrency Patterns**: What unique state management patterns emerge from Go's concurrency model?

### Supplemental Package Extraction

Once examples validate patterns, extract to standalone libraries:

**go-agents-document-context**:
- Migrate `examples/pkg/document-context/` to standalone repo
- Publish as independent library for document processing
- Reusable across any Go application (not agent-specific)

**go-agents-orchestration**:
- Migrate `examples/pkg/orchestration/` to standalone repo
- Publish as agent coordination library
- Provides hub, messaging, state, and pattern primitives

**go-agents-services**:
- Migrate `examples/pkg/services/` to standalone repo
- Publish as HTTP service primitives for agent-backed APIs
- Containerization templates and patterns

**Documentation Classification Service**:
- Reference implementation using all three libraries
- Production-ready example for enterprise adoption
- Demonstrates full capability of tooling ecosystem

## Future Enhancements

These features and capabilities are envisioned for development after the Examples Roadmap completes:

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

### Publishing and Versioning

Once the Examples Roadmap completes and supplemental packages are validated, the library ecosystem will be published.

### Pre-Release Strategy

**Version Numbering**:
- Pre-release versions: `v0.1.0`, `v0.2.0`, etc.
- Release candidates: `v1.0.0-rc.1`, `v1.0.0-rc.2`, etc.
- Stable release: `v1.0.0`

**Publishing Checklist**:
- [ ] All unit tests passing with 80%+ coverage
- [ ] README examples verified via `tools/prompt-agent`
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

### Additional Supplemental Packages

These packages extend the go-agents ecosystem beyond the three core libraries developed through the Examples Roadmap:

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
