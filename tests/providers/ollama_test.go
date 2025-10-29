package providers_test

import (
	"context"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/providers"
	"github.com/JaimeStill/go-agents/pkg/types"
)

func TestNewOllama(t *testing.T) {
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
			name:        "URL without /v1 suffix",
			baseURL:     "http://localhost:11434",
			expectedURL: "http://localhost:11434/v1/chat/completions",
		},
		{
			name:        "URL with /v1 suffix",
			baseURL:     "http://localhost:11434/v1",
			expectedURL: "http://localhost:11434/v1/chat/completions",
		},
		{
			name:        "URL with trailing slash",
			baseURL:     "http://localhost:11434/",
			expectedURL: "http://localhost:11434/v1/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ProviderConfig{
				Name:    "ollama",
				BaseURL: tt.baseURL,
				Model: &config.ModelConfig{
					Name: "llama2",
					Capabilities: map[string]map[string]any{
						"chat": {},
					},
				},
			}

			provider, err := providers.NewOllama(cfg)
			if err != nil {
				t.Fatalf("NewOllama failed: %v", err)
			}

			// Test endpoint construction instead of BaseURL
			endpoint, err := provider.GetEndpoint(types.Chat)
			if err != nil {
				t.Fatalf("GetEndpoint failed: %v", err)
			}

			if endpoint != tt.expectedURL {
				t.Errorf("got endpoint %q, want %q", endpoint, tt.expectedURL)
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
			Capabilities: map[string]map[string]any{
				"chat":       {},
				"vision":     {},
				"tools":      {},
				"embeddings": {},
			},
		},
	}

	provider, err := providers.NewOllama(cfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	tests := []struct {
		protocol types.Protocol
		expected string
	}{
		{
			types.Chat,
			"http://localhost:11434/v1/chat/completions",
		},
		{
			types.Vision,
			"http://localhost:11434/v1/chat/completions",
		},
		{
			types.Tools,
			"http://localhost:11434/v1/chat/completions",
		},
		{
			types.Embeddings,
			"http://localhost:11434/v1/embeddings",
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

func TestOllama_PrepareRequest(t *testing.T) {
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

	provider, err := providers.NewOllama(cfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	chatRequest := &types.ChatRequest{
		Messages: []types.Message{
			types.NewMessage("user", "Hello"),
		},
		Options: map[string]any{"model": "llama2"},
	}

	request, err := provider.PrepareRequest(context.Background(), chatRequest)

	if err != nil {
		t.Fatalf("PrepareRequest failed: %v", err)
	}

	if request == nil {
		t.Fatal("PrepareRequest returned nil request")
	}

	expectedURL := "http://localhost:11434/v1/chat/completions"
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

func TestOllama_PrepareStreamRequest(t *testing.T) {
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

	provider, err := providers.NewOllama(cfg)
	if err != nil {
		t.Fatalf("NewOllama failed: %v", err)
	}

	chatRequest := &types.ChatRequest{
		Messages: []types.Message{
			types.NewMessage("user", "Hello"),
		},
		Options: map[string]any{"model": "llama2", "stream": true},
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
