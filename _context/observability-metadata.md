# Observability Metadata Implementation Guide

## Problem Context

The go-agents library currently processes LLM requests and responses but does not preserve metadata necessary for debugging, performance analysis, and agent tuning. While responses contain provider metadata (IDs, timestamps, token usage), this information is not systematically captured or made available to library consumers.

**Key Requirements**:
- Preserve request/response metadata at each layer (capability, provider, transport)
- Distinguish between observability metadata (passive, format-agnostic), extended metadata (provider-specific), and statistical metadata (computed)
- Maintain clean separation of concerns with layer-specific metadata enrichment methods
- Zero API changes - preserve existing return types and interfaces
- Enable comprehensive observability through embedded Response struct

**Current Limitations**:
- Protocol response structures have unused metadata fields (ID, Object, Created)
- No request correlation or timing information preserved
- Provider-specific metadata (request IDs, deployment info) not captured
- No raw request/response data available for debugging

## Architecture Approach

**Response Embedding Pattern**: Create a `Response` struct that all protocol responses embed. Each layer enriches the metadata through dedicated methods that keep execution flows clean.

**Layer Responsibility Model**:
- **Capability Layer**: Initializes metadata structure, captures format-standard metadata (finish_reason, tool calls)
- **Provider Layer**: Enriches with provider-specific metadata (request IDs, deployment info, rate limits)
- **Transport Layer**: Adds timing, correlation, retry information, and raw data preservation

**Core Structures**:
```go
// Response embeds in all protocol responses
type Response struct {
    // Provider response metadata (already exists in responses)
    ID      string `json:"id,omitempty"`
    Object  string `json:"object,omitempty"`
    Created int64  `json:"created,omitempty"`
    Model   string `json:"model"`

    // Observability metadata (never serialized)
    Meta     ResponseMeta    `json:"-"`
    Extended map[string]any  `json:"-"`
}

// ResponseMeta contains observability data
type ResponseMeta struct {
    // Request correlation
    RequestID         string
    ProviderRequestID string

    // Context
    Provider string
    Protocol Protocol

    // Timing information
    TotalDuration   time.Duration
    RequestDuration time.Duration

    // Retry information
    Attempts     int
    RetryReasons []string

    // Raw data (suppressible via config)
    Raw *RawData
}

// RawData preserves original request/response
type RawData struct {
    Request  []byte
    Response []byte
}
```

**Metadata Flow**:
1. **Transport.Execute()**: Creates RequestID, captures timing start
2. **Capability.ParseResponse()**: Initializes Response, populates format-standard Extended fields via `enrichMetadata()`
3. **Provider.ProcessResponse()**: Enriches with provider metadata via `enrichMetadata()`
4. **Transport.Execute()**: Completes timing, adds retry info, preserves raw data via `completeMetadata()`
5. **Consumer**: Accesses metadata through embedded Response fields after type assertion

**Field Access Pattern**:
```go
result, _ := capability.ParseResponse(data)
chatResp := result.(*protocols.ChatResponse)

// Access protocol-specific fields
chatResp.Choices

// Access metadata via embedding promotion
chatResp.Meta.RequestID
chatResp.Meta.TotalDuration
chatResp.Extended["finish_reason"]
```

## Implementation Phases

### Preparation Phase

This phase refactors existing structures without changing functionality, establishing the foundation for metadata preservation.

#### Step 1: Define Response Structure and Metadata Types

**File**: `pkg/protocols/protocol.go`

**After existing Protocol type definition** (~line 30), add:

