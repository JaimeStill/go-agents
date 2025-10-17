package prompt_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/config"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/prompt"
)

func TestGenerate_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultClassifyConfig()

	_, err := prompt.Generate(context.Background(), cfg, tmpDir)
	if err == nil {
		t.Fatal("expected error for empty directory")
	}

	if !strings.Contains(err.Error(), "no PDF files found") {
		t.Errorf("expected 'no PDF files found' in error, got: %v", err)
	}
}

func TestGenerate_NonexistentDirectory(t *testing.T) {
	cfg := config.DefaultClassifyConfig()

	_, err := prompt.Generate(context.Background(), cfg, "/nonexistent/path/to/directory")
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("expected 'does not exist' in error, got: %v", err)
	}
}

func TestGenerate_NotADirectory(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "notadir.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cfg := config.DefaultClassifyConfig()

	_, err := prompt.Generate(context.Background(), cfg, tmpFile)
	if err == nil {
		t.Fatal("expected error when path is not a directory")
	}

	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("expected 'not a directory' in error, got: %v", err)
	}
}
