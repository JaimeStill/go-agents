package config

type CacheConfig struct {
	Enabled *bool  `json:"enabled,omitempty"`
	Path    string `json:"path,omitempty"`
}

func DefaultCacheConfig() CacheConfig {
	enabled := true
	return CacheConfig{
		Enabled: &enabled,
		Path:    ".cache/system-prompt.json",
	}
}

func (c *CacheConfig) Merge(source *CacheConfig) {
	// Only merge Enabled if explicitly set in source
	if source.Enabled != nil {
		c.Enabled = source.Enabled
	}
	if source.Path != "" {
		c.Path = source.Path
	}
}

// IsEnabled returns the cache enabled status, handling nil pointer
func (c *CacheConfig) IsEnabled() bool {
	if c.Enabled == nil {
		return true // Default to enabled
	}
	return *c.Enabled
}
