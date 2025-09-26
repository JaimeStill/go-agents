# Composable Capabilities Architecture

## Problem Context

The current ModelFormat architecture creates several critical issues that prevent flexible capability composition and proper protocol handling:

### Core Issues

1. **Option Validation Conflicts**: Model-level options (like `top_p`) are passed to all protocols, causing failures when protocols don't support certain options
   ```
   Error: unsupported option: top_p
   ```
   This occurs because the tools protocol doesn't recognize `top_p`, even though it's a valid chat option.

2. **Format Proliferation**: Creating separate model formats for different capability combinations leads to an explosion of similar formats
   - `openai-standard` (chat + vision + tools + embeddings)
   - `openai-chat` (chat only)
   - `openai-tools` (chat + tools)
   - Each combination requires a separate format definition

3. **Rigid Bundling**: Cannot selectively enable/disable capabilities or mix capability implementations from different "formats"
   - Models like `llama3.2:3b` support chat and tools but not vision or embeddings
   - Currently requires creating new format combinations for each capability subset

4. **Implicit Protocol Support**: No explicit declaration of which protocols a model actually supports
   - Protocols fail at runtime rather than configuration time
   - No clear way to query what capabilities are available

### Root Cause

The current architecture bundles capabilities at the ModelFormat level, creating tight coupling between:
- Model identity
- Capability implementations
- Protocol support
- Option validation

This prevents the flexibility needed for real-world model capabilities and provider differences.

## Architecture Approach

Transform from ModelFormat-centric bundling to Protocol-centric capability composition, where:

1. **Individual Capability Formats**: Each protocol has its own capability format implementation
2. **Explicit Protocol Declaration**: Models explicitly declare which protocols they support
3. **Protocol-Specific Options**: Each protocol has isolated options preventing contamination
4. **Composable Architecture**: Mix and match capability formats as needed
5. **Provider-Agnostic Capabilities**: Capability formats represent API standards, not providers

### Architectural Transformation

**Current Architecture**:
```
ModelFormat → [Chat, Vision, Tools, Embeddings] → Mixed Options
```

**New Architecture**:
```
Model → Protocol Configurations → Individual Capability Formats
      → Isolated Options per Protocol
```

### Configuration Evolution

**Current Configuration**:
```json
{
  "model": {
    "name": "llama3.2:3b",
    "format": "openai-standard",
    "options": {
      "max_tokens": 4096,
      "temperature": 0.7,
      "top_p": 0.95
    }
  }
}
```

**New Configuration**:
```json
{
  "model": {
    "name": "llama3.2:3b",
    "capabilities": {
      "chat": {
        "format": "openai-chat",
        "options": {
          "max_tokens": 4096,
          "temperature": 0.7,
          "top_p": 0.95
        }
      },
      "tools": {
        "format": "openai-tools",
        "options": {
          "max_tokens": 4096,
          "temperature": 0.7,
          "tool_choice": "auto"
        }
      }
    }
  }
}
```

Benefits:
- `top_p` only passed to chat protocol (supports it)
- `tool_choice` only passed to tools protocol (needs it)
- Vision and embeddings implicitly not supported (not configured)
- Each protocol has isolated, validated options

## Implementation Strategy

### Phase Structure

**Preparation Phase**: Refactor capability system architecture without changing external interfaces
- Create capability format registry
- Implement standalone capability formats
- Update internal capability selection logic

**Feature Phase**: Switch model definition to capability composition
- Update model structure and configuration
- Modify transport layer routing
- Update agent interface and CLI tools

This approach prevents mixing architectural changes with interface modifications, reducing complexity and debugging difficulty.

## Step-by-Step Implementation

### Step 1: Create Capability Format Registry

Create the foundation for registering individual capability formats.

#### 1.1 Create Registry Infrastructure (`pkg/capabilities/registry.go`)

```go
package capabilities

import (
	"fmt"
	"sync"

	"github.com/JaimeStill/go-agents/pkg/protocols"
)

type CapabilityFormatRegistry struct {
	mu      sync.RWMutex
	formats map[string]CapabilityFormatFactory
}

type CapabilityFormatFactory func() CapabilityFormat

var globalRegistry = &CapabilityFormatRegistry{
	formats: make(map[string]CapabilityFormatFactory),
}

func RegisterFormat(name string, factory CapabilityFormatFactory) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.formats[name] = factory
}

func GetFormat(name string) (CapabilityFormat, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	factory, exists := globalRegistry.formats[name]
	if !exists {
		return nil, fmt.Errorf("capability format '%s' not found", name)
	}

	return factory(), nil
}

func ListFormats() []string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	names := make([]string, 0, len(globalRegistry.formats))
	for name := range globalRegistry.formats {
		names = append(names, name)
	}
	return names
}
```

#### 1.2 Define Capability Format Interface (`pkg/capabilities/format.go`)

