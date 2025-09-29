package capabilities

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/JaimeStill/go-agents/pkg/protocols"
)

type CapabilityRequest struct {
	Protocol protocols.Protocol
	Messages []protocols.Message
	Options  map[string]any
}

type CapabilityOption struct {
	Option       string
	Required     bool
	DefaultValue any
}

type Capability interface {
	Name() string
	Protocol() protocols.Protocol
	Options() []CapabilityOption

	ValidateOptions(options map[string]any) error
	ProcessOptions(options map[string]any) (map[string]any, error)

	CreateRequest(req *CapabilityRequest, model ModelInfo) (*protocols.Request, error)
	ParseResponse(data []byte) (any, error)

	SupportsStreaming() bool
}

type StreamingCapability interface {
	Capability

	CreateStreamingRequest(req *CapabilityRequest, model ModelInfo) (*protocols.Request, error)
	ParseStreamingChunk(data []byte) (*protocols.StreamingChunk, error)
	IsStreamComplete(data string) bool
}

type ModelInfo interface {
	Name() string
	Options() map[string]any
}

type StandardCapability struct {
	name     string
	protocol protocols.Protocol
	options  []CapabilityOption
}

func NewStandardCapability(name string, protocol protocols.Protocol, options []CapabilityOption) *StandardCapability {
	return &StandardCapability{
		name:     name,
		protocol: protocol,
		options:  options,
	}
}

func (s *StandardCapability) Name() string {
	return s.name
}

func (s *StandardCapability) Protocol() protocols.Protocol {
	return s.protocol
}

func (s *StandardCapability) Options() []CapabilityOption {
	return s.options
}

func (s *StandardCapability) ValidateOptions(options map[string]any) error {
	accepted := make(map[string]bool)
	required := make([]string, 0)

	for _, opt := range s.options {
		accepted[opt.Option] = true
		if opt.Required {
			required = append(required, opt.Option)
		}
	}

	for key := range options {
		if !accepted[key] {
			return fmt.Errorf("unsupported option: %s", key)
		}
	}

	for _, req := range required {
		if _, provided := options[req]; !provided {
			return fmt.Errorf("required option missing: %s", req)
		}
	}

	return nil
}

func (s *StandardCapability) ProcessOptions(options map[string]any) (map[string]any, error) {
	if err := s.ValidateOptions(options); err != nil {
		return nil, err
	}

	result := make(map[string]any)
	for _, opt := range s.options {
		if value, provided := options[opt.Option]; provided {
			result[opt.Option] = value
		} else if opt.DefaultValue != nil {
			result[opt.Option] = opt.DefaultValue
		}
	}

	return result, nil
}

func (s *StandardCapability) SupportsStreaming() bool {
	return false
}

type StandardStreamingCapability struct {
	*StandardCapability
}

func NewStandardStreamingCapability(name string, protocol protocols.Protocol, options []CapabilityOption) *StandardStreamingCapability {
	return &StandardStreamingCapability{
		StandardCapability: NewStandardCapability(
			name,
			protocol,
			options,
		),
	}
}

func (s *StandardStreamingCapability) IsStreamComplete(data string) bool {
	return strings.Contains(data, "[DONE]")
}

func (s *StandardStreamingCapability) ParseStreamingChunk(data []byte) (*protocols.StreamingChunk, error) {
	line := string(data)

	// Handle Server-Sent Events format
	if after, ok := strings.CutPrefix(line, "data: "); ok {
		line = after
	}

	// Skip empty lines or [DONE] markers
	if line == "" || strings.Contains(line, "[DONE]") {
		return nil, fmt.Errorf("skip line")
	}

	var chunk protocols.StreamingChunk
	if err := json.Unmarshal([]byte(line), &chunk); err != nil {
		return nil, err
	}
	return &chunk, nil
}

func (s *StandardStreamingCapability) SupportsStreaming() bool {
	return true
}
