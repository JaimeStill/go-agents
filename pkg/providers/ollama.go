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

type OllamaProvider struct {
	*BaseProvider
	options map[string]any
}

func NewOllama(c *config.ProviderConfig) (Provider, error) {
	baseURL := c.BaseURL
	if !strings.HasSuffix(baseURL, "/v1") {
		baseURL = strings.TrimSuffix(baseURL, "/") + "/v1"
	}

	model, err := models.New(c.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to create model: %w", err)
	}

	return &OllamaProvider{
		BaseProvider: NewBaseProvider(c.Name, baseURL, model),
		options:      c.Options,
	}, nil
}

func (p *OllamaProvider) GetEndpoint(protocol protocols.Protocol) (string, error) {
	endpoints := map[protocols.Protocol]string{
		protocols.Chat:       "/chat/completions",
		protocols.Vision:     "/chat/completions",
		protocols.Tools:      "/chat/completions",
		protocols.Embeddings: "/embeddings",
	}

	endpoint, exists := endpoints[protocol]
	if !exists {
		return "", fmt.Errorf("protocol %s not supported by Ollama", protocol)
	}

	return fmt.Sprintf("%s%s", p.BaseURL(), endpoint), nil
}

func (p *OllamaProvider) PrepareRequest(ctx context.Context, protocol protocols.Protocol, request *protocols.Request) (*Request, error) {
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

func (p *OllamaProvider) PrepareStreamRequest(ctx context.Context, protocol protocols.Protocol, request *protocols.Request) (*Request, error) {
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

func (p *OllamaProvider) ProcessResponse(resp *http.Response, capability capabilities.Capability) (any, error) {
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

func (p *OllamaProvider) ProcessStreamResponse(ctx context.Context, resp *http.Response, capability capabilities.StreamingCapability) (<-chan any, error) {
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

			if capability.IsStreamComplete(line) {
				return
			}

			chunk, err := capability.ParseStreamingChunk([]byte(line))
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