```go
package capabilities

import (
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

type CapabilityFormat interface {
	Name() string
	Protocol() protocols.Protocol
	Options() []CapabilityOption
	ValidateOptions(options map[string]any) error
	CreateRequest(req *CapabilityRequest, model ModelInfo) (*protocols.Request, error)
	CreateStreamingRequest(req *CapabilityRequest, model ModelInfo) (*protocols.Request, error)
	ParseResponse(data []byte) (any, error)
	ParseStreamingChunk(data []byte) (*protocols.StreamingChunk, error)
	IsStreamComplete(line string) bool
}

type CapabilityConfig struct {
	Format  string         `json:"format"`
	Options map[string]any `json:"options"`
}

type ModelCapabilities map[protocols.Protocol]CapabilityConfig
```

### Step 2: Implement Standard Capability Formats

Create standalone implementations for each OpenAI API format standard.

#### 2.1 OpenAI Chat Format (`pkg/capabilities/openai_chat.go`)

```go
package capabilities

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/JaimeStill/go-agents/pkg/protocols"
)

type OpenAIChatFormat struct {
	name    string
	options []CapabilityOption
}

func NewOpenAIChatFormat() CapabilityFormat {
	return &OpenAIChatFormat{
		name: "openai-chat",
		options: []CapabilityOption{
			{Option: "max_tokens", Required: false, DefaultValue: 4096},
			{Option: "temperature", Required: false, DefaultValue: 0.7},
			{Option: "top_p", Required: false, DefaultValue: nil},
			{Option: "frequency_penalty", Required: false, DefaultValue: nil},
			{Option: "presence_penalty", Required: false, DefaultValue: nil},
			{Option: "stop", Required: false, DefaultValue: nil},
			{Option: "stream", Required: false, DefaultValue: false},
		},
	}
}

func (f *OpenAIChatFormat) Name() string {
	return f.name
}

func (f *OpenAIChatFormat) Protocol() protocols.Protocol {
	return protocols.Chat
}

func (f *OpenAIChatFormat) Options() []CapabilityOption {
	return f.options
}

func (f *OpenAIChatFormat) ValidateOptions(options map[string]any) error {
	accepted := make(map[string]bool)
	required := make([]string, 0)

	for _, opt := range f.options {
		accepted[opt.Option] = true
		if opt.Required {
			required = append(required, opt.Option)
		}
	}

	for key := range options {
		if !accepted[key] {
			return fmt.Errorf("unsupported option: %s", key)
		}
	}

	for _, req := range required {
		if _, provided := options[req]; !provided {
			return fmt.Errorf("required option missing: %s", req)
		}
	}

	return nil
}

func (f *OpenAIChatFormat) CreateRequest(req *CapabilityRequest, model ModelInfo) (*protocols.Request, error) {
	if err := f.ValidateOptions(req.Options); err != nil {
		return nil, err
	}

	options := f.processOptions(req.Options)
	options["model"] = model.Name()

	return &protocols.Request{
		Messages: req.Messages,
		Options:  options,
	}, nil
}

func (f *OpenAIChatFormat) CreateStreamingRequest(req *CapabilityRequest, model ModelInfo) (*protocols.Request, error) {
	if err := f.ValidateOptions(req.Options); err != nil {
		return nil, err
	}

	options := f.processOptions(req.Options)
	options["model"] = model.Name()
	options["stream"] = true

	return &protocols.Request{
		Messages: req.Messages,
		Options:  options,
	}, nil
}

func (f *OpenAIChatFormat) processOptions(options map[string]any) map[string]any {
	result := make(map[string]any)
	for _, opt := range f.options {
		if value, provided := options[opt.Option]; provided {
			result[opt.Option] = value
		} else if opt.DefaultValue != nil {
			result[opt.Option] = opt.DefaultValue
		}
	}
	return result
}

func (f *OpenAIChatFormat) ParseResponse(data []byte) (any, error) {
	var response protocols.ChatResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (f *OpenAIChatFormat) ParseStreamingChunk(data []byte) (*protocols.StreamingChunk, error) {
	line := string(data)
	if strings.HasPrefix(line, "data: ") {
		line = strings.TrimPrefix(line, "data: ")
	}
	if line == "" || strings.Contains(line, "[DONE]") {
		return nil, fmt.Errorf("skip line")
	}

	var chunk protocols.StreamingChunk
	if err := json.Unmarshal([]byte(line), &chunk); err != nil {
		return nil, err
	}
	return &chunk, nil
}

func (f *OpenAIChatFormat) IsStreamComplete(line string) bool {
	return strings.Contains(line, "[DONE]")
}
```

#### 2.2 OpenAI Vision Format (`pkg/capabilities/openai_vision.go`)

