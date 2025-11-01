package types_test

import (
	"encoding/json"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/types"
)

func TestChatRequest_GetProtocol(t *testing.T) {
	req := &types.ChatRequest{}
	if req.GetProtocol() != types.Chat {
		t.Errorf("got protocol %s, want %s", req.GetProtocol(), types.Chat)
	}
}

func TestChatRequest_GetHeaders(t *testing.T) {
	req := &types.ChatRequest{}
	headers := req.GetHeaders()

	if headers["Content-Type"] != "application/json" {
		t.Errorf("got Content-Type %q, want %q", headers["Content-Type"], "application/json")
	}
}

func TestChatRequest_Marshal(t *testing.T) {
	req := &types.ChatRequest{
		Messages: []types.Message{
			types.NewMessage("user", "Hello"),
		},
		Options: map[string]any{
			"model":       "gpt-4",
			"temperature": 0.7,
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

	// Check messages field
	if _, exists := result["messages"]; !exists {
		t.Error("messages field missing")
	}

	// Check options are merged at root level
	if result["model"] != "gpt-4" {
		t.Errorf("got model %v, want gpt-4", result["model"])
	}

	if result["temperature"] != 0.7 {
		t.Errorf("got temperature %v, want 0.7", result["temperature"])
	}
}

func TestVisionRequest_GetProtocol(t *testing.T) {
	req := &types.VisionRequest{}
	if req.GetProtocol() != types.Vision {
		t.Errorf("got protocol %s, want %s", req.GetProtocol(), types.Vision)
	}
}

func TestVisionRequest_Marshal(t *testing.T) {
	req := &types.VisionRequest{
		Messages: []types.Message{
			types.NewMessage("user", "What's in this image?"),
		},
		Images: []string{
			"data:image/png;base64,iVBORw0KGgoAAAANSUhEUg==",
		},
		VisionOptions: map[string]any{
			"detail": "high",
		},
		Options: map[string]any{
			"model":       "gpt-4o",
			"max_tokens": 300,
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

	// Check messages field exists
	messages, ok := result["messages"].([]any)
	if !ok {
		t.Fatal("messages field missing or wrong type")
	}

	// Check last message has structured content with images
	lastMsg := messages[len(messages)-1].(map[string]any)
	content, ok := lastMsg["content"].([]any)
	if !ok {
		t.Fatal("last message content is not structured")
	}

	// Should have text + image content
	if len(content) < 2 {
		t.Errorf("got %d content blocks, want at least 2 (text + image)", len(content))
	}

	// Check image content has image_options embedded
	imageBlock := content[1].(map[string]any)
	if imageBlock["type"] != "image_url" {
		t.Errorf("got type %v, want image_url", imageBlock["type"])
	}

	imageURL, ok := imageBlock["image_url"].(map[string]any)
	if !ok {
		t.Fatal("image_url field missing or wrong type")
	}

	if imageURL["detail"] != "high" {
		t.Errorf("got detail %v, want high", imageURL["detail"])
	}

	// Check model options merged at root
	if result["model"] != "gpt-4o" {
		t.Errorf("got model %v, want gpt-4o", result["model"])
	}
}

func TestVisionRequest_Marshal_MultipleImages(t *testing.T) {
	req := &types.VisionRequest{
		Messages: []types.Message{
			types.NewMessage("user", "Compare these images"),
		},
		Images: []string{
			"data:image/png;base64,image1",
			"data:image/png;base64,image2",
		},
		VisionOptions: map[string]any{
			"detail": "low",
		},
		Options: map[string]any{
			"model": "gpt-4o",
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

	messages := result["messages"].([]any)
	lastMsg := messages[len(messages)-1].(map[string]any)
	content := lastMsg["content"].([]any)

	// Should have text + 2 images = 3 content blocks
	if len(content) != 3 {
		t.Errorf("got %d content blocks, want 3 (text + 2 images)", len(content))
	}
}

func TestToolsRequest_GetProtocol(t *testing.T) {
	req := &types.ToolsRequest{}
	if req.GetProtocol() != types.Tools {
		t.Errorf("got protocol %s, want %s", req.GetProtocol(), types.Tools)
	}
}

func TestToolsRequest_Marshal_OpenAIFormat(t *testing.T) {
	req := &types.ToolsRequest{
		Messages: []types.Message{
			types.NewMessage("user", "What's the weather in Boston?"),
		},
		Tools: []types.ToolDefinition{
			{
				Name:        "get_weather",
				Description: "Get current weather for a location",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "City name",
						},
					},
					"required": []string{"location"},
				},
			},
		},
		Options: map[string]any{
			"model": "gpt-4",
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

	// Check tools are wrapped in OpenAI format
	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatal("tools field missing or wrong type")
	}

	if len(tools) != 1 {
		t.Fatalf("got %d tools, want 1", len(tools))
	}

	tool := tools[0].(map[string]any)

	// Check OpenAI wrapping: {"type": "function", "function": {...}}
	if tool["type"] != "function" {
		t.Errorf("got type %v, want function", tool["type"])
	}

	function, ok := tool["function"].(map[string]any)
	if !ok {
		t.Fatal("function field missing or wrong type")
	}

	if function["name"] != "get_weather" {
		t.Errorf("got name %v, want get_weather", function["name"])
	}

	if function["description"] != "Get current weather for a location" {
		t.Errorf("got description %v", function["description"])
	}

	// Check model option merged at root
	if result["model"] != "gpt-4" {
		t.Errorf("got model %v, want gpt-4", result["model"])
	}
}

func TestEmbeddingsRequest_GetProtocol(t *testing.T) {
	req := &types.EmbeddingsRequest{}
	if req.GetProtocol() != types.Embeddings {
		t.Errorf("got protocol %s, want %s", req.GetProtocol(), types.Embeddings)
	}
}

func TestEmbeddingsRequest_Marshal_StringInput(t *testing.T) {
	req := &types.EmbeddingsRequest{
		Input: "Hello, world!",
		Options: map[string]any{
			"model": "text-embedding-ada-002",
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

	if result["input"] != "Hello, world!" {
		t.Errorf("got input %v, want 'Hello, world!'", result["input"])
	}

	if result["model"] != "text-embedding-ada-002" {
		t.Errorf("got model %v", result["model"])
	}
}

func TestEmbeddingsRequest_Marshal_ArrayInput(t *testing.T) {
	req := &types.EmbeddingsRequest{
		Input: []string{"Hello", "World"},
		Options: map[string]any{
			"model": "text-embedding-ada-002",
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

	input, ok := result["input"].([]any)
	if !ok {
		t.Fatal("input field is not an array")
	}

	if len(input) != 2 {
		t.Errorf("got %d inputs, want 2", len(input))
	}
}

func TestProtocol_SupportsStreaming(t *testing.T) {
	tests := []struct {
		name              string
		protocol          types.Protocol
		supportsStreaming bool
	}{
		{
			name:              "Chat",
			protocol:          types.Chat,
			supportsStreaming: true,
		},
		{
			name:              "Vision",
			protocol:          types.Vision,
			supportsStreaming: true,
		},
		{
			name:              "Tools",
			protocol:          types.Tools,
			supportsStreaming: true,
		},
		{
			name:              "Embeddings",
			protocol:          types.Embeddings,
			supportsStreaming: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.protocol.SupportsStreaming(); got != tt.supportsStreaming {
				t.Errorf("SupportsStreaming() = %v, want %v", got, tt.supportsStreaming)
			}
		})
	}
}
