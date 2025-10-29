# Consolidated Architecture: Protocol-Centric Simplification

## Overview

This implementation guide consolidates the go-agents architecture from 5-6 layers down to 4 essential layers by:

1. **Removing `pkg/capabilities/`** - Redundant after validation removal; logic moved to types and client
2. **Renaming `pkg/protocols/` → `pkg/types/`** - Clarifies purpose as shared type library
3. **Simplifying `pkg/models/`** - Becomes a simple config struct, not a separate package
4. **Renaming `pkg/transport/` → `pkg/client/`** - Better semantic naming
5. **Centralizing provider differentiation** - Providers handle all format-specific logic
6. **Adding transport-layer retry** - Intelligent retry with error detection

## Target Architecture

### Current State (5-6 layers)
```
Agent → Transport → Model → Capability → Protocol → Provider → HTTP
   ↓        ↓         ↓          ↓           ↓          ↓
Config   Retry?   Handler   Registry   SharedTypes  Transform
```

### Target State (4 layers)
```
Agent → Client → Provider → HTTP
   ↓       ↓         ↓
Config  Model   Transform
         ↓         ↓
      Retry    Types.Parse
```

**Eliminated Layers**:
- Capability (redundant passthrough)
- Protocol (renamed to Types - just shared types)
- Model as separate package (becomes config struct)

---

## Phase 1: Rename Protocols → Types

### Goal
Clarify that this package is a shared type library, not protocol logic.

### Step 1.1: Rename Package Directory

**Commands**:
```bash
cd /home/jaime/code/go-agents
mv pkg/protocols pkg/types
```

### Step 1.2: Update Package Declaration

**File**: `pkg/types/*.go` (all files)

**Find and replace** in all files:
```go
// Old
package protocols

// New
package types
```

### Step 1.3: Update All Imports

**Search and replace** across entire codebase:
```bash
# Find all imports
grep -r "github.com/JaimeStill/go-agents/pkg/protocols" pkg/
grep -r "github.com/JaimeStill/go-agents/pkg/protocols" tools/

# Replace with
github.com/JaimeStill/go-agents/pkg/types
```

**Common files**:
- `pkg/agent/*.go`
- `pkg/transport/*.go`
- `pkg/capabilities/*.go`
- `pkg/models/*.go`
- `pkg/providers/*.go`
- `tools/prompt-agent/*.go`

### Step 1.4: Verify Compilation

**Commands**:
```bash
go build ./pkg/...
```

**Expected**: All packages compile with new import paths.

---

## Phase 2: Move Response Parsing to Types

### Goal
Consolidate response parsing logic from capabilities into the types package where response structs are defined.

### Step 2.1: Add Parsing Functions to Types

**File**: `pkg/types/chat.go`

**Add after ChatResponse definition**:
```go
// ParseChatResponse parses a chat completion response.
func ParseChatResponse(body []byte) (*ChatResponse, error) {
	var response ChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse chat response: %w", err)
	}
	return &response, nil
}

// ParseChatStreamChunk parses a streaming chat chunk.
func ParseChatStreamChunk(data []byte) (*StreamingChunk, error) {
	var chunk StreamingChunk
	if err := json.Unmarshal(data, &chunk); err != nil {
		return nil, fmt.Errorf("failed to parse streaming chunk: %w", err)
	}
	return &chunk, nil
}
```

### Step 2.2: Add Parsing Functions for Other Protocols

**File**: `pkg/types/tools.go`

```go
// ParseToolsResponse parses a tools/function calling response.
func ParseToolsResponse(body []byte) (*ToolsResponse, error) {
	var response ToolsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse tools response: %w", err)
	}
	return &response, nil
}

// ParseToolsStreamChunk parses a streaming tools chunk.
func ParseToolsStreamChunk(data []byte) (*StreamingChunk, error) {
	var chunk StreamingChunk
	if err := json.Unmarshal(data, &chunk); err != nil {
		return nil, fmt.Errorf("failed to parse streaming chunk: %w", err)
	}
	return &chunk, nil
}
```

**File**: `pkg/types/embeddings.go`

```go
// ParseEmbeddingsResponse parses an embeddings response.
func ParseEmbeddingsResponse(body []byte) (*EmbeddingsResponse, error) {
	var response EmbeddingsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse embeddings response: %w", err)
	}
	return &response, nil
}
```

