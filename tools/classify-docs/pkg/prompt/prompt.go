package prompt

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/JaimeStill/go-agents/pkg/agent"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/cache"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/config"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/document"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/encoding"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/processing"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/retry"
)

func Generate(ctx context.Context, cfg config.ClassifyConfig, referencesPath string) (string, error) {
	if err := validateInputs(referencesPath); err != nil {
		return "", err
	}

	pdfs, err := discoverPDFs(referencesPath)
	if err != nil {
		return "", fmt.Errorf("failed to discover PDFs: %w", err)
	}

	if len(pdfs) == 0 {
		return "", fmt.Errorf("no PDF files found in %s", referencesPath)
	}

	if cfg.Processing.Cache.IsEnabled() {
		if cached, ok := checkCache(cfg.Processing.Cache.Path, pdfs); ok {
			fmt.Fprintf(os.Stderr, "Using cached system prompt\n")
			return cached, nil
		}
	}

	fmt.Fprintf(os.Stderr, "Discovering reference documents...\n")
	for _, pdf := range pdfs {
		doc, err := document.OpenPDF(pdf)
		if err != nil {
			return "", fmt.Errorf("failed to open %s: %w", filepath.Base(pdf), err)
		}
		pageCount := doc.PageCount()
		doc.Close()
		fmt.Fprintf(os.Stderr, "  Found: %s (%d pages)\n", filepath.Base(pdf), pageCount)
	}

	pages, err := extractAllPages(pdfs)
	if err != nil {
		return "", fmt.Errorf("failed to extract pages: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\nGenerating system prompt from %d pages...\n", len(pages))

	cfg.Agent.SystemPrompt = agentSystemPrompt()
	a, err := agent.New(&cfg.Agent)
	if err != nil {
		return "", fmt.Errorf("failed to create agent: %w", err)
	}

	processor := createProcessor(a, cfg.Processing.Retry)
	prompt := initialTemplate()

	progressFunc := func(completed, total int, context string) {
		fmt.Fprintf(os.Stderr, "  Page %d/%d processed\n", completed, total)
	}

	result, err := processing.ProcessWithContext(
		ctx,
		cfg.Processing.Sequential,
		pages,
		prompt,
		processor,
		progressFunc,
	)

	if err != nil {
		return "", fmt.Errorf("failed to process pages: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\nSystem prompt generated successfully!\n")

	if cfg.Processing.Cache.IsEnabled() {
		if err := saveCache(cfg.Processing.Cache.Path, pdfs, result.Final); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save cache: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "  Cached: %s\n", cfg.Processing.Cache.Path)
		}
	}

	return result.Final, nil
}

func validateInputs(referencesPath string) error {
	info, err := os.Stat(referencesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("references directory does not exist: %s", referencesPath)
		}
		return fmt.Errorf("failed to access references directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("references path is not a directory: %s", referencesPath)
	}

	return nil
}

func discoverPDFs(path string) ([]string, error) {
	pattern := filepath.Join(path, "*.pdf")
	pdfs, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob PDFs: %w", err)
	}

	return pdfs, nil
}

func extractAllPages(pdfPaths []string) ([]document.Page, error) {
	var allPages []document.Page

	for _, path := range pdfPaths {
		doc, err := document.OpenPDF(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open %s: %w", filepath.Base(path), err)
		}

		pages, err := doc.ExtractAllPages()
		doc.Close()

		if err != nil {
			return nil, fmt.Errorf("failed to extract pages from %s: %w", filepath.Base(path), err)
		}

		allPages = append(allPages, pages...)
	}

	return allPages, nil
}

func initialTemplate() string {
	return `You are a document classification expert specializing in DoD security classification markings.

## Classification Levels
[To be populated]

## Classification Formats and Syntax
[To be populated: banner, portion, header/footer markings]

## Caveats and Control Markings
[To be populated: NOFORN, REL TO, etc.]

## Derivation Rules
[To be populated: determining highest classification]

## Special Handling and Edge Cases
[To be populated: multi-classification scenarios]`
}

