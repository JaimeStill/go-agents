package types_test

import (
	"encoding/json"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/types"
)

func TestProtocol_Constants(t *testing.T) {
	tests := []struct {
		name     string
		protocol types.Protocol
		expected string
	}{
		{
			name:     "Chat",
			protocol: types.Chat,
			expected: "chat",
		},
		{
			name:     "Vision",
			protocol: types.Vision,
			expected: "vision",
		},
		{
			name:     "Tools",
			protocol: types.Tools,
			expected: "tools",
		},
		{
			name:     "Embeddings",
			protocol: types.Embeddings,
			expected: "embeddings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.protocol) != tt.expected {
				t.Errorf("got %s, want %s", string(tt.protocol), tt.expected)
			}
		})
	}
}

func TestIsValid_ValidProtocols(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
		expected bool
	}{
		{
			name:     "chat",
			protocol: "chat",
			expected: true,
		},
		{
			name:     "vision",
			protocol: "vision",
			expected: true,
		},
		{
			name:     "tools",
			protocol: "tools",
			expected: true,
		},
		{
			name:     "embeddings",
			protocol: "embeddings",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := types.IsValid(tt.protocol)
			if result != tt.expected {
				t.Errorf("IsValid(%q) = %v, want %v", tt.protocol, result, tt.expected)
			}
		})
	}
}

func TestIsValid_InvalidProtocols(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
	}{
		{
			name:     "invalid",
			protocol: "invalid",
		},
		{
			name:     "empty string",
			protocol: "",
		},
		{
			name:     "uppercase",
			protocol: "CHAT",
		},
		{
			name:     "mixed case",
			protocol: "Chat",
		},
		{
			name:     "unknown",
			protocol: "streaming",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := types.IsValid(tt.protocol)
			if result {
				t.Errorf("IsValid(%q) = true, want false", tt.protocol)
			}
		})
	}
}

func TestValidProtocols(t *testing.T) {
	result := types.ValidProtocols()

	expected := []types.Protocol{
		types.Chat,
		types.Vision,
		types.Tools,
		types.Embeddings,
	}

	if len(result) != len(expected) {
		t.Fatalf("got %d protocols, want %d", len(result), len(expected))
	}

	for i, protocol := range expected {
		if result[i] != protocol {
			t.Errorf("index %d: got %s, want %s", i, result[i], protocol)
		}
	}
}

func TestProtocolStrings(t *testing.T) {
	result := types.ProtocolStrings()
	expected := "chat, vision, tools, embeddings"

	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestNewMessage_StringContent(t *testing.T) {
	msg := types.NewMessage("user", "Hello, world!")

	if msg.Role != "user" {
		t.Errorf("got role %q, want %q", msg.Role, "user")
	}

	if content, ok := msg.Content.(string); !ok {
		t.Errorf("content is not a string")
	} else if content != "Hello, world!" {
		t.Errorf("got content %q, want %q", content, "Hello, world!")
	}
}

func TestNewMessage_StructuredContent(t *testing.T) {
	content := map[string]any{
		"type": "text",
		"text": "Hello",
	}

	msg := types.NewMessage("assistant", content)

	if msg.Role != "assistant" {
		t.Errorf("got role %q, want %q", msg.Role, "assistant")
	}

	if _, ok := msg.Content.(map[string]any); !ok {
		t.Errorf("content is not a map")
	}
}

func TestNewMessage_Roles(t *testing.T) {
	tests := []struct {
		name string
		role string
	}{
		{
			name: "user",
			role: "user",
		},
		{
			name: "assistant",
			role: "assistant",
		},
		{
			name: "system",
			role: "system",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := types.NewMessage(tt.role, "content")
			if msg.Role != tt.role {
				t.Errorf("got role %q, want %q", msg.Role, tt.role)
			}
		})
	}
}

func TestRequest_Marshal_MessagesOnly(t *testing.T) {
	req := &types.Request{
		Messages: []types.Message{
			types.NewMessage("user", "Hello"),
		},
		Options: nil,
	}

	data, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if _, exists := result["messages"]; !exists {
		t.Error("messages field missing from marshaled JSON")
	}
}

func TestRequest_Marshal_MessagesWithOptions(t *testing.T) {
	req := &types.Request{
		Messages: []types.Message{
			types.NewMessage("user", "Hello"),
		},
		Options: map[string]any{
			"temperature": 0.7,
			"max_tokens":  4096,
		},
	}

	data, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if _, exists := result["messages"]; !exists {
		t.Error("messages field missing from marshaled JSON")
	}

	if temp, exists := result["temperature"]; !exists {
		t.Error("temperature option missing from marshaled JSON")
	} else if temp != 0.7 {
		t.Errorf("got temperature %v, want 0.7", temp)
	}

	if tokens, exists := result["max_tokens"]; !exists {
		t.Error("max_tokens option missing from marshaled JSON")
	} else if tokens != float64(4096) {
		t.Errorf("got max_tokens %v, want 4096", tokens)
	}
}

