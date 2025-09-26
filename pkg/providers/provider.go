package providers

import (
	"context"
	"net/http"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/models"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

type Provider interface {
	Name() string
	Model() models.Model

	GetEndpoint(protocol protocols.Protocol) (string, error)
	SetHeaders(req *http.Request)

	PrepareRequest(ctx context.Context, protocol protocols.Protocol, request *protocols.Request) (*Request, error)
	PrepareStreamRequest(ctx context.Context, protocol protocols.Protocol, request *protocols.Request) (*Request, error)
	ProcessResponse(response *http.Response, capability capabilities.Capability) (any, error)
	ProcessStreamResponse(ctx context.Context, response *http.Response, capability capabilities.StreamingCapability) (<-chan any, error)
}

type Request struct {
	URL     string
	Headers map[string]string
	Body    []byte
}
