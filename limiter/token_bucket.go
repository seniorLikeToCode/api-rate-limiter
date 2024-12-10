package limiter

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Error constant
var ErrContextTimeout = errors.New("context canceled or timeout before acquiring token")
var ErrQueueEmpty = errors.New("queue is empty")

// RateLimiter defines the behavior of a rate limiter.
// We have two main methods:
// - TryAcquire: Imediately return whether a token could be acquired or not.
// - Acquire: Blocks (or waits) until a token is available or context is canceled.
type RateLimiter interface {
	TryAcquire() bool
	Acquire(ctx context.Context) error
}

// TokenBucket is an implementation of a token bucket rate limiter.
// key points of the token bucket approach:
// 1. We have a "bucket" that holds a certain number of tokens (capacity).
// 2. Tokens are added to the bucket at a fixed interval (fillInterval).
// 3. Each request (or event) tries to remove one token from the bucket.
// 4. If a token is available, the request is allowed to proceed immediately.
// 5. If no token is available, the request must wait (blocking Acquire) or fail immediately (TryAcquire).
type TokenBucket struct {
	capacity     int           // Maximum number of tokens that the bucket can hold.
	tokens       int           // Current number of tokens available in the bucket.
	fillInterval time.Duration // Interval at which one token is added back into the bucket.
	ticker       *time.Ticker  // A ticker that triggers adding tokens at a regular interval.
	mu           sync.Mutex    // A mutex to ensure safe concurrent access to the bucket state.
	done         chan struct{} // A channel used to signal goroutine termination when stopping the limiter.
}

// NewTokenBucket creates and returns a new TokenBucket rate limiter.
// Parameters:
//   - capacity: The maximum number of tokens in the bucket.
//   - fillInterval: How often a token is added. For example, if fillInterval is 200ms
//     it means every 200mx one token is added to the bucket until the bucket is full.
func NewTokenBucket(capacity int, fillInterval time.Duration) *TokenBucket {
	tb := &TokenBucket{
		capacity:     capacity,
		tokens:       capacity, // start with a full bucket.
		fillInterval: fillInterval,
		ticker:       time.NewTicker(fillInterval),
		done:         make(chan struct{}),
	}

	// Start a separate goroutine to continously refill the bucket over time.
	go tb.refill()

	return tb
}

// refill is a goroutine that runs continously. On each ticker tick, it attempts to
// add one token to the bucket if it's not full. It stops when the "done" channel is closed.
func (tb *TokenBucket) refill() {
	for {
		select {
		case <-tb.ticker.C:
			tb.mu.Lock()
			if tb.tokens < tb.capacity {
				tb.tokens++
			}
			tb.mu.Unlock()
		case <-tb.done:
			// stop refilling when done is closed.
			return
		}

	}
}

// Stop cleanly stops the token bucket from refilling. It should be called when you
// no longer need the limiter, to prevent goroutines and tickers from leaking.
func (tb *TokenBucket) Stop() {
	close(tb.done)
	tb.ticker.Stop()
}

// TryAcquire attempts to take one token from the bucket immediately, without waiting.
// Returns:
// - true: if a token was successfully acquired and now can be used.
// - false: if no token was available at the time, and the caller should back off or return an error.
func (tb *TokenBucket) TryAcquire() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	// No tokens available
	return false
}

// Acquire tries to acquire a token, but if none is available, It waits until one is refilled.
// The waiting is done by periodically checking after each fill interval until a token is obtained
// or the context is canceled.
// Parameters:
//   - ctx: The context allows the caller to set timeouts or cancellations. If ctx is cancelled
//     before a token is acquired, Acquire returns an error.
func (tb *TokenBucket) Acquire(ctx context.Context) error {
	// First, try to get a token without waiting:
	if tb.TryAcquire() {
		return nil
	}

	// if  we didn't get a token, we need to wait for the next refill cycle.
	t := time.NewTicker(tb.fillInterval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			// If the context is canceled or times out before we get a token, return an error.
			return ErrContextTimeout
		case <-t.C:
			// On each tick, try again to acquire a token.
			if tb.TryAcquire() {
				return nil
			}
		}
	}
}
