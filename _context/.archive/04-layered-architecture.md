# Layered Architecture Foundation

## Starting Point

This session began with Phase 1 of the model capabilities infrastructure complete. The architectural restructuring had established the package hierarchy (`pkg/models/`, `pkg/config/`, `pkg/llm/`) and moved format registries to appropriate locations. The system was ready for Phase 2 implementation, which originally included complex capability handlers and composition logic.

The user identified concerns about Phase 2's over-engineering, particularly the complex `CapabilityHandler` interfaces, detection systems, and model composition patterns that added significant complexity without clear benefit.

## Implementation Decisions

### Capability Architecture Simplification

**Decision**: Eliminated complex capability handlers entirely in favor of pure configuration-driven approach.

**Rationale**: If capabilities are configuration-driven and the `Capability` struct already specifies its parameters, global registration and complex handler interfaces add unnecessary abstraction layers.

**Implementation**:
- Removed `CapabilityHandler` interfaces and implementations
- Simplified `Capability` struct to contain only `Name`, `Endpoint`, and `Parameters` fields
- Eliminated capability-specific detection and composition logic

### Registry Architecture

**Initial Attempt**: Create single global registry at `pkg/registry/registry.go` for both formats and providers.

**Problem Discovered**: Import cycle constraint - providers depend on format registry to initialize, creating circular dependencies if both registries are in the same package.

**Final Decision**: Maintain separate registries:
- Format registry remains in `pkg/models/formats/registry.go`
- Provider registry remains in `pkg/llm/providers/registry.go`
- No capability registry needed

### Configuration-Driven Capabilities

**Decision**: Capabilities require no global registration, exist purely as configuration data.

**Implementation**:
- `ModelConfig` contains `[]*models.Capability` directly
- Each capability object includes optional endpoint and parameter overrides
- Providers return configured capabilities from model configuration
- No validation against global capability definitions required

**Benefits**:
- Zero registration overhead for new capabilities
- Maximum configuration flexibility per model
- Simplified provider implementation
- Easy testing with configuration-only approach

### Package Hierarchy Refinement

**Final Structure**:
```
pkg/
├── config/          # Configuration structures and parsing
│   ├── agent.go     # Agent configuration
│   ├── llm.go       # LLM client configuration
│   └── model.go     # Model and capability configuration
├── models/          # Model abstraction layer
│   ├── capability.go # Capability configuration struct
│   └── formats/     # Message format implementations
│       ├── registry.go
│       ├── openai_standard.go
│       └── openai_reasoning.go
├── llm/             # LLM client interfaces and implementations
│   ├── interface.go # Client interface definitions
│   └── providers/   # Provider implementations
│       ├── registry.go
│       ├── ollama/
│       └── azure/
└── agent/           # Agent orchestration layer
    ├── agent.go
    └── error.go
```

## Final Architecture State

### Capability System

**Structure**: Pure configuration-driven with no registration requirements.

**Configuration Format**:
```json
{
  "capabilities": [
    {
      "name": "completion",
      "parameters": {
        "max_tokens": 4096,
        "temperature": 0.7
      }
    },
    {
      "name": "embeddings",
      "endpoint": "/embeddings",
      "parameters": {
        "dimensions": 1024
      }
    }
  ]
}
```

**Provider Integration**: Providers return `ModelConfig.GetCapabilityNames()` for capability queries and use `ModelConfig.HasCapability()` for capability checking.

**Format Validation**: Formats can optionally validate that configured capabilities are supported through `ValidateCapabilities([]*models.Capability)` methods.

### Implementation Guide Updates

The Phase 2 section of `_context/04-model-capabilities-infrastructure.md` was updated to reflect the simplified architecture:

- Step 1: Single registry concept replaced with separate registry maintenance
- Step 2: Simple `Capability` struct for configuration parsing only
- Step 3: Direct `ModelConfig` capability configuration instead of model composition
- Step 4-7: Simplified provider integration and validation steps

## Current Blockers

### Client Interface Design Challenge

**Problem Identified**: The current `llm.Client` interface assumes all models support completion capability through `Chat()` and `ChatStream()` methods. However, embedding-only models (like `text-embedding-ada-002`) do not support completion.

**Impact**: This challenges the fundamental interface design and requires resolution before capability integration can proceed.

**Options Requiring Evaluation**:
1. Separate interfaces (`CompletionClient`, `EmbeddingClient`)
2. Optional methods with capability-based error handling
3. Capability-based client factories

### Phase 2 Implementation Status

**Complete**: Architecture planning and implementation guide updates
**Incomplete**: Actual code implementation of simplified capability system
**Next Session Requirements**:
- Resolve Client interface design for different model types
- Implement capability configuration parsing and validation
- Update provider implementations to use configured capabilities
- Test capability system with different model types

## Session Outcomes

**Architectural Foundation**: Established layered package hierarchy supporting configuration-driven capabilities without complex abstractions.

**Design Clarity**: Eliminated over-engineering concerns through pure configuration approach.

**Registry Strategy**: Resolved import cycle constraints while maintaining clean separation of concerns.

**Documentation**: Updated implementation guide to reflect simplified architecture decisions.

**Next Phase Preparation**: Identified critical interface design decision required for capability integration implementation.