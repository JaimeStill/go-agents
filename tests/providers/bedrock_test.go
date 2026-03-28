package providers_test

import (
	"context"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/providers"
	"github.com/JaimeStill/go-agents/pkg/streaming"
)

func TestNewBedrock(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name: "bedrock",
		Options: map[string]any{
			"region":          "us-east-1",
			"auth_type":       "static",
			"access_key_id":   "AKIAIOSFODNN7EXAMPLE",
			"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	}

	provider, err := providers.NewBedrock(cfg)

	if err != nil {
		t.Fatalf("NewBedrock failed: %v", err)
	}

	if provider == nil {
		t.Fatal("NewBedrock returned nil provider")
	}

	if provider.Name() != "bedrock" {
		t.Errorf("got name %q, want %q", provider.Name(), "bedrock")
	}
}

func TestNewBedrock_AutoBaseURL(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name: "bedrock",
		Options: map[string]any{
			"region":          "us-west-2",
			"auth_type":       "static",
			"access_key_id":   "AKIAIOSFODNN7EXAMPLE",
			"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	}

	provider, err := providers.NewBedrock(cfg)
	if err != nil {
		t.Fatalf("NewBedrock failed: %v", err)
	}

	expectedURL := "https://bedrock-runtime.us-west-2.amazonaws.com"
	if provider.BaseURL() != expectedURL {
		t.Errorf("got base URL %q, want %q", provider.BaseURL(), expectedURL)
	}
}

func TestNewBedrock_CustomBaseURL(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "bedrock",
		BaseURL: "https://custom-endpoint.example.com",
		Options: map[string]any{
			"region":          "us-east-1",
			"auth_type":       "static",
			"access_key_id":   "AKIAIOSFODNN7EXAMPLE",
			"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	}

	provider, err := providers.NewBedrock(cfg)
	if err != nil {
		t.Fatalf("NewBedrock failed: %v", err)
	}

	if provider.BaseURL() != "https://custom-endpoint.example.com" {
		t.Errorf("got base URL %q, want %q", provider.BaseURL(), "https://custom-endpoint.example.com")
	}
}

func TestNewBedrock_MissingRegion(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "bedrock",
		Options: map[string]any{},
	}

	_, err := providers.NewBedrock(cfg)

	if err == nil {
		t.Error("expected error for missing region, got nil")
	}
}

func TestBedrock_Endpoint(t *testing.T) {
	tests := []struct {
		name        string
		protocol    protocol.Protocol
		expectError bool
	}{
		{name: "chat", protocol: protocol.Chat, expectError: false},
		{name: "vision", protocol: protocol.Vision, expectError: false},
		{name: "tools", protocol: protocol.Tools, expectError: false},
		{name: "embeddings", protocol: protocol.Embeddings, expectError: true},
		{name: "unsupported", protocol: protocol.Protocol("unsupported"), expectError: true},
	}

	cfg := &config.ProviderConfig{
		Name: "bedrock",
		Options: map[string]any{
			"region":          "us-east-1",
			"auth_type":       "static",
			"access_key_id":   "AKIAIOSFODNN7EXAMPLE",
			"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	}

	provider, err := providers.NewBedrock(cfg)
	if err != nil {
		t.Fatalf("NewBedrock failed: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := provider.Endpoint(tt.protocol)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestBedrock_PrepareRequest(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name: "bedrock",
		Options: map[string]any{
			"region":          "us-east-1",
			"auth_type":       "static",
			"access_key_id":   "AKIAIOSFODNN7EXAMPLE",
			"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	}

	provider, err := providers.NewBedrock(cfg)
	if err != nil {
		t.Fatalf("NewBedrock failed: %v", err)
	}

	body := []byte(`{"modelId":"anthropic.claude-sonnet-4-6-v1","messages":[]}`)
	headers := map[string]string{"Content-Type": "application/json"}

	request, err := provider.PrepareRequest(context.Background(), protocol.Chat, body, headers)
	if err != nil {
		t.Fatalf("PrepareRequest failed: %v", err)
	}

	expectedURL := "https://bedrock-runtime.us-east-1.amazonaws.com/model/anthropic.claude-sonnet-4-6-v1/converse"
	if request.URL != expectedURL {
		t.Errorf("got URL %q, want %q", request.URL, expectedURL)
	}
}

func TestBedrock_PrepareRequest_ModelWithColon(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name: "bedrock",
		Options: map[string]any{
			"region":          "us-east-1",
			"auth_type":       "static",
			"access_key_id":   "AKIAIOSFODNN7EXAMPLE",
			"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	}

	provider, err := providers.NewBedrock(cfg)
	if err != nil {
		t.Fatalf("NewBedrock failed: %v", err)
	}

	body := []byte(`{"modelId":"anthropic.claude-haiku-4-5-20251001-v1:0","messages":[]}`)
	headers := map[string]string{"Content-Type": "application/json"}

	request, err := provider.PrepareRequest(context.Background(), protocol.Chat, body, headers)
	if err != nil {
		t.Fatalf("PrepareRequest failed: %v", err)
	}

	expectedURL := "https://bedrock-runtime.us-east-1.amazonaws.com/model/anthropic.claude-haiku-4-5-20251001-v1:0/converse"
	if request.URL != expectedURL {
		t.Errorf("got URL %q, want %q", request.URL, expectedURL)
	}
}

func TestBedrock_PrepareStreamRequest(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name: "bedrock",
		Options: map[string]any{
			"region":          "us-east-1",
			"auth_type":       "static",
			"access_key_id":   "AKIAIOSFODNN7EXAMPLE",
			"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	}

	provider, err := providers.NewBedrock(cfg)
	if err != nil {
		t.Fatalf("NewBedrock failed: %v", err)
	}

	body := []byte(`{"modelId":"anthropic.claude-sonnet-4-6-v1","messages":[]}`)
	headers := map[string]string{"Content-Type": "application/json"}

	request, err := provider.PrepareStreamRequest(context.Background(), protocol.Chat, body, headers)
	if err != nil {
		t.Fatalf("PrepareStreamRequest failed: %v", err)
	}

	expectedURL := "https://bedrock-runtime.us-east-1.amazonaws.com/model/anthropic.claude-sonnet-4-6-v1/converse-stream"
	if request.URL != expectedURL {
		t.Errorf("got URL %q, want %q", request.URL, expectedURL)
	}

	if request.Headers["Accept"] != streaming.EventStreamMedia {
		t.Errorf("got Accept header %q, want %q", request.Headers["Accept"], streaming.EventStreamMedia)
	}
}

func TestBedrock_Stream(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name: "bedrock",
		Options: map[string]any{
			"region":          "us-east-1",
			"auth_type":       "static",
			"access_key_id":   "AKIAIOSFODNN7EXAMPLE",
			"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	}

	provider, err := providers.NewBedrock(cfg)
	if err != nil {
		t.Fatalf("NewBedrock failed: %v", err)
	}

	stream := provider.Stream()
	if stream == nil {
		t.Fatal("Stream() returned nil")
	}

	if _, ok := stream.(*streaming.EventStreamReader); !ok {
		t.Errorf("expected *streaming.EventStreamReader, got %T", stream)
	}
}
