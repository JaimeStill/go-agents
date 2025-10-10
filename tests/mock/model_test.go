package mock_test

import (
	"errors"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/mock"
	"github.com/JaimeStill/go-agents/pkg/models"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

func TestNewMockModel_DefaultConfiguration(t *testing.T) {
	m := mock.NewMockModel()

	if m.Name() != "mock-model" {
		t.Errorf("expected name 'mock-model', got %q", m.Name())
	}

	// Should support all protocols by default
	if !m.SupportsProtocol(protocols.Chat) {
		t.Error("expected Chat protocol support")
	}
	if !m.SupportsProtocol(protocols.Vision) {
		t.Error("expected Vision protocol support")
	}
	if !m.SupportsProtocol(protocols.Tools) {
		t.Error("expected Tools protocol support")
	}
	if !m.SupportsProtocol(protocols.Embeddings) {
		t.Error("expected Embeddings protocol support")
	}
}

func TestMockModel_WithModelName(t *testing.T) {
	customName := "custom-model"
	m := mock.NewMockModel(mock.WithModelName(customName))

	if m.Name() != customName {
		t.Errorf("expected name %q, got %q", customName, m.Name())
	}
}

func TestMockModel_WithSupportedProtocols(t *testing.T) {
	m := mock.NewMockModel(mock.WithSupportedProtocols(
		protocols.Chat,
		protocols.Vision,
	))

	if !m.SupportsProtocol(protocols.Chat) {
		t.Error("expected Chat protocol support")
	}
	if !m.SupportsProtocol(protocols.Vision) {
		t.Error("expected Vision protocol support")
	}
	if m.SupportsProtocol(protocols.Tools) {
		t.Error("expected Tools protocol not supported")
	}
	if m.SupportsProtocol(protocols.Embeddings) {
		t.Error("expected Embeddings protocol not supported")
	}
}

func TestMockModel_GetCapability_Success(t *testing.T) {
	customCapability := mock.NewMockCapability(
		mock.WithCapabilityName("custom-capability"),
	)

	m := mock.NewMockModel(mock.WithProtocolCapability(
		protocols.Chat,
		customCapability,
	))

	capability, err := m.GetCapability(protocols.Chat)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if capability != customCapability {
		t.Error("capability does not match expected")
	}

	if capability.Name() != "custom-capability" {
		t.Errorf("expected capability name 'custom-capability', got %q", capability.Name())
	}
}

func TestMockModel_GetCapability_UnsupportedProtocol(t *testing.T) {
	m := mock.NewMockModel(mock.WithSupportedProtocols(protocols.Chat))

	capability, err := m.GetCapability(protocols.Vision)

	if err == nil {
		t.Error("expected error for unsupported protocol")
	}

	if capability != nil {
		t.Error("expected nil capability")
	}
}

func TestMockModel_GetCapability_WithError(t *testing.T) {
	expectedError := errors.New("capability error")
	m := mock.NewMockModel(mock.WithGetCapabilityError(expectedError))

	capability, err := m.GetCapability(protocols.Chat)

	if err != expectedError {
		t.Errorf("expected error %v, got %v", expectedError, err)
	}

	if capability != nil {
		t.Error("expected nil capability")
	}
}

func TestMockModel_GetProtocolOptions(t *testing.T) {
	options := map[string]any{
		"temperature": 0.8,
		"max_tokens":  1000,
	}

	m := mock.NewMockModel(mock.WithProtocolOptions(protocols.Chat, options))

	result := m.GetProtocolOptions(protocols.Chat)

	if result["temperature"] != 0.8 {
		t.Error("expected temperature option")
	}

	if result["max_tokens"] != 1000 {
		t.Error("expected max_tokens option")
	}
}

func TestMockModel_GetProtocolOptions_UnsupportedProtocol(t *testing.T) {
	m := mock.NewMockModel(mock.WithSupportedProtocols(protocols.Chat))

	result := m.GetProtocolOptions(protocols.Vision)

	if len(result) != 0 {
		t.Error("expected empty options for unsupported protocol")
	}
}

func TestMockModel_UpdateProtocolOptions_Success(t *testing.T) {
	m := mock.NewMockModel()

	newOptions := map[string]any{
		"temperature": 0.9,
	}

	err := m.UpdateProtocolOptions(protocols.Chat, newOptions)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	result := m.GetProtocolOptions(protocols.Chat)

	if result["temperature"] != 0.9 {
		t.Error("expected updated temperature option")
	}
}

func TestMockModel_UpdateProtocolOptions_UnsupportedProtocol(t *testing.T) {
	m := mock.NewMockModel(mock.WithSupportedProtocols(protocols.Chat))

	err := m.UpdateProtocolOptions(protocols.Vision, map[string]any{})

	if err == nil {
		t.Error("expected error for unsupported protocol")
	}
}

func TestMockModel_UpdateProtocolOptions_WithError(t *testing.T) {
	expectedError := errors.New("update error")
	m := mock.NewMockModel(mock.WithUpdateOptionsError(expectedError))

	err := m.UpdateProtocolOptions(protocols.Chat, map[string]any{})

	if err != expectedError {
		t.Errorf("expected error %v, got %v", expectedError, err)
	}
}

func TestMockModel_MergeRequestOptions(t *testing.T) {
	modelOptions := map[string]any{
		"temperature": 0.7,
		"max_tokens":  500,
	}

	m := mock.NewMockModel(mock.WithProtocolOptions(protocols.Chat, modelOptions))

	requestOptions := map[string]any{
		"temperature": 0.9, // Override
		"top_p":       0.95, // Additional
	}

	merged := m.MergeRequestOptions(protocols.Chat, requestOptions)

	// Request options should override model options
	if merged["temperature"] != 0.9 {
		t.Error("expected request temperature to override model temperature")
	}

	// Model options should be preserved
	if merged["max_tokens"] != 500 {
		t.Error("expected model max_tokens to be preserved")
	}

	// Request options should be added
	if merged["top_p"] != 0.95 {
		t.Error("expected request top_p to be added")
	}
}

func TestMockModel_MergeRequestOptions_UnsupportedProtocol(t *testing.T) {
	m := mock.NewMockModel(mock.WithSupportedProtocols(protocols.Chat))

	requestOptions := map[string]any{
		"temperature": 0.9,
	}

	merged := m.MergeRequestOptions(protocols.Vision, requestOptions)

	// Should return request options unchanged for unsupported protocol
	if merged["temperature"] != 0.9 {
		t.Error("expected request options to be returned unchanged")
	}
}

func TestMockModel_ImplementsInterface(t *testing.T) {
	var _ models.Model = (*mock.MockModel)(nil)
}
