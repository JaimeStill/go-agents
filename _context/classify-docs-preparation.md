# Classify-Docs Preparation: Architecture Simplification and Component Extraction

## Overview

This implementation guide prepares the go-agents architecture for the classify-docs production project by:

1. **Simplifying the capabilities system** - Remove validation and option management, keeping only transformation logic
2. **Implementing transport-layer retry** - Add intelligent retry with HTTP status and network error detection
3. **Streamlining configuration** - Remove format-specific validation, flatten to pure options map
4. **Preparing for library extraction** - Identify and prepare components for go-agents-document-context

This refactoring follows the **Preparation Phase** pattern: architectural changes without adding new features, ensuring each step results in a compilable state.

---

## Target Architecture

### Current State
- Capabilities validate options and apply format-specific rules
- No retry infrastructure in transport layer (only unused config fields)
- Configuration has `CapabilityConfig.Format` field for validation
- classify-docs has all document/processing infrastructure embedded

### Target State
- Capabilities only transform requests/responses (no validation)
- Transport layer handles retry with intelligent error detection
- Configuration simplified: capability name IS the format
- Clean separation of concerns for library extraction

---

## Phase 1: Simplify Capabilities System

### Goal
Remove validation and option management from capabilities while preserving transformation logic.

### Step 1.1: Update Capability Interface

**File**: `pkg/capabilities/capability.go`

**Current Interface** (lines 27-41):
```go
type Capability interface {
	Name() string
	Protocol() protocols.Protocol
	Options() []string
	ValidateOptions(options map[string]any) error
	ProcessOptions(options map[string]any) (map[string]any, error)
	CreateRequest(req *CapabilityRequest, model string) (protocols.Request, error)
	ParseResponse(body []byte) (any, error)
	SupportsStreaming() bool
}
```

**Replace with**:
```go
type Capability interface {
	// Identity
	Protocol() protocols.Protocol

	// Request/Response Transformation
	CreateRequest(req *CapabilityRequest, model string) (*protocols.Request, error)
	ParseResponse(body []byte) (any, error)

	// Streaming Support
	SupportsStreaming() bool
}
```

**Rationale**:
- Removed `Options()`, `ValidateOptions()`, and `ProcessOptions()` - options flow through untouched
- Removed redundant `Name() string` method - protocol is the capability identity
- `Protocol()` returns `protocols.Protocol` - semantic match between method name and return type

### Step 1.2: Update StreamingCapability Interface

**File**: `pkg/capabilities/capability.go`

**Current Interface** (lines 43-51):
```go
type StreamingCapability interface {
	Capability
	CreateStreamingRequest(req *CapabilityRequest, model string) (*protocols.Request, error)
	ParseStreamingChunk(line []byte) (any, error)
	IsStreamComplete(chunk any) bool
}
```

**Keep As-Is**: No changes needed for streaming interface (pointer return already correct).

### Step 1.3: Remove StandardCapability Base Implementation

**File**: `pkg/capabilities/capability.go`

**Delete** (lines 53-138): Remove entire `StandardCapability` type and all its methods:
- `StandardCapability` struct
- `newStandardCapability()`
- All method implementations (`Name()`, `Protocol()`, `Options()`, `ValidateOptions()`, `ProcessOptions()`, `CreateRequest()`, `ParseResponse()`, `SupportsStreaming()`)

**Rationale**: Base implementation no longer needed without validation logic.

### Step 1.4: Remove StandardStreamingCapability Base Implementation

**File**: `pkg/capabilities/capability.go`

**Delete** (lines 140-244): Remove entire `StandardStreamingCapability` type and all its methods:
- `StandardStreamingCapability` struct
- `newStandardStreamingCapability()`
- All method implementations
- SSE parsing helpers

**Rationale**: Streaming capabilities will implement interface directly.

### Step 1.5: Update Chat Capability

**File**: `pkg/capabilities/chat.go`

**Current Structure**:
```go
type chat struct {
	*StandardCapability
}

func newChat() Capability {
	return &chat{
		StandardCapability: newStandardCapability(
			"chat",
			protocols.Chat,
			[]string{/* option names */},
			map[string]any{/* defaults */},
			[]string{/* required */},
		),
	}
}
```

**Replace Entire File**:
```go
package capabilities

import (
	"encoding/json"
	"fmt"

	"github.com/JaimeStill/go-agents/pkg/protocols"
)

// chat implements the Capability interface for chat completions.
type chat struct{}

// NewChat creates a new chat capability.
// Exported for direct instantiation by model package.
func NewChat() Capability {
	return &chat{}
}

// Protocol returns the protocol this capability implements.
func (c *chat) Protocol() protocols.Protocol {
	return protocols.Chat
}

// CreateRequest creates a chat request from a capability request.
func (c *chat) CreateRequest(req *CapabilityRequest, model string) (*protocols.Request, error) {
	req.Options["model"] = model

	return &protocols.Request{
		Messages: req.Messages,
		Options:  req.Options,
	}, nil
}

// ParseResponse parses a chat response.
func (c *chat) ParseResponse(body []byte) (any, error) {
	var response protocols.ChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse chat response: %w", err)
	}
	return response, nil
}

// SupportsStreaming returns whether this capability supports streaming.
func (c *chat) SupportsStreaming() bool {
	return true
}

// CreateStreamingRequest creates a streaming chat request.
func (c *chat) CreateStreamingRequest(req *CapabilityRequest, model string) (*protocols.Request, error) {
	req.Options["model"] = model
	req.Options["stream"] = true

	return &protocols.Request{
		Messages: req.Messages,
		Options:  req.Options,
	}, nil
}

// ParseStreamingChunk parses a streaming chunk.
func (c *chat) ParseStreamingChunk(line []byte) (any, error) {
	var chunk protocols.StreamingChunk
	if err := json.Unmarshal(line, &chunk); err != nil {
		return nil, fmt.Errorf("failed to parse streaming chunk: %w", err)
	}
	return &chunk, nil
}

// IsStreamComplete checks if streaming is complete.
func (c *chat) IsStreamComplete(chunk any) bool {
	if sc, ok := chunk.(*protocols.StreamingChunk); ok {
		if len(sc.Choices) > 0 {
			// FinishReason is *string, check for non-nil and non-empty
			return sc.Choices[0].FinishReason != nil && *sc.Choices[0].FinishReason != ""
		}
	}
	return false
}
```

