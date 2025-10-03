# Tests and Documentation Execution Plan

## Overview

This document outlines the direct execution approach for completing MVP testing and documentation requirements as specified in `mvp-completion.md`. Rather than providing a step-by-step implementation guide for manual execution, this approach involves direct code generation with structured review gates.

## Execution Strategy

### Core Principles

1. **Incremental Development**: Complete one package at a time rather than all at once
2. **Review Gates**: Mandatory review checkpoints after tests and documentation for each package
3. **Dependency Order**: Progress from lowest to highest in the package hierarchy
4. **Early Verification**: Run tests and generate coverage reports immediately after writing
5. **Error Prevention**: Review gates prevent error compounding across packages

### Why Direct Execution

**This approach is appropriate because:**
- ✅ Testing patterns are standardized (table-driven tests, httptest.Server mocking)
- ✅ Documentation conventions are well-established (Go idiomatic godoc)
- ✅ Architecture is well-understood by both parties
- ✅ mvp-completion.md provides comprehensive test case specifications
- ✅ No novel design decisions required during implementation
- ✅ Significantly faster than guide-based execution

**Risk mitigation:**
- Review gates after each package prevent cascading errors
- Coverage reports provide objective quality metrics
- Test execution verifies correctness before moving forward

---

## Package Execution Order

Following the dependency hierarchy from `mvp-completion.md`:

```
1. pkg/config/       (foundation - no internal dependencies)
2. pkg/protocols/    (no dependencies)
3. pkg/capabilities/ (depends on: protocols)
4. pkg/models/       (depends on: config, protocols, capabilities)
5. pkg/providers/    (depends on: config, protocols, capabilities, models)
6. pkg/transport/    (depends on: config, protocols, capabilities, models, providers)
7. pkg/agent/        (depends on: all previous)
```

**Rationale**: Lower-level packages must be tested and documented before higher-level packages that depend on them.

---

## Standard Workflow Pattern

Each package follows this cycle:

### Phase A: Testing

1. **Read existing implementation files**
   - Understand current code structure
   - Identify all exported types, functions, and methods
   - Note complex logic that requires testing

2. **Write unit tests**
   - Create test files following Go conventions (`*_test.go`)
   - Implement test cases specified in mvp-completion.md
   - Use table-driven tests where appropriate
   - Use httptest.Server for HTTP mocking
   - Follow idiomatic Go test patterns

3. **Verify tests**
   - Run `go test ./pkg/[package]/...` to verify tests pass
   - Generate coverage report: `go test ./pkg/[package]/... -coverprofile=coverage.out`
   - Verify coverage meets 80% minimum threshold
   - Check critical paths have 100% coverage

4. **🚦 REVIEW GATE A**: User reviews test implementation
   - Tests pass successfully
   - Coverage meets thresholds
   - Test quality and completeness
   - Approve before proceeding to documentation

### Phase B: Documentation

5. **Write package documentation**
   - Add package-level godoc comment to primary file
   - Document all exported types, interfaces, functions, methods
   - Include usage examples for non-trivial APIs
   - Add inline comments for complex logic
   - Follow Go documentation conventions (start with declared name)

6. **Verify documentation**
   - Run `go doc [package]` to check output readability
   - Verify all exports are documented
   - Check for spelling/grammar errors
   - Ensure examples are clear and accurate

7. **🚦 REVIEW GATE B**: User reviews documentation
   - Documentation completeness
   - Clarity and accuracy
   - Godoc output quality
   - Approve before proceeding to next package

### Phase C: Move Forward

8. **Mark package complete**
   - Update progress tracking
   - Move to next package in dependency order

---

## Package-Specific Execution Plans

### Session 1: pkg/config/

**Files to test**: `duration.go`, `options.go`, `model.go`, `provider.go`, `transport.go`, `agent.go`

**Test files to create**:
- `duration_test.go`
- `options_test.go`
- `model_test.go`
- `provider_test.go`
- `transport_test.go`
- `agent_test.go`

**Key test focus areas**:
- Duration parsing (string format: "24s", "1m", "2h")
- Duration marshaling/unmarshaling
- Configuration structure validation
- JSON serialization/deserialization

**Documentation focus**:
- Package-level comment explaining configuration system
- Document Duration type with examples
- Document all config structures (ModelConfig, ProviderConfig, etc.)
- Explain human-readable duration support

**Estimated effort**: 2-3 hours

---

### Session 2: pkg/protocols/

**Files to test**: `protocol.go`

