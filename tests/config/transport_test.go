package config_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/JaimeStill/go-agents/pkg/config"
)

func TestTransportConfig_Unmarshal(t *testing.T) {
	jsonData := `{
		"provider": {
			"name": "ollama",
			"base_url": "http://localhost:11434",
			"model": {
				"name": "llama3.2:3b"
			}
		},
		"timeout": "24s",
		"max_retries": 3,
		"retry_backoff_base": "1s",
		"connection_pool_size": 10,
		"connection_timeout": "9s"
	}`

	var cfg config.TransportConfig
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.Provider == nil {
		t.Fatal("provider is nil")
	}

	if cfg.Provider.Name != "ollama" {
		t.Errorf("got provider name %s, want ollama", cfg.Provider.Name)
	}

	if cfg.Timeout.ToDuration() != 24*time.Second {
		t.Errorf("got timeout %v, want 24s", cfg.Timeout.ToDuration())
	}

	if cfg.MaxRetries != 3 {
		t.Errorf("got max_retries %d, want 3", cfg.MaxRetries)
	}

	if cfg.RetryBackoffBase.ToDuration() != 1*time.Second {
		t.Errorf("got retry_backoff_base %v, want 1s", cfg.RetryBackoffBase.ToDuration())
	}

	if cfg.ConnectionPoolSize != 10 {
		t.Errorf("got connection_pool_size %d, want 10", cfg.ConnectionPoolSize)
	}

	if cfg.ConnectionTimeout.ToDuration() != 9*time.Second {
		t.Errorf("got connection_timeout %v, want 9s", cfg.ConnectionTimeout.ToDuration())
	}
}

func TestTransportConfig_Defaults(t *testing.T) {
	cfg := config.DefaultTransportConfig()

	if cfg == nil {
		t.Fatal("DefaultTransportConfig returned nil")
	}

	if cfg.Provider == nil {
		t.Fatal("provider is nil")
	}

	if cfg.Timeout.ToDuration() != 2*time.Minute {
		t.Errorf("got timeout %v, want 2m", cfg.Timeout.ToDuration())
	}

	if cfg.MaxRetries != 3 {
		t.Errorf("got max_retries %d, want 3", cfg.MaxRetries)
	}

	if cfg.RetryBackoffBase.ToDuration() != 1*time.Second {
		t.Errorf("got retry_backoff_base %v, want 1s", cfg.RetryBackoffBase.ToDuration())
	}

	if cfg.ConnectionPoolSize != 10 {
		t.Errorf("got connection_pool_size %d, want 10", cfg.ConnectionPoolSize)
	}

	if cfg.ConnectionTimeout.ToDuration() != 90*time.Second {
		t.Errorf("got connection_timeout %v, want 90s", cfg.ConnectionTimeout.ToDuration())
	}
}

func TestTransportConfig_ConnectionPooling(t *testing.T) {
	tests := []struct {
		name     string
		poolSize int
	}{
		{
			name:     "default pool size",
			poolSize: 10,
		},
		{
			name:     "custom pool size",
			poolSize: 20,
		},
		{
			name:     "small pool size",
			poolSize: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.TransportConfig{
				ConnectionPoolSize: tt.poolSize,
			}

			if cfg.ConnectionPoolSize != tt.poolSize {
				t.Errorf("got connection_pool_size %d, want %d", cfg.ConnectionPoolSize, tt.poolSize)
			}
		})
	}
}

func TestTransportConfig_Merge(t *testing.T) {
	tests := []struct {
		name     string
		base     *config.TransportConfig
		source   *config.TransportConfig
		expected *config.TransportConfig
	}{
		{
			name: "merge timeout",
			base: &config.TransportConfig{
				Timeout: config.Duration(1 * time.Minute),
			},
			source: &config.TransportConfig{
				Timeout: config.Duration(2 * time.Minute),
			},
			expected: &config.TransportConfig{
				Timeout: config.Duration(2 * time.Minute),
			},
		},
		{
			name: "merge max_retries",
			base: &config.TransportConfig{
				MaxRetries: 3,
			},
			source: &config.TransportConfig{
				MaxRetries: 5,
			},
			expected: &config.TransportConfig{
				MaxRetries: 5,
			},
		},
		{
			name: "merge retry_backoff_base",
			base: &config.TransportConfig{
				RetryBackoffBase: config.Duration(1 * time.Second),
			},
			source: &config.TransportConfig{
				RetryBackoffBase: config.Duration(2 * time.Second),
			},
			expected: &config.TransportConfig{
				RetryBackoffBase: config.Duration(2 * time.Second),
			},
		},
		{
			name: "merge connection_pool_size",
			base: &config.TransportConfig{
				ConnectionPoolSize: 10,
			},
			source: &config.TransportConfig{
				ConnectionPoolSize: 20,
			},
			expected: &config.TransportConfig{
				ConnectionPoolSize: 20,
			},
		},
		{
			name: "merge connection_timeout",
			base: &config.TransportConfig{
				ConnectionTimeout: config.Duration(60 * time.Second),
			},
			source: &config.TransportConfig{
				ConnectionTimeout: config.Duration(90 * time.Second),
			},
			expected: &config.TransportConfig{
				ConnectionTimeout: config.Duration(90 * time.Second),
			},
		},
		{
			name: "merge provider",
			base: &config.TransportConfig{
				Provider: &config.ProviderConfig{
					Name: "base-provider",
				},
			},
			source: &config.TransportConfig{
				Provider: &config.ProviderConfig{
					Name: "source-provider",
				},
			},
			expected: &config.TransportConfig{
				Provider: &config.ProviderConfig{
					Name: "source-provider",
				},
			},
		},
		{
			name: "zero values preserve base",
			base: &config.TransportConfig{
				Timeout:            config.Duration(1 * time.Minute),
				MaxRetries:         3,
				ConnectionPoolSize: 10,
			},
			source: &config.TransportConfig{
				Timeout:            0,
				MaxRetries:         0,
				ConnectionPoolSize: 0,
			},
			expected: &config.TransportConfig{
				Timeout:            config.Duration(1 * time.Minute),
				MaxRetries:         3,
				ConnectionPoolSize: 10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.base.Merge(tt.source)

			if tt.base.Timeout != tt.expected.Timeout {
				t.Errorf("got timeout %v, want %v", tt.base.Timeout, tt.expected.Timeout)
			}

			if tt.base.MaxRetries != tt.expected.MaxRetries {
				t.Errorf("got max_retries %d, want %d", tt.base.MaxRetries, tt.expected.MaxRetries)
			}

			if tt.base.RetryBackoffBase != tt.expected.RetryBackoffBase {
				t.Errorf("got retry_backoff_base %v, want %v", tt.base.RetryBackoffBase, tt.expected.RetryBackoffBase)
			}

			if tt.base.ConnectionPoolSize != tt.expected.ConnectionPoolSize {
				t.Errorf("got connection_pool_size %d, want %d", tt.base.ConnectionPoolSize, tt.expected.ConnectionPoolSize)
			}

			if tt.base.ConnectionTimeout != tt.expected.ConnectionTimeout {
				t.Errorf("got connection_timeout %v, want %v", tt.base.ConnectionTimeout, tt.expected.ConnectionTimeout)
			}

			if tt.expected.Provider != nil {
				if tt.base.Provider == nil {
					t.Fatal("provider is nil after merge")
				}
				if tt.base.Provider.Name != tt.expected.Provider.Name {
					t.Errorf("got provider name %s, want %s", tt.base.Provider.Name, tt.expected.Provider.Name)
				}
			}
		})
	}
}
