package capabilities

import (
	"encoding/json"

	"github.com/JaimeStill/go-agents/pkg/protocols"
)

type ChatCapability struct {
	*StandardStreamingCapability
	capabilityType string
}

func NewChatCapability(name string, options []CapabilityOption) *ChatCapability {
	return &ChatCapability{
		StandardStreamingCapability: NewStandardStreamingCapability(
			name,
			protocols.Chat,
			options,
		),
	}
}

func (c *ChatCapability) CreateRequest(req *CapabilityRequest, model string) (*protocols.Request, error) {
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

func (c *ChatCapability) CreateStreamingRequest(req *CapabilityRequest, model string) (*protocols.Request, error) {
	options, err := c.ProcessOptions(req.Options)
	if err != nil {
		return nil, err
	}

	options["model"] = model
	options["stream"] = true

	return &protocols.Request{
		Messages: req.Messages,
		Options:  options,
	}, nil
}

func (c *ChatCapability) ParseResponse(data []byte) (any, error) {
	var response protocols.ChatResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
