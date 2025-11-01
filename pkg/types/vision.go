package types

import (
	"encoding/json"
	"fmt"

	"maps"
)

// VisionRequest represents a vision protocol request with image inputs.
// Separates images and vision-specific options from model configuration options.
type VisionRequest struct {
	Messages      []Message
	Images        []string           // URLs or base64 data URIs
	VisionOptions map[string]any     // Vision-specific options (e.g., detail: "high")
	Options       map[string]any     // Model configuration options
}

// GetProtocol returns the Vision protocol identifier.
func (r *VisionRequest) GetProtocol() Protocol {
	return Vision
}

// GetHeaders returns the HTTP headers for a vision request.
func (r *VisionRequest) GetHeaders() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

// Marshal converts the vision request to JSON.
// Transforms the last message to embed images in multimodal content structure:
//
//	{
//	  "messages": [{
//	    "role": "user",
//	    "content": [
//	      {"type": "text", "text": "What's in this image?"},
//	      {"type": "image_url", "image_url": {"url": "...", "detail": "high"}}
//	    ]
//	  }],
//	  "temperature": 0.7,
//	  "max_tokens": 4096
//	}
func (r *VisionRequest) Marshal() ([]byte, error) {
	if len(r.Messages) == 0 {
		return nil, fmt.Errorf("messages cannot be empty for vision requests")
	}

	if len(r.Images) == 0 {
		return nil, fmt.Errorf("images cannot be empty for vision requests")
	}

	// Transform the last message to embed images
	lastIdx := len(r.Messages) - 1
	message := r.Messages[lastIdx]

	var textContent string
	switch v := message.Content.(type) {
	case string:
		textContent = v
	default:
		return nil, fmt.Errorf("message content must be a string for vision transformation")
	}

	// Build structured content starting with text
	content := []map[string]any{
		{"type": "text", "text": textContent},
	}

	// Add each image with embedded options
	for _, imgURL := range r.Images {
		imageURL := map[string]any{
			"url": imgURL,
		}

		// Embed vision_options into image_url map
		if r.VisionOptions != nil {
			for key, value := range r.VisionOptions {
				imageURL[key] = value
			}
		}

		content = append(content, map[string]any{
			"type":      "image_url",
			"image_url": imageURL,
		})
	}

	// Create transformed messages
	transformedMessages := make([]Message, len(r.Messages))
	copy(transformedMessages, r.Messages)
	transformedMessages[lastIdx] = Message{
		Role:    message.Role,
		Content: content,
	}

	// Combine messages with options at root level
	combined := make(map[string]any)
	combined["messages"] = transformedMessages
	maps.Copy(combined, r.Options)

	return json.Marshal(combined)
}

// ParseVisionResponse parses a vision response from JSON.
// Vision protocol uses the same response format as chat.
// Returns the parsed ChatResponse or an error if parsing fails.
func ParseVisionResponse(body []byte) (*ChatResponse, error) {
	return ParseChatResponse(body)
}

// ParseVisionStreamChunk parses a streaming vision chunk from JSON.
// Vision protocol uses the same streaming format as chat.
// Returns the parsed StreamingChunk or an error if parsing fails.
func ParseVisionStreamChunk(data []byte) (*StreamingChunk, error) {
	return ParseChatStreamChunk(data)
}
