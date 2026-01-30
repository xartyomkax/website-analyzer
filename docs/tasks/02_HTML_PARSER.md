# Task 02: HTML Parser

## Objective
Implement HTML parsing functionality to extract HTML version, title, headings, and detect login forms.

## Prerequisites
- Task 01 (Project Setup) completed
- `goquery` library installed

## Implementation Steps

### 1. Define Data Models

**File: `internal/models/models.go`**

```go
package models

// AnalysisResult contains all analysis data for a webpage
type AnalysisResult struct {
	URL               string         `json:"url"`
	HTMLVersion       string         `json:"html_version"`
	Title             string         `json:"title"`
	Headings          map[string]int `json:"headings"`
	InternalLinks     int            `json:"internal_links"`
	ExternalLinks     int            `json:"external_links"`
	InaccessibleLinks []LinkError    `json:"inaccessible_links"`
	HasLoginForm      bool           `json:"has_login_form"`
}

// LinkError represents a link that could not be accessed
type LinkError struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code,omitempty"`
	Error      string `json:"error"`
}
```

### 2. Implement HTML Version Detection

**File: `internal/analyzer/html.go`**

```go
package analyzer

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// DetectHTMLVersion parses the DOCTYPE and returns the HTML version
func DetectHTMLVersion(doc *goquery.Document) string {
	// Get the HTML node
	htmlNode := doc.Find("html").First()
	if htmlNode.Length() == 0 {
		return "HTML5" // Default
	}

	// Try to get DOCTYPE from document
	// goquery doesn't directly expose DOCTYPE, so we check common patterns
	html, _ := doc.Html()
	htmlLower := strings.ToLower(html)

	// HTML5
	if strings.Contains(htmlLower, "<!doctype html>") {
		return "HTML5"
	}

	// HTML 4.01 Strict
	if strings.Contains(htmlLower, "html 4.01") && strings.Contains(htmlLower, "strict") {
		return "HTML 4.01 Strict"
	}

	// HTML 4.01 Transitional
	if strings.Contains(htmlLower, "html 4.01") && strings.Contains(htmlLower, "transitional") {
		return "HTML 4.01 Transitional"
	}

	// XHTML 1.0 Strict
	if strings.Contains(htmlLower, "xhtml 1.0") && strings.Contains(htmlLower, "strict") {
		return "XHTML 1.0 Strict"
	}

	// XHTML 1.0 Transitional
	if strings.Contains(htmlLower, "xhtml 1.0") && strings.Contains(htmlLower, "transitional") {
		return "XHTML 1.0 Transitional"
	}

	// Default to HTML5 for modern pages
	return "HTML5"
}
```

### 3. Implement Title Extraction

**Add to `internal/analyzer/html.go`**

```go
// ExtractTitle returns the page title, or "No title" if not found
func ExtractTitle(doc *goquery.Document) string {
	title := doc.Find("title").First().Text()
	title = strings.TrimSpace(title)
	
	if title == "" {
		return "No title"
	}
	
	return title
}
```

### 4. Implement Heading Counter

**Add to `internal/analyzer/html.go`**

```go
// CountHeadings counts headings by level (h1-h6)
func CountHeadings(doc *goquery.Document) map[string]int {
	headings := map[string]int{
		"h1": 0,
		"h2": 0,
		"h3": 0,
		"h4": 0,
		"h5": 0,
		"h6": 0,
	}

	// Count each heading level
	for level := 1; level <= 6; level++ {
		selector := fmt.Sprintf("h%d", level)
		count := doc.Find(selector).Length()
		headings[selector] = count
	}

	return headings
}
```

### 5. Implement Login Form Detection

**Add to `internal/analyzer/html.go`**

```go
// HasLoginForm checks if the page contains a login form
// (a form with a password input field)
func HasLoginForm(doc *goquery.Document) bool {
	// Find all forms
	hasPasswordInput := false
	
	doc.Find("form").Each(func(i int, form *goquery.Selection) {
		// Check if this form has a password input
		if form.Find("input[type='password']").Length() > 0 {
			hasPasswordInput = true
		}
	})
	
	return hasPasswordInput
}
```

### 6. Create Unit Tests

**File: `internal/analyzer/html_test.go`**

```go
package analyzer

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestDetectHTMLVersion(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "HTML5",
			html:     `<!DOCTYPE html><html><head></head><body></body></html>`,
			expected: "HTML5",
		},
		{
			name:     "HTML 4.01 Strict",
			html:     `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd"><html></html>`,
			expected: "HTML 4.01 Strict",
		},
		{
			name:     "No DOCTYPE",
			html:     `<html><head></head><body></body></html>`,
			expected: "HTML5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			result := DetectHTMLVersion(doc)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "Normal title",
			html:     `<html><head><title>Test Page</title></head></html>`,
			expected: "Test Page",
		},
		{
			name:     "Title with whitespace",
			html:     `<html><head><title>  Spaced Title  </title></head></html>`,
			expected: "Spaced Title",
		},
		{
			name:     "No title",
			html:     `<html><head></head></html>`,
			expected: "No title",
		},
		{
			name:     "Empty title",
			html:     `<html><head><title></title></head></html>`,
			expected: "No title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			result := ExtractTitle(doc)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestCountHeadings(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected map[string]int
	}{
		{
			name: "Various headings",
			html: `
				<html><body>
					<h1>Title</h1>
					<h2>Section 1</h2>
					<h2>Section 2</h2>
					<h3>Subsection</h3>
				</body></html>
			`,
			expected: map[string]int{
				"h1": 1,
				"h2": 2,
				"h3": 1,
				"h4": 0,
				"h5": 0,
				"h6": 0,
			},
		},
		{
			name: "No headings",
			html: `<html><body><p>No headings here</p></body></html>`,
			expected: map[string]int{
				"h1": 0,
				"h2": 0,
				"h3": 0,
				"h4": 0,
				"h5": 0,
				"h6": 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			result := CountHeadings(doc)
			
			for level, expected := range tt.expected {
				if result[level] != expected {
					t.Errorf("Heading %s: expected %d, got %d", level, expected, result[level])
				}
			}
		})
	}
}

