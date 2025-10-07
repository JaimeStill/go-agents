package providers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/protocols"
	"github.com/JaimeStill/go-agents/pkg/providers"
)

func TestNewAzure(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Model: &config.ModelConfig{
			Name: "gpt-4",
			Capabilities: map[string]config.CapabilityConfig{
				"chat": {Format: "openai-chat"},
			},
		},
		Options: map[string]any{
			"deployment":  "gpt-4-deployment",
			"auth_type":   "api_key",
			"token":       "test-key",
			"api_version": "2024-02-01",
		},
	}

	provider, err := providers.NewAzure(cfg)

	if err != nil {
		t.Fatalf("NewAzure failed: %v", err)
	}

	if provider == nil {
		t.Fatal("NewAzure returned nil provider")
	}

	if provider.Name() != "azure" {
		t.Errorf("got name %q, want %q", provider.Name(), "azure")
	}
}

func TestNewAzure_MissingDeployment(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Model: &config.ModelConfig{
			Name: "gpt-4",
			Capabilities: map[string]config.CapabilityConfig{
				"chat": {Format: "openai-chat"},
			},
		},
		Options: map[string]any{
			"auth_type":   "api_key",
			"token":       "test-key",
			"api_version": "2024-02-01",
		},
	}

	_, err := providers.NewAzure(cfg)

	if err == nil {
		t.Error("expected error for missing deployment, got nil")
	}
}

func TestNewAzure_MissingAuthType(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Model: &config.ModelConfig{
			Name: "gpt-4",
			Capabilities: map[string]config.CapabilityConfig{
				"chat": {Format: "openai-chat"},
			},
		},
		Options: map[string]any{
			"deployment":  "gpt-4-deployment",
			"token":       "test-key",
			"api_version": "2024-02-01",
		},
	}

	_, err := providers.NewAzure(cfg)

	if err == nil {
		t.Error("expected error for missing auth_type, got nil")
	}
}

func TestNewAzure_MissingToken(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Model: &config.ModelConfig{
			Name: "gpt-4",
			Capabilities: map[string]config.CapabilityConfig{
				"chat": {Format: "openai-chat"},
			},
		},
		Options: map[string]any{
			"deployment":  "gpt-4-deployment",
			"auth_type":   "api_key",
			"api_version": "2024-02-01",
		},
	}

	_, err := providers.NewAzure(cfg)

	if err == nil {
		t.Error("expected error for missing token, got nil")
	}
}

func TestNewAzure_MissingAPIVersion(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Model: &config.ModelConfig{
			Name: "gpt-4",
			Capabilities: map[string]config.CapabilityConfig{
				"chat": {Format: "openai-chat"},
			},
		},
		Options: map[string]any{
			"deployment": "gpt-4-deployment",
			"auth_type":  "api_key",
			"token":      "test-key",
		},
	}

	_, err := providers.NewAzure(cfg)

	if err == nil {
		t.Error("expected error for missing api_version, got nil")
	}
}

func TestAzure_GetEndpoint(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Model: &config.ModelConfig{
			Name: "gpt-4",
			Capabilities: map[string]config.CapabilityConfig{
				"chat":       {Format: "openai-chat"},
				"vision":     {Format: "openai-vision"},
				"tools":      {Format: "openai-tools"},
				"embeddings": {Format: "openai-embeddings"},
			},
		},
		Options: map[string]any{
			"deployment":  "gpt-4-deployment",
			"auth_type":   "api_key",
			"token":       "test-key",
			"api_version": "2024-02-01",
		},
	}

	provider, err := providers.NewAzure(cfg)
	if err != nil {
		t.Fatalf("NewAzure failed: %v", err)
	}

	tests := []struct {
		protocol protocols.Protocol
		expected string
	}{
		{
			protocols.Chat,
			"https://my-resource.openai.azure.com/deployments/gpt-4-deployment/chat/completions?api-version=2024-02-01",
		},
		{
			protocols.Vision,
			"https://my-resource.openai.azure.com/deployments/gpt-4-deployment/chat/completions?api-version=2024-02-01",
		},
		{
			protocols.Tools,
			"https://my-resource.openai.azure.com/deployments/gpt-4-deployment/chat/completions?api-version=2024-02-01",
		},
		{
			protocols.Embeddings,
			"https://my-resource.openai.azure.com/deployments/gpt-4-deployment/embeddings?api-version=2024-02-01",
		},
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

func TestAzure_PrepareRequest(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Model: &config.ModelConfig{
			Name: "gpt-4",
			Capabilities: map[string]config.CapabilityConfig{
				"chat": {Format: "openai-chat"},
			},
		},
		Options: map[string]any{
			"deployment":  "gpt-4-deployment",
			"auth_type":   "api_key",
			"token":       "test-key",
			"api_version": "2024-02-01",
		},
	}

	provider, err := providers.NewAzure(cfg)
	if err != nil {
		t.Fatalf("NewAzure failed: %v", err)
	}

	capability := capabilities.NewChatCapability("openai-chat", nil)

	capabilityReq := &capabilities.CapabilityRequest{
		Protocol: protocols.Chat,
		Messages: []protocols.Message{
			protocols.NewMessage("user", "Hello"),
		},
		Options: map[string]any{},
	}

	protocolReq, err := capability.CreateRequest(capabilityReq, "gpt-4")
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

	expectedURL := "https://my-resource.openai.azure.com/deployments/gpt-4-deployment/chat/completions?api-version=2024-02-01"
	if request.URL != expectedURL {
		t.Errorf("got URL %q, want %q", request.URL, expectedURL)
	}

	if len(request.Body) == 0 {
		t.Error("request body is empty")
	}

	if request.Headers == nil {
		t.Error("request headers is nil")
	}
}

