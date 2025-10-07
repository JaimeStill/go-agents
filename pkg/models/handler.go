package models

import (
	"maps"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
)

// ProtocolHandler manages a capability and its associated options for a protocol.
// Handles option storage, updates, and merging for protocol-specific configuration.
type ProtocolHandler struct {
	capability capabilities.Capability
	options    map[string]any
}

// NewProtocolHandler creates a new ProtocolHandler with a capability and options.
// The options map is cloned to prevent external modifications.
func NewProtocolHandler(capability capabilities.Capability, options map[string]any) *ProtocolHandler {
	return &ProtocolHandler{
		capability: capability,
		options:    maps.Clone(options),
	}
}

// Capability returns the capability instance managed by this handler.
func (h *ProtocolHandler) Capability() capabilities.Capability {
	return h.capability
}

// Options returns the handler's current options.
func (h *ProtocolHandler) Options() map[string]any {
	return h.options
}

// UpdateOptions updates the handler's options by copying provided options into the existing map.
// Existing options are preserved unless overridden by the provided options.
func (h *ProtocolHandler) UpdateOptions(options map[string]any) {
	maps.Copy(h.options, options)
}

// MergeOptions creates a new options map by merging handler options with provided options.
// The provided options take precedence over handler options.
// Neither the handler options nor the provided options are modified.
func (h *ProtocolHandler) MergeOptions(options map[string]any) map[string]any {
	merged := maps.Clone(h.options)
	maps.Copy(merged, options)
	return merged
}
