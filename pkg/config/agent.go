package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type AgentConfig struct {
	Name         string           `json:"name"`
	SystemPrompt string           `json:"system_prompt,omitempty"`
	Transport    *TransportConfig `json:"transport,omitempty"`
}

func DefaultAgentConfig() AgentConfig {
	transport := DefaultTransportConfig()
	return AgentConfig{
		Name:         "default-agent",
		SystemPrompt: "",
		Transport:    transport,
	}
}

func (c *AgentConfig) Merge(source *AgentConfig) {
	if source.Name != "" {
		c.Name = source.Name
	}

	if source.SystemPrompt != "" {
		c.SystemPrompt = source.SystemPrompt
	}

	if source.Transport != nil {
		if c.Transport == nil {
			c.Transport = source.Transport
		} else {
			c.Transport.Merge(source.Transport)
		}
	}
}

func LoadAgentConfig(filename string) (*AgentConfig, error) {
	config := DefaultAgentConfig()

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var loaded AgentConfig
	if err := json.Unmarshal(data, &loaded); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	config.Merge(&loaded)

	return &config, nil
}
