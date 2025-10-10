package agent_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JaimeStill/go-agents/pkg/agent"
	"github.com/JaimeStill/go-agents/pkg/config"
)

func TestNew(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.AgentConfig{
		SystemPrompt: "You are a helpful assistant.",
		Transport: &config.TransportConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]config.CapabilityConfig{
						"chat": {
							Format:  "openai-chat",
							Options: map[string]any{},
						},
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
}

func TestNew_InvalidTransport(t *testing.T) {
	cfg := &config.AgentConfig{
		SystemPrompt: "Test",
		Transport: &config.TransportConfig{
			Provider: &config.ProviderConfig{
				Name:    "unknown-provider",
				BaseURL: "http://localhost",
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]config.CapabilityConfig{
						"chat": {
							Format:  "openai-chat",
							Options: map[string]any{},
						},
					},
				},
			},
			Timeout:            config.Duration(30 * time.Second),
			ConnectionTimeout:  config.Duration(10 * time.Second),
			ConnectionPoolSize: 10,
		},
	}

	_, err := agent.New(cfg)

	if err == nil {
		t.Error("expected error for invalid transport, got nil")
	}
}

func TestAgent_Client(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.AgentConfig{
		Transport: &config.TransportConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]config.CapabilityConfig{
						"chat": {
							Format:  "openai-chat",
							Options: map[string]any{},
						},
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
		t.Fatal("Client() returned nil")
	}
}

func TestAgent_Provider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.AgentConfig{
		Transport: &config.TransportConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]config.CapabilityConfig{
						"chat": {
							Format:  "openai-chat",
							Options: map[string]any{},
						},
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
		t.Fatal("Provider() returned nil")
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
		Transport: &config.TransportConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]config.CapabilityConfig{
						"chat": {
							Format:  "openai-chat",
							Options: map[string]any{},
						},
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
		t.Fatal("Model() returned nil")
	}

	if model.Name() != "test-model" {
		t.Errorf("got model name %q, want %q", model.Name(), "test-model")
	}
}

func TestAgent_Chat(t *testing.T) {
	// Create mock server that verifies request and returns response
	requestReceived := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true

		// Parse request to verify structure
		var reqBody map[string]any
		json.NewDecoder(r.Body).Decode(&reqBody)

		// Verify messages were included
		if messages, ok := reqBody["messages"].([]any); ok {
			if len(messages) == 0 {
				t.Error("expected messages in request, got none")
			}
		}

		response := map[string]any{
			"model": "test-model",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": "Hello from agent!",
					},
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := &config.AgentConfig{
		Transport: &config.TransportConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]config.CapabilityConfig{
						"chat": {
							Format:  "openai-chat",
							Options: map[string]any{},
						},
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

	ctx := context.Background()
	response, err := a.Chat(ctx, "Hello")

	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if response == nil {
		t.Fatal("Chat returned nil response")
	}

	if !requestReceived {
		t.Error("request was not sent to server")
	}

	content := response.Content()
	if content != "Hello from agent!" {
		t.Errorf("got content %q, want %q", content, "Hello from agent!")
	}
}

func TestAgent_Chat_WithSystemPrompt(t *testing.T) {
	// Verify system prompt is injected
	var receivedMessages []any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]any
		json.NewDecoder(r.Body).Decode(&reqBody)

		if messages, ok := reqBody["messages"].([]any); ok {
			receivedMessages = messages
		}

		response := map[string]any{
			"model": "test-model",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": "Response",
					},
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := &config.AgentConfig{
		SystemPrompt: "You are a helpful assistant.",
		Transport: &config.TransportConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]config.CapabilityConfig{
						"chat": {
							Format:  "openai-chat",
							Options: map[string]any{},
						},
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

	ctx := context.Background()
	_, err = a.Chat(ctx, "Hello")
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	// Verify system message is first
	if len(receivedMessages) < 2 {
		t.Fatalf("expected at least 2 messages, got %d", len(receivedMessages))
	}

	firstMsg, ok := receivedMessages[0].(map[string]any)
	if !ok {
		t.Fatal("first message is not a map")
	}

	if role := firstMsg["role"]; role != "system" {
		t.Errorf("first message role is %q, want %q", role, "system")
	}

	if content := firstMsg["content"]; content != "You are a helpful assistant." {
		t.Errorf("system prompt is %q, want %q", content, "You are a helpful assistant.")
	}
}

func TestAgent_Chat_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]any
		json.NewDecoder(r.Body).Decode(&reqBody)

		// Verify options were passed
		if temp, ok := reqBody["temperature"].(float64); !ok || temp != 0.9 {
			t.Errorf("expected temperature 0.9, got %v", temp)
		}

		response := map[string]any{
			"model": "test-model",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": "Response",
					},
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := &config.AgentConfig{
		Transport: &config.TransportConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]config.CapabilityConfig{
						"chat": {
							Format:  "openai-chat",
							Options: map[string]any{}, // Initialize to avoid nil map
						},
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

	ctx := context.Background()
	options := map[string]any{
		"temperature": 0.9,
	}
	_, err = a.Chat(ctx, "Hello", options)

	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}
}

