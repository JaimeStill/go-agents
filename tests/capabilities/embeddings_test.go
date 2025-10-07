package capabilities_test

import (
	"testing"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

func TestNewEmbeddingsCapability(t *testing.T) {
	options := []capabilities.CapabilityOption{
		{Option: "input", Required: true, DefaultValue: nil},
		{Option: "dimensions", Required: false, DefaultValue: nil},
	}

	cap := capabilities.NewEmbeddingsCapability("openai-embeddings", options)

	if cap.Name() != "openai-embeddings" {
		t.Errorf("got name %q, want %q", cap.Name(), "openai-embeddings")
	}

	if cap.Protocol() != protocols.Embeddings {
		t.Errorf("got protocol %q, want %q", cap.Protocol(), protocols.Embeddings)
	}

	if cap.SupportsStreaming() {
		t.Error("EmbeddingsCapability should not support streaming")
	}
}

func TestEmbeddingsCapability_CreateRequest(t *testing.T) {
	options := []capabilities.CapabilityOption{
		{Option: "input", Required: true, DefaultValue: nil},
		{Option: "dimensions", Required: false, DefaultValue: 1536},
		{Option: "encoding_format", Required: false, DefaultValue: "float"},
	}

	cap := capabilities.NewEmbeddingsCapability("openai-embeddings", options)

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Embeddings,
		Messages: []protocols.Message{},
		Options: map[string]any{
			"input":      "The quick brown fox",
			"dimensions": 768,
		},
	}

	protocolReq, err := cap.CreateRequest(req, "text-embedding-ada-002")
	if err != nil {
		t.Fatalf("CreateRequest failed: %v", err)
	}

	if model, exists := protocolReq.Options["model"]; !exists {
		t.Error("model option missing")
	} else if model != "text-embedding-ada-002" {
		t.Errorf("got model %q, want %q", model, "text-embedding-ada-002")
	}

	if input, exists := protocolReq.Options["input"]; !exists {
		t.Error("input option missing")
	} else if input != "The quick brown fox" {
		t.Errorf("got input %q, want %q", input, "The quick brown fox")
	}

	if dimensions, exists := protocolReq.Options["dimensions"]; !exists {
		t.Error("dimensions option missing")
	} else if dimensions != 768 {
		t.Errorf("got dimensions %v, want 768", dimensions)
	}

	if encodingFormat, exists := protocolReq.Options["encoding_format"]; !exists {
		t.Error("encoding_format default missing")
	} else if encodingFormat != "float" {
		t.Errorf("got encoding_format %v, want %q", encodingFormat, "float")
	}
}

func TestEmbeddingsCapability_CreateRequest_WithDefaults(t *testing.T) {
	options := []capabilities.CapabilityOption{
		{Option: "input", Required: true, DefaultValue: nil},
		{Option: "encoding_format", Required: false, DefaultValue: "float"},
	}

	cap := capabilities.NewEmbeddingsCapability("openai-embeddings", options)

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Embeddings,
		Messages: []protocols.Message{},
		Options: map[string]any{
			"input": "Test text",
		},
	}

	protocolReq, err := cap.CreateRequest(req, "text-embedding-ada-002")
	if err != nil {
		t.Fatalf("CreateRequest failed: %v", err)
	}

	if encodingFormat, exists := protocolReq.Options["encoding_format"]; !exists {
		t.Error("encoding_format default should be applied")
	} else if encodingFormat != "float" {
		t.Errorf("got encoding_format %v, want %q", encodingFormat, "float")
	}
}

func TestEmbeddingsCapability_ParseResponse(t *testing.T) {
	cap := capabilities.NewEmbeddingsCapability("openai-embeddings", nil)

	responseData := []byte(`{
		"object": "list",
		"data": [{
			"object": "embedding",
			"embedding": [0.1, 0.2, 0.3, 0.4, 0.5],
			"index": 0
		}],
		"model": "text-embedding-ada-002",
		"usage": {
			"prompt_tokens": 8,
			"total_tokens": 8
		}
	}`)

	result, err := cap.ParseResponse(responseData)
	if err != nil {
		t.Fatalf("ParseResponse failed: %v", err)
	}

	response, ok := result.(*protocols.EmbeddingsResponse)
	if !ok {
		t.Fatal("result is not an EmbeddingsResponse")
	}

	if response.Object != "list" {
		t.Errorf("got object %q, want %q", response.Object, "list")
	}

	if response.Model != "text-embedding-ada-002" {
		t.Errorf("got model %q, want %q", response.Model, "text-embedding-ada-002")
	}

	if len(response.Data) != 1 {
		t.Fatalf("got %d data items, want 1", len(response.Data))
	}

	if len(response.Data[0].Embedding) != 5 {
		t.Fatalf("got %d embedding dimensions, want 5", len(response.Data[0].Embedding))
	}

	if response.Data[0].Embedding[0] != 0.1 {
		t.Errorf("got embedding[0] %f, want 0.1", response.Data[0].Embedding[0])
	}

	if response.Usage == nil {
		t.Error("usage should not be nil")
	} else if response.Usage.PromptTokens != 8 {
		t.Errorf("got prompt tokens %d, want 8", response.Usage.PromptTokens)
	}
}

func TestEmbeddingsCapability_ParseResponse_InvalidJSON(t *testing.T) {
	cap := capabilities.NewEmbeddingsCapability("openai-embeddings", nil)

	responseData := []byte(`{invalid json}`)

	_, err := cap.ParseResponse(responseData)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestEmbeddingsCapability_CreateRequest_ValidationError(t *testing.T) {
	options := []capabilities.CapabilityOption{
		{Option: "input", Required: true, DefaultValue: nil},
	}

	cap := capabilities.NewEmbeddingsCapability("openai-embeddings", options)

	req := &capabilities.CapabilityRequest{
		Protocol: protocols.Embeddings,
		Messages: []protocols.Message{},
		Options: map[string]any{
			// Missing required input
		},
	}

	_, err := cap.CreateRequest(req, "text-embedding-ada-002")
	if err == nil {
		t.Error("expected validation error for missing input, got nil")
	}
}
