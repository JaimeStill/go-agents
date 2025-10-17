package processing_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/config"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/document"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/processing"
)

// mockPage implements document.Page for testing
type mockPage struct {
	number int
}

func (m *mockPage) Number() int {
	return m.number
}

func (m *mockPage) ToImage(opts document.ImageOptions) ([]byte, error) {
	return []byte(fmt.Sprintf("image-%d", m.number)), nil
}

func createMockPages(count int) []document.Page {
	pages := make([]document.Page, count)
	for i := 0; i < count; i++ {
		pages[i] = &mockPage{number: i + 1}
	}
	return pages
}

func TestProcessPages_Success(t *testing.T) {
	cfg := config.DefaultParallelConfig()
	pages := createMockPages(10)
	ctx := context.Background()

	results, err := processing.ProcessPages(ctx, cfg, pages,
		func(ctx context.Context, page document.Page) (int, error) {
			// Simulate processing time
			time.Sleep(10 * time.Millisecond)
			return page.Number() * 2, nil
		},
		nil,
	)

	if err != nil {
		t.Fatalf("ProcessPages failed: %v", err)
	}

	if len(results) != 10 {
		t.Errorf("expected 10 results, got %d", len(results))
	}

	// Verify results are in correct order
	for i, result := range results {
		expected := (i + 1) * 2
		if result != expected {
			t.Errorf("result[%d]: expected %d, got %d", i, expected, result)
		}
	}
}

func TestProcessPages_EmptyPages(t *testing.T) {
	cfg := config.DefaultParallelConfig()

	results, err := processing.ProcessPages(
		context.Background(),
		cfg,
		[]document.Page{},
		func(ctx context.Context, page document.Page) (int, error) {
			return 0, nil
		},
		nil,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}

func TestProcessPages_Error(t *testing.T) {
	cfg := config.DefaultParallelConfig()
	pages := createMockPages(10)
	ctx := context.Background()

	results, err := processing.ProcessPages(ctx, cfg, pages,
		func(ctx context.Context, page document.Page) (int, error) {
			if page.Number() == 5 {
				return 0, errors.New("processing error on page 5")
			}
			time.Sleep(10 * time.Millisecond)
			return page.Number(), nil
		},
		nil,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if results != nil {
		t.Error("expected nil results on error")
	}
}

func TestProcessPages_ContextCancellation(t *testing.T) {
	cfg := config.DefaultParallelConfig()
	pages := createMockPages(100)

	// Create an already-cancelled context for deterministic test
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately before processing

	results, err := processing.ProcessPages(ctx, cfg, pages,
		func(ctx context.Context, page document.Page) (int, error) {
			// Check context and return error if cancelled
			if err := ctx.Err(); err != nil {
				return 0, err
			}
			return page.Number(), nil
		},
		nil,
	)

	if err == nil {
		t.Fatal("expected error from context cancellation")
	}

	if results != nil {
		t.Error("expected nil results on cancellation")
	}
}

func TestProcessPages_Progress(t *testing.T) {
	cfg := config.DefaultParallelConfig()
	pages := createMockPages(5)
	ctx := context.Background()

	progressCalls := 0
	var lastCompleted, lastTotal int

	_, err := processing.ProcessPages(ctx, cfg, pages,
		func(ctx context.Context, page document.Page) (int, error) {
			time.Sleep(10 * time.Millisecond)
			return page.Number(), nil
		},
		func(completed, total int, result int) {
			progressCalls++
			lastCompleted = completed
			lastTotal = total
		},
	)

	if err != nil {
		t.Fatalf("ProcessPages failed: %v", err)
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

func TestProcessPages_Concurrency(t *testing.T) {
	cfg := config.ParallelConfig{WorkerCap: 4}
	pages := createMockPages(20)
	ctx := context.Background()

	start := time.Now()

	_, err := processing.ProcessPages(ctx, cfg, pages,
		func(ctx context.Context, page document.Page) (int, error) {
			time.Sleep(50 * time.Millisecond)
			return page.Number(), nil
		},
		nil,
	)

	duration := time.Since(start)

	if err != nil {
		t.Fatalf("ProcessPages failed: %v", err)
	}

	// With 4 workers and 20 pages, should take ~250ms (5 batches * 50ms)
	// Sequential would take 1000ms (20 * 50ms)
	// Allow some overhead
	if duration > 600*time.Millisecond {
		t.Errorf("parallel processing too slow: %v (expected < 600ms)", duration)
	}

	if duration < 200*time.Millisecond {
		t.Errorf("parallel processing too fast: %v (expected > 200ms)", duration)
	}
}
