package config

// ProviderConfig defines the configuration for an LLM provider.
// It includes the provider name, base URL, model configuration, and
// provider-specific options (e.g., deployment, API version, authentication type).
type ProviderConfig struct {
	Name    string         `json:"name"`
	BaseURL string         `json:"base_url"`
	Model   *ModelConfig   `json:"model"`
	Options map[string]any `json:"options"`
}

// DefaultProviderConfig creates a ProviderConfig with Ollama defaults.
func DefaultProviderConfig() *ProviderConfig {
	return &ProviderConfig{
		Name:    "ollama",
		BaseURL: "http://localhost:11434",
		Model:   DefaultModelConfig(),
		Options: make(map[string]any),
	}
}

// Merge combines the source ProviderConfig into this ProviderConfig.
// Non-empty name, base_url, model, and options from source override the current values.
func (c *ProviderConfig) Merge(source *ProviderConfig) {
	if source.Name != "" {
		c.Name = source.Name
	}

	if source.BaseURL != "" {
		c.BaseURL = source.BaseURL
	}

	if source.Model != nil {
		if c.Model == nil {
			c.Model = source.Model
		} else {
			c.Model.Merge(source.Model)
		}
	}

	if source.Options != nil {
		c.Options = MergeOptions(c.Options, source.Options)
	}
}
