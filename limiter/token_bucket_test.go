package limiter

import (
	"context"
	"testing"
	"time"
)

func TestTokenBucketImmediateAcquire(t *testing.T) {
	// Create a bucket with capacity=2, refill every 100ms.
	tb := NewTokenBucket(2, 100*time.Millisecond)
	defer tb.Stop()

	// Initially, the bucket has 2 tokens.
	if !tb.TryAcquire() {
		t.Error("expected to acquire a token immediately from a full bucket")
	}
	if !tb.TryAcquire() {
		t.Error("expected to acquire the second token immediately")
	}

	// Now the bucket should be empty.
	// Trying to acquire again should return false immediately.
	if tb.TryAcquire() {
		t.Error("expected no tokens available after depleting the bucket")
	}
}

func TestTokenBucketRefill(t *testing.T) {
	// Create a bucket with capacity=1, refills every 50ms.
	tb := NewTokenBucket(1, 50*time.Millisecond)
	defer tb.Stop()

	// Initially, we can acquire the one token.
	if !tb.TryAcquire() {
		t.Error("expected to acquire token immediately from a full bucket")
	}

	// Bucket is now empty. Let's wait for a bit more than one fill interval to let it refill.
	time.Sleep(100 * time.Millisecond) // More than one interval to ensure refill has occurred.

	if !tb.TryAcquire() {
		t.Error("expected to acquire a token after the bucket refilled")
	}
}

func TestTokenBucketAcquireBlocking(t *testing.T) {
	// Create a bucket with capacity=1, refills every 50ms.
	tb := NewTokenBucket(1, 50*time.Millisecond)
	defer tb.Stop()

	// Acquire the initial token immediately.
	if err := tb.Acquire(context.Background()); err != nil {
		t.Errorf("unexpected error acquiring initial token: %v", err)
	}

	// Now the bucket is empty. The next Acquire should block until a token is refilled.
	// We'll set a timeout context to avoid waiting indefinitely in case of failure.
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	start := time.Now()
	if err := tb.Acquire(ctx); err != nil {
		t.Errorf("expected to eventually acquire a token after waiting for a refill, got error: %v", err)
	}
	elapsed := time.Since(start)

	// Ensure we actually waited at least one refill interval (50ms) before acquiring.
	if elapsed < 50*time.Millisecond {
		t.Error("expected Acquire to block until token refill, but it returned too quickly")
	}
}

func TestTokenBucketContextCancellation(t *testing.T) {
	// Create a bucket with capacity=1, refills every 500ms (quite slow).
	tb := NewTokenBucket(1, 500*time.Millisecond)
	defer tb.Stop()

	// Acquire the initial token.
	if err := tb.Acquire(context.Background()); err != nil {
		t.Errorf("unexpected error acquiring initial token: %v", err)
	}

	// Now empty. Attempt to Acquire with a short timeout context.
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := tb.Acquire(ctx); err == nil {
		t.Error("expected error due to context timeout, got no error")
	}
}
