package config

type ProcessingConfig struct {
	Sequential SequentialConfig `json:"sequential"`
	Cache      CacheConfig      `json:"cache"`
}

func DefaultProcessingConfig() ProcessingConfig {
	return ProcessingConfig{
		Sequential: DefaultSequentialConfig(),
		Cache:      DefaultCacheConfig(),
	}
}

func (c *ProcessingConfig) Merge(source *ProcessingConfig) {
	c.Sequential.Merge(&source.Sequential)
	c.Cache.Merge(&source.Cache)
}

type SequentialConfig struct {
	ExposeIntermediateContexts bool `json:"expose_intermediate_contexts"`
}

func DefaultSequentialConfig() SequentialConfig {
	return SequentialConfig{
		ExposeIntermediateContexts: false,
	}
}

func (c *SequentialConfig) Merge(source *SequentialConfig) {
	c.ExposeIntermediateContexts = source.ExposeIntermediateContexts
}
