package transport_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/protocols"
	"github.com/JaimeStill/go-agents/pkg/transport"
)

func TestNew(t *testing.T) {
	// Create mock server for Ollama
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.TransportConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: server.URL,
			Model: &config.ModelConfig{
				Name: "test-model",
				Capabilities: map[string]config.CapabilityConfig{
					"chat": {Format: "openai-chat"},
				},
			},
		},
		Timeout:           config.Duration(30 * time.Second),
		ConnectionTimeout: config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
	}

	client, err := transport.New(cfg)

	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if client == nil {
		t.Fatal("New returned nil client")
	}
}

func TestNew_InvalidProvider(t *testing.T) {
	cfg := &config.TransportConfig{
		Provider: &config.ProviderConfig{
			Name:    "unknown-provider",
			BaseURL: "http://localhost",
			Model: &config.ModelConfig{
				Name: "test-model",
				Capabilities: map[string]config.CapabilityConfig{
					"chat": {Format: "openai-chat"},
				},
			},
		},
		Timeout:           config.Duration(30 * time.Second),
		ConnectionTimeout: config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
	}

	_, err := transport.New(cfg)

	if err == nil {
		t.Error("expected error for unknown provider, got nil")
	}
}

func TestClient_Provider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.TransportConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: server.URL,
			Model: &config.ModelConfig{
				Name: "test-model",
				Capabilities: map[string]config.CapabilityConfig{
					"chat": {Format: "openai-chat"},
				},
			},
		},
		Timeout:           config.Duration(30 * time.Second),
		ConnectionTimeout: config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
	}

	client, err := transport.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	provider := client.Provider()

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

	cfg := &config.TransportConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: server.URL,
			Model: &config.ModelConfig{
				Name: "test-model",
				Capabilities: map[string]config.CapabilityConfig{
					"chat": {Format: "openai-chat"},
				},
			},
		},
		Timeout:           config.Duration(30 * time.Second),
		ConnectionTimeout: config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
	}

	client, err := transport.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	model := client.Model()

	if model == nil {
		t.Fatal("Model() returned nil")
	}

	if model.Name() != "test-model" {
		t.Errorf("got model name %q, want %q", model.Name(), "test-model")
	}
}

func TestClient_HTTPClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.TransportConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: server.URL,
			Model: &config.ModelConfig{
				Name: "test-model",
				Capabilities: map[string]config.CapabilityConfig{
					"chat": {Format: "openai-chat"},
				},
			},
		},
		Timeout:           config.Duration(30 * time.Second),
		ConnectionTimeout: config.Duration(10 * time.Second),
		ConnectionPoolSize: 5,
	}

	client, err := transport.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	httpClient := client.HTTPClient()

	if httpClient == nil {
		t.Fatal("HTTPClient() returned nil")
	}

	if httpClient.Timeout != 30*time.Second {
		t.Errorf("got timeout %v, want %v", httpClient.Timeout, 30*time.Second)
	}

	if transport, ok := httpClient.Transport.(*http.Transport); ok {
		if transport.MaxIdleConns != 5 {
			t.Errorf("got MaxIdleConns %d, want 5", transport.MaxIdleConns)
		}
		if transport.MaxIdleConnsPerHost != 5 {
			t.Errorf("got MaxIdleConnsPerHost %d, want 5", transport.MaxIdleConnsPerHost)
		}
		if transport.IdleConnTimeout != 10*time.Second {
			t.Errorf("got IdleConnTimeout %v, want %v", transport.IdleConnTimeout, 10*time.Second)
		}
	} else {
		t.Error("HTTPClient.Transport is not *http.Transport")
	}
}

