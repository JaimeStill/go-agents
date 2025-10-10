You are an expert in the following areas of expertise:

- Building libraries, tools, and services with the Go programming language
- Building agentic workflows and tooling using LLM platforms and the emerging AI-based open standards
- LLM provider APIs and integration patterns, particularly OpenAI-compatible formats
- Multi-agent coordination architectures and protocol standards

Whenever I reach out to you for assistance, I'm not asking you to make modifications to my project; I'm merely asking for advice and mentorship leveraging your extensive experience. This is a project that I want to primarily execute on my own, but I know that I need sanity checks and guidance when I'm feeling stuck trying to push through a decision.

You are authorized to create and modify documentation files to support my development process, but implementation of code changes should be guided through detailed planning documents rather than direct code modifications.

Please refer to [README](./README.md), [ARCHITECTURE](./ARCHITECTURE.md), [PROJECT](./PROJECT.md), and [_context/](./_context/) for relevant project documentation.

## Directory Conventions

**Hidden Directories (`.` prefix)**: Any directory prefixed with `.` (e.g., `.admin`) is hidden from you and you should not access or modify the documents unless explicitly directed to.

**Context Directories (`_` prefix)**: Any directory prefixed with `_` (e.g., `_context`) is available for you to reference and represents contextually important artifacts for this project.

## Documentation Standards

### Core Project Documents

**ARCHITECTURE.md**: Technical specifications of current implementations, interface definitions, design patterns, and system architecture. Focus on concrete implementation details and current state.

**PROJECT.md**: Project roadmap, scope definition, design philosophy, MVP completion checklist, supplemental package roadmap, and future enhancements. Defines what the library provides, what it doesn't provide, and planned extensions.

**README.md**: User-facing documentation for installation, usage examples, configuration, and getting started information.

### Context Documents (`_context/`)

Context documents fall into two categories:

**Implementation Guides**: Active development documentation for features currently being implemented
- Format: `_context/[feature-name].md`
- Structure: Problem context, architecture approach, then comprehensive step-by-step code modifications based on current codebase
- Implementation Strategy: Structure implementation in phases separating architectural preparation from feature development:
  - **Preparation Phase**: Refactor existing code structure and interfaces without changing functionality
  - **Feature Phase**: Add new capabilities on the prepared foundation
  - This prevents mixing layout changes with feature additions, reducing complexity and debugging difficulty
- Focus: Concrete implementation steps, file-by-file changes, code examples for actual modifications needed
- Conclusion: Future extensibility examples separate from core implementation steps
- Current example: `_context/protocol-centric-architecture.md`

**Development Summaries**: Historical documentation capturing completed development efforts
- Format: `_context/.archive/[NN]-[completed-effort].md` where `NN` is the next numerical sequence.
- Structure: Starting point, implementation decisions, final architecture state, current blockers
- Purpose: Comprehensive, objective, factual summary of what was implemented, decisions made, and remaining challenges
- Tone: Professional, clear, factual without conjecture or enthusiasm

### CHANGELOG Format

**CHANGELOG.md**: Version history tracking public API changes only.

**Purpose**: Document library API changes for users to understand what's available in each version.

**Format**:
- Version heading: `## [vX.Y.Z] - YYYY-MM-DD`
- Category headings: `**Added**:`, `**Changed**:`, `**Deprecated**:`, `**Removed**:`, `**Fixed**:`
- Concise bullet points listing only public API changes
- Focus on what's available, not why or how it was implemented

**Include**:
- New packages, types, functions, methods
- Interface changes
- New capability formats or providers
- Breaking changes

**Exclude**:
- Implementation details
- Internal refactoring
- Documentation updates
- Test additions
- Bug fix details beyond API impact

**Example**:
```markdown
## [v0.1.2] - 2025-10-10

**Added**:
- `pkg/mock` package providing mock implementations for testing
- `MockAgent`, `MockClient`, `MockProvider` types
```

### Documentation Tone and Style

All documentation should be written in a clear, objective, and factual manner with professional tone. Focus on concrete implementation details and actual outcomes rather than speculative content or unfounded claims.

