package processing

import (
	"context"
	"fmt"

	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/config"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/document"
)

type ContextProcessor[TContext any] func(
	ctx context.Context,
	page document.Page,
	initial TContext,
) (current TContext, err error)

type SequentialResult[TContext any] struct {
	Final        TContext
	Intermediate []TContext
}

func ProcessWithContext[TContext any](
	ctx context.Context,
	cfg config.SequentialConfig,
	pages []document.Page,
	initial TContext,
	processor ContextProcessor[TContext],
	progress ProgressFunc,
) (SequentialResult[TContext], error) {
	var result SequentialResult[TContext]

	if len(pages) == 0 {
		result.Final = initial
		return result, nil
	}

	var intermediate []TContext
	if cfg.ExposeIntermediateContexts {
		intermediate = make([]TContext, 0, len(pages)+1)
		intermediate = append(intermediate, initial)
	}

	current := initial

	for i, page := range pages {
		if err := ctx.Err(); err != nil {
			return result, fmt.Errorf("sequential processing cancelled: %w", err)
		}

		updated, err := processor(ctx, page, current)
		if err != nil {
			return result, fmt.Errorf("failed on page %d: %w", page.Number(), err)
		}

		current = updated

		if cfg.ExposeIntermediateContexts {
			intermediate = append(intermediate, current)
		}

		if progress != nil {
			progress(i+1, len(pages))
		}
	}

	result.Final = current
	if cfg.ExposeIntermediateContexts {
		result.Intermediate = intermediate
	}

	return result, nil
}
