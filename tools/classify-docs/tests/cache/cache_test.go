package cache_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/cache"
)

func TestNew(t *testing.T) {
	refDocs := []string{"doc1.pdf", "doc2.pdf"}
	prompt := "Test system prompt"

	c := cache.New(refDocs, prompt)

	if c.SystemPrompt != prompt {
		t.Errorf("expected prompt %q, got %q", prompt, c.SystemPrompt)
	}

	if len(c.ReferenceDocuments) != 2 {
		t.Errorf("expected 2 reference documents, got %d", len(c.ReferenceDocuments))
	}

	if c.GeneratedAt.IsZero() {
		t.Error("expected non-zero generation time")
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "test-cache.json")

	// Create cache
	refDocs := []string{"doc1.pdf", "doc2.pdf"}
	prompt := "Test system prompt"
	original := cache.New(refDocs, prompt)

	// Save
	if err := original.Save(cachePath); err != nil {
		t.Fatalf("failed to save cache: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Fatal("cache file was not created")
	}

	// Load
	loaded, err := cache.Load(cachePath)
	if err != nil {
		t.Fatalf("failed to load cache: %v", err)
	}

	// Verify contents
	if loaded.SystemPrompt != original.SystemPrompt {
		t.Errorf("expected prompt %q, got %q", original.SystemPrompt, loaded.SystemPrompt)
	}

	if len(loaded.ReferenceDocuments) != len(original.ReferenceDocuments) {
		t.Errorf("expected %d reference documents, got %d",
			len(original.ReferenceDocuments), len(loaded.ReferenceDocuments))
	}

	for i, doc := range original.ReferenceDocuments {
		if loaded.ReferenceDocuments[i] != doc {
			t.Errorf("reference document %d: expected %q, got %q",
				i, doc, loaded.ReferenceDocuments[i])
		}
	}

	// Verify timestamp preserved (with some tolerance for JSON marshaling)
	timeDiff := loaded.GeneratedAt.Sub(original.GeneratedAt)
	if timeDiff > time.Second || timeDiff < -time.Second {
		t.Errorf("timestamp not preserved: original=%v, loaded=%v",
			original.GeneratedAt, loaded.GeneratedAt)
	}
}

func TestLoad_NonexistentFile(t *testing.T) {
	_, err := cache.Load("/nonexistent/cache.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestSave_CreatesDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "nested", "dirs", "cache.json")

	c := cache.New([]string{"doc.pdf"}, "prompt")

	if err := c.Save(cachePath); err != nil {
		t.Fatalf("failed to save cache: %v", err)
	}

	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Error("cache file was not created in nested directory")
	}
}
