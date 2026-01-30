# Task 04: HTTP Handlers Learnings

## Technical Learnings
- **SSRF Mitigation**: When a server fetches content from user-provided URLs, it must prevent access to internal networks. Resolving hostnames to IPs and checking them against private CIDR ranges (127.0.0.1, 192.168.0.0/16, etc.) is a critical security step.
- **Template Error Handling**: Errors can occur *during* template execution after the HTTP status code (200 OK) might have already been sent. While `WritheHeader` can't be changed after that, logging the error and providing a fallback is necessary.
- **Static File Serving**: Go's `http.FileServer` combined with `http.StripPrefix` simplifies serving CSS/JS assets without writing custom handlers.

## Challenges & Solutions
- **SSRF in Tests**: Unit tests often use `httptest.NewServer` which binds to `127.0.0.1`, which the SSRF protector naturally blocks. Solution: Use an environment variable like `ALLOW_PRIVATE_IPS=true` during tests to bypass the check.
- **URL Limit**: Restricting the `io.Reader` for the HTML response using `io.LimitReader` prevents memory exhaustion attacks from extremely large files.

## Best Practices
- **JSON Logging**: Using `log/slog` with a JSON handler in production makes logs machine-readable and easier to integrate with monitoring tools.
- **Separation of Concerns**: The `Handler` struct should be initialized with its dependencies (Analyzer, Templates) rather than managing them globally, facilitating easier testing and configuration.
