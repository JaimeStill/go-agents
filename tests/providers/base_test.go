package providers_test

import (
	"testing"

	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/models"
	"github.com/JaimeStill/go-agents/pkg/providers"
)

func TestNewBaseProvider(t *testing.T) {
	modelCfg := &config.ModelConfig{
		Name: "test-model",
		Capabilities: map[string]config.CapabilityConfig{
			"chat": {
				Format:  "openai-chat",
				Options: map[string]any{},
			},
		},
	}

	model, err := models.New(modelCfg)
	if err != nil {
		t.Fatalf("failed to create model: %v", err)
	}

	provider := providers.NewBaseProvider("test-provider", "https://api.example.com", model)

	if provider == nil {
		t.Fatal("NewBaseProvider returned nil")
	}

	if provider.Name() != "test-provider" {
		t.Errorf("got name %q, want %q", provider.Name(), "test-provider")
	}

	if provider.BaseURL() != "https://api.example.com" {
		t.Errorf("got baseURL %q, want %q", provider.BaseURL(), "https://api.example.com")
	}

	if provider.Model() == nil {
		t.Error("Model() returned nil")
	}

	if provider.Model().Name() != "test-model" {
		t.Errorf("got model name %q, want %q", provider.Model().Name(), "test-model")
	}
}

func TestBaseProvider_Name(t *testing.T) {
	modelCfg := &config.ModelConfig{
		Name: "test-model",
		Capabilities: map[string]config.CapabilityConfig{
			"chat": {Format: "openai-chat"},
		},
	}

	model, _ := models.New(modelCfg)
	provider := providers.NewBaseProvider("my-provider", "https://api.test.com", model)

	if provider.Name() != "my-provider" {
		t.Errorf("got name %q, want %q", provider.Name(), "my-provider")
	}
}

func TestBaseProvider_BaseURL(t *testing.T) {
	modelCfg := &config.ModelConfig{
		Name: "test-model",
		Capabilities: map[string]config.CapabilityConfig{
			"chat": {Format: "openai-chat"},
		},
	}

	model, _ := models.New(modelCfg)
	provider := providers.NewBaseProvider("test", "https://custom.api.com/v2", model)

	if provider.BaseURL() != "https://custom.api.com/v2" {
		t.Errorf("got baseURL %q, want %q", provider.BaseURL(), "https://custom.api.com/v2")
	}
}

func TestBaseProvider_Model(t *testing.T) {
	modelCfg := &config.ModelConfig{
		Name: "gpt-4",
		Capabilities: map[string]config.CapabilityConfig{
			"chat": {
				Format: "openai-chat",
				Options: map[string]any{
					"temperature": 0.7,
				},
			},
		},
	}

	model, _ := models.New(modelCfg)
	provider := providers.NewBaseProvider("test", "https://api.test.com", model)

	result := provider.Model()

	if result == nil {
		t.Fatal("Model() returned nil")
	}

	if result.Name() != "gpt-4" {
		t.Errorf("got model name %q, want %q", result.Name(), "gpt-4")
	}
}
