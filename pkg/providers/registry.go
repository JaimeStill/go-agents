package providers

import (
	"fmt"
	"sync"

	"github.com/JaimeStill/go-agents/pkg/config"
)

type Factory func(c *config.ProviderConfig) (Provider, error)

type registry struct {
	factories map[string]Factory
	mu        sync.RWMutex
}

var register = &registry{
	factories: make(map[string]Factory),
}

func Register(name string, factory Factory) {
	register.mu.Lock()
	defer register.mu.Unlock()
	register.factories[name] = factory
}

func Create(c *config.ProviderConfig) (Provider, error) {
	register.mu.RLock()
	factory, exists := register.factories[c.Name]
	register.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unknown provider: %s", c.Name)
	}

	return factory(c)
}

func ListProviders() []string {
	register.mu.RLock()
	defer register.mu.RUnlock()

	names := make([]string, 0, len(register.factories))
	for name := range register.factories {
		names = append(names, name)
	}
	return names
}

func init() {
	Register("ollama", NewOllama)
	Register("azure", NewAzure)
}
