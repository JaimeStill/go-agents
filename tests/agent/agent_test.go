package agent_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JaimeStill/go-agents/pkg/agent"
	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/types"
)

func TestNew(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.AgentConfig{
		Name:         "test-agent",
		SystemPrompt: "You are a helpful assistant.",
		Client: &config.ClientConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]map[string]any{
						"chat": {"temperature": 0.7},
					},
				},
			},
			Timeout:            config.Duration(30 * time.Second),
			ConnectionTimeout:  config.Duration(10 * time.Second),
			ConnectionPoolSize: 10,
		},
	}

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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.AgentConfig{
		Name: "test-agent",
		Client: &config.ClientConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]map[string]any{
						"chat": {},
					},
				},
			},
			Timeout:            config.Duration(30 * time.Second),
			ConnectionTimeout:  config.Duration(10 * time.Second),
			ConnectionPoolSize: 10,
		},
	}

	a1, _ := agent.New(cfg)
	a2, _ := agent.New(cfg)

	if a1.ID() == a2.ID() {
		t.Error("two agents have the same ID")
	}

	// ID should be stable
	id1 := a1.ID()
	id2 := a1.ID()

	if id1 != id2 {
		t.Error("agent ID changed between calls")
	}
}

func TestAgent_Chat(t *testing.T) {
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
					Message: types.NewMessage("assistant", "Hello, how can I help you?"),
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := &config.AgentConfig{
		Name:         "test-agent",
		SystemPrompt: "You are helpful.",
		Client: &config.ClientConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]map[string]any{
						"chat": {},
					},
				},
			},
			Timeout:            config.Duration(30 * time.Second),
			ConnectionTimeout:  config.Duration(10 * time.Second),
			ConnectionPoolSize: 10,
			Retry: config.RetryConfig{
				MaxRetries: 0,
			},
		},
	}

	a, err := agent.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	response, err := a.Chat(context.Background(), "Hello")
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if response == nil {
		t.Fatal("Chat returned nil response")
	}

	if response.Content() != "Hello, how can I help you?" {
		t.Errorf("got content %q, want %q", response.Content(), "Hello, how can I help you?")
	}
}

func TestAgent_Vision(t *testing.T) {
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
					Message: types.NewMessage("assistant", "I see a cat in the image."),
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := &config.AgentConfig{
		Name: "test-agent",
		Client: &config.ClientConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]map[string]any{
						"vision": {},
					},
				},
			},
			Timeout:            config.Duration(30 * time.Second),
			ConnectionTimeout:  config.Duration(10 * time.Second),
			ConnectionPoolSize: 10,
			Retry: config.RetryConfig{
				MaxRetries: 0,
			},
		},
	}

	a, err := agent.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	images := []string{"data:image/png;base64,iVBORw0KGgoAAAANSUhEUg=="}
	response, err := a.Vision(context.Background(), "What's in this image?", images)
	if err != nil {
		t.Fatalf("Vision failed: %v", err)
	}

	if response == nil {
		t.Fatal("Vision returned nil response")
	}

	if response.Content() != "I see a cat in the image." {
		t.Errorf("got content %q, want %q", response.Content(), "I see a cat in the image.")
	}
}

func TestAgent_Tools(t *testing.T) {
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

	cfg := &config.AgentConfig{
		Name: "test-agent",
		Client: &config.ClientConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]map[string]any{
						"tools": {},
					},
				},
			},
			Timeout:            config.Duration(30 * time.Second),
			ConnectionTimeout:  config.Duration(10 * time.Second),
			ConnectionPoolSize: 10,
			Retry: config.RetryConfig{
				MaxRetries: 0,
			},
		},
	}

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

	response, err := a.Tools(context.Background(), "What's the weather in Boston?", tools)
	if err != nil {
		t.Fatalf("Tools failed: %v", err)
	}

	if response == nil {
		t.Fatal("Tools returned nil response")
	}

	if len(response.Choices) == 0 {
		t.Fatal("response has no choices")
	}

	if len(response.Choices[0].Message.ToolCalls) == 0 {
		t.Fatal("response has no tool calls")
	}

	toolCall := response.Choices[0].Message.ToolCalls[0]
	if toolCall.Function.Name != "get_weather" {
		t.Errorf("got function name %q, want %q", toolCall.Function.Name, "get_weather")
	}
}

func TestAgent_Embed(t *testing.T) {
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

	cfg := &config.AgentConfig{
		Name: "test-agent",
		Client: &config.ClientConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]map[string]any{
						"embeddings": {},
					},
				},
			},
			Timeout:            config.Duration(30 * time.Second),
			ConnectionTimeout:  config.Duration(10 * time.Second),
			ConnectionPoolSize: 10,
			Retry: config.RetryConfig{
				MaxRetries: 0,
			},
		},
	}

	a, err := agent.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	response, err := a.Embed(context.Background(), "Hello, world!")
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if response == nil {
		t.Fatal("Embed returned nil response")
	}

	if len(response.Data) == 0 {
		t.Fatal("response has no embeddings")
	}

	if len(response.Data[0].Embedding) != 3 {
		t.Errorf("got %d dimensions, want 3", len(response.Data[0].Embedding))
	}
}

func TestAgent_Client(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.AgentConfig{
		Name: "test-agent",
		Client: &config.ClientConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]map[string]any{
						"chat": {},
					},
				},
			},
			Timeout:            config.Duration(30 * time.Second),
			ConnectionTimeout:  config.Duration(10 * time.Second),
			ConnectionPoolSize: 10,
		},
	}

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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.AgentConfig{
		Name: "test-agent",
		Client: &config.ClientConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]map[string]any{
						"chat": {},
					},
				},
			},
			Timeout:            config.Duration(30 * time.Second),
			ConnectionTimeout:  config.Duration(10 * time.Second),
			ConnectionPoolSize: 10,
		},
	}

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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.AgentConfig{
		Name: "test-agent",
		Client: &config.ClientConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]map[string]any{
						"chat": {},
					},
				},
			},
			Timeout:            config.Duration(30 * time.Second),
			ConnectionTimeout:  config.Duration(10 * time.Second),
			ConnectionPoolSize: 10,
		},
	}

	a, err := agent.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	model := a.Model()

	if model == nil {
		t.Error("Model() returned nil")
	}

	if model.Name != "test-model" {
		t.Errorf("got model name %q, want %q", model.Name, "test-model")
	}
}
