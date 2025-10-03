package config

import "maps"

// CapabilityConfig defines the configuration for a specific protocol capability.
// Format specifies the capability format name (e.g., "openai-chat"), and Options
// contains capability format-specific configuration parameters.
type CapabilityConfig struct {
	Format  string         `json:"format"`
	Options map[string]any `json:"options,omitempty"`
}

// ModelCapabilities maps protocol names to their capability configurations.
type ModelCapabilities map[string]CapabilityConfig

// ModelConfig defines the configuration for an LLM model.
// It includes the model name and a map of protocol capabilities.
type ModelConfig struct {
	Name         string            `json:"name,omitempty"`
	Capabilities ModelCapabilities `json:"capabilities,omitempty"`
}

// DefaultModelConfig creates a ModelConfig with initialized empty capabilities.
func DefaultModelConfig() *ModelConfig {
	return &ModelConfig{
		Capabilities: make(ModelCapabilities),
	}
}

// Merge combines the source ModelConfig into this ModelConfig.
// Non-empty name and capabilities from source override the current values.
func (c *ModelConfig) Merge(source *ModelConfig) {
	if source.Name != "" {
		c.Name = source.Name
	}

	if source.Capabilities != nil {
		if c.Capabilities == nil {
			c.Capabilities = make(ModelCapabilities)
		}
		maps.Copy(c.Capabilities, source.Capabilities)
	}
}
