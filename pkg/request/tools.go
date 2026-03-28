package request

import (
	"github.com/JaimeStill/go-agents/pkg/format"
	"github.com/JaimeStill/go-agents/pkg/model"
	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/providers"
)

// ToolsRequest represents a tools (function calling) protocol request.
// Separates tool definitions (protocol input data) from model configuration options.
type ToolsRequest struct {
	provider providers.Provider
	fmt      format.Format
	model    *model.Model
	messages []protocol.Message
	tools    []format.ToolDefinition
	options  map[string]any
}

// NewTools creates a new ToolsRequest with the given components.
// Messages contain the conversation history.
// Tools define the available functions the model can call.
// Options specify model configuration (temperature, max_tokens, etc.).
func NewTools(
	p providers.Provider,
	f format.Format,
	m *model.Model,
	messages []protocol.Message,
	tools []format.ToolDefinition,
	opts map[string]any,
) *ToolsRequest {
	return &ToolsRequest{
		provider: p,
		fmt:      f,
		model:    m,
		messages: messages,
		tools:    tools,
		options:  opts,
	}
}

// Protocol returns the Tools protocol identifier.
func (r *ToolsRequest) Protocol() protocol.Protocol {
	return protocol.Tools
}

// Headers returns the HTTP headers for a tools request.
func (r *ToolsRequest) Headers() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

// Marshal delegates to the provider for provider-specific JSON formatting.
// Different providers use different tool formats (OpenAI, Anthropic, Google).
func (r *ToolsRequest) Marshal() ([]byte, error) {
	return r.fmt.Marshal(protocol.Tools, &format.ToolsData{
		Model:    r.model.Name,
		Messages: r.messages,
		Tools:    r.tools,
		Options:  r.options,
	})
}

// Provider returns the provider for this request.
func (r *ToolsRequest) Provider() providers.Provider {
	return r.provider
}

func (r *ToolsRequest) Format() format.Format {
	return r.fmt
}

// Model returns the model for this request.
func (r *ToolsRequest) Model() *model.Model {
	return r.model
}
