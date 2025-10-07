package providers

import "github.com/JaimeStill/go-agents/pkg/models"

// BaseProvider provides common functionality for provider implementations.
// It stores the provider name, base URL, and model instance.
// Provider implementations typically embed BaseProvider to inherit this functionality.
type BaseProvider struct {
	name    string
	baseURL string
	model   models.Model
}

// NewBaseProvider creates a new BaseProvider with the given name, base URL, and model.
// This is typically called by provider constructors to initialize common fields.
func NewBaseProvider(name, baseURL string, model models.Model) *BaseProvider {
	return &BaseProvider{
		name:    name,
		baseURL: baseURL,
		model:   model,
	}
}

// Name returns the provider's identifier.
func (p *BaseProvider) Name() string {
	return p.name
}

// BaseURL returns the provider's base URL.
// Provider implementations use this to construct full endpoint URLs.
func (p *BaseProvider) BaseURL() string {
	return p.baseURL
}

// Model returns the model instance managed by this provider.
func (p *BaseProvider) Model() models.Model {
	return p.model
}
