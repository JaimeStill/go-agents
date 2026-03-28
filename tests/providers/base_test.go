package providers_test

import (
	"testing"

	"github.com/JaimeStill/go-agents/pkg/providers"
)

func TestNewBaseProvider(t *testing.T) {
	provider := providers.NewBaseProvider("test-provider", "https://api.example.com")

	if provider == nil {
		t.Fatal("NewBaseProvider returned nil")
	}

	if provider.Name() != "test-provider" {
		t.Errorf("got name %q, want %q", provider.Name(), "test-provider")
	}

	if provider.BaseURL() != "https://api.example.com" {
		t.Errorf("got baseURL %q, want %q", provider.BaseURL(), "https://api.example.com")
	}
}

func TestBaseProvider_Name(t *testing.T) {
	provider := providers.NewBaseProvider("my-provider", "https://api.test.com")

	if provider.Name() != "my-provider" {
		t.Errorf("got name %q, want %q", provider.Name(), "my-provider")
	}
}

func TestBaseProvider_BaseURL(t *testing.T) {
	provider := providers.NewBaseProvider("test", "https://custom.api.com/v2")

	if provider.BaseURL() != "https://custom.api.com/v2" {
		t.Errorf("got baseURL %q, want %q", provider.BaseURL(), "https://custom.api.com/v2")
	}
}
