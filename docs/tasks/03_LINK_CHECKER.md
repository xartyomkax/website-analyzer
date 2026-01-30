# Task 03: Link Checker

## Objective
Implement concurrent link extraction, classification (internal/external), and accessibility checking using goroutines and channels.

## Prerequisites
- Task 02 (HTML Parser) completed
- Understanding of Go concurrency (goroutines, channels)

## Implementation Steps

### 1. Add Link-Related Models

**Update `internal/models/models.go`**

```go
package models

// LinkType represents the category of a link
type LinkType int

const (
	LinkTypeInternal LinkType = iota
	LinkTypeExternal
	LinkTypeInvalid
)

func (lt LinkType) String() string {
	switch lt {
	case LinkTypeInternal:
		return "internal"
	case LinkTypeExternal:
		return "external"
	default:
		return "invalid"
	}
}

// Link represents a hyperlink found in the document
type Link struct {
	URL  string   `json:"url"`
	Type LinkType `json:"type"`
}
```

### 2. Implement Link Extraction

**File: `internal/analyzer/links.go`**

```go
package analyzer

import (
	"fmt"
	"net/url"
	"strings"

	"webpage-analyzer/internal/models"

	"github.com/PuerkitoBio/goquery"
)

// ExtractLinks finds all <a href> tags and returns their URLs
func ExtractLinks(doc *goquery.Document, baseURL string) ([]models.Link, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	var links []models.Link
	seen := make(map[string]bool) // Deduplicate

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" {
			return
		}

		// Resolve relative URLs
		resolved, err := resolveURL(base, href)
		if err != nil || resolved == "" {
			return
		}

		// Skip duplicates
		if seen[resolved] {
			return
		}
		seen[resolved] = true

		// Classify link
		linkType := classifyLink(resolved, base)
		
		links = append(links, models.Link{
			URL:  resolved,
			Type: linkType,
		})
	})

	return links, nil
}

// resolveURL converts relative URLs to absolute
func resolveURL(base *url.URL, href string) (string, error) {
	href = strings.TrimSpace(href)

	// Skip invalid schemes
	if strings.HasPrefix(href, "javascript:") ||
		strings.HasPrefix(href, "mailto:") ||
		strings.HasPrefix(href, "tel:") ||
		href == "#" ||
		strings.HasPrefix(href, "#") {
		return "", nil
	}

	// Parse href
	parsed, err := url.Parse(href)
	if err != nil {
		return "", err
	}

	// Resolve against base
	resolved := base.ResolveReference(parsed)
	
	// Only return http/https URLs
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return "", nil
	}

	return resolved.String(), nil
}

// classifyLink determines if a link is internal or external
func classifyLink(link string, base *url.URL) models.LinkType {
	parsed, err := url.Parse(link)
	if err != nil {
		return models.LinkTypeInvalid
	}

	// Same host (including subdomains) = internal
	if parsed.Host == base.Host {
		return models.LinkTypeInternal
	}

	return models.LinkTypeExternal
}
```

### 3. Implement Concurrent Link Checker

**File: `internal/analyzer/checker.go`**

```go
package analyzer

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"webpage-analyzer/internal/models"
)

// CheckLinksConfig holds configuration for link checking
type CheckLinksConfig struct {
	Timeout    time.Duration
	MaxWorkers int
}

// checkResult is used internally for worker communication
type checkResult struct {
	url        string
	statusCode int
	err        error
}

// CheckLinks verifies accessibility of links concurrently
func CheckLinks(links []models.Link, config CheckLinksConfig) []models.LinkError {
	if len(links) == 0 {
		return nil
	}

	// Channels for work distribution
	jobs := make(chan models.Link, len(links))
	results := make(chan checkResult, len(links))

	// Start worker pool
	var wg sync.WaitGroup
	workerCount := config.MaxWorkers
	if workerCount <= 0 {
		workerCount = 10
	}

	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go worker(jobs, results, config.Timeout, &wg)
	}

	// Send jobs
	for _, link := range links {
		jobs <- link
	}
	close(jobs)

	// Wait for workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect errors
	var errors []models.LinkError
	for result := range results {
		if result.err != nil {
			errors = append(errors, models.LinkError{
				URL:        result.url,
				StatusCode: result.statusCode,
				Error:      result.err.Error(),
			})
		}
	}

	return errors
}

// worker processes link checking jobs
func worker(jobs <-chan models.Link, results chan<- checkResult, timeout time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	for link := range jobs {
		result := checkLink(client, link.URL)
		results <- result
	}
}

// checkLink performs a single link check
func checkLink(client *http.Client, url string) checkResult {
	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return checkResult{
			url:        url,
			statusCode: 0,
			err:        err,
		}
	}

	req.Header.Set("User-Agent", "WebPageAnalyzer/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return checkResult{
			url:        url,
			statusCode: 0,
			err:        err,
		}
	}
	defer resp.Body.Close()

	// Consider 2xx and 3xx as success
	if resp.StatusCode >= 400 {
		return checkResult{
			url:        url,
			statusCode: resp.StatusCode,
			err:        fmt.Errorf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode)),
		}
	}

	return checkResult{
		url:        url,
		statusCode: resp.StatusCode,
		err:        nil,
	}
}
```

### 4. Create Unit Tests

