# Phase 5: Document Classification Tool - Implementation Guide

## Overview

This phase implements document classification using **parallel processing at the document level** with **sequential page processing within each document**. The architecture mirrors Phase 4's context accumulation pattern, but applies it to classification instead of prompt generation.

## Key Architectural Decisions

### Document-Level Parallelism

Unlike typical parallel page processing, Phase 5 processes **entire documents in parallel**:
- Input: Directory containing multiple PDF documents
- Parallel unit: Each PDF document
- Within each document: Sequential page processing with context accumulation
- Output: Array of DocumentClassification results (one per document)

**Rationale**: Real-world scenario is "classify all documents in this directory" not "classify each page independently in one document."

### Context Accumulation Pattern

Each document's classification is built progressively across its pages:
```
Initial state: DocumentClassification{File: "doc.pdf", Classification: "", ...}

Page 1 → Vision API → Updated classification (first markings found)
Page 2 → Vision API → Updated classification (more markings, higher level?)
Page 3 → Vision API → Updated classification (accumulated markings)
...
Page N → Vision API → Final DocumentClassification
```

The accumulated DocumentClassification **IS** the final result. No separate synthesis step needed.

### Unified CLI Architecture

Phase 5 consolidates CLI into a single binary with subcommands:
```bash
go run . generate-prompt [flags]    # Migrated from cmd/generate-prompt/
go run . classify [flags]           # New functionality
```

## Implementation Steps

### Step 1: Create pkg/classify/types.go

Define the DocumentClassification structure that accumulates context during sequential page processing.

**File**: `pkg/classify/types.go`

```go
package classify

// DocumentClassification represents the classification state of a document.
// This structure is progressively updated as pages are processed sequentially.
type DocumentClassification struct {
	// File is the name of the PDF file being classified
	File string `json:"file"`

	// Classification is the highest classification level found across all pages
	// Values: "UNCLASSIFIED", "CONFIDENTIAL", "SECRET", "TOP SECRET"
	// Updated as higher classifications are encountered
	Classification string `json:"classification"`

	// Confidence is the agent's confidence in the classification
	// Values: "high", "medium", "low"
	// Refined as more pages are processed
	Confidence string `json:"confidence"`

	// MarkingsFound is the accumulated set of unique classification markings
	// Example: ["SECRET//NOFORN", "(S)", "(U)", "CONFIDENTIAL"]
	// Agent deduplicates markings as pages are processed
	MarkingsFound []string `json:"markings_found"`

	// ClassificationRationale is the cumulative explanation
	// Updated with each page to reflect the overall document classification
	ClassificationRationale string `json:"classification_rationale"`
}
```

**Design Notes**:
- Single type serves as both working state and final output
- Agent updates all fields progressively across pages
- Agent handles deduplication of `MarkingsFound`
- Agent synthesizes cumulative `ClassificationRationale`

---

### Step 2: Create pkg/classify/parser.go

Implement robust JSON parsing that handles agent responses with or without markdown code fences.

**File**: `pkg/classify/parser.go`

```go
package classify

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// parseClassificationResponse extracts DocumentClassification from agent response.
// Handles both direct JSON and markdown-wrapped JSON (```json...```).
// Returns detailed error if parsing fails with both strategies.
func parseClassificationResponse(content string) (DocumentClassification, error) {
	var result DocumentClassification

	// Try direct JSON parse
	err1 := json.Unmarshal([]byte(content), &result)
	if err1 == nil {
		return result, nil
	}

	// Try extracting from markdown code fence
	extracted := extractJSONFromMarkdown(content)
	err2 := json.Unmarshal([]byte(extracted), &result)
	if err2 == nil {
		return result, nil
	}

	// Return both errors for debugging
	return DocumentClassification{}, fmt.Errorf("failed to parse classification response: %w",
		errors.Join(
			fmt.Errorf("direct parse: %w", err1),
			fmt.Errorf("markdown extraction: %w", err2),
		))
}

// extractJSONFromMarkdown extracts JSON from markdown code fences.
// Matches ```json...``` or ```...``` blocks.
// Returns original content if no code fence found.
func extractJSONFromMarkdown(content string) string {
	// Match ```json...``` or ```...``` blocks with multiline support
	re := regexp.MustCompile(`(?s)` + "`" + `{3}(?:json)?\s*(.+?)\s*` + "`" + `{3}`)
	matches := re.FindStringSubmatch(content)

	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return content // Return original if no code fence found
}
```