```go
// Response embeds in all protocol responses to provide metadata
type Response struct {
    // Provider response metadata
    ID      string `json:"id,omitempty"`
    Object  string `json:"object,omitempty"`
    Created int64  `json:"created,omitempty"`
    Model   string `json:"model"`

    // Observability metadata (never serialized)
    Meta     ResponseMeta    `json:"-"`
    Extended map[string]any  `json:"-"`
}

// ResponseMeta contains observability metadata
type ResponseMeta struct {
    // Request correlation
    RequestID         string
    ProviderRequestID string

    // Context
    Provider string
    Protocol Protocol

    // Timing
    TotalDuration   time.Duration
    RequestDuration time.Duration

    // Retry information
    Attempts     int
    RetryReasons []string

    // Raw data preservation
    Raw *RawData
}

// RawData preserves original request/response for debugging
type RawData struct {
    Request  []byte
    Response []byte
}

// InitializeResponse creates a Response with basic metadata
func InitializeResponse(protocol Protocol) Response {
    return Response{
        Meta: ResponseMeta{
            Protocol: protocol,
        },
        Extended: make(map[string]any),
    }
}
```

**Rationale**: Defining the base structures in the protocols package establishes the foundation. The `InitializeResponse` helper ensures consistent initialization across capabilities. The `json:"-"` tags prevent metadata serialization while preserving existing JSON behavior.

#### Step 2: Embed Response in Protocol Responses

**File**: `pkg/protocols/protocol.go`

**Replace existing ChatResponse** (~line 50):

```go
type ChatResponse struct {
    Response
    Choices []struct {
        Index        int     `json:"index"`
        Message      Message `json:"message"`
        FinishReason string  `json:"finish_reason,omitempty"`
    } `json:"choices"`
    Usage *TokenUsage `json:"usage,omitempty"`
}
```

**Replace existing ToolsResponse** (~line 80):

```go
type ToolsResponse struct {
    Response
    Choices []struct {
        Index        int           `json:"index"`
        Message      ToolMessage   `json:"message"`
        FinishReason string        `json:"finish_reason,omitempty"`
    } `json:"choices"`
    Usage *TokenUsage `json:"usage,omitempty"`
}
```

**Replace existing EmbeddingsResponse** (~line 110):

```go
type EmbeddingsResponse struct {
    Response
    Data []struct {
        Index     int       `json:"index"`
        Embedding []float64 `json:"embedding"`
    } `json:"data"`
    Usage *TokenUsage `json:"usage,omitempty"`
}
```

**Rationale**: Embedding `Response` in all protocol responses provides metadata fields through promotion while preserving existing JSON serialization. The duplicate fields (ID, Object, Created, Model) are removed from the embedding since Response already defines them.

### Feature Phase

This phase implements metadata capture and enrichment at each layer.

#### Step 3: Implement Capability Layer Metadata Initialization

**File**: `pkg/capabilities/chat.go`

**Locate OpenAIChatCapability.ParseResponse** (~line 100) and replace:

```go
func (c *OpenAIChatCapability) ParseResponse(data []byte) (any, error) {
    var response protocols.ChatResponse
    if err := json.Unmarshal(data, &response); err != nil {
        return nil, fmt.Errorf("failed to parse chat response: %w", err)
    }

    c.enrichMetadata(&response)
    return &response, nil
}

func (c *OpenAIChatCapability) enrichMetadata(response *protocols.ChatResponse) {
    response.Response = protocols.InitializeResponse(c.protocol)

    // Capture format-standard extended metadata
    if len(response.Choices) > 0 {
        if response.Choices[0].FinishReason != "" {
            response.Extended["finish_reason"] = response.Choices[0].FinishReason
        }
    }

    // Preserve token usage if present
    if response.Usage != nil {
        response.Extended["usage"] = map[string]any{
            "prompt_tokens":     response.Usage.PromptTokens,
            "completion_tokens": response.Usage.CompletionTokens,
            "total_tokens":      response.Usage.TotalTokens,
        }
    }
}
```

**File**: `pkg/capabilities/tools.go`

**Locate OpenAIToolsCapability.ParseResponse** (~line 120) and replace:

