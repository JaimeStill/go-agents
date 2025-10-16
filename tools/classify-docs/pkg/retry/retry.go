package retry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/config"
)

type Retryable[T any] func(ctx context.Context, attempt int) (T, error)

func Do[T any](
	ctx context.Context,
	cfg config.RetryConfig,
	fn Retryable[T],
) (T, error) {
	var zero T
	var lastErr error

	backoff := cfg.InitialBackoff.ToDuration()

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return zero, fmt.Errorf("retry cancelled: %w", ctx.Err())
		default:
		}

		result, err := fn(ctx, attempt)
		if err == nil {
			return result, nil
		}

		if !IsRetryable(err) {
			return zero, fmt.Errorf("non-retryable error on attempt %d: %w", attempt, err)
		}

		lastErr = err

		if attempt < cfg.MaxAttempts {
			select {
			case <-time.After(backoff):
				backoff = time.Duration(float64(backoff) * cfg.BackoffMultiplier)
				maxBackoff := cfg.MaxBackoff.ToDuration()
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			case <-ctx.Done():
				return zero, fmt.Errorf("retry cancelled during backoff: %w", ctx.Err())
			}
		}
	}

	return zero, fmt.Errorf("all %d attempts failed, last error: %w", cfg.MaxAttempts, lastErr)
}

func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	var nonRetryable *NonRetryableError
	if errors.As(err, &nonRetryable) {
		return false
	}

	return true
}

type NonRetryableError struct {
	Err error
}

func (e *NonRetryableError) Error() string {
	return fmt.Sprintf("non-retryable: %v", e.Err)
}

func (e *NonRetryableError) Unwrap() error {
	return e.Err
}

func MarkNonRetryable(err error) error {
	if err == nil {
		return nil
	}
	return &NonRetryableError{Err: err}
}