func TestRequest_GetHeaders(t *testing.T) {
	req := &types.Request{}
	headers := req.GetHeaders()

	if contentType, exists := headers["Content-Type"]; !exists {
		t.Error("Content-Type header missing")
	} else if contentType != "application/json" {
		t.Errorf("got Content-Type %q, want %q", contentType, "application/json")
	}
}

func TestChatResponse_Content_StringContent(t *testing.T) {
	jsonData := `{
		"model": "gpt-4",
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": "Hello, world!"
			}
		}]
	}`

	var response types.ChatResponse
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	content := response.Content()
	if content != "Hello, world!" {
		t.Errorf("got content %q, want %q", content, "Hello, world!")
	}
}

func TestChatResponse_Content_StructuredContent(t *testing.T) {
	jsonData := `{
		"model": "gpt-4",
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": {"type": "text", "text": "Hello"}
			}
		}]
	}`

	var response types.ChatResponse
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	content := response.Content()
	if content == "" {
		t.Error("expected non-empty content for structured content")
	}
}

func TestChatResponse_Content_EmptyChoices(t *testing.T) {
	jsonData := `{
		"model": "gpt-4",
		"choices": []
	}`

	var response types.ChatResponse
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	content := response.Content()
	if content != "" {
		t.Errorf("got content %q, want empty string", content)
	}
}

func TestStreamingChunk_Content(t *testing.T) {
	jsonData := `{
		"model": "gpt-4",
		"choices": [{
			"index": 0,
			"delta": {
				"content": "Hello"
			}
		}]
	}`

	var chunk types.StreamingChunk
	if err := json.Unmarshal([]byte(jsonData), &chunk); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	content := chunk.Content()
	if content != "Hello" {
		t.Errorf("got content %q, want %q", content, "Hello")
	}
}

func TestStreamingChunk_Content_EmptyChoices(t *testing.T) {
	jsonData := `{
		"model": "gpt-4",
		"choices": []
	}`

	var chunk types.StreamingChunk
	if err := json.Unmarshal([]byte(jsonData), &chunk); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	content := chunk.Content()
	if content != "" {
		t.Errorf("got content %q, want empty string", content)
	}
}

func TestStreamingChunk_Content_EmptyDelta(t *testing.T) {
	jsonData := `{
		"model": "gpt-4",
		"choices": [{
			"index": 0,
			"delta": {
				"content": ""
			}
		}]
	}`

	var chunk types.StreamingChunk
	if err := json.Unmarshal([]byte(jsonData), &chunk); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	content := chunk.Content()
	if content != "" {
		t.Errorf("got content %q, want empty string", content)
	}
}

func TestExtractOption_ExistsWithCorrectType(t *testing.T) {
	options := map[string]any{
		"temperature": 0.7,
		"max_tokens":  4096,
		"model":       "gpt-4",
		"stream":      true,
	}

	tests := []struct {
		name         string
		key          string
		defaultValue any
		expected     any
	}{
		{
			name:         "float64",
			key:          "temperature",
			defaultValue: 0.5,
			expected:     0.7,
		},
		{
			name:         "int",
			key:          "max_tokens",
			defaultValue: 1000,
			expected:     4096,
		},
		{
			name:         "string",
			key:          "model",
			defaultValue: "default",
			expected:     "gpt-4",
		},
		{
			name:         "bool",
			key:          "stream",
			defaultValue: false,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch expected := tt.expected.(type) {
			case float64:
				result := types.ExtractOption(options, tt.key, tt.defaultValue.(float64))
				if result != expected {
					t.Errorf("got %v, want %v", result, expected)
				}
			case int:
				result := types.ExtractOption(options, tt.key, tt.defaultValue.(int))
				if result != expected {
					t.Errorf("got %v, want %v", result, expected)
				}
			case string:
				result := types.ExtractOption(options, tt.key, tt.defaultValue.(string))
				if result != expected {
					t.Errorf("got %v, want %v", result, expected)
				}
			case bool:
				result := types.ExtractOption(options, tt.key, tt.defaultValue.(bool))
				if result != expected {
					t.Errorf("got %v, want %v", result, expected)
				}
			}
		})
	}
}

func TestExtractOption_ExistsWithWrongType(t *testing.T) {
	options := map[string]any{
		"temperature": "0.7", // string instead of float64
	}

	result := types.ExtractOption(options, "temperature", 0.5)
	if result != 0.5 {
		t.Errorf("expected default value 0.5, got %v", result)
	}
}