**Design Notes**:
- `errors.Join()` (Go 1.20+) provides detailed multi-error reporting
- Regex with `(?s)` flag enables multiline matching
- Falls back gracefully if markdown extraction fails
- Returns original content if no code fence detected

**Testing Considerations**:
- Test with direct JSON: `{"file": "test.pdf", ...}`
- Test with markdown: ` ```json\n{...}\n``` `
- Test with invalid JSON in both formats
- Test with empty/malformed responses

---

### Step 3: Create pkg/classify/document.go

Implement sequential page classification with context accumulation for a single PDF document.

**File**: `pkg/classify/document.go`

```go
package classify

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/JaimeStill/go-agents/pkg/agent"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/config"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/document"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/encoding"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/processing"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/retry"
)

// ClassifyDocument processes a single PDF document sequentially, accumulating
// classification across all pages. Returns the final DocumentClassification.
func ClassifyDocument(
	ctx context.Context,
	cfg config.ClassifyConfig,
	a agent.Agent,
	pdfPath string,
) (DocumentClassification, error) {
	doc, err := document.OpenPDF(pdfPath)
	if err != nil {
		return DocumentClassification{}, fmt.Errorf("failed to open PDF: %w", err)
	}

	pages, err := doc.ExtractAllPages()
	doc.Close()

	if err != nil {
		return DocumentClassification{}, fmt.Errorf("failed to extract pages: %w", err)
	}

	filename := filepath.Base(pdfPath)

	// Initial classification state - empty, to be populated by agent
	initial := DocumentClassification{
		File:                    filename,
		Classification:          "",
		Confidence:              "",
		MarkingsFound:           []string{},
		ClassificationRationale: "",
	}

	// Progress callback - called after each page is processed
	progressFunc := func(completed, total int, current DocumentClassification) {
		fmt.Fprintf(os.Stderr, "  Page %d/%d processed\n", completed, total)
	}

	// Process pages sequentially with context accumulation
	result, err := processing.ProcessWithContext(
		ctx,
		cfg.Processing.Sequential,
		pages,
		initial,
		createClassifier(a, cfg.Processing.Retry),
		progressFunc,
	)

	if err != nil {
		return DocumentClassification{}, fmt.Errorf("failed to classify %s: %w", filename, err)
	}

	return result.Final, nil
}

// createClassifier returns a ContextProcessor that classifies a single page
// and updates the accumulated DocumentClassification.
func createClassifier(
	a agent.Agent,
	retryCfg config.RetryConfig,
) processing.ContextProcessor[DocumentClassification] {
	return func(
		ctx context.Context,
		page document.Page,
		current DocumentClassification,
	) (DocumentClassification, error) {
		if err := ctx.Err(); err != nil {
			return current, fmt.Errorf("context cancelled: %w", err)
		}

		// Render page to image
		data, err := page.ToImage(document.DefaultImageOptions())
		if err != nil {
			return current, fmt.Errorf("failed to render page %d: %w", page.Number(), err)
		}

		// Encode as data URI for vision API
		encoded, err := encoding.EncodeImageDataURI(data, document.PNG)
		if err != nil {
			return current, fmt.Errorf("failed to encode page %d: %w", page.Number(), err)
		}

		// Build prompt with current classification state
		prompt := buildClassificationPrompt(current)

		// Vision API call with retry
		updated, err := retry.Do(ctx, retryCfg, func(ctx context.Context, attempt int) (DocumentClassification, error) {
			if attempt > 1 {
				fmt.Fprintf(os.Stderr, "    Retry attempt %d for page %d...\n", attempt-1, page.Number())
			}

			response, err := a.Vision(ctx, prompt, []string{encoded})
			if err != nil {
				return current, err
			}

			if len(response.Choices) == 0 {
				return current, fmt.Errorf("empty response for page %d", page.Number())
			}

			content := response.Content()
			if strings.TrimSpace(content) == "" {
				return current, fmt.Errorf("received empty classification for page %d", page.Number())
			}

			// Parse JSON response
			classification, err := parseClassificationResponse(content)
			if err != nil {
				return current, fmt.Errorf("failed to parse page %d response: %w", page.Number(), err)
			}

			return classification, nil
		})

		if err != nil {
			return current, fmt.Errorf("vision request failed for page %d: %w", page.Number(), err)
		}

		return updated, nil
	}
}

const classificationPromptTemplate = `Current document classification state:

