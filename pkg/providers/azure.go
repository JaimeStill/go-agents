package providers

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"

	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/identities"
	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/response"
)

// AzureProvider implements Provider for Azure OpenAI Service.
// Supports deployment-based routing with API key, Entra ID (bearer token),
// and managed identity authentication.
type AzureProvider struct {
	*BaseProvider
	deployment  string
	authType    string
	token       string
	apiVersion  string
	tokenSource *identities.AzureTokenSource
}

// NewAzure creates a new AzureProvider from configuration.
// Requires "deployment", "auth_type", and "api_version" in options.
// For "api_key" and "bearer" auth types, "token" is also required.
// For "managed_identity", optional "resource" (token scope) and "client_id"
// (user-assigned identity) are supported.
// Returns an error if any required option is missing or auth_type is unsupported.
func NewAzure(c *config.ProviderConfig) (Provider, error) {
	deployment, ok := c.Options["deployment"].(string)
	if !ok || deployment == "" {
		return nil, fmt.Errorf("deployment is required for Azure provider")
	}

	authType, ok := c.Options["auth_type"].(string)
	if !ok || authType == "" {
		return nil, fmt.Errorf("auth_type is required for Azure provider")
	}

	apiVersion, ok := c.Options["api_version"].(string)
	if !ok || apiVersion == "" {
		return nil, fmt.Errorf("api_version is required for Azure provider")
	}

	p := &AzureProvider{
		BaseProvider: NewBaseProvider(c.Name, c.BaseURL),
		deployment:   deployment,
		authType:     authType,
		apiVersion:   apiVersion,
	}

	if err := p.initAuth(c); err != nil {
		return nil, err
	}

	return p, nil
}

// Endpoint returns the full Azure OpenAI endpoint URL for a protocol.
// Includes deployment name in path and api-version as query parameter.
// Supports chat, vision, tools (all use /deployments/{deployment}/chat/completions),
// and embeddings (/deployments/{deployment}/embeddings).
// Returns an error if the protocol is not supported.
func (p *AzureProvider) Endpoint(proto protocol.Protocol) (string, error) {
	basePath := fmt.Sprintf("/deployments/%s", p.deployment)

	endpoints := map[protocol.Protocol]string{
		protocol.Chat:       basePath + "/chat/completions",
		protocol.Vision:     basePath + "/chat/completions",
		protocol.Tools:      basePath + "/chat/completions",
		protocol.Embeddings: basePath + "/embeddings",
	}

	endpoint, exists := endpoints[proto]
	if !exists {
		return "", fmt.Errorf("protocol %s not supported by Azure", proto)
	}

	return fmt.Sprintf("%s%s?api-version=%s", p.BaseURL(), endpoint, p.apiVersion), nil
}

// PrepareRequest prepares a standard (non-streaming) Azure request.
// Returns an error if the endpoint is invalid.
func (p *AzureProvider) PrepareRequest(ctx context.Context, proto protocol.Protocol, body []byte, headers map[string]string) (*Request, error) {
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

// PrepareStreamRequest prepares a streaming Azure request.
// Adds streaming-specific headers (Accept: text/event-stream, Cache-Control: no-cache).
// Returns an error if the endpoint is invalid.
func (p *AzureProvider) PrepareStreamRequest(ctx context.Context, proto protocol.Protocol, body []byte, headers map[string]string) (*Request, error) {
	endpoint, err := p.Endpoint(proto)
	if err != nil {
		return nil, err
	}

	// Clone headers to avoid mutating the original
	streamHeaders := make(map[string]string)
	maps.Copy(streamHeaders, headers)
	streamHeaders["Accept"] = "text/event-stream"
	streamHeaders["Cache-Control"] = "no-cache"

	return &Request{
		URL:     endpoint,
		Headers: streamHeaders,
		Body:    body,
	}, nil
}

// ProcessResponse processes a standard Azure HTTP response.
// Returns an error if the HTTP status is not OK.
// Uses response.Parse for protocol-aware parsing.
func (p *AzureProvider) ProcessResponse(ctx context.Context, resp *http.Response, proto protocol.Protocol) (any, error) {
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return response.Parse(proto, body)
}

// ProcessStreamResponse processes a streaming Azure HTTP response with SSE format.
// Azure uses "data: " prefix for server-sent events.
// Returns a channel that emits parsed streaming chunks.
// The channel is closed when the stream completes or context is cancelled.
// Returns an error if the HTTP status is not OK.
func (p *AzureProvider) ProcessStreamResponse(ctx context.Context, resp *http.Response, proto protocol.Protocol) (<-chan any, error) {
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
				case output <- &response.StreamingChunk{Error: err}:
				case <-ctx.Done():
				}
				return
			}

			line = strings.TrimSpace(line)

			if line == "" {
				continue
			}

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			// Check for stream completion marker
			if data == "[DONE]" {
				return
			}

			chunk, err := response.ParseStreamChunk(proto, []byte(data))
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
// Supports "api_key" (api-key header), "bearer" (Authorization: Bearer),
// and "managed_identity" (dynamic token acquisition via Azure identity).
func (p *AzureProvider) SetHeaders(ctx context.Context, req *http.Request) error {
	switch p.authType {
	case "api_key":
		if p.token != "" {
			req.Header.Set("api-key", p.token)
		}
	case "bearer":
		if p.token != "" {
			req.Header.Set("Authorization", "Bearer "+p.token)
		}
	case "managed_identity":
		token, err := p.tokenSource.GetToken(ctx)
		if err != nil {
			return fmt.Errorf("acquire managed identity token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return nil
}

// initAuth initializes authentication based on the configured auth_type.
func (p *AzureProvider) initAuth(c *config.ProviderConfig) error {
	switch p.authType {
	case "api_key", "bearer":
		token, ok := c.Options["token"].(string)
		if !ok || token == "" {
			return fmt.Errorf("token is required for Azure provider with auth_type %q", p.authType)
		}
		p.token = token
	case "managed_identity":
		resource, _ := c.Options["resource"].(string)
		clientID, _ := c.Options["client_id"].(string)

		tokenSource, err := identities.NewAzureTokenSource(resource, clientID)
		if err != nil {
			return fmt.Errorf("initialize managed identity: %w", err)
		}
		p.tokenSource = tokenSource
	default:
		return fmt.Errorf("unsupported auth_type %q for Azure provider", p.authType)
	}

	return nil
}