**File**: `pkg/types/vision.go` (if separate file, otherwise add to chat.go)

```go
// Vision responses use ChatResponse structure

// ParseVisionResponse parses a vision completion response.
func ParseVisionResponse(body []byte) (*ChatResponse, error) {
	return ParseChatResponse(body) // Vision uses same response format as chat
}

// ParseVisionStreamChunk parses a streaming vision chunk.
func ParseVisionStreamChunk(data []byte) (*StreamingChunk, error) {
	return ParseChatStreamChunk(data) // Vision uses same streaming format
}
```

### Step 2.3: Add Protocol-Aware Parsing Helper

**File**: `pkg/types/protocol.go`

**Add function**:
```go
// ParseResponse parses a response based on protocol type.
func ParseResponse(protocol Protocol, body []byte) (any, error) {
	switch protocol {
	case Chat:
		return ParseChatResponse(body)
	case Vision:
		return ParseVisionResponse(body)
	case Tools:
		return ParseToolsResponse(body)
	case Embeddings:
		return ParseEmbeddingsResponse(body)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

// ParseStreamChunk parses a streaming chunk based on protocol type.
func ParseStreamChunk(protocol Protocol, data []byte) (*StreamingChunk, error) {
	switch protocol {
	case Chat:
		return ParseChatStreamChunk(data)
	case Vision:
		return ParseVisionStreamChunk(data)
	case Tools:
		return ParseToolsStreamChunk(data)
	default:
		return nil, fmt.Errorf("protocol %s does not support streaming", protocol)
	}
}
```

### Step 2.4: Verify Compilation

**Commands**:
```bash
go build ./pkg/types/...
```

---

## Phase 3: Remove Capabilities Package

### Goal
Eliminate the capabilities layer entirely since it's now redundant.

### Step 3.1: Delete Capability Files

**Delete**:
- `pkg/capabilities/capability.go`
- `pkg/capabilities/chat.go`
- `pkg/capabilities/vision.go`
- `pkg/capabilities/tools.go`
- `pkg/capabilities/embeddings.go`
- `pkg/capabilities/registry.go` (if exists)
- `pkg/capabilities/init.go` (if exists)
- `pkg/capabilities/doc.go`

**Commands**:
```bash
cd /home/jaime/code/go-agents
rm -rf pkg/capabilities/
```

### Step 3.2: Remove Capabilities Tests

**Delete**:
```bash
rm -rf tests/capabilities/
```

### Step 3.3: Remove Capabilities Imports

**Search for**:
```bash
grep -r "github.com/JaimeStill/go-agents/pkg/capabilities" pkg/
```

**Files that will need updating**:
- `pkg/transport/client.go`
- `pkg/models/model.go`
- `pkg/providers/*.go`

We'll fix these in the next phases.

---

## Phase 4: Simplify Model Package

### Goal
Convert Model from a complex package with handlers to a simple configuration struct.

### Step 4.1: Create Model Domain Type

**New File**: `pkg/types/model.go`

```go
package types

// Model represents a configured LLM model at runtime.
type Model struct {
	// Name is the model identifier (e.g., "gpt-4o", "claude-3-opus")
	Name string

	// Options holds protocol-specific default options.
	Options map[Protocol]map[string]any
}

// NewModel creates a new Model with the given name.
func NewModel(name string) *Model {
	return &Model{
		Name:    name,
		Options: make(map[Protocol]map[string]any),
	}
}

// MergeOptions merges protocol-specific default options with request options.
// Request options take precedence over defaults.
func (m *Model) MergeOptions(protocol Protocol, requestOptions map[string]any) map[string]any {
	merged := make(map[string]any)

	// Start with model defaults for this protocol
	if defaults, exists := m.Options[protocol]; exists {
		for k, v := range defaults {
			merged[k] = v
		}
	}

	// Override with request options
	for k, v := range requestOptions {
		merged[k] = v
	}

	return merged
}

// SetProtocolOptions sets default options for a specific protocol.
func (m *Model) SetProtocolOptions(protocol Protocol, options map[string]any) {
	if m.Options == nil {
		m.Options = make(map[Protocol]map[string]any)
	}
	m.Options[protocol] = options
}
```

### Step 4.2: Keep ModelConfig in Config Package

**File**: `pkg/config/model.go`

