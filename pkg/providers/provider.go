package providers

import (
	"context"
	"net/http"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/models"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

// Provider defines the interface for LLM service provider implementations.
// Providers handle endpoint routing, authentication, request preparation,
// and response processing for their specific service.
type Provider interface {
	// Name returns the provider identifier.
	Name() string

	// Model returns the model instance managed by this provider.
	Model() models.Model

	// GetEndpoint returns the full endpoint URL for a protocol.
	// Returns an error if the protocol is not supported by this provider.
	GetEndpoint(protocol protocols.Protocol) (string, error)

	// SetHeaders sets provider-specific authentication and custom headers on an HTTP request.
	// This is called after the request is created but before it is executed.
	SetHeaders(req *http.Request)

	// PrepareRequest creates a Request for standard (non-streaming) protocol execution.
	// Converts a protocols.Request into a provider-specific HTTP request structure.
	PrepareRequest(ctx context.Context, protocol protocols.Protocol, request *protocols.Request) (*Request, error)

	// PrepareStreamRequest creates a Request for streaming protocol execution.
	// Adds streaming-specific headers (Accept: text/event-stream, Cache-Control: no-cache).
	PrepareStreamRequest(ctx context.Context, protocol protocols.Protocol, request *protocols.Request) (*Request, error)

	// ProcessResponse processes a standard HTTP response and returns the parsed result.
	// Delegates parsing to the capability's ParseResponse method.
	// Returns an error if the HTTP status is not OK or parsing fails.
	ProcessResponse(response *http.Response, capability capabilities.Capability) (any, error)

	// ProcessStreamResponse processes a streaming HTTP response and returns a channel of chunks.
	// The channel is closed when the stream completes or an error occurs.
	// Context cancellation stops processing and closes the channel.
	ProcessStreamResponse(ctx context.Context, response *http.Response, capability capabilities.StreamingCapability) (<-chan any, error)
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
