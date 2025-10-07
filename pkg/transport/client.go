package transport

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/models"
	"github.com/JaimeStill/go-agents/pkg/protocols"
	"github.com/JaimeStill/go-agents/pkg/providers"
)

// ClientRequest is reserved for future use in request batching or queuing.
// Currently not exposed in the public API.
type ClientRequest struct {
	ctx        context.Context
	capability capabilities.Capability
	input      any
}

// Client provides the interface for executing LLM protocol requests.
// It orchestrates the flow from capability selection through HTTP execution
// to response processing, with option management and health tracking.
type Client interface {
	// Provider returns the provider instance managed by this client.
	Provider() providers.Provider

	// Model returns the model instance managed by this client.
	Model() models.Model

	// HTTPClient returns a configured HTTP client for this transport.
	// Creates a new client on each call with timeout and connection pool settings.
	HTTPClient() *http.Client

	// ExecuteProtocol executes a protocol request and returns the parsed response.
	// Performs option merging, validation, and complete request/response flow.
	// Returns an error if protocol is not supported, options are invalid, or request fails.
	ExecuteProtocol(ctx context.Context, req *capabilities.CapabilityRequest) (any, error)

	// ExecuteProtocolStream executes a streaming protocol request and returns a channel of chunks.
	// The channel is closed when streaming completes or context is cancelled.
	// Returns an error if protocol is not supported, doesn't support streaming, or request fails.
	ExecuteProtocolStream(ctx context.Context, req *capabilities.CapabilityRequest) (<-chan protocols.StreamingChunk, error)

	// IsHealthy returns the current health status of the client.
	// Set to false after request failures, true after successful requests.
	// Thread-safe for concurrent access.
	IsHealthy() bool
}

// client implements the Client interface with provider and model orchestration.
type client struct {
	provider providers.Provider
	model    models.Model
	config   *config.TransportConfig

	mutex      sync.RWMutex
	healthy    bool
	lastHealth time.Time
}

// New creates a new transport Client from configuration.
// Creates the provider from configuration and initializes health tracking.
// Returns an error if provider creation fails.
func New(config *config.TransportConfig) (Client, error) {
	provider, err := providers.Create(config.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	return &client{
		provider:   provider,
		model:      provider.Model(),
		config:     config,
		healthy:    true,
		lastHealth: time.Now(),
	}, nil
}

// Provider returns the provider instance managed by this client.
func (c *client) Provider() providers.Provider {
	return c.provider
}

// Model returns the model instance managed by this client.
func (c *client) Model() models.Model {
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
// Merges request options with model defaults, validates options, and executes the complete request flow.
// Returns the parsed response or an error if any step fails.
func (c *client) ExecuteProtocol(ctx context.Context, req *capabilities.CapabilityRequest) (any, error) {
	capability, err := c.model.GetCapability(req.Protocol)
	if err != nil {
		return nil, fmt.Errorf("capability selection failed: %w", err)
	}

	options := c.model.MergeRequestOptions(req.Protocol, req.Options)

	if err := capability.ValidateOptions(options); err != nil {
		return nil, fmt.Errorf("invalid options for %s protocol: %w", req.Protocol, err)
	}

	request := &capabilities.CapabilityRequest{
		Protocol: req.Protocol,
		Messages: req.Messages,
		Options:  options,
	}

	return c.execute(ctx, capability, request)
}

// ExecuteProtocolStream executes a streaming protocol request.
// Verifies capability supports streaming, merges options, validates, and executes streaming flow.
// Returns a channel of streaming chunks that closes when stream completes or context is cancelled.
// Returns an error if capability doesn't support streaming or initial request setup fails.
func (c *client) ExecuteProtocolStream(ctx context.Context, req *capabilities.CapabilityRequest) (<-chan protocols.StreamingChunk, error) {
	capability, err := c.provider.Model().GetCapability(req.Protocol)
	if err != nil {
		return nil, fmt.Errorf("capability selection failed: %w", err)
	}

	streaming, ok := capability.(capabilities.StreamingCapability)
	if !ok {
		return nil, fmt.Errorf("capability %s does not support streaming", capability.Name())
	}

	options := c.model.MergeRequestOptions(req.Protocol, req.Options)

	if err := capability.ValidateOptions(options); err != nil {
		return nil, fmt.Errorf("invalid options for %s protocol: %w", req.Protocol, err)
	}

	request := &capabilities.CapabilityRequest{
		Protocol: req.Protocol,
		Messages: req.Messages,
		Options:  options,
	}

	return c.executeStream(ctx, streaming, request)
}

// IsHealthy returns the current health status.
// Thread-safe for concurrent access via read mutex.
func (c *client) IsHealthy() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.healthy
}

// execute performs the complete execution flow for standard requests.
// Creates capability request, prepares provider request, executes HTTP request,
// processes response, and updates health status.
func (c *client) execute(ctx context.Context, capability capabilities.Capability, req *capabilities.CapabilityRequest) (any, error) {
	capRequest, err := capability.CreateRequest(req, c.model.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	providerRequest, err := c.provider.PrepareRequest(ctx, req.Protocol, capRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		providerRequest.URL,
		bytes.NewBuffer(providerRequest.Body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	for key, value := range providerRequest.Headers {
		httpReq.Header.Set(key, value)
	}

	c.provider.SetHeaders(httpReq)

	httpClient := c.HTTPClient()
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		c.setHealthy(false)
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	result, err := c.provider.ProcessResponse(resp, capability)
	if err != nil {
		c.setHealthy(false)
		return nil, err
	}

	c.setHealthy(true)
	return result, nil
}

// executeStream performs the complete execution flow for streaming requests.
// Creates streaming request, prepares provider request, executes HTTP request,
// processes streaming response, and manages goroutine for chunk forwarding.
// Updates health status based on stream success/failure.
func (c *client) executeStream(ctx context.Context, capability capabilities.StreamingCapability, req *capabilities.CapabilityRequest) (<-chan protocols.StreamingChunk, error) {
	capRequest, err := capability.CreateStreamingRequest(req, c.model.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to create streaming request: %w", err)
	}

	providerRequest, err := c.provider.PrepareStreamRequest(ctx, req.Protocol, capRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare streaming request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		providerRequest.URL,
		bytes.NewBuffer(providerRequest.Body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	for key, value := range providerRequest.Headers {
		httpReq.Header.Set(key, value)
	}

	c.provider.SetHeaders(httpReq)

	httpClient := c.HTTPClient()
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		c.setHealthy(false)
		return nil, fmt.Errorf("streaming request failed: %w", err)
	}

	stream, err := c.provider.ProcessStreamResponse(ctx, resp, capability)
	if err != nil {
		c.setHealthy(false)
		resp.Body.Close()
		return nil, err
	}

	output := make(chan protocols.StreamingChunk)
	go func() {
		defer close(output)
		defer resp.Body.Close()

		for data := range stream {
			if chunk, ok := data.(*protocols.StreamingChunk); ok {
				output <- *chunk
			}
		}
		c.setHealthy(true)
	}()

	return output, nil
}

// setHealthy updates the health status with timestamp.
// Thread-safe via write mutex.
func (c *client) setHealthy(healthy bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.healthy = healthy
	c.lastHealth = time.Now()
}