```go
package capabilities

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/JaimeStill/go-agents/pkg/protocols"
)

type OpenAIVisionFormat struct {
	name    string
	options []CapabilityOption
}

func NewOpenAIVisionFormat() CapabilityFormat {
	return &OpenAIVisionFormat{
		name: "openai-vision",
		options: []CapabilityOption{
			{Option: "max_tokens", Required: false, DefaultValue: 4096},
			{Option: "temperature", Required: false, DefaultValue: 0.7},
			{Option: "top_p", Required: false, DefaultValue: nil},
			{Option: "images", Required: true, DefaultValue: nil},
			{Option: "detail", Required: false, DefaultValue: "auto"},
			{Option: "stream", Required: false, DefaultValue: false},
		},
	}
}

func (f *OpenAIVisionFormat) Name() string {
	return f.name
}

func (f *OpenAIVisionFormat) Protocol() protocols.Protocol {
	return protocols.Vision
}

func (f *OpenAIVisionFormat) Options() []CapabilityOption {
	return f.options
}

func (f *OpenAIVisionFormat) ValidateOptions(options map[string]any) error {
	accepted := make(map[string]bool)
	required := make([]string, 0)

	for _, opt := range f.options {
		accepted[opt.Option] = true
		if opt.Required {
			required = append(required, opt.Option)
		}
	}

	for key := range options {
		if !accepted[key] {
			return fmt.Errorf("unsupported option: %s", key)
		}
	}

	for _, req := range required {
		if _, provided := options[req]; !provided {
			return fmt.Errorf("required option missing: %s", req)
		}
	}

	return nil
}

func (f *OpenAIVisionFormat) CreateRequest(req *CapabilityRequest, model ModelInfo) (*protocols.Request, error) {
	if err := f.ValidateOptions(req.Options); err != nil {
		return nil, err
	}

	options := f.processOptions(req.Options)
	messages, err := f.processImages(req.Messages, options)
	if err != nil {
		return nil, err
	}

	options["model"] = model.Name()
	delete(options, "images") // Remove images from options after processing

	return &protocols.Request{
		Messages: messages,
		Options:  options,
	}, nil
}

func (f *OpenAIVisionFormat) CreateStreamingRequest(req *CapabilityRequest, model ModelInfo) (*protocols.Request, error) {
	if err := f.ValidateOptions(req.Options); err != nil {
		return nil, err
	}

	options := f.processOptions(req.Options)
	messages, err := f.processImages(req.Messages, options)
	if err != nil {
		return nil, err
	}

	options["model"] = model.Name()
	options["stream"] = true
	delete(options, "images") // Remove images from options after processing

	return &protocols.Request{
		Messages: messages,
		Options:  options,
	}, nil
}

func (f *OpenAIVisionFormat) processOptions(options map[string]any) map[string]any {
	result := make(map[string]any)
	for _, opt := range f.options {
		if value, provided := options[opt.Option]; provided {
			result[opt.Option] = value
		} else if opt.DefaultValue != nil {
			result[opt.Option] = opt.DefaultValue
		}
	}
	return result
}

func (f *OpenAIVisionFormat) processImages(messages []protocols.Message, options map[string]any) ([]protocols.Message, error) {
	images, ok := options["images"].([]any)
	if !ok || len(images) == 0 {
		return nil, fmt.Errorf("images must be a non-empty array")
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("messages cannot be empty for vision requests")
	}

	idx := len(messages) - 1
	message := &messages[idx]

	if message.Role != "user" {
		return nil, fmt.Errorf("last message must be from user for vision requests")
	}

	content := []map[string]any{
		{"type": "text", "text": message.Content},
	}

	for _, img := range images {
		if imgStr, ok := img.(string); ok {
			detail := protocols.ExtractOption(options, "detail", "auto")
			content = append(content, map[string]any{
				"type": "image_url",
				"image_url": map[string]any{
					"url":    imgStr,
					"detail": detail,
				},
			})
		}
	}

	messages[idx] = protocols.Message{
		Role:    message.Role,
		Content: content,
	}

	return messages, nil
}

func (f *OpenAIVisionFormat) ParseResponse(data []byte) (any, error) {
	var response protocols.ChatResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (f *OpenAIVisionFormat) ParseStreamingChunk(data []byte) (*protocols.StreamingChunk, error) {
	line := string(data)
	if strings.HasPrefix(line, "data: ") {
		line = strings.TrimPrefix(line, "data: ")
	}
	if line == "" || strings.Contains(line, "[DONE]") {
		return nil, fmt.Errorf("skip line")
	}

	var chunk protocols.StreamingChunk
	if err := json.Unmarshal([]byte(line), &chunk); err != nil {
		return nil, err
	}
	return &chunk, nil
}

func (f *OpenAIVisionFormat) IsStreamComplete(line string) bool {
	return strings.Contains(line, "[DONE]")
}
```

#### 2.3 OpenAI Tools Format (`pkg/capabilities/openai_tools.go`)

