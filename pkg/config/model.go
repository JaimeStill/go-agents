package config

type ModelConfig struct {
	Format  string         `json:"format,omitempty"`
	Name    string         `json:"name,omitempty"`
	Options map[string]any `json:"options,omitempty"`
}

func DefaultModelConfig() *ModelConfig {
	return &ModelConfig{
		Format:  "openai-standard",
		Options: make(map[string]any),
	}
}

func (c *ModelConfig) Merge(source *ModelConfig) {
	if source.Format != "" {
		c.Format = source.Format
	}

	if source.Name != "" {
		c.Name = source.Name
	}

	if source.Options != nil {
		c.Options = MergeOptions(c.Options, source.Options)
	}
}
