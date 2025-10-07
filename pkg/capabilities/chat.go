package capabilities

import (
	"encoding/json"

	"github.com/JaimeStill/go-agents/pkg/protocols"
)

// ChatCapability implements the chat protocol with streaming support.
// Supports standard text-based conversations with configurable options
// like temperature, max_tokens, and other model parameters.
type ChatCapability struct {
	*StandardStreamingCapability
	capabilityType string
}

// NewChatCapability creates a new ChatCapability with the specified options.
// Common options include temperature, max_tokens, top_p, and frequency_penalty.
func NewChatCapability(name string, options []CapabilityOption) *ChatCapability {
	return &ChatCapability{
		StandardStreamingCapability: NewStandardStreamingCapability(
			name,
			protocols.Chat,
			options,
		),
	}
}

// CreateRequest creates a protocol request for non-streaming chat.
// Processes options, validates them, and adds the model parameter.
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

// CreateStreamingRequest creates a protocol request for streaming chat.
// Sets the stream option to true and processes other options.
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

// ParseResponse parses a non-streaming chat response.
// Returns a ChatResponse containing the model's reply and metadata.
func (c *ChatCapability) ParseResponse(data []byte) (any, error) {
	var response protocols.ChatResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