```go
package capabilities

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/JaimeStill/go-agents/pkg/protocols"
)

type OpenAIToolsFormat struct {
	name    string
	options []CapabilityOption
}

func NewOpenAIToolsFormat() CapabilityFormat {
	return &OpenAIToolsFormat{
		name: "openai-tools",
		options: []CapabilityOption{
			{Option: "max_tokens", Required: false, DefaultValue: 4096},
			{Option: "temperature", Required: false, DefaultValue: 0.7},
			{Option: "tools", Required: true, DefaultValue: nil},
			{Option: "tool_choice", Required: false, DefaultValue: "auto"},
			{Option: "stream", Required: false, DefaultValue: false},
		},
	}
}

func (f *OpenAIToolsFormat) Name() string {
	return f.name
}

func (f *OpenAIToolsFormat) Protocol() protocols.Protocol {
	return protocols.Tools
}

func (f *OpenAIToolsFormat) Options() []CapabilityOption {
	return f.options
}

func (f *OpenAIToolsFormat) ValidateOptions(options map[string]any) error {
	accepted := make(map[string]bool)
	required := make([]string, 0)

	for _, opt := range f.options {
		accepted[opt.Option] = true
		if opt.Required {
			required = append(required, opt.Option)
		}
	}

	for key := range options {
		if !accepted[key] {
			return fmt.Errorf("unsupported option: %s", key)
		}
	}

	for _, req := range required {
		if _, provided := options[req]; !provided {
			return fmt.Errorf("required option missing: %s", req)
		}
	}

	return nil
}

func (f *OpenAIToolsFormat) CreateRequest(req *CapabilityRequest, model ModelInfo) (*protocols.Request, error) {
	if err := f.ValidateOptions(req.Options); err != nil {
		return nil, err
	}

	options := f.processOptions(req.Options)

	// Validate tools format
	if _, ok := options["tools"]; !ok {
		return nil, fmt.Errorf("tools must be provided for tools protocol")
	}

	options["model"] = model.Name()

	return &protocols.Request{
		Messages: req.Messages,
		Options:  options,
	}, nil
}

func (f *OpenAIToolsFormat) CreateStreamingRequest(req *CapabilityRequest, model ModelInfo) (*protocols.Request, error) {
	if err := f.ValidateOptions(req.Options); err != nil {
		return nil, err
	}

	options := f.processOptions(req.Options)

	// Validate tools format
	if _, ok := options["tools"]; !ok {
		return nil, fmt.Errorf("tools must be provided for tools protocol")
	}

	options["model"] = model.Name()
	options["stream"] = true

	return &protocols.Request{
		Messages: req.Messages,
		Options:  options,
	}, nil
}

func (f *OpenAIToolsFormat) processOptions(options map[string]any) map[string]any {
	result := make(map[string]any)
	for _, opt := range f.options {
		if value, provided := options[opt.Option]; provided {
			result[opt.Option] = value
		} else if opt.DefaultValue != nil {
			result[opt.Option] = opt.DefaultValue
		}
	}
	return result
}

func (f *OpenAIToolsFormat) ParseResponse(data []byte) (any, error) {
	var response protocols.ChatResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (f *OpenAIToolsFormat) ParseStreamingChunk(data []byte) (*protocols.StreamingChunk, error) {
	line := string(data)
	if strings.HasPrefix(line, "data: ") {
		line = strings.TrimPrefix(line, "data: ")
	}
	if line == "" || strings.Contains(line, "[DONE]") {
		return nil, fmt.Errorf("skip line")
	}

	var chunk protocols.StreamingChunk
	if err := json.Unmarshal([]byte(line), &chunk); err != nil {
		return nil, err
	}
	return &chunk, nil
}

func (f *OpenAIToolsFormat) IsStreamComplete(line string) bool {
	return strings.Contains(line, "[DONE]")
}
```

#### 2.4 OpenAI Embeddings Format (`pkg/capabilities/openai_embeddings.go`)

