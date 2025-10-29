package mock

import (
	"context"

	"github.com/JaimeStill/go-agents/pkg/agent"
	"github.com/JaimeStill/go-agents/pkg/client"
	"github.com/JaimeStill/go-agents/pkg/providers"
	"github.com/JaimeStill/go-agents/pkg/types"
)

// MockAgent implements agent.Agent interface for testing.
// All methods return predetermined responses configured during construction.
type MockAgent struct {
	id string

	// Protocol responses
	chatResponse       *types.ChatResponse
	chatError          error
	visionResponse     *types.ChatResponse
	visionError        error
	toolsResponse      *types.ToolsResponse
	toolsError         error
	embeddingsResponse *types.EmbeddingsResponse
	embeddingsError    error

	// Streaming responses
	streamChunks []types.StreamingChunk
	streamError  error

	// Dependencies
	client   client.Client
	provider providers.Provider
}

// NewMockAgent creates a new MockAgent with default configuration.
// Use option functions to configure specific behaviors.
func NewMockAgent(opts ...MockAgentOption) *MockAgent {
	m := &MockAgent{
		id:           "mock-agent-id",
		client:       NewMockClient(),
		provider:     NewMockProvider(),
		streamChunks: []types.StreamingChunk{},
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
func WithChatResponse(response *types.ChatResponse, err error) MockAgentOption {
	return func(m *MockAgent) {
		m.chatResponse = response
		m.chatError = err
	}
}

// WithVisionResponse sets the vision response and error.
func WithVisionResponse(response *types.ChatResponse, err error) MockAgentOption {
	return func(m *MockAgent) {
		m.visionResponse = response
		m.visionError = err
	}
}

// WithToolsResponse sets the tools response and error.
func WithToolsResponse(response *types.ToolsResponse, err error) MockAgentOption {
	return func(m *MockAgent) {
		m.toolsResponse = response
		m.toolsError = err
	}
}

// WithEmbeddingsResponse sets the embeddings response and error.
func WithEmbeddingsResponse(response *types.EmbeddingsResponse, err error) MockAgentOption {
	return func(m *MockAgent) {
		m.embeddingsResponse = response
		m.embeddingsError = err
	}
}

// WithStreamChunks sets the streaming chunks for stream methods.
func WithStreamChunks(chunks []types.StreamingChunk, err error) MockAgentOption {
	return func(m *MockAgent) {
		m.streamChunks = chunks
		m.streamError = err
	}
}

// WithClient sets a custom client.
func WithClient(c client.Client) MockAgentOption {
	return func(m *MockAgent) {
		m.client = c
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

// Client returns the mock client.
func (m *MockAgent) Client() client.Client {
	return m.client
}

// Provider returns the mock provider.
func (m *MockAgent) Provider() providers.Provider {
	return m.provider
}

// Model returns the mock model from the provider.
func (m *MockAgent) Model() *types.Model {
	if m.provider != nil {
		return m.provider.Model()
	}
	// Return default model if no provider
	return &types.Model{
		Name:    "mock-model",
		Options: make(map[types.Protocol]map[string]any),
	}
}

// Chat returns the predetermined chat response.
func (m *MockAgent) Chat(ctx context.Context, prompt string, opts ...map[string]any) (*types.ChatResponse, error) {
	return m.chatResponse, m.chatError
}

// ChatStream returns a channel with predetermined streaming chunks.
func (m *MockAgent) ChatStream(ctx context.Context, prompt string, opts ...map[string]any) (<-chan *types.StreamingChunk, error) {
	if m.streamError != nil {
		return nil, m.streamError
	}

	ch := make(chan *types.StreamingChunk, len(m.streamChunks))
	for i := range m.streamChunks {
		ch <- &m.streamChunks[i]
	}
	close(ch)

	return ch, nil
}

// Vision returns the predetermined vision response.
func (m *MockAgent) Vision(ctx context.Context, prompt string, images []string, opts ...map[string]any) (*types.ChatResponse, error) {
	return m.visionResponse, m.visionError
}

// VisionStream returns a channel with predetermined streaming chunks.
func (m *MockAgent) VisionStream(ctx context.Context, prompt string, images []string, opts ...map[string]any) (<-chan *types.StreamingChunk, error) {
	if m.streamError != nil {
		return nil, m.streamError
	}

	ch := make(chan *types.StreamingChunk, len(m.streamChunks))
	for i := range m.streamChunks {
		ch <- &m.streamChunks[i]
	}
	close(ch)

	return ch, nil
}

// Tools returns the predetermined tools response.
func (m *MockAgent) Tools(ctx context.Context, prompt string, tools []agent.Tool, opts ...map[string]any) (*types.ToolsResponse, error) {
	return m.toolsResponse, m.toolsError
}

// Embed returns the predetermined embeddings response.
func (m *MockAgent) Embed(ctx context.Context, input string, opts ...map[string]any) (*types.EmbeddingsResponse, error) {
	return m.embeddingsResponse, m.embeddingsError
}

// Verify MockAgent implements agent.Agent interface.
var _ agent.Agent = (*MockAgent)(nil)
