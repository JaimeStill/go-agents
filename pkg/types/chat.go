package types

import (
	"encoding/json"
	"fmt"

	"maps"
)

// ChatRequest represents a chat protocol request.
// Separates conversation messages from model configuration options.
type ChatRequest struct {
	Messages []Message
	Options  map[string]any
}

// GetProtocol returns the Chat protocol identifier.
func (r *ChatRequest) GetProtocol() Protocol {
	return Chat
}

// GetHeaders returns the HTTP headers for a chat request.
func (r *ChatRequest) GetHeaders() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

// Marshal converts the chat request to JSON.
// Combines messages and options at the root level:
//
//	{
//	  "messages": [...],
//	  "temperature": 0.7,
//	  "max_tokens": 4096
//	}
func (r *ChatRequest) Marshal() ([]byte, error) {
	combined := make(map[string]any)
	combined["messages"] = r.Messages
	maps.Copy(combined, r.Options)
	return json.Marshal(combined)
}

func ParseChatResponse(body []byte) (*ChatResponse, error) {
	var response ChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse chat response: %w", err)
	}
	return &response, nil
}

func ParseChatStreamChunk(data []byte) (*StreamingChunk, error) {
	var chunk StreamingChunk
	if err := json.Unmarshal(data, &chunk); err != nil {
		return nil, fmt.Errorf("failed to parse streaming chunk: %w", err)
	}
	return &chunk, nil
}
