package format_test

import (
	"encoding/json"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/format"
	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/response"
)

func TestOpenAIFormat_Name(t *testing.T) {
	f, err := format.OpenAIFactory()
	if err != nil {
		t.Fatalf("OpenAIFactory failed: %v", err)
	}

	if f.Name() != "openai" {
		t.Errorf("got name %q, want %q", f.Name(), "openai")
	}
}

func TestOpenAIFormat_Marshal_Chat(t *testing.T) {
	f, _ := format.OpenAIFactory()

	chatData := &format.ChatData{
		Model: "gpt-4",
		Messages: []protocol.Message{
			protocol.NewMessage("user", "Hello"),
		},
		Options: map[string]any{
			"temperature": 0.7,
		},
	}

	body, err := f.Marshal(protocol.Chat, chatData)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if result["model"] != "gpt-4" {
		t.Errorf("got model %v, want gpt-4", result["model"])
	}

	if result["temperature"] != 0.7 {
		t.Errorf("got temperature %v, want 0.7", result["temperature"])
	}

	messages, ok := result["messages"].([]any)
	if !ok {
		t.Fatal("messages is not an array")
	}
	if len(messages) != 1 {
		t.Errorf("got %d messages, want 1", len(messages))
	}
}

func TestOpenAIFormat_Marshal_Vision(t *testing.T) {
	f, _ := format.OpenAIFactory()

	visionData := &format.VisionData{
		Model: "gpt-4-vision",
		Messages: []protocol.Message{
			protocol.NewMessage("user", "What is in this image?"),
		},
		Images: []format.Image{
			{URL: "https://example.com/image.jpg"},
		},
		Options: map[string]any{
			"max_tokens": 1024,
		},
	}

	body, err := f.Marshal(protocol.Vision, visionData)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if result["model"] != "gpt-4-vision" {
		t.Errorf("got model %v, want gpt-4-vision", result["model"])
	}

	if result["max_tokens"] != float64(1024) {
		t.Errorf("got max_tokens %v, want 1024", result["max_tokens"])
	}
}

func TestOpenAIFormat_Marshal_Tools(t *testing.T) {
	f, _ := format.OpenAIFactory()

	toolsData := &format.ToolsData{
		Model: "gpt-4",
		Messages: []protocol.Message{
			protocol.NewMessage("user", "What's the weather?"),
		},
		Tools: []format.ToolDefinition{
			{
				Name:        "get_weather",
				Description: "Get weather for a location",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The city name",
						},
					},
				},
			},
		},
		Options: map[string]any{},
	}

	body, err := f.Marshal(protocol.Tools, toolsData)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if result["model"] != "gpt-4" {
		t.Errorf("got model %v, want gpt-4", result["model"])
	}

	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatal("tools is not an array")
	}
	if len(tools) != 1 {
		t.Errorf("got %d tools, want 1", len(tools))
	}
}

func TestOpenAIFormat_Marshal_Embeddings(t *testing.T) {
	f, _ := format.OpenAIFactory()

	embeddingsData := &format.EmbeddingsData{
		Model:   "text-embedding-ada-002",
		Input:   "Hello world",
		Options: map[string]any{},
	}

	body, err := f.Marshal(protocol.Embeddings, embeddingsData)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if result["model"] != "text-embedding-ada-002" {
		t.Errorf("got model %v, want text-embedding-ada-002", result["model"])
	}

	if result["input"] != "Hello world" {
		t.Errorf("got input %v, want 'Hello world'", result["input"])
	}
}

func TestOpenAIFormat_Marshal_UnsupportedProtocol(t *testing.T) {
	f, _ := format.OpenAIFactory()

	_, err := f.Marshal(protocol.Protocol("unsupported"), nil)
	if err == nil {
		t.Error("expected error for unsupported protocol, got nil")
	}
}

func TestOpenAIFormat_Parse_Chat(t *testing.T) {
	f, _ := format.OpenAIFactory()

	chatJSON := `{
		"choices": [{
			"message": {"role": "assistant", "content": "Hello!"},
			"finish_reason": "stop"
		}],
		"usage": {"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15}
	}`

	result, err := f.Parse(protocol.Chat, []byte(chatJSON))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	resp, ok := result.(*response.Response)
	if !ok {
		t.Fatalf("expected *response.Response, got %T", result)
	}

	if resp.Text() != "Hello!" {
		t.Errorf("got text %q, want %q", resp.Text(), "Hello!")
	}

	if resp.StopReason != "stop" {
		t.Errorf("got stop reason %q, want %q", resp.StopReason, "stop")
	}

	if resp.Usage == nil {
		t.Fatal("usage is nil")
	}

	if resp.Usage.InputTokens != 10 {
		t.Errorf("got input tokens %d, want 10", resp.Usage.InputTokens)
	}

	if resp.Usage.OutputTokens != 5 {
		t.Errorf("got output tokens %d, want 5", resp.Usage.OutputTokens)
	}

	if resp.Usage.TotalTokens != 15 {
		t.Errorf("got total tokens %d, want 15", resp.Usage.TotalTokens)
	}
}