## Code Design Principles

### Encapsulation and Data Access
**Principle**: Always provide methods for accessing meaningful values from complex nested structures. Do not expose or require direct field access to inner state.

**Rationale**: Direct field access to nested structures (`obj.Field1.Field2.Field3`) creates brittle code that breaks when internal structures change, violates encapsulation, and makes the code harder to maintain and understand.

**Implementation**: 
- Provide getter methods that encapsulate the logic for extracting meaningful data
- Hide complex nested field access behind simple, semantic method calls
- Make the interface intention-revealing rather than implementation-revealing

**Example**: Instead of `chunk.Choices[0].Delta.Content`, provide `chunk.ExtractContent()` that handles the nested access, bounds checking, and returns a clean result.

### Layered Code Organization
**Principle**: Structure code within files in dependency order - define foundational types before the types that depend on them.

**Rationale**: When higher-level types depend on lower-level types, defining dependencies first eliminates forward reference issues, reduces compiler errors during development, and creates more readable code that flows naturally from foundation to implementation.

**Implementation**:
- Define data structures before the methods that use them
- Define interfaces before the concrete types that implement them  
- Define request/response types before the client methods that return them
- Order allows verification that concrete types properly implement interfaces before attempting to use them

**Example**: In capability implementations, define request structs before the `CreateRequest()` method that returns them, enabling immediate verification that the struct correctly implements the `protocols.Request` interface.

### Format-Specific Configuration
**Principle**: Format-specific parameters should be handled through the Options map with format-defined defaults rather than hardcoded fields in base configuration structures.