func TestClient_ExecuteProtocol(t *testing.T) {
	// Create mock server that returns valid chat response
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		response := map[string]any{
			"model": "test-model",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": "Hello from test!",
					},
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := &config.TransportConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: server.URL,
			Model: &config.ModelConfig{
				Name: "test-model",
				Capabilities: map[string]config.CapabilityConfig{
					"chat": {
						Format: "openai-chat",
						Options: map[string]any{
							"temperature": 0.7,
						},
					},
				},
			},
		},
		Timeout:           config.Duration(30 * time.Second),
		ConnectionTimeout: config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
	}

	client, err := transport.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Chat,
		Messages: []protocols.Message{
			protocols.NewMessage("user", "Hello"),
		},
		Options: map[string]any{},
	}

	ctx := context.Background()
	result, err := client.ExecuteProtocol(ctx, req)

	if err != nil {
		t.Fatalf("ExecuteProtocol failed: %v", err)
	}

	if result == nil {
		t.Fatal("ExecuteProtocol returned nil result")
	}

	if requestCount != 1 {
		t.Errorf("expected 1 HTTP request, got %d", requestCount)
	}
}

func TestClient_ExecuteProtocol_OptionMerging(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse the request to verify options were merged
		var reqBody map[string]any
		json.NewDecoder(r.Body).Decode(&reqBody)

		// Verify temperature from request options (0.9) overrides model default (0.7)
		if temp, ok := reqBody["temperature"].(float64); ok {
			if temp != 0.9 {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "expected temperature 0.9, got %v", temp)
				return
			}
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

	cfg := &config.TransportConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: server.URL,
			Model: &config.ModelConfig{
				Name: "test-model",
				Capabilities: map[string]config.CapabilityConfig{
					"chat": {
						Format: "openai-chat",
						Options: map[string]any{
							"temperature": 0.7,
						},
					},
				},
			},
		},
		Timeout:           config.Duration(30 * time.Second),
		ConnectionTimeout: config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
	}

	client, err := transport.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Chat,
		Messages: []protocols.Message{
			protocols.NewMessage("user", "Hello"),
		},
		Options: map[string]any{
			"temperature": 0.9, // Override model default
		},
	}

	ctx := context.Background()
	_, err = client.ExecuteProtocol(ctx, req)

	if err != nil {
		t.Fatalf("ExecuteProtocol failed: %v", err)
	}
}

func TestClient_ExecuteProtocol_UnsupportedProtocol(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.TransportConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: server.URL,
			Model: &config.ModelConfig{
				Name: "test-model",
				Capabilities: map[string]config.CapabilityConfig{
					"chat": {Format: "openai-chat"},
				},
			},
		},
		Timeout:           config.Duration(30 * time.Second),
		ConnectionTimeout: config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
	}

	client, err := transport.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Tools, // Not supported by model
		Messages: []protocols.Message{
			protocols.NewMessage("user", "Hello"),
		},
		Options: map[string]any{},
	}

	ctx := context.Background()
	_, err = client.ExecuteProtocol(ctx, req)

	if err == nil {
		t.Error("expected error for unsupported protocol, got nil")
	}
}

func TestClient_ExecuteProtocol_InvalidOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.TransportConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: server.URL,
			Model: &config.ModelConfig{
				Name: "test-model",
				Capabilities: map[string]config.CapabilityConfig{
					"chat": {
						Format:  "openai-chat",
						Options: map[string]any{}, // Initialize to empty map to avoid nil issues
					},
				},
			},
		},
		Timeout:           config.Duration(30 * time.Second),
		ConnectionTimeout: config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
	}

	client, err := transport.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Chat,
		Messages: []protocols.Message{
			protocols.NewMessage("user", "Hello"),
		},
		Options: map[string]any{
			"invalid_option": "value",
		},
	}

	ctx := context.Background()
	_, err = client.ExecuteProtocol(ctx, req)

	if err == nil {
		t.Error("expected error for invalid options, got nil")
	}
}