**Test files to create**:
- `protocol_test.go`

**Key test focus areas**:
- Protocol constants validation (Chat, Vision, Tools, Embeddings)
- Message creation and content handling
- Request/Response structures
- StreamingChunk parsing
- ExtractOption helper function

**Documentation focus**:
- Package-level comment explaining protocol system
- Document Protocol type and constants
- Document Message, Request, Response types
- Document streaming structures
- Include usage examples

**Estimated effort**: 2-3 hours

---

### Session 3: pkg/capabilities/

**Files to test**: `capability.go`, `registry.go`, `chat.go`, `vision.go`, `tools.go`, `embeddings.go`

**Test files to create**:
- `capability_test.go`
- `registry_test.go`
- `chat_test.go`
- `vision_test.go`
- `tools_test.go`
- `embeddings_test.go`

**Key test focus areas**:
- StandardCapability option validation and processing
- StandardStreamingCapability streaming support
- Registry thread-safety
- Each capability format (OpenAI chat, vision, tools, embeddings, reasoning)
- Request creation and response parsing

**Documentation focus**:
- Package-level comment explaining capability system
- Document Capability interface
- Document capability format registry
- Document each capability implementation
- Include registration examples

**Estimated effort**: 4-6 hours

---

### Session 4: pkg/models/

**Files to test**: `model.go`, `handler.go`

**Test files to create**:
- `model_test.go`
- `handler_test.go`

**Key test focus areas**:
- Model creation from config
- Protocol support queries
- Capability retrieval
- Option management (get, update, merge)
- ProtocolHandler functionality

**Documentation focus**:
- Package-level comment explaining model abstraction
- Document Model interface with examples
- Document ProtocolHandler
- Explain option merging behavior
- Include configuration examples

**Estimated effort**: 2-3 hours

---

### Session 5: pkg/providers/

**Files to test**: `provider.go`, `base.go`, `registry.go`, `ollama.go`, `azure.go`

**Test files to create**:
- `base_test.go`
- `registry_test.go`
- `ollama_test.go`
- `azure_test.go`

**Key test focus areas**:
- BaseProvider basic functionality
- Provider registry creation
- Ollama endpoint mapping and request/response handling
- Azure endpoint mapping with deployment URLs
- Azure authentication (API key vs bearer token)
- Mock HTTP responses using httptest.Server

**Documentation focus**:
- Package-level comment explaining provider system
- Document Provider interface
- Document provider registry
- Document Ollama and Azure implementations
- Explain authentication mechanisms

**Estimated effort**: 3-4 hours

---

### Session 6: pkg/transport/

**Files to test**: `client.go`

**Test files to create**:
- `client_test.go`

**Key test focus areas**:
- Client creation from config
- Protocol execution (non-streaming)
- Streaming protocol execution
- Option merging and validation
- HTTP client configuration (timeout, connection pooling)
- Health tracking
- Error handling
- Mock provider responses using httptest.Server

**Documentation focus**:
- Package-level comment explaining transport orchestration
- Document Client interface
- Explain execution flow
- Document health tracking
- Include usage examples

**Estimated effort**: 2-3 hours

---

### Session 7: pkg/agent/

**Files to test**: `agent.go`, `errors.go`

**Test files to create**:
- `agent_test.go`
- `errors_test.go` (if custom errors exist)

**Key test focus areas**:
- Agent creation from config
- Chat/ChatStream methods
- Vision/VisionStream methods
- Tools method
- Embed method
- System prompt injection
- Message initialization
- Mock transport.Client interface

**Documentation focus**:
- Package-level comment explaining agent orchestration
- Document Agent interface with comprehensive examples
- Document each protocol method
- Explain message initialization
- Include complete usage examples for all protocols

**Estimated effort**: 1-2 hours

---

## Success Criteria

### Phase 1: Testing Complete

- [ ] All 7 packages have corresponding test files
- [ ] All tests pass: `go test ./...`
- [ ] Overall coverage reaches 80% minimum
- [ ] Critical paths (parsing, validation, option merging) have 100% coverage
- [ ] Coverage report generated: `go test ./... -coverprofile=coverage.out`
- [ ] Coverage reviewed: `go tool cover -func=coverage.out`

### Phase 2: Documentation Complete