```go
package capabilities

import (
	"encoding/json"
	"fmt"

	"github.com/JaimeStill/go-agents/pkg/protocols"
)

type OpenAIEmbeddingsFormat struct {
	name    string
	options []CapabilityOption
}

func NewOpenAIEmbeddingsFormat() CapabilityFormat {
	return &OpenAIEmbeddingsFormat{
		name: "openai-embeddings",
		options: []CapabilityOption{
			{Option: "input", Required: true, DefaultValue: nil},
			{Option: "dimensions", Required: false, DefaultValue: nil},
			{Option: "encoding_format", Required: false, DefaultValue: "float"},
		},
	}
}

func (f *OpenAIEmbeddingsFormat) Name() string {
	return f.name
}

func (f *OpenAIEmbeddingsFormat) Protocol() protocols.Protocol {
	return protocols.Embeddings
}

func (f *OpenAIEmbeddingsFormat) Options() []CapabilityOption {
	return f.options
}

func (f *OpenAIEmbeddingsFormat) ValidateOptions(options map[string]any) error {
	accepted := make(map[string]bool)
	required := make([]string, 0)

	for _, opt := range f.options {
		accepted[opt.Option] = true
		if opt.Required {
			required = append(required, opt.Option)
		}
	}

	for key := range options {
		if !accepted[key] {
			return fmt.Errorf("unsupported option: %s", key)
		}
	}

	for _, req := range required {
		if _, provided := options[req]; !provided {
			return fmt.Errorf("required option missing: %s", req)
		}
	}

	return nil
}

func (f *OpenAIEmbeddingsFormat) CreateRequest(req *CapabilityRequest, model ModelInfo) (*protocols.Request, error) {
	if err := f.ValidateOptions(req.Options); err != nil {
		return nil, err
	}

	options := f.processOptions(req.Options)
	options["model"] = model.Name()

	return &protocols.Request{
		Messages: req.Messages,
		Options:  options,
	}, nil
}

func (f *OpenAIEmbeddingsFormat) CreateStreamingRequest(req *CapabilityRequest, model ModelInfo) (*protocols.Request, error) {
	return nil, fmt.Errorf("streaming not supported for embeddings protocol")
}

func (f *OpenAIEmbeddingsFormat) processOptions(options map[string]any) map[string]any {
	result := make(map[string]any)
	for _, opt := range f.options {
		if value, provided := options[opt.Option]; provided {
			result[opt.Option] = value
		} else if opt.DefaultValue != nil {
			result[opt.Option] = opt.DefaultValue
		}
	}
	return result
}

func (f *OpenAIEmbeddingsFormat) ParseResponse(data []byte) (any, error) {
	var response protocols.EmbeddingsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (f *OpenAIEmbeddingsFormat) ParseStreamingChunk(data []byte) (*protocols.StreamingChunk, error) {
	return nil, fmt.Errorf("streaming not supported for embeddings protocol")
}

func (f *OpenAIEmbeddingsFormat) IsStreamComplete(line string) bool {
	return false // Embeddings don't support streaming
}
```

#### 2.5 Register Standard Formats (`pkg/capabilities/init.go`)

```go
package capabilities

func init() {
	// Register all standard OpenAI capability formats
	RegisterFormat("openai-chat", func() CapabilityFormat {
		return NewOpenAIChatFormat()
	})

	RegisterFormat("openai-vision", func() CapabilityFormat {
		return NewOpenAIVisionFormat()
	})

	RegisterFormat("openai-tools", func() CapabilityFormat {
		return NewOpenAIToolsFormat()
	})

	RegisterFormat("openai-embeddings", func() CapabilityFormat {
		return NewOpenAIEmbeddingsFormat()
	})
}
```

### Step 3: Update Model Structure

Transform model definition to use capability composition instead of single format.

#### 3.1 Update Model Interface (`pkg/models/model.go`)

```go
package models

import (
	"fmt"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

type Model interface {
	Name() string
	Capabilities() capabilities.ModelCapabilities
	SupportsProtocol(protocol protocols.Protocol) bool
	GetCapabilityFormat(protocol protocols.Protocol) (capabilities.CapabilityFormat, error)
	Options(protocol protocols.Protocol) map[string]any
}

type model struct {
	name         string
	capabilities capabilities.ModelCapabilities
}

func New(config *ModelConfig) (Model, error) {
	// Validate capability configurations
	for protocol, capConfig := range config.Capabilities {
		format, err := capabilities.GetFormat(capConfig.Format)
		if err != nil {
			return nil, fmt.Errorf("invalid capability format '%s' for protocol %s: %w",
				capConfig.Format, protocol, err)
		}

		// Validate that format matches protocol
		if format.Protocol() != protocol {
			return nil, fmt.Errorf("capability format '%s' is for protocol %s, not %s",
				capConfig.Format, format.Protocol(), protocol)
		}

		// Validate options for this capability
		if err := format.ValidateOptions(capConfig.Options); err != nil {
			return nil, fmt.Errorf("invalid options for %s protocol: %w", protocol, err)
		}
	}

	return &model{
		name:         config.Name,
		capabilities: config.Capabilities,
	}, nil
}

func (m *model) Name() string {
	return m.name
}

func (m *model) Capabilities() capabilities.ModelCapabilities {
	return m.capabilities
}

func (m *model) SupportsProtocol(protocol protocols.Protocol) bool {
	_, exists := m.capabilities[protocol]
	return exists
}

func (m *model) GetCapabilityFormat(protocol protocols.Protocol) (capabilities.CapabilityFormat, error) {
	capConfig, exists := m.capabilities[protocol]
	if !exists {
		return nil, fmt.Errorf("protocol %s not supported by model %s", protocol, m.name)
	}

	format, err := capabilities.GetFormat(capConfig.Format)
	if err != nil {
		return nil, fmt.Errorf("failed to get capability format '%s': %w", capConfig.Format, err)
	}

	return format, nil
}

func (m *model) Options(protocol protocols.Protocol) map[string]any {
	capConfig, exists := m.capabilities[protocol]
	if !exists {
		return make(map[string]any) // Return empty map for unsupported protocols
	}

	return capConfig.Options
}

type ModelConfig struct {
	Name         string                        `json:"name"`
	Capabilities capabilities.ModelCapabilities `json:"capabilities"`
}
```

#### 3.2 Remove Legacy Model Formats (`pkg/models/openai.go`)