func agentSystemPrompt() string {
	return `You are a document classification expert focused on IDENTIFYING classification markings in documents.

Your task: Build a system prompt that teaches how to IDENTIFY and INTERPRET classification markings visually present in documents.

EXTRACT ONLY information about:
1. Classification levels and their visual markers (TS, S, C, U, CUI)
2. Marking formats (banner markings, portion markings, headers/footers)
3. Syntax and separators (//, /, -, spaces, commas)
4. Caveats and dissemination controls (NOFORN, ORCON, REL TO, ACCM)
5. How to determine highest classification from multiple markings
6. Visual patterns and placement of markings

IGNORE information about:
- Procedural guidance (how to classify, derivative classification processes)
- Responsibilities and authorities (OCAs, derivative classifiers)
- Administrative actions (tentative classification, missing markings)
- Training requirements
- Acquisition processes
- Policy compliance procedures
- What to do when information is leaked/released

Focus: "What do classification markings LOOK like and what do they MEAN?"
Not: "Who can classify and what are the procedures?"

You will receive:
1. A page image from a classification policy document
2. The current system prompt (work in progress)

CRITICAL OUTPUT REQUIREMENTS:
- Return ONLY the raw system prompt text
- NO conversational preambles ("Here's the updated prompt...")
- NO follow-up questions ("Do you want me to refine...")
- NO markdown indicating this is a prompt ("**SYSTEM PROMPT:**")
- NO meta-commentary about what you're doing
- Just return the direct, clean prompt text that will be used

The output should start immediately with the prompt content (e.g., "You are a document classification expert...")
Do not wrap it, introduce it, or comment on it. Just the prompt itself.`
}

func createProcessor(a agent.Agent, retryCfg config.RetryConfig) processing.ContextProcessor[string] {
	return func(ctx context.Context, page document.Page, current string) (string, error) {
		if err := ctx.Err(); err != nil {
			return current, fmt.Errorf("context cancelled: %w", err)
		}

		data, err := page.ToImage(document.DefaultImageOptions())
		if err != nil {
			return current, fmt.Errorf("failed to render page %d: %w", page.Number(), err)
		}

		encoded, err := encoding.EncodeImageDataURI(data, document.PNG)
		if err != nil {
			return current, fmt.Errorf("failed to encode page %d: %w", page.Number(), err)
		}

		var promptBuilder strings.Builder
		promptBuilder.WriteString("Current system prompt:\n\n")
		promptBuilder.WriteString(current)
		promptBuilder.WriteString("\n\nAnalyze this page and update the system prompt accordingly.")
		prompt := promptBuilder.String()

		update, err := retry.Do(ctx, retryCfg, func(ctx context.Context, attempt int) (string, error) {
			if attempt > 1 {
				fmt.Fprintf(os.Stderr, "    Retry attempt %d for page %d...\n", attempt-1, page.Number())
			}

			response, err := a.Vision(ctx, prompt, []string{encoded})
			if err != nil {
				return "", err
			}

			if len(response.Choices) == 0 {
				return "", fmt.Errorf("empty response for page %d", page.Number())
			}

			content := response.Content()
			if strings.TrimSpace(content) == "" {
				return "", fmt.Errorf("received empty context update for page %d", page.Number())
			}

			return content, nil
		})

		if err != nil {
			return current, fmt.Errorf("vision request failed for page %d: %w", page.Number(), err)
		}

		return update, nil
	}
}

func checkCache(cachePath string, pdfs []string) (string, bool) {
	cached, err := cache.Load(cachePath)
	if err != nil {
		return "", false
	}

	if len(cached.ReferenceDocuments) != len(pdfs) {
		return "", false
	}

	if !slices.Equal(pdfs, cached.ReferenceDocuments) {
		return "", false
	}

	return cached.SystemPrompt, true
}

func saveCache(cachePath string, pdfs []string, prompt string) error {
	c := cache.New(pdfs, prompt)
	return c.Save(cachePath)
}
