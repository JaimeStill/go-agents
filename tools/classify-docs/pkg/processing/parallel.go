package processing

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/config"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/document"
)

func ProcessPages[T any](
	ctx context.Context,
	cfg config.ParallelConfig,
	pages []document.Page,
	processor func(context.Context, document.Page) (T, error),
	progress ProgressFunc[T],
) ([]T, error) {
	if len(pages) == 0 {
		return []T{}, nil
	}

	workers := min(min(runtime.NumCPU()*2, cfg.WorkerCap), len(pages))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make([]T, len(pages))

	type workItem struct {
		page  document.Page
		index int
	}

	type result struct {
		index int
		value T
		err   error
	}

	workCh := make(chan workItem, workers)
	resultCh := make(chan result, len(pages))

	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for work := range workCh {
				select {
				case <-ctx.Done():
					return
				default:
				}

				value, err := processor(ctx, work.page)

				resultCh <- result{
					index: work.index,
					value: value,
					err:   err,
				}

				if err != nil {
					cancel()
					return
				}
			}
		})
	}

	go func() {
		defer close(workCh)
		for i, page := range pages {
			select {
			case workCh <- workItem{page: page, index: i}:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Collect results in background
	var firstErr error
	completed := 0
	done := make(chan struct{})

	go func() {
		defer close(done)
		for res := range resultCh {
			if res.err != nil {
				if firstErr == nil {
					firstErr = res.err
					cancel()
				}
			} else {
				results[res.index] = res.value
			}

			completed++
			if progress != nil {
				progress(completed, len(pages), res.value)
			}
		}
	}()

	// Wait for workers to finish
	wg.Wait()
	close(resultCh)

	// Wait for result collection to complete
	<-done

	if firstErr != nil {
		return nil, fmt.Errorf("parallel processing failed: %w", firstErr)
	}

	// Check if context was cancelled without processing all pages
	if completed < len(pages) {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("parallel processing cancelled: %w", err)
		}
	}

	return results, nil
}
