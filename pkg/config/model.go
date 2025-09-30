package config

import "maps"

type CapabilityConfig struct {
	Format  string         `json:"format"`
	Options map[string]any `json:"options,omitempty"`
}

type ModelCapabilities map[string]CapabilityConfig

type ModelConfig struct {
	Name         string            `json:"name,omitempty"`
	Capabilities ModelCapabilities `json:"capabilities,omitempty"`
}

func DefaultModelConfig() *ModelConfig {
	return &ModelConfig{
		Capabilities: make(ModelCapabilities),
	}
}

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
