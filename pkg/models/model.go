package models

import (
	"fmt"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

// Model provides a protocol-agnostic interface for working with LLM models.
// Each model supports one or more protocols (chat, vision, tools, embeddings)
// and manages protocol-specific capabilities and options.
type Model interface {
	// Name returns the model's identifier.
	Name() string

	// SupportsProtocol checks if the model supports a given protocol.
	SupportsProtocol(p protocols.Protocol) bool

	// GetCapability returns the capability instance for a protocol.
	// Returns an error if the protocol is not supported.
	GetCapability(p protocols.Protocol) (capabilities.Capability, error)

	// GetProtocolOptions returns the model's default options for a protocol.
	// Returns an empty map if the protocol is not supported.
	GetProtocolOptions(p protocols.Protocol) map[string]any

	// UpdateProtocolOptions updates the model's default options for a protocol.
	// Validates options against the capability's requirements.
	// Returns an error if the protocol is not supported or options are invalid.
	UpdateProtocolOptions(p protocols.Protocol, options map[string]any) error

	// MergeRequestOptions merges model options with request options for a protocol.
	// Request options take precedence over model defaults.
	// Returns the request options unchanged if the protocol is not supported.
	MergeRequestOptions(p protocols.Protocol, options map[string]any) map[string]any
}

type model struct {
	name string

	chat       *ProtocolHandler
	vision     *ProtocolHandler
	tools      *ProtocolHandler
	embeddings *ProtocolHandler
}

// New creates a Model from configuration.
// Validates that all protocols are valid and capability formats exist.
// Returns an error if configuration is invalid or capabilities cannot be loaded.
func New(cfg *config.ModelConfig) (Model, error) {
	m := &model{
		name: cfg.Name,
	}

	for proto, cap := range cfg.Capabilities {
		if !protocols.IsValid(proto) {
			return nil, fmt.Errorf(
				"invalid protocol in configuration: %s (valid protocols: %s)",
				proto,
				protocols.ProtocolStrings(),
			)
		}

		protocol := protocols.Protocol(proto)

		capability, err := capabilities.GetFormat(cap.Format)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to get capability format '%s' for protocol %s: %w",
				cap.Format, protocol, err,
			)
		}

		handler := NewProtocolHandler(capability, cap.Options)

		switch protocol {
		case protocols.Chat:
			m.chat = handler
		case protocols.Vision:
			m.vision = handler
		case protocols.Tools:
			m.tools = handler
		case protocols.Embeddings:
			m.embeddings = handler
		default:
			return nil, fmt.Errorf("unhandled protocol: %s", protocol)
		}
	}

	return m, nil
}

func (m *model) Name() string {
	return m.name
}

func (m *model) SupportsProtocol(p protocols.Protocol) bool {
	return m.getHandler(p) != nil
}

func (m *model) GetCapability(p protocols.Protocol) (capabilities.Capability, error) {
	handler := m.getHandler(p)
	if handler == nil {
		return nil, fmt.Errorf("protocol %s not supported by model %s", p, m.name)
	}
	return handler.Capability(), nil
}

func (m *model) GetProtocolOptions(p protocols.Protocol) map[string]any {
	handler := m.getHandler(p)
	if handler == nil {
		return make(map[string]any)
	}
	return handler.Options()
}

func (m *model) UpdateProtocolOptions(p protocols.Protocol, options map[string]any) error {
	handler := m.getHandler(p)
	if handler == nil {
		return fmt.Errorf("protocol %s not supported by model %s", p, m.name)
	}

	if err := handler.Capability().ValidateOptions(options); err != nil {
		return fmt.Errorf("invalid options for %s protocol: %w", p, err)
	}

	handler.UpdateOptions(options)
	return nil
}

func (m *model) MergeRequestOptions(p protocols.Protocol, options map[string]any) map[string]any {
	handler := m.getHandler(p)
	if handler == nil {
		return options
	}
	return handler.MergeOptions(options)
}

func (m *model) getHandler(p protocols.Protocol) *ProtocolHandler {
	switch p {
	case protocols.Chat:
		return m.chat
	case protocols.Vision:
		return m.vision
	case protocols.Tools:
		return m.tools
	case protocols.Embeddings:
		return m.embeddings
	default:
		return nil
	}
}
