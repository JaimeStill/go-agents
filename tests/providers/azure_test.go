package providers_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/format"
	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/providers"
	"github.com/JaimeStill/go-agents/pkg/streaming"
)

func TestNewAzure(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
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

func TestNewAzure_MissingToken_Bearer(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Options: map[string]any{
			"deployment":  "gpt-4-deployment",
			"auth_type":   "bearer",
			"api_version": "2024-02-01",
		},
	}

	_, err := providers.NewAzure(cfg)

	if err == nil {
		t.Error("expected error for missing token with bearer auth, got nil")
	}

	if !strings.Contains(err.Error(), `auth_type "bearer"`) {
		t.Errorf("error should mention auth_type, got: %v", err)
	}
}

func TestNewAzure_MissingAPIVersion(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
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

func TestNewAzure_UnsupportedAuthType(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Options: map[string]any{
			"deployment":  "gpt-4-deployment",
			"auth_type":   "unknown",
			"api_version": "2024-02-01",
		},
	}

	_, err := providers.NewAzure(cfg)

	if err == nil {
		t.Error("expected error for unsupported auth_type, got nil")
	}

	if !strings.Contains(err.Error(), "unsupported auth_type") {
		t.Errorf("error should mention unsupported auth_type, got: %v", err)
	}
}

func TestAzure_Endpoint(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
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
		protocol protocol.Protocol
		expected string
	}{
		{
			protocol.Chat,
			"https://my-resource.openai.azure.com/deployments/gpt-4-deployment/chat/completions?api-version=2024-02-01",
		},
		{
			protocol.Vision,
			"https://my-resource.openai.azure.com/deployments/gpt-4-deployment/chat/completions?api-version=2024-02-01",
		},
		{
			protocol.Tools,
			"https://my-resource.openai.azure.com/deployments/gpt-4-deployment/chat/completions?api-version=2024-02-01",
		},
		{
			protocol.Embeddings,
			"https://my-resource.openai.azure.com/deployments/gpt-4-deployment/embeddings?api-version=2024-02-01",
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.protocol), func(t *testing.T) {
			endpoint, err := provider.Endpoint(tt.protocol)

			if err != nil {
				t.Fatalf("Endpoint failed: %v", err)
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

	f, err := format.OpenAIFactory()
	if err != nil {
		t.Fatalf("OpenAIFactory failed: %v", err)
	}

	chatData := &format.ChatData{
		Model: "gpt-4",
		Messages: []protocol.Message{
			protocol.NewMessage("user", "Hello"),
		},
		Options: map[string]any{},
	}

	body, err := f.Marshal(protocol.Chat, chatData)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	request, err := provider.PrepareRequest(context.Background(), protocol.Chat, body, headers)

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

	f, err := format.OpenAIFactory()
	if err != nil {
		t.Fatalf("OpenAIFactory failed: %v", err)
	}

	chatData := &format.ChatData{
		Model: "gpt-4",
		Messages: []protocol.Message{
			protocol.NewMessage("user", "Hello"),
		},
		Options: map[string]any{"stream": true},
	}

	body, err := f.Marshal(protocol.Chat, chatData)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	request, err := provider.PrepareStreamRequest(context.Background(), protocol.Chat, body, headers)

	if err != nil {
		t.Fatalf("PrepareStreamRequest failed: %v", err)
	}

	if request == nil {
		t.Fatal("PrepareStreamRequest returned nil request")
	}

	if request.Headers["Accept"] != streaming.SSEMedia {
		t.Errorf("got Accept header %q, want %q", request.Headers["Accept"], streaming.SSEMedia)
	}

	if request.Headers["Cache-Control"] != "no-cache" {
		t.Errorf("got Cache-Control header %q, want %q", request.Headers["Cache-Control"], "no-cache")
	}
}

func TestAzure_Stream(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
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

	stream := provider.Stream()
	if stream == nil {
		t.Fatal("Stream() returned nil")
	}

	if _, ok := stream.(*streaming.SSEReader); !ok {
		t.Errorf("expected *streaming.SSEReader, got %T", stream)
	}
}

func TestAzure_SetHeaders_ApiKey(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
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

	req, _ := http.NewRequest("POST", "https://example.com", nil)
	err = provider.SetHeaders(context.Background(), req)

	if err != nil {
		t.Fatalf("SetHeaders failed: %v", err)
	}

	if req.Header.Get("api-key") != "test-api-key" {
		t.Errorf("got api-key header %q, want %q", req.Header.Get("api-key"), "test-api-key")
	}
}

func TestAzure_SetHeaders_Bearer(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
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

	req, _ := http.NewRequest("POST", "https://example.com", nil)
	err = provider.SetHeaders(context.Background(), req)

	if err != nil {
		t.Fatalf("SetHeaders failed: %v", err)
	}

	expected := "Bearer test-bearer-token"
	if req.Header.Get("Authorization") != expected {
		t.Errorf("got Authorization header %q, want %q", req.Header.Get("Authorization"), expected)
	}
}
