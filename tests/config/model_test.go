package config_test

import (
	"encoding/json"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/config"
)

func TestModelConfig_Unmarshal(t *testing.T) {
	jsonData := `{
		"name": "gpt-4",
		"capabilities": {
			"chat": {
				"format": "openai-chat",
				"options": {
					"temperature": 0.7,
					"max_tokens": 4096
				}
			},
			"vision": {
				"format": "openai-vision",
				"options": {
					"detail": "auto"
				}
			}
		}
	}`

	var cfg config.ModelConfig
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.Name != "gpt-4" {
		t.Errorf("got name %s, want gpt-4", cfg.Name)
	}

	if len(cfg.Capabilities) != 2 {
		t.Errorf("got %d capabilities, want 2", len(cfg.Capabilities))
	}

	chatCap, exists := cfg.Capabilities["chat"]
	if !exists {
		t.Fatal("chat capability not found")
	}
	if chatCap.Format != "openai-chat" {
		t.Errorf("got format %s, want openai-chat", chatCap.Format)
	}

	temp, exists := chatCap.Options["temperature"]
	if !exists {
		t.Fatal("temperature option not found")
	}
	if temp != 0.7 {
		t.Errorf("got temperature %v, want 0.7", temp)
	}
}

func TestModelConfig_Capabilities(t *testing.T) {
	cfg := &config.ModelConfig{
		Name: "test-model",
		Capabilities: config.ModelCapabilities{
			"chat": config.CapabilityConfig{
				Format: "openai-chat",
				Options: map[string]any{
					"temperature": 0.7,
				},
			},
		},
	}

	if len(cfg.Capabilities) != 1 {
		t.Errorf("got %d capabilities, want 1", len(cfg.Capabilities))
	}

	chatCap, exists := cfg.Capabilities["chat"]
	if !exists {
		t.Fatal("chat capability not found")
	}

	if chatCap.Format != "openai-chat" {
		t.Errorf("got format %s, want openai-chat", chatCap.Format)
	}
}

func TestDefaultModelConfig(t *testing.T) {
	cfg := config.DefaultModelConfig()

	if cfg == nil {
		t.Fatal("DefaultModelConfig returned nil")
	}

	if cfg.Capabilities == nil {
		t.Fatal("Capabilities map is nil")
	}

	if len(cfg.Capabilities) != 0 {
		t.Errorf("expected empty capabilities, got %d", len(cfg.Capabilities))
	}
}

func TestModelConfig_Merge(t *testing.T) {
	tests := []struct {
		name     string
		base     *config.ModelConfig
		source   *config.ModelConfig
		expected *config.ModelConfig
	}{
		{
			name: "merge name",
			base: &config.ModelConfig{
				Name: "base-model",
			},
			source: &config.ModelConfig{
				Name: "source-model",
			},
			expected: &config.ModelConfig{
				Name: "source-model",
			},
		},
		{
			name: "merge capabilities",
			base: &config.ModelConfig{
				Name: "test-model",
				Capabilities: config.ModelCapabilities{
					"chat": config.CapabilityConfig{
						Format: "openai-chat",
					},
				},
			},
			source: &config.ModelConfig{
				Capabilities: config.ModelCapabilities{
					"vision": config.CapabilityConfig{
						Format: "openai-vision",
					},
				},
			},
			expected: &config.ModelConfig{
				Name: "test-model",
				Capabilities: config.ModelCapabilities{
					"chat": config.CapabilityConfig{
						Format: "openai-chat",
					},
					"vision": config.CapabilityConfig{
						Format: "openai-vision",
					},
				},
			},
		},
		{
			name: "source empty name preserves base",
			base: &config.ModelConfig{
				Name: "base-model",
			},
			source: &config.ModelConfig{
				Name: "",
			},
			expected: &config.ModelConfig{
				Name: "base-model",
			},
		},
		{
			name: "nil capabilities initialized",
			base: &config.ModelConfig{
				Name: "test-model",
			},
			source: &config.ModelConfig{
				Capabilities: config.ModelCapabilities{
					"chat": config.CapabilityConfig{
						Format: "openai-chat",
					},
				},
			},
			expected: &config.ModelConfig{
				Name: "test-model",
				Capabilities: config.ModelCapabilities{
					"chat": config.CapabilityConfig{
						Format: "openai-chat",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.base.Merge(tt.source)

			if tt.base.Name != tt.expected.Name {
				t.Errorf("got name %s, want %s", tt.base.Name, tt.expected.Name)
			}

			if len(tt.base.Capabilities) != len(tt.expected.Capabilities) {
				t.Errorf("got %d capabilities, want %d", len(tt.base.Capabilities), len(tt.expected.Capabilities))
			}

			for key, expectedCap := range tt.expected.Capabilities {
				baseCap, exists := tt.base.Capabilities[key]
				if !exists {
					t.Errorf("capability %s missing from result", key)
					continue
				}
				if baseCap.Format != expectedCap.Format {
					t.Errorf("capability %s: got format %s, want %s", key, baseCap.Format, expectedCap.Format)
				}
			}
		})
	}
}
