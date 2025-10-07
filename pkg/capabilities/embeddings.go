package capabilities

import (
	"encoding/json"

	"github.com/JaimeStill/go-agents/pkg/protocols"
)

// EmbeddingsCapability implements the embeddings protocol.
// Converts text into vector representations for semantic search and similarity comparisons.
// Does not support streaming as embeddings are generated for complete input.
type EmbeddingsCapability struct {
	*StandardCapability
}

// NewEmbeddingsCapability creates a new EmbeddingsCapability with the specified options.
// Required options typically include "input" (text to embed).
// Optional options include "dimensions" and "encoding_format".
func NewEmbeddingsCapability(name string, options []CapabilityOption) *EmbeddingsCapability {
	return &EmbeddingsCapability{
		StandardCapability: NewStandardCapability(
			name,
			protocols.Embeddings,
			options,
		),
	}
}

// CreateRequest creates a protocol request for generating embeddings.
// Processes options and adds the model parameter.
func (c *EmbeddingsCapability) CreateRequest(req *CapabilityRequest, model string) (*protocols.Request, error) {
	options, err := c.ProcessOptions(req.Options)
	if err != nil {
		return nil, err
	}

	options["model"] = model

	return &protocols.Request{
		Messages: req.Messages,
		Options:  options,
	}, nil
}

// ParseResponse parses an embeddings response.
// Returns an EmbeddingsResponse containing vector representations of the input.
func (c *EmbeddingsCapability) ParseResponse(data []byte) (any, error) {
	var response protocols.EmbeddingsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
