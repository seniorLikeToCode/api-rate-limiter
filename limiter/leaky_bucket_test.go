package limiter

import (
	"testing"
	"time"
)

// TestBasicFunctionality tests that the limiter allows requests up to its capacity
// and then rejects further requests until it leaks tokens.
func TestBasicFunctionality(t *testing.T) {
	// Create a bucket that leaks 1 token every 100 milliseconds.
	lb := NewLeakyBucket(3, 100*time.Millisecond, 1)
	defer lb.Stop()

	// Initially, the bucket should be empty.
	if lb.CurrentSize() != 0 {
		t.Fatalf("expected empty bucket, got %d tokens", lb.CurrentSize())
	}

	// Fill the bucket up to capacity.
	for i := 0; i < 3; i++ {
		if !lb.Allow() {
			t.Fatalf("expected to allow request %d, but got denied", i)
		}
	}

	// Now the bucket should be full.
	if lb.CurrentSize() != 3 {
		t.Fatalf("expected bucket size 3, got %d", lb.CurrentSize())
	}

	// One more request should fail as the bucket is full.
	if lb.Allow() {
		t.Fatalf("expected deny but got allow")
	}

	// Wait for a leak interval so that one token is removed.
	time.Sleep(150 * time.Millisecond) // a bit more than leak interval for reliability

	// Now one token should have leaked out.
	if lb.CurrentSize() != 2 {
		t.Fatalf("expected bucket size 2 after leaking, got %d", lb.CurrentSize())
	}

	// Try again, we should now be able to add one more token.
	if !lb.Allow() {
		t.Fatalf("expected to allow request after leaking, but got denied")
	}

	// Cleanup
	lb.Stop()
}

// TestStop tests that after calling Stop, no more requests are allowed.
func TestStop(t *testing.T) {
	lb := NewLeakyBucket(2, 50*time.Millisecond, 1)

	// Add some tokens
	lb.Allow()
	lb.Allow()

	// Stop the bucket
	lb.Stop()

	// After stopping, no new tokens should be allowed.
	if lb.Allow() {
		t.Fatalf("expected deny after stop, but got allow")
	}
}

// TestLeakCount tests that multiple tokens can leak per interval if configured.
func TestLeakCount(t *testing.T) {
	// Create a bucket that can hold 5 tokens and leaks 2 tokens every 100ms.
	lb := NewLeakyBucket(5, 100*time.Millisecond, 2)
	defer lb.Stop()

	for i := 0; i < 5; i++ {
		if !lb.Allow() {
			t.Fatalf("expected to allow request %d, got denied", i)
		}
	}

	// Bucket should now be full.
	if lb.CurrentSize() != 5 {
		t.Fatalf("expected 5 tokens, got %d", lb.CurrentSize())
	}

	// Wait enough time for at least one leak event
	time.Sleep(120 * time.Millisecond)

	// Expect 2 tokens to have leaked out
	size := lb.CurrentSize()
	if size != 3 {
		t.Fatalf("expected 3 tokens after leak, got %d", size)
	}
}
