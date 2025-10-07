package providers_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/pkg/providers"
)

func TestRegister(t *testing.T) {
	// Register a test provider factory using Ollama implementation
	providers.Register("test-provider-1", providers.NewOllama)

	// Verify by creating it
	cfg := &config.ProviderConfig{
		Name:    "test-provider-1",
		BaseURL: "https://test.com",
		Model: &config.ModelConfig{
			Name: "test-model",
			Capabilities: map[string]config.CapabilityConfig{
				"chat": {Format: "openai-chat"},
			},
		},
	}

	provider, err := providers.Create(cfg)
	if err != nil {
		t.Errorf("failed to create registered provider: %v", err)
	}

	if provider == nil {
		t.Error("Create returned nil for registered provider")
	}
}

func TestCreate_RegisteredProvider(t *testing.T) {
	// Use the built-in "ollama" provider (registered in init)
	cfg := &config.ProviderConfig{
		Name:    "ollama",
		BaseURL: "http://localhost:11434",
		Model: &config.ModelConfig{
			Name: "llama2",
			Capabilities: map[string]config.CapabilityConfig{
				"chat": {Format: "openai-chat"},
			},
		},
	}

	provider, err := providers.Create(cfg)

	if err != nil {
		t.Fatalf("Create failed for registered provider: %v", err)
	}

	if provider == nil {
		t.Fatal("Create returned nil provider")
	}

	if provider.Name() != "ollama" {
		t.Errorf("got provider name %q, want %q", provider.Name(), "ollama")
	}
}

func TestCreate_UnknownProvider(t *testing.T) {
	cfg := &config.ProviderConfig{
		Name:    "unknown-provider",
		BaseURL: "https://test.com",
		Model: &config.ModelConfig{
			Name: "test-model",
			Capabilities: map[string]config.CapabilityConfig{
				"chat": {Format: "openai-chat"},
			},
		},
	}

	provider, err := providers.Create(cfg)

	if err == nil {
		t.Error("expected error for unknown provider, got nil")
	}

	if provider != nil {
		t.Error("expected nil provider for unknown provider")
	}

	expectedErrMsg := "unknown provider: unknown-provider"
	if err != nil && err.Error() != expectedErrMsg {
		t.Errorf("got error %q, want %q", err.Error(), expectedErrMsg)
	}
}

func TestListProviders(t *testing.T) {
	// Register a test provider to ensure the list is not empty
	providers.Register("test-provider-2", providers.NewOllama)

	list := providers.ListProviders()

	if len(list) == 0 {
		t.Fatal("ListProviders returned empty list")
	}

	// Check that built-in providers are in the list
	hasOllama := false
	hasAzure := false

	for _, name := range list {
		if name == "ollama" {
			hasOllama = true
		}
		if name == "azure" {
			hasAzure = true
		}
	}

	if !hasOllama {
		t.Error("ListProviders missing 'ollama' provider")
	}

	if !hasAzure {
		t.Error("ListProviders missing 'azure' provider")
	}
}

func TestRegistry_ConcurrentRegistration(t *testing.T) {
	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(index int) {
			defer wg.Done()

			providerName := fmt.Sprintf("concurrent-test-%d", index)
			providers.Register(providerName, providers.NewOllama)
		}(i)
	}

	wg.Wait()

	// Verify all providers were registered
	list := providers.ListProviders()

	for i := 0; i < goroutines; i++ {
		providerName := fmt.Sprintf("concurrent-test-%d", i)
		found := false
		for _, name := range list {
			if name == providerName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("provider %q not found in list after concurrent registration", providerName)
		}
	}
}

func TestRegistry_ConcurrentCreate(t *testing.T) {
	// Register a provider for concurrent creation testing
	providers.Register("concurrent-create-test", providers.NewOllama)

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	errors := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			cfg := &config.ProviderConfig{
				Name:    "concurrent-create-test",
				BaseURL: "https://test.com",
				Model: &config.ModelConfig{
					Name: "test-model",
					Capabilities: map[string]config.CapabilityConfig{
						"chat": {Format: "openai-chat"},
					},
				},
			}

			_, err := providers.Create(cfg)
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent Create failed: %v", err)
	}
}
