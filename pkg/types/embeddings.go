package types

import (
	"encoding/json"
	"fmt"

	"maps"
)

// EmbeddingsRequest represents an embeddings protocol request.
// Separates input text (protocol data) from model configuration options.
// Does not use messages array - input is the primary data field.
type EmbeddingsRequest struct {
	Input   any            // string or []string for batch embeddings
	Options map[string]any // Model configuration options
}

// GetProtocol returns the Embeddings protocol identifier.
func (r *EmbeddingsRequest) GetProtocol() Protocol {
	return Embeddings
}

// GetHeaders returns the HTTP headers for an embeddings request.
func (r *EmbeddingsRequest) GetHeaders() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

// Marshal converts the embeddings request to JSON.
// Combines input and options at the root level:
//
//	{
//	  "input": "text to embed",
//	  "encoding_format": "float"
//	}
//
// Note: Providers may transform input to their specific format
// (e.g., Google uses {"content": {"parts": [{"text": "..."}]}}).
func (r *EmbeddingsRequest) Marshal() ([]byte, error) {
	combined := make(map[string]any)
	combined["input"] = r.Input
	maps.Copy(combined, r.Options)
	return json.Marshal(combined)
}

// ParseEmbeddingsResponse parses an embeddings response from JSON.
// Returns the parsed EmbeddingsResponse or an error if parsing fails.
func ParseEmbeddingsResponse(body []byte) (*EmbeddingsResponse, error) {
	var response EmbeddingsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse embeddings response: %w", err)
	}
	return &response, nil
}
