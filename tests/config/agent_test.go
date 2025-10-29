package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JaimeStill/go-agents/pkg/config"
)

func TestAgentConfig_Unmarshal(t *testing.T) {
	jsonData := `{
		"name": "test-agent",
		"system_prompt": "You are a helpful assistant",
		"client": {
			"provider": {
				"name": "ollama",
				"base_url": "http://localhost:11434",
				"model": {
					"name": "llama3.2:3b",
					"capabilities": {
						"chat": {
							"temperature": 0.7
						}
					}
				}
			},
			"timeout": "24s",
			"retry": {
				"max_retries": 3
			}
		}
	}`

	var cfg config.AgentConfig
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.Name != "test-agent" {
		t.Errorf("got name %s, want test-agent", cfg.Name)
	}

	if cfg.SystemPrompt != "You are a helpful assistant" {
		t.Errorf("got system_prompt %s, want 'You are a helpful assistant'", cfg.SystemPrompt)
	}

	if cfg.Client == nil {
		t.Fatal("client is nil")
	}

	if cfg.Client.Provider == nil {
		t.Fatal("provider is nil")
	}

	if cfg.Client.Provider.Name != "ollama" {
		t.Errorf("got provider name %s, want ollama", cfg.Client.Provider.Name)
	}
}

func TestAgentConfig_FullConfiguration(t *testing.T) {
	cfg := &config.AgentConfig{
		Name:         "full-agent",
		SystemPrompt: "Test system prompt",
		Client: &config.ClientConfig{
			Provider: &config.ProviderConfig{
				Name:    "azure",
				BaseURL: "https://example.openai.azure.com",
				Model: &config.ModelConfig{
					Name: "gpt-4",
					Capabilities: map[string]map[string]any{
						"chat": {
							"temperature": 0.7,
							"max_tokens":  4096,
						},
						"vision": {
							"detail": "auto",
						},
					},
				},
				Options: map[string]any{
					"deployment":  "gpt-4-deployment",
					"api_version": "2024-08-01",
					"auth_type":   "api_key",
				},
			},
			Timeout: config.Duration(24 * time.Second),
			Retry: config.RetryConfig{
				MaxRetries:     3,
				InitialBackoff: config.Duration(1 * time.Second),
			},
			ConnectionPoolSize: 10,
			ConnectionTimeout:  config.Duration(9 * time.Second),
		},
	}

	if cfg.Name != "full-agent" {
		t.Errorf("got name %s, want full-agent", cfg.Name)
	}

	if cfg.SystemPrompt != "Test system prompt" {
		t.Errorf("got system_prompt %s, want 'Test system prompt'", cfg.SystemPrompt)
	}

	if cfg.Client.Provider.Model.Name != "gpt-4" {
		t.Errorf("got model name %s, want gpt-4", cfg.Client.Provider.Model.Name)
	}

	if len(cfg.Client.Provider.Model.Capabilities) != 2 {
		t.Errorf("got %d capabilities, want 2", len(cfg.Client.Provider.Model.Capabilities))
	}
}

func TestDefaultAgentConfig(t *testing.T) {
	cfg := config.DefaultAgentConfig()

	if cfg.Name != "default-agent" {
		t.Errorf("got name %s, want default-agent", cfg.Name)
	}

	if cfg.SystemPrompt != "" {
		t.Errorf("got system_prompt %s, want empty string", cfg.SystemPrompt)
	}

	if cfg.Client == nil {
		t.Fatal("client is nil")
	}

	if cfg.Client.Provider == nil {
		t.Fatal("provider is nil")
	}

	if cfg.Client.Provider.Name != "ollama" {
		t.Errorf("got provider name %s, want ollama", cfg.Client.Provider.Name)
	}
}

