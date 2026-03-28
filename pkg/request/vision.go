package request

import (
	"github.com/JaimeStill/go-agents/pkg/format"
	"github.com/JaimeStill/go-agents/pkg/model"
	"github.com/JaimeStill/go-agents/pkg/protocol"
	"github.com/JaimeStill/go-agents/pkg/providers"
)

// VisionRequest represents a vision protocol request with image inputs.
// Separates images and vision-specific options from model configuration options.
type VisionRequest struct {
	provider      providers.Provider
	fmt           format.Format
	model         *model.Model
	messages      []protocol.Message
	images        []format.Image // URLs or base64 data URIs
	visionOptions map[string]any // Vision-specific options (e.g., detail: "high")
	options       map[string]any // Model configuration options
}

// NewVision creates a new VisionRequest with the given components.
// Messages contain the conversation history.
// Images are URLs or base64 data URIs to analyze.
// VisionOptions are vision-specific settings (e.g., detail level).
// Options specify model configuration (temperature, max_tokens, etc.).
func NewVision(
	p providers.Provider,
	f format.Format,
	m *model.Model,
	messages []protocol.Message,
	images []format.Image,
	visionOpts,
	opts map[string]any,
) *VisionRequest {
	return &VisionRequest{
		provider:      p,
		fmt:           f,
		model:         m,
		messages:      messages,
		images:        images,
		visionOptions: visionOpts,
		options:       opts,
	}
}

// Protocol returns the Vision protocol identifier.
func (r *VisionRequest) Protocol() protocol.Protocol {
	return protocol.Vision
}

// Headers returns the HTTP headers for a vision request.
func (r *VisionRequest) Headers() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

// Marshal delegates to the provider for provider-specific JSON formatting.
func (r *VisionRequest) Marshal() ([]byte, error) {
	return r.fmt.Marshal(protocol.Vision, &format.VisionData{
		Model:         r.model.Name,
		Messages:      r.messages,
		Images:        r.images,
		VisionOptions: r.visionOptions,
		Options:       r.options,
	})
}

// Provider returns the provider for this request.
func (r *VisionRequest) Provider() providers.Provider {
	return r.provider
}

func (r *VisionRequest) Format() format.Format {
	return r.fmt
}

// Model returns the model for this request.
func (r *VisionRequest) Model() *model.Model {
	return r.model
}
