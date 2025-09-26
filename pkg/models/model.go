package models

import (
	"fmt"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

type Model interface {
	Name() string
	Options() map[string]any
	Format() string

	SupportsProtocol(p protocols.Protocol) bool
	GetCapability(p protocols.Protocol) (capabilities.Capability, error)
}

type model struct {
	name    string
	options map[string]any
	format  *ModelFormat
}

func New(config *config.ModelConfig) (Model, error) {
	format, exists := GetFormat(config.Format)
	if !exists {
		return nil, fmt.Errorf("unknown model format: %s", config.Format)
	}

	return &model{
		name:    config.Name,
		options: config.Options,
		format:  format,
	}, nil
}

func (m *model) Name() string {
	return m.name
}

func (m *model) Options() map[string]any {
	return m.options
}

func (m *model) Format() string {
	return m.format.Name
}

func (m *model) SupportsProtocol(p protocols.Protocol) bool {
	return m.format.SupportsProtocol(p)
}

func (m *model) GetCapability(p protocols.Protocol) (capabilities.Capability, error) {
	capability, exists := m.format.GetCapability(p)
	if !exists {
		return nil, fmt.Errorf(
			"protocol %s not supported by model %s (format: %s)",
			p,
			m.name,
			m.format.Name,
		)
	}

	return capability, nil
}
