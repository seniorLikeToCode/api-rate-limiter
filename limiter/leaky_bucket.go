package limiter

import (
	"sync"
	"time"
)

// LeakyBucket is a rate limiter that uses the leaky bucket algorithm.
//
// The leaky bucket algorithm works as follows:
//   - There is a bucket with a fixed capacity that represents the maximum
//     number of tokens (requests) that can be "in-flight".
//   - Tokens typically represent opportunities to handle requests.
//   - Tokens "leak" (i.e., are removed) from the bucket at a fixed rate (leakInterval)
//     and in fixed increments (leakCount) to simulate a steady outflow of capacity over time.
//   - When a new request arrives, it tries to acquire a token. If the bucket is not full,
//     a token is added immediately (and thus the request is allowed). If the bucket is full,
//     the request must wait until there is room (in some approaches) or simply be rejected.
//
// In this implementation:
//   - `Allow()` attempts to add a token to the bucket. If successful, it returns true.
//     If the bucket is full, it returns false, indicating the request should be rejected or
//     deferred.
//   - The bucket automatically leaks tokens at the specified interval.
type leakyBucket struct {
	capacity  int           // Maximum number of tokens the bucket can hold at once.
	leakRate  time.Duration // Interval at which tokens are removed (leaked).
	mu        sync.Mutex    // Protects access to the bucket state.
	queue     []struct{}    // Represents the current tokens in the bucket.
	stopCh    chan struct{} // Signals that the leaking goroutine should stop.
	stopped   bool          // Indicates if the limiter has been stopped.
	leakCount int           // Number of tokens to remove each leak interval.
}

// NewLeakyBucket creates a new leaky bucket limiter with the given capacity,
// leakInterval, and leakCount. The leakCount is how many tokens are removed
// from the bucket each leak interval.
// For example, NewLeakyBucket(10, 100*time.Millisecond, 1) would create a bucket
// that can hold up to 10 tokens and leaks 1 token every 100 ms.
func NewLeakyBucket(capacity int, leakInterval time.Duration, leakCount int) *leakyBucket {
	lb := &leakyBucket{
		capacity:  capacity,
		leakRate:  leakInterval,
		queue:     make([]struct{}, 0, capacity),
		stopCh:    make(chan struct{}),
		leakCount: leakCount,
	}

	// Start a background goroutine to continuously remove tokens at the specified leak rate.
	go lb.startLeaking()
	return lb
}

// startLeaking runs in a dedicated goroutine. It periodically removes up to
// `leakCount` tokens from the bucket at each interval. If the limiter is stopped,
// it exits.
func (lb *leakyBucket) startLeaking() {
	ticker := time.NewTicker(lb.leakRate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			lb.mu.Lock()
			if lb.stopped {
				lb.mu.Unlock()
				return
			}

			// Remove up to `leakCount` tokens if available.
			removeCount := lb.leakCount
			if removeCount > len(lb.queue) {
				removeCount = len(lb.queue)
			}

			lb.queue = lb.queue[removeCount:]
			lb.mu.Unlock()

		case <-lb.stopCh:
			// The limiter has been stopped, exit the goroutine.
			return
		}
	}
}

// Allow attempts to add a token to the bucket. If there's capacity, it returns true,
// meaning the request can proceed. If not, it returns false, indicating the request
// should be rejected or delayed.
func (lb *leakyBucket) Allow() bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if lb.stopped {
		return false
	}

	if len(lb.queue) < lb.capacity {
		// There is room for a new token
		lb.queue = append(lb.queue, struct{}{})
		return true
	}

	// Bucket is full
	return false
}

// Stop stops the leaking goroutine and prevents any further tokens from being added.
func (lb *leakyBucket) Stop() {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if !lb.stopped {
		lb.stopped = true
		close(lb.stopCh)
	}
}

// CurrentSize returns the current number of tokens in the bucket.
// This can be useful for monitoring or debugging.
func (lb *leakyBucket) CurrentSize() int {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	return len(lb.queue)
}

// Capacity returns the maximum capacity of the bucket.
func (lb *leakyBucket) Capacity() int {
	return lb.capacity
}
