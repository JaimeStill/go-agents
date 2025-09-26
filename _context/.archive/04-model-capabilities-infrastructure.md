# Model Capabilities Infrastructure Implementation

## Problem Context

The current architecture uses a format-based abstraction where models are grouped by API format (e.g., "openai-standard", "openai-reasoning"), with each format assuming uniform parameter support across all models. This approach has proven insufficient for several reasons:

### Current Limitations

1. **Format-Only Abstraction**: The system assumes provider-level consistency but evidence shows model-level variation within providers. For example, Azure reasoning models (o1/o3 series) require `max_completion_tokens` while traditional models use `max_tokens`.

2. **No Capability Infrastructure**: There is no mechanism to detect, validate, or handle model-specific capabilities like tools, vision, embedding, or thinking modes.

3. **Hardcoded Parameter Handling**: Format implementations blindly apply parameters without checking model support, leading to errors like "Unsupported parameter: 'max_tokens' is not supported with this model."

4. **Explicit Model Registration**: Any solution requiring explicit model registration contradicts the compositional philosophy that makes the current configuration system elegant.

### Research Findings

**Provider Model Uniformity**: Models are largely uniform across providers but with important variations:
- Same model (e.g., GPT-4) has identical core capabilities regardless of host
- Provider differences exist in performance, system prompts, safety filters, deployment configurations
- Function calling uses same modern parameters (`tools`, `tool_choice`) across providers
- Azure uses deployment names while OpenAI uses model names directly

**Capability-Endpoint Architecture**: Mixed architecture patterns discovered:
- **Single Endpoint**: Chat/Tools use `/chat/completions` or `/messages` with capability-specific parameters
- **Separate Endpoints**: Embeddings require distinct `/embeddings` endpoint with vector response format
- **Hybrid Approach Required**: System needs to support both patterns

## Architecture Approach

