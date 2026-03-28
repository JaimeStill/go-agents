// Package format defines the wire format abstraction for marshaling and parsing
// LLM provider requests and responses across different API formats.
package format

import (
	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/response"
)

// Format defines the interface for marshaling requests and parsing responses
// in a provider-specific wire format such as OpenAI or AWS Converse.
type Format interface {
	// Name returns the identifier for this format (e.g., "openai", "converse").
	Name() string

	// Marshal serializes the given protocol-specific data into a JSON request body.
	Marshal(p protocol.Protocol, data any) ([]byte, error)

	// Parse deserializes a JSON response body into a protocol-specific response type.
	Parse(p protocol.Protocol, body []byte) (any, error)

	// ParseStreamChunk parses a single streaming event into a StreamingResponse.
	ParseStreamChunk(p protocol.Protocol, data []byte) (*response.StreamingResponse, error)
}