**Keep as pure configuration structure**:
```go
package config

// ModelConfig is the JSON configuration structure for a model.
type ModelConfig struct {
	Name         string                      `json:"name"`
	Capabilities map[string]CapabilityConfig `json:"capabilities"`
}

// DefaultModelConfig returns a default model configuration.
func DefaultModelConfig() *ModelConfig {
	return &ModelConfig{
		Name:         "",
		Capabilities: make(map[string]CapabilityConfig),
	}
}

// Merge merges source configuration into this configuration.
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

### Step 4.3: Add Model Conversion Helper

**File**: `pkg/types/model.go`

**Add conversion function**:
```go
// FromConfig creates a Model from a ModelConfig.
// Handles conversion from string-keyed config to Protocol-keyed runtime model.
func FromConfig(cfg *config.ModelConfig) *Model {
	model := &Model{
		Name:    cfg.Name,
		Options: make(map[Protocol]map[string]any),
	}

	// Convert string keys to Protocol constants
	for protocolName, capConfig := range cfg.Capabilities {
		protocol := Protocol(protocolName) // e.g., "chat" → types.Chat
		model.Options[protocol] = capConfig.Options
	}

	return model
}
```

**Note**: This requires importing `pkg/config` in `pkg/types`, which is acceptable since types is consuming config (lower-level package).

**Add import**:
```go
import "github.com/JaimeStill/go-agents/pkg/config"
```

### Step 4.4: Delete Old Model Package

**Delete**:
```bash
rm -rf pkg/models/
rm -rf tests/models/
```

### Step 4.5: Update Config to Use New Model

**File**: `pkg/config/provider.go`

**Update ProviderConfig**:
```go
import "github.com/JaimeStill/go-agents/pkg/types"

type ProviderConfig struct {
	Name    string         `json:"name"`
	BaseURL string         `json:"base_url"`
	Model   *ModelConfig   `json:"model"`  // Keep as config for JSON loading
	Options map[string]any `json:"options,omitempty"`
}

