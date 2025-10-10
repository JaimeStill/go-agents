package mock

import (
	"context"

	"github.com/JaimeStill/go-agents/pkg/agent"
	"github.com/JaimeStill/go-agents/pkg/models"
	"github.com/JaimeStill/go-agents/pkg/protocols"
	"github.com/JaimeStill/go-agents/pkg/providers"
	"github.com/JaimeStill/go-agents/pkg/transport"
)

// MockAgent implements agent.Agent interface for testing.
// All methods return predetermined responses configured during construction.
type MockAgent struct {
	id string

	// Protocol responses
	chatResponse       *protocols.ChatResponse
	chatError          error
	visionResponse     *protocols.ChatResponse
	visionError        error
	toolsResponse      *protocols.ToolsResponse
	toolsError         error
	embeddingsResponse *protocols.EmbeddingsResponse
	embeddingsError    error

	// Streaming responses
	streamChunks []protocols.StreamingChunk
	streamError  error

	// Dependencies
	client   transport.Client
	provider providers.Provider
}

// NewMockAgent creates a new MockAgent with default configuration.
// Use option functions to configure specific behaviors.
func NewMockAgent(opts ...MockAgentOption) *MockAgent {
	m := &MockAgent{
		id:           "mock-agent-id",
		client:       NewMockClient(),
		provider:     NewMockProvider(),
		streamChunks: []protocols.StreamingChunk{},
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// MockAgentOption configures a MockAgent.
type MockAgentOption func(*MockAgent)

// WithID sets the agent ID.
func WithID(id string) MockAgentOption {
	return func(m *MockAgent) {
		m.id = id
	}
}

// WithChatResponse sets the chat response and error.
func WithChatResponse(response *protocols.ChatResponse, err error) MockAgentOption {
	return func(m *MockAgent) {
		m.chatResponse = response
		m.chatError = err
	}
}

// WithVisionResponse sets the vision response and error.
func WithVisionResponse(response *protocols.ChatResponse, err error) MockAgentOption {
	return func(m *MockAgent) {
		m.visionResponse = response
		m.visionError = err
	}
}

// WithToolsResponse sets the tools response and error.
func WithToolsResponse(response *protocols.ToolsResponse, err error) MockAgentOption {
	return func(m *MockAgent) {
		m.toolsResponse = response
		m.toolsError = err
	}
}

// WithEmbeddingsResponse sets the embeddings response and error.
func WithEmbeddingsResponse(response *protocols.EmbeddingsResponse, err error) MockAgentOption {
	return func(m *MockAgent) {
		m.embeddingsResponse = response
		m.embeddingsError = err
	}
}

// WithStreamChunks sets the streaming chunks for stream methods.
func WithStreamChunks(chunks []protocols.StreamingChunk, err error) MockAgentOption {
	return func(m *MockAgent) {
		m.streamChunks = chunks
		m.streamError = err
	}
}

// WithClient sets a custom transport client.
func WithClient(client transport.Client) MockAgentOption {
	return func(m *MockAgent) {
		m.client = client
	}
}

// WithProvider sets a custom provider.
func WithProvider(provider providers.Provider) MockAgentOption {
	return func(m *MockAgent) {
		m.provider = provider
	}
}

// ID returns the mock agent's unique identifier.
func (m *MockAgent) ID() string {
	return m.id
}

// Client returns the mock transport client.
func (m *MockAgent) Client() transport.Client {
	return m.client
}

// Provider returns the mock provider.
func (m *MockAgent) Provider() providers.Provider {
	return m.provider
}

// Model returns the mock model from the client.
func (m *MockAgent) Model() models.Model {
	if m.client != nil {
		return m.client.Model()
	}
	if m.provider != nil {
		return m.provider.Model()
	}
	return NewMockModel()
}

// Chat returns the predetermined chat response.
func (m *MockAgent) Chat(ctx context.Context, prompt string, opts ...map[string]any) (*protocols.ChatResponse, error) {
	return m.chatResponse, m.chatError
}

// ChatStream returns a channel with predetermined streaming chunks.
func (m *MockAgent) ChatStream(ctx context.Context, prompt string, opts ...map[string]any) (<-chan protocols.StreamingChunk, error) {
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

// Vision returns the predetermined vision response.
func (m *MockAgent) Vision(ctx context.Context, prompt string, images []string, opts ...map[string]any) (*protocols.ChatResponse, error) {
	return m.visionResponse, m.visionError
}

// VisionStream returns a channel with predetermined streaming chunks.
func (m *MockAgent) VisionStream(ctx context.Context, prompt string, images []string, opts ...map[string]any) (<-chan protocols.StreamingChunk, error) {
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

// Tools returns the predetermined tools response.
func (m *MockAgent) Tools(ctx context.Context, prompt string, tools []agent.Tool, opts ...map[string]any) (*protocols.ToolsResponse, error) {
	return m.toolsResponse, m.toolsError
}

// Embed returns the predetermined embeddings response.
func (m *MockAgent) Embed(ctx context.Context, input string, opts ...map[string]any) (*protocols.EmbeddingsResponse, error) {
	return m.embeddingsResponse, m.embeddingsError
}

// Verify MockAgent implements agent.Agent interface.
var _ agent.Agent = (*MockAgent)(nil)
