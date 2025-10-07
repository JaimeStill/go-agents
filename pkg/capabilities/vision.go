package capabilities

import (
	"encoding/json"
	"fmt"

	"github.com/JaimeStill/go-agents/pkg/protocols"
)

// VisionCapability implements the vision protocol with streaming support.
// Handles multimodal inputs combining text prompts with images.
// Transforms image URLs into the structured content format required by vision models.
type VisionCapability struct {
	*StandardStreamingCapability
}

// NewVisionCapability creates a new VisionCapability with the specified options.
// Required options typically include "images". Optional options include "detail" for image quality.
func NewVisionCapability(name string, options []CapabilityOption) *VisionCapability {
	return &VisionCapability{
		StandardStreamingCapability: NewStandardStreamingCapability(
			name,
			protocols.Vision,
			options,
		),
	}
}

// CreateRequest creates a protocol request for non-streaming vision.
// Processes images and transforms the last user message to include structured image content.
func (c *VisionCapability) CreateRequest(req *CapabilityRequest, model string) (*protocols.Request, error) {
	options, err := c.ProcessOptions(req.Options)
	if err != nil {
		return nil, err
	}

	messages, err := c.ProcessImages(req.Messages, options)
	if err != nil {
		return nil, err
	}

	options["model"] = model

	return &protocols.Request{
		Messages: messages,
		Options:  options,
	}, nil
}

// CreateStreamingRequest creates a protocol request for streaming vision.
// Processes images and sets the stream option to true.
func (c *VisionCapability) CreateStreamingRequest(req *CapabilityRequest, model string) (*protocols.Request, error) {
	options, err := c.ProcessOptions(req.Options)
	if err != nil {
		return nil, err
	}

	messages, err := c.ProcessImages(req.Messages, options)
	if err != nil {
		return nil, err
	}

	options["model"] = model
	options["stream"] = true

	return &protocols.Request{
		Messages: messages,
		Options:  options,
	}, nil
}

// ProcessImages transforms the last user message to include image content.
// Takes image URLs from options and creates structured content with text and images.
// The detail option controls image quality (low, high, auto).
// Removes images and detail from options after embedding them in message content.
func (c *VisionCapability) ProcessImages(messages []protocols.Message, options map[string]any) ([]protocols.Message, error) {
	images, ok := options["images"].([]any)
	if !ok || len(images) == 0 {
		return nil, fmt.Errorf("images must be a non-empty array")
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("messages cannot be empty for vision requests")
	}

	idx := len(messages) - 1
	message := &messages[idx]

	if message.Role != "user" {
		return nil, fmt.Errorf("last message must be from user for vision requests")
	}

	content := []map[string]any{
		{"type": "text", "text": message.Content},
	}

	for _, img := range images {
		if imgStr, ok := img.(string); ok {
			detail := protocols.ExtractOption(options, "detail", "auto")
			content = append(content, map[string]any{
				"type": "image_url",
				"image_url": map[string]any{
					"url":    imgStr,
					"detail": detail,
				},
			})
		}
	}

	messages[idx] = protocols.Message{
		Role:    message.Role,
		Content: content,
	}

	// Remove vision-specific options that are now embedded in message content
	// These should not be sent as top-level request parameters
	delete(options, "images")
	delete(options, "detail")

	return messages, nil
}

// ParseResponse parses a non-streaming vision response.
// Returns a ChatResponse with the model's analysis of the images.
func (c *VisionCapability) ParseResponse(data []byte) (any, error) {
	var response protocols.ChatResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