func TestAgentConfig_Merge(t *testing.T) {
	tests := []struct {
		name     string
		base     *config.AgentConfig
		source   *config.AgentConfig
		expected *config.AgentConfig
	}{
		{
			name: "merge name",
			base: &config.AgentConfig{
				Name: "base-agent",
			},
			source: &config.AgentConfig{
				Name: "source-agent",
			},
			expected: &config.AgentConfig{
				Name: "source-agent",
			},
		},
		{
			name: "merge system_prompt",
			base: &config.AgentConfig{
				SystemPrompt: "base prompt",
			},
			source: &config.AgentConfig{
				SystemPrompt: "source prompt",
			},
			expected: &config.AgentConfig{
				SystemPrompt: "source prompt",
			},
		},
		{
			name: "merge client",
			base: &config.AgentConfig{
				Client: &config.ClientConfig{
					Retry: config.RetryConfig{
						MaxRetries: 3,
					},
				},
			},
			source: &config.AgentConfig{
				Client: &config.ClientConfig{
					Retry: config.RetryConfig{
						MaxRetries: 5,
					},
				},
			},
			expected: &config.AgentConfig{
				Client: &config.ClientConfig{
					Retry: config.RetryConfig{
						MaxRetries: 5,
					},
				},
			},
		},
		{
			name: "source empty name preserves base",
			base: &config.AgentConfig{
				Name: "base-agent",
			},
			source: &config.AgentConfig{
				Name: "",
			},
			expected: &config.AgentConfig{
				Name: "base-agent",
			},
		},
		{
			name: "source empty system_prompt preserves base",
			base: &config.AgentConfig{
				SystemPrompt: "base prompt",
			},
			source: &config.AgentConfig{
				SystemPrompt: "",
			},
			expected: &config.AgentConfig{
				SystemPrompt: "base prompt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.base.Merge(tt.source)

			if tt.base.Name != tt.expected.Name {
				t.Errorf("got name %s, want %s", tt.base.Name, tt.expected.Name)
			}

			if tt.base.SystemPrompt != tt.expected.SystemPrompt {
				t.Errorf("got system_prompt %s, want %s", tt.base.SystemPrompt, tt.expected.SystemPrompt)
			}

			if tt.expected.Client != nil {
				if tt.base.Client == nil {
					t.Fatal("client is nil after merge")
				}
				if tt.base.Client.Retry.MaxRetries != tt.expected.Client.Retry.MaxRetries {
					t.Errorf("got max_retries %d, want %d", tt.base.Client.Retry.MaxRetries, tt.expected.Client.Retry.MaxRetries)
				}
			}
		})
	}
}

func TestLoadAgentConfig(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		configJSON  string
		expectError bool
		validate    func(*testing.T, *config.AgentConfig)
	}{
		{
			name: "valid config",
			configJSON: `{
				"name": "test-agent",
				"system_prompt": "Test prompt",
				"client": {
					"provider": {
						"name": "ollama",
						"base_url": "http://localhost:11434",
						"model": {
							"name": "llama3.2:3b"
						}
					},
					"timeout": "24s"
				}
			}`,
			expectError: false,
			validate: func(t *testing.T, cfg *config.AgentConfig) {
				if cfg.Name != "test-agent" {
					t.Errorf("got name %s, want test-agent", cfg.Name)
				}
				if cfg.SystemPrompt != "Test prompt" {
					t.Errorf("got system_prompt %s, want 'Test prompt'", cfg.SystemPrompt)
				}
			},
		},
		{
			name:        "invalid json",
			configJSON:  `{invalid json}`,
			expectError: true,
		},
		{
			name:        "file not found",
			configJSON:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filename string

			if tt.name == "file not found" {
				filename = filepath.Join(tempDir, "nonexistent.json")
			} else {
				filename = filepath.Join(tempDir, tt.name+".json")
				if err := os.WriteFile(filename, []byte(tt.configJSON), 0644); err != nil {
					t.Fatalf("failed to write test config file: %v", err)
				}
			}

			cfg, err := config.LoadAgentConfig(filename)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cfg == nil {
				t.Fatal("config is nil")
			}

			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestLoadAgentConfig_MergesWithDefaults(t *testing.T) {
	tempDir := t.TempDir()

	configJSON := `{
		"name": "custom-agent"
	}`

	filename := filepath.Join(tempDir, "config.json")
	if err := os.WriteFile(filename, []byte(configJSON), 0644); err != nil {
		t.Fatalf("failed to write test config file: %v", err)
	}

	cfg, err := config.LoadAgentConfig(filename)
	if err != nil {
		t.Fatalf("LoadAgentConfig failed: %v", err)
	}

	// Should have custom name
	if cfg.Name != "custom-agent" {
		t.Errorf("got name %s, want custom-agent", cfg.Name)
	}

	// Should have default client
	if cfg.Client == nil {
		t.Fatal("client is nil")
	}

	// Should have default provider
	if cfg.Client.Provider == nil {
		t.Fatal("provider is nil")
	}

	if cfg.Client.Provider.Name != "ollama" {
		t.Errorf("got provider name %s, want ollama (from defaults)", cfg.Client.Provider.Name)
	}
}