```go
func (c *OpenAIToolsCapability) ParseResponse(data []byte) (any, error) {
    var response protocols.ToolsResponse
    if err := json.Unmarshal(data, &response); err != nil {
        return nil, fmt.Errorf("failed to parse tools response: %w", err)
    }

    c.enrichMetadata(&response)
    return &response, nil
}

func (c *OpenAIToolsCapability) enrichMetadata(response *protocols.ToolsResponse) {
    response.Response = protocols.InitializeResponse(c.protocol)

    // Capture format-standard extended metadata
    if len(response.Choices) > 0 {
        if response.Choices[0].FinishReason != "" {
            response.Extended["finish_reason"] = response.Choices[0].FinishReason
        }

        if len(response.Choices[0].Message.ToolCalls) > 0 {
            response.Extended["tool_calls_count"] = len(response.Choices[0].Message.ToolCalls)
        }
    }

    if response.Usage != nil {
        response.Extended["usage"] = map[string]any{
            "prompt_tokens":     response.Usage.PromptTokens,
            "completion_tokens": response.Usage.CompletionTokens,
            "total_tokens":      response.Usage.TotalTokens,
        }
    }
}
```

**File**: `pkg/capabilities/embeddings.go`

**Locate OpenAIEmbeddingsCapability.ParseResponse** (~line 90) and replace:

```go
func (c *OpenAIEmbeddingsCapability) ParseResponse(data []byte) (any, error) {
    var response protocols.EmbeddingsResponse
    if err := json.Unmarshal(data, &response); err != nil {
        return nil, fmt.Errorf("failed to parse embeddings response: %w", err)
    }

    c.enrichMetadata(&response)
    return &response, nil
}

func (c *OpenAIEmbeddingsCapability) enrichMetadata(response *protocols.EmbeddingsResponse) {
    response.Response = protocols.InitializeResponse(c.protocol)

    // Capture embeddings-specific metadata
    response.Extended["embeddings_count"] = len(response.Data)

    if response.Usage != nil {
        response.Extended["usage"] = map[string]any{
            "prompt_tokens": response.Usage.PromptTokens,
            "total_tokens":  response.Usage.TotalTokens,
        }
    }
}
```

**Rationale**: Separate `enrichMetadata` methods keep parsing clean and focused. Capability layer captures format-standard metadata that applies across all providers supporting that format. The pattern is consistent across all capability types.

#### Step 4: Implement Provider Layer Metadata Enrichment

**File**: `pkg/providers/openai.go`

**Locate OpenAIProvider.ProcessResponse** (~line 80) and replace:

```go
func (p *OpenAIProvider) ProcessResponse(resp *http.Response, capability capabilities.Capability) (any, error) {
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %w", err)
    }

    response, err := capability.ParseResponse(body)
    if err != nil {
        return nil, err
    }

    p.enrichMetadata(response, resp)
    return response, nil
}

func (p *OpenAIProvider) enrichMetadata(response any, resp *http.Response) {
    // Type assertion to access Response fields
    // All protocol responses embed Response, but we need concrete type
    type hasResponse interface {
        getResponse() *protocols.Response
    }

    // Use reflection helper or type switches
    switch r := response.(type) {
    case *protocols.ChatResponse:
        p.enrichResponseMeta(&r.Response, resp)
    case *protocols.ToolsResponse:
        p.enrichResponseMeta(&r.Response, resp)
    case *protocols.EmbeddingsResponse:
        p.enrichResponseMeta(&r.Response, resp)
    }
}

func (p *OpenAIProvider) enrichResponseMeta(response *protocols.Response, resp *http.Response) {
    // Provider context
    response.Meta.Provider = p.Name()

    // OpenAI-specific request ID
    if reqID := resp.Header.Get("x-request-id"); reqID != "" {
        response.Meta.ProviderRequestID = reqID
        response.Extended["openai_request_id"] = reqID
    }

    // Rate limiting headers
    if remaining := resp.Header.Get("x-ratelimit-remaining-requests"); remaining != "" {
        response.Extended["ratelimit_remaining_requests"] = remaining
    }
    if resetRequests := resp.Header.Get("x-ratelimit-reset-requests"); resetRequests != "" {
        response.Extended["ratelimit_reset_requests"] = resetRequests
    }
    if remaining := resp.Header.Get("x-ratelimit-remaining-tokens"); remaining != "" {
        response.Extended["ratelimit_remaining_tokens"] = remaining
    }
    if resetTokens := resp.Header.Get("x-ratelimit-reset-tokens"); resetTokens != "" {
        response.Extended["ratelimit_reset_tokens"] = resetTokens
    }
}
```