**File: `internal/analyzer/links_test.go`**

```go
package analyzer

import (
	"strings"
	"testing"

	"webpage-analyzer/internal/models"

	"github.com/PuerkitoBio/goquery"
)

func TestExtractLinks(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		baseURL  string
		expected int
		internal int
		external int
	}{
		{
			name: "Internal and external links",
			html: `
				<html><body>
					<a href="/about">About</a>
					<a href="https://example.com/contact">Contact</a>
					<a href="https://google.com">Google</a>
				</body></html>
			`,
			baseURL:  "https://example.com",
			expected: 3,
			internal: 2,
			external: 1,
		},
		{
			name: "Skip invalid links",
			html: `
				<html><body>
					<a href="javascript:void(0)">JS</a>
					<a href="mailto:test@example.com">Email</a>
					<a href="#">Anchor</a>
					<a href="/valid">Valid</a>
				</body></html>
			`,
			baseURL:  "https://example.com",
			expected: 1,
			internal: 1,
			external: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			links, err := ExtractLinks(doc, tt.baseURL)

			if err != nil {
				t.Fatalf("ExtractLinks failed: %v", err)
			}

			if len(links) != tt.expected {
				t.Errorf("Expected %d links, got %d", tt.expected, len(links))
			}

			internal := 0
			external := 0
			for _, link := range links {
				if link.Type == models.LinkTypeInternal {
					internal++
				} else if link.Type == models.LinkTypeExternal {
					external++
				}
			}

			if internal != tt.internal {
				t.Errorf("Expected %d internal links, got %d", tt.internal, internal)
			}
			if external != tt.external {
				t.Errorf("Expected %d external links, got %d", tt.external, external)
			}
		})
	}
}

func TestClassifyLink(t *testing.T) {
	baseURL := mustParseURL("https://example.com")

	tests := []struct {
		name     string
		link     string
		expected models.LinkType
	}{
		{"Internal same path", "https://example.com/about", models.LinkTypeInternal},
		{"Internal root", "https://example.com/", models.LinkTypeInternal},
		{"External", "https://google.com", models.LinkTypeExternal},
		{"External subdomain", "https://blog.example.com", models.LinkTypeExternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifyLink(tt.link, baseURL)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// Helper
func mustParseURL(s string) *url.URL {
	u, _ := url.Parse(s)
	return u
}
```

**File: `internal/analyzer/checker_test.go`**

```go
package analyzer

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"webpage-analyzer/internal/models"
)

func TestCheckLinks(t *testing.T) {
	// Create test servers
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
		t.Errorf("Expected 1 error, got %d", len(errors))
	}

	if len(errors) > 0 && errors[0].StatusCode != 404 {
		t.Errorf("Expected status code 404, got %d", errors[0].StatusCode)
	}
}

func TestCheckLinksTimeout(t *testing.T) {
	// Create slow server
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer slowServer.Close()

	links := []models.Link{
		{URL: slowServer.URL, Type: models.LinkTypeExternal},
	}

	config := CheckLinksConfig{
		Timeout:    100 * time.Millisecond, // Very short timeout
		MaxWorkers: 1,
	}

	errors := CheckLinks(links, config)

	// Should timeout
	if len(errors) != 1 {
		t.Errorf("Expected timeout error, got %d errors", len(errors))
	}
}
```

### 5. Run Tests

```bash
# Test link extraction
go test -v ./internal/analyzer/ -run TestExtractLinks

# Test link checking
go test -v ./internal/analyzer/ -run TestCheckLinks

# Test all
make test
```

## Acceptance Criteria

- ✅ `ExtractLinks()` finds all `<a href>` tags
- ✅ Resolves relative URLs to absolute
- ✅ Skips invalid links (javascript:, mailto:, #)
- ✅ Classifies links as internal/external correctly
- ✅ `CheckLinks()` verifies link accessibility concurrently
- ✅ Uses worker pool with configurable number of workers
- ✅ Handles timeouts gracefully
- ✅ Returns proper error details (status code, message)
- ✅ All tests pass
- ✅ Test coverage ≥80%

## Verification

```bash
make test
make test-coverage
```

Expected behavior:
- Link extraction completes successfully
- Worker pool processes links concurrently
- Timeouts are respected (5s per link)
- HTTP errors (404, 500) are captured

## Performance Notes

**Worker Pool Benefits**:
- 50 links with 10 workers: ~5s (vs ~250s sequential)
- Respects timeouts individually per link
- Doesn't overwhelm target servers

**Configuration**:
```go
config := CheckLinksConfig{
    Timeout:    5 * time.Second,  // Per link
    MaxWorkers: 10,                // Concurrent checkers
}
```

## Next Steps

Once this task is complete, proceed to:
- [Task 04: HTTP Handlers](04_HTTP_HANDLERS.md)

## Troubleshooting

### Too many open files
Reduce MaxWorkers if hitting OS limits:
```go
MaxWorkers: 5  // Instead of 20
```

### Context deadline exceeded
Increase timeout for slow networks:
```go
Timeout: 10 * time.Second
```

## Related Documentation
- [ARCHITECTURE.md](../specs/ARCHITECTURE.md) - Concurrency model
- [GO_BEST_PRACTICES.md](../guides/GO_BEST_PRACTICES.md) - Goroutine patterns