func TestClient_ExecuteProtocolStream(t *testing.T) {
	// Create mock streaming server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		// Send streaming chunks
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

	cfg := &config.TransportConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: server.URL,
			Model: &config.ModelConfig{
				Name: "test-model",
				Capabilities: map[string]config.CapabilityConfig{
					"chat": {Format: "openai-chat"},
				},
			},
		},
		Timeout:           config.Duration(30 * time.Second),
		ConnectionTimeout: config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
	}

	client, err := transport.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Chat,
		Messages: []protocols.Message{
			protocols.NewMessage("user", "Hello"),
		},
		Options: map[string]any{},
	}

	ctx := context.Background()
	chunks, err := client.ExecuteProtocolStream(ctx, req)

	if err != nil {
		t.Fatalf("ExecuteProtocolStream failed: %v", err)
	}

	if chunks == nil {
		t.Fatal("ExecuteProtocolStream returned nil channel")
	}

	// Read chunks
	chunkCount := 0
	for range chunks {
		chunkCount++
	}

	if chunkCount == 0 {
		t.Error("expected to receive streaming chunks, got 0")
	}
}

func TestClient_ExecuteProtocolStream_UnsupportedStreaming(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.TransportConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: server.URL,
			Model: &config.ModelConfig{
				Name: "test-model",
				Capabilities: map[string]config.CapabilityConfig{
					"embeddings": {Format: "openai-embeddings"}, // Doesn't support streaming
				},
			},
		},
		Timeout:           config.Duration(30 * time.Second),
		ConnectionTimeout: config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
	}

	client, err := transport.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Embeddings,
		Messages: []protocols.Message{},
		Options: map[string]any{
			"input": "test text",
		},
	}

	ctx := context.Background()
	_, err = client.ExecuteProtocolStream(ctx, req)

	if err == nil {
		t.Error("expected error for unsupported streaming, got nil")
	}
}

func TestClient_IsHealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	cfg := &config.TransportConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: server.URL,
			Model: &config.ModelConfig{
				Name: "test-model",
				Capabilities: map[string]config.CapabilityConfig{
					"chat": {Format: "openai-chat"},
				},
			},
		},
		Timeout:           config.Duration(30 * time.Second),
		ConnectionTimeout: config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
	}

	client, err := transport.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	// Client should be healthy initially
	if !client.IsHealthy() {
		t.Error("client should be healthy initially")
	}

	// Execute successful request
	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Chat,
		Messages: []protocols.Message{
			protocols.NewMessage("user", "Hello"),
		},
		Options: map[string]any{},
	}

	ctx := context.Background()
	_, err = client.ExecuteProtocol(ctx, req)
	if err != nil {
		t.Fatalf("ExecuteProtocol failed: %v", err)
	}

	// Client should still be healthy
	if !client.IsHealthy() {
		t.Error("client should be healthy after successful request")
	}
}

func TestClient_IsHealthy_AfterFailure(t *testing.T) {
	// Create server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	}))
	defer server.Close()

	cfg := &config.TransportConfig{
		Provider: &config.ProviderConfig{
			Name:    "ollama",
			BaseURL: server.URL,
			Model: &config.ModelConfig{
				Name: "test-model",
				Capabilities: map[string]config.CapabilityConfig{
					"chat": {Format: "openai-chat"},
				},
			},
		},
		Timeout:           config.Duration(30 * time.Second),
		ConnectionTimeout: config.Duration(10 * time.Second),
		ConnectionPoolSize: 10,
	}

	client, err := transport.New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	// Client should be healthy initially
	if !client.IsHealthy() {
		t.Error("client should be healthy initially")
	}

	// Execute request that will fail
	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Chat,
		Messages: []protocols.Message{
			protocols.NewMessage("user", "Hello"),
		},
		Options: map[string]any{},
	}

	ctx := context.Background()
	_, err = client.ExecuteProtocol(ctx, req)

	if err == nil {
		t.Error("expected error from failed request")
	}

	// Client should be unhealthy after failure
	if client.IsHealthy() {
		t.Error("client should be unhealthy after failed request")
	}
}
