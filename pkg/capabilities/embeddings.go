package capabilities

import (
	"encoding/json"

	"github.com/JaimeStill/go-agents/pkg/protocols"
)

type EmbeddingsCapability struct {
	*StandardCapability
}

func NewEmbeddingsCapability(name string, options []CapabilityOption) *EmbeddingsCapability {
	return &EmbeddingsCapability{
		StandardCapability: NewStandardCapability(
			name,
			protocols.Embeddings,
			options,
		),
	}
}

func (c *EmbeddingsCapability) CreateRequest(req *CapabilityRequest, model ModelInfo) (*protocols.Request, error) {
	options, err := c.ProcessOptions(req.Options)
	if err != nil {
		return nil, err
	}

	options["model"] = model.Name()

	return &protocols.Request{
		Messages: req.Messages,
		Options:  options,
	}, nil
}

func (c *EmbeddingsCapability) ParseResponse(data []byte) (any, error) {
	var response protocols.EmbeddingsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
