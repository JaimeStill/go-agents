package providers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/protocols"
	"github.com/JaimeStill/go-agents/pkg/providers"
)

func TestNewOllama(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: "http://localhost:11434",
		Model: &config.ModelConfig{
			Name: "llama2",
			Capabilities: map[string]config.CapabilityConfig{
				"chat": {Format: "openai-chat"},
			},
		},
	}

	provider, err := providers.NewOllama(cfg)

	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	if provider == nil {
		t.Fatal("NewOllama returned nil provider")
	}

	if provider.Name() != "ollama" {
		t.Errorf("got name %q, want %q", provider.Name(), "ollama")
	}
}

func TestNewOllama_URLSuffixHandling(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		expectedURL string
	}{
		{
			name:        "URL without /v1",
			baseURL:     "http://localhost:11434",
			expectedURL: "http://localhost:11434/v1",
		},
		{
			name:        "URL with /v1",
			baseURL:     "http://localhost:11434/v1",
			expectedURL: "http://localhost:11434/v1",
		},
		{
			name:        "URL with trailing slash",
			baseURL:     "http://localhost:11434/",
			expectedURL: "http://localhost:11434/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: tt.baseURL,
				Model: &config.ModelConfig{
					Name: "llama2",
					Capabilities: map[string]config.CapabilityConfig{
						"chat": {Format: "openai-chat"},
					},
				},
			}

			provider, err := providers.NewOllama(cfg)
			if err != nil {
				t.Fatalf("NewOllama failed: %v", err)
			}

			// Access baseURL through GetEndpoint to verify it was set correctly
			endpoint, err := provider.GetEndpoint(protocols.Chat)
			if err != nil {
				t.Fatalf("GetEndpoint failed: %v", err)
			}

			if !strings.HasPrefix(endpoint, tt.expectedURL) {
				t.Errorf("got endpoint %q, expected to start with %q", endpoint, tt.expectedURL)
			}
		})
	}
}

func TestOllama_GetEndpoint(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: "http://localhost:11434",
		Model: &config.ModelConfig{
			Name: "llama2",
			Capabilities: map[string]config.CapabilityConfig{
				"chat":       {Format: "openai-chat"},
				"vision":     {Format: "openai-vision"},
				"tools":      {Format: "openai-tools"},
				"embeddings": {Format: "openai-embeddings"},
			},
		},
	}

	provider, err := providers.NewOllama(cfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	tests := []struct {
		protocol protocols.Protocol
		expected string
	}{
		{protocols.Chat, "http://localhost:11434/v1/chat/completions"},
		{protocols.Vision, "http://localhost:11434/v1/chat/completions"},
		{protocols.Tools, "http://localhost:11434/v1/chat/completions"},
		{protocols.Embeddings, "http://localhost:11434/v1/embeddings"},
	}

	for _, tt := range tests {
		t.Run(string(tt.protocol), func(t *testing.T) {
			endpoint, err := provider.GetEndpoint(tt.protocol)

			if err != nil {
				t.Fatalf("GetEndpoint failed: %v", err)
			}

			if endpoint != tt.expected {
				t.Errorf("got endpoint %q, want %q", endpoint, tt.expected)
			}
		})
	}
}

func TestOllama_PrepareRequest(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: "http://localhost:11434",
		Model: &config.ModelConfig{
			Name: "llama2",
			Capabilities: map[string]config.CapabilityConfig{
				"chat": {Format: "openai-chat"},
			},
		},
	}

	provider, err := providers.NewOllama(cfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	capability := capabilities.NewChatCapability("openai-chat", nil)

	capabilityReq := &capabilities.CapabilityRequest{
		Protocol: protocols.Chat,
		Messages: []protocols.Message{
			protocols.NewMessage("user", "Hello"),
		},
		Options: map[string]any{},
	}

	protocolReq, err := capability.CreateRequest(capabilityReq, "llama2")
	if err != nil {
		t.Fatalf("CreateRequest failed: %v", err)
	}

	request, err := provider.PrepareRequest(context.Background(), protocols.Chat, protocolReq)

	if err != nil {
		t.Fatalf("PrepareRequest failed: %v", err)
	}

	if request == nil {
		t.Fatal("PrepareRequest returned nil request")
	}

	if request.URL != "http://localhost:11434/v1/chat/completions" {
		t.Errorf("got URL %q, want %q", request.URL, "http://localhost:11434/v1/chat/completions")
	}

	if len(request.Body) == 0 {
		t.Error("request body is empty")
	}

	if request.Headers == nil {
		t.Error("request headers is nil")
	}
}

func TestOllama_PrepareStreamRequest(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: "http://localhost:11434",
		Model: &config.ModelConfig{
			Name: "llama2",
			Capabilities: map[string]config.CapabilityConfig{
				"chat": {Format: "openai-chat"},
			},
		},
	}

	provider, err := providers.NewOllama(cfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	capability := capabilities.NewChatCapability("openai-chat", nil)

	capabilityReq := &capabilities.CapabilityRequest{
		Protocol: protocols.Chat,
		Messages: []protocols.Message{
			protocols.NewMessage("user", "Hello"),
		},
		Options: map[string]any{},
	}

	protocolReq, err := capability.CreateRequest(capabilityReq, "llama2")
	if err != nil {
		t.Fatalf("CreateRequest failed: %v", err)
	}

	request, err := provider.PrepareStreamRequest(context.Background(), protocols.Chat, protocolReq)

	if err != nil {
		t.Fatalf("PrepareStreamRequest failed: %v", err)
	}

	if request == nil {
		t.Fatal("PrepareStreamRequest returned nil request")
	}

	if request.URL != "http://localhost:11434/v1/chat/completions" {
		t.Errorf("got URL %q, want %q", request.URL, "http://localhost:11434/v1/chat/completions")
	}

	if request.Headers["Accept"] != "text/event-stream" {
		t.Errorf("got Accept header %q, want %q", request.Headers["Accept"], "text/event-stream")
	}

	if request.Headers["Cache-Control"] != "no-cache" {
		t.Errorf("got Cache-Control header %q, want %q", request.Headers["Cache-Control"], "no-cache")
	}
}

