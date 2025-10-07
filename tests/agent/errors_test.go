package agent_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/JaimeStill/go-agents/pkg/agent"
	"github.com/JaimeStill/go-agents/pkg/config"
)

func TestNewAgentError(t *testing.T) {
	err := agent.NewAgentError(agent.ErrorTypeInit, "test error")

	if err == nil {
		t.Fatal("NewAgentError returned nil")
	}

	if err.Error() == "" {
		t.Error("Error() returned empty string")
	}

	if err.Message != "test error" {
		t.Errorf("got message %q, want %q", err.Message, "test error")
	}

	if err.Type != agent.ErrorTypeInit {
		t.Errorf("got type %q, want %q", err.Type, agent.ErrorTypeInit)
	}

	if err.Timestamp.IsZero() {
		t.Error("Timestamp was not set")
	}
}

func TestNewAgentError_WithCode(t *testing.T) {
	err := agent.NewAgentError(agent.ErrorTypeLLM, "test error", agent.WithCode("E001"))

	if err.Code != "E001" {
		t.Errorf("got code %q, want %q", err.Code, "E001")
	}
}

func TestNewAgentError_WithCause(t *testing.T) {
	cause := errors.New("underlying error")
	err := agent.NewAgentError(agent.ErrorTypeLLM, "test error", agent.WithCause(cause))

	if err.Cause != cause {
		t.Error("Cause was not set correctly")
	}

	if err.Unwrap() != cause {
		t.Error("Unwrap() did not return cause")
	}
}

func TestNewAgentError_WithName(t *testing.T) {
	err := agent.NewAgentError(agent.ErrorTypeInit, "test error", agent.WithName("test-agent"))

	if err.Name != "test-agent" {
		t.Errorf("got name %q, want %q", err.Name, "test-agent")
	}
}

func TestNewAgentError_WithClient(t *testing.T) {
	cfg := &config.TransportConfig{
		Provider: &config.ProviderConfig{
			Name: "ollama",
			Model: &config.ModelConfig{
				Name: "test-model",
			},
		},
	}

	err := agent.NewAgentError(agent.ErrorTypeLLM, "test error", agent.WithClient(cfg))

	expected := "ollama/test-model"
	if err.Client != expected {
		t.Errorf("got client %q, want %q", err.Client, expected)
	}
}

func TestNewAgentError_WithClient_ProviderOnly(t *testing.T) {
	cfg := &config.TransportConfig{
		Provider: &config.ProviderConfig{
			Name: "ollama",
			Model: &config.ModelConfig{
				Name: "",
			},
		},
	}

	err := agent.NewAgentError(agent.ErrorTypeLLM, "test error", agent.WithClient(cfg))

	if err.Client != "ollama" {
		t.Errorf("got client %q, want %q", err.Client, "ollama")
	}
}

func TestNewAgentError_WithClient_ModelOnly(t *testing.T) {
	cfg := &config.TransportConfig{
		Provider: &config.ProviderConfig{
			Name: "",
			Model: &config.ModelConfig{
				Name: "test-model",
			},
		},
	}

	err := agent.NewAgentError(agent.ErrorTypeLLM, "test error", agent.WithClient(cfg))

	if err.Client != "test-model" {
		t.Errorf("got client %q, want %q", err.Client, "test-model")
	}
}

func TestNewAgentError_WithClient_Unknown(t *testing.T) {
	cfg := &config.TransportConfig{
		Provider: &config.ProviderConfig{
			Name: "",
			Model: &config.ModelConfig{
				Name: "",
			},
		},
	}

	err := agent.NewAgentError(agent.ErrorTypeLLM, "test error", agent.WithClient(cfg))

	if err.Client != "unknown" {
		t.Errorf("got client %q, want %q", err.Client, "unknown")
	}
}

func TestNewAgentError_WithID(t *testing.T) {
	id := uuid.New()
	err := agent.NewAgentError(agent.ErrorTypeInit, "test error", agent.WithID(id))

	if err.ID != id {
		t.Errorf("got ID %v, want %v", err.ID, id)
	}
}

func TestNewAgentError_MultipleOptions(t *testing.T) {
	cause := errors.New("underlying")
	id := uuid.New()

	err := agent.NewAgentError(
		agent.ErrorTypeLLM,
		"test error",
		agent.WithCode("E001"),
		agent.WithCause(cause),
		agent.WithName("test-agent"),
		agent.WithID(id),
	)

	if err.Code != "E001" {
		t.Errorf("got code %q, want %q", err.Code, "E001")
	}

	if err.Cause != cause {
		t.Error("Cause was not set")
	}

	if err.Name != "test-agent" {
		t.Errorf("got name %q, want %q", err.Name, "test-agent")
	}

	if err.ID != id {
		t.Errorf("got ID %v, want %v", err.ID, id)
	}
}

