package format_test

import (
	"encoding/json"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/format"
	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/response"
)

func TestConverseFormat_Name(t *testing.T) {
	f, err := format.ConverseFactory()
	if err != nil {
		t.Fatalf("ConverseFactory failed: %v", err)
	}

	if f.Name() != "converse" {
		t.Errorf("got name %q, want %q", f.Name(), "converse")
	}
}

func TestConverseFormat_Marshal_Chat(t *testing.T) {
	f, _ := format.ConverseFactory()

	chatData := &format.ChatData{
		Model: "anthropic.claude-sonnet-4-6-v1",
		Messages: []protocol.Message{
			protocol.NewMessage("user", "Hello"),
		},
		Options: map[string]any{},
	}

	body, err := f.Marshal(protocol.Chat, chatData)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if result["modelId"] != "anthropic.claude-sonnet-4-6-v1" {
		t.Errorf("got modelId %v, want anthropic.claude-sonnet-4-6-v1", result["modelId"])
	}

	messages, ok := result["messages"].([]any)
	if !ok {
		t.Fatal("messages is not an array")
	}
	if len(messages) != 1 {
		t.Errorf("got %d messages, want 1", len(messages))
	}
}

func TestConverseFormat_Marshal_Chat_SystemMessage(t *testing.T) {
	f, _ := format.ConverseFactory()

	chatData := &format.ChatData{
		Model: "anthropic.claude-sonnet-4-6-v1",
		Messages: []protocol.Message{
			protocol.NewMessage("system", "You are helpful."),
			protocol.NewMessage("user", "Hello"),
		},
		Options: map[string]any{},
	}

	body, err := f.Marshal(protocol.Chat, chatData)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// System messages should be separated into the top-level system field
	system, ok := result["system"].([]any)
	if !ok {
		t.Fatal("system field is missing or not an array")
	}
	if len(system) != 1 {
		t.Fatalf("got %d system blocks, want 1", len(system))
	}

	systemBlock, ok := system[0].(map[string]any)
	if !ok {
		t.Fatal("system block is not a map")
	}
	if systemBlock["text"] != "You are helpful." {
		t.Errorf("got system text %v, want %q", systemBlock["text"], "You are helpful.")
	}

	// User message should be in the messages array (system excluded)
	messages, ok := result["messages"].([]any)
	if !ok {
		t.Fatal("messages is not an array")
	}
	if len(messages) != 1 {
		t.Errorf("got %d messages, want 1 (system should be separated)", len(messages))
	}
}

func TestConverseFormat_Marshal_Vision(t *testing.T) {
	f, _ := format.ConverseFactory()

	visionData := &format.VisionData{
		Model: "anthropic.claude-sonnet-4-6-v1",
		Messages: []protocol.Message{
			protocol.NewMessage("user", "What is in this image?"),
		},
		Images: []format.Image{
			{Data: []byte("fakepng"), Format: "png"},
		},
		Options: map[string]any{},
	}

	body, err := f.Marshal(protocol.Vision, visionData)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if result["modelId"] != "anthropic.claude-sonnet-4-6-v1" {
		t.Errorf("got modelId %v, want anthropic.claude-sonnet-4-6-v1", result["modelId"])
	}

	// Verify the last message contains both text and image content blocks
	messages := result["messages"].([]any)
	lastMsg := messages[0].(map[string]any)
	content := lastMsg["content"].([]any)

	if len(content) != 2 {
		t.Fatalf("got %d content blocks, want 2 (text + image)", len(content))
	}
}

