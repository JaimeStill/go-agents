package agent_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JaimeStill/go-agents/pkg/agent"
	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/format"
	"github.com/JaimeStill/go-agents/pkg/response"
)

// openAIChatServer returns a mock server that responds with OpenAI-format chat JSON.
func openAIChatServer(content string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{
					"message":       map[string]any{"role": "assistant", "content": content},
					"finish_reason": "stop",
				},
			},
		})
	}))
}

// openAIToolsServer returns a mock server that responds with OpenAI-format tools JSON.
func openAIToolsServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"role":    "assistant",
						"content": "",
						"tool_calls": []map[string]any{
							{
								"id":   "call_123",
								"type": "function",
								"function": map[string]any{
									"name":      "get_weather",
									"arguments": `{"location":"Boston"}`,
								},
							},
						},
					},
					"finish_reason": "tool_calls",
				},
			},
		})
	}))
}

// openAIEmbeddingsServer returns a mock server that responds with OpenAI-format embeddings JSON.
func openAIEmbeddingsServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"embedding": []float64{0.1, 0.2, 0.3}},
			},
			"model": "test-model",
		})
	}))
}

func newTestConfig(serverURL string, capabilities map[string]map[string]any) *config.AgentConfig {
	return &config.AgentConfig{
		Name: "test-agent",
		Client: &config.ClientConfig{
			Timeout:            config.Duration(30 * time.Second),
			ConnectionTimeout:  config.Duration(10 * time.Second),
			ConnectionPoolSize: 10,
			Retry: config.RetryConfig{
				MaxRetries: 0,
			},
		},
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: serverURL,
		},
		Model: &config.ModelConfig{
			Name:         "test-model",
			Capabilities: capabilities,
		},
	}
}

func TestNew(t *testing.T) {
	server := openAIChatServer("test")
	defer server.Close()

	cfg := newTestConfig(server.URL, map[string]map[string]any{"chat": {"temperature": 0.7}})
	cfg.SystemPrompt = "You are a helpful assistant."

	a, err := agent.New(cfg)

	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if a == nil {
		t.Fatal("New returned nil agent")
	}

	if a.ID() == "" {
		t.Error("agent ID is empty")
	}
}

func TestAgent_ID(t *testing.T) {
	server := openAIChatServer("test")
	defer server.Close()

	cfg := newTestConfig(server.URL, map[string]map[string]any{"chat": {}})

	a1, _ := agent.New(cfg)
	a2, _ := agent.New(cfg)

	if a1.ID() == a2.ID() {
		t.Error("two agents have the same ID")
	}

	id1 := a1.ID()
	id2 := a1.ID()

	if id1 != id2 {
		t.Error("agent ID changed between calls")
	}
}

func TestAgent_Chat(t *testing.T) {
	server := openAIChatServer("Hello, how can I help you?")
	defer server.Close()

	cfg := newTestConfig(server.URL, map[string]map[string]any{"chat": {}})
	cfg.SystemPrompt = "You are helpful."

	a, err := agent.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	resp, err := a.Chat(context.Background(), "Hello")
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Chat returned nil response")
	}

	if resp.Text() != "Hello, how can I help you?" {
		t.Errorf("got text %q, want %q", resp.Text(), "Hello, how can I help you?")
	}
}

func TestAgent_Vision(t *testing.T) {
	server := openAIChatServer("I see a cat in the image.")
	defer server.Close()

	cfg := newTestConfig(server.URL, map[string]map[string]any{"vision": {}})

	a, err := agent.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	pngData, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUg==")
	images := []format.Image{
		{Data: pngData, Format: "png"},
	}
	resp, err := a.Vision(context.Background(), "What's in this image?", images)
	if err != nil {
		t.Fatalf("Vision failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Vision returned nil response")
	}

	if resp.Text() != "I see a cat in the image." {
		t.Errorf("got text %q, want %q", resp.Text(), "I see a cat in the image.")
	}
}

func TestAgent_Tools(t *testing.T) {
	server := openAIToolsServer()
	defer server.Close()

	cfg := newTestConfig(server.URL, map[string]map[string]any{"tools": {}})

	a, err := agent.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	tools := []agent.Tool{
		{
			Name:        "get_weather",
			Description: "Get weather for a location",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"location": map[string]any{
						"type": "string",
					},
				},
			},
		},
	}

	resp, err := a.Tools(context.Background(), "What's the weather in Boston?", tools)
	if err != nil {
		t.Fatalf("Tools failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Tools returned nil response")
	}

	calls := resp.ToolCalls()
	if len(calls) == 0 {
		t.Fatal("response has no tool calls")
	}

	if calls[0].Name != "get_weather" {
		t.Errorf("got function name %q, want %q", calls[0].Name, "get_weather")
	}
}

func TestAgent_Embed(t *testing.T) {
	server := openAIEmbeddingsServer()
	defer server.Close()

	cfg := newTestConfig(server.URL, map[string]map[string]any{"embeddings": {}})

	a, err := agent.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	resp, err := a.Embed(context.Background(), "Hello, world!")
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Embed returned nil response")
	}

	if len(resp.Embeddings) == 0 {
		t.Fatal("response has no embeddings")
	}

	if len(resp.Embeddings[0]) != 3 {
		t.Errorf("got %d dimensions, want 3", len(resp.Embeddings[0]))
	}
}

func TestAgent_Client(t *testing.T) {
	server := openAIChatServer("test")
	defer server.Close()

	cfg := newTestConfig(server.URL, map[string]map[string]any{"chat": {}})

	a, err := agent.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	client := a.Client()

	if client == nil {
		t.Error("Client() returned nil")
	}
}

func TestAgent_Provider(t *testing.T) {
	server := openAIChatServer("test")
	defer server.Close()

	cfg := newTestConfig(server.URL, map[string]map[string]any{"chat": {}})

	a, err := agent.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	provider := a.Provider()

	if provider == nil {
		t.Error("Provider() returned nil")
	}

	if provider.Name() != "ollama" {
		t.Errorf("got provider name %q, want %q", provider.Name(), "ollama")
	}
}

func TestAgent_Model(t *testing.T) {
	server := openAIChatServer("test")
	defer server.Close()

	cfg := newTestConfig(server.URL, map[string]map[string]any{"chat": {}})

	a, err := agent.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	mdl := a.Model()

	if mdl == nil {
		t.Error("Model() returned nil")
	}

	if mdl.Name != "test-model" {
		t.Errorf("got model name %q, want %q", mdl.Name, "test-model")
	}
}

// Verify agent.New defaults format to "openai" when not specified
func TestNew_DefaultFormat(t *testing.T) {
	server := openAIChatServer("test")
	defer server.Close()

	cfg := newTestConfig(server.URL, map[string]map[string]any{"chat": {}})

	a, err := agent.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if a.Format() == nil {
		t.Fatal("Format() returned nil")
	}

	if a.Format().Name() != "openai" {
		t.Errorf("got format name %q, want %q", a.Format().Name(), "openai")
	}
}

// Verify unused import guard
var _ = response.Response{}
