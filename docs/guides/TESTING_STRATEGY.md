# Testing Strategy

## Overview
This project targets **80%+ code coverage** with comprehensive unit tests. We do **not** implement E2E tests for the MVP.

## Test Organization

### Directory Structure
```
internal/
├── analyzer/
│   ├── analyzer.go
│   ├── analyzer_test.go
│   ├── html.go
│   ├── html_test.go
│   ├── links.go
│   ├── links_test.go
│   ├── checker.go
│   └── checker_test.go
├── handler/
│   ├── handler.go
│   └── handler_test.go
└── validator/
    ├── validator.go
    └── validator_test.go
```

## Test Types

### 1. Unit Tests

Test individual functions in isolation.

**Example: HTML Parser**
```go
func TestExtractTitle(t *testing.T) {
    tests := []struct {
        name     string
        html     string
        expected string
    }{
        {
            name:     "normal title",
            html:     `<html><head><title>Test</title></head></html>`,
            expected: "Test",
        },
        {
            name:     "no title",
            html:     `<html><head></head></html>`,
            expected: "No title",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            doc, _ := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
            result := ExtractTitle(doc)
            
            if result != tt.expected {
                t.Errorf("got %q, want %q", result, tt.expected)
            }
        })
    }
}
```

### 2. Integration Tests

Test multiple components working together.

**Example: Handler + Analyzer**
```go
func TestAnalyzeHandler(t *testing.T) {
    // Create test server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`<html><head><title>Test</title></head></html>`))
    }))
    defer server.Close()

    // Setup
    config := &analyzer.Config{
        RequestTimeout: 5 * time.Second,
        LinkTimeout:    2 * time.Second,
        MaxWorkers:     5,
        MaxResponseSize: 10 * 1024 * 1024,
    }
    
    analyzer := analyzer.NewAnalyzer(config)
    handler, _ := handler.NewHandler(analyzer, "../../web/templates")

    // Test request
    form := url.Values{}
    form.Add("url", server.URL)
    
    req := httptest.NewRequest("POST", "/analyze", strings.NewReader(form.Encode()))
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    
    w := httptest.NewRecorder()
    handler.AnalyzeHandler(w, req)

    // Assertions
    if w.Code != http.StatusOK {
        t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
    }
}
```

### 3. Mock HTTP Tests

Use `httptest` to mock external HTTP calls.

**Example: Link Checker**
```go
func TestCheckLinks(t *testing.T) {
    // Mock servers
    server200 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    defer server200.Close()

    server404 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusNotFound)
    }))
    defer server404.Close()

    links := []models.Link{
        {URL: server200.URL, Type: models.LinkTypeExternal},
        {URL: server404.URL, Type: models.LinkTypeExternal},
    }

    config := CheckLinksConfig{
        Timeout:    5 * time.Second,
        MaxWorkers: 2,
    }

    errors := CheckLinks(links, config)

    // Should have 1 error (404)
    if len(errors) != 1 {
        t.Errorf("expected 1 error, got %d", len(errors))
    }
}
```

## Testing Patterns

### Table-Driven Tests

**Why**: Test multiple scenarios efficiently.

```go
func TestValidateURL(t *testing.T) {
    tests := []struct {
        name      string
        url       string
        shouldErr bool
        errMsg    string
    }{
        {
            name:      "valid http",
            url:       "http://example.com",
            shouldErr: false,
        },
        {
            name:      "valid https",
            url:       "https://example.com",
            shouldErr: false,
        },
        {
            name:      "invalid scheme",
            url:       "ftp://example.com",
            shouldErr: true,
            errMsg:    "scheme must be http or https",
        },
        {
            name:      "empty url",
            url:       "",
            shouldErr: true,
            errMsg:    "URL is required",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateURL(tt.url)
            
            if tt.shouldErr && err == nil {
                t.Error("expected error but got nil")
            }
            
            if !tt.shouldErr && err != nil {
                t.Errorf("unexpected error: %v", err)
            }
            
            if tt.shouldErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
                t.Errorf("error message %q doesn't contain %q", err.Error(), tt.errMsg)
            }
        })
    }
}
```

