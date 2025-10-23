package capabilities

func init() {
	// Standard Formats Supported by Most LLMs
	RegisterFormat("chat", func() Capability {
		return NewChatCapability("chat", []CapabilityOption{
			{Option: "max_tokens", Required: false, DefaultValue: 4096},
			{Option: "temperature", Required: false, DefaultValue: 0.7},
			{Option: "top_p", Required: false, DefaultValue: nil},
			{Option: "frequency_penalty", Required: false, DefaultValue: nil},
			{Option: "presence_penalty", Required: false, DefaultValue: nil},
			{Option: "stop", Required: false, DefaultValue: nil},
			{Option: "stream", Required: false, DefaultValue: false},
		})
	})

	RegisterFormat("vision", func() Capability {
		return NewVisionCapability("vision", []CapabilityOption{
			{Option: "images", Required: true, DefaultValue: nil},
			{Option: "max_tokens", Required: false, DefaultValue: 4096},
			{Option: "temperature", Required: false, DefaultValue: 0.7},
			{Option: "detail", Required: false, DefaultValue: "auto"},
			{Option: "stream", Required: false, DefaultValue: false},
		})
	})

	RegisterFormat("tools", func() Capability {
		return NewToolsCapability("tools", []CapabilityOption{
			{Option: "tools", Required: true, DefaultValue: nil},
			{Option: "tool_choice", Required: false, DefaultValue: "auto"},
			{Option: "max_tokens", Required: false, DefaultValue: 4096},
			{Option: "temperature", Required: false, DefaultValue: 0.7},
		})
	})

	RegisterFormat("embeddings", func() Capability {
		return NewEmbeddingsCapability("embeddings", []CapabilityOption{
			{Option: "input", Required: true, DefaultValue: nil},
			{Option: "dimensions", Required: false, DefaultValue: nil},
			{Option: "encoding_format", Required: false, DefaultValue: "float"},
		})
	})

	// OpenAI o-series reasoning models
	RegisterFormat("o-chat", func() Capability {
		return NewChatCapability("o-chat", []CapabilityOption{
			{Option: "max_completion_tokens", Required: false, DefaultValue: 4096},
			{Option: "reasoning_effort", Required: false, DefaultValue: "medium"},
			{Option: "stream", Required: false, DefaultValue: false},
		})
	})

	RegisterFormat("o-vision", func() Capability {
		return NewVisionCapability("o-vision", []CapabilityOption{
			{Option: "max_completion_tokens", Required: false, DefaultValue: 4096},
			{Option: "images", Required: true, DefaultValue: nil},
			{Option: "detail", Required: false, DefaultValue: "auto"},
			{Option: "reasoning_effort", Required: false, DefaultValue: "medium"},
			{Option: "stream", Required: false, DefaultValue: false},
		})
	})
}
