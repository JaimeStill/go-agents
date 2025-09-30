package capabilities

func init() {
	RegisterFormat("openai-chat", func() Capability {
		return NewChatCapability("openai-chat", []CapabilityOption{
			{Option: "max_tokens", Required: false, DefaultValue: 4096},
			{Option: "temperature", Required: false, DefaultValue: 0.7},
			{Option: "top_p", Required: false, DefaultValue: nil},
			{Option: "frequency_penalty", Required: false, DefaultValue: nil},
			{Option: "presence_penalty", Required: false, DefaultValue: nil},
			{Option: "stop", Required: false, DefaultValue: nil},
			{Option: "stream", Required: false, DefaultValue: false},
		})
	})

	RegisterFormat("openai-vision", func() Capability {
		return NewVisionCapability("openai-vision", []CapabilityOption{
			{Option: "images", Required: true, DefaultValue: nil},
			{Option: "max_tokens", Required: false, DefaultValue: 4096},
			{Option: "temperature", Required: false, DefaultValue: 0.7},
			{Option: "detail", Required: false, DefaultValue: "auto"},
			{Option: "stream", Required: false, DefaultValue: false},
		})
	})

	RegisterFormat("openai-tools", func() Capability {
		return NewToolsCapability("openai-tools", []CapabilityOption{
			{Option: "tools", Required: true, DefaultValue: nil},
			{Option: "tool_choice", Required: false, DefaultValue: "auto"},
			{Option: "max_tokens", Required: false, DefaultValue: 4096},
			{Option: "temperature", Required: false, DefaultValue: 0.7},
		})
	})

	RegisterFormat("openai-embeddings", func() Capability {
		return NewEmbeddingsCapability("openai-embeddings", []CapabilityOption{
			{Option: "input", Required: true, DefaultValue: nil},
			{Option: "dimensions", Required: false, DefaultValue: nil},
			{Option: "encoding_format", Required: false, DefaultValue: "float"},
		})
	})

	RegisterFormat("openai-reasoning", func() Capability {
		return NewChatCapability("openai-reasoning", []CapabilityOption{
			{Option: "max_completion_tokens", Required: true, DefaultValue: nil},
			{Option: "stream", Required: false, DefaultValue: false},
		})
	})
}