func TestConverseFormat_Marshal_Tools(t *testing.T) {
	f, _ := format.ConverseFactory()

	toolsData := &format.ToolsData{
		Model: "anthropic.claude-sonnet-4-6-v1",
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
						"location": map[string]any{"type": "string"},
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

	toolConfig, ok := result["toolConfig"].(map[string]any)
	if !ok {
		t.Fatal("toolConfig is missing or not a map")
	}

	tools, ok := toolConfig["tools"].([]any)
	if !ok {
		t.Fatal("tools is not an array")
	}
	if len(tools) != 1 {
		t.Errorf("got %d tools, want 1", len(tools))
	}

	tool := tools[0].(map[string]any)
	toolSpec, ok := tool["toolSpec"].(map[string]any)
	if !ok {
		t.Fatal("toolSpec is missing")
	}
	if toolSpec["name"] != "get_weather" {
		t.Errorf("got tool name %v, want %q", toolSpec["name"], "get_weather")
	}
}

func TestConverseFormat_Marshal_Embeddings_Error(t *testing.T) {
	f, _ := format.ConverseFactory()

	_, err := f.Marshal(protocol.Embeddings, nil)
	if err == nil {
		t.Error("expected error for embeddings protocol, got nil")
	}
}

func TestConverseFormat_Parse_ChatResponse(t *testing.T) {
	f, _ := format.ConverseFactory()

	converseJSON := `{
		"output": {
			"message": {
				"role": "assistant",
				"content": [{"text": "Hello! How can I help?"}]
			}
		},
		"stopReason": "end_turn",
		"usage": {
			"inputTokens": 10,
			"outputTokens": 8,
			"totalTokens": 18
		}
	}`

	result, err := f.Parse(protocol.Chat, []byte(converseJSON))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	resp, ok := result.(*response.Response)
	if !ok {
		t.Fatalf("expected *response.Response, got %T", result)
	}

	if resp.Text() != "Hello! How can I help?" {
		t.Errorf("got text %q, want %q", resp.Text(), "Hello! How can I help?")
	}

	if resp.StopReason != "end_turn" {
		t.Errorf("got stop reason %q, want %q", resp.StopReason, "end_turn")
	}

	if resp.Usage == nil {
		t.Fatal("usage is nil")
	}

	if resp.Usage.InputTokens != 10 {
		t.Errorf("got input tokens %d, want 10", resp.Usage.InputTokens)
	}

	if resp.Usage.OutputTokens != 8 {
		t.Errorf("got output tokens %d, want 8", resp.Usage.OutputTokens)
	}

	if resp.Usage.TotalTokens != 18 {
		t.Errorf("got total tokens %d, want 18", resp.Usage.TotalTokens)
	}
}

func TestConverseFormat_Parse_ToolsResponse(t *testing.T) {
	f, _ := format.ConverseFactory()

	converseJSON := `{
		"output": {
			"message": {
				"role": "assistant",
				"content": [
					{"text": "Let me check that."},
					{
						"toolUse": {
							"toolUseId": "tool_abc",
							"name": "get_weather",
							"input": {"location": "Boston"}
						}
					}
				]
			}
		},
		"stopReason": "tool_use",
		"usage": {
			"inputTokens": 20,
			"outputTokens": 15,
			"totalTokens": 35
		}
	}`

	result, err := f.Parse(protocol.Tools, []byte(converseJSON))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	resp, ok := result.(*response.Response)
	if !ok {
		t.Fatalf("expected *response.Response, got %T", result)
	}

	if resp.Text() != "Let me check that." {
		t.Errorf("got text %q, want %q", resp.Text(), "Let me check that.")
	}

	calls := resp.ToolCalls()
	if len(calls) != 1 {
		t.Fatalf("got %d tool calls, want 1", len(calls))
	}

	if calls[0].ID != "tool_abc" {
		t.Errorf("got tool ID %q, want %q", calls[0].ID, "tool_abc")
	}

	if calls[0].Name != "get_weather" {
		t.Errorf("got tool name %q, want %q", calls[0].Name, "get_weather")
	}

	location, ok := calls[0].Input["location"].(string)
	if !ok {
		t.Fatal("location is not a string")
	}
	if location != "Boston" {
		t.Errorf("got location %q, want %q", location, "Boston")
	}

	if resp.StopReason != "tool_use" {
		t.Errorf("got stop reason %q, want %q", resp.StopReason, "tool_use")
	}
}

func TestConverseFormat_Parse_Embeddings_Error(t *testing.T) {
	f, _ := format.ConverseFactory()

	_, err := f.Parse(protocol.Embeddings, []byte(`{}`))
	if err == nil {
		t.Error("expected error for embeddings protocol, got nil")
	}
}

func TestConverseFormat_ParseStreamChunk_ContentBlockDelta(t *testing.T) {
	f, _ := format.ConverseFactory()

	eventJSON := `{"contentBlockDelta": {"contentBlockIndex": 0, "delta": {"text": "Hello"}}}`

	chunk, err := f.ParseStreamChunk(protocol.Chat, []byte(eventJSON))
	if err != nil {
		t.Fatalf("ParseStreamChunk failed: %v", err)
	}

	if chunk.Text() != "Hello" {
		t.Errorf("got text %q, want %q", chunk.Text(), "Hello")
	}
}

func TestConverseFormat_ParseStreamChunk_MessageStop(t *testing.T) {
	f, _ := format.ConverseFactory()

	eventJSON := `{"messageStop": {"stopReason": "end_turn"}}`

	chunk, err := f.ParseStreamChunk(protocol.Chat, []byte(eventJSON))
	if err != nil {
		t.Fatalf("ParseStreamChunk failed: %v", err)
	}

	if chunk.StopReason != "end_turn" {
		t.Errorf("got stop reason %q, want %q", chunk.StopReason, "end_turn")
	}
}

func TestConverseFormat_ParseStreamChunk_SkippableEvent(t *testing.T) {
	f, _ := format.ConverseFactory()

	eventJSON := `{"messageStart": {"role": "assistant"}}`

	chunk, err := f.ParseStreamChunk(protocol.Chat, []byte(eventJSON))
	if err != nil {
		t.Errorf("expected nil error for skippable event, got %v", err)
	}
	if chunk != nil {
		t.Error("expected nil chunk for skippable event")
	}
}
