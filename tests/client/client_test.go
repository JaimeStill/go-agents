package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JaimeStill/go-agents/pkg/client"
	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/format"
	"github.com/JaimeStill/go-agents/pkg/model"
	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/providers"
	"github.com/JaimeStill/go-agents/pkg/request"
	"github.com/JaimeStill/go-agents/pkg/response"
)

func TestNew(t *testing.T) {
	cfg := &config.ClientConfig{
		Timeout:            config.Duration(30 * time.Second),
		ConnectionTimeout:  config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
	}

	c := client.New(cfg)

	if c == nil {
		t.Fatal("New returned nil client")
	}
}

func TestClient_Execute_Chat(t *testing.T) {
	// Server returns OpenAI-format JSON — the format layer parses it
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{
					"message":       map[string]any{"role": "assistant", "content": "Hello, world!"},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]any{
				"prompt_tokens":     10,
				"completion_tokens": 5,
				"total_tokens":      15,
			},
		})
	}))
	defer server.Close()

	providerCfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: server.URL,
	}
	provider, err := providers.NewOllama(providerCfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	mdl := model.New(&config.ModelConfig{
		Name: "test-model",
	})

	cfg := &config.ClientConfig{
		Timeout:            config.Duration(30 * time.Second),
		ConnectionTimeout:  config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
		Retry: config.RetryConfig{
			MaxRetries: 0,
		},
	}
	c := client.New(cfg)

	messages := []protocol.Message{
		protocol.NewMessage("user", "Hello"),
	}
	f, _ := format.OpenAIFactory()
	req := request.NewChat(provider, f, mdl, messages, map[string]any{})

	result, err := c.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resp, ok := result.(*response.Response)
	if !ok {
		t.Fatalf("expected *response.Response, got %T", result)
	}

	if resp.Text() != "Hello, world!" {
		t.Errorf("got text %q, want %q", resp.Text(), "Hello, world!")
	}
}

func TestClient_Execute_Tools(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	defer server.Close()

	providerCfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: server.URL,
	}
	provider, err := providers.NewOllama(providerCfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	mdl := model.New(&config.ModelConfig{
		Name: "test-model",
	})

	cfg := &config.ClientConfig{
		Timeout:            config.Duration(30 * time.Second),
		ConnectionTimeout:  config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
		Retry: config.RetryConfig{
			MaxRetries: 0,
		},
	}
	c := client.New(cfg)

	messages := []protocol.Message{
		protocol.NewMessage("user", "What's the weather in Boston?"),
	}

	tools := []format.ToolDefinition{
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
			},
		},
	}

	f, _ := format.OpenAIFactory()
	req := request.NewTools(provider, f, mdl, messages, tools, map[string]any{})

	result, err := c.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resp, ok := result.(*response.Response)
	if !ok {
		t.Fatalf("expected *response.Response, got %T", result)
	}

	calls := resp.ToolCalls()
	if len(calls) == 0 {
		t.Fatal("no tool calls in response")
	}

	if calls[0].Name != "get_weather" {
		t.Errorf("got function name %q, want %q", calls[0].Name, "get_weather")
	}
}

func TestClient_Execute_Embeddings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"embedding": []float64{0.1, 0.2, 0.3}},
			},
			"model": "test-model",
			"usage": map[string]any{
				"prompt_tokens": 5,
				"total_tokens":  5,
			},
		})
	}))
	defer server.Close()

	providerCfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: server.URL,
	}
	provider, err := providers.NewOllama(providerCfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	mdl := model.New(&config.ModelConfig{
		Name: "test-model",
	})

	cfg := &config.ClientConfig{
		Timeout:            config.Duration(30 * time.Second),
		ConnectionTimeout:  config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
		Retry: config.RetryConfig{
			MaxRetries: 0,
		},
	}
	c := client.New(cfg)

	f, _ := format.OpenAIFactory()
	req := request.NewEmbeddings(provider, f, mdl, "Hello, world!", map[string]any{})

	result, err := c.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	embResp, ok := result.(*response.EmbeddingsResponse)
	if !ok {
		t.Fatalf("expected *response.EmbeddingsResponse, got %T", result)
	}

	if len(embResp.Embeddings) == 0 {
		t.Fatal("no embeddings in response")
	}

	if len(embResp.Embeddings[0]) != 3 {
		t.Errorf("got %d dimensions, want 3", len(embResp.Embeddings[0]))
	}
}

func TestClient_Execute_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	providerCfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: server.URL,
	}
	provider, err := providers.NewOllama(providerCfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	mdl := model.New(&config.ModelConfig{
		Name: "test-model",
	})

	cfg := &config.ClientConfig{
		Timeout:            config.Duration(30 * time.Second),
		ConnectionTimeout:  config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
		Retry: config.RetryConfig{
			MaxRetries: 0,
		},
	}
	c := client.New(cfg)

	messages := []protocol.Message{
		protocol.NewMessage("user", "Hello"),
	}
	f, _ := format.OpenAIFactory()
	req := request.NewChat(provider, f, mdl, messages, map[string]any{})

	_, err = c.Execute(context.Background(), req)
	if err == nil {
		t.Error("expected error for HTTP 500, got nil")
	}
}

func TestClient_ExecuteStream_UnsupportedProtocol(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	providerCfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: server.URL,
	}
	provider, err := providers.NewOllama(providerCfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	mdl := model.New(&config.ModelConfig{
		Name: "test-model",
	})

	cfg := &config.ClientConfig{
		Timeout:            config.Duration(30 * time.Second),
		ConnectionTimeout:  config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
	}
	c := client.New(cfg)

	f, _ := format.OpenAIFactory()
	req := request.NewEmbeddings(provider, f, mdl, "test", map[string]any{})

	_, err = c.ExecuteStream(context.Background(), req)
	if err == nil {
		t.Error("expected error for unsupported streaming protocol, got nil")
	}
}

func TestClient_IsHealthy(t *testing.T) {
	cfg := &config.ClientConfig{
		Timeout:            config.Duration(30 * time.Second),
		ConnectionTimeout:  config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
	}

	c := client.New(cfg)

	if !c.IsHealthy() {
		t.Error("expected client to be healthy initially")
	}
}

func TestClient_HTTPClient(t *testing.T) {
	cfg := &config.ClientConfig{
		Timeout:            config.Duration(5 * time.Second),
		ConnectionTimeout:  config.Duration(2 * time.Second),
		ConnectionPoolSize: 20,
	}

	c := client.New(cfg)

	httpClient := c.HTTPClient()

	if httpClient == nil {
		t.Fatal("HTTPClient() returned nil")
	}

	if httpClient.Timeout != 5*time.Second {
		t.Errorf("got timeout %v, want %v", httpClient.Timeout, 5*time.Second)
	}
}
