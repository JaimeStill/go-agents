package models

import (
	"maps"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
)

type ProtocolHandler struct {
	capability capabilities.Capability
	options    map[string]any
}

func NewProtocolHandler(capability capabilities.Capability, options map[string]any) *ProtocolHandler {
	return &ProtocolHandler{
		capability: capability,
		options:    maps.Clone(options),
	}
}

func (h *ProtocolHandler) Capability() capabilities.Capability {
	return h.capability
}

func (h *ProtocolHandler) Options() map[string]any {
	return h.options
}

func (h *ProtocolHandler) UpdateOptions(options map[string]any) {
	maps.Copy(h.options, options)
}

func (h *ProtocolHandler) MergeOptions(options map[string]any) map[string]any {
	merged := maps.Clone(h.options)
	maps.Copy(merged, options)
	return merged
}