**Changes**:
- Removed embedding of `StandardCapability`
- Implement all interface methods directly
- Removed redundant `Name() string` method - only `Protocol()` remains for identity
- Factory function exported: `NewChat()` instead of `newChat()`
- No validation or option processing
- Add model name directly to options map
- Return pointer to `protocols.Request`
- Preserve streaming support
- Handle `FinishReason` as `*string` pointer in `IsStreamComplete`

### Step 1.6: Update Vision Capability

**File**: `pkg/capabilities/vision.go`

**Replace Entire File**:
```go
package capabilities

import (
	"encoding/json"
	"fmt"

	"github.com/JaimeStill/go-agents/pkg/protocols"
)

// vision implements the Capability interface for vision completions.
type vision struct{}

// NewVision creates a new vision capability.
// Exported for direct instantiation by model package.
func NewVision() Capability {
	return &vision{}
}

// Protocol returns the protocol this capability implements.
func (v *vision) Protocol() protocols.Protocol {
	return protocols.Vision
}

// CreateRequest creates a vision request from a capability request.
func (v *vision) CreateRequest(req *CapabilityRequest, model string) (*protocols.Request, error) {
	req.Options["model"] = model

	return &protocols.Request{
		Messages: req.Messages,
		Options:  req.Options,
	}, nil
}

// ParseResponse parses a vision response.
func (v *vision) ParseResponse(body []byte) (any, error) {
	var response protocols.ChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse vision response: %w", err)
	}
	return response, nil
}

// SupportsStreaming returns whether this capability supports streaming.
func (v *vision) SupportsStreaming() bool {
	return true
}

// CreateStreamingRequest creates a streaming vision request.
func (v *vision) CreateStreamingRequest(req *CapabilityRequest, model string) (*protocols.Request, error) {
	req.Options["model"] = model
	req.Options["stream"] = true

	return &protocols.Request{
		Messages: req.Messages,
		Options:  req.Options,
	}, nil
}

// ParseStreamingChunk parses a streaming chunk.
func (v *vision) ParseStreamingChunk(line []byte) (any, error) {
	var chunk protocols.StreamingChunk
	if err := json.Unmarshal(line, &chunk); err != nil {
		return nil, fmt.Errorf("failed to parse streaming chunk: %w", err)
	}
	return &chunk, nil
}

// IsStreamComplete checks if streaming is complete.
func (v *vision) IsStreamComplete(chunk any) bool {
	if sc, ok := chunk.(*protocols.StreamingChunk); ok {
		if len(sc.Choices) > 0 {
			// FinishReason is *string, check for non-nil and non-empty
			return sc.Choices[0].FinishReason != nil && *sc.Choices[0].FinishReason != ""
		}
	}
	return false
}
```

**Changes**:
- Same pattern as chat.go
- `Protocol()` returns `protocols.Vision`
- Factory function exported: `NewVision()`
- Return pointer to `protocols.Request`
- Correct pointer handling in `IsStreamComplete`

### Step 1.7: Update Tools Capability

**File**: `pkg/capabilities/tools.go`

**Replace Entire File**:
```go
package capabilities

import (
	"encoding/json"
	"fmt"

	"github.com/JaimeStill/go-agents/pkg/protocols"
)

// tools implements the Capability interface for tool/function calling.
type tools struct{}

// NewTools creates a new tools capability.
// Exported for direct instantiation by model package.
func NewTools() Capability {
	return &tools{}
}

// Protocol returns the protocol this capability implements.
func (t *tools) Protocol() protocols.Protocol {
	return protocols.Tools
}

// CreateRequest creates a tools request from a capability request.
func (t *tools) CreateRequest(req *CapabilityRequest, model string) (*protocols.Request, error) {
	req.Options["model"] = model

	return &protocols.Request{
		Messages: req.Messages,
		Options:  req.Options,
	}, nil
}

// ParseResponse parses a tools response.
func (t *tools) ParseResponse(body []byte) (any, error) {
	var response protocols.ToolsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse tools response: %w", err)
	}
	return response, nil
}

// SupportsStreaming returns whether this capability supports streaming.
func (t *tools) SupportsStreaming() bool {
	return true
}

// CreateStreamingRequest creates a streaming tools request.
func (t *tools) CreateStreamingRequest(req *CapabilityRequest, model string) (*protocols.Request, error) {
	req.Options["model"] = model
	req.Options["stream"] = true

	return &protocols.Request{
		Messages: req.Messages,
		Options:  req.Options,
	}, nil
}

// ParseStreamingChunk parses a streaming chunk.
func (t *tools) ParseStreamingChunk(line []byte) (any, error) {
	var chunk protocols.StreamingChunk
	if err := json.Unmarshal(line, &chunk); err != nil {
		return nil, fmt.Errorf("failed to parse streaming chunk: %w", err)
	}
	return &chunk, nil
}

// IsStreamComplete checks if streaming is complete.
func (t *tools) IsStreamComplete(chunk any) bool {
	if sc, ok := chunk.(*protocols.StreamingChunk); ok {
		if len(sc.Choices) > 0 {
			// FinishReason is *string, check for non-nil and non-empty
			return sc.Choices[0].FinishReason != nil && *sc.Choices[0].FinishReason != ""
		}
	}
	return false
}
```

**Changes**:
- Same pattern as chat.go and vision.go
- `Protocol()` returns `protocols.Tools`
- Factory function exported: `NewTools()`
- Parses `ToolsResponse` (not ChatResponse)
- Return pointer to `protocols.Request`
- Correct pointer handling in `IsStreamComplete`

### Step 1.8: Update Embeddings Capability

**File**: `pkg/capabilities/embeddings.go`