func TestAzure_PrepareStreamRequest(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Model: &config.ModelConfig{
			Name: "gpt-4",
			Capabilities: map[string]config.CapabilityConfig{
				"chat": {Format: "openai-chat"},
			},
		},
		Options: map[string]any{
			"deployment":  "gpt-4-deployment",
			"auth_type":   "api_key",
			"token":       "test-key",
			"api_version": "2024-02-01",
		},
	}

	provider, err := providers.NewAzure(cfg)
	if err != nil {
		t.Fatalf("NewAzure failed: %v", err)
	}

	capability := capabilities.NewChatCapability("openai-chat", nil)

	capabilityReq := &capabilities.CapabilityRequest{
		Protocol: protocols.Chat,
		Messages: []protocols.Message{
			protocols.NewMessage("user", "Hello"),
		},
		Options: map[string]any{},
	}

	protocolReq, err := capability.CreateRequest(capabilityReq, "gpt-4")
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

	if request.Headers["Accept"] != "text/event-stream" {
		t.Errorf("got Accept header %q, want %q", request.Headers["Accept"], "text/event-stream")
	}

	if request.Headers["Cache-Control"] != "no-cache" {
		t.Errorf("got Cache-Control header %q, want %q", request.Headers["Cache-Control"], "no-cache")
	}
}

func TestAzure_SetHeaders_APIKey(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Model: &config.ModelConfig{
			Name: "gpt-4",
			Capabilities: map[string]config.CapabilityConfig{
				"chat": {Format: "openai-chat"},
			},
		},
		Options: map[string]any{
			"deployment":  "gpt-4-deployment",
			"auth_type":   "api_key",
			"token":       "test-api-key",
			"api_version": "2024-02-01",
		},
	}

	provider, err := providers.NewAzure(cfg)
	if err != nil {
		t.Fatalf("NewAzure failed: %v", err)
	}

	req := httptest.NewRequest("POST", "http://test.com", nil)
	provider.SetHeaders(req)

	apiKeyHeader := req.Header.Get("api-key")
	expected := "test-api-key"

	if apiKeyHeader != expected {
		t.Errorf("got api-key header %q, want %q", apiKeyHeader, expected)
	}
}

func TestAzure_SetHeaders_Bearer(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Model: &config.ModelConfig{
			Name: "gpt-4",
			Capabilities: map[string]config.CapabilityConfig{
				"chat": {Format: "openai-chat"},
			},
		},
		Options: map[string]any{
			"deployment":  "gpt-4-deployment",
			"auth_type":   "bearer",
			"token":       "test-bearer-token",
			"api_version": "2024-02-01",
		},
	}

	provider, err := providers.NewAzure(cfg)
	if err != nil {
		t.Fatalf("NewAzure failed: %v", err)
	}

	req := httptest.NewRequest("POST", "http://test.com", nil)
	provider.SetHeaders(req)

	authHeader := req.Header.Get("Authorization")
	expected := "Bearer test-bearer-token"

	if authHeader != expected {
		t.Errorf("got Authorization header %q, want %q", authHeader, expected)
	}
}

func TestAzure_ProcessResponse(t *testing.T) {
	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]any{
			"model": "gpt-4",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": "Hello from Azure!",
					},
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: server.URL,
		Model: &config.ModelConfig{
			Name: "gpt-4",
			Capabilities: map[string]config.CapabilityConfig{
				"chat": {Format: "openai-chat"},
			},
		},
		Options: map[string]any{
			"deployment":  "gpt-4-deployment",
			"auth_type":   "api_key",
			"token":       "test-key",
			"api_version": "2024-02-01",
		},
	}

	provider, err := providers.NewAzure(cfg)
	if err != nil {
		t.Fatalf("NewAzure failed: %v", err)
	}

	capability := capabilities.NewChatCapability("openai-chat", nil)

	result, err := provider.ProcessResponse(resp, capability)

	if err != nil {
		t.Fatalf("ProcessResponse failed: %v", err)
	}

	if result == nil {
		t.Fatal("ProcessResponse returned nil result")
	}
}

func TestAzure_ProcessResponse_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: server.URL,
		Model: &config.ModelConfig{
			Name: "gpt-4",
			Capabilities: map[string]config.CapabilityConfig{
				"chat": {Format: "openai-chat"},
			},
		},
		Options: map[string]any{
			"deployment":  "gpt-4-deployment",
			"auth_type":   "api_key",
			"token":       "test-key",
			"api_version": "2024-02-01",
		},
	}

	provider, err := providers.NewAzure(cfg)
	if err != nil {
		t.Fatalf("NewAzure failed: %v", err)
	}

	capability := capabilities.NewChatCapability("openai-chat", nil)

	_, err = provider.ProcessResponse(resp, capability)

	if err == nil {
		t.Error("expected error for non-OK status, got nil")
	}
}

func TestAzure_ProcessStreamResponse(t *testing.T) {
	// Create mock streaming server with "data: " prefix
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/event-stream")

		// Azure uses "data: " prefix for SSE
		chunks := []string{
			`data: {"choices":[{"delta":{"content":"Hello"}}]}`,
			`data: {"choices":[{"delta":{"content":" from Azure"}}]}`,
			`data: [DONE]`,
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
		Name:    "azure",
		BaseURL: server.URL,
		Model: &config.ModelConfig{
			Name: "gpt-4",
			Capabilities: map[string]config.CapabilityConfig{
				"chat": {Format: "openai-chat"},
			},
		},
		Options: map[string]any{
			"deployment":  "gpt-4-deployment",
			"auth_type":   "api_key",
			"token":       "test-key",
			"api_version": "2024-02-01",
		},
	}

	provider, err := providers.NewAzure(cfg)
	if err != nil {
		t.Fatalf("NewAzure failed: %v", err)
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
