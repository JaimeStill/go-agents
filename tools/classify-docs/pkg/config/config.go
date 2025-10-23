package config

import (
	"encoding/json"
	"fmt"
	"os"

	acfg "github.com/JaimeStill/go-agents/pkg/config"
)

type ClassifyConfig struct {
	Agent      acfg.AgentConfig `json:"agent"`
	Processing ProcessingConfig `json:"processing"`
}

func DefaultClassifyConfig() ClassifyConfig {
	return ClassifyConfig{
		Agent:      acfg.DefaultAgentConfig(),
		Processing: DefaultProcessingConfig(),
	}
}

func LoadClassifyConfig(path string) (*ClassifyConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := DefaultClassifyConfig()

	var fileConfig ClassifyConfig
	if err := json.Unmarshal(data, &fileConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	cfg.Merge(&fileConfig)

	return &cfg, nil
}

func (c *ClassifyConfig) Merge(source *ClassifyConfig) {
	c.Agent.Merge(&source.Agent)
	c.Processing.Merge(&source.Processing)
}
