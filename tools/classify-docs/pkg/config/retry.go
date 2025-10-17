package config

import (
	"time"

	acfg "github.com/JaimeStill/go-agents/pkg/config"
)

type RetryConfig struct {
	MaxAttempts       int           `json:"max_attempts"`
	InitialBackoff    acfg.Duration `json:"initial_backoff"`
	MaxBackoff        acfg.Duration `json:"max_backoff"`
	BackoffMultiplier float64       `json:"backoff_multiplier"`
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    acfg.Duration(13 * time.Second),
		MaxBackoff:        acfg.Duration(50 * time.Second),
		BackoffMultiplier: 1.2,
	}
}

func (c *RetryConfig) Merge(source *RetryConfig) {
	if source.MaxAttempts != 0 {
		c.MaxAttempts = source.MaxAttempts
	}
	if source.InitialBackoff != 0 {
		c.InitialBackoff = source.InitialBackoff
	}
	if source.MaxBackoff != 0 {
		c.MaxBackoff = source.MaxBackoff
	}
	if source.BackoffMultiplier != 0 {
		c.BackoffMultiplier = source.BackoffMultiplier
	}
}
