package providers

import "github.com/JaimeStill/go-agents/pkg/models"

type BaseProvider struct {
	name    string
	baseURL string
	model   models.Model
}

func NewBaseProvider(name, baseURL string, model models.Model) *BaseProvider {
	return &BaseProvider{
		name:    name,
		baseURL: baseURL,
		model:   model,
	}
}

func (p *BaseProvider) Name() string {
	return p.name
}

func (p *BaseProvider) BaseURL() string {
	return p.baseURL
}

func (p *BaseProvider) Model() models.Model {
	return p.model
}