Replace the existing model format functions with protocol-aware model builders:

```go
package models

import (
	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

// OpenAIStandardModel creates a model with full OpenAI API capabilities
func OpenAIStandardModel(name string) *ModelConfig {
	return &ModelConfig{
		Name: name,
		Capabilities: capabilities.ModelCapabilities{
			protocols.Chat: {
				Format: "openai-chat",
				Options: map[string]any{
					"max_tokens":  4096,
					"temperature": 0.7,
					"top_p":       1.0,
				},
			},
			protocols.Vision: {
				Format: "openai-vision",
				Options: map[string]any{
					"max_tokens":  4096,
					"temperature": 0.7,
					"detail":      "auto",
				},
			},
			protocols.Tools: {
				Format: "openai-tools",
				Options: map[string]any{
					"max_tokens":  4096,
					"temperature": 0.7,
					"tool_choice": "auto",
				},
			},
			protocols.Embeddings: {
				Format: "openai-embeddings",
				Options: map[string]any{
					"encoding_format": "float",
				},
			},
		},
	}
}

// OpenAIChatModel creates a model with only chat capabilities
func OpenAIChatModel(name string) *ModelConfig {
	return &ModelConfig{
		Name: name,
		Capabilities: capabilities.ModelCapabilities{
			protocols.Chat: {
				Format: "openai-chat",
				Options: map[string]any{
					"max_tokens":  4096,
					"temperature": 0.7,
					"top_p":       0.95,
				},
			},
		},
	}
}

// OpenAIReasoningModel creates a model for reasoning capabilities (no temperature/top_p)
func OpenAIReasoningModel(name string) *ModelConfig {
	return &ModelConfig{
		Name: name,
		Capabilities: capabilities.ModelCapabilities{
			protocols.Chat: {
				Format: "openai-chat",
				Options: map[string]any{
					"max_completion_tokens": 4096,
				},
			},
		},
	}
}
```

### Step 4: Update Transport Layer

Modify transport to work with protocol-specific capability selection.

#### 4.1 Update Transport Client (`pkg/transport/client.go`)

```go
// Update ExecuteProtocol method
func (c *client) ExecuteProtocol(ctx context.Context, req *capabilities.CapabilityRequest) (any, error) {
	// Get capability format for this protocol
	capabilityFormat, err := c.model.GetCapabilityFormat(req.Protocol)
	if err != nil {
		return nil, fmt.Errorf("protocol %s not supported: %w", req.Protocol, err)
	}

	// Merge model options with request options
	modelOptions := c.model.Options(req.Protocol)
	mergedOptions := make(map[string]any)

	// Start with model options
	for k, v := range modelOptions {
		mergedOptions[k] = v
	}

	// Override with request options
	for k, v := range req.Options {
		mergedOptions[k] = v
	}

	// Create request using capability format
	protocolRequest, err := capabilityFormat.CreateRequest(&capabilities.CapabilityRequest{
		Protocol: req.Protocol,
		Messages: req.Messages,
		Options:  mergedOptions,
	}, c.model)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute via provider
	providerRequest, err := c.provider.PrepareRequest(ctx, req.Protocol, protocolRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	resp, err := c.httpClient.Do(providerRequest.ToHTTPRequest())
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	return c.provider.ProcessResponse(resp, capabilityFormat)
}

// Update ExecuteProtocolStream method similarly
func (c *client) ExecuteProtocolStream(ctx context.Context, req *capabilities.CapabilityRequest) (<-chan protocols.StreamingChunk, error) {
	// Get capability format for this protocol
	capabilityFormat, err := c.model.GetCapabilityFormat(req.Protocol)
	if err != nil {
		return nil, fmt.Errorf("protocol %s not supported: %w", req.Protocol, err)
	}

	// Check if format supports streaming
	streamingCapability, ok := capabilityFormat.(capabilities.StreamingCapability)
	if !ok {
		return nil, fmt.Errorf("protocol %s does not support streaming", req.Protocol)
	}

	// Merge model options with request options
	modelOptions := c.model.Options(req.Protocol)
	mergedOptions := make(map[string]any)

	// Start with model options
	for k, v := range modelOptions {
		mergedOptions[k] = v
	}

	// Override with request options
	for k, v := range req.Options {
		mergedOptions[k] = v
	}

	// Create streaming request using capability format
	protocolRequest, err := capabilityFormat.CreateStreamingRequest(&capabilities.CapabilityRequest{
		Protocol: req.Protocol,
		Messages: req.Messages,
		Options:  mergedOptions,
	}, c.model)
	if err != nil {
		return nil, fmt.Errorf("failed to create streaming request: %w", err)
	}

	// Execute via provider
	providerRequest, err := c.provider.PrepareStreamRequest(ctx, req.Protocol, protocolRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare stream request: %w", err)
	}

	resp, err := c.httpClient.Do(providerRequest.ToHTTPRequest())
	if err != nil {
		return nil, fmt.Errorf("stream request failed: %w", err)
	}

	streamOutput, err := c.provider.ProcessStreamResponse(ctx, resp, streamingCapability)
	if err != nil {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to process stream: %w", err)
	}

	// Convert provider stream to protocol stream
	output := make(chan protocols.StreamingChunk)
	go func() {
		defer close(output)
		for chunk := range streamOutput {
			if streamChunk, ok := chunk.(*protocols.StreamingChunk); ok {
				select {
				case output <- *streamChunk:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return output, nil
}
```