**Replace Entire File**:
```go
package capabilities

import (
	"encoding/json"
	"fmt"

	"github.com/JaimeStill/go-agents/pkg/protocols"
)

// embeddings implements the Capability interface for embeddings generation.
type embeddings struct{}

// NewEmbeddings creates a new embeddings capability.
// Exported for direct instantiation by model package.
func NewEmbeddings() Capability {
	return &embeddings{}
}

// Protocol returns the protocol this capability implements.
func (e *embeddings) Protocol() protocols.Protocol {
	return protocols.Embeddings
}

// CreateRequest creates an embeddings request from a capability request.
func (e *embeddings) CreateRequest(req *CapabilityRequest, model string) (*protocols.Request, error) {
	req.Options["model"] = model

	return &protocols.Request{
		Messages: req.Messages,
		Options:  req.Options,
	}, nil
}

// ParseResponse parses an embeddings response.
func (e *embeddings) ParseResponse(body []byte) (any, error) {
	var response protocols.EmbeddingsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse embeddings response: %w", err)
	}
	return response, nil
}

// SupportsStreaming returns whether this capability supports streaming.
func (e *embeddings) SupportsStreaming() bool {
	return false
}
```

**Changes**:
- Non-streaming capability (simpler than others)
- `Protocol()` returns `protocols.Embeddings`
- Factory function exported: `NewEmbeddings()`
- Parses `EmbeddingsResponse`
- Return pointer to `protocols.Request`
- No streaming methods needed

### Step 1.9: Remove Registry Infrastructure

**Goal**: Eliminate the registry pattern since we no longer need multiple capability formats per protocol.

**Delete Files**:
- `pkg/capabilities/registry.go` - Remove entirely
- `pkg/capabilities/init.go` - Remove entirely

**Rationale**: With option validation removed, "chat" and "o-chat" are functionally identical. We only need one capability implementation per protocol.

### Step 1.10: Update Capability Instantiation

**File**: `pkg/models/model.go`

**Find**: `Model.GetCapability(protocol protocols.Protocol)` method

**Replace capability lookup logic**:

**Old approach**:
```go
// Used registry to get capability by format name
capability, err := capabilities.GetFormat(handler.format)
```

**New approach**:
```go
// Direct instantiation based on protocol
func (m *model) GetCapability(protocol protocols.Protocol) (capabilities.Capability, error) {
	switch protocol {
	case protocols.Chat:
		return capabilities.NewChat(), nil
	case protocols.Vision:
		return capabilities.NewVision(), nil
	case protocols.Tools:
		return capabilities.NewTools(), nil
	case protocols.Embeddings:
		return capabilities.NewEmbeddings(), nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}
```

**Note**: Change factory function names to be exported (capitalized):
- `newChat()` → `NewChat()`
- `newVision()` → `NewVision()`
- `newTools()` → `NewTools()`
- `newEmbeddings()` → `NewEmbeddings()`

### Step 1.11: Update Model Integration

**File**: `pkg/models/model.go`

**Current ProtocolHandler** (approximate):
```go
type ProtocolHandler struct {
	capability capabilities.Capability
	options    map[string]any
}
```

**No structural change needed**, but remove any validation logic in model methods that call `capability.ValidateOptions()` or `capability.ProcessOptions()`.

**Find and Remove**:
- Any calls to `capability.ValidateOptions()`
- Any calls to `capability.ProcessOptions()`
- Any calls to `capability.Options()` for validation purposes

Options should flow directly from config → model → capability without validation.

### Step 1.12: Update Transport Client

**File**: `pkg/transport/client.go`

**Current ExecuteProtocol** (lines 108-130):
```go
func (c *client) ExecuteProtocol(ctx context.Context, req *capabilities.CapabilityRequest) (any, error) {
	capability, err := c.model.GetCapability(req.Protocol)
	if err != nil {
		return nil, fmt.Errorf("capability selection failed: %w", err)
	}

	options := c.model.MergeRequestOptions(req.Protocol, req.Options)

	if err := capability.ValidateOptions(options); err != nil {
		return nil, fmt.Errorf("invalid options for %s protocol: %w", req.Protocol, err)
	}

	request := &capabilities.CapabilityRequest{
		Protocol: req.Protocol,
		Messages: req.Messages,
		Options:  options,
	}

	return c.execute(ctx, capability, request)
}
```

**Replace with**:
```go
func (c *client) ExecuteProtocol(ctx context.Context, req *capabilities.CapabilityRequest) (any, error) {
	capability, err := c.model.GetCapability(req.Protocol)
	if err != nil {
		return nil, fmt.Errorf("capability selection failed: %w", err)
	}

	options := c.model.MergeRequestOptions(req.Protocol, req.Options)

	request := &capabilities.CapabilityRequest{
		Protocol: req.Protocol,
		Messages: req.Messages,
		Options:  options,
	}

	return c.execute(ctx, capability, request)
}
```

**Changes**: Remove `ValidateOptions()` call.

**Apply same change to ExecuteProtocolStream** (lines 132-160): Remove validation call.

**Also update error message** (line 144 in ExecuteProtocolStream):
```go
// Old (referenced capability name string)
return nil, fmt.Errorf("capability %s does not support streaming", capability.Name())

// New (reference protocol from request)
return nil, fmt.Errorf("protocol %s does not support streaming", req.Protocol)
```

### Step 1.13: Verify Compilation

**Commands**:
```bash
cd /home/jaime/code/go-agents
go build ./pkg/...
go test ./tests/capabilities/... -v
```

**Expected**: All packages compile successfully. Tests may need updates if they tested validation logic.

---

## Phase 2: Implement Transport-Layer Retry

### Goal
Add intelligent retry logic to transport client with HTTP status code and network error detection.

### Step 2.1: Create Retry Configuration Types

**File**: `pkg/config/transport.go`

**Current RetryConfig fields** (lines 11-12):
```go
MaxRetries       int      `json:"max_retries"`
RetryBackoffBase Duration `json:"retry_backoff_base"`
```