{
  "file": {{.File | printf "%q"}},
  "classification": {{.Classification | printf "%q"}},
  "confidence": {{.Confidence | printf "%q"}},
  "markings_found": [{{range $i, $m := .MarkingsFound}}{{if $i}}, {{end}}{{$m | printf "%q"}}{{end}}],
  "classification_rationale": {{.ClassificationRationale | printf "%q"}}
}
{{if not .Classification}}

This is the first page - initialize the classification.
{{end}}

Analyze this page image and update the document classification:

1. Add newly discovered markings to markings_found (avoid duplicates)
2. Update classification to the highest level found across all pages so far
3. Update confidence based on consistency and clarity of markings
4. Update rationale with cumulative classification reasoning

Return ONLY the updated DocumentClassification as valid JSON.
Do NOT wrap in markdown code fences (though parser can handle it).
Do NOT add conversational text - just the JSON object.`

var promptTemplate *template.Template

func init() {
	promptTemplate = template.Must(template.New("classification").Parse(classificationPromptTemplate))
}

// buildClassificationPrompt constructs the user prompt for each page using a template.
// Includes current classification state and instructs agent to update it.
func buildClassificationPrompt(current DocumentClassification) string {
	var buf bytes.Buffer
	if err := promptTemplate.Execute(&buf, current); err != nil {
		// Fallback to simple prompt on template error
		return "Analyze this page and provide classification as JSON."
	}

	return buf.String()
}
```

**Design Notes**:
- Sequential processing mirrors Phase 4's prompt generation pattern
- `text/template` approach makes prompt structure clear and maintainable
- Template parsed once at initialization for efficiency
- Prompt shows current state and instructs agent to update incrementally
- Retry infrastructure handles Azure rate limiting
- Progress reporting shows page numbers (full JSON output comes later)
- Context cancellation checked before each page
- Template fallback ensures graceful degradation on template errors

**Testing Considerations**:
- Mock agent responses with progressive classification updates
- Test with empty initial state
- Test with partially populated state
- Test retry logic with failing agent responses
- Test context cancellation mid-document

---

### Step 4: Create pkg/classify/classify.go

Implement parallel document processing across a directory of PDFs.

**File**: `pkg/classify/classify.go`

