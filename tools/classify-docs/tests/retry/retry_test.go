package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	acfg "github.com/JaimeStill/go-agents/pkg/config"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/config"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/retry"
)

func TestDo_SuccessFirstAttempt(t *testing.T) {
	cfg := config.DefaultRetryConfig()
	ctx := context.Background()

	result, err := retry.Do(ctx, cfg, func(ctx context.Context, attempt int) (string, error) {
		return "success", nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != "success" {
		t.Errorf("expected 'success', got %s", result)
	}
}

func TestDo_SuccessAfterRetry(t *testing.T) {
	cfg := config.RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    acfg.Duration(10 * time.Millisecond),
		MaxBackoff:        acfg.Duration(100 * time.Millisecond),
		BackoffMultiplier: 2.0,
	}
	ctx := context.Background()

	attempts := 0
	result, err := retry.Do(ctx, cfg, func(ctx context.Context, attempt int) (string, error) {
		attempts++
		if attempts < 3 {
			return "", errors.New("temporary failure")
		}
		return "success", nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != "success" {
		t.Errorf("expected 'success', got %s", result)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestDo_AllAttemptsFail(t *testing.T) {
	cfg := config.RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    acfg.Duration(10 * time.Millisecond),
		MaxBackoff:        acfg.Duration(100 * time.Millisecond),
		BackoffMultiplier: 2.0,
	}
	ctx := context.Background()

	attempts := 0
	_, err := retry.Do(ctx, cfg, func(ctx context.Context, attempt int) (string, error) {
		attempts++
		return "", errors.New("persistent failure")
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestDo_NonRetryableError(t *testing.T) {
	cfg := config.DefaultRetryConfig()
	ctx := context.Background()

	attempts := 0
	_, err := retry.Do(ctx, cfg, func(ctx context.Context, attempt int) (string, error) {
		attempts++
		return "", retry.MarkNonRetryable(errors.New("validation error"))
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if attempts != 1 {
		t.Errorf("expected 1 attempt for non-retryable error, got %d", attempts)
	}
}

func TestDo_ContextCancellation(t *testing.T) {
	cfg := config.RetryConfig{
		MaxAttempts:       5,
		InitialBackoff:    acfg.Duration(100 * time.Millisecond),
		MaxBackoff:        acfg.Duration(1 * time.Second),
		BackoffMultiplier: 2.0,
	}

	ctx, cancel := context.WithCancel(context.Background())

	attempts := 0
	done := make(chan struct{})

	go func() {
		_, err := retry.Do(ctx, cfg, func(ctx context.Context, attempt int) (string, error) {
			attempts++
			return "", errors.New("error")
		})

		if err == nil {
			t.Error("expected error from context cancellation")
		}
		close(done)
	}()

	// Cancel after short delay
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Wait for completion
	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("retry did not respect context cancellation")
	}

	// Should have stopped early due to cancellation
	if attempts >= 5 {
		t.Errorf("expected fewer than 5 attempts due to cancellation, got %d", attempts)
	}
}

func TestDo_ExponentialBackoff(t *testing.T) {
	cfg := config.RetryConfig{
		MaxAttempts:       4,
		InitialBackoff:    acfg.Duration(10 * time.Millisecond),
		MaxBackoff:        acfg.Duration(100 * time.Millisecond),
		BackoffMultiplier: 2.0,
	}
	ctx := context.Background()

	start := time.Now()
	attempts := 0

	_, err := retry.Do(ctx, cfg, func(ctx context.Context, attempt int) (string, error) {
		attempts++
		return "", errors.New("error")
	})

	duration := time.Since(start)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Expected backoffs: 10ms, 20ms, 40ms = 70ms minimum
	// Allow some overhead for execution
	if duration < 70*time.Millisecond {
		t.Errorf("backoff too short: %v (expected at least 70ms)", duration)
	}

	if duration > 200*time.Millisecond {
		t.Errorf("backoff too long: %v (expected less than 200ms)", duration)
	}
}