**Replace with enhanced retry config**:
```go
// RetryConfig configures retry behavior for failed requests.
type RetryConfig struct {
	MaxRetries      int      `json:"max_retries"`       // Maximum number of retry attempts
	InitialBackoff  Duration `json:"initial_backoff"`   // Initial backoff duration
	MaxBackoff      Duration `json:"max_backoff"`       // Maximum backoff duration
	BackoffMultiplier float64  `json:"backoff_multiplier"` // Backoff multiplier (typically 2.0)
	Jitter          bool     `json:"jitter"`            // Add random jitter to backoff
}

// DefaultRetryConfig returns default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    Duration(time.Second),
		MaxBackoff:        Duration(30 * time.Second),
		BackoffMultiplier: 2.0,
		Jitter:            true,
	}
}
```

**Update TransportConfig**:
```go
type TransportConfig struct {
	Provider          *ProviderConfig `json:"provider"`
	Timeout           Duration        `json:"timeout"`
	Retry             RetryConfig     `json:"retry"`
	ConnectionPoolSize int            `json:"connection_pool_size"`
	ConnectionTimeout Duration        `json:"connection_timeout"`
}
```

**Changes**: Replace individual retry fields with nested `RetryConfig`.

### Step 2.2: Create Retry Infrastructure

**New File**: `pkg/transport/retry.go`

```go
package transport

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"time"

	"github.com/JaimeStill/go-agents/pkg/config"
)

// HTTPStatusError represents an HTTP error with status code.
type HTTPStatusError struct {
	StatusCode int
	Status     string
	Body       []byte
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Status)
}

// isRetryableError determines if an error should trigger a retry attempt.
// Retries on:
// - HTTP 429 (rate limit), 502, 503, 504 (gateway errors)
// - Network errors (connection failures, timeouts)
// Does NOT retry on:
// - Context cancellation/deadline exceeded
// - Client errors (4xx except 429)
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Never retry context errors
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Check for HTTP status errors
	var httpErr *HTTPStatusError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode == 429 || // Rate limit
			httpErr.StatusCode == 502 || // Bad gateway
			httpErr.StatusCode == 503 || // Service unavailable
			httpErr.StatusCode == 504 // Gateway timeout
	}

	// Check for network operation errors
	var netOpErr *net.OpError
	if errors.As(err, &netOpErr) {
		return true // Network operations are generally retryable
	}

	// Check for DNS errors
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return dnsErr.Temporary() || dnsErr.Timeout()
	}

	// Check for URL errors (unwrap and check underlying)
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return isRetryableError(urlErr.Err)
	}

	return false
}

// calculateBackoff computes exponential backoff with optional jitter.
func calculateBackoff(attempt int, cfg config.RetryConfig) time.Duration {
	// Cap attempt to prevent overflow
	maxAttempt := min(attempt, 10)

	// Calculate exponential backoff: InitialBackoff * 2^attempt
	delay := time.Duration(cfg.InitialBackoff) * time.Duration(1<<uint(maxAttempt))

	// Apply jitter (±25% randomization) to prevent thundering herd
	if cfg.Jitter {
		jitterRange := delay / 4
		jitter := time.Duration(rand.Int63n(int64(jitterRange)*2)) - jitterRange
		delay += jitter
	}

	// Cap at MaxBackoff
	return min(delay, time.Duration(cfg.MaxBackoff))
}

// doWithRetry executes an operation with retry logic.
func doWithRetry[T any](
	ctx context.Context,
	cfg config.RetryConfig,
	operation func(context.Context) (T, error),
) (T, error) {
	var result T
	var lastErr error

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		// Check context cancellation before retry
		if err := ctx.Err(); err != nil {
			return result, fmt.Errorf("operation cancelled: %w", err)
		}

		result, lastErr = operation(ctx)
		if lastErr == nil {
			return result, nil
		}

		// Check if error is retryable
		if !isRetryableError(lastErr) {
			return result, lastErr
		}

		// Don't sleep after last attempt
		if attempt < cfg.MaxRetries {
			delay := calculateBackoff(attempt, cfg)

			select {
			case <-time.After(delay):
				// Continue to next retry
			case <-ctx.Done():
				return result, fmt.Errorf("operation cancelled during backoff: %w", ctx.Err())
			}
		}
	}

	return result, fmt.Errorf("max retries (%d) exceeded: %w", cfg.MaxRetries, lastErr)
}
```

**Modern Go 1.25.2 Patterns**:
- Use `errors.Is()` for value comparisons (context errors)
- Use `errors.As()` for type assertions (network errors)
- Use builtin `min()` for backoff calculations
- Direct `ctx.Err()` check instead of select
- Error wrapping with `%w`

### Step 2.3: Update Transport Client Execute Methods

**File**: `pkg/transport/client.go`

**Update execute() method** (lines 170-216):

**Current Implementation**:
```go
func (c *client) execute(ctx context.Context, capability capabilities.Capability, req *capabilities.CapabilityRequest) (any, error) {
	capRequest, err := capability.CreateRequest(req, c.model.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	providerRequest, err := c.provider.PrepareRequest(ctx, req.Protocol, capRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		providerRequest.URL,
		bytes.NewBuffer(providerRequest.Body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	for key, value := range providerRequest.Headers {
		httpReq.Header.Set(key, value)
	}

	c.provider.SetHeaders(httpReq)

	httpClient := c.HTTPClient()
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		c.setHealthy(false)
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	result, err := c.provider.ProcessResponse(resp, capability)
	if err != nil {
		c.setHealthy(false)
		return nil, err
	}

	c.setHealthy(true)
	return result, nil
}
```

