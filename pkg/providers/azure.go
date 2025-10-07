package providers

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/models"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

// AzureProvider implements Provider for Azure OpenAI Service.
// Supports deployment-based routing and both API key and Entra ID authentication.
type AzureProvider struct {
	*BaseProvider
	deployment string
	authType   string
	token      string
	apiVersion string
}

// NewAzure creates a new AzureProvider from configuration.
// Requires "deployment", "auth_type", "token", and "api_version" in options.
// Returns an error if any required option is missing or model creation fails.
func NewAzure(c *config.ProviderConfig) (Provider, error) {
	deployment, ok := c.Options["deployment"].(string)
	if !ok || deployment == "" {
		return nil, fmt.Errorf("deployment is required for Azure provider")
	}

	authType, ok := c.Options["auth_type"].(string)
	if !ok || authType == "" {
		return nil, fmt.Errorf("auth_type is required for Azure provider")
	}

	token, ok := c.Options["token"].(string)
	if !ok || token == "" {
		return nil, fmt.Errorf("token is required for Azure provider")
	}

	apiVersion, ok := c.Options["api_version"].(string)
	if !ok || apiVersion == "" {
		return nil, fmt.Errorf("api_version is required for Azure provider")
	}

	model, err := models.New(c.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to create model: %w", err)
	}

	return &AzureProvider{
		BaseProvider: NewBaseProvider(c.Name, c.BaseURL, model),
		deployment:   deployment,
		authType:     authType,
		token:        token,
		apiVersion:   apiVersion,
	}, nil
}

// GetEndpoint returns the full Azure OpenAI endpoint URL for a protocol.
// Includes deployment name in path and api-version as query parameter.
// Supports chat, vision, tools (all use /deployments/{deployment}/chat/completions),
// and embeddings (/deployments/{deployment}/embeddings).
// Returns an error if the protocol is not supported.
func (p *AzureProvider) GetEndpoint(protocol protocols.Protocol) (string, error) {
	basePath := fmt.Sprintf("/deployments/%s", p.deployment)

	endpoints := map[protocols.Protocol]string{
		protocols.Chat:       basePath + "/chat/completions",
		protocols.Vision:     basePath + "/chat/completions",
		protocols.Tools:      basePath + "/chat/completions",
		protocols.Embeddings: basePath + "/embeddings",
	}

	endpoint, exists := endpoints[protocol]
	if !exists {
		return "", fmt.Errorf("protocol %s not supported by Azure", protocol)
	}

	return fmt.Sprintf("%s%s?api-version=%s", p.BaseURL(), endpoint, p.apiVersion), nil
}

// PrepareRequest prepares a standard (non-streaming) Azure request.
// Marshals the protocol request body and includes protocol headers.
// Returns an error if the endpoint is invalid or marshaling fails.
func (p *AzureProvider) PrepareRequest(ctx context.Context, protocol protocols.Protocol, request *protocols.Request) (*Request, error) {
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

// PrepareStreamRequest prepares a streaming Azure request.
// Adds streaming-specific headers (Accept: text/event-stream, Cache-Control: no-cache).
// Returns an error if the endpoint is invalid or marshaling fails.
func (p *AzureProvider) PrepareStreamRequest(ctx context.Context, protocol protocols.Protocol, request *protocols.Request) (*Request, error) {
	endpoint, err := p.GetEndpoint(protocol)
	if err != nil {
		return nil, err
	}

	body, err := request.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marhsal request: %w", err)
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

// ProcessResponse processes a standard Azure HTTP response.
// Returns an error if the HTTP status is not OK.
// Delegates response parsing to the capability's ParseResponse method.
func (p *AzureProvider) ProcessResponse(resp *http.Response, capability capabilities.Capability) (any, error) {
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return capability.ParseResponse(body)
}

// ProcessStreamResponse processes a streaming Azure HTTP response with SSE format.
// Azure uses "data: " prefix for server-sent events.
// Returns a channel that emits parsed streaming chunks.
// The channel is closed when the stream completes or context is cancelled.
// Returns an error if the HTTP status is not OK.
func (p *AzureProvider) ProcessStreamResponse(ctx context.Context, resp *http.Response, capability capabilities.StreamingCapability) (<-chan any, error) {
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
				case output <- &protocols.StreamingChunk{Error: err}:
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

			if capability.IsStreamComplete(data) {
				return
			}

			chunk, err := capability.ParseStreamingChunk([]byte(data))
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
// Supports "api_key" (api-key header) and "bearer" (Authorization: Bearer <token>).
func (p *AzureProvider) SetHeaders(req *http.Request) {
	switch p.authType {
	case "api_key":
		if p.token != "" {
			req.Header.Set("api-key", p.token)
		}
	case "bearer":
		if p.token != "" {
			req.Header.Set("Authorization", "Bearer "+p.token)
		}
	}
}
