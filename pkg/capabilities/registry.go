package capabilities

import (
	"fmt"
	"sync"
)

type CapabilityFactory func() Capability

type capabilityRegistry struct {
	mu        sync.RWMutex
	factories map[string]CapabilityFactory
}

var registry = &capabilityRegistry{
	factories: make(map[string]CapabilityFactory),
}

func RegisterFormat(name string, factory CapabilityFactory) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.factories[name] = factory
}

func GetFormat(name string) (Capability, error) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	factory, exists := registry.factories[name]
	if !exists {
		return nil, fmt.Errorf("capability format '%s' not registered", name)
	}

	return factory(), nil
}

func ListFormats() []string {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	names := make([]string, 0, len(registry.factories))
	for name := range registry.factories {
		names = append(names, name)
	}
	return names
}
