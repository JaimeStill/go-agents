# AWS Bedrock Provider

This document details the changes required to support AWS Bedrock as a provider in go-agents. The implementation is structured in three stages, each with its own step sequence:

- **Stage 1: Extract Format Layer** — Separates wire formatting from providers into a dedicated `pkg/format` package
- **Stage 2: Response System Refactor** — Replaces OpenAI-shaped response types with a provider-neutral content block model
- **Stage 3: Converse Format + Bedrock Provider** — Implements the Bedrock Converse API format and provider on the new foundation

## Problem Context

AWS Bedrock provides access to Claude and other foundation models via the Converse API, which uses a completely different wire format than OpenAI-compatible APIs. The current architecture couples wire format concerns (request marshaling, response parsing) with transport concerns (endpoints, authentication, HTTP mechanics) inside the Provider interface. This works when all providers share the OpenAI format, but breaks down when Bedrock introduces the Converse API format.

The Converse API differences from OpenAI format include:
- Messages use typed content blocks (`text`, `image`, `toolUse`, `toolResult`) instead of string or structured content
- System prompts are a separate top-level field, not a message with role `"system"`
- Model parameters go into `inferenceConfig` instead of root-level fields
- Tool definitions use `toolSpec.inputSchema.json` instead of OpenAI's `function.parameters`
- Images are base64 bytes with explicit format field, not URLs in `image_url` maps
- Streaming uses AWS event stream binary framing, not Server-Sent Events
- Authentication uses SigV4 request signing instead of API keys or bearer tokens

## Architecture Approach

Extract formatting into a dedicated `pkg/format` package that owns request marshaling and response parsing. The provider interface slims down to transport and authentication only. Format becomes a first-class concept in the agent configuration alongside provider, model, and client.

### Updated Dependency Hierarchy

```
pkg/config          (foundation)
pkg/protocol        (protocol types)
pkg/response        (response types)
pkg/format          (format interface, data types, implementations, registry)
pkg/identities      (managed identity / credential sources)
pkg/providers       (provider implementations — transport/auth only)
pkg/model           (model runtime)
pkg/request         (request interface and types)
pkg/client          (HTTP orchestration)
pkg/agent           (high-level agent)
pkg/mock            (testing)
```

Key changes from current hierarchy:
- `pkg/format` sits between `pkg/response` and `pkg/identities`
- `pkg/providers` no longer imports `pkg/response` (format handles parsing)
- `pkg/request` imports `pkg/format` (for data types and Format interface)

### Updated Agent Configuration

```json
{
  "name": "bedrock-vision-agent",
  "system_prompt": "You are a vision analysis assistant.",
  "client": {
    "timeout": "30s",
    "retry": {
      "max_retries": 3,
      "initial_backoff": "1s",
      "max_backoff": "30s",
      "backoff_multiplier": 2.0,
      "jitter": true
    }
  },
  "provider": {
    "name": "bedrock",
    "base_url": "https://bedrock-runtime.us-east-1.amazonaws.com",
    "options": {
      "region": "us-east-1",
      "auth_type": "default"
    }
  },
  "format": "converse",
  "model": {
    "name": "anthropic.claude-sonnet-4-6-v1",
    "capabilities": {
      "chat": {
        "max_tokens": 4096,
        "temperature": 0.7
      },
      "vision": {
        "max_tokens": 4096,
        "temperature": 0.5
      },
      "tools": {
        "max_tokens": 4096
      }
    }
  }
}
```

Default format is `"openai"` — existing configurations that omit the format field work without changes.

---

## Stage 1: Extract Format Layer

Separates wire formatting from providers into a dedicated `pkg/format` package. Every step produces a compilable state. Changes proceed bottom-up through the dependency hierarchy.

### Step 1: Create `pkg/format/data.go`

Move data types from `pkg/providers/data.go` to `pkg/format/data.go`. These types are format input structures used by both the format layer and the request layer.

**New file: `pkg/format/data.go`**

```go
package format

import "github.com/JaimeStill/go-agents/pkg/protocol"

// ChatData contains the data needed to marshal a chat protocol request.
type ChatData struct {
	Model    string
	Messages []protocol.Message
	Options  map[string]any
}

// Image represents a provider-neutral image for vision requests.
// Consumers provide raw bytes and format from their source (disk, HTTP, render).
// Format implementations use these directly without re-encoding.
type Image struct {
	// Data holds raw image bytes.
	Data []byte
	// Format is the image format: "png", "jpeg", "webp", "gif".
	Format string
	// URL is an optional source URL. Formats that support URL references
	// (e.g., OpenAI) can use this directly instead of encoding Data.
	URL string
}

// VisionData contains the data needed to marshal a vision protocol request.
type VisionData struct {
	Model         string
	Messages      []protocol.Message
	Images        []Image
	VisionOptions map[string]any
	Options       map[string]any
}

// ToolsData contains the data needed to marshal a tools protocol request.
type ToolsData struct {
	Model    string
	Messages []protocol.Message
	Tools    []ToolDefinition
	Options  map[string]any
}

// ToolDefinition defines a function that the model can call.
type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// EmbeddingsData contains the data needed to marshal an embeddings protocol request.
type EmbeddingsData struct {
	Model   string
	Input   any
	Options map[string]any
}
```

### Step 2: Create `pkg/format/format.go`

Define the Format interface. This is the core abstraction that decouples wire format from transport.

**New file: `pkg/format/format.go`**

```go
package format

import (
	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/response"
)

// Format defines the interface for wire format implementations.
// Formats handle the translation between library data structures and
// provider-specific JSON formats for both requests and responses.
//
// The Marshal method converts protocol-specific data (ChatData, VisionData, etc.)
// into the wire format bytes. The Parse and ParseStreamChunk methods convert
// wire format bytes back into library response types.
type Format interface {
	// Name returns the format identifier.
	Name() string

	// Marshal converts request data to format-specific JSON bytes.
	// The data parameter should be *ChatData, *VisionData, *ToolsData,
	// or *EmbeddingsData based on the protocol.
	Marshal(p protocol.Protocol, data any) ([]byte, error)

	// Parse parses a standard response from JSON bytes into a library response type.
	// Returns *response.ChatResponse, *response.ToolsResponse,
	// or *response.EmbeddingsResponse based on the protocol.
	Parse(p protocol.Protocol, body []byte) (any, error)

	// ParseStreamChunk parses a single streaming chunk from data bytes.
	// The input is format-specific: JSON for OpenAI, event stream payload for Converse.
	ParseStreamChunk(p protocol.Protocol, data []byte) (*response.StreamingChunk, error)
}
```

### Step 3: Create `pkg/format/openai.go`

Extract the current `BaseProvider` marshaling logic and `response.Parse*` functions into the OpenAI format implementation. This is a direct extraction — the logic is identical to what currently exists in `pkg/providers/base.go` and `pkg/response/parse.go`.

**New file: `pkg/format/openai.go`**

```go
package format

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"maps"

	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/response"
)

// OpenAIFormat implements Format for OpenAI-compatible wire format.
// Used by Ollama, Azure, and any other OpenAI-compatible provider.
type OpenAIFormat struct{}

func (f *OpenAIFormat) Name() string {
	return "openai"
}

// Marshal converts request data to OpenAI-compatible JSON format.
func (f *OpenAIFormat) Marshal(proto protocol.Protocol, data any) ([]byte, error) {
	switch proto {
	case protocol.Chat:
		return f.marshalChat(data)
	case protocol.Vision:
		return f.marshalVision(data)
	case protocol.Tools:
		return f.marshalTools(data)
	case protocol.Embeddings:
		return f.marshalEmbeddings(data)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", proto)
	}
}

// Parse parses an OpenAI-compatible JSON response.
func (f *OpenAIFormat) Parse(proto protocol.Protocol, body []byte) (any, error) {
	switch proto {
	case protocol.Chat:
		return response.ParseChat(body)
	case protocol.Vision:
		return response.ParseVision(body)
	case protocol.Tools:
		return response.ParseTools(body)
	case protocol.Embeddings:
		return response.ParseEmbeddings(body)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", proto)
	}
}

// ParseStreamChunk parses an OpenAI-compatible streaming chunk.
func (f *OpenAIFormat) ParseStreamChunk(proto protocol.Protocol, data []byte) (*response.StreamingChunk, error) {
	switch proto {
	case protocol.Chat:
		return response.ParseChatStreamChunk(data)
	case protocol.Vision:
		return response.ParseVisionStreamChunk(data)
	case protocol.Tools:
		return response.ParseToolsStreamChunk(data)
	case protocol.Embeddings:
		return nil, fmt.Errorf("protocol %s does not support streaming", proto)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", proto)
	}
}

// OpenAIFactory creates an OpenAIFormat instance.
func OpenAIFactory() (Format, error) {
	return &OpenAIFormat{}, nil
}

func (f *OpenAIFormat) marshalChat(data any) ([]byte, error) {
	d, ok := data.(*ChatData)
	if !ok {
		return nil, fmt.Errorf("expected *ChatData, got %T", data)
	}

	combined := make(map[string]any)
	combined["model"] = d.Model
	combined["messages"] = d.Messages
	maps.Copy(combined, d.Options)
	return json.Marshal(combined)
}

func (f *OpenAIFormat) marshalVision(data any) ([]byte, error) {
	d, ok := data.(*VisionData)
	if !ok {
		return nil, fmt.Errorf("expected *VisionData, got %T", data)
	}

	if len(d.Messages) == 0 {
		return nil, fmt.Errorf("messages cannot be empty for vision requests")
	}

	if len(d.Images) == 0 {
		return nil, fmt.Errorf("images cannot be empty for vision requests")
	}

	// Transform the last message to embed images
	lastIdx := len(d.Messages) - 1
	message := d.Messages[lastIdx]

	var textContent string
	switch v := message.Content.(type) {
	case string:
		textContent = v
	default:
		return nil, fmt.Errorf("message content must be a string for vision transformation")
	}

	// Build structured content starting with text
	content := []map[string]any{
		{"type": "text", "text": textContent},
	}

	// Add each image with embedded options
	for _, img := range d.Images {
		// Resolve image to a URL string for OpenAI format
		var url string
		if img.URL != "" {
			url = img.URL
		} else {
			url = fmt.Sprintf("data:image/%s;base64,%s", img.Format, base64.StdEncoding.EncodeToString(img.Data))
		}

		imageURL := map[string]any{
			"url": url,
		}

		if d.VisionOptions != nil {
			maps.Copy(imageURL, d.VisionOptions)
		}

		content = append(content, map[string]any{
			"type":      "image_url",
			"image_url": imageURL,
		})
	}

	// Create transformed messages
	transformedMessages := make([]protocol.Message, len(d.Messages))
	copy(transformedMessages, d.Messages)
	transformedMessages[lastIdx] = protocol.Message{
		Role:    message.Role,
		Content: content,
	}

	// Combine model, messages, and options at root level
	combined := make(map[string]any)
	combined["model"] = d.Model
	combined["messages"] = transformedMessages
	maps.Copy(combined, d.Options)

	return json.Marshal(combined)
}

func (f *OpenAIFormat) marshalTools(data any) ([]byte, error) {
	d, ok := data.(*ToolsData)
	if !ok {
		return nil, fmt.Errorf("expected *ToolsData, got %T", data)
	}

	combined := make(map[string]any)
	combined["model"] = d.Model
	combined["messages"] = d.Messages

	// Transform tools to OpenAI format: {"type": "function", "function": {...}}
	openAITools := make([]map[string]any, len(d.Tools))
	for i, tool := range d.Tools {
		openAITools[i] = map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.Parameters,
			},
		}
	}
	combined["tools"] = openAITools

	maps.Copy(combined, d.Options)
	return json.Marshal(combined)
}

func (f *OpenAIFormat) marshalEmbeddings(data any) ([]byte, error) {
	d, ok := data.(*EmbeddingsData)
	if !ok {
		return nil, fmt.Errorf("expected *EmbeddingsData, got %T", data)
	}

	combined := make(map[string]any)
	combined["model"] = d.Model
	combined["input"] = d.Input
	maps.Copy(combined, d.Options)
	return json.Marshal(combined)
}
```

