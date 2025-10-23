package main

import (
	"fmt"
	"log"
	"os"

	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/config"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: test-config <config-file>")
	}

	configPath := os.Args[1]

	cfg, err := config.LoadClassifyConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Configuration loaded successfully from: %s\n\n", configPath)
	fmt.Printf("Agent Configuration:\n")
	fmt.Printf("  Name: %s\n", cfg.Agent.Name)
	fmt.Printf("  Provider: %s\n", cfg.Agent.Transport.Provider.Name)
	fmt.Printf("  Model: %s\n", cfg.Agent.Transport.Provider.Model.Name)
	fmt.Printf("  Base URL: %s\n", cfg.Agent.Transport.Provider.BaseURL)

	fmt.Printf("\nProcessing Configuration (with defaults):\n")
	fmt.Printf("  Sequential:\n")
	fmt.Printf("    Expose Intermediate Contexts: %v (default: false)\n", cfg.Processing.Sequential.ExposeIntermediateContexts)
	fmt.Printf("  Retry:\n")
	fmt.Printf("    Max Attempts: %d (default: 3)\n", cfg.Processing.Retry.MaxAttempts)
	fmt.Printf("    Initial Backoff: %s (default: 1s)\n", cfg.Processing.Retry.InitialBackoff.ToDuration())
	fmt.Printf("    Max Backoff: %s (default: 30s)\n", cfg.Processing.Retry.MaxBackoff.ToDuration())
	fmt.Printf("    Backoff Multiplier: %.1f (default: 2.0)\n", cfg.Processing.Retry.BackoffMultiplier)
	fmt.Printf("  Cache:\n")
	fmt.Printf("    Enabled: %v (default: true)\n", cfg.Processing.Cache.IsEnabled())
	fmt.Printf("    Path: %s (default: .cache/system-prompt.json)\n", cfg.Processing.Cache.Path)
}