**Replace with retry-enabled version**:
```go
func (c *client) execute(ctx context.Context, capability capabilities.Capability, req *capabilities.CapabilityRequest) (any, error) {
	return doWithRetry(ctx, c.config.Retry, func(ctx context.Context) (any, error) {
		capRequest, err := capability.CreateRequest(req, c.model.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		providerRequest, err := c.provider.PrepareRequest(ctx, req.Protocol, capRequest)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare request: %w", err)
		}

		httpReq, err := http.NewRequestWithContext(
			ctx,
			"POST",
			providerRequest.URL,
			bytes.NewBuffer(providerRequest.Body),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP request: %w", err)
		}

		for key, value := range providerRequest.Headers {
			httpReq.Header.Set(key, value)
		}

		c.provider.SetHeaders(httpReq)

		httpClient := c.HTTPClient()
		resp, err := httpClient.Do(httpReq)
		if err != nil {
			c.setHealthy(false)
			return nil, fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		// Check for retryable HTTP status codes
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			c.setHealthy(false)
			return nil, &HTTPStatusError{
				StatusCode: resp.StatusCode,
				Status:     resp.Status,
				Body:       bodyBytes,
			}
		}

		result, err := c.provider.ProcessResponse(resp, capability)
		if err != nil {
			c.setHealthy(false)
			return nil, err
		}

		c.setHealthy(true)
		return result, nil
	})
}
```

**Key Changes**:
- Wrap entire execution in `doWithRetry()`
- Check HTTP status codes and return `HTTPStatusError` for retryable codes
- Retry logic handles network errors automatically
- Health tracking remains the same

**Add import**:
```go
import (
	"io"
	// ... existing imports
)
```

### Step 2.4: Update Streaming Execute Method

**File**: `pkg/transport/client.go`

**Update executeStream()** (lines 218-277):

Streaming is more complex because we can't retry once the stream has started. The retry should only apply to establishing the connection.

**Replace with**:
```go
func (c *client) executeStream(ctx context.Context, capability capabilities.StreamingCapability, req *capabilities.CapabilityRequest) (<-chan protocols.StreamingChunk, error) {
	// Retry connection establishment only
	resp, err := doWithRetry(ctx, c.config.Retry, func(ctx context.Context) (*http.Response, error) {
		capRequest, err := capability.CreateStreamingRequest(req, c.model.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to create streaming request: %w", err)
		}

		providerRequest, err := c.provider.PrepareStreamRequest(ctx, req.Protocol, capRequest)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare streaming request: %w", err)
		}

		httpReq, err := http.NewRequestWithContext(
			ctx,
			"POST",
			providerRequest.URL,
			bytes.NewBuffer(providerRequest.Body),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP request: %w", err)
		}

		for key, value := range providerRequest.Headers {
			httpReq.Header.Set(key, value)
		}

		c.provider.SetHeaders(httpReq)

		httpClient := c.HTTPClient()
		resp, err := httpClient.Do(httpReq)
		if err != nil {
			c.setHealthy(false)
			return nil, fmt.Errorf("streaming request failed: %w", err)
		}

		// Check for retryable HTTP status codes
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			c.setHealthy(false)
			return nil, &HTTPStatusError{
				StatusCode: resp.StatusCode,
				Status:     resp.Status,
				Body:       bodyBytes,
			}
		}

		return resp, nil
	})

	if err != nil {
		return nil, err
	}

	// Connection established, now process stream (no retry at this point)
	stream, err := c.provider.ProcessStreamResponse(ctx, resp, capability)
	if err != nil {
		c.setHealthy(false)
		resp.Body.Close()
		return nil, err
	}

	output := make(chan protocols.StreamingChunk)
	go func() {
		defer close(output)
		defer resp.Body.Close()

		for data := range stream {
			if chunk, ok := data.(*protocols.StreamingChunk); ok {
				select {
				case output <- *chunk:
				case <-ctx.Done():
					return
				}
			}
		}
		c.setHealthy(true)
	}()

	return output, nil
}
```

**Key Changes**:
- Retry only connection establishment (up to receiving response)
- Once stream starts, no retry (would duplicate data)
- Check status codes before processing stream
- Use `defer close(output)` pattern

### Step 2.5: Update Configuration Defaults

**File**: `pkg/config/transport.go`

**Update DefaultTransportConfig**:
```go
func DefaultTransportConfig() *TransportConfig {
	return &TransportConfig{
		Provider:          DefaultProviderConfig(),
		Timeout:           Duration(2 * time.Minute),
		Retry:             DefaultRetryConfig(),
		ConnectionPoolSize: 10,
		ConnectionTimeout: Duration(30 * time.Second),
	}
}
```

**Update Merge method** to handle nested RetryConfig:
```go
func (c *TransportConfig) Merge(source *TransportConfig) {
	if source == nil {
		return
	}

	if source.Provider != nil {
		if c.Provider == nil {
			c.Provider = source.Provider
		} else {
			c.Provider.Merge(source.Provider)
		}
	}

	if source.Timeout > 0 {
		c.Timeout = source.Timeout
	}

	// Merge retry config
	if source.Retry.MaxRetries > 0 {
		c.Retry.MaxRetries = source.Retry.MaxRetries
	}
	if source.Retry.InitialBackoff > 0 {
		c.Retry.InitialBackoff = source.Retry.InitialBackoff
	}
	if source.Retry.MaxBackoff > 0 {
		c.Retry.MaxBackoff = source.Retry.MaxBackoff
	}
	if source.Retry.BackoffMultiplier > 0 {
		c.Retry.BackoffMultiplier = source.Retry.BackoffMultiplier
	}
	// Jitter: use source value if explicitly set
	c.Retry.Jitter = source.Retry.Jitter

	if source.ConnectionPoolSize > 0 {
		c.ConnectionPoolSize = source.ConnectionPoolSize
	}

	if source.ConnectionTimeout > 0 {
		c.ConnectionTimeout = source.ConnectionTimeout
	}
}
```

### Step 2.6: Verify Retry Functionality

**Commands**:
```bash
cd /home/jaime/code/go-agents
go build ./pkg/...
go test ./tests/transport/... -v
```

**Manual Testing** (if needed):
- Use tools/prompt-agent with intentionally bad endpoint to verify retry
- Check retry backoff timing with logging

---

## Phase 3: Simplify Configuration Structure

### Goal
Remove `Format` field from `CapabilityConfig`, making capability name the format identifier.

### Step 3.1: Update CapabilityConfig

**File**: `pkg/config/options.go`

**Current Structure**:
```go
type CapabilityConfig struct {
	Format  string         `json:"format"`
	Options map[string]any `json:"options"`
}
```