func TestOllama_SetHeaders_Bearer(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: "http://localhost:11434",
		Model: &config.ModelConfig{
			Name: "llama2",
			Capabilities: map[string]config.CapabilityConfig{
				"chat": {Format: "openai-chat"},
			},
		},
		Options: map[string]any{
			"auth_type": "bearer",
			"token":     "test-token-123",
		},
	}

	provider, err := providers.NewOllama(cfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	req := httptest.NewRequest("POST", "http://test.com", nil)
	provider.SetHeaders(req)

	authHeader := req.Header.Get("Authorization")
	expected := "Bearer test-token-123"

	if authHeader != expected {
		t.Errorf("got Authorization header %q, want %q", authHeader, expected)
	}
}

func TestOllama_SetHeaders_APIKey(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: "http://localhost:11434",
		Model: &config.ModelConfig{
			Name: "llama2",
			Capabilities: map[string]config.CapabilityConfig{
				"chat": {Format: "openai-chat"},
			},
		},
		Options: map[string]any{
			"auth_type": "api_key",
			"token":     "test-api-key",
		},
	}

	provider, err := providers.NewOllama(cfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	req := httptest.NewRequest("POST", "http://test.com", nil)
	provider.SetHeaders(req)

	apiKeyHeader := req.Header.Get("X-API-Key")
	expected := "test-api-key"

	if apiKeyHeader != expected {
		t.Errorf("got X-API-Key header %q, want %q", apiKeyHeader, expected)
	}
}

func TestOllama_SetHeaders_CustomAPIKeyHeader(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: "http://localhost:11434",
		Model: &config.ModelConfig{
			Name: "llama2",
			Capabilities: map[string]config.CapabilityConfig{
				"chat": {Format: "openai-chat"},
			},
		},
		Options: map[string]any{
			"auth_type":   "api_key",
			"token":       "custom-key",
			"auth_header": "X-Custom-Auth",
		},
	}

	provider, err := providers.NewOllama(cfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	req := httptest.NewRequest("POST", "http://test.com", nil)
	provider.SetHeaders(req)

	customHeader := req.Header.Get("X-Custom-Auth")
	expected := "custom-key"

	if customHeader != expected {
		t.Errorf("got X-Custom-Auth header %q, want %q", customHeader, expected)
	}
}

func TestOllama_ProcessResponse(t *testing.T) {
	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]any{
			"model": "llama2",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": "Hello from Ollama!",
					},
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Make request to mock server
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Create provider and capability
	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: server.URL,
		Model: &config.ModelConfig{
			Name: "llama2",
			Capabilities: map[string]config.CapabilityConfig{
				"chat": {Format: "openai-chat"},
			},
		},
	}

	provider, err := providers.NewOllama(cfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	capability := capabilities.NewChatCapability("openai-chat", nil)

	// Process response
	result, err := provider.ProcessResponse(resp, capability)

	if err != nil {
		t.Fatalf("ProcessResponse failed: %v", err)
	}

	if result == nil {
		t.Fatal("ProcessResponse returned nil result")
	}
}

func TestOllama_ProcessResponse_ErrorStatus(t *testing.T) {
	// Create mock HTTP server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request"))
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: server.URL,
		Model: &config.ModelConfig{
			Name: "llama2",
			Capabilities: map[string]config.CapabilityConfig{
				"chat": {Format: "openai-chat"},
			},
		},
	}

	provider, err := providers.NewOllama(cfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	capability := capabilities.NewChatCapability("openai-chat", nil)

	_, err = provider.ProcessResponse(resp, capability)

	if err == nil {
		t.Error("expected error for non-OK status, got nil")
	}
}

func TestOllama_ProcessStreamResponse(t *testing.T) {
	// Create mock streaming server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/event-stream")

		// Send streaming chunks
		chunks := []string{
			`{"choices":[{"delta":{"content":"Hello"}}]}`,
			`{"choices":[{"delta":{"content":" world"}}]}`,
			`[DONE]`,
		}

		for _, chunk := range chunks {
			w.Write([]byte(chunk + "\n"))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: server.URL,
		Model: &config.ModelConfig{
			Name: "llama2",
			Capabilities: map[string]config.CapabilityConfig{
				"chat": {Format: "openai-chat"},
			},
		},
	}

	provider, err := providers.NewOllama(cfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	capability := capabilities.NewChatCapability("openai-chat", nil)

	ctx := context.Background()
	outputChan, err := provider.ProcessStreamResponse(ctx, resp, capability)

	if err != nil {
		t.Fatalf("ProcessStreamResponse failed: %v", err)
	}

	if outputChan == nil {
		t.Fatal("ProcessStreamResponse returned nil channel")
	}

	// Read chunks from channel
	chunkCount := 0
	for range outputChan {
		chunkCount++
	}

	if chunkCount == 0 {
		t.Error("expected to receive streaming chunks, got 0")
	}
}