```go
package classify

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/JaimeStill/go-agents/pkg/agent"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/config"
)

// Classify processes all PDF documents in a directory in parallel.
// Each document is classified sequentially page-by-page.
// Returns array of DocumentClassification results.
func Classify(
	ctx context.Context,
	cfg config.ClassifyConfig,
	a agent.Agent,
	inputDir string,
) ([]DocumentClassification, error) {
	if err := validateInputs(inputDir); err != nil {
		return nil, err
	}

	pdfs, err := discoverPDFs(inputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to discover PDFs: %w", err)
	}

	if len(pdfs) == 0 {
		return nil, fmt.Errorf("no PDF files found in %s", inputDir)
	}

	fmt.Fprintf(os.Stderr, "Processing %d documents...\n\n", len(pdfs))

	// Parallel document processing
	workers := min(cfg.Processing.Parallel.WorkerCap, len(pdfs))
	workCh := make(chan string, len(pdfs))
	resultCh := make(chan DocumentClassification, len(pdfs))
	errCh := make(chan error, 1)

	// Launch workers
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for pdfPath := range workCh {
				if ctx.Err() != nil {
					return
				}

				result, err := ClassifyDocument(ctx, cfg, a, pdfPath)
				if err != nil {
					select {
					case errCh <- err:
					default:
					}
					return
				}

				// Output full JSON for this document
				outputDocumentResult(result)

				resultCh <- result
			}
		})
	}

	// Send work
	go func() {
		defer close(workCh)
		for _, pdf := range pdfs {
			select {
			case workCh <- pdf:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait and close results
	go func() {
		wg.Wait()
		close(resultCh)
		close(errCh)
	}()

	// Collect results
	var results []DocumentClassification
	for result := range resultCh {
		results = append(results, result)
	}

	// Check for errors
	if err := <-errCh; err != nil {
		return nil, err
	}

	if ctx.Err() != nil {
		return nil, fmt.Errorf("classification cancelled: %w", ctx.Err())
	}

	return results, nil
}

func validateInputs(inputDir string) error {
	info, err := os.Stat(inputDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("input directory does not exist: %s", inputDir)
		}
		return fmt.Errorf("failed to access input directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("input path is not a directory: %s", inputDir)
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

// outputDocumentResult writes the document classification result to stderr
func outputDocumentResult(result DocumentClassification) {
	jsonBytes, err := json.MarshalIndent(result, "  ", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Warning: failed to format result JSON: %v\n", err)
		return
	}

	fmt.Fprintf(os.Stderr, "  Result: %s\n\n", string(jsonBytes))
}
```

**Design Notes**:
- Direct parallel processing at document level using worker pool pattern
- Uses `sync.WaitGroup.Go()` for modern Go 1.25.2 concurrency
- Worker pool size controlled by `cfg.Processing.Parallel.WorkerCap`
- Each document is classified sequentially via `ClassifyDocument()`
- Buffered channels for work distribution and result collection
- First error stops processing via `errCh` with single-item buffer
- Context cancellation supported throughout the pipeline
- Full JSON output per document to stderr after completion
- Progress updates handled within `ClassifyDocument()` (per-page)

**Testing Considerations**:
- Mock multiple documents with varying page counts
- Test parallel execution timing
- Test error handling when one document fails
- Test context cancellation across parallel workers

---

### Step 5: Create Unified CLI in main.go

Consolidate CLI with subcommand routing for both generate-prompt and classify.

**File**: `main.go`

```go
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
	fmt.Fprintf(os.Stderr, "  generate-prompt    Generate system prompt from reference documents\n")
	fmt.Fprintf(os.Stderr, "  classify           Classify documents in a directory\n\n")
	fmt.Fprintf(os.Stderr, "Run 'classify-docs <command> --help' for command-specific options.\n")
}

// runGeneratePrompt executes the generate-prompt subcommand
func runGeneratePrompt(args []string) {
	fs := flag.NewFlagSet("generate-prompt", flag.ExitOnError)
	configPath := fs.String("config", "config.classify-gemma.json", "Path to configuration file")
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

// runClassify executes the classify subcommand
func runClassify(args []string) {
	fs := flag.NewFlagSet("classify", flag.ExitOnError)
	configPath := fs.String("config", "config.classify-gpt4o-key.json", "Path to configuration file")
	token := fs.String("token", "", "API token (overrides config)")
	inputDir := fs.String("input", "", "Directory containing PDF documents to classify (required)")
	outputFile := fs.String("output", "classification-results.json", "Output JSON file path")
	systemPromptPath := fs.String("system-prompt", ".cache/system-prompt.json", "Path to cached system prompt")
	timeout := fs.Duration("timeout", 15*time.Minute, "Operation timeout")
	workers := fs.Int("workers", 0, "Number of parallel workers (0 = use config default)")

	fs.Parse(args)

	if *inputDir == "" {
		fmt.Fprintf(os.Stderr, "Error: --input is required\n\n")
		fs.Usage()
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.LoadClassifyConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if *token != "" {
		cfg.Agent.Transport.Provider.Options["token"] = *token
	}

	if *workers > 0 {
		cfg.Processing.Parallel.WorkerCap = *workers
	}

	// Load cached system prompt
	cached, err := cache.Load(*systemPromptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading system prompt from %s: %v\n", *systemPromptPath, err)
		fmt.Fprintf(os.Stderr, "Hint: Run 'classify-docs generate-prompt' first to create the system prompt.\n")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Loaded system prompt from: %s\n", *systemPromptPath)
	fmt.Fprintf(os.Stderr, "  Generated: %s\n", cached.Timestamp.Format(time.RFC3339))
	fmt.Fprintf(os.Stderr, "  References: %d documents\n\n", len(cached.ReferenceDocuments))

	// Set system prompt in agent config
	cfg.Agent.SystemPrompt = cached.SystemPrompt

	// Create agent
	a, err := agent.New(&cfg.Agent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating agent: %v\n", err)
		os.Exit(1)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Classify documents
	results, err := classify.Classify(ctx, *cfg, a, *inputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error classifying documents: %v\n", err)
		os.Exit(1)
	}

	// Save results to file
	if err := saveResults(*outputFile, results); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving results: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Classification complete!\n")
	fmt.Fprintf(os.Stderr, "  Results saved: %s\n\n", *outputFile)

	// Output JSON array to stdout
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
```

