package processing_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/config"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/document"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/processing"
)

func TestProcessWithContext_Success(t *testing.T) {
	cfg := config.DefaultSequentialConfig()
	pages := createMockPages(5)
	ctx := context.Background()

	result, err := processing.ProcessWithContext(
		ctx,
		cfg,
		pages,
		"initial",
		func(ctx context.Context, page document.Page, prevContext string) (string, error) {
			return fmt.Sprintf("%s->page%d", prevContext, page.Number()), nil
		},
		nil,
	)

	if err != nil {
		t.Fatalf("ProcessWithContext failed: %v", err)
	}

	expected := "initial->page1->page2->page3->page4->page5"
	if result.Final != expected {
		t.Errorf("expected final context %q, got %q", expected, result.Final)
	}

	if result.Intermediate != nil {
		t.Error("expected nil intermediate contexts when not enabled")
	}
}

func TestProcessWithContext_EmptyPages(t *testing.T) {
	cfg := config.DefaultSequentialConfig()
	ctx := context.Background()

	result, err := processing.ProcessWithContext(
		ctx,
		cfg,
		[]document.Page{},
		"initial",
		func(ctx context.Context, page document.Page, prevContext string) (string, error) {
			return prevContext, nil
		},
		nil,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Final != "initial" {
		t.Errorf("expected initial context preserved, got %q", result.Final)
	}
}

func TestProcessWithContext_Error(t *testing.T) {
	cfg := config.DefaultSequentialConfig()
	pages := createMockPages(5)
	ctx := context.Background()

	result, err := processing.ProcessWithContext(
		ctx,
		cfg,
		pages,
		"initial",
		func(ctx context.Context, page document.Page, prevContext string) (string, error) {
			if page.Number() == 3 {
				return "", errors.New("processing error on page 3")
			}
			return fmt.Sprintf("%s->page%d", prevContext, page.Number()), nil
		},
		nil,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Result should have zero value for context
	if result.Final != "" {
		t.Errorf("expected empty final context on error, got %q", result.Final)
	}
}

func TestProcessWithContext_IntermediateContexts(t *testing.T) {
	cfg := config.SequentialConfig{ExposeIntermediateContexts: true}
	pages := createMockPages(3)
	ctx := context.Background()

	result, err := processing.ProcessWithContext(
		ctx,
		cfg,
		pages,
		"initial",
		func(ctx context.Context, page document.Page, prevContext string) (string, error) {
			return fmt.Sprintf("%s->page%d", prevContext, page.Number()), nil
		},
		nil,
	)

	if err != nil {
		t.Fatalf("ProcessWithContext failed: %v", err)
	}

	if result.Intermediate == nil {
		t.Fatal("expected intermediate contexts, got nil")
	}

	// Should have initial + 3 pages = 4 contexts
	if len(result.Intermediate) != 4 {
		t.Errorf("expected 4 intermediate contexts, got %d", len(result.Intermediate))
	}

	expectedContexts := []string{
		"initial",
		"initial->page1",
		"initial->page1->page2",
		"initial->page1->page2->page3",
	}

	for i, expected := range expectedContexts {
		if result.Intermediate[i] != expected {
			t.Errorf("intermediate context[%d]: expected %q, got %q",
				i, expected, result.Intermediate[i])
		}
	}
}

func TestProcessWithContext_Progress(t *testing.T) {
	cfg := config.DefaultSequentialConfig()
	pages := createMockPages(5)
	ctx := context.Background()

	progressCalls := 0
	var lastCompleted, lastTotal int

	_, err := processing.ProcessWithContext(
		ctx,
		cfg,
		pages,
		"initial",
		func(ctx context.Context, page document.Page, prevContext string) (string, error) {
			return prevContext, nil
		},
		func(completed, total int, context string) {
			progressCalls++
			lastCompleted = completed
			lastTotal = total
		},
	)

	if err != nil {
		t.Fatalf("ProcessWithContext failed: %v", err)
	}

	if progressCalls != 5 {
		t.Errorf("expected 5 progress calls, got %d", progressCalls)
	}

	if lastCompleted != 5 {
		t.Errorf("expected final completed=5, got %d", lastCompleted)
	}

	if lastTotal != 5 {
		t.Errorf("expected total=5, got %d", lastTotal)
	}
}

func TestProcessWithContext_ContextCancellation(t *testing.T) {
	cfg := config.DefaultSequentialConfig()
	pages := createMockPages(10)
	ctx, cancel := context.WithCancel(context.Background())

	processedPages := 0

	_, err := processing.ProcessWithContext(
		ctx,
		cfg,
		pages,
		"initial",
		func(ctx context.Context, page document.Page, prevContext string) (string, error) {
			processedPages++
			if processedPages == 3 {
				cancel() // Cancel after processing 3 pages
			}
			return prevContext, nil
		},
		nil,
	)

	if err == nil {
		t.Fatal("expected error from context cancellation")
	}

	// Should have processed only 3 pages before cancellation
	if processedPages != 3 {
		t.Errorf("expected 3 processed pages before cancellation, got %d", processedPages)
	}
}