### Step 4: Create `pkg/format/registry.go`

Thread-safe factory registry mirroring the pattern in `pkg/providers/registry.go`.

**New file: `pkg/format/registry.go`**

```go
package format

import (
	"fmt"
	"sync"
)

// Factory is a function that creates a Format instance.
type Factory func() (Format, error)

type registry struct {
	factories map[string]Factory
	mu        sync.RWMutex
}

var register = &registry{
	factories: make(map[string]Factory),
}

// Register registers a format factory with the given name.
// This should be called during package initialization to register custom formats.
func Register(name string, factory Factory) {
	register.mu.Lock()
	defer register.mu.Unlock()
	register.factories[name] = factory
}

// Create creates a Format instance by name.
// Returns an error if the format name is not registered.
func Create(name string) (Format, error) {
	register.mu.RLock()
	factory, exists := register.factories[name]
	register.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unknown format: %s", name)
	}

	return factory()
}

// ListFormats returns a list of all registered format names.
func ListFormats() []string {
	register.mu.RLock()
	defer register.mu.RUnlock()

	names := make([]string, 0, len(register.factories))
	for name := range register.factories {
		names = append(names, name)
	}
	return names
}

func init() {
	Register("openai", OpenAIFactory)
}
```

### Step 5: Update `pkg/config/agent.go`

Add `Format` as a string field on `AgentConfig`. No `FormatConfig` struct or separate file needed — format is a simple name with no options.

**File: `pkg/config/agent.go`**

**Current code (line 12-18):**
```go
type AgentConfig struct {
	Name         string          `json:"name"`
	SystemPrompt string          `json:"system_prompt,omitempty"`
	Client       *ClientConfig   `json:"client,omitempty"`
	Provider     *ProviderConfig `json:"provider"`
	Model        *ModelConfig    `json:"model"`
}
```

**Updated code:**
```go
type AgentConfig struct {
	Name         string          `json:"name"`
	SystemPrompt string          `json:"system_prompt,omitempty"`
	Client       *ClientConfig   `json:"client,omitempty"`
	Provider     *ProviderConfig `json:"provider"`
	Format       string          `json:"format,omitempty"`
	Model        *ModelConfig    `json:"model"`
}
```

**Current code (line 21-29):**
```go
func DefaultAgentConfig() AgentConfig {
	return AgentConfig{
		Name:         "default-agent",
		SystemPrompt: "",
		Client:       DefaultClientConfig(),
		Provider:     DefaultProviderConfig(),
		Model:        DefaultModelConfig(),
	}
}
```

**Updated code:**
```go
func DefaultAgentConfig() AgentConfig {
	return AgentConfig{
		Name:         "default-agent",
		SystemPrompt: "",
		Client:       DefaultClientConfig(),
		Provider:     DefaultProviderConfig(),
		Format:       "openai",
		Model:        DefaultModelConfig(),
	}
}
```

**Add format merge logic after the provider merge block (after line 55):**
```go
	if source.Format != "" {
		c.Format = source.Format
	}
```

### Step 6: Update `pkg/providers/provider.go` and `pkg/providers/base.go`

Remove `Marshal`, `ProcessResponse`, and `ProcessStreamResponse` from the Provider interface. Remove marshaling methods from BaseProvider. Delete `pkg/providers/data.go`.

**File: `pkg/providers/provider.go`**

**Updated Provider interface:**
```go
// Provider defines the interface for LLM service provider implementations.
// Providers handle endpoint routing, authentication, and HTTP request
// preparation for their specific service. Wire format concerns (marshaling
// and response parsing) are handled by the Format interface.
type Provider interface {
	// Name returns the provider identifier.
	Name() string

	// BaseURL returns the provider's base URL.
	BaseURL() string

	// Endpoint returns the full endpoint URL for a protocol.
	// Returns an error if the protocol is not supported by this provider.
	Endpoint(p protocol.Protocol) (string, error)

	// SetHeaders sets provider-specific authentication and custom headers on an HTTP request.
	// This is called after the request is created but before it is executed.
	// Accepts context for operations that require I/O (e.g., token acquisition).
	// Returns an error if authentication setup fails.
	SetHeaders(ctx context.Context, req *http.Request) error

	// PrepareRequest creates a Request for standard (non-streaming) protocol execution.
	// Accepts pre-marshaled request body and headers from the request structure.
	PrepareRequest(ctx context.Context, p protocol.Protocol, body []byte, headers map[string]string) (*Request, error)

	// PrepareStreamRequest creates a Request for streaming protocol execution.
	// Accepts pre-marshaled request body and headers, adds streaming-specific headers.
	PrepareStreamRequest(ctx context.Context, p protocol.Protocol, body []byte, headers map[string]string) (*Request, error)
}
```

Note: The `Request` struct (URL, Headers, Body) remains in this file unchanged.

**File: `pkg/providers/base.go`**

Remove all `marshal*` methods. BaseProvider retains only `Name()` and `BaseURL()`:

**Updated file:**
```go
package providers

// BaseProvider provides common functionality for provider implementations.
// It stores the provider name and base URL.
// Provider implementations typically embed BaseProvider to inherit this functionality.
type BaseProvider struct {
	name    string
	baseURL string
}

// NewBaseProvider creates a new BaseProvider with the given name and base URL.
func NewBaseProvider(name, baseURL string) *BaseProvider {
	return &BaseProvider{
		name:    name,
		baseURL: baseURL,
	}
}

// Name returns the provider's identifier.
func (p *BaseProvider) Name() string {
	return p.name
}

// BaseURL returns the provider's base URL.
func (p *BaseProvider) BaseURL() string {
	return p.baseURL
}
```

**Delete file: `pkg/providers/data.go`** — types have moved to `pkg/format/data.go`.

### Step 7: Update `pkg/providers/ollama.go` and `pkg/providers/azure.go`

Remove `ProcessResponse` and `ProcessStreamResponse` methods. The SSE decoding logic in these methods needs to move somewhere accessible. Since SSE is a transport concern shared by multiple providers, introduce a shared utility.

**New file: `pkg/providers/sse.go`**

Extract the SSE reading logic used by both Ollama and Azure into a shared function:

```go
package providers

import (
	"bufio"
	"context"
	"io"
	"strings"
)

// SSELine represents a parsed Server-Sent Event line.
type SSELine struct {
	Data string
	Done bool
	Err  error
}

// ReadSSE reads Server-Sent Events from a reader and sends parsed data lines to a channel.
// Handles "data: " prefix stripping and "[DONE]" completion marker detection.
// Read errors are sent as SSELine with the Err field set.
// The channel is closed when the stream completes, an error occurs, or context is cancelled.
// The caller is responsible for closing the reader.
func ReadSSE(ctx context.Context, reader io.Reader) <-chan SSELine {
	output := make(chan SSELine)

	go func() {
		defer close(output)

		scanner := bufio.NewReader(reader)

		for {
			line, err := scanner.ReadString('\n')
			if err == io.EOF {
				return
			}
			if err != nil {
				select {
				case output <- SSELine{Err: err}:
				case <-ctx.Done():
				}
				return
			}

			line = strings.TrimSpace(line)

			if line == "" {
				continue
			}

			// Strip SSE "data: " prefix
			data, ok := strings.CutPrefix(line, "data: ")
			if !ok {
				continue
			}

			// Check for completion marker
			if data == "[DONE]" {
				select {
				case output <- SSELine{Done: true}:
				case <-ctx.Done():
				}
				return
			}

			select {
			case output <- SSELine{Data: data}:
			case <-ctx.Done():
				return
			}
		}
	}()

	return output
}
```

**File: `pkg/providers/ollama.go`**

Remove the `response` import (no longer needed). Remove `ProcessResponse` and `ProcessStreamResponse` methods entirely. The remaining methods are: `NewOllama`, `Endpoint`, `PrepareRequest`, `PrepareStreamRequest`, `SetHeaders`.

**Updated imports:**
```go
import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"strings"

	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/protocol"
)
```

Remove the `ProcessResponse` method (current lines 98-110) and `ProcessStreamResponse` method (current lines 117-174). All other methods remain unchanged.

**File: `pkg/providers/azure.go`**

Same treatment as Ollama. Remove `ProcessResponse` (current lines 129-141) and `ProcessStreamResponse` (current lines 148-206). Remove the `response` import. Remove `bufio`, `io`, and `strings` imports (only needed for removed methods). Keep `maps` if used elsewhere.

**Updated imports:**
```go
import (
	"context"
	"fmt"
	"maps"
	"net/http"

	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/identities"
	"github.com/JaimeStill/go-agents/pkg/protocol"
)
```

### Step 8: Update `pkg/request/`

Update Request interface to include `Format()`. Update all request types to hold a format reference and delegate `Marshal()` to format instead of provider.

**File: `pkg/request/interface.go`**

**Updated code:**
```go
package request

import (
	"github.com/JaimeStill/go-agents/pkg/format"
	"github.com/JaimeStill/go-agents/pkg/model"
	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/providers"
)

// Request defines the interface for protocol requests.
type Request interface {
	// Protocol returns the protocol identifier for this request.
	Protocol() protocol.Protocol

	// Headers returns the HTTP headers for this request.
	Headers() map[string]string

	// Marshal converts the request to JSON bytes using the format.
	Marshal() ([]byte, error)

	// Provider returns the provider for this request.
	Provider() providers.Provider

	// Format returns the format for this request.
	Format() format.Format

	// Model returns the model for this request.
	Model() *model.Model
}
```

