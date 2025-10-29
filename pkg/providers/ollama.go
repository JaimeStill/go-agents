package providers

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/types"
)

// OllamaProvider implements Provider for Ollama services with OpenAI-compatible API.
// Supports local and remote Ollama instances with optional authentication.
type OllamaProvider struct {
	*BaseProvider
	options map[string]any
}

// NewOllama creates a new OllamaProvider from configuration.
// Automatically adds /v1 suffix to base URL if not present for OpenAI compatibility.
// Supports optional authentication via "auth_type" and "token" options.
// Returns an error if model creation fails.
func NewOllama(c *config.ProviderConfig) (Provider, error) {
	baseURL := c.BaseURL
	if !strings.HasSuffix(baseURL, "/v1") {
		baseURL = strings.TrimSuffix(baseURL, "/") + "/v1"
	}

	model := types.FromConfig(c.Model)

	return &OllamaProvider{
		BaseProvider: NewBaseProvider(c.Name, baseURL, model),
		options:      c.Options,
	}, nil
}

// GetEndpoint returns the full Ollama endpoint URL for a protocol.
// Supports chat, vision, tools (all use /chat/completions), and embeddings (/embeddings).
// Returns an error if the protocol is not supported.
func (p *OllamaProvider) GetEndpoint(protocol types.Protocol) (string, error) {
	endpoints := map[types.Protocol]string{
		types.Chat:       "/chat/completions",
		types.Vision:     "/chat/completions",
		types.Tools:      "/chat/completions",
		types.Embeddings: "/embeddings",
	}

	endpoint, exists := endpoints[protocol]
	if !exists {
		return "", fmt.Errorf("protocol %s not supported by Ollama", protocol)
	}

	return fmt.Sprintf("%s%s", p.BaseURL(), endpoint), nil
}

// PrepareRequest prepares a standard (non-streaming) Ollama request.
// Marshals the protocol request body and includes protocol headers.
// Returns an error if the endpoint is invalid or marshaling fails.
func (p *OllamaProvider) PrepareRequest(ctx context.Context, request types.ProtocolRequest) (*Request, error) {
	protocol := request.GetProtocol()
	endpoint, err := p.GetEndpoint(protocol)
	if err != nil {
		return nil, err
	}

	body, err := request.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	return &Request{
		URL:     endpoint,
		Headers: request.GetHeaders(),
		Body:    body,
	}, nil
}

// PrepareStreamRequest prepares a streaming Ollama request.
// Adds streaming-specific headers (Accept: text/event-stream, Cache-Control: no-cache).
// Returns an error if the endpoint is invalid or marshaling fails.
func (p *OllamaProvider) PrepareStreamRequest(ctx context.Context, request types.ProtocolRequest) (*Request, error) {
	protocol := request.GetProtocol()
	endpoint, err := p.GetEndpoint(protocol)
	if err != nil {
		return nil, err
	}

	body, err := request.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	headers := request.GetHeaders()
	headers["Accept"] = "text/event-stream"
	headers["Cache-Control"] = "no-cache"

	return &Request{
		URL:     endpoint,
		Headers: headers,
		Body:    body,
	}, nil
}

// ProcessResponse processes a standard Ollama HTTP response.
// Returns an error if the HTTP status is not OK.
// Uses types.ParseResponse for protocol-aware parsing.
func (p *OllamaProvider) ProcessResponse(ctx context.Context, resp *http.Response, protocol types.Protocol) (any, error) {
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return types.ParseResponse(protocol, body)
}

// ProcessStreamResponse processes a streaming Ollama HTTP response.
// Ollama uses SSE format with "data: " prefix.
// Returns a channel that emits parsed streaming chunks.
// The channel is closed when the stream completes or context is cancelled.
// Returns an error if the HTTP status is not OK.
func (p *OllamaProvider) ProcessStreamResponse(ctx context.Context, resp *http.Response, protocol types.Protocol) (<-chan any, error) {
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	output := make(chan any)

	go func() {
		defer close(output)
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)

		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				select {
				case output <- &types.StreamingChunk{Error: err}:
				case <-ctx.Done():
				}
				return
			}

			line = strings.TrimSpace(line)

			if line == "" {
				continue
			}

			// Check for completion marker
			if line == "data: [DONE]" {
				return
			}

			// Strip SSE "data: " prefix
			if strings.HasPrefix(line, "data: ") {
				line = strings.TrimPrefix(line, "data: ")
			}

			chunk, err := types.ParseStreamChunk(protocol, []byte(line))
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

// SetHeaders sets authentication headers on the HTTP request.
// Supports "bearer" token (Authorization: Bearer <token>) and "api_key" (custom header).
// The "auth_header" option allows customizing the API key header name (default: X-API-Key).
func (p *OllamaProvider) SetHeaders(req *http.Request) {
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
}
