package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/providers"
	"github.com/JaimeStill/go-agents/pkg/types"
)

// Client provides the interface for executing LLM protocol requests.
// It orchestrates the flow from protocol selection through HTTP execution
// to response processing, with model option merging, retry logic, and health tracking.
type Client interface {
	// Provider returns the provider instance managed by this client.
	Provider() providers.Provider

	// Model returns the model instance managed by this client.
	Model() *types.Model

	// HTTPClient returns a configured HTTP client.
	// Creates a new client on each call with timeout and connection pool settings.
	HTTPClient() *http.Client

	// ExecuteProtocol executes a protocol request and returns the parsed response.
	// Accepts protocol-specific request types (ChatRequest, VisionRequest, etc.).
	// Automatically retries on transient failures (HTTP 429/502/503/504, network errors).
	// Returns an error if protocol is not supported or request fails.
	ExecuteProtocol(ctx context.Context, request types.ProtocolRequest) (any, error)

	// ExecuteProtocolStream executes a streaming protocol request and returns a channel of chunks.
	// Accepts protocol-specific request types and sets stream:true.
	// The channel is closed when streaming completes or context is cancelled.
	// Returns an error if protocol doesn't support streaming or request fails.
	ExecuteProtocolStream(ctx context.Context, request types.ProtocolRequest) (<-chan *types.StreamingChunk, error)

	// IsHealthy returns the current health status of the client.
	// Set to false after request failures, true after successful requests.
	// Thread-safe for concurrent access.
	IsHealthy() bool
}

// client implements the Client interface with provider and model orchestration.
type client struct {
	provider providers.Provider
	model    *types.Model
	config   *config.ClientConfig

	mutex      sync.RWMutex
	healthy    bool
	lastHealth time.Time
}

// New creates a new Client from configuration.
// Creates the provider from configuration and initializes health tracking.
// Returns an error if provider creation fails.
func New(cfg *config.ClientConfig) (Client, error) {
	provider, err := providers.Create(cfg.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	return &client{
		provider:   provider,
		model:      provider.Model(),
		config:     cfg,
		healthy:    true,
		lastHealth: time.Now(),
	}, nil
}

// Provider returns the provider instance managed by this client.
func (c *client) Provider() providers.Provider {
	return c.provider
}

// Model returns the model instance managed by this client.
func (c *client) Model() *types.Model {
	return c.model
}

// HTTPClient creates and returns a configured HTTP client.
// Each call creates a new client with timeout and connection pool settings from configuration.
func (c *client) HTTPClient() *http.Client {
	return &http.Client{
		Timeout: c.config.Timeout.ToDuration(),
		Transport: &http.Transport{
			MaxIdleConns:        c.config.ConnectionPoolSize,
			MaxIdleConnsPerHost: c.config.ConnectionPoolSize,
			IdleConnTimeout:     c.config.ConnectionTimeout.ToDuration(),
		},
	}
}

// ExecuteProtocol executes a standard (non-streaming) protocol request.
// Accepts protocol-specific request types and executes with retry.
func (c *client) ExecuteProtocol(ctx context.Context, request types.ProtocolRequest) (any, error) {
	// Execute with retry - retry logic determines if errors are retryable
	return doWithRetry(ctx, c.config.Retry, func(ctx context.Context) (any, error) {
		return c.execute(ctx, request)
	})
}

// execute performs a single HTTP request attempt without retry logic.
// Returns HTTPStatusError for bad status codes, which retry logic evaluates.
func (c *client) execute(ctx context.Context, request types.ProtocolRequest) (any, error) {
	protocol := request.GetProtocol()

	// Prepare provider request
	providerRequest, err := c.provider.PrepareRequest(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		providerRequest.URL,
		bytes.NewBuffer(providerRequest.Body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	for key, value := range providerRequest.Headers {
		httpReq.Header.Set(key, value)
	}
	c.provider.SetHeaders(httpReq)

	// Execute HTTP request
	httpClient := c.HTTPClient()
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		c.setHealthy(false)
		return nil, err // Network error - retry logic will evaluate
	}
	defer resp.Body.Close()

	// Check for non-OK status - return HTTPStatusError for retry evaluation
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		c.setHealthy(false)
		return nil, &HTTPStatusError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       bodyBytes,
		}
	}

	// Process response through provider
	result, err := c.provider.ProcessResponse(ctx, resp, protocol)
	if err != nil {
		c.setHealthy(false)
		return nil, err
	}

	c.setHealthy(true)
	return result, nil
}

// ExecuteProtocolStream executes a streaming protocol request.
// Verifies protocol supports streaming and executes streaming flow.
func (c *client) ExecuteProtocolStream(ctx context.Context, request types.ProtocolRequest) (<-chan *types.StreamingChunk, error) {
	protocol := request.GetProtocol()

	// Verify protocol supports streaming using Protocol method
	if !protocol.SupportsStreaming() {
		return nil, fmt.Errorf("protocol %s does not support streaming", protocol)
	}

	return c.executeStream(ctx, request)
}

// executeStream performs the streaming HTTP request.
// Streaming requests are not retried - they fail immediately on error.
func (c *client) executeStream(ctx context.Context, request types.ProtocolRequest) (<-chan *types.StreamingChunk, error) {
	protocol := request.GetProtocol()

	// Prepare streaming request
	providerRequest, err := c.provider.PrepareStreamRequest(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare streaming request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		providerRequest.URL,
		bytes.NewBuffer(providerRequest.Body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	for key, value := range providerRequest.Headers {
		httpReq.Header.Set(key, value)
	}
	c.provider.SetHeaders(httpReq)

	// Execute HTTP request
	httpClient := c.HTTPClient()
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		c.setHealthy(false)
		return nil, fmt.Errorf("streaming request failed: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		c.setHealthy(false)
		return nil, fmt.Errorf("streaming request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Process stream through provider
	stream, err := c.provider.ProcessStreamResponse(ctx, resp, protocol)
	if err != nil {
		c.setHealthy(false)
		resp.Body.Close()
		return nil, err
	}

	// Convert provider stream to typed chunk stream
	output := make(chan *types.StreamingChunk)
	go func() {
		defer close(output)
		defer resp.Body.Close()

		for data := range stream {
			if chunk, ok := data.(*types.StreamingChunk); ok {
				select {
				case output <- chunk:
				case <-ctx.Done():
					return
				}
			}
		}
		c.setHealthy(true)
	}()

	return output, nil
}

// IsHealthy returns the current health status.
// Thread-safe for concurrent access via read mutex.
func (c *client) IsHealthy() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.healthy
}

// setHealthy updates the health status with timestamp.
// Thread-safe via write mutex.
func (c *client) setHealthy(healthy bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.healthy = healthy
	c.lastHealth = time.Now()
}