**File: `pkg/request/chat.go`**

**Updated code:**
```go
package request

import (
	"github.com/JaimeStill/go-agents/pkg/format"
	"github.com/JaimeStill/go-agents/pkg/model"
	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/providers"
)

// ChatRequest represents a chat protocol request.
type ChatRequest struct {
	messages []protocol.Message
	options  map[string]any
	provider providers.Provider
	fmt      format.Format
	model    *model.Model
}

// NewChat creates a new ChatRequest with the given components.
func NewChat(p providers.Provider, f format.Format, m *model.Model, messages []protocol.Message, opts map[string]any) *ChatRequest {
	return &ChatRequest{
		messages: messages,
		options:  opts,
		provider: p,
		fmt:      f,
		model:    m,
	}
}

func (r *ChatRequest) Protocol() protocol.Protocol {
	return protocol.Chat
}

func (r *ChatRequest) Headers() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

// Marshal delegates to the format for wire format encoding.
func (r *ChatRequest) Marshal() ([]byte, error) {
	return r.fmt.Marshal(protocol.Chat, &format.ChatData{
		Model:    r.model.Name,
		Messages: r.messages,
		Options:  r.options,
	})
}

func (r *ChatRequest) Provider() providers.Provider {
	return r.provider
}

func (r *ChatRequest) Format() format.Format {
	return r.fmt
}

func (r *ChatRequest) Model() *model.Model {
	return r.model
}
```

**File: `pkg/request/vision.go`**

Same pattern. Constructor becomes `NewVision(p providers.Provider, f format.Format, m *model.Model, messages []protocol.Message, images []format.Image, visionOpts, opts map[string]any)`. The `images` field changes from `[]string` to `[]format.Image`. Marshal delegates to `r.fmt.Marshal(protocol.Vision, &format.VisionData{...})`. Add `fmt format.Format` field and `Format()` accessor.

**File: `pkg/request/tools.go`**

Same pattern. Constructor becomes `NewTools(p providers.Provider, f format.Format, m *model.Model, messages []protocol.Message, tools []format.ToolDefinition, opts map[string]any)`. Note: `ToolDefinition` now comes from `format` package, not `providers`. Marshal delegates to `r.fmt.Marshal(protocol.Tools, &format.ToolsData{...})`.

**File: `pkg/request/embeddings.go`**

Same pattern. Constructor becomes `NewEmbeddings(p providers.Provider, f format.Format, m *model.Model, input any, opts map[string]any)`. Marshal delegates to `r.fmt.Marshal(protocol.Embeddings, &format.EmbeddingsData{...})`.

### Step 9: Update `pkg/client/client.go`

Replace `provider.ProcessResponse()` with `req.Format().Parse()`. Replace `provider.ProcessStreamResponse()` with format-based stream parsing. The SSE decoding utility from `pkg/providers/sse.go` handles transport framing.

**File: `pkg/client/client.go`**

**Updated `execute` method — complete replacement (currently lines 86-150):**

The `provider.ProcessResponse()` call is replaced with `io.ReadAll` + `req.Format().Parse()`. Note that `body` and `err` are already declared earlier in the method (line 91: `body, err := req.Marshal()`), so the response read uses `=` assignment, not `:=`. The `respBody` name avoids reusing `body` for a different purpose.

```go
func (c *client) execute(ctx context.Context, req request.Request) (any, error) {
	provider := req.Provider()
	proto := req.Protocol()

	// Marshal request body through format
	body, err := req.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Prepare provider request
	providerRequest, err := provider.PrepareRequest(ctx, proto, body, req.Headers())
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		providerRequest.URL,
		bytes.NewBuffer(providerRequest.Body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	for key, value := range providerRequest.Headers {
		httpReq.Header.Set(key, value)
	}
	if err := provider.SetHeaders(ctx, httpReq); err != nil {
		return nil, fmt.Errorf("failed to set headers: %w", err)
	}

	// Execute HTTP request
	httpClient := c.HTTPClient()
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		c.setHealthy(false)
		return nil, err
	}
	defer resp.Body.Close()

	// Check for non-OK status - return HTTPStatusError for retry evaluation
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		c.setHealthy(false)
		return nil, &HTTPStatusError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       bodyBytes,
		}
	}

	// Read and parse response through format
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.setHealthy(false)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	result, err := req.Format().Parse(proto, respBody)
	if err != nil {
		c.setHealthy(false)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	c.setHealthy(true)
	return result, nil
}
```

Add `"io"` to the imports.

**Updated `executeStream` method — complete replacement (currently lines 174-246):**

The `provider.ProcessStreamResponse()` call and its channel adapter goroutine are replaced with `providers.ReadSSE` + `format.ParseStreamChunk`. The client now reads SSE frames directly and parses them through the format layer.

```go
func (c *client) executeStream(ctx context.Context, req request.Request) (<-chan *response.StreamingChunk, error) {
	provider := req.Provider()
	proto := req.Protocol()

	// Marshal request body through format
	body, err := req.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Prepare streaming request
	providerRequest, err := provider.PrepareStreamRequest(ctx, proto, body, req.Headers())
	if err != nil {
		return nil, fmt.Errorf("failed to prepare streaming request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		providerRequest.URL,
		bytes.NewBuffer(providerRequest.Body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	for key, value := range providerRequest.Headers {
		httpReq.Header.Set(key, value)
	}
	if err := provider.SetHeaders(ctx, httpReq); err != nil {
		return nil, fmt.Errorf("failed to set headers: %w", err)
	}

	// Execute HTTP request
	httpClient := c.HTTPClient()
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		c.setHealthy(false)
		return nil, fmt.Errorf("streaming request failed: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		c.setHealthy(false)
		return nil, fmt.Errorf("streaming request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Read SSE stream and parse chunks through format
	f := req.Format()
	sseLines := providers.ReadSSE(ctx, resp.Body)

	output := make(chan *response.StreamingChunk)
	go func() {
		defer close(output)
		defer resp.Body.Close()

		for line := range sseLines {
			if line.Err != nil {
				select {
				case output <- &response.StreamingChunk{Error: line.Err}:
				case <-ctx.Done():
				}
				c.setHealthy(false)
				return
			}

			if line.Done {
				c.setHealthy(true)
				return
			}

			chunk, err := f.ParseStreamChunk(proto, []byte(line.Data))
			if err != nil {
				continue
			}

			select {
			case output <- chunk:
			case <-ctx.Done():
				return
			}
		}
		c.setHealthy(true)
	}()

	return output, nil
```

Add `"github.com/JaimeStill/go-agents/pkg/providers"` to imports for `providers.ReadSSE`.

**Important note on streaming**: This default SSE-based streaming works for Ollama and Azure. For Bedrock, the Converse stream uses AWS event stream binary framing, not SSE. The Bedrock provider will need to handle this differently — see the Feature Phase streaming section below.

### Step 10: Update `pkg/agent/agent.go`

Update `New()` to create format from config. Pass format to all request constructors.

**File: `pkg/agent/agent.go`**

**Updated imports:**
```go
import (
	"context"
	"fmt"
	"maps"

	"github.com/JaimeStill/go-agents/pkg/client"
	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/format"
	"github.com/JaimeStill/go-agents/pkg/model"
	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/providers"
	"github.com/JaimeStill/go-agents/pkg/request"
	"github.com/JaimeStill/go-agents/pkg/response"
	"github.com/google/uuid"
)
```

**Updated Agent interface — add Format accessor (after Provider):**
```go
	// Format returns the format instance.
	Format() format.Format
```

**Updated agent struct (lines 69-75):**
```go
type agent struct {
	id           string
	client       client.Client
	provider     providers.Provider
	fmt          format.Format
	model        *model.Model
	systemPrompt string
}
```

**Updated New function (lines 81-97):**
```go
func New(cfg *config.AgentConfig) (Agent, error) {
	p, err := providers.Create(cfg.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	f, err := format.Create(cfg.Format)
	if err != nil {
		return nil, fmt.Errorf("failed to create format: %w", err)
	}

	m := model.New(cfg.Model)
	c := client.New(cfg.Client)

	return &agent{
		id:           uuid.Must(uuid.NewV7()).String(),
		client:       c,
		provider:     p,
		fmt:          f,
		model:        m,
		systemPrompt: cfg.SystemPrompt,
	}, nil
}
```

**Add Format accessor:**
```go
func (a *agent) Format() format.Format {
	return a.fmt
}
```

**Update all request constructors** to pass format. For example, Chat (line 126):

**Current:**
```go
	req := request.NewChat(a.provider, a.model, messages, options)
```

**Updated:**
```go
	req := request.NewChat(a.provider, a.fmt, a.model, messages, options)
```

Apply the same change to:
- `ChatStream` (line 150)
- `Vision` (line 173) — also update signature from `images []string` to `images []format.Image`
- `VisionStream` (line 207) — also update signature from `images []string` to `images []format.Image`
- `Tools` (line 230) — also update `ToolDefinition` reference from `providers.ToolDefinition` to `format.ToolDefinition`
- `Embed` (line 251)

The Agent interface `Vision` and `VisionStream` method signatures also change:
```go
	Vision(ctx context.Context, prompt string, images []format.Image, opts ...map[string]any) (*response.ChatResponse, error)
	VisionStream(ctx context.Context, prompt string, images []format.Image, opts ...map[string]any) (<-chan *response.StreamingChunk, error)
```

**Update Tools method tool conversion (lines 221-228):**

**Current:**
```go
	toolDefs := make([]providers.ToolDefinition, len(tools))
	for i, tool := range tools {
		toolDefs[i] = providers.ToolDefinition{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  tool.Parameters,
		}
	}
```

**Updated:**
```go
	toolDefs := make([]format.ToolDefinition, len(tools))
	for i, tool := range tools {
		toolDefs[i] = format.ToolDefinition{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  tool.Parameters,
		}
	}
```

### Step 11: Update `pkg/mock/`

**File: `pkg/mock/provider.go`**

Remove `Marshal`, `ProcessResponse`, and `ProcessStreamResponse` methods and their associated fields/options. Remove the `response` import.

**Remove fields (from struct, lines 22-29):**
- `marshalResponse`
- `marshalError`
- `processResponse`
- `processError`
- `streamChunks`
- `streamError`

**Remove option functions:**
- `WithMarshalResponse`
- `WithProcessResponse`
- `WithProviderStreamChunks`

**Remove methods:**
- `Marshal`
- `ProcessResponse`
- `ProcessStreamResponse`

