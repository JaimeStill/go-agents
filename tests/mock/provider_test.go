package mock_test

import (
	"context"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/mock"
	"github.com/JaimeStill/go-agents/pkg/providers"
	"github.com/JaimeStill/go-agents/pkg/types"
)

func TestNewMockProvider(t *testing.T) {
	provider := mock.NewMockProvider()

	if provider == nil {
		t.Fatal("NewMockProvider returned nil")
	}
}

func TestMockProvider_Name(t *testing.T) {
	provider := mock.NewMockProvider()

	if provider.Name() != "mock-provider" {
		t.Errorf("got name %q, want %q", provider.Name(), "mock-provider")
	}
}

func TestMockProvider_GetEndpoint(t *testing.T) {
	customEndpoints := map[types.Protocol]string{
		types.Chat:   "/chat",
		types.Vision: "/vision",
	}

	provider := mock.NewMockProvider(
		mock.WithBaseURL("https://custom.api"),
		mock.WithEndpointMapping(customEndpoints),
	)

	endpoint, err := provider.GetEndpoint(types.Chat)

	if err != nil {
		t.Fatalf("GetEndpoint failed: %v", err)
	}

	if endpoint != "https://custom.api/chat" {
		t.Errorf("got endpoint %q, want %q", endpoint, "https://custom.api/chat")
	}
}

func TestMockProvider_PrepareRequest(t *testing.T) {
	expectedRequest := &providers.Request{
		URL:     "https://test.api/chat",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    []byte(`{"test":"data"}`),
	}

	provider := mock.NewMockProvider(
		mock.WithPrepareResponse(expectedRequest, nil),
	)

	chatRequest := &types.ChatRequest{
		Messages: []types.Message{
			types.NewMessage("user", "Hello"),
		},
		Options: map[string]any{"model": "test"},
	}

	request, err := provider.PrepareRequest(context.Background(), chatRequest)

	if err != nil {
		t.Fatalf("PrepareRequest failed: %v", err)
	}

	if request != expectedRequest {
		t.Error("returned different request than configured")
	}
}

func TestMockProvider_Model(t *testing.T) {
	provider := mock.NewMockProvider()

	model := provider.Model()

	if model == nil {
		t.Error("Model() returned nil")
	}

	if model.Name != "mock-model" {
		t.Errorf("got model name %q, want %q", model.Name, "mock-model")
	}
}
