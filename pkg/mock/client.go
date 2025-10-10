package mock

import (
	"context"
	"net/http"
	"time"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/models"
	"github.com/JaimeStill/go-agents/pkg/protocols"
	"github.com/JaimeStill/go-agents/pkg/providers"
	"github.com/JaimeStill/go-agents/pkg/transport"
)

// MockClient implements transport.Client interface for testing.
type MockClient struct {
	provider providers.Provider
	model    models.Model
	healthy  bool

	// Configurable responses
	executeResponse      any
	executeError         error
	streamChunks         []protocols.StreamingChunk
	streamError          error
	httpClient           *http.Client
	httpClientConfigured bool
}

// NewMockClient creates a new MockClient with default configuration.
func NewMockClient(opts ...MockClientOption) *MockClient {
	m := &MockClient{
		provider: NewMockProvider(),
		model:    NewMockModel(),
		healthy:  true,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Set model on provider
	if mockProv, ok := m.provider.(*MockProvider); ok {
		mockProv.model = m.model
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// MockClientOption configures a MockClient.
type MockClientOption func(*MockClient)

// WithMockProvider sets a custom provider.
func WithMockProvider(provider providers.Provider) MockClientOption {
	return func(m *MockClient) {
		m.provider = provider
	}
}

// WithMockModel sets a custom model.
func WithMockModel(model models.Model) MockClientOption {
	return func(m *MockClient) {
		m.model = model
	}
}

// WithExecuteResponse sets the response for ExecuteProtocol.
func WithExecuteResponse(response any, err error) MockClientOption {
	return func(m *MockClient) {
		m.executeResponse = response
		m.executeError = err
	}
}

// WithStreamResponse sets the chunks for ExecuteProtocolStream.
func WithStreamResponse(chunks []protocols.StreamingChunk, err error) MockClientOption {
	return func(m *MockClient) {
		m.streamChunks = chunks
		m.streamError = err
	}
}

// WithHealthy sets the health status.
func WithHealthy(healthy bool) MockClientOption {
	return func(m *MockClient) {
		m.healthy = healthy
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) MockClientOption {
	return func(m *MockClient) {
		m.httpClient = client
		m.httpClientConfigured = true
	}
}

// Provider returns the mock provider.
func (m *MockClient) Provider() providers.Provider {
	return m.provider
}

// Model returns the mock model.
func (m *MockClient) Model() models.Model {
	return m.model
}

// HTTPClient returns the configured HTTP client.
func (m *MockClient) HTTPClient() *http.Client {
	return m.httpClient
}

// ExecuteProtocol returns the predetermined response.
func (m *MockClient) ExecuteProtocol(ctx context.Context, req *capabilities.CapabilityRequest) (any, error) {
	return m.executeResponse, m.executeError
}

// ExecuteProtocolStream returns a channel with predetermined chunks.
func (m *MockClient) ExecuteProtocolStream(ctx context.Context, req *capabilities.CapabilityRequest) (<-chan protocols.StreamingChunk, error) {
	if m.streamError != nil {
		return nil, m.streamError
	}

	ch := make(chan protocols.StreamingChunk, len(m.streamChunks))
	for _, chunk := range m.streamChunks {
		ch <- chunk
	}
	close(ch)

	return ch, nil
}

// IsHealthy returns the mock health status.
func (m *MockClient) IsHealthy() bool {
	return m.healthy
}

// Verify MockClient implements transport.Client interface.
var _ transport.Client = (*MockClient)(nil)