**New file: `pkg/mock/format.go`**

```go
package mock

import (
	"github.com/JaimeStill/go-agents/pkg/format"
	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/response"
)

// MockFormat implements format.Format for testing.
type MockFormat struct {
	name string

	marshalResponse    []byte
	marshalError       error
	parseResponse      any
	parseError         error
	streamChunk        *response.StreamingChunk
	streamChunkError   error
}

// MockFormatOption configures a MockFormat.
type MockFormatOption func(*MockFormat)

// NewMockFormat creates a new MockFormat with default configuration.
func NewMockFormat(opts ...MockFormatOption) *MockFormat {
	m := &MockFormat{
		name: "mock-format",
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// WithFormatName sets the format name.
func WithFormatName(name string) MockFormatOption {
	return func(m *MockFormat) {
		m.name = name
	}
}

// WithFormatMarshalResponse sets the response for Marshal.
func WithFormatMarshalResponse(body []byte, err error) MockFormatOption {
	return func(m *MockFormat) {
		m.marshalResponse = body
		m.marshalError = err
	}
}

// WithFormatParseResponse sets the response for Parse.
func WithFormatParseResponse(resp any, err error) MockFormatOption {
	return func(m *MockFormat) {
		m.parseResponse = resp
		m.parseError = err
	}
}

// WithFormatStreamChunk sets the response for ParseStreamChunk.
func WithFormatStreamChunk(chunk *response.StreamingChunk, err error) MockFormatOption {
	return func(m *MockFormat) {
		m.streamChunk = chunk
		m.streamChunkError = err
	}
}

func (m *MockFormat) Name() string {
	return m.name
}

func (m *MockFormat) Marshal(proto protocol.Protocol, data any) ([]byte, error) {
	if m.marshalError != nil {
		return nil, m.marshalError
	}

	if m.marshalResponse != nil {
		return m.marshalResponse, nil
	}

	return []byte(`{}`), nil
}

func (m *MockFormat) Parse(proto protocol.Protocol, body []byte) (any, error) {
	if m.parseError != nil {
		return nil, m.parseError
	}

	if m.parseResponse != nil {
		return m.parseResponse, nil
	}

	return response.ParseChat(body)
}

func (m *MockFormat) ParseStreamChunk(proto protocol.Protocol, data []byte) (*response.StreamingChunk, error) {
	if m.streamChunkError != nil {
		return nil, m.streamChunkError
	}

	if m.streamChunk != nil {
		return m.streamChunk, nil
	}

	return response.ParseChatStreamChunk(data)
}

// Verify MockFormat implements format.Format interface.
var _ format.Format = (*MockFormat)(nil)
```

**File: `pkg/mock/agent.go`**

Add `mockFormat format.Format` field, `NewMockFormat()` default in constructor, `WithFormat` option, and `Format()` accessor:

**Add field to MockAgent struct (after `mockProvider`):**
```go
	mockFormat   format.Format
```

**Add default in `NewMockAgent` constructor (after `mockProvider: NewMockProvider()`):**
```go
		mockFormat:   NewMockFormat(),
```

**Add option function (after `WithProvider`):**
```go
// WithFormat sets a custom format.
func WithFormat(f format.Format) MockAgentOption {
	return func(m *MockAgent) {
		m.mockFormat = f
	}
}
```

**Add accessor (after `Provider()`):**
```go
// Format returns the mock format.
func (m *MockAgent) Format() format.Format {
	return m.mockFormat
}
```

**Add import:**
```go
	"github.com/JaimeStill/go-agents/pkg/format"
```

### Step 12: Update Tests

Update all tests bottom-up following the refactoring order. Create `tests/format/` directory with tests for OpenAI format marshaling and parsing (migrated from current `tests/providers/base_test.go`). Update provider tests to remove marshal/process assertions. Update request tests for new constructor signatures.

### Step 13: Validation

```bash
go vet ./...
go test ./tests/...
```

All packages should compile and existing tests should pass with the updated signatures.

---

## Stage 2: Response System Refactor

Replaces the OpenAI-shaped response types with a provider-neutral content block model. The current response types (`ChatResponse`, `ToolsResponse`, `StreamingChunk`) mirror OpenAI's wire format with anonymous struct fields, `Choices` arrays, and separate type hierarchies for chat vs tool responses. This makes programmatic construction from non-OpenAI formats awkward and verbose.

The new model uses typed content blocks — the pattern used natively by Anthropic, Bedrock, and Gemini. OpenAI is the only outlier, and its responses translate cleanly into content blocks.

### Design Decisions

- **Single `Response` type** for Chat, Vision, and Tools protocols. Content blocks determine what the response contains — `TextBlock` for text, `ToolUseBlock` for tool calls. A response can contain both.
- **`EmbeddingsResponse` remains separate** — vector data is fundamentally different from conversational content.
- **Tool call `Input` is `map[string]any`** (parsed). Three of four providers return parsed objects natively. The OpenAI format `json.Unmarshal`s the arguments string during `Parse`.
- **No `Choices` array** — no provider other than OpenAI supports multiple completions, and it's rarely used. The response *is* the single completion.
- **Normalized `TokenUsage`** — `InputTokens`, `OutputTokens`, `TotalTokens` across all providers.
- **Streaming uses the same `Response` type** with partial content blocks, plus an `Error` field for transport errors.

### Step 1: Rewrite `pkg/response/` types

**File: `pkg/response/usage.go`**

Normalize field names:

```go
package response

// TokenUsage tracks token consumption for a request/response cycle.
type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}
```

**New file: `pkg/response/content.go`**

Define the content block types:

```go
package response

// ContentBlock represents a single unit of content in a response.
// Use TextBlock for text content and ToolUseBlock for tool calls.
type ContentBlock interface {
	blockType() string
}

// TextBlock contains text content from the model.
type TextBlock struct {
	Text string
}

func (b TextBlock) blockType() string { return "text" }

// ToolUseBlock contains a tool call requested by the model.
type ToolUseBlock struct {
	ID    string
	Name  string
	Input map[string]any
}

func (b ToolUseBlock) blockType() string { return "tool_use" }
```

**File: `pkg/response/response.go`** (replaces `chat.go` and `tools.go`)

```go
package response

// Response represents the result of a Chat, Vision, or Tools protocol request.
// Content blocks determine what the response contains — TextBlock for text,
// ToolUseBlock for tool calls, or both for mixed responses.
type Response struct {
	Role       string
	Content    []ContentBlock
	StopReason string
	Usage      *TokenUsage
}

// Text returns the concatenated text from all TextBlocks in the response.
// Returns empty string if there are no text blocks.
func (r *Response) Text() string {
	var text string
	for _, block := range r.Content {
		if tb, ok := block.(TextBlock); ok {
			text += tb.Text
		}
	}
	return text
}

// ToolCalls returns all ToolUseBlocks from the response.
// Returns nil if there are no tool use blocks.
func (r *Response) ToolCalls() []ToolUseBlock {
	var calls []ToolUseBlock
	for _, block := range r.Content {
		if tb, ok := block.(ToolUseBlock); ok {
			calls = append(calls, tb)
		}
	}
	return calls
}
```

**File: `pkg/response/streaming.go`** (rewritten)

```go
package response

// StreamingResponse represents a partial response from a streaming request.
// Content blocks accumulate as deltas arrive.
// The Error field is set for transport errors during streaming.
type StreamingResponse struct {
	Content    []ContentBlock
	StopReason string
	Usage      *TokenUsage
	Error      error
}

// Text returns the concatenated text from all TextBlocks in the chunk.
func (r *StreamingResponse) Text() string {
	var text string
	for _, block := range r.Content {
		if tb, ok := block.(TextBlock); ok {
			text += tb.Text
		}
	}
	return text
}
```

**File: `pkg/response/embeddings.go`** (simplified)

```go
package response

// EmbeddingsResponse represents the result of an Embeddings protocol request.
type EmbeddingsResponse struct {
	Embeddings [][]float64
	Model      string
	Usage      *TokenUsage
}
```

**Delete files:** `chat.go`, `tools.go`

**File: `pkg/response/parse.go`** (simplified)

The `Parse` and `ParseStreamChunk` functions are removed — parsing is now entirely the format layer's responsibility. This file can be deleted or reduced to just the `doc.go` package comment.

### Step 2: Update `pkg/format/format.go`

Update the `Format` interface return types:

```go
type Format interface {
	Name() string
	Marshal(p protocol.Protocol, data any) ([]byte, error)
	Parse(p protocol.Protocol, body []byte) (any, error)
	ParseStreamChunk(p protocol.Protocol, data []byte) (*response.StreamingResponse, error)
}
```

`Parse` still returns `any` because it returns either `*response.Response` or `*response.EmbeddingsResponse` depending on protocol.

### Step 3: Update `pkg/format/openai.go`

Rewrite `Parse` and `ParseStreamChunk` to construct the new response types. The marshal side is unchanged. Define named unexported wire format types that describe the OpenAI response shape, then use them in the parse methods.

**Wire format types** — define these before the parse methods:

```go
// OpenAI wire format types for response deserialization.

type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
	Usage   *openAIUsage   `json:"usage,omitempty"`
}

type openAIChoice struct {
	Message      openAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason,omitempty"`
}

type openAIMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []openAIToolCall `json:"tool_calls,omitempty"`
}

type openAIToolCall struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Function openAIToolFunction `json:"function"`
}

type openAIToolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type openAIStreamChunk struct {
	Choices []openAIStreamChoice `json:"choices"`
}

type openAIStreamChoice struct {
	Delta        openAIDelta `json:"delta"`
	FinishReason *string     `json:"finish_reason"`
}

type openAIDelta struct {
	Content string `json:"content"`
}

type openAIEmbeddingsResponse struct {
	Data  []openAIEmbedding `json:"data"`
	Model string            `json:"model"`
	Usage *openAIUsage      `json:"usage,omitempty"`
}