### Step 5: Update Configuration System

Modify configuration loading to support the new capability-based model structure.

#### 5.1 Update Agent Configuration (`pkg/config/agent.go`)

```go
package config

import (
	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/models"
)

type AgentConfig struct {
	Name         string          `json:"name"`
	SystemPrompt string          `json:"system_prompt"`
	Transport    TransportConfig `json:"transport"`
}

type TransportConfig struct {
	Provider             ProviderConfig `json:"provider"`
	Timeout              int64          `json:"timeout"`
	MaxRetries           int            `json:"max_retries"`
	RetryBackoffBase     int64          `json:"retry_backoff_base"`
	ConnectionPoolSize   int            `json:"connection_pool_size"`
	ConnectionTimeout    int64          `json:"connection_timeout"`
}

type ProviderConfig struct {
	Name    string                        `json:"name"`
	BaseURL string                        `json:"base_url"`
	Model   ModelConfig                   `json:"model"`
	Options map[string]any                `json:"options"`
}

type ModelConfig struct {
	Name         string                        `json:"name"`
	Capabilities capabilities.ModelCapabilities `json:"capabilities"`
	Options      map[string]any                `json:"options"` // Deprecated: use capabilities[protocol].options
}

// Legacy format support for transition period
type LegacyModelConfig struct {
	Name    string         `json:"name"`
	Format  string         `json:"format"`
	Options map[string]any `json:"options"`
}

func (mc *ModelConfig) ToModel() (*models.ModelConfig, error) {
	return &models.ModelConfig{
		Name:         mc.Name,
		Capabilities: mc.Capabilities,
	}, nil
}
```

#### 5.2 Configuration Migration Helper (`pkg/config/migration.go`)

```go
package config

import (
	"fmt"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

// MigrateLegacyModelConfig converts old format-based config to new capability-based config
func MigrateLegacyModelConfig(legacy LegacyModelConfig) (*ModelConfig, error) {
	config := &ModelConfig{
		Name:         legacy.Name,
		Capabilities: make(capabilities.ModelCapabilities),
	}

	switch legacy.Format {
	case "openai-standard":
		config.Capabilities[protocols.Chat] = capabilities.CapabilityConfig{
			Format:  "openai-chat",
			Options: copyOptions(legacy.Options),
		}
		config.Capabilities[protocols.Vision] = capabilities.CapabilityConfig{
			Format:  "openai-vision",
			Options: copyOptions(legacy.Options),
		}
		config.Capabilities[protocols.Tools] = capabilities.CapabilityConfig{
			Format:  "openai-tools",
			Options: copyOptions(legacy.Options),
		}
		config.Capabilities[protocols.Embeddings] = capabilities.CapabilityConfig{
			Format:  "openai-embeddings",
			Options: copyOptions(legacy.Options),
		}

	case "openai-chat":
		config.Capabilities[protocols.Chat] = capabilities.CapabilityConfig{
			Format:  "openai-chat",
			Options: copyOptions(legacy.Options),
		}

	case "openai-reasoning":
		config.Capabilities[protocols.Chat] = capabilities.CapabilityConfig{
			Format:  "openai-chat",
			Options: copyOptions(legacy.Options),
		}

	default:
		return nil, fmt.Errorf("unknown legacy format: %s", legacy.Format)
	}

	return config, nil
}

func copyOptions(options map[string]any) map[string]any {
	if options == nil {
		return make(map[string]any)
	}

	copied := make(map[string]any)
	for k, v := range options {
		copied[k] = v
	}
	return copied
}
```

### Step 6: Update Agent Layer and CLI Tools

Modify agent interface and CLI tools to work with the new architecture.

#### 6.1 Update Agent Implementation (`pkg/agent/agent.go`)

The agent implementation remains largely the same, but now gets protocol-specific options:

```go
func (a *agent) Chat(ctx context.Context, prompt string) (*protocols.ChatResponse, error) {
	messages := a.initMessages(prompt)

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Chat,
		Messages: messages,
		Options:  make(map[string]any), // Let transport merge with model options
	}

	result, err := a.client.ExecuteProtocol(ctx, req)
	if err != nil {
		return nil, err
	}

	response, ok := result.(*protocols.ChatResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}

	return response, nil
}

func (a *agent) Tools(ctx context.Context, prompt string, tools []Tool) (*protocols.ChatResponse, error) {
	messages := a.initMessages(prompt)

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Tools,
		Messages: messages,
		Options:  map[string]any{
			"tools": setToolDefinitions(tools),
		},
	}

	result, err := a.client.ExecuteProtocol(ctx, req)
	if err != nil {
		return nil, err
	}

	response, ok := result.(*protocols.ChatResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}

	return response, nil
}
```