**Replace with**:
```go
type CapabilityConfig struct {
	Options map[string]any `json:"options"`
}

// DefaultCapabilityConfig returns a default capability configuration.
func DefaultCapabilityConfig() CapabilityConfig {
	return CapabilityConfig{
		Options: make(map[string]any),
	}
}
```

**Rationale**:
- The key in `Capabilities` map is the protocol name (e.g., "chat", "vision")
- With registry removed (Phase 1), protocol name directly maps to capability
- No `Format` field needed - protocol IS the capability identifier

### Step 3.2: Update ModelConfig

**File**: `pkg/config/model.go`

**Current Structure** (approximate):
```go
type ModelConfig struct {
	Name         string                      `json:"name"`
	Capabilities map[string]CapabilityConfig `json:"capabilities"`
}
```

**Keep As-Is**: No structural changes needed. The map key is the capability name.

**Update Merge method** if it references `Format` field:
```go
func (c *ModelConfig) Merge(source *ModelConfig) {
	if source == nil {
		return
	}

	if source.Name != "" {
		c.Name = source.Name
	}

	if source.Capabilities != nil {
		if c.Capabilities == nil {
			c.Capabilities = make(map[string]CapabilityConfig)
		}
		for name, capConfig := range source.Capabilities {
			existing, exists := c.Capabilities[name]
			if !exists {
				c.Capabilities[name] = capConfig
			} else {
				// Merge options
				if existing.Options == nil {
					existing.Options = make(map[string]any)
				}
				for k, v := range capConfig.Options {
					existing.Options[k] = v
				}
				c.Capabilities[name] = existing
			}
		}
	}
}
```

### Step 3.3: Remove Option Validation Helpers

**File**: `pkg/config/options.go`

**Find and remove** (if they exist):
- `ValidateRequiredOptions()` function
- `FilterSupportedOptions()` function
- Any other validation-specific helpers

**Keep**:
- `ExtractOption[T]()` - Still useful for safe extraction with defaults
- `MergeOptions()` - Still useful for merging maps

### Step 3.4: Update Configuration Examples

**Example Configuration Format**:
```json
{
  "name": "document-classifier",
  "system_prompt": "You are a classification assistant.",
  "transport": {
    "provider": {
      "name": "openai",
      "base_url": "https://api.openai.com/v1",
      "model": {
        "name": "gpt-4o",
        "capabilities": {
          "chat": {
            "options": {
              "max_tokens": 1200,
              "temperature": 0.7
            }
          },
          "vision": {
            "options": {
              "max_tokens": 2000
            }
          }
        }
      }
    },
    "timeout": "2m",
    "retry": {
      "max_retries": 3,
      "initial_backoff": "1s",
      "max_backoff": "30s",
      "backoff_multiplier": 2.0,
      "jitter": true
    }
  }
}
```

**For empty options**:
```json
{
  "capabilities": {
    "chat": {},
    "vision": {}
  }
}
```

### Step 3.5: Verify Configuration Loading

**Commands**:
```bash
cd /home/jaime/code/go-agents
go build ./pkg/...
go test ./tests/config/... -v
```

**Update example configs** in `tools/prompt-agent/` and any test fixtures.

---

## Phase 4: Update Model Package Integration

### Goal
Ensure model package properly integrates simplified capabilities and configuration.

### Step 4.1: Update Model Interface (if needed)

**File**: `pkg/models/model.go`

**Review interface methods**:
- Remove any methods that exposed capability options or validation
- Keep `GetCapability()`, `MergeRequestOptions()`, `Name()`

**ProtocolHandler** should only hold capability and merged options:
```go
type ProtocolHandler struct {
	capability capabilities.Capability
	options    map[string]any
}
```

### Step 4.2: Update MergeRequestOptions

**File**: `pkg/models/model.go`

**Ensure this method**:
1. Gets default options from model config for the protocol
2. Merges request options (request takes precedence)
3. Returns merged map
4. Does NOT validate

**Example Implementation**:
```go
func (m *model) MergeRequestOptions(protocol protocols.Protocol, requestOptions map[string]any) map[string]any {
	merged := make(map[string]any)

	// Start with model defaults for this protocol
	if handler, exists := m.protocols[protocol]; exists {
		for k, v := range handler.options {
			merged[k] = v
		}
	}

	// Override with request options
	for k, v := range requestOptions {
		merged[k] = v
	}

	return merged
}
```

### Step 4.3: Verify Model Tests

**Commands**:
```bash
go test ./tests/models/... -v
```

**Update tests** to remove validation expectations.

---

## Phase 5: Update Documentation

### Goal
Document the architectural changes for users and future development.

### Step 5.1: Update ARCHITECTURE.md

**File**: `ARCHITECTURE.md`

**Update Capabilities Section**:
- Document that capabilities only handle transformation
- Explain that options flow through without validation
- Note that providers may validate if needed
- Update capability interface documentation

**Update Transport Section**:
- Document retry infrastructure
- Explain retryable error detection (HTTP status + network)
- Document retry configuration options
- Note that streaming only retries connection establishment

**Update Configuration Section**:
- Document simplified configuration structure
- Show that capability name IS the format
- Explain option merging flow without validation
- Provide configuration examples

### Step 5.2: Update PROJECT.md

**File**: `PROJECT.md`

**Add to "Recent Changes" or create new section**:
- Note completion of architecture simplification
- List removed features (capability validation)
- List added features (transport retry)
- Update roadmap to reflect new architecture

### Step 5.3: Update README.md Examples

**File**: `README.md`

**Update configuration examples** to use new format:
```go
config := &config.AgentConfig{
    SystemPrompt: "You are a helpful assistant.",
    Transport: &config.TransportConfig{
        Provider: &config.ProviderConfig{
            Name:    "openai",
            BaseURL: "https://api.openai.com/v1",
            Model: &config.ModelConfig{
                Name: "gpt-4o",
                Capabilities: map[string]config.CapabilityConfig{
                    "chat": {
                        Options: map[string]any{
                            "max_tokens": 1200,
                        },
                    },
                },
            },
        },
        Retry: config.DefaultRetryConfig(),
    },
}
```

