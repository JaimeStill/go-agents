package providers_test

import (
	"context"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/providers"
	"github.com/JaimeStill/go-agents/pkg/types"
)

func TestNewAzure(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "azure",
		BaseURL: "https://my-resource.openai.azure.com",
		Model: &config.ModelConfig{
			Name: "gpt-4",
			Capabilities: map[string]map[string]any{
				"chat": {},
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
			Capabilities: map[string]map[string]any{
				"chat": {},
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
			Capabilities: map[string]map[string]any{
				"chat": {},
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
			Capabilities: map[string]map[string]any{
				"chat": {},
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
			Capabilities: map[string]map[string]any{
				"chat": {},
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
			Capabilities: map[string]map[string]any{
				"chat":       {},
				"vision":     {},
				"tools":      {},
				"embeddings": {},
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
		protocol types.Protocol
		expected string
	}{
		{
			types.Chat,
			"https://my-resource.openai.azure.com/deployments/gpt-4-deployment/chat/completions?api-version=2024-02-01",
		},
		{
			types.Vision,
			"https://my-resource.openai.azure.com/deployments/gpt-4-deployment/chat/completions?api-version=2024-02-01",
		},
		{
			types.Tools,
			"https://my-resource.openai.azure.com/deployments/gpt-4-deployment/chat/completions?api-version=2024-02-01",
		},
		{
			types.Embeddings,
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
			Capabilities: map[string]map[string]any{
				"chat": {},
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

	chatRequest := &types.ChatRequest{
		Messages: []types.Message{
			types.NewMessage("user", "Hello"),
		},
		Options: map[string]any{"model": "gpt-4"},
	}

	request, err := provider.PrepareRequest(context.Background(), chatRequest)

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
			Capabilities: map[string]map[string]any{
				"chat": {},
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

	chatRequest := &types.ChatRequest{
		Messages: []types.Message{
			types.NewMessage("user", "Hello"),
		},
		Options: map[string]any{"model": "gpt-4", "stream": true},
	}

	request, err := provider.PrepareStreamRequest(context.Background(), chatRequest)

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
