package mock_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/mock"
	"github.com/JaimeStill/go-agents/pkg/protocols"
	"github.com/JaimeStill/go-agents/pkg/providers"
)

func TestNewMockProvider_DefaultConfiguration(t *testing.T) {
	p := mock.NewMockProvider()

	if p.Name() != "mock-provider" {
		t.Errorf("expected name 'mock-provider', got %q", p.Name())
	}

	if p.Model() == nil {
		t.Error("expected default model")
	}
}

func TestMockProvider_WithProviderName(t *testing.T) {
	customName := "custom-provider"
	p := mock.NewMockProvider(mock.WithProviderName(customName))

	if p.Name() != customName {
		t.Errorf("expected name %q, got %q", customName, p.Name())
	}
}

func TestMockProvider_WithBaseURL(t *testing.T) {
	customURL := "http://custom.local"
	p := mock.NewMockProvider(mock.WithBaseURL(customURL))

	endpoint, err := p.GetEndpoint(protocols.Chat)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if endpoint != customURL+"/mock/endpoint" {
		t.Errorf("expected endpoint to start with %q, got %q", customURL, endpoint)
	}
}

func TestMockProvider_GetEndpoint_DefaultEndpoint(t *testing.T) {
	p := mock.NewMockProvider()

	endpoint, err := p.GetEndpoint(protocols.Chat)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := "http://mock-provider.local/mock/endpoint"
	if endpoint != expected {
		t.Errorf("expected endpoint %q, got %q", expected, endpoint)
	}
}

func TestMockProvider_GetEndpoint_CustomMapping(t *testing.T) {
	mapping := map[protocols.Protocol]string{
		protocols.Chat:       "/chat",
		protocols.Embeddings: "/embeddings",
	}

	p := mock.NewMockProvider(mock.WithEndpointMapping(mapping))

	chatEndpoint, err := p.GetEndpoint(protocols.Chat)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedChat := "http://mock-provider.local/chat"
	if chatEndpoint != expectedChat {
		t.Errorf("expected chat endpoint %q, got %q", expectedChat, chatEndpoint)
	}

	embeddingsEndpoint, err := p.GetEndpoint(protocols.Embeddings)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedEmbeddings := "http://mock-provider.local/embeddings"
	if embeddingsEndpoint != expectedEmbeddings {
		t.Errorf("expected embeddings endpoint %q, got %q", expectedEmbeddings, embeddingsEndpoint)
	}
}

func TestMockProvider_GetEndpoint_WithError(t *testing.T) {
	expectedError := errors.New("endpoint error")
	p := mock.NewMockProvider(mock.WithEndpointError(expectedError))

	endpoint, err := p.GetEndpoint(protocols.Chat)

	if err != expectedError {
		t.Errorf("expected error %v, got %v", expectedError, err)
	}

	if endpoint != "" {
		t.Errorf("expected empty endpoint, got %q", endpoint)
	}
}

func TestMockProvider_SetHeaders(t *testing.T) {
	headers := map[string]string{
		"Authorization": "Bearer test-token",
		"X-Custom":      "custom-value",
	}

	p := mock.NewMockProvider(mock.WithProviderHeaders(headers))

	req := httptest.NewRequest("GET", "http://test.local", nil)
	p.SetHeaders(req)

	if req.Header.Get("Authorization") != "Bearer test-token" {
		t.Error("expected Authorization header to be set")
	}

	if req.Header.Get("X-Custom") != "custom-value" {
		t.Error("expected X-Custom header to be set")
	}
}

func TestMockProvider_PrepareRequest_Success(t *testing.T) {
	expectedRequest := &providers.Request{
		URL:     "http://custom.local/test",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    []byte(`{"test": true}`),
	}

	p := mock.NewMockProvider(mock.WithPrepareResponse(expectedRequest, nil))

	ctx := context.Background()
	req := &protocols.Request{
		Messages: []protocols.Message{},
	}

	response, err := p.PrepareRequest(ctx, protocols.Chat, req)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if response != expectedRequest {
		t.Error("response does not match expected")
	}
}

func TestMockProvider_PrepareRequest_WithError(t *testing.T) {
	expectedError := errors.New("prepare error")
	p := mock.NewMockProvider(mock.WithPrepareResponse(nil, expectedError))

	ctx := context.Background()
	req := &protocols.Request{
		Messages: []protocols.Message{},
	}

	response, err := p.PrepareRequest(ctx, protocols.Chat, req)

	if err != expectedError {
		t.Errorf("expected error %v, got %v", expectedError, err)
	}

	if response != nil {
		t.Error("expected nil response")
	}
}

func TestMockProvider_PrepareStreamRequest_AddsStreamingHeaders(t *testing.T) {
	p := mock.NewMockProvider()

	ctx := context.Background()
	req := &protocols.Request{
		Messages: []protocols.Message{},
	}

	response, err := p.PrepareStreamRequest(ctx, protocols.Chat, req)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if response.Headers["Accept"] != "text/event-stream" {
		t.Error("expected Accept header for streaming")
	}

	if response.Headers["Cache-Control"] != "no-cache" {
		t.Error("expected Cache-Control header for streaming")
	}
}

func TestMockProvider_ProcessResponse_Success(t *testing.T) {
	expectedResponse := &protocols.ChatResponse{
		Model: "test-model",
	}

	p := mock.NewMockProvider(mock.WithProcessResponse(expectedResponse, nil))

	mockHTTPResponse := &http.Response{
		Body: http.NoBody,
	}
	capability := mock.NewMockCapability()

	response, err := p.ProcessResponse(mockHTTPResponse, capability)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if response != expectedResponse {
		t.Error("response does not match expected")
	}
}

func TestMockProvider_ProcessStreamResponse_Success(t *testing.T) {
	chunks := []any{"chunk1", "chunk2"}

	p := mock.NewMockProvider(mock.WithProviderStreamChunks(chunks, nil))

	ctx := context.Background()
	mockHTTPResponse := &http.Response{
		Body: http.NoBody,
	}
	capability := mock.NewMockCapability()

	ch, err := p.ProcessStreamResponse(ctx, mockHTTPResponse, capability)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	received := []any{}
	for chunk := range ch {
		received = append(received, chunk)
	}

	if len(received) != len(chunks) {
		t.Errorf("expected %d chunks, got %d", len(chunks), len(received))
	}
}

func TestMockProvider_ImplementsInterface(t *testing.T) {
	var _ providers.Provider = (*mock.MockProvider)(nil)
}