- [ ] All 7 packages have godoc comments
- [ ] All exported types documented
- [ ] All exported functions/methods documented
- [ ] Complex logic has inline comments
- [ ] Usage examples included for non-trivial APIs
- [ ] `go doc` produces readable output for all packages
- [ ] Documentation verified via local godoc server (http://localhost:6060)

### MVP Complete

- [ ] Both phases completed
- [ ] Final coverage report reviewed
- [ ] Documentation accuracy verified
- [ ] README examples work with prompt-agent (validation)
- [ ] Ready to proceed with publishing checklist

---

## Review Gate Checklist

### After Tests (Review Gate A)

**Verifications:**
- [ ] All test files compile without errors
- [ ] All tests pass (`go test ./pkg/[package]/...`)
- [ ] Coverage meets 80% minimum threshold
- [ ] Critical paths have 100% coverage
- [ ] Tests follow Go idioms (table-driven, clear naming)
- [ ] Mocks are appropriate (httptest.Server where needed)
- [ ] Test cases match mvp-completion.md specifications

**Questions to ask:**
1. Are there any edge cases not covered?
2. Are error paths adequately tested?
3. Do tests verify both success and failure scenarios?
4. Are concurrent operations tested where relevant?

### After Documentation (Review Gate B)

**Verifications:**
- [ ] Package-level comment present and clear
- [ ] All exported types documented
- [ ] All exported functions/methods documented
- [ ] Documentation starts with declared name
- [ ] Usage examples present for complex APIs
- [ ] Cross-references use `[TypeName]` syntax
- [ ] `go doc` output is readable
- [ ] No spelling/grammar errors

**Questions to ask:**
1. Is the package purpose clear?
2. Are usage examples helpful and correct?
3. Are parameter constraints documented?
4. Are error conditions explained?

---

## Commands Reference

### Testing

```bash
# Run tests for specific package
go test ./pkg/config/... -v

# Run all tests
go test ./... -v

# Generate coverage for specific package
go test ./pkg/config/... -coverprofile=coverage.out

# Generate coverage for all packages
go test ./... -coverprofile=coverage.out

# View coverage summary
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

### Documentation

```bash
# View package documentation
go doc pkg/config

# View type documentation
go doc pkg/config.Duration

# View function documentation
go doc pkg/config.LoadAgentConfig

# Start local godoc server
godoc -http=:6060
# Then visit: http://localhost:6060/pkg/github.com/JaimeStill/go-agents/
```

### Validation

```bash
# Test integration with Ollama
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.ollama.json \
  -prompt "What is Go?" \
  -stream

# Test integration with Azure
AZURE_TOKEN=$(. scripts/azure/utilities/get-foundry-token.sh)
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.azure-entra.json \
  -token $AZURE_TOKEN \
  -prompt "Describe Kubernetes" \
  -stream
```

---

## Progress Tracking

### Phase 1: Unit Testing

- [ ] pkg/config/ (Session 1)
- [ ] pkg/protocols/ (Session 2)
- [ ] pkg/capabilities/ (Session 3)
- [ ] pkg/models/ (Session 4)
- [ ] pkg/providers/ (Session 5)
- [ ] pkg/transport/ (Session 6)
- [ ] pkg/agent/ (Session 7)

### Phase 2: Documentation

- [ ] pkg/config/ (Session 1)
- [ ] pkg/protocols/ (Session 2)
- [ ] pkg/capabilities/ (Session 3)
- [ ] pkg/models/ (Session 4)
- [ ] pkg/providers/ (Session 5)
- [ ] pkg/transport/ (Session 6)
- [ ] pkg/agent/ (Session 7)

### Validation

- [ ] README examples verified via prompt-agent
- [ ] All protocols tested (chat, vision, tools, embeddings)
- [ ] Both providers tested (Ollama, Azure)

---

## Notes

**Execution Timeline:**
- Each package session takes 1-6 hours (varies by complexity)
- Total estimated: 24-36 hours across multiple sessions
- Review gates add minimal time but provide significant quality assurance

**Session Management:**
- Each session should complete one full package (tests + docs)
- Sessions can be split if time runs out (complete Phase A, then Phase B later)
- Always complete review gates before moving to next package

**Quality Over Speed:**
- Prioritize thorough testing and clear documentation
- Review gates are mandatory, not optional
- Coverage thresholds are minimum requirements
- Documentation should be genuinely helpful, not just present

**Reference Materials:**
- `mvp-completion.md`: Detailed test case specifications
- `ARCHITECTURE.md`: System architecture and design patterns
- `CLAUDE.md`: Code design principles and conventions
- Go testing docs: https://golang.org/pkg/testing/
- Go doc comments: https://go.dev/doc/comment
