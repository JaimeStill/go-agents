package models

import (
	"sync"
)

type formatRegistry struct {
	formats map[string]*ModelFormat
	mu      sync.RWMutex
}

var (
	registry = &formatRegistry{
		formats: make(map[string]*ModelFormat),
	}
)

func GetFormat(name string) (*ModelFormat, bool) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	format, exists := registry.formats[name]
	return format, exists
}

func ListFormats() []string {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	names := make([]string, 0, len(registry.formats))
	for name := range registry.formats {
		names = append(names, name)
	}

	return names
}

func RegisterFormat(name string, format *ModelFormat) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.formats[name] = format
}

func SetDefaultFormat(provider string) string {
	defaults := map[string]string{
		"ollama": "openai-standard",
		"azure":  "openai-standard",
	}

	if format, exists := defaults[provider]; exists {
		return format
	}
	return "openai-standard"
}

func init() {
	RegisterFormat("openai-standard", OpenAIStandardFormat())
	RegisterFormat("openai-chat", OpenAIChatFormat())
	RegisterFormat("openai-reasoning", OpenAIReasoningFormat())
	RegisterFormat("openai-embeddings", OpenAIEmbeddingsFormat())
}
