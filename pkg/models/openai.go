package models

import (
	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

func OpenAIStandardFormat() *ModelFormat {
	format := NewModelFormat("openai-standard")

	chatOptions := []capabilities.CapabilityOption{
		{Option: "max_tokens", Required: false, DefaultValue: 4096},
		{Option: "temperature", Required: false, DefaultValue: 0.7},
		{Option: "top_p", Required: false, DefaultValue: nil},
		{Option: "frequency_penalty", Required: false, DefaultValue: nil},
		{Option: "presence_penalty", Required: false, DefaultValue: nil},
		{Option: "stop", Required: false, DefaultValue: nil},
		{Option: "stream", Required: false, DefaultValue: false},
	}

	format.SetCapability(
		protocols.Chat,
		capabilities.NewChatCapability("standard", chatOptions),
	)

	visionOptions := []capabilities.CapabilityOption{
		{Option: "images", Required: true, DefaultValue: nil},
		{Option: "max_tokens", Required: false, DefaultValue: 4096},
		{Option: "temperature", Required: false, DefaultValue: 0.7},
		{Option: "detail", Required: false, DefaultValue: "auto"},
		{Option: "stream", Required: false, DefaultValue: false},
	}
	format.SetCapability(protocols.Vision, capabilities.NewVisionCapability("standard", visionOptions))

	// Tools with function calling options
	toolsOptions := []capabilities.CapabilityOption{
		{Option: "tools", Required: true, DefaultValue: nil},
		{Option: "tool_choice", Required: false, DefaultValue: "auto"},
		{Option: "max_tokens", Required: false, DefaultValue: 4096},
		{Option: "temperature", Required: false, DefaultValue: 0.7},
		{Option: "stream", Required: false, DefaultValue: false},
	}
	format.SetCapability(protocols.Tools, capabilities.NewToolsCapability("standard", toolsOptions))

	// Embeddings with OpenAI embedding options
	embeddingsOptions := []capabilities.CapabilityOption{
		{Option: "input", Required: true, DefaultValue: nil},
		{Option: "dimensions", Required: false, DefaultValue: nil},
		{Option: "encoding_format", Required: false, DefaultValue: "float"},
	}
	format.SetCapability(protocols.Embeddings, capabilities.NewEmbeddingsCapability("standard", embeddingsOptions))

	return format
}

func OpenAIChatFormat() *ModelFormat {
	format := NewModelFormat("openai-chat")

	chatOptions := []capabilities.CapabilityOption{
		{Option: "max_tokens", Required: false, DefaultValue: 4096},
		{Option: "temperature", Required: false, DefaultValue: 0.7},
		{Option: "top_p", Required: false, DefaultValue: nil},
		{Option: "frequency_penalty", Required: false, DefaultValue: nil},
		{Option: "presence_penalty", Required: false, DefaultValue: nil},
		{Option: "stop", Required: false, DefaultValue: nil},
		{Option: "stream", Required: false, DefaultValue: false},
	}

	format.SetCapability(
		protocols.Chat,
		capabilities.NewChatCapability("chat", chatOptions),
	)

	return format
}

func OpenAIReasoningFormat() *ModelFormat {
	format := NewModelFormat("openai-reasoning")

	// Chat with reasoning-specific options (no temperature/top_p)
	reasoningOptions := []capabilities.CapabilityOption{
		{Option: "max_completion_tokens", Required: true, DefaultValue: nil},
		{Option: "stream", Required: false, DefaultValue: false},
	}
	format.SetCapability(protocols.Chat, capabilities.NewChatCapability("reasoning", reasoningOptions))

	return format
}

func OpenAIEmbeddingsFormat() *ModelFormat {
	format := NewModelFormat("openai-embeddings")

	// Embeddings-only format with specific options
	embeddingsOptions := []capabilities.CapabilityOption{
		{Option: "input", Required: true, DefaultValue: nil},
		{Option: "dimensions", Required: false, DefaultValue: nil},
		{Option: "encoding_format", Required: false, DefaultValue: "float"},
	}
	format.SetCapability(protocols.Embeddings, capabilities.NewEmbeddingsCapability("embeddings-only", embeddingsOptions))

	return format
}