func TestOpenAIFormat_Parse_Tools(t *testing.T) {
	f, _ := format.OpenAIFactory()

	toolsJSON := `{
		"choices": [{
			"message": {
				"role": "assistant",
				"content": "",
				"tool_calls": [{
					"id": "call_abc",
					"type": "function",
					"function": {
						"name": "get_weather",
						"arguments": "{\"location\":\"London\"}"
					}
				}]
			},
			"finish_reason": "tool_calls"
		}]
	}`

	result, err := f.Parse(protocol.Tools, []byte(toolsJSON))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	resp, ok := result.(*response.Response)
	if !ok {
		t.Fatalf("expected *response.Response, got %T", result)
	}

	calls := resp.ToolCalls()
	if len(calls) != 1 {
		t.Fatalf("got %d tool calls, want 1", len(calls))
	}

	if calls[0].Name != "get_weather" {
		t.Errorf("got function name %q, want %q", calls[0].Name, "get_weather")
	}

	if calls[0].ID != "call_abc" {
		t.Errorf("got ID %q, want %q", calls[0].ID, "call_abc")
	}

	location, ok := calls[0].Input["location"].(string)
	if !ok {
		t.Fatal("location is not a string")
	}

	if location != "London" {
		t.Errorf("got location %q, want %q", location, "London")
	}
}

func TestOpenAIFormat_Parse_Embeddings(t *testing.T) {
	f, _ := format.OpenAIFactory()

	embeddingsJSON := `{
		"data": [{
			"embedding": [0.1, 0.2, 0.3]
		}],
		"model": "text-embedding-ada-002",
		"usage": {"prompt_tokens": 5, "total_tokens": 5}
	}`

	result, err := f.Parse(protocol.Embeddings, []byte(embeddingsJSON))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	embResp, ok := result.(*response.EmbeddingsResponse)
	if !ok {
		t.Fatalf("expected *response.EmbeddingsResponse, got %T", result)
	}

	if len(embResp.Embeddings) != 1 {
		t.Fatalf("got %d embeddings, want 1", len(embResp.Embeddings))
	}

	if len(embResp.Embeddings[0]) != 3 {
		t.Errorf("got %d embedding dimensions, want 3", len(embResp.Embeddings[0]))
	}

	if embResp.Model != "text-embedding-ada-002" {
		t.Errorf("got model %q, want %q", embResp.Model, "text-embedding-ada-002")
	}
}

func TestOpenAIFormat_Parse_UnsupportedProtocol(t *testing.T) {
	f, _ := format.OpenAIFactory()

	_, err := f.Parse(protocol.Protocol("unsupported"), []byte(`{}`))
	if err == nil {
		t.Error("expected error for unsupported protocol, got nil")
	}
}

func TestOpenAIFormat_ParseStreamChunk_Chat(t *testing.T) {
	f, _ := format.OpenAIFactory()

	chunkJSON := `{
		"choices": [{
			"delta": {"content": "Hello"},
			"finish_reason": null
		}]
	}`

	chunk, err := f.ParseStreamChunk(protocol.Chat, []byte(chunkJSON))
	if err != nil {
		t.Fatalf("ParseStreamChunk failed: %v", err)
	}

	if chunk.Text() != "Hello" {
		t.Errorf("got text %q, want %q", chunk.Text(), "Hello")
	}
}

func TestOpenAIFormat_ParseStreamChunk_FinishReason(t *testing.T) {
	f, _ := format.OpenAIFactory()

	chunkJSON := `{
		"choices": [{
			"delta": {},
			"finish_reason": "stop"
		}]
	}`

	chunk, err := f.ParseStreamChunk(protocol.Chat, []byte(chunkJSON))
	if err != nil {
		t.Fatalf("ParseStreamChunk failed: %v", err)
	}

	if chunk.StopReason != "stop" {
		t.Errorf("got stop reason %q, want %q", chunk.StopReason, "stop")
	}
}

func TestOpenAIFormat_ParseStreamChunk_Embeddings(t *testing.T) {
	f, _ := format.OpenAIFactory()

	_, err := f.ParseStreamChunk(protocol.Embeddings, []byte(`{}`))
	if err == nil {
		t.Error("expected error for embeddings streaming, got nil")
	}
}