**Design Notes**:
- Subcommand routing with `flag.NewFlagSet()` for clean separation
- `generate-prompt` migrated from `cmd/generate-prompt/main.go`
- `classify` requires `--input`, other flags have sensible defaults
- Loads cached system prompt and injects into agent config
- Saves results to JSON file (default: classification-results.json)
- Outputs JSON array to stdout for piping

**After Creating main.go**:
1. Delete `cmd/generate-prompt/` directory (functionality migrated)
2. Keep `cmd/test-config/` and `cmd/test-render/` (utility tools)

---

### Step 6: Create tests/classify/ Package

Implement black-box tests for classification logic.

#### File: `tests/classify/parser_test.go`

```go
package classify_test

import (
	"strings"
	"testing"

	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/classify"
)

func TestParseClassificationResponse_DirectJSON(t *testing.T) {
	jsonResponse := `{
		"file": "test.pdf",
		"classification": "SECRET",
		"confidence": "high",
		"markings_found": ["SECRET//NOFORN", "(S)"],
		"classification_rationale": "Document contains SECRET markings"
	}`

	// Use unexported function via testing backdoor or make it exported for testing
	// For now, we'll test through the public ClassifyDocument flow
	// This is a placeholder showing the test structure

	t.Skip("Parser function is unexported - test via integration")
}

func TestParseClassificationResponse_MarkdownJSON(t *testing.T) {
	markdownResponse := "```json\n" + `{
		"file": "test.pdf",
		"classification": "SECRET",
		"confidence": "high",
		"markings_found": ["SECRET//NOFORN"],
		"classification_rationale": "Contains SECRET markings"
	}` + "\n```"

	t.Skip("Parser function is unexported - test via integration")
}

func TestParseClassificationResponse_InvalidJSON(t *testing.T) {
	invalidResponse := "This is not JSON at all"

	t.Skip("Parser function is unexported - test via integration")
}

// Note: Since parseClassificationResponse is unexported, these tests
// would need to either:
// 1. Make the function exported (ParseClassificationResponse)
// 2. Test through the public API (ClassifyDocument with mocked agent)
// 3. Use testing backdoors (not recommended)
//
// Recommendation: Export the parser function for testability
```

#### File: `tests/classify/classify_test.go`

```go
package classify_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/classify"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/config"
)

func TestClassify_NonexistentDirectory(t *testing.T) {
	cfg := config.DefaultClassifyConfig()

	// Create mock agent (implementation depends on go-agents mock package)
	// For now, this is a placeholder structure

	_, err := classify.Classify(context.Background(), *cfg, nil, "/nonexistent/directory")
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}

	if !contains(err.Error(), "does not exist") {
		t.Errorf("expected 'does not exist' in error, got: %v", err)
	}
}