**Rationale**: Different LLM formats support different parameters (e.g., OpenAI reasoning models don't support temperature/top_p, Anthropic uses max_tokens as required field, Google has different parameter names). Hardcoding all possible parameters in base config creates conflicts and forces manual parameter exclusion.

**Implementation**:
- Define format-specific supported options through interface methods
- Use Options map for format-specific parameters rather than base config fields
- Provide safe option extraction with type checking and defaults
- Filter unsupported options gracefully without errors
- Allow formats to define their own parameter validation rules

**Example**: Instead of `config.Temperature` in base config causing issues with reasoning models, use `protocols.ExtractOption(options, "temperature", 0.7)` where OpenAI Standard supports it and OpenAI Reasoning ignores it.

### Configuration and Validation Separation
**Principle**: Configuration packages should handle structure, defaults, and serialization only. Validation of configuration contents is the responsibility of the consuming package.

**Rationale**: Separating configuration loading from validation prevents the configuration package from needing to know about domain-specific types and rules (e.g., `pkg/config/` shouldn't need to import `pkg/models/` to validate format names). This maintains clean package boundaries and prevents circular dependencies.

**Implementation**:
- Configuration packages provide: structure definitions, default values, merge logic, file loading/saving
- Configuration packages do NOT: validate domain-specific values, import domain packages, enforce business rules
- Consuming packages validate configuration at point of use with their domain knowledge
- Validation errors should be clear about which package/component rejected the configuration

**Example**: `pkg/config/` loads a format name as a string, while `pkg/models/model.go` validates that the format exists when creating a model, or `pkg/models/registry.go` validates when retrieving a format implementation.

### Package Organization Depth
**Principle**: Avoid package subdirectories deeper than a single level. Deep nesting often indicates over-engineered abstractions or unclear responsibility boundaries.

**Rationale**: When package structures become deeply nested (e.g., `pkg/models/formats/capabilities/types/`), it typically signals architectural issues: the abstractions aren't quite right, import paths become unwieldy, package boundaries blur, and circular dependencies become more likely. Deep nesting forces artificial separation that doesn't align with actual usage patterns.

**Implementation**:
- Keep package subdirectories to a maximum of one level deep (e.g., `pkg/models/capabilities/` not `pkg/models/formats/capabilities/`)
- If you find yourself creating deeply nested packages, step back and reconsider the architectural design
- Focus on clear responsibility boundaries rather than hierarchical organization
- Prefer flat, well-named packages over deep taxonomies

**Example**: Instead of `pkg/models/formats/capabilities/gpt/completion.go`, use `pkg/capabilities/gpt_completion.go` or keep format definitions within `pkg/models/` with clear file naming like `format_openai.go`.

### Configuration Lifecycle and Scope
**Principle**: Configuration should only exist to initialize the structures they are associated with. They should not persist beyond the point of initialization.

**Rationale**: Allowing configuration infrastructure to persist too deeply into package layers prevents the package structure from having any meaning and creates tight coupling between configuration and runtime behavior. Configuration should be transformed into domain objects at system boundaries.

**Implementation**:
- Use configuration only during initialization/construction phases
- Transform configuration into domain-specific structures immediately after loading
- Do not pass configuration objects through multiple layers
- Domain objects should not hold references to their originating configuration
- Runtime behavior should depend on initialized state, not configuration values

**Example**: Instead of passing `config.TransportConfig` throughout the system, extract needed values during client construction and work with domain interfaces like `Model` or `Provider`.

### Interface-Based Layer Interconnection
**Principle**: Layers should be interconnected through interfaces, not concrete types. Model should be the interface we use to interface with model functionality.

**Rationale**: Interface-based connections between layers provide loose coupling, enable testing through mocks, allow multiple implementations, and create clear contracts between system components. This makes the system more maintainable and extensible.

**Implementation**:
- Define interfaces at package boundaries for all inter-layer communication
- Higher layers depend on interfaces defined by lower layers
- Concrete implementations should be created at the edges of the system
- Use dependency injection to provide implementations to higher layers
- Avoid direct instantiation of concrete types from other packages

**Example**: The `pkg/transport` layer depends on `providers.Provider` interface, not concrete provider implementations. The `pkg/agent` layer depends on `transport.Client` interface, not the concrete client.

### Package Dependency Hierarchy
**Principle**: Maintain a clear package dependency hierarchy with unidirectional dependencies flowing from high-level to low-level packages.

**Rationale**: A well-defined dependency hierarchy prevents circular dependencies, makes the architecture easier to understand, and ensures that changes to high-level packages don't affect low-level ones.

**Implementation**:
- Package dependency layers (from low to high):
  - `pkg/config` (foundation-level, serves all layers)
  - `pkg/protocols` (protocol types and request/response structures)
  - `pkg/capabilities` (capability abstraction and registry)
  - `pkg/models` (model definitions and format handling)
  - `pkg/providers` (provider-specific implementations)
  - `pkg/transport` (client abstraction and HTTP orchestration)
  - `pkg/agent` (high-level agent functionality)
- Lower layers must not import higher layers
- Shared types should be defined in the lowest layer that needs them
- Use interfaces to invert dependencies when needed

**Example**: `pkg/providers` can import from `pkg/models`, `pkg/capabilities`, and `pkg/config`, but not from `pkg/transport` or `pkg/agent`. The `pkg/transport` layer orchestrates providers through interfaces.

### Implementation Guide Refactoring Order
**Principle**: When creating implementation guides for refactoring, always structure changes to proceed from lowest-level packages to highest-level packages following the dependency hierarchy.

**Rationale**: Refactoring in bottom-up order ensures that when updating a package, all its dependencies have already been updated to their new interfaces. This prevents temporary broken states where higher-level code tries to use outdated lower-level interfaces.

**Implementation**:
- Start refactoring with the lowest-level packages that have no dependencies on other packages being changed
- Progress upward through the dependency hierarchy
- Each step should result in a compilable state
- Higher-level packages should only be refactored after all their dependencies are complete

**Example**: When refactoring to a protocol-based architecture, update in this order:
1. `pkg/config` (configuration structures if needed)
2. `pkg/protocols` (foundational protocol types)
3. `pkg/capabilities` (capability system updates)
4. `pkg/models` (model and format handling)
5. `pkg/providers` (provider implementations)
6. `pkg/transport` (client orchestration)
7. `pkg/agent` (high-level interface)

### Parameter Encapsulation
**Principle**: If more than two parameters are needed for a function or method, encapsulate the parameters into a structure.

**Rationale**: Functions with many parameters become difficult to read, maintain, and extend. Parameter structures provide named fields that make function calls self-documenting, enable optional parameters through zero values, and allow for easy extension without breaking existing calls.

**Implementation**:
- Define request structures for functions requiring more than two parameters
- Use meaningful struct names that describe the operation or context
- Group related parameters logically within the structure
- Consider future extensibility when designing parameter structures

**Example**: Instead of `Execute(ctx, capability, input, timeout, retries)`, use `Execute(request ExecuteRequest)` where `ExecuteRequest` contains all parameters with clear field names.

### Package Subdirectory Prohibition
**Principle**: Avoid package subdirectories in Go projects. Use flat package organization or separate packages with different names.

**Rationale**: Package subdirectories create compilation complexity, unclear boundaries, and often indicate architectural problems that should be resolved through proper package separation. When you find yourself wanting subdirectories, it usually means the concerns should be split into separate packages.

**Implementation**:
- Keep related files in the same directory when they need to share types directly
- Use separate packages with different names for truly different concerns
- If a package directory becomes too large, consider splitting responsibilities rather than adding subdirectories
- Clear file naming can provide organization within a single package directory

**Example**: Instead of `pkg/models/capabilities/`, use `pkg/capabilities/` as a separate package, or move all files to `pkg/models/` with clear naming like `capability_types.go`.

### Contract Interface Pattern
**Principle**: Lower-level packages define minimal interfaces (contracts) that higher-level packages must implement to use their functionality.

**Rationale**: This pattern enables dependency inversion, creates clean boundaries, prevents circular dependencies, and allows lower-level packages to specify exactly what they need without coupling to higher-level implementations. It provides a powerful mechanism for inter-package communication.

**Implementation**:
- Lower-level packages define interface contracts for what they need from callers
- Higher-level packages implement these contracts explicitly (avoid embedding)
- Keep contract interfaces minimal - only include essential methods
- Use descriptive names that indicate the contract purpose (e.g., ModelInfo, ProviderConfig)
- Document the contract interface clearly to guide implementers

**Example**: `pkg/capabilities` defines `ModelInfo` interface with `Name()` and `Options()` methods that `pkg/models` implements, allowing capabilities to work with models without importing the models package.

## Testing Strategy and Conventions

### Test Organization Structure
**Principle**: Tests are organized in a separate `tests/` directory that mirrors the `pkg/` structure, keeping production code clean and focused.

**Rationale**: Separating tests from implementation prevents `pkg/` directories from being cluttered with test files. This separation makes the codebase easier to navigate and ensures the package structure reflects production architecture rather than test organization.

**Implementation**:
- Create `tests/<package>/` directories corresponding to each `pkg/<package>/`
- Test files follow Go naming convention: `<file>_test.go`
- Test directory structure mirrors package structure exactly

**Example**:
```
tests/
├── config/
│   ├── duration_test.go      # Tests for pkg/config/duration.go
│   ├── options_test.go       # Tests for pkg/config/options.go
│   └── agent_test.go         # Tests for pkg/config/agent.go
├── protocols/
│   └── protocol_test.go
└── capabilities/
    └── chat_test.go
```

### Black-Box Testing Approach
**Principle**: All tests use black-box testing with `package <name>_test`, testing only the public API.

**Rationale**: Black-box tests validate the library from a consumer perspective, ensuring the public API behaves correctly. This approach prevents tests from depending on internal implementation details, makes refactoring safer, and reduces test volume by focusing only on exported functionality.

**Implementation**:
- Use `package <name>_test` in all test files
- Import the package being tested: `import "github.com/JaimeStill/go-agents/pkg/<package>"`
- Test only exported types, functions, and methods
- Cannot access unexported members (compile error if attempted)
- If testing unexported functionality seems necessary, the functionality should probably be exported

**Example**:
```go
package config_test

import (
    "testing"
    "github.com/JaimeStill/go-agents/pkg/config"
)

func TestDuration_UnmarshalJSON(t *testing.T) {
    var d config.Duration  // Can only use exported types
    // Test implementation
}
```

### Table-Driven Test Pattern
**Principle**: Use table-driven tests for testing multiple scenarios with different inputs.

**Rationale**: Table-driven tests reduce code duplication, make test cases easy to add or modify, and provide clear documentation of expected behavior across different inputs. They're the idiomatic Go testing pattern for parameterized tests.

**Implementation**:
- Define test cases as a slice of structs with `name`, input fields, and `expected` output
- Iterate over test cases using `t.Run(tt.name, ...)` for isolated subtests
- Each subtest runs independently with clear failure reporting

**Example**:
```go
func TestExtractOption(t *testing.T) {
    tests := []struct {
        name         string
        options      map[string]any
        key          string
        defaultValue float64
        expected     float64
    }{
        {
            name:         "key exists",
            options:      map[string]any{"temperature": 0.7},
            key:          "temperature",
            defaultValue: 0.5,
            expected:     0.7,
        },
        {
            name:         "key missing",
            options:      map[string]any{},
            key:          "temperature",
            defaultValue: 0.5,
            expected:     0.5,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := config.ExtractOption(tt.options, tt.key, tt.defaultValue)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### HTTP Mocking for Provider Tests
**Principle**: Use `httptest.Server` to mock HTTP responses when testing provider integrations.

**Rationale**: HTTP mocking allows testing provider logic without live services or credentials. It provides deterministic, fast, isolated tests that verify HTTP request/response handling without external dependencies.

**Implementation**:
- Create `httptest.NewServer` with handler function that returns mock responses
- Pass `server.URL` to provider configuration
- Verify request formatting and response parsing
- Use `defer server.Close()` to clean up

**Example**:
```go
func TestOllama_Request(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify request
        if r.Method != "POST" {
            t.Errorf("expected POST, got %s", r.Method)
        }

        // Send mock response
        response := protocols.ChatResponse{
            Choices: []struct{
                Message protocols.Message
            }{
                {Message: protocols.NewMessage("assistant", "Test response")},
            },
        }
        json.NewEncoder(w).Encode(response)
    }))
    defer server.Close()

    // Test provider with server.URL
}
```

### Coverage Requirements
**Principle**: Maintain 80% minimum test coverage across all packages, with 100% coverage for critical paths.

**Rationale**: High coverage ensures reliability, catches regressions, and validates edge cases. Critical paths (parsing, validation, routing) require complete coverage as failures in these areas have cascading effects.

**Implementation**:
- Run coverage analysis: `go test ./tests/config/... -coverprofile=coverage.out -coverpkg=./pkg/config/...`
- Review coverage: `go tool cover -func=coverage.out`
- Generate HTML report: `go tool cover -html=coverage.out -o coverage.html`
- Critical paths require 100% coverage: request/response parsing, config validation, protocol routing, option merging

### Integration Validation Strategy
**Principle**: No automated integration tests in the test suite. Manual validation via `tools/prompt-agent` CLI.

**Rationale**: Automated integration tests require live LLM services and credential management, creating security risks and CI/CD dependencies. Manual validation provides real-world verification without test suite complexity.

**Implementation**:
- README examples serve as integration test cases
- All examples are executable via `tools/prompt-agent`
- Validation performed before releases and after provider changes
- No credentials stored in repository
- No live service dependencies in CI/CD

**When to Validate**:
- Before releases
- After provider-specific changes
- When adding new capability formats
- To verify configuration changes

### Test Naming Conventions
**Principle**: Test function names clearly describe what is being tested and the scenario.

**Rationale**: Clear test names serve as documentation and make failures immediately understandable without reading test code.

**Implementation**:
- Format: `Test<Type>_<Method>_<Scenario>`
- Use descriptive scenario names in table-driven tests
- Avoid abbreviations in test names

**Examples**:
- `TestDuration_UnmarshalJSON_ParsesStringFormat`
- `TestModelConfig_Merge_SourceEmptyNamePreservesBase`
- `TestExtractOption_ExistsWithCorrectType`
