package agent

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/JaimeStill/go-agents/pkg/config"
)

type ErrorType string

const (
	ErrorTypeInit ErrorType = "init"
	ErrorTypeLLM  ErrorType = "llm"
)

type AgentError struct {
	Type      ErrorType `json:"type"`
	ID        uuid.UUID `json:"uuid,omitempty"`
	Name      string    `json:"name,omitempty"`
	Code      string    `json:"code,omitempty"`
	Message   string    `json:"message"`
	Cause     error     `json:"-"`
	Client    string    `json:"client,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

func NewAgentError(errorType ErrorType, message string, options ...ErrorOption) *AgentError {
	err := &AgentError{
		Type:      errorType,
		Message:   message,
		Timestamp: time.Now(),
	}

	for _, option := range options {
		option(err)
	}

	return err
}

func (e *AgentError) Error() string {
	if e.Client != "" && e.Name != "" {
		return fmt.Sprintf("Agent error [%s/%s]: %s", e.Client, e.Name, e.Message)
	}
	if e.Name != "" {
		return fmt.Sprintf("Agent error [%s]: %s", e.Name, e.Message)
	}

	return fmt.Sprintf("Agent error: %s", e.Message)
}

func (e *AgentError) Unwrap() error {
	return e.Cause
}

type ErrorOption func(*AgentError)

func WithCode(code string) ErrorOption {
	return func(e *AgentError) {
		e.Code = code
	}
}

func WithCause(cause error) ErrorOption {
	return func(e *AgentError) {
		e.Cause = cause
	}
}

func WithName(name string) ErrorOption {
	return func(e *AgentError) {
		e.Name = name
	}
}

func WithClient(client *config.TransportConfig) ErrorOption {
	return func(e *AgentError) {
		if client.Provider.Name != "" && client.Provider.Model.Name != "" {
			e.Client = fmt.Sprintf("%s/%s", client.Provider.Name, client.Provider.Model.Name)
		} else if client.Provider.Name != "" {
			e.Client = client.Provider.Name
		} else if client.Provider.Model.Name != "" {
			e.Client = client.Provider.Model.Name
		} else {
			e.Client = "unknown"
		}
	}
}

func WithID(id uuid.UUID) ErrorOption {
	return func(e *AgentError) {
		e.ID = id
	}
}

func NewAgentInitError(message string, options ...ErrorOption) *AgentError {
	return NewAgentError(ErrorTypeInit, message, options...)
}

func NewAgentLLMError(message string, options ...ErrorOption) *AgentError {
	return NewAgentError(ErrorTypeLLM, message, options...)
}
