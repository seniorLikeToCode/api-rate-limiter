# Rate Limiter Example in Go

This project provides a basic example of implementing a rate limiter in Go using a **token bucket algorithm**, integrated with a simple HTTP server. The rate limiter ensures that incoming requests do not exceed a specified rate, helping prevent server overload and ensuring fair usage patterns.

## Key Concepts

### Token Bucket Algorithm

The **token bucket** algorithm is a commonly used technique for rate limiting. It works as follows:

-   A "bucket" is filled with tokens at a constant rate, up to a maximum capacity.
-   Each incoming request must acquire a token from the bucket before proceeding.
-   If a token is available, the request continues; if not, the request is denied (or must wait, depending on the implementation).

This approach allows short bursts up to the bucket's capacity but enforces an overall average rate determined by the token refill interval.

### Leaky Bucket Algorithm (Not Yet Implemented)

The **leaky bucket** algorithm is another rate limiting strategy:

-   Imagine a bucket with a small hole at the bottom where water (requests) leaks out at a steady rate.
-   Requests pour into the bucket at varying rates.
-   If the bucket overflows, excess requests are dropped.

Leaky bucket ensures a smooth, constant outflow (processing rate), but it may discard bursts of traffic that exceed the bucket capacity.

### Sliding Window Algorithm (Not Yet Implemented)

The **sliding window** algorithm is typically used with time-based counters:

-   It measures the number of requests received over the last fixed time window (e.g., the past N seconds).
-   If the count exceeds a threshold, new requests are denied.
-   The window "slides" forward in time, always considering only the most recent period.

This approach can be implemented using in-memory counters or distributed data stores, but it tends to be more complex and less burst-friendly compared to token or leaky buckets.

### Future Implementations

We’ve started with the token bucket algorithm. However, you may want to extend this project with:

-   **Leaky Bucket**: Implement a queuing mechanism that processes requests at a constant rate, discarding any overflow.
-   **Sliding Window**: Maintain a time-based rolling window of request counts to smooth out variance and ensure a strict limit over time.
-   **Fixed Window**: A simpler variant of the sliding window, where you bucket requests into discrete intervals (like per second or per minute). This can be easier to implement but less flexible.

## Project Structure

```
rate-limiter-example/
    main.go            // Entry point of the server
    limiter/
        limiter.go     // Token bucket implementation
        limiter_test.go// Tests for the rate limiter
```

-   **`main.go`**: Sets up an HTTP server and applies the token bucket rate limiter to incoming requests.
-   **`limiter.go`**: Contains the `TokenBucket` type, responsible for token-based rate limiting.
-   **`limiter_test.go`**: Unit tests verifying the functionality of the token bucket algorithm.

## Running the Project

1. **Initialize the module:**

    ```bash
    go mod init ratelimiter
    go mod tidy
    ```

2. **Run the server:**

    ```bash
    go run main.go
    ```

    The server will start listening on `:4000`.

3. **Test requests:**
   Use `curl` or a browser to access `http://localhost:4000`.  
   If you send multiple requests rapidly, you will eventually receive HTTP 429 responses ("Too Many Requests") once the token bucket is depleted.

## Testing

To run the tests for the token bucket rate limiter:

```bash
go test ./limiter
```

## Example Scripts

We have provided a Bash script to send multiple requests:

-   `script.sh`: Sends 500 requests to `http://localhost:4000` and counts the number of successful (200) and rate-limited (429) responses.

Run the script:

```bash
chmod +x script.sh
./script.sh
```

## Future Enhancements

-   **Leaky Bucket Implementation**: Introduce a mechanism to queue requests and process them at a steady outflow rate.
-   **Sliding Window or Fixed Window Algorithms**: Implement a window-based approach to rate limiting, controlling request bursts over defined intervals.
-   **Distributed State**: Extend the design to work across multiple servers, potentially using a distributed cache or message queue for rate limiting at scale.
-   **Configuration & Metrics**: Add configuration options for capacities, intervals, and reporting metrics for observability and tuning.

With these extensions, you can build a more comprehensive and flexible rate-limiting solution, adapting to various application requirements and traffic patterns.