**File**: `pkg/providers/azure.go`

**Locate AzureProvider.ProcessResponse** (~line 90) and replace:

```go
func (p *AzureProvider) ProcessResponse(resp *http.Response, capability capabilities.Capability) (any, error) {
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %w", err)
    }

    response, err := capability.ParseResponse(body)
    if err != nil {
        return nil, err
    }

    p.enrichMetadata(response, resp)
    return response, nil
}

func (p *AzureProvider) enrichMetadata(response any, resp *http.Response) {
    switch r := response.(type) {
    case *protocols.ChatResponse:
        p.enrichResponseMeta(&r.Response, resp)
    case *protocols.ToolsResponse:
        p.enrichResponseMeta(&r.Response, resp)
    case *protocols.EmbeddingsResponse:
        p.enrichResponseMeta(&r.Response, resp)
    }
}

func (p *AzureProvider) enrichResponseMeta(response *protocols.Response, resp *http.Response) {
    // Provider context
    response.Meta.Provider = p.Name()

    // Azure-specific metadata
    response.Extended["azure_deployment"] = p.deployment

    if reqID := resp.Header.Get("apim-request-id"); reqID != "" {
        response.Meta.ProviderRequestID = reqID
        response.Extended["azure_request_id"] = reqID
    }

    // Azure rate limiting
    if remaining := resp.Header.Get("x-ratelimit-remaining-requests"); remaining != "" {
        response.Extended["ratelimit_remaining_requests"] = remaining
    }
    if remaining := resp.Header.Get("x-ratelimit-remaining-tokens"); remaining != "" {
        response.Extended["ratelimit_remaining_tokens"] = remaining
    }
}
```

**File**: `pkg/providers/ollama.go`

**Locate OllamaProvider.ProcessResponse** (~line 70) and replace:

```go
func (p *OllamaProvider) ProcessResponse(resp *http.Response, capability capabilities.Capability) (any, error) {
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %w", err)
    }

    response, err := capability.ParseResponse(body)
    if err != nil {
        return nil, err
    }

    p.enrichMetadata(response, resp)
    return response, nil
}

func (p *OllamaProvider) enrichMetadata(response any, resp *http.Response) {
    switch r := response.(type) {
    case *protocols.ChatResponse:
        p.enrichResponseMeta(&r.Response, resp)
    case *protocols.ToolsResponse:
        p.enrichResponseMeta(&r.Response, resp)
    case *protocols.EmbeddingsResponse:
        p.enrichResponseMeta(&r.Response, resp)
    }
}

func (p *OllamaProvider) enrichResponseMeta(response *protocols.Response, resp *http.Response) {
    // Provider context
    response.Meta.Provider = p.Name()

    // Ollama doesn't provide request IDs in headers
    // Extended metadata for local model context
    response.Extended["ollama_local"] = true
}
```

**Rationale**: Each provider's `enrichMetadata` method uses type switching to access the embedded Response field, then delegates to `enrichResponseMeta` for the actual enrichment. This pattern keeps the type assertion logic centralized while allowing provider-specific metadata capture. The type switch is unavoidable given Go's type system and the `any` return type.

#### Step 5: Implement Transport Layer Metadata Completion

**File**: `pkg/transport/client.go`

**Add import** at top of file:

```go
import (
    "github.com/google/uuid"
    // ... existing imports
)
```

**Locate Client struct** (~line 20) and add config field:

```go
type Client struct {
    httpClient *http.Client
    config     config.TransportConfig
}
```

**Locate Client.Execute** (~line 120) and replace:

