package config

type ProcessingConfig struct {
	Parallel   ParallelConfig   `json:"parallel"`
	Sequential SequentialConfig `json:"sequential"`
	Retry      RetryConfig      `json:"retry"`
	Cache      CacheConfig      `json:"cache"`
}

func DefaultProcessingConfig() ProcessingConfig {
	return ProcessingConfig{
		Parallel:   DefaultParallelConfig(),
		Sequential: DefaultSequentialConfig(),
		Retry:      DefaultRetryConfig(),
		Cache:      DefaultCacheConfig(),
	}
}

func (c *ProcessingConfig) Merge(source *ProcessingConfig) {
	c.Parallel.Merge(&source.Parallel)
	c.Sequential.Merge(&source.Sequential)
	c.Retry.Merge(&source.Retry)
	c.Cache.Merge(&source.Cache)
}

type ParallelConfig struct {
	WorkerCap int `json:"worker_cap"`
}

func DefaultParallelConfig() ParallelConfig {
	return ParallelConfig{
		WorkerCap: 16,
	}
}

func (c *ParallelConfig) Merge(source *ParallelConfig) {
	if source.WorkerCap != 0 {
		c.WorkerCap = source.WorkerCap
	}
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
