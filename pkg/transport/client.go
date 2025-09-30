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

type ClientRequest struct {
	ctx        context.Context
	capability capabilities.Capability
	input      any
}

type Client interface {
	Provider() providers.Provider
	Model() models.Model
	HTTPClient() *http.Client

	ExecuteProtocol(ctx context.Context, req *capabilities.CapabilityRequest) (any, error)
	ExecuteProtocolStream(ctx context.Context, req *capabilities.CapabilityRequest) (<-chan protocols.StreamingChunk, error)

	IsHealthy() bool
}

type client struct {
	provider providers.Provider
	model    models.Model
	config   *config.TransportConfig

	mutex      sync.RWMutex
	healthy    bool
	lastHealth time.Time
}

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

func (c *client) Provider() providers.Provider {
	return c.provider
}

func (c *client) Model() models.Model {
	return c.model
}

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

func (c *client) IsHealthy() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.healthy
}

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

func (c *client) setHealthy(healthy bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.healthy = healthy
	c.lastHealth = time.Now()
}
