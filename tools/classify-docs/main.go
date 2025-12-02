package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/JaimeStill/go-agents/pkg/agent"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/cache"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/classify"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/config"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/prompt"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "generate-prompt":
		runGeneratePrompt(os.Args[2:])
	case "classify":
		runClassify(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: classify-docs <command> [options]\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  generate-prompt\tGenerate system prompt from reference documents\n")
	fmt.Fprintf(os.Stderr, "  classify\tClassify documents in a directory\n\n")
	fmt.Fprintf(os.Stderr, "Run classify-docs <command> --help for command-specific options.\n")
}

func runGeneratePrompt(args []string) {
	fs := flag.NewFlagSet("generate-prompt", flag.ExitOnError)
	configPath := fs.String("config", "config.classify-gpt4o-key.json", "Path to configuration file")
	token := fs.String("token", "", "API token (overrides config)")
	referencesPath := fs.String("references", "_context", "Directory containing reference PDFs")
	noCache := fs.Bool("no-cache", false, "Disable cache usage")
	timeout := fs.Duration("timeout", 30*time.Minute, "Operation timeout")

	fs.Parse(args)

	cfg, err := config.LoadClassifyConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if *token != "" {
		if cfg.Agent.Provider.Options == nil {
			cfg.Agent.Provider.Options = make(map[string]any)
		}
		cfg.Agent.Provider.Options["token"] = *token
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

func runClassify(args []string) {
	fs := flag.NewFlagSet("classify", flag.ExitOnError)
	configPath := fs.String("config", "config.classify-o4-mini.json", "Path to configuration file")
	token := fs.String("token", "", "API token (overrides config)")
	inputDir := fs.String("input", "_context/marked-documents", "Directory containing PDF documents to classify")
	outputFile := fs.String("output", "classification-results.json", "Output JSON file path")
	systemPromptPath := fs.String("system-prompt", ".cache/system-prompt.json", "Path to cached system prompt")
	timeout := fs.Duration("timeout", 15*time.Minute, "Operation timeout")

	fs.Parse(args)

	if *inputDir == "" {
		fmt.Fprintf(os.Stderr, "Error: --input is required\n\n")
		fs.Usage()
		os.Exit(1)
	}

	cfg, err := config.LoadClassifyConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if *token != "" {
		if cfg.Agent.Provider.Options == nil {
			cfg.Agent.Provider.Options = make(map[string]any)
		}
		cfg.Agent.Provider.Options["token"] = *token
	}

	cached, err := cache.Load(*systemPromptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading system prompt from %s: %v\n", *systemPromptPath, err)
		fmt.Fprintf(os.Stderr, "Hint: Run 'classify-docs generate-prompt' first to create the system prompt.\n")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Loaded system prompt from: %s\n", *systemPromptPath)
	fmt.Fprintf(os.Stderr, "  Generated: %s\n", cached.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(os.Stderr, "  References: %d documents\n\n", len(cached.ReferenceDocuments))

	cfg.Agent.SystemPrompt = cached.SystemPrompt

	a, err := agent.New(&cfg.Agent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating agent: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	results, err := classify.Classify(ctx, *cfg, a, *inputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error classifying documents: %v\n", err)
		os.Exit(1)
	}

	if err := saveResults(*outputFile, results); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving results: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Classification complete!\n")
	fmt.Fprintf(os.Stderr, "  Results saved: %s\n\n", *outputFile)

	outputJSON(results)
}

func saveResults(outputFile string, results []classify.DocumentClassification) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write results file: %w", err)
	}

	return nil
}

func outputJSON(results []classify.DocumentClassification) {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output JSON: %v\n", err)
		return
	}

	fmt.Println("---")
	fmt.Println(string(data))
	fmt.Println("---")
}
