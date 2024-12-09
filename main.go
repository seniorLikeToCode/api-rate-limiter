package main

import (
	"fmt"
	"net/http"
	"ratelimiter/limiter"
	"time"
	// limiter "./ratelimiter/limiter/limiter.go"
)

// rateLimitedHandler is an HTTP handler function that uses the rate limiter
// to control incoming requests. In this example, we use TryAcquire, which
// means if there's no available tokek right away, we return a 429 Too Many Request/.
func rateLimitedHandler(rl limiter.RateLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Here we use TryAcquire for a non-blocking approach.
		// If we prefer to wait, we could use Acquire with a timeout context, e.g.:
		// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		// defer cancel()
		// if err := rl.Acquire(ctx); err != nil{...}

		if !rl.TryAcquire() {
			// No token available immediately, send a HTTP 429 response.
			http.Error(w, "Too many Requests", http.StatusTooManyRequests)
			return
		}

		// If a token was acquire, proceed with normal request handling.
		fmt.Fprintln(w, "Request succeeded!")
	}
}

func main() {
	// Create a new TokenBucket rate limiter.
	// Let's configure it to:
	// - Have a capacity of 5 tokens maximum.
	// - Add one token every 200 milliseconds.
	// This roughly allowd up to 5 immediate requests, and then replenished one every 200ms.
	rl := limiter.NewTokenBucket(5, 20*time.Millisecond)
	defer rl.Stop()

	// create a new HTTP server
	mux := http.NewServeMux()

	// Assign our rate-limited handler to the root endpoint.
	mux.HandleFunc("/", rateLimitedHandler(rl))

	server := &http.Server{
		Addr:    ":4000", // Listem on port 4000
		Handler: mux,
	}

	fmt.Println("Starting server on :4000")
	// Start the server. Visit http://localhost:4000 in your browser or use curl:
	// curl http://localhost:4000

	// If you hit it rapidly you'll start seeing 429 respones after the capacity is exhausted.
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
