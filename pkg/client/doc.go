// Package client provides HTTP orchestration for LLM protocol execution.
// It handles request marshaling, HTTP transport, retry logic, response parsing,
// and health tracking. The client coordinates between the request, provider,
// format, and streaming layers.
//
// # Client Interface
//
//	type Client interface {
//	    HTTPClient() *http.Client
//	    Execute(ctx context.Context, req request.Request) (any, error)
//	    ExecuteStream(ctx context.Context, req request.Request) (<-chan *response.StreamingResponse, error)
//	    IsHealthy() bool
//	}
//
// # Request Flow
//
//  1. Marshal request body through the format layer (req.Marshal)
//  2. Prepare provider request with endpoint URL and headers (provider.PrepareRequest)
//  3. Create HTTP request and apply provider authentication (provider.SetHeaders)
//  4. Execute with retry logic for transient failures
//  5. Parse response through the format layer (format.Parse)
//  6. Update health status based on result
//
// # Streaming Flow
//
//  1. Same preparation as standard flow, using PrepareStreamRequest
//  2. Provider's StreamReader decodes the transport framing (SSE or event stream)
//  3. Format's ParseStreamChunk translates payloads to StreamingResponse
//  4. Chunks delivered via channel; nil chunks are skipped
//  5. Channel closes on completion, error, or context cancellation
//
// # Retry Logic
//
// Standard requests are retried with exponential backoff and jitter for
// transient failures (HTTP 429, 502, 503, 504, network errors).
// Streaming requests are not retried.
package client
