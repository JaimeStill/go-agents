package protocols

import (
	"encoding/json"
	"fmt"
	"maps"
	"strings"
)

// Protocol represents the type of LLM interaction operation.
// Each protocol defines a specific capability for model interaction.
type Protocol string

const (
	// Chat represents standard text-based conversation protocol.
	Chat Protocol = "chat"

	// Vision represents image understanding with multimodal inputs.
	Vision Protocol = "vision"

	// Tools represents function calling and tool execution protocol.
	Tools Protocol = "tools"

	// Embeddings represents text vectorization for semantic search.
	Embeddings Protocol = "embeddings"
)

// Request represents a protocol request with messages and options.
// The Messages field contains the conversation history, and Options
// contains protocol-specific parameters (e.g., temperature, max_tokens).
//
// When marshaled to JSON, options are merged at the root level alongside messages:
//
//	{
//	  "messages": [...],
//	  "temperature": 0.7,
//	  "max_tokens": 4096
//	}
type Request struct {
	Messages []Message
	Options  map[string]any
}

// Marshal converts the Request to JSON with messages and options combined at the root level.
// Returns the JSON bytes or an error if marshaling fails.
func (r *Request) Marshal() ([]byte, error) {
	combined := make(map[string]any)

	combined["messages"] = r.Messages
	maps.Copy(combined, r.Options)

	return json.Marshal(combined)
}

// GetHeaders returns the HTTP headers required for this request.
// Currently returns Content-Type: application/json for all protocols.
func (r *Request) GetHeaders() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

// Message represents a single message in a conversation.
// The Role indicates the message sender (user, assistant, system),
// and Content can be either a string for text or a structured object
// for multimodal content (e.g., vision protocol with images).
type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

// NewMessage creates a new Message with the specified role and content.
// Content can be a string for text or a structured object for multimodal inputs.
//
// Example:
//
//	msg := protocols.NewMessage("user", "Hello, world!")
//	visionMsg := protocols.NewMessage("user", map[string]any{"type": "image_url", "image_url": url})
func NewMessage(role string, content any) Message {
	return Message{Role: role, Content: content}
}

// ChatResponse represents the response from a non-streaming chat protocol request.
// Contains the model output, metadata, and optional token usage information.
type ChatResponse struct {
	ID      string `json:"id,omitempty"`
	Object  string `json:"object,omitempty"`
	Created int64  `json:"created,omitempty"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int     `json:"index"`
		Message Message `json:"message"`
		Delta   *struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta,omitempty"`
		FinishReason string `json:"finish_reason,omitempty"`
	} `json:"choices"`
	Usage *TokenUsage `json:"usage,omitempty"`
}

// Content extracts the text content from the first choice in the response.
// Handles both string content and structured content (e.g., vision responses).
// Returns empty string if there are no choices.
func (r *ChatResponse) Content() string {
	if len(r.Choices) > 0 {
		// Handle both string content and structured content
		switch v := r.Choices[0].Message.Content.(type) {
		case string:
			return v
		default:
			// For structured content, convert to string representation
			// This handles vision responses that might have complex content
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

// StreamingChunk represents a single chunk from a streaming protocol response.
// Each chunk contains incremental content in the Delta field and metadata.
// The Error field can be set during streaming to indicate processing errors.
type StreamingChunk struct {
	ID      string `json:"id,omitempty"`
	Object  string `json:"object,omitempty"`
	Created int64  `json:"created,omitempty"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Error error `json:"-"`
}

// Content extracts the incremental content from the delta in the first choice.
// Returns empty string if there are no choices or no content in the delta.
func (c *StreamingChunk) Content() string {
	if len(c.Choices) > 0 {
		return c.Choices[0].Delta.Content
	}
	return ""
}

// TokenUsage tracks token consumption for a request/response cycle.
// Provides counts for prompt tokens, completion tokens, and total tokens used.
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// EmbeddingsResponse represents the response from an embeddings protocol request.
// Contains vector embeddings for the input text along with metadata and token usage.
type EmbeddingsResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
		Object    string    `json:"object"`
	}
	Model string      `json:"model"`
	Usage *TokenUsage `json:"usage,omitempty"`
}

// ToolsResponse represents the response from a tools (function calling) protocol request.
// Contains function calls requested by the model along with metadata and token usage.
type ToolsResponse struct {
	ID      string      `json:"id,omitempty"`
	Object  string      `json:"object,omitempty"`
	Created int64       `json:"created,omitempty"`
	Model   string      `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role      string     `json:"role"`
			Content   string     `json:"content"`
			ToolCalls []ToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason,omitempty"`
	} `json:"choices"`
	Usage *TokenUsage `json:"usage,omitempty"`
}

// ToolCall represents a function call requested by the model.
// Contains the call ID, type, and function details.
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction contains the details of a function to be called.
// Name specifies the function name, and Arguments contains JSON-encoded parameters.
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ExtractOption retrieves a typed value from an options map with type safety.
// If the key exists and the value is of type T, it returns the value.
// Otherwise, it returns the provided default value.
//
// This function provides safe option extraction without panics or type assertion errors.
//
// Example:
//
//	temperature := protocols.ExtractOption(options, "temperature", 0.7)
//	maxTokens := protocols.ExtractOption(options, "max_tokens", 4096)
func ExtractOption[T any](options map[string]any, key string, defaultValue T) T {
	if options == nil {
		return defaultValue
	}
	if value, exists := options[key]; exists {
		if typed, ok := value.(T); ok {
			return typed
		}
	}
	return defaultValue
}

// IsValid checks if a protocol string is valid.
// Returns true if the protocol is one of: chat, vision, tools, embeddings.
// The check is case-sensitive.
func IsValid(p string) bool {
	switch Protocol(p) {
	case Chat, Vision, Tools, Embeddings:
		return true
	default:
		return false
	}
}

// ProtocolStrings returns a comma-separated string of all valid protocols.
// Useful for displaying available protocols in error messages or help text.
func ProtocolStrings() string {
	valid := ValidProtocols()
	strs := make([]string, len(valid))
	for i, p := range valid {
		strs[i] = string(p)
	}
	return strings.Join(strs, ", ")
}

// ValidProtocols returns a slice of all supported protocol values.
// Returns protocols in order: Chat, Vision, Tools, Embeddings.
func ValidProtocols() []Protocol {
	return []Protocol{
		Chat,
		Vision,
		Tools,
		Embeddings,
	}
}
