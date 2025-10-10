package mock

import (
	"fmt"
	"maps"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/models"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

// MockModel implements models.Model interface for testing.
type MockModel struct {
	name string

	// Protocol support configuration
	supportedProtocols map[protocols.Protocol]bool

	// Protocol capabilities and options
	capabilities map[protocols.Protocol]capabilities.Capability
	options      map[protocols.Protocol]map[string]any

	// Error configuration
	getCapabilityError error
	updateOptionsError error
}

// NewMockModel creates a new MockModel with default configuration.
// By default, supports all protocols with mock capabilities.
func NewMockModel(opts ...MockModelOption) *MockModel {
	m := &MockModel{
		name:               "mock-model",
		supportedProtocols: make(map[protocols.Protocol]bool),
		capabilities:       make(map[protocols.Protocol]capabilities.Capability),
		options:            make(map[protocols.Protocol]map[string]any),
	}

	// Default: support all protocols
	m.supportedProtocols[protocols.Chat] = true
	m.supportedProtocols[protocols.Vision] = true
	m.supportedProtocols[protocols.Tools] = true
	m.supportedProtocols[protocols.Embeddings] = true

	// Set default mock capabilities
	m.capabilities[protocols.Chat] = NewMockCapability()
	m.capabilities[protocols.Vision] = NewMockCapability()
	m.capabilities[protocols.Tools] = NewMockCapability()
	m.capabilities[protocols.Embeddings] = NewMockCapability()

	// Set default empty options
	m.options[protocols.Chat] = make(map[string]any)
	m.options[protocols.Vision] = make(map[string]any)
	m.options[protocols.Tools] = make(map[string]any)
	m.options[protocols.Embeddings] = make(map[string]any)

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// MockModelOption configures a MockModel.
type MockModelOption func(*MockModel)

// WithModelName sets the model name.
func WithModelName(name string) MockModelOption {
	return func(m *MockModel) {
		m.name = name
	}
}

// WithSupportedProtocols sets which protocols are supported.
func WithSupportedProtocols(protos ...protocols.Protocol) MockModelOption {
	return func(m *MockModel) {
		// Clear existing
		m.supportedProtocols = make(map[protocols.Protocol]bool)
		for _, p := range protos {
			m.supportedProtocols[p] = true
		}
	}
}

// WithProtocolCapability sets a capability for a specific protocol.
func WithProtocolCapability(protocol protocols.Protocol, capability capabilities.Capability) MockModelOption {
	return func(m *MockModel) {
		m.capabilities[protocol] = capability
		m.supportedProtocols[protocol] = true
	}
}

// WithProtocolOptions sets default options for a protocol.
func WithProtocolOptions(protocol protocols.Protocol, options map[string]any) MockModelOption {
	return func(m *MockModel) {
		m.options[protocol] = options
	}
}

// WithGetCapabilityError sets an error for GetCapability.
func WithGetCapabilityError(err error) MockModelOption {
	return func(m *MockModel) {
		m.getCapabilityError = err
	}
}

// WithUpdateOptionsError sets an error for UpdateProtocolOptions.
func WithUpdateOptionsError(err error) MockModelOption {
	return func(m *MockModel) {
		m.updateOptionsError = err
	}
}

// Name returns the model name.
func (m *MockModel) Name() string {
	return m.name
}

// SupportsProtocol checks if the protocol is supported.
func (m *MockModel) SupportsProtocol(p protocols.Protocol) bool {
	return m.supportedProtocols[p]
}

// GetCapability returns the capability for a protocol.
func (m *MockModel) GetCapability(p protocols.Protocol) (capabilities.Capability, error) {
	if m.getCapabilityError != nil {
		return nil, m.getCapabilityError
	}

	if !m.SupportsProtocol(p) {
		return nil, fmt.Errorf("protocol %s not supported by model %s", p, m.name)
	}

	capability, ok := m.capabilities[p]
	if !ok {
		return nil, fmt.Errorf("no capability configured for protocol %s", p)
	}

	return capability, nil
}

// GetProtocolOptions returns the default options for a protocol.
func (m *MockModel) GetProtocolOptions(p protocols.Protocol) map[string]any {
	if !m.SupportsProtocol(p) {
		return make(map[string]any)
	}

	opts, ok := m.options[p]
	if !ok {
		return make(map[string]any)
	}

	return opts
}

// UpdateProtocolOptions updates the options for a protocol.
func (m *MockModel) UpdateProtocolOptions(p protocols.Protocol, options map[string]any) error {
	if m.updateOptionsError != nil {
		return m.updateOptionsError
	}

	if !m.SupportsProtocol(p) {
		return fmt.Errorf("protocol %s not supported by model %s", p, m.name)
	}

	m.options[p] = options
	return nil
}

// MergeRequestOptions merges model options with request options.
// Request options take precedence.
func (m *MockModel) MergeRequestOptions(p protocols.Protocol, options map[string]any) map[string]any {
	if !m.SupportsProtocol(p) {
		return options
	}

	modelOpts := m.GetProtocolOptions(p)
	merged := make(map[string]any)

	// Copy model defaults
	maps.Copy(merged, modelOpts)

	// Override with request options
	maps.Copy(merged, options)

	return merged
}

// Verify MockModel implements models.Model interface.
var _ models.Model = (*MockModel)(nil)
