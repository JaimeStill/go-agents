package mock

import (
	"github.com/JaimeStill/go-agents/pkg/format"
	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/response"
)

// MockFormat implements format.Format for testing.
type MockFormat struct {
	name string

	marshalResponse  []byte
	marshalError     error
	parseResponse    any
	parseError       error
	streamChunk      *response.StreamingResponse
	streamChunkError error
}

// MockFormatOption configures a MockFormat.
type MockFormatOption func(*MockFormat)

// NewMockFormat creates a new MockFormat with default configuration.
func NewMockFormat(opts ...MockFormatOption) *MockFormat {
	m := &MockFormat{
		name: "mock-format",
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// WithFormatName sets the format name.
func WithFormatName(name string) MockFormatOption {
	return func(m *MockFormat) {
		m.name = name
	}
}

// WithFormatMarshalResponse sets the response for Marshal.
func WithFormatMarshalResponse(body []byte, err error) MockFormatOption {
	return func(m *MockFormat) {
		m.marshalResponse = body
		m.marshalError = err
	}
}

// WithFormatParseResponse sets the response for Parse.
func WithFormatParseResponse(resp any, err error) MockFormatOption {
	return func(m *MockFormat) {
		m.parseResponse = resp
		m.parseError = err
	}
}

// WithFormatStreamChunk sets the response for ParseStreamChunk.
func WithFormatStreamChunk(chunk *response.StreamingResponse, err error) MockFormatOption {
	return func(m *MockFormat) {
		m.streamChunk = chunk
		m.streamChunkError = err
	}
}

func (m *MockFormat) Name() string {
	return m.name
}

func (m *MockFormat) Marshal(proto protocol.Protocol, data any) ([]byte, error) {
	if m.marshalError != nil {
		return nil, m.marshalError
	}

	if m.marshalResponse != nil {
		return m.marshalResponse, nil
	}

	return []byte(`{}`), nil
}

func (m *MockFormat) Parse(proto protocol.Protocol, body []byte) (any, error) {
	if m.parseError != nil {
		return nil, m.parseError
	}

	if m.parseResponse != nil {
		return m.parseResponse, nil
	}

	return &response.Response{Role: "assistant"}, nil
}

func (m *MockFormat) ParseStreamChunk(proto protocol.Protocol, data []byte) (*response.StreamingResponse, error) {
	if m.streamChunkError != nil {
		return nil, m.streamChunkError
	}

	if m.streamChunk != nil {
		return m.streamChunk, nil
	}

	return &response.StreamingResponse{}, nil
}

// Verify MockFormat implements format.Format interface.
var _ format.Format = (*MockFormat)(nil)
