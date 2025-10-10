package mock

import (
	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

// MockCapability implements both capabilities.Capability and capabilities.StreamingCapability for testing.
// Provides configurable responses for all capability methods.
type MockCapability struct {
	name            string
	protocol        protocols.Protocol
	options         []capabilities.CapabilityOption
	supportsStream  bool

	// Configurable responses
	validateError      error
	processedOptions   map[string]any
	processError       error
	request            *protocols.Request
	createRequestError error
	response           any
	parseError         error
	streamChunk        *protocols.StreamingChunk
	parseChunkError    error
	streamComplete     bool
}

// NewMockCapability creates a new MockCapability with default configuration.
func NewMockCapability(opts ...MockCapabilityOption) *MockCapability {
	m := &MockCapability{
		name:             "mock-capability",
		protocol:         protocols.Chat,
		options:          []capabilities.CapabilityOption{},
		supportsStream:   true,
		processedOptions: make(map[string]any),
		request: &protocols.Request{
			Messages: []protocols.Message{},
			Options:  make(map[string]any),
		},
		streamChunk: createStreamingChunk("mock chunk"),
		streamComplete: false,
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// MockCapabilityOption configures a MockCapability.
type MockCapabilityOption func(*MockCapability)

// WithCapabilityName sets the capability name.
func WithCapabilityName(name string) MockCapabilityOption {
	return func(m *MockCapability) {
		m.name = name
	}
}

// WithCapabilityProtocol sets the protocol.
func WithCapabilityProtocol(protocol protocols.Protocol) MockCapabilityOption {
	return func(m *MockCapability) {
		m.protocol = protocol
	}
}

// WithCapabilityOptions sets the capability options.
func WithCapabilityOptions(options []capabilities.CapabilityOption) MockCapabilityOption {
	return func(m *MockCapability) {
		m.options = options
	}
}

// WithSupportsStreaming sets whether streaming is supported.
func WithSupportsStreaming(supports bool) MockCapabilityOption {
	return func(m *MockCapability) {
		m.supportsStream = supports
	}
}

// WithValidateError sets an error for ValidateOptions.
func WithValidateError(err error) MockCapabilityOption {
	return func(m *MockCapability) {
		m.validateError = err
	}
}

// WithProcessedOptions sets the processed options result.
func WithProcessedOptions(options map[string]any, err error) MockCapabilityOption {
	return func(m *MockCapability) {
		m.processedOptions = options
		m.processError = err
	}
}

// WithCreateRequest sets the request and error for CreateRequest.
func WithCreateRequest(request *protocols.Request, err error) MockCapabilityOption {
	return func(m *MockCapability) {
		m.request = request
		m.createRequestError = err
	}
}

// WithParseResponse sets the response and error for ParseResponse.
func WithParseResponse(response any, err error) MockCapabilityOption {
	return func(m *MockCapability) {
		m.response = response
		m.parseError = err
	}
}

// WithStreamChunk sets the chunk and error for ParseStreamingChunk.
func WithStreamChunk(chunk *protocols.StreamingChunk, err error) MockCapabilityOption {
	return func(m *MockCapability) {
		m.streamChunk = chunk
		m.parseChunkError = err
	}
}

// WithStreamComplete sets whether streaming is complete.
func WithStreamComplete(complete bool) MockCapabilityOption {
	return func(m *MockCapability) {
		m.streamComplete = complete
	}
}

// Name returns the capability name.
func (m *MockCapability) Name() string {
	return m.name
}

// Protocol returns the protocol.
func (m *MockCapability) Protocol() protocols.Protocol {
	return m.protocol
}

// Options returns the capability options.
func (m *MockCapability) Options() []capabilities.CapabilityOption {
	return m.options
}

// ValidateOptions returns the configured error.
func (m *MockCapability) ValidateOptions(options map[string]any) error {
	return m.validateError
}

// ProcessOptions returns the configured processed options.
func (m *MockCapability) ProcessOptions(options map[string]any) (map[string]any, error) {
	return m.processedOptions, m.processError
}

// CreateRequest returns the configured request.
func (m *MockCapability) CreateRequest(req *capabilities.CapabilityRequest, model string) (*protocols.Request, error) {
	return m.request, m.createRequestError
}

// ParseResponse returns the configured response.
func (m *MockCapability) ParseResponse(data []byte) (any, error) {
	return m.response, m.parseError
}

// SupportsStreaming returns whether streaming is supported.
func (m *MockCapability) SupportsStreaming() bool {
	return m.supportsStream
}

// CreateStreamingRequest returns the configured streaming request.
func (m *MockCapability) CreateStreamingRequest(req *capabilities.CapabilityRequest, model string) (*protocols.Request, error) {
	return m.request, m.createRequestError
}

// ParseStreamingChunk returns the configured chunk.
func (m *MockCapability) ParseStreamingChunk(data []byte) (*protocols.StreamingChunk, error) {
	return m.streamChunk, m.parseChunkError
}

// IsStreamComplete returns whether streaming is complete.
func (m *MockCapability) IsStreamComplete(data string) bool {
	return m.streamComplete
}

// Verify MockCapability implements both interfaces.
var _ capabilities.Capability = (*MockCapability)(nil)
var _ capabilities.StreamingCapability = (*MockCapability)(nil)

// createStreamingChunk is a helper to create a StreamingChunk with proper structure.
func createStreamingChunk(content string) *protocols.StreamingChunk {
	chunk := &protocols.StreamingChunk{
		Model: "mock-model",
	}
	chunk.Choices = append(chunk.Choices, struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	}{
		Index: 0,
		Delta: struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		}{
			Content: content,
		},
	})
	return chunk
}
