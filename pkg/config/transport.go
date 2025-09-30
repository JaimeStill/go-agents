package config

import "time"

type TransportConfig struct {
	Provider           *ProviderConfig `json:"provider"`
	Timeout            Duration        `json:"timeout"`
	MaxRetries         int             `json:"max_retries"`
	RetryBackoffBase   Duration        `json:"retry_backoff_base"`
	ConnectionPoolSize int             `json:"connection_pool_size"`
	ConnectionTimeout  Duration        `json:"connection_timeout"`
}

func DefaultTransportConfig() *TransportConfig {
	return &TransportConfig{
		Provider:           DefaultProviderConfig(),
		Timeout:            Duration(2 * time.Minute),
		MaxRetries:         3,
		RetryBackoffBase:   Duration(1 * time.Second),
		ConnectionPoolSize: 10,
		ConnectionTimeout:  Duration(90 * time.Second),
	}
}

func (c *TransportConfig) Merge(source *TransportConfig) {
	if source.Timeout > 0 {
		c.Timeout = source.Timeout
	}

	if source.MaxRetries > 0 {
		c.MaxRetries = source.MaxRetries
	}

	if source.RetryBackoffBase > 0 {
		c.RetryBackoffBase = source.RetryBackoffBase
	}

	if source.ConnectionPoolSize > 0 {
		c.ConnectionPoolSize = source.ConnectionPoolSize
	}

	if source.ConnectionTimeout > 0 {
		c.ConnectionTimeout = source.ConnectionTimeout
	}

	if source.Provider != nil {
		if c.Provider == nil {
			c.Provider = source.Provider
		} else {
			c.Provider.Merge(source.Provider)
		}
	}
}
