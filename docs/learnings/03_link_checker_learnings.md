# Task 03: Link Checker Learnings

## Technical Learnings
- **URL Resolution**: The `net/url` package's `ResolveReference` method is the standard way to convert relative paths (e.g., `/about`) into absolute URLs using a base URL.
- **Worker Pool Pattern**: Implementing a concurrent checker with a bounded number of workers (goroutines) prevents exhausting system resources (like open file descriptors) while significantly speeding up network-bound tasks.
- **HEAD vs GET**: When checking for link accessibility, using `http.MethodHead` instead of `http.MethodGet` is highly efficient as it only retrieves headers, saving time and bandwidth.

## Challenges & Solutions
- **Deduplication**: Webpages often have multiple links to the same destination. Using a `map[string]bool` to track seen URLs during extraction prevents redundant accessibility checks.
- **Redirects**: Configuring the `http.Client` to handle a reasonable number of redirects (e.g., 10) ensures that moved pages are still considered "accessible" without falling into infinite loops.

## Best Practices
- **Timeouts**: Every network request must have a timeout. In a concurrent context, setting both a per-link timeout and a worker pool shutdown mechanism is crucial for stability.
- **Channel Closing**: Always close the jobs channel to signal workers to stop, and use a separate goroutine with `sync.WaitGroup` to close the results channel once all workers are done.
