package types

import "github.com/JaimeStill/go-agents/pkg/config"

// Model represents a configured LLM model at runtime.
// It stores the model name and protocol-specific default options.
// This is the domain type used during execution, separate from JSON configuration.
type Model struct {
	// Name is the model identifier (e.g., "gpt-4o", "claude-3-opus", "llama3.1:8b")
	Name string

	// Options holds protocol-specific default options.
	// Keys are protocols (Chat, Vision, Tools, Embeddings).
	// Values are option maps for that protocol (temperature, max_tokens, etc.)
	Options map[Protocol]map[string]any
}

// NewModel creates a new Model with the given name and empty options.
func NewModel(name string) *Model {
	return &Model{
		Name:    name,
		Options: make(map[Protocol]map[string]any),
	}
}

// FromConfig creates a Model from a ModelConfig.
// Handles conversion from string-keyed configuration to Protocol-keyed runtime model.
// This bridges the gap between JSON configuration structure and runtime domain type.
func FromConfig(cfg *config.ModelConfig) *Model {
	model := &Model{
		Name:    cfg.Name,
		Options: make(map[Protocol]map[string]any),
	}

	// Convert string keys to Protocol constants
	for protocolName, options := range cfg.Capabilities {
		protocol := Protocol(protocolName) // e.g., "chat" → types.Chat
		model.Options[protocol] = options
	}

	return model
}