type openAIEmbedding struct {
	Embedding []float64 `json:"embedding"`
}
```

**Parse routing:**

```go
func (f *OpenAIFormat) Parse(proto protocol.Protocol, body []byte) (any, error) {
	switch proto {
	case protocol.Chat, protocol.Vision:
		return f.parseChat(body)
	case protocol.Tools:
		return f.parseTools(body)
	case protocol.Embeddings:
		return f.parseEmbeddings(body)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", proto)
	}
}
```

**Parse for Chat/Vision** — translate `choices[0].message` into content blocks:

```go
func (f *OpenAIFormat) parseChat(body []byte) (*response.Response, error) {
	var raw openAIResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse chat response: %w", err)
	}

	resp := &response.Response{Role: "assistant"}

	if len(raw.Choices) > 0 {
		choice := raw.Choices[0]
		if choice.Message.Content != "" {
			resp.Content = append(resp.Content, response.TextBlock{
				Text: choice.Message.Content,
			})
		}
		resp.StopReason = choice.FinishReason
	}

	resp.Usage = f.mapUsage(raw.Usage)

	return resp, nil
}
```

**Parse for Tools** — translate `tool_calls` into `ToolUseBlock`s, parsing the arguments JSON string into `map[string]any`:

```go
func (f *OpenAIFormat) parseTools(body []byte) (*response.Response, error) {
	var raw openAIResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse tools response: %w", err)
	}

	resp := &response.Response{Role: "assistant"}

	if len(raw.Choices) > 0 {
		choice := raw.Choices[0]

		if choice.Message.Content != "" {
			resp.Content = append(resp.Content, response.TextBlock{
				Text: choice.Message.Content,
			})
		}

		for _, tc := range choice.Message.ToolCalls {
			var input map[string]any
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &input); err != nil {
				return nil, fmt.Errorf("failed to parse tool arguments: %w", err)
			}

			resp.Content = append(resp.Content, response.ToolUseBlock{
				ID:    tc.ID,
				Name:  tc.Function.Name,
				Input: input,
			})
		}

		resp.StopReason = choice.FinishReason
	}

	resp.Usage = f.mapUsage(raw.Usage)

	return resp, nil
}
```

**Parse for Embeddings:**

```go
func (f *OpenAIFormat) parseEmbeddings(body []byte) (*response.EmbeddingsResponse, error) {
	var raw openAIEmbeddingsResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse embeddings response: %w", err)
	}

	embeddings := make([][]float64, len(raw.Data))
	for i, d := range raw.Data {
		embeddings[i] = d.Embedding
	}

	resp := &response.EmbeddingsResponse{
		Embeddings: embeddings,
		Model:      raw.Model,
		Usage:      f.mapUsage(raw.Usage),
	}

	return resp, nil
}
```

**ParseStreamChunk:**

```go
func (f *OpenAIFormat) ParseStreamChunk(proto protocol.Protocol, data []byte) (*response.StreamingResponse, error) {
	var raw openAIStreamChunk
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse streaming chunk: %w", err)
	}

	chunk := &response.StreamingResponse{}

	if len(raw.Choices) > 0 {
		choice := raw.Choices[0]
		if choice.Delta.Content != "" {
			chunk.Content = append(chunk.Content, response.TextBlock{
				Text: choice.Delta.Content,
			})
		}
		if choice.FinishReason != nil {
			chunk.StopReason = *choice.FinishReason
		}
	}

	return chunk, nil
}
```

**Shared usage mapping helper:**

```go
func (f *OpenAIFormat) mapUsage(usage *openAIUsage) *response.TokenUsage {
	if usage == nil {
		return nil
	}

	return &response.TokenUsage{
		InputTokens:  usage.PromptTokens,
		OutputTokens: usage.CompletionTokens,
		TotalTokens:  usage.TotalTokens,
	}
}
```

### Step 4: Update `pkg/client/client.go`

Update the streaming chunk type in `executeStream`:

**Current:**
```go
	output := make(chan *response.StreamingChunk)
```

**Updated:**
```go
	output := make(chan *response.StreamingResponse)
```

Update the `ExecuteStream` return type:
```go
	ExecuteStream(ctx context.Context, req request.Request) (<-chan *response.StreamingResponse, error)
```

Update the error chunk construction in the SSE error handler:
```go
	case output <- &response.StreamingResponse{Error: line.Err}:
```

### Step 5: Update `pkg/agent/agent.go`

Return types change for all protocol methods:

```go
	Chat(ctx context.Context, prompt string, opts ...map[string]any) (*response.Response, error)
	ChatStream(ctx context.Context, prompt string, opts ...map[string]any) (<-chan *response.StreamingResponse, error)
	Vision(ctx context.Context, prompt string, images []format.Image, opts ...map[string]any) (*response.Response, error)
	VisionStream(ctx context.Context, prompt string, images []format.Image, opts ...map[string]any) (<-chan *response.StreamingResponse, error)
	Tools(ctx context.Context, prompt string, tools []Tool, opts ...map[string]any) (*response.Response, error)
```

Update the type assertions in the method implementations. For Chat/Vision:

**Current:**
```go
	resp, ok := result.(*response.ChatResponse)
```

**Updated:**
```go
	resp, ok := result.(*response.Response)
```

For Tools — same change, now returns `*response.Response` instead of `*response.ToolsResponse`.

### Step 6: Update `pkg/mock/`

#### `pkg/mock/agent.go`

**Struct fields** — unify the separate chat/vision/tools response fields into a single `*response.Response`, and update streaming chunks:

Current fields:
```go
	chatResponse       *response.ChatResponse
	visionResponse     *response.ChatResponse
	toolsResponse      *response.ToolsResponse
	embeddingsResponse *response.EmbeddingsResponse
	streamChunks       []response.StreamingChunk
```

Updated fields:
```go
	chatResponse       *response.Response
	visionResponse     *response.Response
	toolsResponse      *response.Response
	embeddingsResponse *response.EmbeddingsResponse
	streamChunks       []response.StreamingResponse
```

**Constructor default** — update the `streamChunks` initialization:

Current:
```go
	streamChunks: []response.StreamingChunk{},
```

Updated:
```go
	streamChunks: []response.StreamingResponse{},
