package config_test

import (
	"encoding/json"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/config"
)

func TestProviderConfig_Unmarshal(t *testing.T) {
	jsonData := `{
		"name": "azure",
		"base_url": "https://example.openai.azure.com",
		"model": {
			"name": "gpt-4",
			"capabilities": {
				"chat": {
					"format": "openai-chat"
				}
			}
		},
		"options": {
			"deployment": "gpt-4-deployment",
			"api_version": "2024-08-01"
		}
	}`

	var cfg config.ProviderConfig
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.Name != "azure" {
		t.Errorf("got name %s, want azure", cfg.Name)
	}

	if cfg.BaseURL != "https://example.openai.azure.com" {
		t.Errorf("got base_url %s, want https://example.openai.azure.com", cfg.BaseURL)
	}

	if cfg.Model == nil {
		t.Fatal("model is nil")
	}

	if cfg.Model.Name != "gpt-4" {
		t.Errorf("got model name %s, want gpt-4", cfg.Model.Name)
	}

	if len(cfg.Options) != 2 {
		t.Errorf("got %d options, want 2", len(cfg.Options))
	}
}

func TestProviderConfig_Options(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "test-provider",
		BaseURL: "http://localhost",
		Options: map[string]any{
			"deployment":  "test-deployment",
			"api_version": "2024-08-01",
			"auth_type":   "api_key",
		},
	}

	if len(cfg.Options) != 3 {
		t.Errorf("got %d options, want 3", len(cfg.Options))
	}

	deployment, exists := cfg.Options["deployment"]
	if !exists {
		t.Fatal("deployment option not found")
	}
	if deployment != "test-deployment" {
		t.Errorf("got deployment %v, want test-deployment", deployment)
	}
}

func TestDefaultProviderConfig(t *testing.T) {
	cfg := config.DefaultProviderConfig()

	if cfg == nil {
		t.Fatal("DefaultProviderConfig returned nil")
	}

	if cfg.Name != "ollama" {
		t.Errorf("got name %s, want ollama", cfg.Name)
	}

	if cfg.BaseURL != "http://localhost:11434" {
		t.Errorf("got base_url %s, want http://localhost:11434", cfg.BaseURL)
	}

	if cfg.Model == nil {
		t.Fatal("model is nil")
	}

	if cfg.Options == nil {
		t.Fatal("options is nil")
	}
}

func TestProviderConfig_Merge(t *testing.T) {
	tests := []struct {
		name     string
		base     *config.ProviderConfig
		source   *config.ProviderConfig
		expected *config.ProviderConfig
	}{
		{
			name: "merge name",
			base: &config.ProviderConfig{
				Name: "base-provider",
			},
			source: &config.ProviderConfig{
				Name: "source-provider",
			},
			expected: &config.ProviderConfig{
				Name: "source-provider",
			},
		},
		{
			name: "merge base_url",
			base: &config.ProviderConfig{
				BaseURL: "http://base",
			},
			source: &config.ProviderConfig{
				BaseURL: "http://source",
			},
			expected: &config.ProviderConfig{
				BaseURL: "http://source",
			},
		},
		{
			name: "merge options",
			base: &config.ProviderConfig{
				Options: map[string]any{
					"option1": "value1",
				},
			},
			source: &config.ProviderConfig{
				Options: map[string]any{
					"option2": "value2",
				},
			},
			expected: &config.ProviderConfig{
				Options: map[string]any{
					"option1": "value1",
					"option2": "value2",
				},
			},
		},
		{
			name: "merge model",
			base: &config.ProviderConfig{
				Model: &config.ModelConfig{
					Name: "base-model",
				},
			},
			source: &config.ProviderConfig{
				Model: &config.ModelConfig{
					Name: "source-model",
				},
			},
			expected: &config.ProviderConfig{
				Model: &config.ModelConfig{
					Name: "source-model",
				},
			},
		},
		{
			name: "source empty name preserves base",
			base: &config.ProviderConfig{
				Name: "base-provider",
			},
			source: &config.ProviderConfig{
				Name: "",
			},
			expected: &config.ProviderConfig{
				Name: "base-provider",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.base.Merge(tt.source)

			if tt.base.Name != tt.expected.Name {
				t.Errorf("got name %s, want %s", tt.base.Name, tt.expected.Name)
			}

			if tt.base.BaseURL != tt.expected.BaseURL {
				t.Errorf("got base_url %s, want %s", tt.base.BaseURL, tt.expected.BaseURL)
			}

			if tt.expected.Model != nil {
				if tt.base.Model == nil {
					t.Fatal("model is nil after merge")
				}
				if tt.base.Model.Name != tt.expected.Model.Name {
					t.Errorf("got model name %s, want %s", tt.base.Model.Name, tt.expected.Model.Name)
				}
			}

			for key, expectedValue := range tt.expected.Options {
				baseValue, exists := tt.base.Options[key]
				if !exists {
					t.Errorf("option %s missing from result", key)
					continue
				}
				if baseValue != expectedValue {
					t.Errorf("option %s: got %v, want %v", key, baseValue, expectedValue)
				}
			}
		})
	}
}
