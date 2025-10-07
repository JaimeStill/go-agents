package capabilities

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/JaimeStill/go-agents/pkg/protocols"
)

// CapabilityRequest represents a request to be processed by a capability.
// It contains the protocol type, conversation messages, and capability-specific options.
type CapabilityRequest struct {
	Protocol protocols.Protocol
	Messages []protocols.Message
	Options  map[string]any
}

// CapabilityOption defines a configuration option for a capability.
// Options can be required or optional, and can specify default values.
type CapabilityOption struct {
	Option       string
	Required     bool
	DefaultValue any
}

// Capability defines the interface for all capability implementations.
// Each capability handles a specific protocol and manages request/response processing.
type Capability interface {
	Name() string
	Protocol() protocols.Protocol
	Options() []CapabilityOption

	ValidateOptions(options map[string]any) error
	ProcessOptions(options map[string]any) (map[string]any, error)

	CreateRequest(req *CapabilityRequest, model string) (*protocols.Request, error)
	ParseResponse(data []byte) (any, error)

	SupportsStreaming() bool
}

// StreamingCapability extends Capability with streaming support.
// Capabilities implementing this interface can handle both standard and streaming requests.
type StreamingCapability interface {
	Capability

	CreateStreamingRequest(req *CapabilityRequest, model string) (*protocols.Request, error)
	ParseStreamingChunk(data []byte) (*protocols.StreamingChunk, error)
	IsStreamComplete(data string) bool
}

// StandardCapability provides a base implementation for non-streaming capabilities.
// It handles option validation, processing, and default value application.
type StandardCapability struct {
	name     string
	protocol protocols.Protocol
	options  []CapabilityOption
}

// NewStandardCapability creates a new StandardCapability with the specified configuration.
// The name identifies the capability format, protocol specifies the operation type,
// and options define the available configuration parameters.
func NewStandardCapability(name string, protocol protocols.Protocol, options []CapabilityOption) *StandardCapability {
	return &StandardCapability{
		name:     name,
		protocol: protocol,
		options:  options,
	}
}

// Name returns the capability's format name.
func (s *StandardCapability) Name() string {
	return s.name
}

// Protocol returns the protocol type this capability implements.
func (s *StandardCapability) Protocol() protocols.Protocol {
	return s.protocol
}

// Options returns the list of configuration options supported by this capability.
func (s *StandardCapability) Options() []CapabilityOption {
	return s.options
}

// ValidateOptions verifies that provided options are supported and required options are present.
// Returns an error if unsupported options are provided or required options are missing.
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

// ProcessOptions validates and processes options, applying default values where needed.
// Returns a map containing all provided options plus defaults for unprovided optional parameters.
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

// SupportsStreaming returns false for StandardCapability.
// Override this method in capabilities that support streaming.
func (s *StandardCapability) SupportsStreaming() bool {
	return false
}

// StandardStreamingCapability extends StandardCapability with streaming support.
// It provides SSE (Server-Sent Events) parsing and stream completion detection.
type StandardStreamingCapability struct {
	*StandardCapability
}

// NewStandardStreamingCapability creates a new StandardStreamingCapability.
// This base implementation is suitable for most streaming protocols.
func NewStandardStreamingCapability(name string, protocol protocols.Protocol, options []CapabilityOption) *StandardStreamingCapability {
	return &StandardStreamingCapability{
		StandardCapability: NewStandardCapability(
			name,
			protocol,
			options,
		),
	}
}

// IsStreamComplete checks if the streaming data indicates completion.
// Returns true if the data contains the [DONE] marker.
func (s *StandardStreamingCapability) IsStreamComplete(data string) bool {
	return strings.Contains(data, "[DONE]")
}

// ParseStreamingChunk parses a single chunk from a streaming response.
// Handles both SSE format (with "data: " prefix) and plain JSON.
// Returns an error for empty lines or [DONE] markers which should be skipped.
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

// SupportsStreaming returns true for StandardStreamingCapability.
func (s *StandardStreamingCapability) SupportsStreaming() bool {
	return true
}