```

**Option functions** — update parameter types:

`WithChatResponse`:
```go
func WithChatResponse(resp *response.Response, err error) MockAgentOption {
```

`WithVisionResponse`:
```go
func WithVisionResponse(resp *response.Response, err error) MockAgentOption {
```

`WithToolsResponse`:
```go
func WithToolsResponse(resp *response.Response, err error) MockAgentOption {
```

`WithStreamChunks`:
```go
func WithStreamChunks(chunks []response.StreamingResponse, err error) MockAgentOption {
```

`WithEmbeddingsResponse` is unchanged — `*response.EmbeddingsResponse` stays the same.

**Protocol method return types:**

`Chat` and `Vision` — change return type from `*response.ChatResponse` to `*response.Response`:
```go
func (m *MockAgent) Chat(ctx context.Context, prompt string, opts ...map[string]any) (*response.Response, error) {
	return m.chatResponse, m.chatError
}

func (m *MockAgent) Vision(ctx context.Context, prompt string, images []format.Image, opts ...map[string]any) (*response.Response, error) {
	return m.visionResponse, m.visionError
}
```

`Tools` — change return type from `*response.ToolsResponse` to `*response.Response`:
```go
func (m *MockAgent) Tools(ctx context.Context, prompt string, tools []agent.Tool, opts ...map[string]any) (*response.Response, error) {
	return m.toolsResponse, m.toolsError
}
```

`ChatStream` and `VisionStream` — change channel type from `*response.StreamingChunk` to `*response.StreamingResponse`:
```go
func (m *MockAgent) ChatStream(ctx context.Context, prompt string, opts ...map[string]any) (<-chan *response.StreamingResponse, error) {
	if m.streamError != nil {
		return nil, m.streamError
	}

	ch := make(chan *response.StreamingResponse, len(m.streamChunks))
	for i := range m.streamChunks {
		ch <- &m.streamChunks[i]
	}
	close(ch)

	return ch, nil
}
```

Same pattern for `VisionStream`.

`Embed` is unchanged.

#### `pkg/mock/format.go`

**Struct field** — change `streamChunk` type:

Current:
```go
	streamChunk      *response.StreamingChunk
	streamChunkError error
```

Updated:
```go
	streamChunk      *response.StreamingResponse
	streamChunkError error
```

**Option function** — update `WithFormatStreamChunk` parameter:
```go
func WithFormatStreamChunk(chunk *response.StreamingResponse, err error) MockFormatOption {
```

**`Parse` default fallback** — the current fallback calls `response.ParseChat(body)` which no longer exists. Replace with returning a minimal `*response.Response`:
```go
func (m *MockFormat) Parse(proto protocol.Protocol, body []byte) (any, error) {
	if m.parseError != nil {
		return nil, m.parseError
	}

	if m.parseResponse != nil {
		return m.parseResponse, nil
	}

	return &response.Response{Role: "assistant"}, nil
}
```

**`ParseStreamChunk` return type and default fallback** — update return type and remove the `response.ParseChatStreamChunk` call:
```go
func (m *MockFormat) ParseStreamChunk(proto protocol.Protocol, data []byte) (*response.StreamingResponse, error) {
	if m.streamChunkError != nil {
		return nil, m.streamChunkError
	}

	if m.streamChunk != nil {
		return m.streamChunk, nil
	}

	return &response.StreamingResponse{}, nil
}
```

#### `pkg/mock/helpers.go`

All helper constructors simplify dramatically with the new response types.

**`NewSimpleChatAgent`:**
```go
func NewSimpleChatAgent(id string, content string) *MockAgent {
	return NewMockAgent(
		WithID(id),
		WithChatResponse(&response.Response{
			Role: "assistant",
			Content: []response.ContentBlock{
				response.TextBlock{Text: content},
			},
		}, nil),
	)
}
```

**`NewStreamingChatAgent`:**
```go
func NewStreamingChatAgent(id string, chunks []string) *MockAgent {
	streamChunks := make([]response.StreamingResponse, len(chunks))
	for i, content := range chunks {
		streamChunks[i] = response.StreamingResponse{
			Content: []response.ContentBlock{
				response.TextBlock{Text: content},
			},
		}
	}

	return NewMockAgent(
		WithID(id),
		WithStreamChunks(streamChunks, nil),
	)
}
```

**`NewToolsAgent`** — parameter changes from `[]response.ToolCall` to `[]response.ToolUseBlock`:
```go
func NewToolsAgent(id string, toolCalls []response.ToolUseBlock) *MockAgent {
	content := make([]response.ContentBlock, len(toolCalls))
	for i, tc := range toolCalls {
		content[i] = tc
	}

	return NewMockAgent(
		WithID(id),
		WithToolsResponse(&response.Response{
			Role:    "assistant",
			Content: content,
		}, nil),
	)
}
```

**`NewEmbeddingsAgent`:**
```go
func NewEmbeddingsAgent(id string, embedding []float64) *MockAgent {
	return NewMockAgent(
		WithID(id),
		WithEmbeddingsResponse(&response.EmbeddingsResponse{
			Embeddings: [][]float64{embedding},
			Model:      "mock-model",
		}, nil),
	)
}
```

**`NewMultiProtocolAgent`:**
```go
func NewMultiProtocolAgent(id string) *MockAgent {
	chatResponse := &response.Response{
		Role: "assistant",
		Content: []response.ContentBlock{
			response.TextBlock{Text: "Mock chat response"},
		},
	}

	toolsResponse := &response.Response{
		Role:    "assistant",
		Content: []response.ContentBlock{},
	}

	return NewMockAgent(
		WithID(id),
		WithChatResponse(chatResponse, nil),
		WithVisionResponse(chatResponse, nil),
		WithToolsResponse(toolsResponse, nil),
		WithEmbeddingsResponse(&response.EmbeddingsResponse{
			Embeddings: [][]float64{{0.1, 0.2, 0.3}},
			Model:      "mock-model",
		}, nil),
	)
}
```

**`NewFailingAgent`** is unchanged — it only passes `nil` responses and errors.

**Remove `protocol` import** from helpers.go — it's no longer needed since we don't call `protocol.NewMessage`.

Note how the anonymous struct gymnastics are completely eliminated. The new `Response` type with content blocks makes construction clean and readable.

### Step 7: Update tests

All test files that reference `ChatResponse`, `ToolsResponse`, `StreamingChunk`, or `TokenUsage` fields need updating:

- `tests/format/openai_test.go` — Parse tests assert on `*response.Response`, use `resp.Text()` and `resp.ToolCalls()`
- `tests/agent/agent_test.go` — type assertions change from `*response.ChatResponse` to `*response.Response`
- `tests/client/client_test.go` — same type assertion changes
- `tests/mock/agent_test.go` — construct `Response` instead of `ChatResponse`/`ToolsResponse`
- `tests/mock/format_test.go` — update mock parse response types
- `tests/response/response_test.go` — rewrite for new types, test `Text()` and `ToolCalls()` methods
- `tools/prompt-agent/main.go` — use `resp.Text()` instead of `resp.Content()`, update usage field names

### Step 8: Validation

```bash
go vet ./...
go test ./tests/...
```

---

## Stage 3: Converse Format + Bedrock Provider

### Step 1: Create `pkg/format/converse.go`

Implement the Converse API format for AWS Bedrock. This handles the translation between library data types and Bedrock's Converse wire format.

**New file: `pkg/format/converse.go`**

```go
package format

import (
	"encoding/json"
	"fmt"
	"maps"

	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/response"
)

// ConverseFormat implements Format for the AWS Bedrock Converse API.
type ConverseFormat struct{}

func (f *ConverseFormat) Name() string {
	return "converse"
}

// ConverseFactory creates a ConverseFormat instance.
func ConverseFactory() (Format, error) {
	return &ConverseFormat{}, nil
}
```

**Marshal — Chat protocol:**

Translates `ChatData` to Converse request format:
- System messages extracted to top-level `system` field as `[{"text": "..."}]`
- Non-system messages use `content` blocks: `[{"text": "..."}]`
- Options mapped to `inferenceConfig`: `maxTokens`, `temperature`, `topP`, `stopSequences`

```go
func (f *ConverseFormat) Marshal(proto protocol.Protocol, data any) ([]byte, error) {
	switch proto {
	case protocol.Chat:
		return f.marshalChat(data)
	case protocol.Vision:
		return f.marshalVision(data)
	case protocol.Tools:
		return f.marshalTools(data)
	case protocol.Embeddings:
		return nil, fmt.Errorf("embeddings protocol not supported by Converse API")
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", proto)
	}
}

func (f *ConverseFormat) marshalChat(data any) ([]byte, error) {
	d, ok := data.(*ChatData)
	if !ok {
		return nil, fmt.Errorf("expected *ChatData, got %T", data)
	}

	req := make(map[string]any)
	req["modelId"] = d.Model

	system, messages := f.splitSystemMessages(d.Messages)
	if len(system) > 0 {
		req["system"] = system
	}
	req["messages"] = messages

	maps.Copy(req, d.Options)

	return json.Marshal(req)
}
```

**Marshal — Vision protocol:**

Translates `VisionData` to Converse format. Images must be base64 bytes with format specification. URLs are not directly supported by Converse — images passed as URLs would need to be fetched and converted, or passed as S3 URIs.

```go
func (f *ConverseFormat) marshalVision(data any) ([]byte, error) {
	d, ok := data.(*VisionData)
	if !ok {
		return nil, fmt.Errorf("expected *VisionData, got %T", data)
	}

	if len(d.Messages) == 0 {
		return nil, fmt.Errorf("messages cannot be empty for vision requests")
	}

	if len(d.Images) == 0 {
		return nil, fmt.Errorf("images cannot be empty for vision requests")
	}

	req := make(map[string]any)
	req["modelId"] = d.Model

	system, messages := f.splitSystemMessages(d.Messages)
	if len(system) > 0 {
		req["system"] = system
	}

	// Transform last message to include image content blocks
	lastIdx := len(messages) - 1
	lastMsg := messages[lastIdx].(map[string]any)
	content := lastMsg["content"].([]map[string]any)

	for _, img := range d.Images {
		content = append(content, f.buildImageBlock(img))
	}
	lastMsg["content"] = content
	messages[lastIdx] = lastMsg

	req["messages"] = messages

	maps.Copy(req, d.Options)

	return json.Marshal(req)
}
```

**Marshal — Tools protocol:**

Translates `ToolsData` to Converse format with `toolConfig`:

```go
func (f *ConverseFormat) marshalTools(data any) ([]byte, error) {
	d, ok := data.(*ToolsData)
	if !ok {
		return nil, fmt.Errorf("expected *ToolsData, got %T", data)
	}

	req := make(map[string]any)
	req["modelId"] = d.Model

	system, messages := f.splitSystemMessages(d.Messages)
	if len(system) > 0 {
		req["system"] = system
	}
	req["messages"] = messages

	// Build toolConfig
	tools := make([]map[string]any, len(d.Tools))
	for i, tool := range d.Tools {
		tools[i] = map[string]any{
			"toolSpec": map[string]any{
				"name":        tool.Name,
				"description": tool.Description,
				"inputSchema": map[string]any{
					"json": tool.Parameters,
				},
			},
		}
	}
	req["toolConfig"] = map[string]any{
		"tools": tools,
	}

	maps.Copy(req, d.Options)

	return json.Marshal(req)
}
```

**Helper methods:**

```go
// splitSystemMessages separates system messages from conversation messages.
// Returns system content blocks and transformed conversation messages.
func (f *ConverseFormat) splitSystemMessages(msgs []protocol.Message) ([]map[string]any, []any) {
	var system []map[string]any
	var messages []any

	for _, msg := range msgs {
		if msg.Role == "system" {
			if text, ok := msg.Content.(string); ok {
				system = append(system, map[string]any{"text": text})
			}
			continue
		}

		content := []map[string]any{}
		if text, ok := msg.Content.(string); ok {
			content = append(content, map[string]any{"text": text})
		}

		messages = append(messages, map[string]any{
			"role":    msg.Role,
			"content": content,
		})
	}

	return system, messages
}

// buildImageBlock converts an Image to a Converse image content block.
// Uses Data and Format directly — no encoding or decoding needed.
func (f *ConverseFormat) buildImageBlock(img Image) map[string]any {
	return map[string]any{
		"image": map[string]any{
			"format": img.Format,
			"source": map[string]any{
				"bytes": img.Data,
			},
		},
	}
}
```

**Wire format types** — define these before the parse methods:

```go
// Converse wire format types for response deserialization.

type converseResponse struct {
	Output     converseOutput `json:"output"`
	StopReason string         `json:"stopReason"`
	Usage      converseUsage  `json:"usage"`
}

type converseOutput struct {
	Message converseMessage `json:"message"`
}

type converseMessage struct {
	Role    string                `json:"role"`
	Content []converseContentBlock `json:"content"`
}

type converseContentBlock struct {
	Text    string              `json:"text,omitempty"`
	ToolUse *converseToolUse    `json:"toolUse,omitempty"`
}

type converseToolUse struct {
	ToolUseID string         `json:"toolUseId"`
	Name      string         `json:"name"`
	Input     map[string]any `json:"input"`
}

type converseUsage struct {
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
	TotalTokens  int `json:"totalTokens"`
}

// Converse streaming event types.

type converseStreamEvent struct {
	ContentBlockDelta *converseContentBlockDelta `json:"contentBlockDelta,omitempty"`
	MessageStop       *converseMessageStop       `json:"messageStop,omitempty"`
}

type converseContentBlockDelta struct {
	ContentBlockIndex int            `json:"contentBlockIndex"`
	Delta             converseDelta  `json:"delta"`
}

type converseDelta struct {
	Text string `json:"text"`
}

type converseMessageStop struct {
	StopReason string `json:"stopReason"`
}
```

**Parse routing:**

```go
func (f *ConverseFormat) Parse(proto protocol.Protocol, body []byte) (any, error) {
	switch proto {
	case protocol.Chat, protocol.Vision, protocol.Tools:
		return f.parseResponse(body)
	case protocol.Embeddings:
		return nil, fmt.Errorf("embeddings protocol not supported by Converse API")
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", proto)
	}
}
```

Note: Chat, Vision, and Tools all use the same parse method. Converse returns the same response shape regardless of whether the model produced text, tool calls, or both — content blocks are the universal unit. This is a natural fit for the unified `Response` type.

**Parse response** — translate Converse content blocks directly to library content blocks:

```go
func (f *ConverseFormat) parseResponse(body []byte) (*response.Response, error) {
	var raw converseResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse Converse response: %w", err)
	}

	resp := &response.Response{
		Role:       raw.Output.Message.Role,
		StopReason: raw.StopReason,
		Usage: &response.TokenUsage{
			InputTokens:  raw.Usage.InputTokens,
			OutputTokens: raw.Usage.OutputTokens,
			TotalTokens:  raw.Usage.TotalTokens,
		},
	}

	for _, block := range raw.Output.Message.Content {
		if block.Text != "" {
			resp.Content = append(resp.Content, response.TextBlock{
				Text: block.Text,
			})
		}
		if block.ToolUse != nil {
			resp.Content = append(resp.Content, response.ToolUseBlock{
				ID:    block.ToolUse.ToolUseID,
				Name:  block.ToolUse.Name,
				Input: block.ToolUse.Input,
			})
		}
	}

	return resp, nil
}
```

**ParseStreamChunk — Converse streaming:**

AWS ConverseStream returns event stream events. By the time `ParseStreamChunk` is called, the event stream binary framing has been decoded (by the Bedrock provider's stream handler) and the JSON payload is passed in.

```go
func (f *ConverseFormat) ParseStreamChunk(proto protocol.Protocol, data []byte) (*response.StreamingResponse, error) {
	var event converseStreamEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("failed to parse stream event: %w", err)
	}

	if event.ContentBlockDelta != nil {
		return &response.StreamingResponse{
			Content: []response.ContentBlock{
				response.TextBlock{Text: event.ContentBlockDelta.Delta.Text},
			},
		}, nil
	}

	if event.MessageStop != nil {
		return &response.StreamingResponse{
			StopReason: event.MessageStop.StopReason,
		}, nil
	}

	// Other event types (messageStart, contentBlockStart, contentBlockStop, metadata)
	// are informational — return nil to skip
	return nil, nil
}
```

**Register Converse format in `pkg/format/registry.go` init():**

```go
func init() {
	Register("openai", OpenAIFactory)
	Register("converse", ConverseFactory)
}
```

### Step 2: Create `pkg/identities/aws.go`

AWS credential source with SigV4 signing capability. Mirrors the `azure.go` pattern, isolating the AWS SDK dependency below the provider layer.

**New file: `pkg/identities/aws.go`**

```go
package identities

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

// AWSAuthType defines the authentication method for AWS credentials.
type AWSAuthType string

const (
	// AWSAuthDefault uses the AWS default credential chain
	// (env vars, shared config, instance roles, ECS task roles).
	AWSAuthDefault AWSAuthType = "default"

	// AWSAuthStatic uses explicit access key, secret key, and optional session token.
	AWSAuthStatic AWSAuthType = "static"

	// AWSAuthProfile uses a named AWS profile from shared config.
	AWSAuthProfile AWSAuthType = "profile"
)

// AWSCredentialSource provides AWS credentials and SigV4 request signing.
// Supports the default credential chain, static credentials, and named profiles.
type AWSCredentialSource struct {
	creds  aws.CredentialsProvider
	signer *v4.Signer
	region string
}

// NewAWSCredentialSource creates a new AWSCredentialSource.
// options may contain:
//   - "access_key_id", "secret_access_key", "session_token" (for AWSAuthStatic)
//   - "profile" (for AWSAuthProfile)
func NewAWSCredentialSource(ctx context.Context, region string, authType AWSAuthType, options map[string]any) (*AWSCredentialSource, error) {
	var creds aws.CredentialsProvider

	switch authType {
	case AWSAuthDefault, "":
		cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
		if err != nil {
			return nil, fmt.Errorf("load default AWS config: %w", err)
		}
		creds = cfg.Credentials

	case AWSAuthStatic:
		accessKey, _ := options["access_key_id"].(string)
		secretKey, _ := options["secret_access_key"].(string)
		sessionToken, _ := options["session_token"].(string)

		if accessKey == "" || secretKey == "" {
			return nil, fmt.Errorf("access_key_id and secret_access_key are required for static auth")
		}

		creds = credentials.NewStaticCredentialsProvider(accessKey, secretKey, sessionToken)

	case AWSAuthProfile:
		profile, _ := options["profile"].(string)
		if profile == "" {
			return nil, fmt.Errorf("profile is required for profile auth")
		}

		cfg, err := awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithRegion(region),
			awsconfig.WithSharedConfigProfile(profile),
		)
		if err != nil {
			return nil, fmt.Errorf("load AWS config for profile %q: %w", profile, err)
		}
		creds = cfg.Credentials

	default:
		return nil, fmt.Errorf("unsupported AWS auth_type: %q", authType)
	}

	return &AWSCredentialSource{
		creds:  creds,
		signer: v4.NewSigner(),
		region: region,
	}, nil
}

// SignRequest signs an HTTP request using AWS SigV4.
// The service parameter identifies the AWS service for the signing scope (e.g., "bedrock").
// This reads the request body to compute the payload hash,
// then replaces the body so the HTTP client can read it again.
func (s *AWSCredentialSource) SignRequest(ctx context.Context, req *http.Request, service string) error {
	// Read the body for signing
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return fmt.Errorf("read request body for signing: %w", err)
		}
		// Replace the body for the HTTP client to read
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	// Compute payload hash
	hash := sha256.Sum256(bodyBytes)
	payloadHash := fmt.Sprintf("%x", hash)

	// Retrieve credentials
	cred, err := s.creds.Retrieve(ctx)
	if err != nil {
		return fmt.Errorf("retrieve AWS credentials: %w", err)
	}

	// Sign the request
	err = s.signer.SignHTTP(ctx, cred, req, payloadHash, service, s.region, time.Now())
	if err != nil {
		return fmt.Errorf("sign request: %w", err)
	}

	return nil
}
```

### Step 3: Create `pkg/streaming/` and update existing providers

Extract streaming transport into a dedicated package. Move SSE decoding from `pkg/providers/sse.go`, define the `StreamReader` interface, add the AWS event stream implementation, and wire existing providers to use it.

**New file: `pkg/streaming/streaming.go`**

```go
package streaming

import (
	"context"
	"io"
)

// StreamLine represents a decoded frame from a streaming transport.
// Data contains the payload bytes, Done signals stream completion,
// and Err signals a transport error.
type StreamLine struct {
	Data []byte
	Done bool
	Err  error
}

// StreamReader decodes a provider-specific streaming transport into StreamLines.
// SSE and AWS event stream are the two built-in implementations.
type StreamReader interface {
	ReadStream(ctx context.Context, reader io.Reader) <-chan StreamLine
}
```

**New file: `pkg/streaming/sse.go`**

Move logic from `pkg/providers/sse.go` into a `StreamReader` implementation:

```go
package streaming

import (
	"bufio"
	"context"
	"io"
	"strings"
)

// SSEMedia is the content type for Server-Sent Events streams.
const SSEMedia = "text/event-stream"

// SSEReader decodes Server-Sent Events streams.
// Used by OpenAI-compatible providers (Ollama, Azure).
type SSEReader struct{}

func NewSSEReader() *SSEReader {
	return &SSEReader{}
}

func (r *SSEReader) ReadStream(ctx context.Context, reader io.Reader) <-chan StreamLine {
	output := make(chan StreamLine)

	go func() {
		defer close(output)

		scanner := bufio.NewReader(reader)

		for {
			line, err := scanner.ReadString('\n')
			if err == io.EOF {
				return
			}
			if err != nil {
				select {
				case output <- StreamLine{Err: err}:
				case <-ctx.Done():
				}
				return
			}

			line = strings.TrimSpace(line)

			if line == "" {
				continue
			}

			data, ok := strings.CutPrefix(line, "data: ")
			if !ok {
				continue
			}

			if data == "[DONE]" {
				select {
				case output <- StreamLine{Done: true}:
				case <-ctx.Done():
				}
				return
			}

			select {
			case output <- StreamLine{Data: []byte(data)}:
			case <-ctx.Done():
				return
			}
		}
	}()

	return output
}
```

**New file: `pkg/streaming/eventstream.go`**

AWS event stream binary framing decoder. Uses `github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream` for frame decoding.

```go
package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream"
)

