package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/config"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/prompt"
)

func main() {
	configPath := flag.String("config", "config.classify-gemma.json", "Path to configuration file")
	token := flag.String("token", "", "API token (overrides config)")
	referencesPath := flag.String("references", "_context", "Directory containing reference PDFs")
	noCache := flag.Bool("no-cache", false, "Disable cache usage")
	timeout := flag.Duration("timeout", 30*time.Minute, "Operation timeout")

	flag.Parse()

	cfg, err := config.LoadClassifyConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if *token != "" {
		cfg.Agent.Transport.Provider.Options["token"] = *token
	}

	if *noCache {
		enabled := false
		cfg.Processing.Cache.Enabled = &enabled
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	result, err := prompt.Generate(ctx, *cfg, *referencesPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating prompt: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("---")
	fmt.Println(result)
	fmt.Println("---")
}
