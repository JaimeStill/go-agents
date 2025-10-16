package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type SystemPromptCache struct {
	GeneratedAt        time.Time `json:"generated_at"`
	ReferenceDocuments []string  `json:"reference_documents"`
	SystemPrompt       string    `json:"system_prompt"`
}

func New(docs []string, prompt string) *SystemPromptCache {
	return &SystemPromptCache{
		GeneratedAt:        time.Now(),
		ReferenceDocuments: docs,
		SystemPrompt:       prompt,
	}
}

func Load(path string) (*SystemPromptCache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache: %w", err)
	}

	var cache SystemPromptCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to parse cache: %w", err)
	}

	return &cache, nil
}

func (c *SystemPromptCache) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}

	return nil
}