// EventStreamMedia is the content type for AWS event stream responses.
const EventStreamMedia = "application/vnd.amazon.eventstream"

// EventStreamReader decodes AWS event stream binary framing.
// Used by the Bedrock provider for ConverseStream responses.
type EventStreamReader struct{}

func NewEventStreamReader() *EventStreamReader {
	return &EventStreamReader{}
}

func (r *EventStreamReader) ReadStream(ctx context.Context, reader io.Reader) <-chan StreamLine {
	output := make(chan StreamLine)

	go func() {
		defer close(output)

		decoder := eventstream.NewDecoder()

		for {
			msg, err := decoder.Decode(reader, nil)
			if err == io.EOF {
				return
			}
			if err != nil {
				select {
				case output <- StreamLine{Err: fmt.Errorf("decode event stream frame: %w", err)}:
				case <-ctx.Done():
				}
				return
			}

			// Extract message type and event type from headers
			var messageType, eventType string
			for _, h := range msg.Headers {
				switch h.Name {
				case ":message-type":
					messageType = h.Value.String()
				case ":event-type":
					eventType = h.Value.String()
				}
			}

			// Handle exceptions
			if messageType == "exception" {
				select {
				case output <- StreamLine{Err: fmt.Errorf("event stream exception: %s: %s", eventType, string(msg.Payload))}:
				case <-ctx.Done():
				}
				return
			}

			// Wrap the payload in a JSON envelope keyed by event type
			// so the format's ParseStreamChunk can identify the event
			envelope, err := json.Marshal(map[string]json.RawMessage{
				eventType: msg.Payload,
			})
			if err != nil {
				continue
			}

			select {
			case output <- StreamLine{Data: envelope}:
			case <-ctx.Done():
				return
			}
		}
	}()

	return output
}
```

Note: The event stream decoder wraps each payload in a `{eventType: payload}` JSON envelope. This matches what the Converse format's `ParseStreamChunk` expects — it checks for `contentBlockDelta`, `messageStop`, etc. as top-level keys.

**Delete file: `pkg/providers/sse.go`** — logic moved to `pkg/streaming/sse.go`.

**Update `pkg/providers/provider.go`** — add `Stream()` to the Provider interface:

```go
	// Stream returns the streaming transport reader for this provider.
	// Used by the client to decode streaming responses.
	Stream() streaming.StreamReader
```

Add import: `"github.com/JaimeStill/go-agents/pkg/streaming"`

**Update `pkg/providers/ollama.go`** — add stream field and accessor:

Add field to struct:
```go
type OllamaProvider struct {
	*BaseProvider
	options map[string]any
	stream  streaming.StreamReader
}
```

Initialize in constructor:
```go
	return &OllamaProvider{
		BaseProvider: NewBaseProvider(c.Name, baseURL),
		options:      c.Options,
		stream:       streaming.NewSSEReader(),
	}, nil
```

Add accessor:
```go
func (p *OllamaProvider) Stream() streaming.StreamReader {
	return p.stream
}
```

Add import: `"github.com/JaimeStill/go-agents/pkg/streaming"`

Also update `PrepareStreamRequest` to use the constant: replace `"text/event-stream"` with `streaming.SSEMedia`.

**Update `pkg/providers/azure.go`** — same pattern as Ollama:

Add `stream streaming.StreamReader` field, initialize with `streaming.NewSSEReader()`, add `Stream()` accessor. Update `PrepareStreamRequest` to use `streaming.SSEMedia`.

**Update `pkg/mock/provider.go`** — add stream field with SSE default and accessor:

Add field:
```go
	stream streaming.StreamReader
```

Initialize in constructor:
```go
	stream: streaming.NewSSEReader(),
```

Add accessor:
```go
func (m *MockProvider) Stream() streaming.StreamReader {
	return m.stream
}
```

Update `PrepareStreamRequest` to use `streaming.SSEMedia` instead of `"text/event-stream"`.

**Update `pkg/client/client.go`** — replace `providers.ReadSSE` with `provider.Stream().ReadStream`:

Replace:
```go
	f := req.Format()
	sseLines := providers.ReadSSE(ctx, resp.Body)
