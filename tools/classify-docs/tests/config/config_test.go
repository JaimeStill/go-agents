package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	acfg "github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/config"
)

func TestDefaultConfigs(t *testing.T) {
	t.Run("DefaultRetryConfig", func(t *testing.T) {
		cfg := config.DefaultRetryConfig()

		if cfg.MaxAttempts != 3 {
			t.Errorf("expected MaxAttempts=3, got %d", cfg.MaxAttempts)
		}

		if cfg.InitialBackoff.ToDuration() != 13*time.Second {
			t.Errorf("expected InitialBackoff=13s, got %v", cfg.InitialBackoff.ToDuration())
		}

		if cfg.MaxBackoff.ToDuration() != 50*time.Second {
			t.Errorf("expected MaxBackoff=50s, got %v", cfg.MaxBackoff.ToDuration())
		}

		if cfg.BackoffMultiplier != 1.2 {
			t.Errorf("expected BackoffMultiplier=1.2, got %v", cfg.BackoffMultiplier)
		}
	})

	t.Run("DefaultParallelConfig", func(t *testing.T) {
		cfg := config.DefaultParallelConfig()

		if cfg.WorkerCap != 16 {
			t.Errorf("expected WorkerCap=16, got %d", cfg.WorkerCap)
		}
	})

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
			"parallel": {
				"worker_cap": 8
			},
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

	if loaded.Processing.Parallel.WorkerCap != 8 {
		t.Errorf("expected WorkerCap=8, got %d", loaded.Processing.Parallel.WorkerCap)
	}

	if loaded.Processing.Cache.Path != ".cache/custom.json" {
		t.Errorf("expected cache path '.cache/custom.json', got %s",
			loaded.Processing.Cache.Path)
	}

	// Verify defaults for unspecified values
	if loaded.Processing.Retry.MaxAttempts != 3 {
		t.Errorf("expected default MaxAttempts=3, got %d",
			loaded.Processing.Retry.MaxAttempts)
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
	base.Processing.Parallel.WorkerCap = 4

	override := config.ClassifyConfig{
		Agent: acfg.AgentConfig{
			Name: "override-agent",
		},
		Processing: config.ProcessingConfig{
			Parallel: config.ParallelConfig{
				WorkerCap: 8,
			},
		},
	}

	base.Merge(&override)

	// Check merged values
	if base.Agent.Name != "override-agent" {
		t.Errorf("expected merged name 'override-agent', got %s", base.Agent.Name)
	}

	if base.Processing.Parallel.WorkerCap != 8 {
		t.Errorf("expected merged WorkerCap=8, got %d", base.Processing.Parallel.WorkerCap)
	}

	// Check that unspecified values retain defaults
	if base.Processing.Retry.MaxAttempts != 3 {
		t.Errorf("expected default MaxAttempts=3, got %d",
			base.Processing.Retry.MaxAttempts)
	}
}
