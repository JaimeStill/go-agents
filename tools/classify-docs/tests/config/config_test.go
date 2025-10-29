package config_test

import (
	"os"
	"path/filepath"
	"testing"

	acfg "github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/config"
)

func TestDefaultConfigs(t *testing.T) {
	t.Run("DefaultSequentialConfig", func(t *testing.T) {
		cfg := config.DefaultSequentialConfig()

		if cfg.ExposeIntermediateContexts != false {
			t.Error("expected ExposeIntermediateContexts=false")
		}
	})

	t.Run("DefaultCacheConfig", func(t *testing.T) {
		cfg := config.DefaultCacheConfig()

		if !cfg.IsEnabled() {
			t.Error("expected Enabled=true")
		}

		if cfg.Path != ".cache/system-prompt.json" {
			t.Errorf("expected Path=.cache/system-prompt.json, got %s", cfg.Path)
		}
	})
}

func TestLoadClassifyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.json")

	// Write test config file
	testConfig := `{
		"agent": {
			"name": "test-agent"
		},
		"processing": {
			"cache": {
				"enabled": true,
				"path": ".cache/custom.json"
			}
		}
	}`
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Load config
	loaded, err := config.LoadClassifyConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify merged values
	if loaded.Agent.Name != "test-agent" {
		t.Errorf("expected agent name 'test-agent', got %s", loaded.Agent.Name)
	}

	if loaded.Processing.Cache.Path != ".cache/custom.json" {
		t.Errorf("expected cache path '.cache/custom.json', got %s",
			loaded.Processing.Cache.Path)
	}
}

func TestLoadClassifyConfig_NonexistentFile(t *testing.T) {
	_, err := config.LoadClassifyConfig("/nonexistent/config.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestClassifyConfig_Merge(t *testing.T) {
	base := config.DefaultClassifyConfig()
	base.Agent.Name = "base-agent"

	override := config.ClassifyConfig{
		Agent: acfg.AgentConfig{
			Name: "override-agent",
		},
	}

	base.Merge(&override)

	// Check merged values
	if base.Agent.Name != "override-agent" {
		t.Errorf("expected merged name 'override-agent', got %s", base.Agent.Name)
	}
}