```

With:
```go
	f := req.Format()
	streamLines := provider.Stream().ReadStream(ctx, resp.Body)
```

Update the goroutine to use `StreamLine`:
```go
	output := make(chan *response.StreamingResponse)
	go func() {
		defer close(output)
		defer resp.Body.Close()

		for line := range streamLines {
			if line.Err != nil {
				select {
				case output <- &response.StreamingResponse{Error: line.Err}:
				case <-ctx.Done():
				}
				c.setHealthy(false)
				return
			}

			if line.Done {
				c.setHealthy(true)
				return
			}

			chunk, err := f.ParseStreamChunk(proto, line.Data)
			if err != nil || chunk == nil {
				continue
			}

			select {
			case output <- chunk:
			case <-ctx.Done():
				return
			}
		}
		c.setHealthy(true)
	}()
```

Remove `"github.com/JaimeStill/go-agents/pkg/providers"` import from client.go (no longer needed for `ReadSSE`). Note: `line.Data` is already `[]byte` — no `[]byte(line.Data)` conversion needed.

**Update `pkg/streaming/` dependency hierarchy:**

```
pkg/streaming    (StreamReader interface, SSE, EventStream implementations)
pkg/providers    (imports streaming — providers hold a StreamReader)
pkg/client       (reads StreamReader from provider via request)
```

---

### Step 4: Create `pkg/providers/bedrock.go`

Bedrock provider handles transport and authentication only. Endpoint construction includes the model ID in the URL path. Authentication delegates to the AWS credential source for SigV4 signing. Uses `EventStreamReader` for streaming transport.

**New file: `pkg/providers/bedrock.go`**

```go
package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"strings"

	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/identities"
	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/streaming"
)

const bedrockService = "bedrock"

// BedrockProvider implements Provider for AWS Bedrock.
// Handles endpoint routing with model ID in the URL path and SigV4 authentication.
type BedrockProvider struct {
	*BaseProvider
	region     string
	credSource *identities.AWSCredentialSource
	stream     streaming.StreamReader
}

// NewBedrock creates a new BedrockProvider from configuration.
// Requires "region" in options. Auth type defaults to "default" (AWS credential chain).
// Supports "default", "static", and "profile" auth types.
func NewBedrock(c *config.ProviderConfig) (Provider, error) {
	region, ok := c.Options["region"].(string)
	if !ok || region == "" {
		return nil, fmt.Errorf("region is required for Bedrock provider")
	}

	// Auto-construct base URL from region if not provided
	baseURL := c.BaseURL
	if baseURL == "" {
		baseURL = fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com", region)
	}

	authTypeStr, _ := c.Options["auth_type"].(string)
	authType := identities.AWSAuthType(authTypeStr)

	credSource, err := identities.NewAWSCredentialSource(context.Background(), region, authType, c.Options)
	if err != nil {
		return nil, fmt.Errorf("initialize AWS credentials: %w", err)
	}

	return &BedrockProvider{
		BaseProvider: NewBaseProvider(c.Name, baseURL),
		region:       region,
		credSource:   credSource,
		stream:       streaming.NewEventStreamReader(),
	}, nil
}

// Stream returns the event stream reader for Bedrock's binary streaming transport.
func (p *BedrockProvider) Stream() streaming.StreamReader {
	return p.stream
}

// Endpoint returns the Bedrock endpoint path template.
// The actual URL is constructed in PrepareRequest using the model ID from the body.
// Chat, Vision, and Tools use the Converse API. Embeddings is not supported.
func (p *BedrockProvider) Endpoint(proto protocol.Protocol) (string, error) {
	switch proto {
	case protocol.Chat, protocol.Vision, protocol.Tools:
		return "/model/%s/converse", nil
	case protocol.Embeddings:
		return "", fmt.Errorf("embeddings protocol not supported by Bedrock Converse API")
	default:
		return "", fmt.Errorf("protocol %s not supported by Bedrock", proto)
	}
}

// PrepareRequest creates a Bedrock request with model ID in the URL path.
// Extracts modelId from the marshaled body to construct the full endpoint URL.
func (p *BedrockProvider) PrepareRequest(ctx context.Context, proto protocol.Protocol, body []byte, headers map[string]string) (*Request, error) {
	pathTemplate, err := p.Endpoint(proto)
	if err != nil {
		return nil, err
	}

	modelID, err := extractModelID(body)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s"+pathTemplate, p.BaseURL(), modelID)

	return &Request{
		URL:     url,
		Headers: headers,
		Body:    body,
	}, nil
}

// PrepareStreamRequest creates a Bedrock streaming request.
// Uses /converse-stream endpoint instead of /converse.
func (p *BedrockProvider) PrepareStreamRequest(ctx context.Context, proto protocol.Protocol, body []byte, headers map[string]string) (*Request, error) {
	modelID, err := extractModelID(body)
	if err != nil {
		return nil, err
	}

	switch proto {
	case protocol.Chat, protocol.Vision, protocol.Tools:
		// Use converse-stream endpoint
	default:
		return nil, fmt.Errorf("protocol %s does not support streaming on Bedrock", proto)
	}

	url := fmt.Sprintf("%s/model/%s/converse-stream", p.BaseURL(), modelID)

	streamHeaders := make(map[string]string)
	maps.Copy(streamHeaders, headers)
	streamHeaders["Accept"] = streaming.EventStreamMedia

	return &Request{
		URL:     url,
		Headers: streamHeaders,
		Body:    body,
	}, nil
}

// SetHeaders applies SigV4 signing to the request.
// Must be called after all other headers are set, as signing covers the full request.
func (p *BedrockProvider) SetHeaders(ctx context.Context, req *http.Request) error {
	return p.credSource.SignRequest(ctx, req, bedrockService)
}

// extractModelID extracts the modelId field from a marshaled Converse request body.
func extractModelID(body []byte) (string, error) {
	var envelope struct {
		ModelID string `json:"modelId"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return "", fmt.Errorf("failed to extract modelId from request body: %w", err)
	}
	if envelope.ModelID == "" {
		return "", fmt.Errorf("modelId is empty in request body")
	}

	return envelope.ModelID, nil
}
```

**Register Bedrock in `pkg/providers/registry.go` init():**

```go
func init() {
	Register("ollama", NewOllama)
	Register("azure", NewAzure)
	Register("bedrock", NewBedrock)
}
```

### Step 5: Tests

**New: `tests/streaming/sse_test.go`** — Test SSEReader with mock data: text lines, `[DONE]` completion, read errors, empty lines skipped, non-data lines skipped.

**New: `tests/streaming/eventstream_test.go`** — Test EventStreamReader with mock binary frames: event payloads wrapped in JSON envelope, exception handling, EOF completion.

**New: `tests/format/converse_test.go`** — Test Converse `Marshal` for Chat, Vision, Tools. Test system message extraction. Test image block building. Test `Parse` for Chat and Tools responses. Test `ParseStreamChunk` for contentBlockDelta and messageStop events, and `nil, nil` for skippable events.

**New: `tests/identities/aws_test.go`** — Test `NewAWSCredentialSource` with static credentials. Test `SignRequest` adds authorization headers.

**New: `tests/providers/bedrock_test.go`** — Test `NewBedrock` with valid/invalid configs. Test `Endpoint` for each protocol. Test `PrepareRequest` URL construction with model ID extraction. Test `PrepareStreamRequest` uses converse-stream endpoint and event stream Accept header. Test `Stream()` returns `EventStreamReader`.

**Updated: `tests/providers/ollama_test.go`** — Test `Stream()` returns `SSEReader`.

**Updated: `tests/providers/azure_test.go`** — Test `Stream()` returns `SSEReader`.

---

## New Dependencies

```
github.com/aws/aws-sdk-go-v2
github.com/aws/aws-sdk-go-v2/config
github.com/aws/aws-sdk-go-v2/credentials
github.com/aws/aws-sdk-go-v2/aws/signer/v4
github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream
```

---

## Configuration Examples

### Bedrock + Converse (Claude)

```json
{
  "name": "bedrock-claude",
  "system_prompt": "You are a helpful assistant.",
  "provider": {
    "name": "bedrock",
    "options": {
      "region": "us-east-1",
      "auth_type": "default"
    }
  },
  "format": "converse",
  "model": {
    "name": "anthropic.claude-sonnet-4-6-v1",
    "capabilities": {
      "chat": { "max_tokens": 4096, "temperature": 0.7 },
      "vision": { "max_tokens": 4096, "temperature": 0.5 },
      "tools": { "max_tokens": 4096 }
    }
  }
}
```

### Bedrock with Static Credentials

```json
{
  "name": "bedrock-static",
  "provider": {
    "name": "bedrock",
    "options": {
      "region": "us-west-2",
      "auth_type": "static",
      "access_key_id": "AKIA...",
      "secret_access_key": "...",
      "session_token": "..."
    }
  },
  "format": "converse",
  "model": {
    "name": "anthropic.claude-haiku-4-5-20251001-v1:0",
    "capabilities": {
      "chat": { "max_tokens": 2048, "temperature": 0.7 }
    }
  }
}
```

### Bedrock with Named Profile

```json
{
  "name": "bedrock-profile",
  "provider": {
    "name": "bedrock",
    "options": {
      "region": "us-east-1",
      "auth_type": "profile",
      "profile": "dev-account"
    }
  },
  "format": "converse",
  "model": {
    "name": "anthropic.claude-sonnet-4-6-v1",
    "capabilities": {
      "chat": { "max_tokens": 4096 }
    }
  }
}
```

### Ollama Backward Compatibility (no format specified)

```json
{
  "name": "local-ollama",
  "provider": {
    "name": "ollama",
    "base_url": "http://localhost:11434"
  },
  "model": {
    "name": "llama3.2:3b",
    "capabilities": {
      "chat": { "temperature": 0.7, "max_tokens": 4096 }
    }
  }
}
```

This config works without changes — `DefaultFormatConfig()` provides `"openai"` format.

---

## Verification

After each step:
```bash
go build ./...
go test ./tests/...
```

End-to-end validation (after all steps complete):
1. Existing Ollama configs work with no format specified (backward compat)
2. Existing Azure configs work with no format specified (backward compat)
3. New Bedrock config with Converse format creates agent successfully
4. `format.ListFormats()` returns `["openai", "converse"]`
5. `providers.ListProviders()` returns `["ollama", "azure", "bedrock"]`
6. Bedrock Chat, Vision, and Tools produce correct Converse API JSON
7. Converse response parsing produces valid `ChatResponse` / `ToolsResponse`
8. All tests pass with 80%+ coverage maintained