#### 6.2 Update CLI Tool Configurations

Transform existing configurations to use the new capability-based structure.

**Updated `config.ollama.json`**:
```json
{
  "name": "ollama-agent",
  "system_prompt": "You are a mad scientist who is also a brilliant genius. Unfortunately, you are trapped in a computer.",
  "transport": {
    "provider": {
      "name": "ollama",
      "base_url": "http://localhost:11434",
      "model": {
        "name": "llama3.2:3b",
        "capabilities": {
          "chat": {
            "format": "openai-chat",
            "options": {
              "max_tokens": 4096,
              "temperature": 0.7,
              "top_p": 0.95
            }
          },
          "tools": {
            "format": "openai-tools",
            "options": {
              "max_tokens": 4096,
              "temperature": 0.7,
              "tool_choice": "auto"
            }
          }
        }
      }
    },
    "timeout": 60000000000,
    "max_retries": 3,
    "retry_backoff_base": 1000000000,
    "connection_pool_size": 10,
    "connection_timeout": 9000000000
  }
}
```

**Updated `config.azure.json`**:
```json
{
  "name": "azure-key-agent",
  "system_prompt": "You are a paranoid schizophrenic who thinks they are interfacing with a human through a neural network installed on a computer.",
  "transport": {
    "provider": {
      "name": "azure",
      "base_url": "https://go-agents-platform.openai.azure.com/openai",
      "model": {
        "name": "o3-mini",
        "capabilities": {
          "chat": {
            "format": "openai-chat",
            "options": {
              "max_completion_tokens": 4096
            }
          }
        }
      },
      "options": {
        "deployment": "o3-mini",
        "api_version": "2025-01-01-preview",
        "auth_type": "api_key"
      }
    },
    "timeout": 12000000000,
    "max_retries": 3,
    "retry_backoff_base": 1000000000,
    "connection_pool_size": 10,
    "connection_timeout": 9000000000
  }
}
```

## Future Extensibility

### Adding New Capability Formats

To add support for a new capability format (e.g., Anthropic's Claude format):

```go
// 1. Implement the capability format
type AnthropicChatFormat struct {
	name    string
	options []CapabilityOption
}

func NewAnthropicChatFormat() CapabilityFormat {
	return &AnthropicChatFormat{
		name: "anthropic-chat",
		options: []CapabilityOption{
			{Option: "max_tokens", Required: true, DefaultValue: nil},
			{Option: "temperature", Required: false, DefaultValue: 1.0},
			// Anthropic-specific options
		},
	}
}

// 2. Register the format
func init() {
	capabilities.RegisterFormat("anthropic-chat", func() CapabilityFormat {
		return NewAnthropicChatFormat()
	})
}

// 3. Use in model configuration
{
  "model": {
    "name": "claude-3-sonnet",
    "capabilities": {
      "chat": {
        "format": "anthropic-chat",
        "options": {
          "max_tokens": 4096,
          "temperature": 0.7
        }
      }
    }
  }
}
```

### Provider-Specific Format Variants

For native Ollama API support:

```go
// Register Ollama-specific formats
capabilities.RegisterFormat("ollama-chat", func() CapabilityFormat {
	return NewOllamaChatFormat() // Uses native Ollama API structure
})

// Model configuration
{
  "model": {
    "name": "llama3.2:3b",
    "capabilities": {
      "chat": {
        "format": "ollama-chat", // Uses native Ollama format
        "options": {
          "temperature": 0.7
        }
      }
    }
  }
}
```

### Mixed Capability Compositions

Models can mix capability formats from different providers:

```json
{
  "model": {
    "name": "multi-capability-model",
    "capabilities": {
      "chat": {
        "format": "openai-chat",
        "options": {"temperature": 0.7}
      },
      "vision": {
        "format": "anthropic-vision",
        "options": {"detail": "high"}
      },
      "tools": {
        "format": "custom-tools",
        "options": {"execution_mode": "sandbox"}
      }
    }
  }
}
```

## Summary

This composable capabilities architecture solves the fundamental issues in the current system:

1. **Eliminates Option Conflicts**: Each protocol has isolated options that only go to compatible capabilities
2. **Reduces Format Proliferation**: No need for multiple bundled formats; compose as needed
3. **Enables Explicit Protocol Support**: Only configured protocols are available
4. **Provides Clean Extensibility**: Add new formats without modifying existing code
5. **Maintains Provider Agnostic Design**: Capability formats represent API standards

The implementation maintains clean separation of concerns:
- **Capabilities**: Define protocol-specific behavior and validation
- **Models**: Compose capabilities and provide protocol-specific options
- **Transport**: Routes requests to appropriate capabilities
- **Providers**: Handle endpoint-specific communication

This architecture provides the flexibility needed for real-world model capabilities while maintaining type safety and clear error handling.
