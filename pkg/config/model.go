package config

import "maps"

// ModelConfig defines the configuration for an LLM model.
// Name is the model identifier (e.g., "gpt-4o", "claude-3-opus", "llama3.1:8b").
// Capabilities maps protocol names to their default options.
//
// Example JSON:
//
//	{
//	  "name": "gpt-4o",
//	  "capabilities": {
//	    "chat": {
//	      "temperature": 0.7,
//	      "max_tokens": 4096
//	    },
//	    "vision": {
//	      "temperature": 0.5,
//	      "max_tokens": 2048
//	    }
//	  }
//	}
type ModelConfig struct {
	Name         string                      `json:"name,omitempty"`
	Capabilities map[string]map[string]any   `json:"capabilities,omitempty"`
}

// DefaultModelConfig creates a ModelConfig with initialized empty capabilities.
func DefaultModelConfig() *ModelConfig {
	return &ModelConfig{
		Capabilities: make(map[string]map[string]any),
	}
}

// Merge combines the source ModelConfig into this ModelConfig.
// Non-empty name from source overrides the current value.
// Capabilities are merged at the protocol level.
func (c *ModelConfig) Merge(source *ModelConfig) {
	if source.Name != "" {
		c.Name = source.Name
	}

	if source.Capabilities != nil {
		if c.Capabilities == nil {
			c.Capabilities = make(map[string]map[string]any)
		}

		// Merge each protocol's options
		for protocol, options := range source.Capabilities {
			if c.Capabilities[protocol] == nil {
				// Protocol doesn't exist, copy entire options map
				c.Capabilities[protocol] = options
			} else {
				// Protocol exists, merge options
				maps.Copy(c.Capabilities[protocol], options)
			}
		}
	}
}
