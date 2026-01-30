# Go Best Practices for This Project

## Code Organization

### Package Structure
```go
// internal/ - Application-specific code (not importable)
// internal/analyzer - Core business logic
// internal/handler - HTTP layer
// internal/models - Data structures
// internal/validator - Input validation
```

### Naming Conventions
- **Packages**: Short, lowercase, single word (`analyzer`, `handler`)
- **Files**: Lowercase with underscores (`html_parser.go`, `link_checker.go`)
- **Interfaces**: Suffix with "-er" (`Analyzer`, `Checker`)
- **Types**: PascalCase (`AnalysisResult`, `LinkType`)
- **Functions**: PascalCase for exported, camelCase for private

## Error Handling

### Wrap Errors with Context
```go
// Good
if err != nil {
    return fmt.Errorf("failed to fetch URL %s: %w", url, err)
}

// Bad
if err != nil {
    return err
}
```

### Custom Error Types
```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
```

## Concurrency

### Use Worker Pools
```go
// Create channels
jobs := make(chan Job, 100)
results := make(chan Result, 100)

// Spawn workers
var wg sync.WaitGroup
for w := 0; w < numWorkers; w++ {
    wg.Add(1)
    go worker(jobs, results, &wg)
}

// Send jobs
for _, job := range jobList {
    jobs <- job
}
close(jobs)

// Wait and close
go func() {
    wg.Wait()
    close(results)
}()
```

### Always Use Context for Timeouts
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
```

## HTTP Clients

### Configure Timeouts
```go
client := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        DialContext: (&net.Dialer{
            Timeout:   10 * time.Second,
            KeepAlive: 30 * time.Second,
        }).DialContext,
        TLSHandshakeTimeout: 10 * time.Second,
    },
}
```

### Set User-Agent
```go
req.Header.Set("User-Agent", "WebPageAnalyzer/1.0")
```

## Testing

### Table-Driven Tests
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"case 1", "input1", "output1"},
        {"case 2", "input2", "output2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Function(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Use httptest for HTTP Handlers
```go
func TestHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/", nil)
    w := httptest.NewRecorder()

    handler(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
    }
}
```

## Logging

### Use Structured Logging (slog)
```go
import "log/slog"

slog.Info("analysis started",
    "url", targetURL,
    "user_ip", remoteAddr)

slog.Error("fetch failed",
    "url", url,
    "error", err,
    "duration", elapsed)
```

## Configuration

### Use Environment Variables with Defaults
```go
func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

port := getEnv("PORT", "8080")
```

## Security

### SSRF Protection
```go
func isPrivateIP(ip net.IP) bool {
    privateRanges := []string{
        "10.0.0.0/8",
        "172.16.0.0/12",
        "192.168.0.0/16",
        "127.0.0.0/8",
    }

    for _, cidr := range privateRanges {
        _, network, _ := net.ParseCIDR(cidr)
        if network.Contains(ip) {
            return true
        }
    }
    return false
}
```

### Input Validation
```go
func ValidateURL(rawURL string) error {
    parsed, err := url.Parse(rawURL)
    if err != nil {
        return fmt.Errorf("invalid URL: %w", err)
    }

    if parsed.Scheme != "http" && parsed.Scheme != "https" {
        return fmt.Errorf("scheme must be http or https")
    }

    return nil
}
```

## Code Style

### Use gofmt and golangci-lint
```bash
go fmt ./...
golangci-lint run ./...
```

### Prefer Composition Over Inheritance
```go
// Good
type Service struct {
    analyzer Analyzer
    cache    Cache
}

// Avoid complex inheritance hierarchies
```

### Keep Functions Small
- Single responsibility
- 50 lines or fewer
- Clear function names

### Document Exported Types
```go
// Analyzer performs webpage analysis
type Analyzer interface {
    Analyze(url string) (*Result, error)
}
```

## Performance

### Reuse HTTP Clients
```go
// Create once
var httpClient = &http.Client{Timeout: 30 * time.Second}

// Reuse
resp, err := httpClient.Get(url)
```

### Limit Memory Usage
```go
// Limit response size
limitedReader := io.LimitReader(resp.Body, maxBytes)
```

### Use Buffering
```go
buf := make([]byte, 4096)
reader := bufio.NewReader(file)
```

## Don'ts

❌ Don't ignore errors  
❌ Don't use panic in libraries (only in main)  
❌ Don't use global mutable state  
❌ Don't create goroutines without cleanup  
❌ Don't use init() for complex initialization  
❌ Don't use blank identifier for unused variables in production code
