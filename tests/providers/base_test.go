package providers_test

import (
	"encoding/json"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/providers"
)

func TestNewBaseProvider(t *testing.T) {
	provider := providers.NewBaseProvider("test-provider", "https://api.example.com")

	if provider == nil {
		t.Fatal("NewBaseProvider returned nil")
	}

	if provider.Name() != "test-provider" {
		t.Errorf("got name %q, want %q", provider.Name(), "test-provider")
	}

	if provider.BaseURL() != "https://api.example.com" {
		t.Errorf("got baseURL %q, want %q", provider.BaseURL(), "https://api.example.com")
	}
}

func TestBaseProvider_Name(t *testing.T) {
	provider := providers.NewBaseProvider("my-provider", "https://api.test.com")

	if provider.Name() != "my-provider" {
		t.Errorf("got name %q, want %q", provider.Name(), "my-provider")
	}
}

func TestBaseProvider_BaseURL(t *testing.T) {
	provider := providers.NewBaseProvider("test", "https://custom.api.com/v2")

	if provider.BaseURL() != "https://custom.api.com/v2" {
		t.Errorf("got baseURL %q, want %q", provider.BaseURL(), "https://custom.api.com/v2")
	}
}

func TestBaseProvider_Marshal_Chat(t *testing.T) {
	provider := providers.NewBaseProvider("test", "https://api.test.com")

	chatData := &providers.ChatData{
		Model: "gpt-4",
		Messages: []protocol.Message{
			protocol.NewMessage("user", "Hello"),
		},
		Options: map[string]any{
			"temperature": 0.7,
		},
	}

	body, err := provider.Marshal(protocol.Chat, chatData)
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

func TestBaseProvider_Marshal_Vision(t *testing.T) {
	provider := providers.NewBaseProvider("test", "https://api.test.com")

	visionData := &providers.VisionData{
		Model: "gpt-4-vision",
		Messages: []protocol.Message{
			protocol.NewMessage("user", "What is in this image?"),
		},
		Images: []string{"https://example.com/image.jpg"},
		Options: map[string]any{
			"max_tokens": 1024,
		},
	}

	body, err := provider.Marshal(protocol.Vision, visionData)
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

func TestBaseProvider_Marshal_Tools(t *testing.T) {
	provider := providers.NewBaseProvider("test", "https://api.test.com")

	toolsData := &providers.ToolsData{
		Model: "gpt-4",
		Messages: []protocol.Message{
			protocol.NewMessage("user", "What's the weather?"),
		},
		Tools: []providers.ToolDefinition{
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

	body, err := provider.Marshal(protocol.Tools, toolsData)
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

func TestBaseProvider_Marshal_Embeddings(t *testing.T) {
	provider := providers.NewBaseProvider("test", "https://api.test.com")

	embeddingsData := &providers.EmbeddingsData{
		Model:   "text-embedding-ada-002",
		Input:   "Hello world",
		Options: map[string]any{},
	}

	body, err := provider.Marshal(protocol.Embeddings, embeddingsData)
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

func TestBaseProvider_Marshal_UnsupportedProtocol(t *testing.T) {
	provider := providers.NewBaseProvider("test", "https://api.test.com")

	_, err := provider.Marshal(protocol.Protocol("unsupported"), nil)
	if err == nil {
		t.Error("expected error for unsupported protocol, got nil")
	}
}
