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
	"github.com/JaimeStill/go-agents/pkg/types"
)

func TestNew(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.ClientConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: server.URL,
			Model: &config.ModelConfig{
				Name: "test-model",
				Capabilities: map[string]map[string]any{
					"chat": {
						"temperature": 0.7,
					},
				},
			},
		},
		Timeout:            config.Duration(30 * time.Second),
		ConnectionTimeout:  config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
	}

	c, err := client.New(cfg)

	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if c == nil {
		t.Fatal("New returned nil client")
	}
}

func TestNew_InvalidProvider(t *testing.T) {
	cfg := &config.ClientConfig{
		Provider: &config.ProviderConfig{
			Name:    "unknown-provider",
			BaseURL: "http://localhost",
			Model: &config.ModelConfig{
				Name: "test-model",
			},
		},
		Timeout:            config.Duration(30 * time.Second),
		ConnectionTimeout:  config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
	}

	_, err := client.New(cfg)

	if err == nil {
		t.Error("expected error for unknown provider, got nil")
	}
}

func TestClient_Provider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.ClientConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: server.URL,
			Model: &config.ModelConfig{
				Name: "test-model",
			},
		},
		Timeout:            config.Duration(30 * time.Second),
		ConnectionTimeout:  config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
	}

	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	provider := c.Provider()

	if provider == nil {
		t.Fatal("Provider() returned nil")
	}

	if provider.Name() != "ollama" {
		t.Errorf("got provider name %q, want %q", provider.Name(), "ollama")
	}
}

func TestClient_Model(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.ClientConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: server.URL,
			Model: &config.ModelConfig{
				Name: "test-model",
			},
		},
		Timeout:            config.Duration(30 * time.Second),
		ConnectionTimeout:  config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
	}

	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	model := c.Model()

	if model == nil {
		t.Fatal("Model() returned nil")
	}

	if model.Name != "test-model" {
		t.Errorf("got model name %q, want %q", model.Name, "test-model")
	}
}

func TestClient_ExecuteProtocol_Chat(t *testing.T) {
	// Create mock server that returns a valid chat response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := types.ChatResponse{
			Model: "test-model",
			Choices: []struct {
				Index        int            `json:"index"`
				Message      types.Message  `json:"message"`
				Delta        *struct {
					Role    string `json:"role,omitempty"`
					Content string `json:"content,omitempty"`
				} `json:"delta,omitempty"`
				FinishReason string `json:"finish_reason,omitempty"`
			}{
				{
					Index:   0,
					Message: types.NewMessage("assistant", "Hello, world!"),
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := &config.ClientConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: server.URL,
			Model: &config.ModelConfig{
				Name: "test-model",
			},
		},
		Timeout:            config.Duration(30 * time.Second),
		ConnectionTimeout:  config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
		Retry: config.RetryConfig{
			MaxRetries: 0, // Disable retry for this test
		},
	}

	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	messages := []types.Message{
		types.NewMessage("user", "Hello"),
	}

	request := &types.ChatRequest{
		Messages: messages,
		Options:  map[string]any{"model": "test-model"},
	}

	result, err := c.ExecuteProtocol(context.Background(), request)
	if err != nil {
		t.Fatalf("ExecuteProtocol failed: %v", err)
	}

	response, ok := result.(*types.ChatResponse)
	if !ok {
		t.Fatalf("expected *types.ChatResponse, got %T", result)
	}

	if response.Content() != "Hello, world!" {
		t.Errorf("got content %q, want %q", response.Content(), "Hello, world!")
	}
}

func TestClient_ExecuteProtocol_Tools(t *testing.T) {
	// Create mock server that returns a valid tools response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := types.ToolsResponse{
			Model: "test-model",
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role      string           `json:"role"`
					Content   string           `json:"content"`
					ToolCalls []types.ToolCall `json:"tool_calls,omitempty"`
				} `json:"message"`
				FinishReason string `json:"finish_reason,omitempty"`
			}{
				{
					Index: 0,
					Message: struct {
						Role      string           `json:"role"`
						Content   string           `json:"content"`
						ToolCalls []types.ToolCall `json:"tool_calls,omitempty"`
					}{
						Role:    "assistant",
						Content: "",
						ToolCalls: []types.ToolCall{
							{
								ID:   "call_123",
								Type: "function",
								Function: types.ToolCallFunction{
									Name:      "get_weather",
									Arguments: `{"location":"Boston"}`,
								},
							},
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := &config.ClientConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: server.URL,
			Model: &config.ModelConfig{
				Name: "test-model",
			},
		},
		Timeout:            config.Duration(30 * time.Second),
		ConnectionTimeout:  config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
		Retry: config.RetryConfig{
			MaxRetries: 0,
		},
	}

	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	messages := []types.Message{
		types.NewMessage("user", "What's the weather in Boston?"),
	}

	request := &types.ToolsRequest{
		Messages: messages,
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
				},
			},
		},
		Options: map[string]any{"model": "test-model"},
	}

	result, err := c.ExecuteProtocol(context.Background(), request)
	if err != nil {
		t.Fatalf("ExecuteProtocol failed: %v", err)
	}

	response, ok := result.(*types.ToolsResponse)
	if !ok {
		t.Fatalf("expected *types.ToolsResponse, got %T", result)
	}

	if len(response.Choices) == 0 {
		t.Fatal("no choices in response")
	}

	if len(response.Choices[0].Message.ToolCalls) == 0 {
		t.Fatal("no tool calls in response")
	}

	toolCall := response.Choices[0].Message.ToolCalls[0]
	if toolCall.Function.Name != "get_weather" {
		t.Errorf("got function name %q, want %q", toolCall.Function.Name, "get_weather")
	}
}

