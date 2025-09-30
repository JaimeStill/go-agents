package protocols

import (
	"encoding/json"
	"fmt"
	"maps"
	"strings"
)

type Protocol string

const (
	Chat       Protocol = "chat"
	Vision     Protocol = "vision"
	Tools      Protocol = "tools"
	Embeddings Protocol = "embeddings"
	Audio      Protocol = "audio"
	Realtime   Protocol = "realtime"
)

type Request struct {
	Messages []Message
	Options  map[string]any
}

func (r *Request) Marshal() ([]byte, error) {
	combined := make(map[string]any)

	combined["messages"] = r.Messages
	maps.Copy(combined, r.Options)

	return json.Marshal(combined)
}

func (r *Request) GetHeaders() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

func NewMessage(role string, content any) Message {
	return Message{Role: role, Content: content}
}

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

func (c *StreamingChunk) Content() string {
	if len(c.Choices) > 0 {
		return c.Choices[0].Delta.Content
	}
	return ""
}

type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

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

type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

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

func IsValid(p string) bool {
	switch Protocol(p) {
	case Chat, Vision, Tools, Embeddings, Audio, Realtime:
		return true
	default:
		return false
	}
}

func ProtocolStrings() string {
	valid := ValidProtocols()
	strs := make([]string, len(valid))
	for i, p := range valid {
		strs[i] = string(p)
	}
	return strings.Join(strs, ", ")
}

func ValidProtocols() []Protocol {
	return []Protocol{
		Chat,
		Vision,
		Tools,
		Embeddings,
		Audio,
		Realtime,
	}
}
