package providers_test

import (
	"testing"

	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/providers"
)

func TestCreate_Ollama(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: "http://localhost:11434",
		Model: &config.ModelConfig{
			Name: "llama2",
			Capabilities: map[string]map[string]any{
				"chat": {},
			},
		},
	}

	provider, err := providers.Create(cfg)

	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if provider == nil {
		t.Fatal("Create returned nil provider")
	}

	if provider.Name() != "ollama" {
		t.Errorf("got name %q, want %q", provider.Name(), "ollama")
	}
}

func TestCreate_Azure(t *testing.T) {
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

	provider, err := providers.Create(cfg)

	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if provider == nil {
		t.Fatal("Create returned nil provider")
	}

	if provider.Name() != "azure" {
		t.Errorf("got name %q, want %q", provider.Name(), "azure")
	}
}

func TestCreate_UnknownProvider(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "unknown-provider",
		BaseURL: "http://localhost",
		Model: &config.ModelConfig{
			Name: "test",
			Capabilities: map[string]map[string]any{
				"chat": {},
			},
		},
	}

	_, err := providers.Create(cfg)

	if err == nil {
		t.Error("expected error for unknown provider, got nil")
	}
}