func TestClient_ExecuteProtocol_Embeddings(t *testing.T) {
	// Create mock server that returns a valid embeddings response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := types.EmbeddingsResponse{
			Object: "list",
			Model:  "test-model",
			Data: []struct {
				Embedding []float64 `json:"embedding"`
				Index     int       `json:"index"`
				Object    string    `json:"object"`
			}{
				{
					Embedding: []float64{0.1, 0.2, 0.3},
					Index:     0,
					Object:    "embedding",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := &config.ClientConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: server.URL,
			Model: &config.ModelConfig{
				Name: "test-model",
			},
		},
		Timeout:            config.Duration(30 * time.Second),
		ConnectionTimeout:  config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
		Retry: config.RetryConfig{
			MaxRetries: 0,
		},
	}

	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	request := &types.EmbeddingsRequest{
		Input:   "Hello, world!",
		Options: map[string]any{"model": "test-model"},
	}

	result, err := c.ExecuteProtocol(context.Background(), request)
	if err != nil {
		t.Fatalf("ExecuteProtocol failed: %v", err)
	}

	response, ok := result.(*types.EmbeddingsResponse)
	if !ok {
		t.Fatalf("expected *types.EmbeddingsResponse, got %T", result)
	}

	if len(response.Data) == 0 {
		t.Fatal("no embeddings in response")
	}

	if len(response.Data[0].Embedding) != 3 {
		t.Errorf("got %d dimensions, want 3", len(response.Data[0].Embedding))
	}
}

func TestClient_ExecuteProtocol_HTTPError(t *testing.T) {
	// Create mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	cfg := &config.ClientConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: server.URL,
			Model: &config.ModelConfig{
				Name: "test-model",
			},
		},
		Timeout:            config.Duration(30 * time.Second),
		ConnectionTimeout:  config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
		Retry: config.RetryConfig{
			MaxRetries: 0, // Disable retry to get immediate error
		},
	}

	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	request := &types.ChatRequest{
		Messages: []types.Message{
			types.NewMessage("user", "Hello"),
		},
		Options: map[string]any{"model": "test-model"},
	}

	_, err = c.ExecuteProtocol(context.Background(), request)
	if err == nil {
		t.Error("expected error for HTTP 500, got nil")
	}
}

func TestClient_ExecuteProtocolStream_UnsupportedProtocol(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.ClientConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: server.URL,
			Model: &config.ModelConfig{
				Name: "test-model",
			},
		},
		Timeout:            config.Duration(30 * time.Second),
		ConnectionTimeout:  config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
	}

	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	// Embeddings does not support streaming
	request := &types.EmbeddingsRequest{
		Input:   "test",
		Options: map[string]any{"model": "test-model"},
	}

	_, err = c.ExecuteProtocolStream(context.Background(), request)
	if err == nil {
		t.Error("expected error for unsupported streaming protocol, got nil")
	}
}

func TestClient_IsHealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.ClientConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: server.URL,
			Model: &config.ModelConfig{
				Name: "test-model",
			},
		},
		Timeout:            config.Duration(30 * time.Second),
		ConnectionTimeout:  config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
	}

	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if !c.IsHealthy() {
		t.Error("expected client to be healthy initially")
	}
}

func TestClient_HTTPClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.ClientConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: server.URL,
			Model: &config.ModelConfig{
				Name: "test-model",
			},
		},
		Timeout:            config.Duration(5 * time.Second),
		ConnectionTimeout:  config.Duration(2 * time.Second),
		ConnectionPoolSize: 20,
	}

	c, err := client.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	httpClient := c.HTTPClient()

	if httpClient == nil {
		t.Fatal("HTTPClient() returned nil")
	}

	if httpClient.Timeout != 5*time.Second {
		t.Errorf("got timeout %v, want %v", httpClient.Timeout, 5*time.Second)
	}
}