func TestExtractOption_DoesNotExist(t *testing.T) {
	options := map[string]any{
		"temperature": 0.7,
	}

	result := types.ExtractOption(options, "nonexistent", 0.5)
	if result != 0.5 {
		t.Errorf("expected default value 0.5, got %v", result)
	}
}

func TestExtractOption_NilOptions(t *testing.T) {
	result := types.ExtractOption[float64](nil, "temperature", 0.5)
	if result != 0.5 {
		t.Errorf("expected default value 0.5, got %v", result)
	}
}

func TestChatResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"id": "chatcmpl-123",
		"object": "chat.completion",
		"created": 1677652288,
		"model": "gpt-4",
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": "Hello there!"
			},
			"finish_reason": "stop"
		}],
		"usage": {
			"prompt_tokens": 9,
			"completion_tokens": 12,
			"total_tokens": 21
		}
	}`

	var response types.ChatResponse
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if response.ID != "chatcmpl-123" {
		t.Errorf("got ID %q, want %q", response.ID, "chatcmpl-123")
	}

	if response.Model != "gpt-4" {
		t.Errorf("got model %q, want %q", response.Model, "gpt-4")
	}

	if len(response.Choices) != 1 {
		t.Fatalf("got %d choices, want 1", len(response.Choices))
	}

	if response.Content() != "Hello there!" {
		t.Errorf("got content %q, want %q", response.Content(), "Hello there!")
	}

	if response.Usage == nil {
		t.Fatal("usage is nil")
	}

	if response.Usage.TotalTokens != 21 {
		t.Errorf("got total tokens %d, want 21", response.Usage.TotalTokens)
	}
}

func TestStreamingChunk_Unmarshal(t *testing.T) {
	jsonData := `{
		"id": "chatcmpl-123",
		"object": "chat.completion.chunk",
		"created": 1677652288,
		"model": "gpt-4",
		"choices": [{
			"index": 0,
			"delta": {
				"content": "Hello"
			},
			"finish_reason": null
		}]
	}`

	var chunk types.StreamingChunk
	if err := json.Unmarshal([]byte(jsonData), &chunk); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if chunk.ID != "chatcmpl-123" {
		t.Errorf("got ID %q, want %q", chunk.ID, "chatcmpl-123")
	}

	if chunk.Model != "gpt-4" {
		t.Errorf("got model %q, want %q", chunk.Model, "gpt-4")
	}

	if chunk.Content() != "Hello" {
		t.Errorf("got content %q, want %q", chunk.Content(), "Hello")
	}
}

func TestEmbeddingsResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"object": "list",
		"data": [{
			"object": "embedding",
			"embedding": [0.1, 0.2, 0.3],
			"index": 0
		}],
		"model": "text-embedding-ada-002",
		"usage": {
			"prompt_tokens": 8,
			"total_tokens": 8
		}
	}`

	var response types.EmbeddingsResponse
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if response.Object != "list" {
		t.Errorf("got object %q, want %q", response.Object, "list")
	}

	if response.Model != "text-embedding-ada-002" {
		t.Errorf("got model %q, want %q", response.Model, "text-embedding-ada-002")
	}

	if len(response.Data) != 1 {
		t.Fatalf("got %d data items, want 1", len(response.Data))
	}

	if len(response.Data[0].Embedding) != 3 {
		t.Fatalf("got %d embedding dimensions, want 3", len(response.Data[0].Embedding))
	}

	if response.Data[0].Embedding[0] != 0.1 {
		t.Errorf("got embedding[0] %f, want 0.1", response.Data[0].Embedding[0])
	}
}

func TestToolsResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"id": "chatcmpl-123",
		"object": "chat.completion",
		"created": 1677652288,
		"model": "gpt-4",
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": "",
				"tool_calls": [{
					"id": "call_123",
					"type": "function",
					"function": {
						"name": "get_weather",
						"arguments": "{\"location\": \"Boston\"}"
					}
				}]
			},
			"finish_reason": "tool_calls"
		}]
	}`

	var response types.ToolsResponse
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if response.ID != "chatcmpl-123" {
		t.Errorf("got ID %q, want %q", response.ID, "chatcmpl-123")
	}

	if len(response.Choices) != 1 {
		t.Fatalf("got %d choices, want 1", len(response.Choices))
	}

	if len(response.Choices[0].Message.ToolCalls) != 1 {
		t.Fatalf("got %d tool calls, want 1", len(response.Choices[0].Message.ToolCalls))
	}

	toolCall := response.Choices[0].Message.ToolCalls[0]

	if toolCall.ID != "call_123" {
		t.Errorf("got tool call ID %q, want %q", toolCall.ID, "call_123")
	}

	if toolCall.Function.Name != "get_weather" {
		t.Errorf("got function name %q, want %q", toolCall.Function.Name, "get_weather")
	}
}
