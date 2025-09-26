package models

import (
	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

type ModelFormat struct {
	Name         string
	Capabilities map[protocols.Protocol]capabilities.Capability
}

func NewModelFormat(name string) *ModelFormat {
	return &ModelFormat{
		Name:         name,
		Capabilities: make(map[protocols.Protocol]capabilities.Capability),
	}
}

func (f *ModelFormat) GetCapability(protocol protocols.Protocol) (capabilities.Capability, bool) {
	capability, exists := f.Capabilities[protocol]
	return capability, exists
}

func (f *ModelFormat) GetSupportedProtocols() []protocols.Protocol {
	protocols := make([]protocols.Protocol, 0, len(f.Capabilities))
	for protocol := range f.Capabilities {
		protocols = append(protocols, protocol)
	}
	return protocols
}

func (f *ModelFormat) SetCapability(protocol protocols.Protocol, capability capabilities.Capability) {
	if f.Capabilities == nil {
		f.Capabilities = make(map[protocols.Protocol]capabilities.Capability)
	}

	f.Capabilities[protocol] = capability
}

func (f *ModelFormat) SupportsProtocol(protocol protocols.Protocol) bool {
	_, exists := f.Capabilities[protocol]
	return exists
}