// GetModel creates a Model from the configuration.
func (c *ProviderConfig) GetModel() *types.Model {
	return types.FromConfig(c.Model)
}
```

---

## Phase 5: Rename Transport → Client

### Goal
Rename transport package to client for better semantic clarity.

### Step 5.1: Rename Package Directory

**Commands**:
```bash
cd /home/jaime/code/go-agents
mv pkg/transport pkg/client
```

### Step 5.2: Update Package Declaration

**File**: `pkg/client/*.go`

```go
// Old
package transport

// New
package client
```

### Step 5.3: Update All Imports

**Search and replace**:
```bash
grep -r "github.com/JaimeStill/go-agents/pkg/transport" pkg/
grep -r "github.com/JaimeStill/go-agents/pkg/transport" tools/
```

**Replace with**:
```
github.com/JaimeStill/go-agents/pkg/client
```

### Step 5.4: Rename TransportConfig → ClientConfig

**File**: `pkg/config/transport.go` → `pkg/config/client.go`

**Rename type**:
```go
// ClientConfig configures the HTTP client for LLM requests.
type ClientConfig struct {
	Provider          *ProviderConfig `json:"provider"`
	Timeout           Duration        `json:"timeout"`
	Retry             RetryConfig     `json:"retry"`
	ConnectionPoolSize int            `json:"connection_pool_size"`
	ConnectionTimeout Duration        `json:"connection_timeout"`
}

// DefaultClientConfig returns default client configuration.
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		Provider:          DefaultProviderConfig(),
		Timeout:           Duration(2 * time.Minute),
		Retry:             DefaultRetryConfig(),
		ConnectionPoolSize: 10,
		ConnectionTimeout: Duration(30 * time.Second),
	}
}
```

### Step 5.5: Update AgentConfig

**File**: `pkg/config/agent.go`

```go
type AgentConfig struct {
	SystemPrompt string        `json:"system_prompt"`
	Client       *ClientConfig `json:"client"`  // Was Transport
}

// DefaultAgentConfig returns default agent configuration.
func DefaultAgentConfig() *AgentConfig {
	return &AgentConfig{
		SystemPrompt: "",
		Client:       DefaultClientConfig(),
	}
}
```

### Step 5.6: Verify Compilation

```bash
go build ./pkg/...
```

---

## Phase 6: Implement Simplified Client

### Goal
Reimplement client to work without capabilities, using types directly.

### Step 6.1: Update Client Interface

**File**: `pkg/client/client.go`

**Replace interface**:
```go
package client

import (
	"context"
	"net/http"

	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/providers"
	"github.com/JaimeStill/go-agents/pkg/types"
)

// Client provides the interface for executing LLM protocol requests.
type Client interface {
	// Provider returns the provider instance managed by this client.
	Provider() providers.Provider

	// Model returns the model configuration.
	Model() *config.Model

	// ExecuteProtocol executes a protocol request and returns the parsed response.
	ExecuteProtocol(ctx context.Context, req *types.ProtocolRequest) (any, error)

	// ExecuteProtocolStream executes a streaming protocol request.
	ExecuteProtocolStream(ctx context.Context, req *types.ProtocolRequest) (<-chan *types.StreamingChunk, error)

	// IsHealthy returns the current health status of the client.
	IsHealthy() bool
}
```

### Step 6.2: Update ProtocolRequest Type

**File**: `pkg/types/protocol.go`

**Add new request type** (if doesn't exist):
```go
// ProtocolRequest represents a high-level request from agent layer.
type ProtocolRequest struct {
	Protocol Protocol
	Messages []Message
	Options  map[string]any
}
```

### Step 6.3: Implement New Client

**File**: `pkg/client/client.go`

**Implementation**:
```go
type client struct {
	provider providers.Provider
	model    *config.Model
	config   *config.ClientConfig

	mutex      sync.RWMutex
	healthy    bool
	lastHealth time.Time
}

// New creates a new client from configuration.
func New(cfg *config.ClientConfig) (Client, error) {
	provider, err := providers.Create(cfg.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	return &client{
		provider:   provider,
		model:      cfg.Provider.GetModel(),
		config:     cfg,
		healthy:    true,
		lastHealth: time.Now(),
	}, nil
}

func (c *client) Provider() providers.Provider {
	return c.provider
}

func (c *client) Model() *config.Model {
	return c.model
}

func (c *client) IsHealthy() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.healthy
}

func (c *client) setHealthy(healthy bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.healthy = healthy
	c.lastHealth = time.Now()
}

func (c *client) HTTPClient() *http.Client {
	return &http.Client{
		Timeout: c.config.Timeout.ToDuration(),
		Transport: &http.Transport{
			MaxIdleConns:        c.config.ConnectionPoolSize,
			MaxIdleConnsPerHost: c.config.ConnectionPoolSize,
			IdleConnTimeout:     c.config.ConnectionTimeout.ToDuration(),
		},
	}
}
```

### Step 6.4: Implement ExecuteProtocol Without Capabilities

**File**: `pkg/client/client.go`

```go
func (c *client) ExecuteProtocol(ctx context.Context, req *types.ProtocolRequest) (any, error) {
	// Merge model defaults with request options
	options := c.model.MergeOptions(req.Protocol, req.Options)

	// Add model name (was done by capability)
	options["model"] = c.model.Name

	// Create protocol request
	protocolReq := &types.Request{
		Messages: req.Messages,
		Options:  options,
	}

	return c.execute(ctx, req.Protocol, protocolReq)
}

func (c *client) execute(ctx context.Context, protocol types.Protocol, req *types.Request) (any, error) {
	// Prepare provider request
	providerRequest, err := c.provider.PrepareRequest(ctx, protocol, req)
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
	c.provider.SetHeaders(httpReq)

	// Execute HTTP request
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

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.setHealthy(false)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		c.setHealthy(false)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response using types package
	result, err := types.ParseResponse(protocol, bodyBytes)
	if err != nil {
		c.setHealthy(false)
		return nil, err
	}

	c.setHealthy(true)
	return result, nil
}
```

### Step 6.5: Implement ExecuteProtocolStream

**File**: `pkg/client/client.go`

```go
func (c *client) ExecuteProtocolStream(ctx context.Context, req *types.ProtocolRequest) (<-chan *types.StreamingChunk, error) {
	// Check if protocol supports streaming
	switch req.Protocol {
	case types.Chat, types.Vision, types.Tools:
		// Supported
	default:
		return nil, fmt.Errorf("protocol %s does not support streaming", req.Protocol)
	}

	// Merge options and add streaming flag
	options := c.model.MergeOptions(req.Protocol, req.Options)
	options["model"] = c.model.Name
	options["stream"] = true

	// Create protocol request
	protocolReq := &types.Request{
		Messages: req.Messages,
		Options:  options,
	}

	return c.executeStream(ctx, req.Protocol, protocolReq)
}

func (c *client) executeStream(ctx context.Context, protocol types.Protocol, req *types.Request) (<-chan *types.StreamingChunk, error) {
	// Prepare streaming request
	providerRequest, err := c.provider.PrepareStreamRequest(ctx, protocol, req)
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
	c.provider.SetHeaders(httpReq)

	// Execute HTTP request
	httpClient := c.HTTPClient()
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		c.setHealthy(false)
		return nil, fmt.Errorf("streaming request failed: %w", err)
	}

	// Check for retryable status codes
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

	// Check status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		c.setHealthy(false)
		return nil, fmt.Errorf("streaming request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Process stream using provider
	stream, err := c.provider.ProcessStreamResponse(ctx, resp, protocol)
	if err != nil {
		c.setHealthy(false)
		resp.Body.Close()
		return nil, err
	}

	// Convert provider stream to typed chunk stream
	output := make(chan *types.StreamingChunk)
	go func() {
		defer close(output)
		defer resp.Body.Close()

		for data := range stream {
			if chunk, ok := data.(*types.StreamingChunk); ok {
				select {
				case output <- chunk:
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

---

## Phase 7: Add Retry Infrastructure to Client

### Goal
Add intelligent retry with HTTP status code and network error detection.

### Step 7.1: Create Retry Configuration

**File**: `pkg/config/client.go`

**Add retry config** (if not exists):
```go
// RetryConfig configures retry behavior for failed requests.
type RetryConfig struct {
	MaxRetries        int      `json:"max_retries"`
	InitialBackoff    Duration `json:"initial_backoff"`
	MaxBackoff        Duration `json:"max_backoff"`
	BackoffMultiplier float64  `json:"backoff_multiplier"`
	Jitter            bool     `json:"jitter"`
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

### Step 7.2: Create Retry Infrastructure

**New File**: `pkg/client/retry.go`

```go
package client

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
		return true
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

	// Calculate exponential backoff
	delay := time.Duration(cfg.InitialBackoff) * time.Duration(1<<uint(maxAttempt))

	// Apply jitter (±25% randomization)
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

### Step 7.3: Wrap Client Execution with Retry

**File**: `pkg/client/client.go`

**Update execute method**:
```go
func (c *client) execute(ctx context.Context, protocol types.Protocol, req *types.Request) (any, error) {
	return doWithRetry(ctx, c.config.Retry, func(ctx context.Context) (any, error) {
		// Prepare provider request
		providerRequest, err := c.provider.PrepareRequest(ctx, protocol, req)
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
		c.provider.SetHeaders(httpReq)

		// Execute HTTP request
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

		// Read response body
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			c.setHealthy(false)
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		// Check status code
		if resp.StatusCode != http.StatusOK {
			c.setHealthy(false)
			return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
		}

		// Parse response using types package
		result, err := types.ParseResponse(protocol, bodyBytes)
		if err != nil {
			c.setHealthy(false)
			return nil, err
		}

		c.setHealthy(true)
		return result, nil
	})
}
```

**Update executeStream** to only retry connection establishment (streaming can't be retried mid-stream).

---

## Phase 8: Update Provider Interface

### Goal
Update provider interface to work with simplified architecture.

### Step 8.1: Update Provider Interface

**File**: `pkg/providers/provider.go`

**Review interface** - should already work with types.Request:
```go
type Provider interface {
	Model() *config.Model
	PrepareRequest(ctx context.Context, protocol types.Protocol, request *types.Request) (*Request, error)
	PrepareStreamRequest(ctx context.Context, protocol types.Protocol, request *types.Request) (*Request, error)
	SetHeaders(req *http.Request)
	ProcessResponse(resp *http.Response) (any, error)
	ProcessStreamResponse(ctx context.Context, resp *http.Response, protocol types.Protocol) (<-chan any, error)
}
```

### Step 8.2: Update Provider Implementations

**File**: `pkg/providers/*.go`

**Update ProcessResponse** to use types.ParseResponse:
```go
func (p *provider) ProcessResponse(resp *http.Response, protocol types.Protocol) (any, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Use types package for parsing
	return types.ParseResponse(protocol, body)
}
```

**Update ProcessStreamResponse** to use types.ParseStreamChunk:
```go
func (p *provider) ProcessStreamResponse(ctx context.Context, resp *http.Response, protocol types.Protocol) (<-chan any, error) {
	output := make(chan any)

	go func() {
		defer close(output)
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line, err := reader.ReadBytes('\n')
			if err != nil {
				return
			}

			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				continue
			}

			// Handle SSE format (Azure/OpenAI) vs raw JSON (Ollama)
			if bytes.HasPrefix(line, []byte("data: ")) {
				line = bytes.TrimPrefix(line, []byte("data: "))
			}

			if bytes.Equal(line, []byte("[DONE]")) {
				return
			}

			// Use types package for parsing
			chunk, err := types.ParseStreamChunk(protocol, line)
			if err != nil {
				continue
			}

			select {
			case output <- chunk:
			case <-ctx.Done():
				return
			}
		}
	}()

	return output, nil
}
```

---

## Phase 9: Update Agent Layer

### Goal
Update agent to work with new client interface.

### Step 9.1: Update Agent Constructor

**File**: `pkg/agent/agent.go`

```go
import (
	"github.com/JaimeStill/go-agents/pkg/client"  // Was transport
	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/types"   // Was protocols
)

type agent struct {
	systemPrompt string
	client       client.Client  // Was transport.Client
}

func New(cfg *config.AgentConfig) (Agent, error) {
	client, err := client.New(cfg.Client)  // Was cfg.Transport
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &agent{
		systemPrompt: cfg.SystemPrompt,
		client:       client,
	}, nil
}
```

### Step 9.2: Update Agent Methods

**File**: `pkg/agent/agent.go`

**Update to use types.ProtocolRequest**:
```go
func (a *agent) Chat(ctx context.Context, prompt string, opts ...map[string]any) (string, error) {
	messages := a.initMessages(prompt)

	options := make(map[string]any)
	if len(opts) > 0 && opts[0] != nil {
		options = opts[0]
	}

	request := &types.ProtocolRequest{
		Protocol: types.Chat,
		Messages: messages,
		Options:  options,
	}

	response, err := a.client.ExecuteProtocol(ctx, request)
	if err != nil {
		return "", err
	}

	chatResp, ok := response.(*types.ChatResponse)
	if !ok {
		return "", fmt.Errorf("unexpected response type: %T", response)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return chatResp.Choices[0].Message.Content, nil
}
```

**Apply same pattern** to Vision, Tools, ChatStream, VisionStream methods.

---

## Phase 10: Update Tests

### Goal
Update test suite to work with consolidated architecture.

### Step 10.1: Update Import Paths

**All test files**:
- `pkg/protocols` → `pkg/types`
- `pkg/transport` → `pkg/client`
- `pkg/capabilities` → (remove references)

### Step 10.2: Update Config Tests

**File**: `tests/config/*.go`

Update to test new Model struct and ClientConfig.

### Step 10.3: Update Client Tests

**File**: `tests/client/*.go` (was tests/transport/)

Update to test:
- Direct protocol execution without capabilities
- Retry infrastructure
- Option merging

### Step 10.4: Remove Capability Tests

**Delete**: `tests/capabilities/` (already done in Phase 3)

### Step 10.5: Add Types Parsing Tests

**New File**: `tests/types/parsing_test.go`

```go
package types_test

import (
	"testing"

	"github.com/JaimeStill/go-agents/pkg/types"
)

func TestParseChatResponse(t *testing.T) {
	jsonData := []byte(`{
		"choices": [
			{
				"message": {
					"role": "assistant",
					"content": "Hello!"
				},
				"finish_reason": "stop"
			}
		]
	}`)

	resp, err := types.ParseChatResponse(jsonData)
	if err != nil {
		t.Fatalf("ParseChatResponse failed: %v", err)
	}

	if len(resp.Choices) != 1 {
		t.Errorf("expected 1 choice, got %d", len(resp.Choices))
	}

	if resp.Choices[0].Message.Content != "Hello!" {
		t.Errorf("unexpected content: %s", resp.Choices[0].Message.Content)
	}
}
```

---

## Phase 11: Update Documentation

### Goal
Document the new consolidated architecture.

### Step 11.1: Update ARCHITECTURE.md

**File**: `ARCHITECTURE.md`

**Major updates**:
- Document 4-layer architecture (Agent → Client → Provider → HTTP)
- Explain removal of capabilities layer
- Document types package as shared type library
- Update package dependency diagram
- Document where provider differentiation happens

### Step 11.2: Update PROJECT.md

**File**: `PROJECT.md`

**Updates**:
- Note completion of architecture consolidation
- Update MVP status
- Document simplified package structure

### Step 11.3: Update README.md

**File**: `README.md`

**Updates**:
- Update package imports in examples
- Update configuration examples to use ClientConfig
- Update code examples to reflect new architecture

### Step 11.4: Update CHANGELOG.md

**File**: `CHANGELOG.md`

```markdown
## [v0.2.0] - 2025-10-29

**Major Architectural Simplification**:

**Removed**:
- `pkg/capabilities/` - Redundant layer after validation removal
- `pkg/models/` as separate package - Now a simple config struct
- Capability registry and format system
- Option validation infrastructure

**Added**:
- `pkg/types/` - Renamed from protocols, now includes response parsing
- `pkg/client/` - Renamed from transport for clarity
- Transport-layer retry with intelligent error detection
- Protocol-aware parsing functions in types package

**Changed**:
- Model is now a simple configuration struct in `pkg/config`
- Client directly handles protocol execution without capability intermediary
- Providers are the primary differentiation point for LLM formats
- Configuration simplified: protocol name directly identifies capability

**Breaking Changes**:
- All `pkg/protocols` imports → `pkg/types`
- All `pkg/transport` imports → `pkg/client`
- `TransportConfig` → `ClientConfig`
- Capability interface removed (internal only)
- Model interface simplified to struct

**Migration Guide**:
See `_context/migration-v0.2.md` for detailed migration instructions.
```

### Step 11.5: Create Migration Guide

**New File**: `_context/migration-v0.2.md`

Document step-by-step migration for users upgrading from v0.1.x.

---

## Phase 12: Final Verification

### Step 12.1: Build All Packages

**Commands**:
```bash
cd /home/jaime/code/go-agents
go mod tidy
go build ./...
```

### Step 12.2: Run Test Suite

**Commands**:
```bash
go test ./tests/... -v -cover
```

### Step 12.3: Manual Integration Test

**Test with tools/prompt-agent**:
```bash
cd tools/prompt-agent
go run main.go -config config.json
```

---

## Success Criteria

### Architecture
- [x] 4-layer architecture (Agent → Client → Provider → HTTP)
- [x] Capabilities layer removed
- [x] Types package with parsing functions
- [x] Model simplified to config struct
- [x] Client with retry infrastructure
- [x] Modern Go 1.25.2 patterns

### Compilation
- [ ] All packages compile
- [ ] All tests pass
- [ ] Coverage ≥ 80%

### Documentation
- [ ] ARCHITECTURE.md updated
- [ ] PROJECT.md updated
- [ ] README.md updated
- [ ] CHANGELOG.md updated
- [ ] Migration guide created

### Code Reduction
- [ ] ~1300 lines removed
- [ ] ~30% code reduction
- [ ] Simplified package structure

---

## Benefits of Consolidated Architecture

### Simplified
- 4 layers instead of 6
- No capability registry/factory pattern
- No intermediate transformation layers
- Clear responsibility separation

### Maintainable
- Provider is THE differentiation point
- Types package is pure shared types
- Client orchestrates HTTP without format logic
- Easy to add new providers

### Extensible
- Anthropic: Custom PrepareRequest transformation
- Gemini: Different request structure in provider
- Future providers: Implement Provider interface
- No capability registration needed

### Performance
- Fewer allocations (no capability instances)
- Direct protocol-based routing
- Simpler call stack
- Less indirection

---

## Next Steps After Completion

1. **Test with classify-docs prototype**
   - Ensure document classification still works
   - Verify retry behavior
   - Test streaming with vision

2. **Add Anthropic provider**
   - Implement Provider interface
   - Handle system message separation
   - Test with Claude models

3. **Add Google Gemini provider**
   - Transform to contents/parts structure
   - Map parameter names
   - Test with Gemini models

4. **Extract to go-agents-document-context**
   - Create standalone library
   - Extract document, encoding, processing
   - Release v0.1.0

5. **Create standalone classify-docs project**
   - Use consolidated go-agents v0.2.0
   - Integrate document-context library
   - Production hardening