func TestAgentError_Error_WithClientAndName(t *testing.T) {
	cfg := &config.TransportConfig{
		Provider: &config.ProviderConfig{
			Name: "ollama",
			Model: &config.ModelConfig{
				Name: "test-model",
			},
		},
	}

	err := agent.NewAgentError(
		agent.ErrorTypeLLM,
		"test error",
		agent.WithClient(cfg),
		agent.WithName("test-agent"),
	)

	errStr := err.Error()
	expected := "Agent error [ollama/test-model/test-agent]: test error"

	if errStr != expected {
		t.Errorf("got error string %q, want %q", errStr, expected)
	}
}

func TestAgentError_Error_WithName(t *testing.T) {
	err := agent.NewAgentError(
		agent.ErrorTypeInit,
		"test error",
		agent.WithName("test-agent"),
	)

	errStr := err.Error()
	expected := "Agent error [test-agent]: test error"

	if errStr != expected {
		t.Errorf("got error string %q, want %q", errStr, expected)
	}
}

func TestAgentError_Error_NoName(t *testing.T) {
	err := agent.NewAgentError(agent.ErrorTypeLLM, "test error")

	errStr := err.Error()
	expected := "Agent error: test error"

	if errStr != expected {
		t.Errorf("got error string %q, want %q", errStr, expected)
	}
}

func TestAgentError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := agent.NewAgentError(agent.ErrorTypeLLM, "test error", agent.WithCause(cause))

	unwrapped := err.Unwrap()

	if unwrapped != cause {
		t.Error("Unwrap did not return cause")
	}
}

func TestAgentError_Unwrap_NoCause(t *testing.T) {
	err := agent.NewAgentError(agent.ErrorTypeLLM, "test error")

	unwrapped := err.Unwrap()

	if unwrapped != nil {
		t.Error("Unwrap should return nil when no cause is set")
	}
}

func TestNewAgentInitError(t *testing.T) {
	err := agent.NewAgentInitError("initialization failed")

	if err == nil {
		t.Fatal("NewAgentInitError returned nil")
	}

	if err.Type != agent.ErrorTypeInit {
		t.Errorf("got type %q, want %q", err.Type, agent.ErrorTypeInit)
	}

	if err.Message != "initialization failed" {
		t.Errorf("got message %q, want %q", err.Message, "initialization failed")
	}
}

func TestNewAgentInitError_WithOptions(t *testing.T) {
	err := agent.NewAgentInitError(
		"initialization failed",
		agent.WithCode("INIT001"),
		agent.WithName("test-agent"),
	)

	if err.Type != agent.ErrorTypeInit {
		t.Errorf("got type %q, want %q", err.Type, agent.ErrorTypeInit)
	}

	if err.Code != "INIT001" {
		t.Errorf("got code %q, want %q", err.Code, "INIT001")
	}

	if err.Name != "test-agent" {
		t.Errorf("got name %q, want %q", err.Name, "test-agent")
	}
}

func TestNewAgentLLMError(t *testing.T) {
	err := agent.NewAgentLLMError("LLM request failed")

	if err == nil {
		t.Fatal("NewAgentLLMError returned nil")
	}

	if err.Type != agent.ErrorTypeLLM {
		t.Errorf("got type %q, want %q", err.Type, agent.ErrorTypeLLM)
	}

	if err.Message != "LLM request failed" {
		t.Errorf("got message %q, want %q", err.Message, "LLM request failed")
	}
}

func TestNewAgentLLMError_WithOptions(t *testing.T) {
	cause := errors.New("network error")
	err := agent.NewAgentLLMError(
		"LLM request failed",
		agent.WithCode("LLM500"),
		agent.WithCause(cause),
	)

	if err.Type != agent.ErrorTypeLLM {
		t.Errorf("got type %q, want %q", err.Type, agent.ErrorTypeLLM)
	}

	if err.Code != "LLM500" {
		t.Errorf("got code %q, want %q", err.Code, "LLM500")
	}

	if err.Cause != cause {
		t.Error("Cause was not set")
	}
}

func TestAgentError_Timestamp(t *testing.T) {
	before := time.Now()
	err := agent.NewAgentError(agent.ErrorTypeInit, "test error")
	after := time.Now()

	if err.Timestamp.Before(before) || err.Timestamp.After(after) {
		t.Error("Timestamp is not within expected range")
	}
}