```go
func (c *Client) Execute(ctx context.Context, request ExecuteRequest) (any, error) {
    // Generate request ID for correlation
    requestID := uuid.New().String()

    // Start timing
    startTime := time.Now()

    // Create HTTP request
    httpReq, err := c.createHTTPRequest(ctx, request)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    // Preserve raw request if enabled
    var rawRequest []byte
    if c.config.PreserveRawData {
        rawRequest, _ = io.ReadAll(httpReq.Body)
        httpReq.Body = io.NopCloser(bytes.NewReader(rawRequest))
    }

    // Execute request with retry
    requestStart := time.Now()
    resp, attempts, retryReasons, err := c.executeWithRetry(httpReq)
    requestDuration := time.Since(requestStart)

    if err != nil {
        return nil, err
    }

    // Process response through provider
    response, err := request.Provider.ProcessResponse(resp, request.Capability)
    if err != nil {
        return nil, err
    }

    // Complete metadata
    c.completeMetadata(response, requestID, startTime, requestDuration, attempts, retryReasons, rawRequest)

    return response, nil
}

func (c *Client) completeMetadata(
    response any,
    requestID string,
    startTime time.Time,
    requestDuration time.Duration,
    attempts int,
    retryReasons []string,
    rawRequest []byte,
) {
    // Type switch to access Response field
    switch r := response.(type) {
    case *protocols.ChatResponse:
        c.completeResponseMeta(&r.Response, requestID, startTime, requestDuration, attempts, retryReasons, rawRequest)
    case *protocols.ToolsResponse:
        c.completeResponseMeta(&r.Response, requestID, startTime, requestDuration, attempts, retryReasons, rawRequest)
    case *protocols.EmbeddingsResponse:
        c.completeResponseMeta(&r.Response, requestID, startTime, requestDuration, attempts, retryReasons, rawRequest)
    }
}

func (c *Client) completeResponseMeta(
    response *protocols.Response,
    requestID string,
    startTime time.Time,
    requestDuration time.Duration,
    attempts int,
    retryReasons []string,
    rawRequest []byte,
) {
    // Request correlation
    response.Meta.RequestID = requestID

    // Timing
    response.Meta.TotalDuration = time.Since(startTime)
    response.Meta.RequestDuration = requestDuration

    // Retry information
    response.Meta.Attempts = attempts
    response.Meta.RetryReasons = retryReasons

    // Raw data preservation
    if c.config.PreserveRawData && rawRequest != nil {
        response.Meta.Raw = &protocols.RawData{
            Request: rawRequest,
        }

        // Note: Response body already consumed by provider
        // Could preserve if needed by reading before provider processing
    }
}

func (c *Client) executeWithRetry(req *http.Request) (*http.Response, int, []string, error) {
    var attempts int
    var retryReasons []string
    var lastErr error

    for i := 0; i < c.config.RetryAttempts; i++ {
        attempts++

        resp, err := c.httpClient.Do(req)
        if err == nil && resp.StatusCode < 500 {
            return resp, attempts, retryReasons, nil
        }

        // Capture retry reason
        if err != nil {
            retryReasons = append(retryReasons, fmt.Sprintf("attempt %d: %v", i+1, err))
            lastErr = err
        } else {
            retryReasons = append(retryReasons, fmt.Sprintf("attempt %d: status %d", i+1, resp.StatusCode))
            lastErr = fmt.Errorf("status code %d", resp.StatusCode)
            resp.Body.Close()
        }

        // Wait before retry
        if i < c.config.RetryAttempts-1 {
            time.Sleep(time.Duration(c.config.RetryDelay))
        }
    }

    return nil, attempts, retryReasons, lastErr
}
```

**Rationale**: Transport layer completes the metadata with timing, correlation, and retry information. The `completeMetadata` method uses type switching to access the embedded Response field, keeping the main execution flow clean. The `executeWithRetry` method now returns attempts and retry reasons for metadata enrichment.

#### Step 6: Update Configuration for Raw Data Control

**File**: `pkg/config/transport.go`

**Locate TransportConfig struct** (~line 20) and add field:

```go
type TransportConfig struct {
    Timeout         Duration `json:"timeout"`
    RetryAttempts   int      `json:"retry_attempts"`
    RetryDelay      Duration `json:"retry_delay"`
    PreserveRawData bool     `json:"preserve_raw_data"`
}
```