### Subtests

**Why**: Better test organization and selective running.

```go
func TestAnalyzer(t *testing.T) {
    t.Run("HTML parsing", func(t *testing.T) {
        t.Run("extract title", func(t *testing.T) {
            // Test title extraction
        })
        
        t.Run("count headings", func(t *testing.T) {
            // Test heading counts
        })
    })
    
    t.Run("Link processing", func(t *testing.T) {
        // Test link extraction
    })
}
```

Run specific subtest:
```bash
go test -v -run TestAnalyzer/HTML_parsing/extract_title
```

## Coverage Requirements

### Target: 80%+

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage
go tool cover -func=coverage.out

# HTML report
go tool cover -html=coverage.out -o coverage.html
```

### What to Cover

✅ **Must Cover**:
- All public functions
- Error paths
- Edge cases
- Boundary conditions

❌ **Can Skip**:
- Simple getters/setters
- Constants
- Type definitions

### Coverage by Package

**Target coverage per package**:
- `internal/analyzer`: 85%+
- `internal/handler`: 75%+
- `internal/validator`: 90%+
- `internal/models`: 50%+ (mostly structs)

## Test Helpers

### Creating Test HTML
```go
func createTestHTML(title, body string) string {
    return fmt.Sprintf(`
        <!DOCTYPE html>
        <html>
            <head><title>%s</title></head>
            <body>%s</body>
        </html>
    `, title, body)
}
```

### Test Fixtures
```go
var testCases = map[string]string{
    "simple": `<html><head><title>Simple</title></head></html>`,
    "complex": `<html><head><title>Complex</title></head><body>
        <h1>Title</h1>
        <h2>Section</h2>
        <a href="/link">Link</a>
    </body></html>`,
}
```

## Running Tests

### All Tests
```bash
go test ./...
go test -v ./...  # Verbose
```

### Specific Package
```bash
go test ./internal/analyzer
go test -v ./internal/analyzer
```

### With Coverage
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
```

### Race Detection
```bash
go test -race ./...
```

### Benchmarks (Future)
```bash
go test -bench=. ./...
```

## Continuous Integration

### GitHub Actions Example
```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.21
      
      - name: Run tests
        run: |
          go test -v -race -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out
      
      - name: Upload coverage
        uses: codecov/codecov-action@v2
        with:
          files: ./coverage.out
```

## Best Practices

### ✅ Do

- Write tests first (TDD) or alongside code
- Use table-driven tests for multiple scenarios
- Test both success and error paths
- Use meaningful test names
- Keep tests simple and focused
- Mock external dependencies
- Use `t.Helper()` for test utilities
- Clean up resources (defer close)

### ❌ Don't

- Test implementation details
- Write flaky tests (timing-dependent)
- Share state between tests
- Skip cleanup
- Ignore race conditions
- Test standard library behavior
- Over-mock (test too much in isolation)

## Debugging Tests

### Print Debug Info
```go
t.Logf("got: %v, want: %v", got, want)
```

### Skip Long Tests
```go
if testing.Short() {
    t.Skip("skipping test in short mode")
}
```

Run short tests only:
```bash
go test -short ./...
```

### Parallel Tests
```go
func TestSomething(t *testing.T) {
    t.Parallel()  // Run in parallel with other tests
    
    // Test code
}
```

## Test Checklist

Before committing:
- [ ] All tests pass
- [ ] Coverage ≥80%
- [ ] No race conditions
- [ ] Tests are deterministic
- [ ] Error cases tested
- [ ] Edge cases covered
- [ ] Tests are readable

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Table Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Test Fixtures](https://dave.cheney.net/2016/05/10/test-fixtures-in-go)
