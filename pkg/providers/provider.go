package providers

import (
	"context"
	"net/http"

	"github.com/JaimeStill/go-agents/pkg/types"
)

// Provider defines the interface for LLM service provider implementations.
// Providers handle endpoint routing, authentication, request preparation,
// and response processing for their specific service.
type Provider interface {
	// Name returns the provider identifier.
	Name() string

	// Model returns the model instance managed by this provider.
	Model() *types.Model

	// GetEndpoint returns the full endpoint URL for a protocol.
	// Returns an error if the protocol is not supported by this provider.
	GetEndpoint(protocol types.Protocol) (string, error)

	// SetHeaders sets provider-specific authentication and custom headers on an HTTP request.
	// This is called after the request is created but before it is executed.
	SetHeaders(req *http.Request)

	// PrepareRequest creates a Request for standard (non-streaming) protocol execution.
	// Accepts protocol-specific request types (ChatRequest, VisionRequest, etc.).
	// Marshals the request and prepares it for HTTP transmission.
	PrepareRequest(ctx context.Context, request types.ProtocolRequest) (*Request, error)

	// PrepareStreamRequest creates a Request for streaming protocol execution.
	// Accepts protocol-specific request types and adds streaming-specific headers.
	// Adds Accept: text/event-stream and Cache-Control: no-cache headers.
	PrepareStreamRequest(ctx context.Context, request types.ProtocolRequest) (*Request, error)

	// ProcessResponse processes a standard HTTP response and returns the parsed result.
	// Uses types.ParseResponse for protocol-aware parsing.
	// Returns an error if the HTTP status is not OK or parsing fails.
	ProcessResponse(ctx context.Context, response *http.Response, protocol types.Protocol) (any, error)

	// ProcessStreamResponse processes a streaming HTTP response and returns a channel of chunks.
	// The channel is closed when the stream completes or an error occurs.
	// Context cancellation stops processing and closes the channel.
	ProcessStreamResponse(ctx context.Context, response *http.Response, protocol types.Protocol) (<-chan any, error)
}

// Request represents a prepared provider request with all necessary components for HTTP execution.
// This structure decouples request preparation from HTTP client execution.
type Request struct {
	// URL is the complete endpoint URL including query parameters.
	URL string

	// Headers contains protocol-specific and provider-specific headers.
	Headers map[string]string

	// Body is the marshaled request body ready for HTTP transmission.
	Body []byte
}
