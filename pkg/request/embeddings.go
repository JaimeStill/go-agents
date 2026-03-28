package request

import (
	"github.com/JaimeStill/go-agents/pkg/format"
	"github.com/JaimeStill/go-agents/pkg/model"
	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/providers"
)

// EmbeddingsRequest represents an embeddings protocol request.
// Separates input text (protocol data) from model configuration options.
// Does not use messages array - input is the primary data field.
type EmbeddingsRequest struct {
	provider providers.Provider
	fmt      format.Format
	model    *model.Model
	input    any // string or []string for batch embeddings
	options  map[string]any
}

// NewEmbeddings creates a new EmbeddingsRequest with the given components.
// Input is the text to embed (string or []string for batch).
// Options specify model configuration (encoding_format, dimensions, etc.).
func NewEmbeddings(
	p providers.Provider,
	f format.Format,
	m *model.Model,
	input any,
	opts map[string]any,
) *EmbeddingsRequest {
	return &EmbeddingsRequest{
		provider: p,
		fmt:      f,
		model:    m,
		input:    input,
		options:  opts,
	}
}

// Protocol returns the Embeddings protocol identifier.
func (r *EmbeddingsRequest) Protocol() protocol.Protocol {
	return protocol.Embeddings
}

// Headers returns the HTTP headers for an embeddings request.
func (r *EmbeddingsRequest) Headers() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

// Marshal delegates to the provider for provider-specific JSON formatting.
func (r *EmbeddingsRequest) Marshal() ([]byte, error) {
	return r.fmt.Marshal(protocol.Embeddings, &format.EmbeddingsData{
		Model:   r.model.Name,
		Input:   r.input,
		Options: r.options,
	})
}

// Provider returns the provider for this request.
func (r *EmbeddingsRequest) Provider() providers.Provider {
	return r.provider
}

func (r *EmbeddingsRequest) Format() format.Format {
	return r.fmt
}

// Model returns the model for this request.
func (r *EmbeddingsRequest) Model() *model.Model {
	return r.model
}