The solution is a **compositional model architecture** where models emerge from configuration composition rather than explicit registration. Models are defined by the composition of:
- Format (message structure)
- Capabilities (what it can do)
- Parameters (how it behaves)
- Provider context (where/how it's hosted)

### Compositional Design Principles

1. **Configuration-Based Models**: Models defined purely through configuration composition, no registration required
2. **Layered Architecture**: `AgentConfig.Client.Model` provides natural architectural flow
3. **Capability-Aware Formats**: Message formats understand and validate capabilities
4. **Provider Transport Layer**: Providers become pure transport/authentication layers
5. **Registry for Abstractions**: Register formats and capabilities, not models

### Package Restructuring Strategy

Current structure scatters model concerns across packages. The new structure consolidates model concerns while maintaining clean separation:

```
pkg/config/
├── model.go        # NEW: ModelConfig and model-specific configuration
├── llm.go          # LLM client configuration with embedded ModelConfig
├── agent.go        # Agent configuration
└── options.go      # Generic option handling utilities

pkg/models/         # NEW: Consolidated model concerns
├── models.go       # Core model interfaces and composition logic
├── registry.go     # Capability/format registries (NOT model registry)
├── formats/        # Moved from pkg/formats/
│   ├── models.go               # Format interfaces and shared types
│   ├── openai_standard.go      # Standard OpenAI format
│   ├── openai_reasoning.go     # Reasoning OpenAI format
│   └── anthropic_messages.go   # Anthropic messages format (future)
├── capabilities/   # NEW: Capability definitions and handlers
│   ├── completion.go   # Standard text completion
│   ├── tools.go       # Function/tool calling
│   ├── embeddings.go  # Vector embeddings (separate endpoint)
│   └── vision.go      # Image input handling
└── parameters.go   # NEW: Parameter validation and defaults

pkg/llm/
├── client.go       # Enhanced client interface
├── providers/      # Moved from pkg/providers/
│   ├── ollama/
│   ├── azure/
│   └── anthropic/  # (future)
└── errors.go

pkg/agent/
├── agent.go        # Enhanced with model composition
└── errors.go
```

## Phase 1: Architectural Preparation

This phase restructures existing code without adding new features. All existing functionality must continue working after these changes.

### Step 1: Create New Package Structure

Create the new directory structure:

```bash
mkdir -p pkg/models/formats
mkdir -p pkg/models/capabilities
mkdir -p pkg/llm/providers
```

### Step 2: Move Format Package

**2.1 Move Format Files**

Move all files from `pkg/formats/` to `pkg/models/formats/`:

```bash
mv pkg/formats/models.go pkg/models/formats/
mv pkg/formats/registry.go pkg/models/formats/
mv pkg/formats/openai_standard.go pkg/models/formats/
mv pkg/formats/openai_reasoning.go pkg/models/formats/
```

**2.2 Update Package Declarations**

In each moved file, update the package declaration:

```go
// Change from:
package formats

// To:
package formats
```

**2.3 Update Import Statements**

Update all imports throughout the codebase:

```go
// Change from:
import "github.com/s2va/agentic-toolkit/pkg/formats"

// To:
import "github.com/s2va/agentic-toolkit/pkg/models/formats"
```

Files requiring import updates:
- `pkg/llm/interface.go`
- `pkg/providers/ollama/ollama.go`
- `pkg/providers/azure/azure.go`
- `pkg/agent/agent.go`
- `tools/prompt-agent/main.go`

### Step 3: Move Provider Package

**3.1 Move Provider Files**

Move all provider directories:

```bash
mv pkg/providers/ollama pkg/llm/providers/
mv pkg/providers/azure pkg/llm/providers/
mv pkg/providers/registry.go pkg/llm/providers/
```

**3.2 Update Package Declarations**

In `pkg/llm/providers/registry.go`:
```go
// Change from:
package providers

// To:
package providers
```

In provider subdirectories, update imports:
```go
// In pkg/llm/providers/ollama/ollama.go, change:
import (
    "github.com/s2va/agentic-toolkit/pkg/formats"
)

// To:
import (
    "github.com/s2va/agentic-toolkit/pkg/models/formats"
)
```

**3.3 Update Import Statements**

Update imports throughout codebase:

```go
// Change from:
import "github.com/s2va/agentic-toolkit/pkg/providers"

// To:
import "github.com/s2va/agentic-toolkit/pkg/llm/providers"
```

### Step 4: Create Model Configuration Foundation

**4.1 Create `pkg/config/model.go`**

```go
package config

// ModelConfig represents the configuration for a language model
// This defines how a model is composed from its characteristics
type ModelConfig struct {
    Format       string         `json:"format,omitempty"`       // Message format: "openai-standard", "openai-reasoning"
    Name         string         `json:"name,omitempty"`         // Provider's model identifier
    Capabilities []string       `json:"capabilities,omitempty"` // Optional capability specification
    Options      map[string]any `json:"options,omitempty"`      // Model-specific parameters
}

// DefaultModelConfig returns default model configuration values
func DefaultModelConfig() *ModelConfig {
    return &ModelConfig{
        Format:       "openai-standard", // Default format
        Name:         "llama3.2:3b",     // Default model name
        Capabilities: []string{},        // Empty - will be detected or inferred
        Options:      make(map[string]any),
    }
}

// Merge overwrites non-zero values from source ModelConfig
func (c *ModelConfig) Merge(source *ModelConfig) {
    if source.Format != "" {
        c.Format = source.Format
    }
    if source.Name != "" {
        c.Name = source.Name
    }
    if len(source.Capabilities) > 0 {
        c.Capabilities = source.Capabilities
    }
    if source.Options != nil {
        c.Options = MergeOptions(c.Options, source.Options)
    }
}
```

**4.2 Update `pkg/config/llm.go`**

Add ModelConfig to LLMConfig structure:

```go
// Add this field to the existing LLMConfig struct:
type LLMConfig struct {
    Provider           string         `json:"provider"`
    Endpoint           string         `json:"endpoint"`
    Model              *ModelConfig   `json:"model,omitempty"`
    Timeout            time.Duration  `json:"timeout"`
    MaxRetries         int            `json:"max_retries"`
    RetryBackoffBase   time.Duration  `json:"retry_backoff_base"`
    ConnectionPoolSize int            `json:"connection_pool_size"`
    ConnectionTimeout  time.Duration  `json:"connection_timeout"`
    Options            map[string]any `json:"options,omitempty"`
}
```

**4.3 Update `DefaultLLMConfig()` function**

Replace the existing function to work with the new structure:

```go
func DefaultLLMConfig() LLMConfig {
    return LLMConfig{
        Provider:           "ollama",
        Endpoint:           "http://localhost:11434",
        Model:              DefaultModelConfig(),
        Timeout:            60 * time.Second,
        MaxRetries:         3,
        RetryBackoffBase:   time.Second,
        ConnectionPoolSize: 10,
        ConnectionTimeout:  90 * time.Second,
        Options:            make(map[string]any),
    }
}
```

**4.4 Update `Merge()` method**

Replace the existing Merge method to handle the new ModelConfig field:

```go
func (c *LLMConfig) Merge(source *LLMConfig) {
    if source.Provider != "" {
        c.Provider = source.Provider
    }

    if source.Endpoint != "" {
        c.Endpoint = source.Endpoint
    }

    if source.Model != nil {
        if c.Model == nil {
            c.Model = source.Model
        } else {
            c.Model.Merge(source.Model)
        }
    }

    if source.Timeout > 0 {
        c.Timeout = source.Timeout
    }

    if source.MaxRetries > 0 {
        c.MaxRetries = source.MaxRetries
    }

    if source.RetryBackoffBase > 0 {
        c.RetryBackoffBase = source.RetryBackoffBase
    }

    if source.ConnectionPoolSize > 0 {
        c.ConnectionPoolSize = source.ConnectionPoolSize
    }

    if source.ConnectionTimeout > 0 {
        c.ConnectionTimeout = source.ConnectionTimeout
    }

    if source.Options != nil {
        if c.Options == nil {
            c.Options = source.Options
        } else {
            maps.Copy(c.Options, source.Options)
        }
    }
}
```

### Step 5: Update Format Interfaces to Use ModelConfig

The format interfaces and implementations should only receive `ModelConfig`, not the entire `LLMConfig`, to maintain proper separation of concerns.

**5.1 Update Format Interface in `pkg/models/formats/models.go`**

```go
type MessageFormat interface {
    Format() string
    GetEndpointSuffix() string
    CreateRequest(messages []Message, model *config.ModelConfig, streaming bool) (FormatRequest, error)
    ParseResponse(data []byte) (*ChatResponse, error)
    ParseStreamingChunk(data []byte) (*StreamingChunk, error)
    IsStreamComplete(data string) bool

    // NEW: Placeholder methods for future capability support
    SupportedCapabilities() []string
    ValidateCapabilities(capabilities []string) error
}
```

**5.2 Update `pkg/models/formats/openai_standard.go`**

```go
func (f *OpenAIStandard) CreateRequest(messages []Message, model *config.ModelConfig, streaming bool) (FormatRequest, error) {
    req := &OpenAIStandardRequest{
        Model:    model.Name,
        Messages: messages,
        Stream:   streaming,
    }

    // Extract options from ModelConfig.Options
    req.MaxTokens = config.ExtractOption(model.Options, "max_tokens", 4096)
    req.Temperature = config.ExtractOption(model.Options, "temperature", 0.7)
    req.TopP = config.ExtractOption(model.Options, "top_p", 0.95)

    return req, nil
}
```

**5.3 Update `pkg/models/formats/openai_reasoning.go`**

```go
func (f *OpenAIReasoning) CreateRequest(messages []Message, model *config.ModelConfig, streaming bool) (FormatRequest, error) {
    req := &OpenAIReasoningRequest{
        Model:    model.Name,
        Messages: messages,
        Stream:   streaming,
    }

    // Extract options from ModelConfig.Options
    req.MaxCompletionTokens = config.ExtractOption(model.Options, "max_completion_tokens", 4096)

    return req, nil
}
```

**5.4 Update Provider Implementations**

Providers will need to pass `ModelConfig` to format methods instead of `LLMConfig`. For example, in `pkg/llm/providers/ollama/ollama.go`:

```go
func (c *Client) Chat(ctx context.Context, request *formats.ChatRequest) (*formats.ChatResponse, error) {
    // Get model configuration
    modelConfig := c.config.Model
    if modelConfig == nil {
        modelConfig = config.DefaultModelConfig()
    }

    // Pass ModelConfig instead of LLMConfig
    formatRequest, err := c.format.CreateRequest(request.Messages, modelConfig, false)
    if err != nil {
        return nil, err
    }

    // ... rest of implementation
}
```

### Step 6: Add Placeholder Capability Methods to Formats

**6.1 Add to `pkg/models/formats/openai_standard.go`**

```go
// SupportedCapabilities returns the capabilities supported by OpenAI Standard format
func (f *OpenAIStandard) SupportedCapabilities() []string {
    // Placeholder implementation - will be enhanced in Phase 2
    return []string{"completion", "tools"}
}

// ValidateCapabilities validates if the provided capabilities are supported
func (f *OpenAIStandard) ValidateCapabilities(capabilities []string) error {
    // Placeholder implementation - will be enhanced in Phase 2
    return nil
}
```

**6.2 Add to `pkg/models/formats/openai_reasoning.go`**

```go
// SupportedCapabilities returns the capabilities supported by OpenAI Reasoning format
func (f *OpenAIReasoning) SupportedCapabilities() []string {
    // Placeholder implementation - will be enhanced in Phase 2
    return []string{"completion", "thinking"}
}

// ValidateCapabilities validates if the provided capabilities are supported
func (f *OpenAIReasoning) ValidateCapabilities(capabilities []string) error {
    // Placeholder implementation - will be enhanced in Phase 2
    return nil
}
```

### Step 7: Update Client Interfaces

**7.1 Update `pkg/llm/interface.go`**

Add placeholder methods for future capability support (these are already shown as added in the system reminders):

```go
type Client interface {
    Provider() string
    Model() string
    Endpoint() string
    Chat(ctx context.Context, request *formats.ChatRequest) (*formats.ChatResponse, error)
    ChatStream(ctx context.Context, request *formats.ChatRequest) (<-chan formats.StreamingChunk, error)
    IsHealthy() bool
    Close() error

    // NEW: Placeholder methods for future capability support
    SupportedCapabilities(ctx context.Context) ([]string, error)
    SupportsCapability(capability string) bool
}
```

**7.2 Add Placeholder Implementations to Providers**

In `pkg/llm/providers/ollama/ollama.go`:
```go
// Add these methods to the Client struct:

// SupportedCapabilities returns the capabilities supported by this model
func (c *Client) SupportedCapabilities(ctx context.Context) ([]string, error) {
    // Placeholder implementation - will be enhanced in Phase 2
    return []string{"completion"}, nil
}

// SupportsCapability checks if the model supports a specific capability
func (c *Client) SupportsCapability(capability string) bool {
    // Placeholder implementation - will be enhanced in Phase 2
    return capability == "completion"
}
```

In `pkg/llm/providers/azure/azure.go`:
```go
// Add these methods to the Client struct:

// SupportedCapabilities returns the capabilities supported by this model
func (c *Client) SupportedCapabilities(ctx context.Context) ([]string, error) {
    // Placeholder implementation - will be enhanced in Phase 2
    return []string{"completion"}, nil
}

// SupportsCapability checks if the model supports a specific capability
func (c *Client) SupportsCapability(capability string) bool {
    // Placeholder implementation - will be enhanced in Phase 2
    return capability == "completion"
}
```

### Step 7: Update Agent Integration

**7.1 Update `pkg/agent/agent.go`**

Update agent to use the new model configuration structure:

```go
// Update the NewAgent function to handle ModelConfig:
func NewAgent(config *config.AgentConfig, token string) (*Agent, error) {
    // ... existing validation logic ...

    // Get model configuration
    modelConfig := config.Client.Model
    if modelConfig == nil {
        modelConfig = DefaultModelConfig()
    }

    // Validate format exists (using new path)
    format, exists := formats.GetFormat(modelConfig.Format)
    if !exists {
        return nil, &AgentError{
            Type:    ValidationError,
            Message: fmt.Sprintf("unknown format: %s", modelConfig.Format),
        }
    }

    // ... rest of existing logic unchanged ...
}
```

### Step 8: Update Tool Configuration Files

Update the existing configuration files in `tools/prompt-agent/` to use the new structure:

**8.1 Update `tools/prompt-agent/config.ollama.json`**

```json
{
  "name": "ollama-agent",
  "system_prompt": "You are a helpful assistant",
  "client": {
    "provider": "ollama",
    "endpoint": "http://localhost:11434",
    "model": {
      "format": "openai-standard",
      "name": "llama3.2:3b",
      "options": {
        "max_tokens": 4096,
        "temperature": 0.7,
        "top_p": 0.95
      }
    },
    "timeout": 60000000000,
    "max_retries": 3,
    "retry_backoff_base": 1000000000,
    "connection_pool_size": 10,
    "connection_timeout": 90000000000,
    "options": {}
  }
}
```

**8.2 Update `tools/prompt-agent/config.azure.json`**

```json
{
  "name": "azure-key-agent",
  "client": {
    "provider": "azure",
    "endpoint": "https://agentic-toolkit-platform.openai.azure.com",
    "model": {
      "format": "openai-reasoning",
      "name": "o3-mini",
      "options": {
        "max_completion_tokens": 4096
      }
    },
    "options": {
      "deployment": "o3-mini",
      "api_version": "2025-01-01-preview",
      "auth_type": "api_key"
    }
  }
}
```

**8.3 Update `tools/prompt-agent/config.azure-entra.json`**

```json
{
  "name": "azure-token-agent",
  "client": {
    "provider": "azure",
    "endpoint": "https://agentic-toolkit-platform.openai.azure.com",
    "model": {
      "format": "openai-reasoning",
      "name": "o3-mini",
      "options": {
        "max_completion_tokens": 4096
      }
    },
    "options": {
      "deployment": "o3-mini",
      "api_version": "2025-01-01-preview",
      "auth_type": "bearer"
    }
  }
}
```

Note: The model-specific options (like `max_tokens`, `temperature`, `max_completion_tokens`) are moved into `model.options`, while provider-specific options (like `deployment`, `api_version`, `auth_type`) remain in the client's `options` field.

### Step 9: Remove Old Package Directories

After confirming all imports are updated and tests pass:

```bash
rm -rf pkg/formats/
rm -rf pkg/providers/
```

### Step 10: Phase 1 Validation

**10.1 Build Validation**
```bash
go build ./...
```

**10.2 Test Validation**
```bash
go test ./...
```

**10.3 Tool Validation**

Test that existing tools continue to work with the updated configuration files:

```bash
# Test with Ollama
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.ollama.json \
  -prompt "Test message" \
  -stream

# Test with Azure
AZURE_API_KEY=$(. scripts/azure/utilities/get-foundry-key.sh)
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.azure.json \
  -token $AZURE_API_KEY \
  -prompt "Test message" \
  -stream
```

**10.4 Configuration Validation**

Verify that the updated configuration files work properly with the new model structure.

### Phase 1 Success Criteria

- [ ] All existing tests pass
- [ ] All imports successfully updated to new package paths
- [ ] Existing configuration files work without modification
- [ ] CLI tools function identically to before
- [ ] No new features implemented - only structural changes
- [ ] All placeholder methods added but return basic implementations

This completes Phase 1, preparing the foundation for Phase 2 feature implementation.

## Phase 2: Feature Implementation

This phase implements the model capabilities infrastructure on the prepared architectural foundation. Phase 1 must be completed and validated before beginning Phase 2.

### Step 1: Create Single Global Registry

**1.1 Create `pkg/registry/registry.go`**

Create a single, centralized registry for all extensible components:

```go
package registry

import (
    "sync"
    "github.com/s2va/agentic-toolkit/pkg/models/formats"
    "github.com/s2va/agentic-toolkit/pkg/llm"
)

// Registry manages all extensible components
type Registry struct {
    mu        sync.RWMutex
    formats   map[string]formats.MessageFormat
    providers map[string]ProviderFactory
}

// ProviderFactory creates LLM provider clients
type ProviderFactory func(config *config.LLMConfig) (llm.Client, error)

// Global registry instance
var globalRegistry = &Registry{
    formats:   make(map[string]formats.MessageFormat),
    providers: make(map[string]ProviderFactory),
}

// RegisterFormat registers a message format
func RegisterFormat(name string, format formats.MessageFormat) {
    globalRegistry.mu.Lock()
    defer globalRegistry.mu.Unlock()
    globalRegistry.formats[name] = format
}

// GetFormat retrieves a format by name
func GetFormat(name string) (formats.MessageFormat, bool) {
    globalRegistry.mu.RLock()
    defer globalRegistry.mu.RUnlock()
    format, exists := globalRegistry.formats[name]
    return format, exists
}

// RegisterProvider registers a provider factory
func RegisterProvider(name string, factory ProviderFactory) {
    globalRegistry.mu.Lock()
    defer globalRegistry.mu.Unlock()
    globalRegistry.providers[name] = factory
}

// GetProvider retrieves a provider factory by name
func GetProvider(name string) (ProviderFactory, bool) {
    globalRegistry.mu.RLock()
    defer globalRegistry.mu.RUnlock()
    factory, exists := globalRegistry.providers[name]
    return factory, exists
}

// ListFormats returns all registered format names
func ListFormats() []string {
    globalRegistry.mu.RLock()
    defer globalRegistry.mu.RUnlock()

    names := make([]string, 0, len(globalRegistry.formats))
    for name := range globalRegistry.formats {
        names = append(names, name)
    }
    return names
}

// ListProviders returns all registered provider names
func ListProviders() []string {
    globalRegistry.mu.RLock()
    defer globalRegistry.mu.RUnlock()

    names := make([]string, 0, len(globalRegistry.providers))
    for name := range globalRegistry.providers {
        names = append(names, name)
    }
    return names
}
```

### Step 2: Define Simple Capability Structure

**2.1 Create `pkg/models/capability.go`**

Define a minimal capability structure for configuration parsing only:

```go
package models

// Capability represents a model capability configuration
type Capability struct {
    Name                string            `json:"name"`
    Endpoint            string            `json:"endpoint,omitempty"`           // Custom endpoint (empty = use provider default)
    Parameters          map[string]any    `json:"parameters,omitempty"`         // Capability-specific parameters
}

// HasCustomEndpoint returns true if the capability requires a custom endpoint
func (c *Capability) HasCustomEndpoint() bool {
    return c.Endpoint != ""
}

// GetParameter retrieves a capability parameter with type checking
func (c *Capability) GetParameter(key string, defaultValue any) any {
    if c.Parameters == nil {
        return defaultValue
    }
    if value, exists := c.Parameters[key]; exists {
        return value
    }
    return defaultValue
}
```

No registration needed - capabilities are purely configuration-driven.

### Step 3: Update Model Configuration

**3.1 Update `pkg/config/model.go`**

Enhance ModelConfig to support detailed capability configuration:

```go
import "github.com/s2va/agentic-toolkit/pkg/models"

// ModelConfig represents model configuration with capabilities
type ModelConfig struct {
    Format       string                    `json:"format,omitempty"`
    Name         string                    `json:"name,omitempty"`
    Capabilities []*models.Capability      `json:"capabilities,omitempty"`     // Full capability configs
    Options      map[string]any           `json:"options,omitempty"`
}

// HasCapability checks if a capability is configured
func (m *ModelConfig) HasCapability(name string) bool {
    for _, cap := range m.Capabilities {
        if cap.Name == name {
            return true
        }
    }
    return false
}

// GetCapability returns a capability by name
func (m *ModelConfig) GetCapability(name string) (*models.Capability, bool) {
    for _, cap := range m.Capabilities {
        if cap.Name == name {
            return cap, true
        }
    }
    return nil, false
}

// GetCapabilityNames returns just the capability names
func (m *ModelConfig) GetCapabilityNames() []string {
    names := make([]string, len(m.Capabilities))
    for i, cap := range m.Capabilities {
        names[i] = cap.Name
    }
    return names
}
```

The model configuration now directly contains capability configurations, eliminating the need for capability registration or model composition complexity.

### Step 4: Update Provider Client Interfaces

**4.1 Update client capability methods**

Providers now simply return the configured capabilities from the model configuration:

```go
// In pkg/llm/providers/ollama/ollama.go and azure/azure.go

// SupportedCapabilities returns capabilities from model configuration
func (c *Client) SupportedCapabilities(ctx context.Context) ([]string, error) {
    if c.config.Model == nil {
        return []string{"completion"}, nil
    }
    return c.config.Model.GetCapabilityNames(), nil
}

// SupportsCapability checks if a capability is configured
func (c *Client) SupportsCapability(capability string) bool {
    if c.config.Model == nil {
        return capability == "completion"
    }
    return c.config.Model.HasCapability(capability)
}
```

Providers no longer need complex detection logic - they simply use the configured capabilities.

### Step 5: Enhanced Configuration Examples

Update configuration files with full capability objects. For example, in `tools/prompt-agent/config.ollama.json`:

```json
{
  "name": "ollama-agent",
  "system_prompt": "You are a helpful assistant",
  "client": {
    "provider": "ollama",
    "endpoint": "http://localhost:11434",
    "model": {
      "format": "openai-standard",
      "name": "llama3.2:3b",
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
      ],
      "options": {
        "max_tokens": 4096,
        "temperature": 0.7,
        "top_p": 0.95
      }
    },
    "timeout": 60000000000,
    "max_retries": 3,
    "retry_backoff_base": 1000000000,
    "connection_pool_size": 10,
    "connection_timeout": 90000000000,
    "options": {}
  }
}
```

And for Azure reasoning models in `tools/prompt-agent/config.azure.json`:

```json
{
  "name": "azure-key-agent",
  "client": {
    "provider": "azure",
    "endpoint": "https://agentic-toolkit-platform.openai.azure.com",
    "model": {
      "format": "openai-reasoning",
      "name": "o3-mini",
      "capabilities": [
        {
          "name": "completion",
          "parameters": {
            "max_completion_tokens": 4096
          }
        },
        {
          "name": "thinking"
        }
      ],
      "options": {
        "max_completion_tokens": 4096
      }
    },
    "options": {
      "deployment": "o3-mini",
      "api_version": "2025-01-01-preview",
      "auth_type": "api_key"
    }
  }
}
```

### Step 6: Update Format Capability Validation

Formats can optionally validate that configured capabilities are supported:

**6.1 Update `pkg/models/formats/openai_standard.go`**

```go
// SupportedCapabilities returns the capabilities supported by OpenAI Standard format
func (f *OpenAIStandard) SupportedCapabilities() []string {
    return []string{"completion", "tools", "embeddings"}
}

// ValidateCapabilities validates if the provided capabilities are supported
func (f *OpenAIStandard) ValidateCapabilities(capabilities []*models.Capability) error {
    supported := f.SupportedCapabilities()
    supportedMap := make(map[string]bool)
    for _, cap := range supported {
        supportedMap[cap] = true
    }

    for _, cap := range capabilities {
        if !supportedMap[cap.Name] {
            return fmt.Errorf("capability %s not supported by OpenAI Standard format", cap.Name)
        }
    }
    return nil
}
```

**6.2 Update `pkg/models/formats/openai_reasoning.go`**

```go
// SupportedCapabilities returns the capabilities supported by OpenAI Reasoning format
func (f *OpenAIReasoning) SupportedCapabilities() []string {
    return []string{"completion", "thinking"}
}

// ValidateCapabilities validates if the provided capabilities are supported
func (f *OpenAIReasoning) ValidateCapabilities(capabilities []*models.Capability) error {
    supported := f.SupportedCapabilities()
    supportedMap := make(map[string]bool)
    for _, cap := range supported {
        supportedMap[cap] = true
    }

    for _, cap := range capabilities {
        if !supportedMap[cap.Name] {
            return fmt.Errorf("capability %s not supported by OpenAI Reasoning format", cap.Name)
        }
    }
    return nil
}
```

### Step 5: Configuration File Updates

Update configuration files to include capabilities. For example, in `tools/prompt-agent/config.ollama.json`:

```json
{
  "name": "ollama-agent",
  "system_prompt": "You are a helpful assistant",
  "client": {
    "provider": "ollama",
    "endpoint": "http://localhost:11434",
    "model": {
      "format": "openai-standard",
      "name": "llama3.2:3b",
      "capabilities": ["completion", "tools"],
      "options": {
        "max_tokens": 4096,
        "temperature": 0.7,
        "top_p": 0.95
      }
    },
    "timeout": 60000000000,
    "max_retries": 3,
    "retry_backoff_base": 1000000000,
    "connection_pool_size": 10,
    "connection_timeout": 90000000000,
    "options": {}
  }
}
```

And in `tools/prompt-agent/config.azure.json`:

```json
{
  "name": "azure-key-agent",
  "client": {
    "provider": "azure",
    "endpoint": "https://agentic-toolkit-platform.openai.azure.com",
    "model": {
      "format": "openai-reasoning",
      "name": "o3-mini",
      "capabilities": ["completion", "thinking"],
      "options": {
        "max_completion_tokens": 4096
      }
    },
    "options": {
      "deployment": "o3-mini",
      "api_version": "2025-01-01-preview",
      "auth_type": "api_key"
    }
  }
}

    // Build capability set from response
    capSet := make(CapabilitySet)

    // Add capabilities from direct API response
    for _, cap := range showResp.Capabilities {
        capSet[Capability(cap)] = CapabilityInfo{
            Enabled:  true,
            Settings: make(map[string]any),
        }
    }

    // Infer additional capabilities from model families
    for _, family := range showResp.Details.Families {
        switch family {
        case "clip":
            capSet[CapabilityVision] = CapabilityInfo{
                Enabled:  true,
                Settings: make(map[string]any),
            }
        }
    }

    // All Ollama models support completion
    capSet[CapabilityCompletion] = CapabilityInfo{
        Enabled:  true,
        Settings: make(map[string]any),
    }

    return capSet, nil
}

// SupportsDetection returns true since Ollama provides capability detection
func (d *OllamaDetector) SupportsDetection() bool {
    return true
}
```

### Step 6: Optional Capability Detection Enhancement

Providers can optionally implement enhanced capability detection. This step is optional - the simple configuration-based approach is sufficient for most use cases.

For providers that support runtime capability detection (like Ollama), you can enhance the existing methods from Step 4 of Phase 2. The basic implementation already handles the core functionality needed.

### Step 7: Update Agent Integration

**7.1 Update `pkg/agent/agent.go`**

Agents can optionally access model capabilities through the client interface:

```go
// Add convenience methods to access capabilities (optional):

// HasCapability checks if the agent's model supports a capability
func (a *Agent) HasCapability(capability string) bool {
    return a.client.SupportsCapability(capability)
}

// GetCapabilities returns all supported capabilities
func (a *Agent) GetCapabilities() ([]string, error) {
    return a.client.SupportedCapabilities(context.Background())
}
```

This integration is minimal since the core capability logic is handled by the LLM client and model configuration.

### Phase 2 Success Criteria

- [ ] Single global registry at pkg/registry manages formats and providers
- [ ] Simple Capability struct used for configuration parsing only
- [ ] ModelConfig directly contains capability configurations
- [ ] Format validation works with capability objects
- [ ] Provider clients return configured capabilities from model config
- [ ] Configuration files use full capability objects with parameters
- [ ] Agent capability access methods work through client interface
- [ ] No capability registration needed - purely configuration-driven
- [ ] All existing functionality preserved

## Future Extensibility Examples

*[This section will be completed after core implementation]*