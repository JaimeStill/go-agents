package request

import (
	"github.com/JaimeStill/go-agents/pkg/format"
	"github.com/JaimeStill/go-agents/pkg/model"
	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/providers"
)

// ChatRequest represents a chat protocol request.
// Encapsulates conversation messages, model configuration options,
// and the provider/model needed for execution.
type ChatRequest struct {
	provider providers.Provider
	fmt      format.Format
	model    *model.Model
	messages []protocol.Message
	options  map[string]any
}

// NewChat creates a new ChatRequest with the given components.
// Messages contain the conversation history.
// Options specify model configuration (temperature, max_tokens, etc.).
func NewChat(
	p providers.Provider,
	f format.Format,
	m *model.Model,
	messages []protocol.Message,
	opts map[string]any,
) *ChatRequest {
	return &ChatRequest{
		provider: p,
		fmt:      f,
		model:    m,
		messages: messages,
		options:  opts,
	}
}

// Protocol returns the Chat protocol identifier.
func (r *ChatRequest) Protocol() protocol.Protocol {
	return protocol.Chat
}

// Headers returns the HTTP headers for a chat request.
func (r *ChatRequest) Headers() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

// Marshal delegates to the provider for provider-specific JSON formatting.
func (r *ChatRequest) Marshal() ([]byte, error) {
	return r.fmt.Marshal(protocol.Chat, &format.ChatData{
		Model:    r.model.Name,
		Messages: r.messages,
		Options:  r.options,
	})
}

// Provider returns the provider for this request.
func (r *ChatRequest) Provider() providers.Provider {
	return r.provider
}

func (r *ChatRequest) Format() format.Format {
	return r.fmt
}

// Model returns the model for this request.
func (r *ChatRequest) Model() *model.Model {
	return r.model
}