**Update usage examples** to reflect new patterns.

### Step 5.4: Create Migration Guide

**New File**: `_context/migration-v0.2.md`

Document breaking changes for users upgrading from v0.1.x:

1. **Capability Configuration Changes**:
   - Remove `format` field from capability config
   - Capability name in map is the format

2. **Option Validation Removed**:
   - Library no longer validates options
   - Invalid options will be caught by provider APIs
   - Update code if it relied on validation errors

3. **Retry Configuration Changes**:
   - New retry fields replace old ones
   - `RetryBackoffBase` → `InitialBackoff`
   - New fields: `MaxBackoff`, `BackoffMultiplier`, `Jitter`

4. **Code Changes Required**:
   - Update configuration JSON files
   - Remove any code that called `capability.ValidateOptions()`
   - Update retry configuration if customized

---

## Phase 6: Update Tests

### Goal
Ensure test suite covers new behavior and doesn't test removed features.

### Step 6.1: Update Capability Tests

**Directory**: `tests/capabilities/`

**Remove**:
- Tests for `ValidateOptions()` - no longer exists
- Tests for `ProcessOptions()` - no longer exists
- Tests for `Options()` - no longer exists

**Keep/Update**:
- Tests for `CreateRequest()` - verify options pass through
- Tests for `ParseResponse()` - unchanged
- Tests for streaming - verify stream flag added

**Add**:
- Test that invalid options don't cause errors at capability layer
- Test that options are copied, not mutated

### Step 6.2: Create Retry Tests

**New File**: `tests/transport/retry_test.go`

```go
package transport_test

import (
	"context"
	"errors"
	"net"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/transport"
)

func TestIsRetryableError_HTTPStatus(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		retryable  bool
	}{
		{"429 Rate Limit", 429, true},
		{"502 Bad Gateway", 502, true},
		{"503 Service Unavailable", 503, true},
		{"504 Gateway Timeout", 504, true},
		{"400 Bad Request", 400, false},
		{"401 Unauthorized", 401, false},
		{"404 Not Found", 404, false},
		{"500 Internal Server", 500, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &transport.HTTPStatusError{
				StatusCode: tt.statusCode,
				Status:     "test",
			}

			// This requires making isRetryableError exported or using reflection
			// For now, test via actual retry behavior
			// TODO: Export isRetryableError for testing or use integration test
		})
	}
}

func TestRetry_NetworkError(t *testing.T) {
	// Test that network errors trigger retry
	// Use httptest server that closes connection
}

func TestRetry_MaxAttempts(t *testing.T) {
	// Test that retry stops after max attempts
}

func TestRetry_ContextCancellation(t *testing.T) {
	// Test that context cancellation stops retry
}

func TestRetry_BackoffCalculation(t *testing.T) {
	// Test exponential backoff calculation
}
```

**Note**: Some retry functions may need to be exported for testing, or use integration tests.

### Step 6.3: Update Config Tests

**File**: `tests/config/options_test.go`

**Remove**:
- Tests for validation functions that were deleted

**Update**:
- Tests for `ExtractOption` - should still work
- Tests for `MergeOptions` - should still work

**Add**:
- Test for new retry config structure
- Test for capability config without format field

### Step 6.4: Update Integration Tests

**Directory**: `tests/integration/` (if exists)

**Update**:
- Tests should work with new config format
- Tests may see retry behavior (add assertions if needed)
- Update mock servers to test retry (return 503, then 200)

### Step 6.5: Verify Test Coverage

**Commands**:
```bash
cd /home/jaime/code/go-agents

# Test all packages
go test ./tests/... -v -cover

# Generate coverage report
go test ./tests/... -coverprofile=coverage.out -coverpkg=./pkg/...
go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Target: 80% overall, 100% for critical paths
```

---

## Phase 7: Prepare for Library Extraction

### Goal
Identify and document components ready for extraction to separate libraries.

### Step 7.1: Document Document Context Components

**Create**: `_context/document-context-library.md`

Document components to extract to go-agents-document-context:

**From classify-docs/pkg/document/**:
- `Document` interface
- `Page` interface
- `ImageOptions` struct
- PDF implementation using pdfcpu
- Future: Word, image, text implementations

**From classify-docs/pkg/encoding/**:
- `EncodeImageDataURI()` function
- Format support (PNG, JPEG)

**From classify-docs/pkg/processing/**:
- `ContextProcessor[TContext]` type
- `ProcessWithContext()` function
- `SequentialResult[TContext]` type
- `ProgressFunc[TContext]` type
- Sequential processing configuration

**Package Structure**:
```
go-agents-document-context/
├── pkg/
│   ├── document/
│   │   ├── document.go      # Interfaces
│   │   ├── pdf.go           # PDF implementation
│   │   └── options.go       # ImageOptions
│   ├── encoding/
│   │   └── datauri.go       # Base64 data URI encoding
│   └── processing/
│       └── sequential.go     # Generic sequential processor
├── go.mod
├── README.md
└── LICENSE
```

**Dependencies**:
- pdfcpu for PDF support
- No dependency on go-agents (standalone)

### Step 7.2: Document Orchestration Integration

**Create**: `_context/orchestration-integration.md`

Document where classify-docs patterns fit in go-agents-orchestration:

**Sequential Processing Pattern**:
- Could live in document-context (document-specific)
- Could live in orchestration (generic workflow)
- **Decision Point**: Start in document-context, extract to orchestration if proven generic

**Retry Pattern**:
- Now lives in go-agents transport layer
- Orchestration can use same pattern for workflow retry
- Consider extracting as shared retry package if orchestration needs different retry semantics

**Hub Integration**:
- Sequential chain workflow uses hub.Request() for agent communication
- Each step in chain is an agent handler
- Context accumulation maps to workflow state

### Step 7.3: Document Classify-Docs Standalone Project

**Create**: `_context/classify-docs-project.md`

Document structure for standalone classify-docs project:

**Project Structure**:
```
classify-docs/
├── cmd/classify-docs/       # CLI entry point
├── pkg/
│   ├── classifier/          # Classification orchestration
│   ├── prompts/             # Prompt generation (template-based)
│   └── cache/               # System prompt caching
├── configs/                 # Example configurations
├── tests/                   # Integration tests
├── go.mod
├── README.md
└── LICENSE
```

**Dependencies**:
- go-agents (v0.2.0+)
- go-agents-document-context
- go-agents-orchestration (future: if using workflow patterns)

**Component Mapping**:
- `pkg/classify/` → `pkg/classifier/` (renamed for clarity)
- `pkg/prompt/` → `pkg/prompts/` (templates may stay tool-specific)
- `pkg/cache/` → Keep as-is (tool-specific)
- `pkg/config/` → Minimal tool-specific config (most in go-agents)
- `pkg/retry/` → REMOVED (now in go-agents transport)
- `pkg/document/` → MOVED to go-agents-document-context
- `pkg/encoding/` → MOVED to go-agents-document-context
- `pkg/processing/` → MOVED to go-agents-document-context

---

## Phase 8: Final Verification

### Step 8.1: Build All Packages

**Commands**:
```bash
cd /home/jaime/code/go-agents
go mod tidy
go build ./...
go build ./cmd/...
go build ./tools/...
```

**Expected**: All packages compile without errors.

### Step 8.2: Run Full Test Suite

**Commands**:
```bash
go test ./tests/... -v -cover
```

**Expected**: All tests pass with acceptable coverage.

### Step 8.3: Manual Integration Testing

**Test classify-docs prototype**:
```bash
cd /home/jaime/code/go-agents/tools/classify-docs

# Update config to new format
# Run classification
go run cmd/classify-docs/main.go -config config.classify-gpt4o-key.json

# Verify retry behavior (intentionally use bad endpoint first)
# Verify classification still works with simplified architecture
```

### Step 8.4: Update CHANGELOG

**File**: `CHANGELOG.md`

**Add new version entry**:
```markdown
## [v0.2.0] - 2025-10-29

**Changed**:
- Simplified capability system: removed option validation and processing
- Capabilities now only handle request/response transformation
- Configuration structure simplified: capability name is the format identifier
- Retry configuration structure updated with enhanced backoff options

**Added**:
- Transport-layer retry infrastructure with intelligent error detection
- Retry on HTTP 429, 502, 503, 504 status codes
- Retry on network errors (connection failures, timeouts)
- Exponential backoff with jitter support
- `pkg/transport/retry.go` - Retry infrastructure
- `HTTPStatusError` type for HTTP error handling

**Removed**:
- `Capability.ValidateOptions()` method
- `Capability.ProcessOptions()` method
- `Capability.Options()` method
- `StandardCapability` base implementation
- `StandardStreamingCapability` base implementation
- `CapabilityConfig.Format` field
- Option validation helper functions

**Fixed**:
- Retry configuration fields now properly utilized (previously unused)

**Breaking Changes**:
- Configuration format changed: remove `format` field from capability config
- Capability interface methods removed: `ValidateOptions`, `ProcessOptions`, `Options`
- Retry configuration fields renamed: `RetryBackoffBase` → `InitialBackoff`
- Code that relied on capability validation must be updated
```

---

## Success Criteria

### Architecture
- [x] Capabilities simplified to transformation only
- [x] Transport layer has retry infrastructure
- [x] Configuration structure flattened
- [x] Modern Go 1.25.2 patterns used throughout

### Compilation
- [ ] All packages compile without errors
- [ ] All tests pass
- [ ] Coverage ≥ 80% overall, 100% for critical paths

### Documentation
- [ ] ARCHITECTURE.md updated
- [ ] PROJECT.md updated
- [ ] README.md examples updated
- [ ] Migration guide created
- [ ] CHANGELOG.md updated
- [ ] Component extraction documented

### Testing
- [ ] Capability tests updated
- [ ] Retry tests created
- [ ] Config tests updated
- [ ] Integration tests pass
- [ ] Manual classify-docs verification

### Preparation for Extraction
- [ ] Document context components identified
- [ ] Orchestration integration documented
- [ ] Classify-docs project structure planned

---

## Next Steps After Completion

1. **Create go-agents-document-context repository**
   - Extract document, encoding, processing packages
   - Create standalone module
   - Release v0.1.0

2. **Create standalone classify-docs repository**
   - Migrate classification logic
   - Update to use new library dependencies
   - Establish as production project

3. **Optimization iteration**
   - Baseline performance testing with gpt-5-mini
   - Identify failure patterns (e.g., marked-documents_19.pdf)
   - Experiment with classification strategies
   - Tune confidence thresholds

4. **Integrate with go-agents-orchestration**
   - Implement sequential chain patterns
   - Use hub for agent coordination
   - Explore parallel processing patterns

---

## Troubleshooting

### Compilation Errors

**"undefined: StandardCapability"**
- Ensure all capability files updated to remove embedding
- Search for `StandardCapability` references and remove

**"cannot use capability (type Capability) as type ... missing method Options"**
- Remove calls to `capability.Options()`
- Remove validation logic that used options list

### Test Failures

**Validation tests fail**
- Remove tests for deleted validation methods
- Update tests to check pass-through behavior instead

**Retry tests fail**
- Ensure `isRetryableError` exported or use integration tests
- Check mock server setup for retry scenarios

### Runtime Issues

**Retry not triggering**
- Verify retry config loaded properly
- Check that errors are being wrapped correctly
- Ensure `HTTPStatusError` returned for retryable status codes

**Invalid options not caught**
- Expected: Options validation removed from library
- Providers will return errors for invalid options
- Update application code to handle provider errors

---

## References

- **Go 1.25.2 Documentation**: Context7 MCP analysis
- **Current Architecture**: `ARCHITECTURE.md`
- **Project Roadmap**: `PROJECT.md`
- **Design Principles**: `CLAUDE.md`
- **Classify-Docs Prototype**: `tools/classify-docs/`
- **Go-Agents-Orchestration**: `~/code/go-agents-orchestration/`