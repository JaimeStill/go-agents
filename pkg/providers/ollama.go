package providers

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"strings"

	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/streaming"
)

// OllamaProvider implements Provider for Ollama services with OpenAI-compatible API.
// Supports local and remote Ollama instances with optional authentication.
type OllamaProvider struct {
	*BaseProvider
	options map[string]any
	stream  streaming.StreamReader
}

// NewOllama creates a new OllamaProvider from configuration.
// Automatically adds /v1 suffix to base URL if not present for OpenAI compatibility.
// Supports optional authentication via "auth_type" and "token" options.
func NewOllama(c *config.ProviderConfig) (Provider, error) {
	baseURL := c.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if !strings.HasSuffix(baseURL, "/v1") {
		baseURL = strings.TrimSuffix(baseURL, "/") + "/v1"
	}

	return &OllamaProvider{
		BaseProvider: NewBaseProvider(c.Name, baseURL),
		options:      c.Options,
		stream:       streaming.NewSSEReader(),
	}, nil
}

// Endpoint returns the full Ollama endpoint URL for a protocol.
// Supports chat, vision, tools (all use /chat/completions), and embeddings (/embeddings).
// Returns an error if the protocol is not supported.
func (p *OllamaProvider) Endpoint(proto protocol.Protocol) (string, error) {
	endpoints := map[protocol.Protocol]string{
		protocol.Chat:       "/chat/completions",
		protocol.Vision:     "/chat/completions",
		protocol.Tools:      "/chat/completions",
		protocol.Embeddings: "/embeddings",
	}

	endpoint, exists := endpoints[proto]
	if !exists {
		return "", fmt.Errorf("protocol %s not supported by Ollama", proto)
	}

	return fmt.Sprintf("%s%s", p.BaseURL(), endpoint), nil
}

// Stream returns the SSE reader for Ollama streaming responses.
func (p *OllamaProvider) Stream() streaming.StreamReader {
	return p.stream
}

// PrepareRequest prepares a standard (non-streaming) Ollama request.
// Returns an error if the endpoint is invalid.
func (p *OllamaProvider) PrepareRequest(ctx context.Context, proto protocol.Protocol, body []byte, headers map[string]string) (*Request, error) {
	endpoint, err := p.Endpoint(proto)
	if err != nil {
		return nil, err
	}

	return &Request{
		URL:     endpoint,
		Headers: headers,
		Body:    body,
	}, nil
}

// PrepareStreamRequest prepares a streaming Ollama request.
// Adds streaming-specific headers (Accept: text/event-stream, Cache-Control: no-cache).
// Returns an error if the endpoint is invalid.
func (p *OllamaProvider) PrepareStreamRequest(ctx context.Context, proto protocol.Protocol, body []byte, headers map[string]string) (*Request, error) {
	endpoint, err := p.Endpoint(proto)
	if err != nil {
		return nil, err
	}

	// Clone headers to avoid mutating the original
	streamHeaders := make(map[string]string)
	maps.Copy(streamHeaders, headers)
	streamHeaders["Accept"] = streaming.SSEMedia
	streamHeaders["Cache-Control"] = "no-cache"

	return &Request{
		URL:     endpoint,
		Headers: streamHeaders,
		Body:    body,
	}, nil
}

// SetHeaders sets optional authentication headers based on provider options.
func (p *OllamaProvider) SetHeaders(ctx context.Context, req *http.Request) error {
	if authType, ok := p.options["auth_type"].(string); ok {
		if token, ok := p.options["token"].(string); ok && token != "" {
			switch authType {
			case "bearer":
				req.Header.Set("Authorization", "Bearer "+token)
			case "api_key":
				headerName := "X-API-Key"
				if head, ok := p.options["auth_header"].(string); ok && head != "" {
					headerName = head
				}
				req.Header.Set(headerName, token)
			}
		}
	}

	return nil
}