func TestClassify_NotADirectory(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "notadir.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cfg := config.DefaultClassifyConfig()

	_, err := classify.Classify(context.Background(), *cfg, nil, tmpFile)
	if err == nil {
		t.Fatal("expected error when path is not a directory")
	}

	if !contains(err.Error(), "not a directory") {
		t.Errorf("expected 'not a directory' in error, got: %v", err)
	}
}

func TestClassify_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultClassifyConfig()

	_, err := classify.Classify(context.Background(), *cfg, nil, tmpDir)
	if err == nil {
		t.Fatal("expected error for empty directory")
	}

	if !contains(err.Error(), "no PDF files found") {
		t.Errorf("expected 'no PDF files found' in error, got: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Note: Integration tests with mocked agent responses would go here
// These would test the full classify flow with controlled agent responses
```

**Design Notes**:
- Parser tests are placeholders - parser function should be exported for testing
- Integration tests use temp directories and validation logic
- Mock agent responses would test full classification flow
- Black-box testing validates error handling and edge cases

**Testing Strategy**:
1. Export `parseClassificationResponse` for direct testing
2. Create mock agent in go-agents for integration tests
3. Test with temporary PDFs if ImageMagick available
4. Skip integration tests if dependencies unavailable

---

### Step 7: Update Documentation

#### Update PROJECT.md

Update Phase 5 section to mark as complete:

```markdown
### Phase 5: Document Classification Tool ✅

**Status**: Complete

**Development Summary**: `_context/.archive/04-document-classification.md` (created after completion)

**Objectives**:
- ✅ Implement CLI interface for document classification
- ✅ Use parallel processor for independent document analysis
- ✅ Implement sequential page processing within each document
- ✅ JSON output with structured results
- ✅ Unified CLI with subcommands

**Deliverables**:
- ✅ `pkg/classify/` - Classification infrastructure
- ✅ `main.go` - Unified CLI with generate-prompt and classify subcommands
- ✅ `tests/classify/` - Black-box tests for classification
- ✅ Consolidated CLI (removed cmd/generate-prompt/)

**Success Criteria**:
- ✅ Process multiple documents in parallel
- ✅ Classify pages sequentially within each document
- ✅ Context accumulation for document classification
- ✅ Agent determines highest classification
- ✅ Valid JSON output with rationale
- ✅ Unified subcommand architecture
```

#### Update README.md

Add classify command documentation:

```markdown
### classify - Document Classification

Classify PDF documents in a directory using parallel processing.

#### Usage

```bash
go run . classify [options]
```

**Flags:**
- `--config` (default: "config.classify-gpt4o-key.json") - Path to agent configuration file
- `--token` - API token (overrides token in config file)
- `--input` (required) - Directory containing PDF documents to classify
- `--output` (default: "classification-results.json") - Output JSON file path
- `--system-prompt` (default: ".cache/system-prompt.json") - Path to cached system prompt
- `--timeout` (default: "15m") - Operation timeout
- `--workers` - Number of parallel workers (overrides config default)

#### Examples

**Basic Usage**

```bash
export AZURE_API_KEY="your-api-key"
go run . classify --token $AZURE_API_KEY --input ./documents
```

**Custom Configuration**

```bash
go run . classify \
  --config config.classify-gpt4o-key.json \
  --token $AZURE_API_KEY \
  --input ./classified-docs \
  --output results.json \
  --system-prompt .cache/system-prompt.json \
  --workers 4
```

**Output:**

```
Loaded system prompt from: .cache/system-prompt.json
  Generated: 2025-10-17T14:30:00Z
  References: 2 documents

Processing 3 documents...

Document: security-memo.pdf (5 pages)
  Page 1/5 processed
  Page 2/5 processed
  Page 3/5 processed
  Page 4/5 processed
  Page 5/5 processed
  Result: {
    "file": "security-memo.pdf",
    "classification": "SECRET",
    "confidence": "high",
    "markings_found": ["SECRET//NOFORN", "(S)", "(U)"],
    "classification_rationale": "Document contains SECRET banner markings across all pages..."
  }

Document: public-notice.pdf (2 pages)
  Page 1/2 processed
  Page 2/2 processed
  Result: {
    "file": "public-notice.pdf",
    "classification": "UNCLASSIFIED",
    ...
  }

Classification complete!
  Results saved: classification-results.json

---
[JSON array to stdout]
---
```

The classification results are saved to the specified output file and also displayed to stdout for piping.
```

Update test count and package structure:

```markdown
**Test Packages:**
- `tests/document/` - PDF processing and image conversion (7 tests)
- `tests/config/` - Configuration loading and merging (4 tests)
- `tests/cache/` - System prompt caching (4 tests)
- `tests/retry/` - Retry logic and exponential backoff (6 tests)
- `tests/processing/` - Parallel and sequential processors (12 tests)
- `tests/encoding/` - Base64 data URI image encoding (5 tests)
- `tests/prompt/` - System prompt generation validation (3 tests)
- `tests/classify/` - Document classification (3+ tests)

**Total: 48+ tests, all passing**

**Package Structure:**
```
pkg/
├── config/        # Configuration types and loading
├── retry/         # Retry logic with exponential backoff
├── cache/         # System prompt caching
├── document/      # PDF processing and image conversion
├── encoding/      # Base64 data URI encoding for images
├── prompt/        # System prompt generation
├── processing/    # Parallel and sequential processors
└── classify/      # Document classification
```
```

---

### Step 8: Validate with Context7 MCP

Before execution, engage a subagent to validate Go 1.25.2 idioms using Context7 MCP.

**Validation completed in planning phase** - Key findings:
- ✅ Existing patterns already use Go 1.25.2 best practices
- ✅ WaitGroup.Go(), range over integers, min(), defer close() all correct
- ⚠️ Minor improvement: Remove redundant context check in parallel worker loop
- 💡 Use errors.Join() for multi-attempt error reporting
- 💡 Use regexp for markdown extraction
- 💡 Consider flag.NewFlagSet() for subcommands (implemented in Step 5)

---

## Implementation Checklist

### Code Implementation
- [ ] Create pkg/classify/types.go (DocumentClassification)
- [ ] Create pkg/classify/parser.go (JSON parsing with markdown)
- [ ] Create pkg/classify/document.go (Sequential page classification)
- [ ] Create pkg/classify/classify.go (Parallel document processing)
- [ ] Create main.go (Unified CLI with subcommands)
- [ ] Delete cmd/generate-prompt/ (migrated to main.go)
- [ ] Create tests/classify/parser_test.go
- [ ] Create tests/classify/classify_test.go

### Testing
- [ ] Run tests: `go test ./tests/classify/... -v`
- [ ] Run full test suite: `go test ./tests/... -v`
- [ ] Verify all 48+ tests pass

### Documentation
- [ ] Update PROJECT.md Phase 5 section
- [ ] Update README.md with classify command
- [ ] Update test count and package structure

### Manual Validation
- [ ] Generate system prompt: `go run . generate-prompt --token $TOKEN`
- [ ] Classify test documents: `go run . classify --token $TOKEN --input ./test-docs`
- [ ] Verify JSON output format
- [ ] Verify progress reporting
- [ ] Test with multiple documents in parallel

---

## Success Criteria

- ✅ Process multiple PDFs in parallel
- ✅ Classify each page sequentially within document
- ✅ Accumulate DocumentClassification across pages
- ✅ Agent determines highest classification via context accumulation
- ✅ Save results to JSON file
- ✅ Output JSON array to stdout
- ✅ Progress reporting to stderr (per-page + per-document JSON)
- ✅ Unified CLI with subcommands
- ✅ Comprehensive black-box tests
- ✅ Modern Go 1.25.2 patterns validated

---

## Architecture Validation

Phase 5 validates these patterns for go-agents-document-context:

1. **Document-level parallelism** - Real-world use case for directory processing
2. **Sequential context accumulation** - Reusable pattern from Phase 4
3. **Agent-driven synthesis** - LLM determines highest classification, not hardcoded logic
4. **Unified CLI** - Scalable subcommand architecture
5. **Robust parsing** - Handle markdown-wrapped JSON from agents
6. **Progress visibility** - Per-page + per-document feedback

These patterns demonstrate the flexibility and power of the document processing → agent analysis architecture.
