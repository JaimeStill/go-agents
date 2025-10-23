package classify

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/JaimeStill/go-agents/pkg/agent"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/config"
)

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

	fmt.Fprintf(os.Stderr, "Processing %d documents sequentially...\n\n", len(pdfs))

	var results []DocumentClassification

	for i, pdfPath := range pdfs {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("classification cancelled: %w", ctx.Err())
		}

		fmt.Fprintf(os.Stderr, "Document %d/%d: %s\n", i+1, len(pdfs), filepath.Base(pdfPath))

		result, err := ClassifyDocument(ctx, cfg, a, pdfPath)
		if err != nil {
			return nil, fmt.Errorf("failed to classify %s: %w", filepath.Base(pdfPath), err)
		}

		outputDocumentResult(result)
		results = append(results, result)
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

func outputDocumentResult(result DocumentClassification) {
	jsonBytes, err := json.MarshalIndent(result, "  ", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Warning: failed to format result JSON: %v\n", err)
	}

	fmt.Fprintf(os.Stderr, "  Result: %s\n\n", string(jsonBytes))
}