**Locate DefaultTransportConfig** (~line 40) and update:

```go
func DefaultTransportConfig() TransportConfig {
    return TransportConfig{
        Timeout:         Duration(30 * time.Second),
        RetryAttempts:   3,
        RetryDelay:      Duration(1 * time.Second),
        PreserveRawData: true,
    }
}
```

**Rationale**: Configuration option allows suppressing raw data preservation if memory is a concern, while defaulting to preservation for maximum debugging value. The boolean flag integrates cleanly with existing transport configuration.

## Testing Considerations

**Unit Tests** (`tests/protocols/`):
- Verify `Response` fields accessible through embedding
- Verify `InitializeResponse()` creates proper structure
- Verify metadata fields are not serialized (`json:"-"` tags effective)
- Test that existing JSON serialization behavior preserved

**Unit Tests** (`tests/capabilities/`):
- Verify `enrichMetadata` initializes Extended map
- Verify format-standard metadata captured (finish_reason, usage)
- Verify protocol set correctly in metadata
- Test each capability type (chat, tools, embeddings)

**Unit Tests** (`tests/providers/`):
- Verify `enrichMetadata` enriches provider metadata
- Verify provider-specific headers captured (request IDs, rate limits)
- Verify provider name set in metadata
- Mock HTTP responses to test header extraction
- Test type switching works for all response types

**Unit Tests** (`tests/transport/`):
- Verify request ID generation
- Verify timing metadata accuracy
- Verify retry attempts and reasons captured
- Verify raw data preservation when enabled
- Verify raw data suppression when disabled
- Test type switching in `completeMetadata`

**Integration Validation** (Manual via `tools/prompt-agent`):
- Execute chat request, verify complete metadata chain
- Execute with retries, verify retry information captured
- Execute with different providers, verify provider-specific metadata
- Verify rate limit headers captured (OpenAI, Azure)
- Verify raw request preservation
- Access metadata after type assertion: `response.(*protocols.ChatResponse).Meta`

## Future Extensibility

**Streaming Metadata Preservation**:
When streaming support is added, metadata for streaming responses can be captured similarly:

```go
type StreamChunk struct {
    Response
    Choices []struct {
        Index int                 `json:"index"`
        Delta protocols.Message   `json:"delta"`
    } `json:"choices"`
}

// Metadata accumulated across chunks
func (c *Client) ExecuteStream(ctx context.Context, request ExecuteRequest) (<-chan StreamChunk, error) {
    // Initialize metadata once at stream start
    // Update timing metadata at stream completion
}
```

**Statistical Metadata**:
Computed statistics can be added to Extended map by higher layers:

```go
// In agent layer after receiving response
chatResp := response.(*protocols.ChatResponse)
chatResp.Extended["confidence_score"] = calculateConfidence(chatResp)
chatResp.Extended["cost_estimate"] = estimateCost(chatResp)
```

**Content Filtering Metadata**:
Provider-specific content filtering can be captured in Extended:

```go
// In Azure provider enrichResponseMeta
if filterResult := resp.Header.Get("content-filter-result"); filterResult != "" {
    response.Extended["content_filter_result"] = filterResult
}
```

**Reasoning Traces**:
Extended metadata supports variable provider features:

```go
// In OpenAI provider for o1/o3 models
var parsed map[string]any
json.Unmarshal(body, &parsed)

if reasoning, ok := parsed["reasoning"]; ok {
    response.Extended["reasoning_trace"] = reasoning
}
```

**Response Body Preservation**:
Raw response body can be preserved by reading before provider processing:

```go
// In transport layer before provider.ProcessResponse
body, _ := io.ReadAll(resp.Body)
resp.Body = io.NopCloser(bytes.NewReader(body))

response, err := request.Provider.ProcessResponse(resp, request.Capability)

// Store raw response
if c.config.PreserveRawData {
    // Access through type switch
    chatResp := response.(*protocols.ChatResponse)
    chatResp.Meta.Raw.Response = body
}
```
