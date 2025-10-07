package capabilities

import (
	"fmt"
	"sync"
)

// CapabilityFactory is a function that creates a new Capability instance.
// Used by the registry to instantiate capabilities on demand.
type CapabilityFactory func() Capability

type capabilityRegistry struct {
	mu        sync.RWMutex
	factories map[string]CapabilityFactory
}

var registry = &capabilityRegistry{
	factories: make(map[string]CapabilityFactory),
}

// RegisterFormat registers a capability format with the global registry.
// The factory function will be called each time the format is retrieved,
// creating a new capability instance. Thread-safe for concurrent registration.
func RegisterFormat(name string, factory CapabilityFactory) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.factories[name] = factory
}

// GetFormat retrieves a capability by format name from the registry.
// Returns a new instance of the capability created by the registered factory.
// Returns an error if the format is not registered. Thread-safe for concurrent access.
func GetFormat(name string) (Capability, error) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	factory, exists := registry.factories[name]
	if !exists {
		return nil, fmt.Errorf("capability format '%s' not registered", name)
	}

	return factory(), nil
}

// ListFormats returns a list of all registered format names.
// The order of formats in the returned slice is not guaranteed.
// Thread-safe for concurrent access.
func ListFormats() []string {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	names := make([]string, 0, len(registry.factories))
	for name := range registry.factories {
		names = append(names, name)
	}
	return names
}