func TestHasLoginForm(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected bool
	}{
		{
			name: "Has login form",
			html: `
				<html><body>
					<form action="/login" method="post">
						<input type="text" name="username">
						<input type="password" name="password">
						<button type="submit">Login</button>
					</form>
				</body></html>
			`,
			expected: true,
		},
		{
			name: "No password input",
			html: `
				<html><body>
					<form action="/search" method="get">
						<input type="text" name="q">
						<button type="submit">Search</button>
					</form>
				</body></html>
			`,
			expected: false,
		},
		{
			name:     "No forms",
			html:     `<html><body><p>No forms here</p></body></html>`,
			expected: false,
		},
		{
			name: "Multiple forms, one with password",
			html: `
				<html><body>
					<form action="/search"><input type="text" name="q"></form>
					<form action="/login"><input type="password" name="pass"></form>
				</body></html>
			`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			result := HasLoginForm(doc)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
```

### 7. Run Tests

```bash
# Run tests for the analyzer package
go test -v ./internal/analyzer/

# Run with coverage
go test -v -cover ./internal/analyzer/
```

## Acceptance Criteria

- ✅ `DetectHTMLVersion()` correctly identifies HTML5, HTML4.01, XHTML
- ✅ `ExtractTitle()` extracts page title and handles missing titles
- ✅ `CountHeadings()` counts all heading levels (h1-h6)
- ✅ `HasLoginForm()` detects forms with password inputs
- ✅ All functions have unit tests
- ✅ Test coverage ≥80% for `html.go`
- ✅ All tests pass

## Verification

```bash
# Run all tests
make test

# Check coverage
make test-coverage
# Open coverage.html in browser to verify >80% coverage

# Format code
make fmt
```

Expected output:
```
=== RUN   TestDetectHTMLVersion
=== RUN   TestDetectHTMLVersion/HTML5
=== RUN   TestDetectHTMLVersion/HTML_4.01_Strict
...
--- PASS: TestDetectHTMLVersion (0.00s)
...
PASS
coverage: 85.7% of statements
```

## Next Steps

Once this task is complete, proceed to:
- [Task 03: Link Checker](03_LINK_CHECKER.md)

## Troubleshooting

### goquery import errors
```bash
go get github.com/PuerkitoBio/goquery
go mod tidy
```

### Test failures
- Check HTML structure in test cases
- Verify goquery selectors are correct
- Ensure whitespace handling in title extraction

## Related Documentation
- [DATA_MODELS.md](../specs/DATA_MODELS.md) - Data structure definitions
- [TESTING_STRATEGY.md](../guides/TESTING_STRATEGY.md) - Testing patterns