func TestAgent_ChatStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		chunks := []string{
			`{"choices":[{"delta":{"content":"Hello"}}]}`,
			`{"choices":[{"delta":{"content":" world"}}]}`,
			`[DONE]`,
		}

		for _, chunk := range chunks {
			fmt.Fprintf(w, "%s\n", chunk)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
	defer server.Close()

	cfg := &config.AgentConfig{
		Transport: &config.TransportConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]config.CapabilityConfig{
						"chat": {
							Format:  "openai-chat",
							Options: map[string]any{},
						},
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

	ctx := context.Background()
	chunks, err := a.ChatStream(ctx, "Hello")

	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}

	if chunks == nil {
		t.Fatal("ChatStream returned nil channel")
	}

	chunkCount := 0
	for range chunks {
		chunkCount++
	}

	if chunkCount == 0 {
		t.Error("expected streaming chunks, got 0")
	}
}

func TestAgent_Vision(t *testing.T) {
	requestReceived := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true

		response := map[string]any{
			"model": "test-model",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": "I see an image",
					},
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := &config.AgentConfig{
		Transport: &config.TransportConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]config.CapabilityConfig{
						"vision": {
							Format:  "openai-vision",
							Options: map[string]any{},
						},
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

	ctx := context.Background()
	images := []string{"image1.jpg", "image2.jpg"}
	response, err := a.Vision(ctx, "Describe these images", images)

	if err != nil {
		t.Fatalf("Vision failed: %v", err)
	}

	if response == nil {
		t.Fatal("Vision returned nil response")
	}

	if !requestReceived {
		t.Error("request was not sent to server")
	}

	content := response.Content()
	if content != "I see an image" {
		t.Errorf("got content %q, want %q", content, "I see an image")
	}
}

func TestAgent_VisionStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		chunks := []string{
			`{"choices":[{"delta":{"content":"I see"}}]}`,
			`{"choices":[{"delta":{"content":" an image"}}]}`,
			`[DONE]`,
		}

		for _, chunk := range chunks {
			fmt.Fprintf(w, "%s\n", chunk)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
	defer server.Close()

	cfg := &config.AgentConfig{
		Transport: &config.TransportConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]config.CapabilityConfig{
						"vision": {
							Format:  "openai-vision",
							Options: map[string]any{},
						},
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

	ctx := context.Background()
	images := []string{"image.jpg"}
	chunks, err := a.VisionStream(ctx, "Describe", images)

	if err != nil {
		t.Fatalf("VisionStream failed: %v", err)
	}

	if chunks == nil {
		t.Fatal("VisionStream returned nil channel")
	}

	chunkCount := 0
	for range chunks {
		chunkCount++
	}

	if chunkCount == 0 {
		t.Error("expected streaming chunks, got 0")
	}
}

func TestAgent_Tools(t *testing.T) {
	var receivedOptions map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]any
		json.NewDecoder(r.Body).Decode(&reqBody)
		receivedOptions = reqBody

		response := map[string]any{
			"model": "test-model",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role": "assistant",
						"tool_calls": []map[string]any{
							{
								"id":   "call_1",
								"type": "function",
								"function": map[string]any{
									"name":      "get_weather",
									"arguments": `{"location":"San Francisco"}`,
								},
							},
						},
					},
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := &config.AgentConfig{
		Transport: &config.TransportConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]config.CapabilityConfig{
						"tools": {
							Format:  "openai-tools",
							Options: map[string]any{},
						},
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

	tools := []agent.Tool{
		{
			Name:        "get_weather",
			Description: "Get weather for a location",
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

	ctx := context.Background()
	response, err := a.Tools(ctx, "What's the weather?", tools)

	if err != nil {
		t.Fatalf("Tools failed: %v", err)
	}

	if response == nil {
		t.Fatal("Tools returned nil response")
	}

	// Verify tools were included in request
	if tools, ok := receivedOptions["tools"].([]any); !ok {
		t.Error("tools not found in request options")
	} else if len(tools) != 1 {
		t.Errorf("got %d tools, want 1", len(tools))
	}
}

func TestAgent_Embed(t *testing.T) {
	var receivedOptions map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]any
		json.NewDecoder(r.Body).Decode(&reqBody)
		receivedOptions = reqBody

		response := map[string]any{
			"model": "test-model",
			"data": []map[string]any{
				{
					"embedding": []float64{0.1, 0.2, 0.3},
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := &config.AgentConfig{
		Transport: &config.TransportConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]config.CapabilityConfig{
						"embeddings": {
							Format:  "openai-embeddings",
							Options: map[string]any{},
						},
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

	ctx := context.Background()
	response, err := a.Embed(ctx, "test text")

	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if response == nil {
		t.Fatal("Embed returned nil response")
	}

	// Verify input was included in request
	if input, ok := receivedOptions["input"].(string); !ok {
		t.Error("input not found in request options")
	} else if input != "test text" {
		t.Errorf("got input %q, want %q", input, "test text")
	}
}

// Orchestration-focused tests for Agent ID behavior

func TestAgent_ID_Uniqueness(t *testing.T) {
	// Test that multiple agents get unique IDs
	// Critical for hub registration without collisions
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.AgentConfig{
		Transport: &config.TransportConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]config.CapabilityConfig{
						"chat": {
							Format:  "openai-chat",
							Options: map[string]any{},
						},
					},
				},
			},
			Timeout:            config.Duration(30 * time.Second),
			ConnectionTimeout:  config.Duration(10 * time.Second),
			ConnectionPoolSize: 10,
		},
	}

	// Create multiple agents
	agents := make([]agent.Agent, 10)
	ids := make(map[string]bool)

	for i := 0; i < 10; i++ {
		a, err := agent.New(cfg)
		if err != nil {
			t.Fatalf("failed to create agent %d: %v", i, err)
		}
		agents[i] = a

		id := a.ID()
		if ids[id] {
			t.Errorf("duplicate ID found: %s", id)
		}
		ids[id] = true
	}

	if len(ids) != 10 {
		t.Errorf("expected 10 unique IDs, got %d", len(ids))
	}
}

func TestAgent_ID_Stability(t *testing.T) {
	// Test that ID() returns the same value across multiple calls
	// Critical for using agent as stable registry key
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.AgentConfig{
		Transport: &config.TransportConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]config.CapabilityConfig{
						"chat": {
							Format:  "openai-chat",
							Options: map[string]any{},
						},
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
		t.Fatalf("failed to create agent: %v", err)
	}

	// Call ID() multiple times
	firstID := a.ID()
	for i := 0; i < 100; i++ {
		id := a.ID()
		if id != firstID {
			t.Errorf("ID changed on call %d: got %s, want %s", i, id, firstID)
		}
	}
}

func TestAgent_ID_Format(t *testing.T) {
	// Test that ID is non-empty and valid UUID format
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.AgentConfig{
		Transport: &config.TransportConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]config.CapabilityConfig{
						"chat": {
							Format:  "openai-chat",
							Options: map[string]any{},
						},
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
		t.Fatalf("failed to create agent: %v", err)
	}

	id := a.ID()

	// Verify non-empty
	if id == "" {
		t.Error("ID is empty")
	}

	// Verify valid UUID format
	// UUIDs are 36 characters: 8-4-4-4-12
	if len(id) != 36 {
		t.Errorf("ID length is %d, want 36 (standard UUID format)", len(id))
	}

	// Verify contains hyphens in correct positions
	if id[8] != '-' || id[13] != '-' || id[18] != '-' || id[23] != '-' {
		t.Errorf("ID does not match UUID format: %s", id)
	}
}

func TestAgent_ID_Concurrent(t *testing.T) {
	// Test that ID() is thread-safe for concurrent access
	// Critical for hub usage with concurrent goroutines
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.AgentConfig{
		Transport: &config.TransportConfig{
			Provider: &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: server.URL,
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]config.CapabilityConfig{
						"chat": {
							Format:  "openai-chat",
							Options: map[string]any{},
						},
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
		t.Fatalf("failed to create agent: %v", err)
	}

	expectedID := a.ID()

	// Access ID concurrently from multiple goroutines
	const numGoroutines = 100
	done := make(chan bool, numGoroutines)
	errors := make(chan string, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			id := a.ID()
			if id != expectedID {
				errors <- fmt.Sprintf("goroutine %d got ID %s, want %s", goroutineID, id, expectedID)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	close(errors)

	// Check for any errors
	for err := range errors {
		t.Error(err)
	}
}
