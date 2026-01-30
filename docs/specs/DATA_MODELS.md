# Data Models

This document defines all Go structs, interfaces, and type definitions used in the application.

## Package: internal/models

### AnalysisRequest
User input for analysis.

```go
package models

// AnalysisRequest represents the user's request to analyze a URL
type AnalysisRequest struct {
    URL string `json:"url" form:"url"`
}

// Validate checks if the request is valid
func (r *AnalysisRequest) Validate() error {
    if r.URL == "" {
        return errors.New("URL is required")
    }
    if len(r.URL) > 2048 {
        return errors.New("URL too long (max 2048 characters)")
    }
    return nil
}
```

### AnalysisResult
Complete analysis results.

```go
// AnalysisResult contains all analysis data for a webpage
type AnalysisResult struct {
    URL               string            `json:"url"`
    HTMLVersion       string            `json:"html_version"`
    Title             string            `json:"title"`
    Headings          map[string]int    `json:"headings"`
    InternalLinks     int               `json:"internal_links"`
    ExternalLinks     int               `json:"external_links"`
    InaccessibleLinks []LinkError       `json:"inaccessible_links"`
    HasLoginForm      bool              `json:"has_login_form"`
}

// Example:
// {
//   "url": "https://example.com",
//   "html_version": "HTML5",
//   "title": "Example Domain",
//   "headings": {"h1": 1, "h2": 3, "h3": 0, "h4": 0, "h5": 0, "h6": 0},
//   "internal_links": 5,
//   "external_links": 2,
//   "inaccessible_links": [{"url": "https://broken.com", "status_code": 404, "error": "Not Found"}],
//   "has_login_form": false
// }
```

### LinkError
Information about inaccessible links.

```go
// LinkError represents a link that could not be accessed
type LinkError struct {
    URL        string `json:"url"`
    StatusCode int    `json:"status_code,omitempty"` // 0 if no HTTP response
    Error      string `json:"error"`
}

// Examples:
// LinkError{URL: "https://example.com/404", StatusCode: 404, Error: "Not Found"}
// LinkError{URL: "https://timeout.com", StatusCode: 0, Error: "context deadline exceeded"}
```

### LinkType
Enumeration for link classification.

```go
// LinkType represents the category of a link
type LinkType int

const (
    LinkTypeInternal LinkType = iota // Same domain
    LinkTypeExternal                  // Different domain
    LinkTypeInvalid                   // Invalid or unsupported URL
)

func (lt LinkType) String() string {
    switch lt {
    case LinkTypeInternal:
        return "internal"
    case LinkTypeExternal:
        return "external"
    case LinkTypeInvalid:
        return "invalid"
    default:
        return "unknown"
    }
}
```

### Link
Represents a parsed link with metadata.

```go
// Link represents a hyperlink found in the document
type Link struct {
    URL      string   `json:"url"`
    Type     LinkType `json:"type"`
    Text     string   `json:"text,omitempty"` // Anchor text (optional)
}
```

## Package: internal/analyzer

### AnalysisConfig
Configuration for the analyzer.

```go
package analyzer

import "time"

// Config holds configuration for the analyzer
type Config struct {
	RequestTimeout  time.Duration // Timeout for fetching main page
	LinkTimeout     time.Duration // Timeout per link check
	MaxWorkers      int           // Number of concurrent link checkers
	MaxResponseSize int64         // Maximum response size in bytes
	MaxURLLength    int           // Maximum URL length allowed
	MaxRedirects    int           // Maximum number of redirects to follow
}
```

### Analyzer
Main analyzer interface.

```go
// Analyzer is the standard implementation
type Analyzer struct {
	config     *Config
	httpClient *http.Client
}

// NewAnalyzer creates a new analyzer instance
func NewAnalyzer(config *Config) *Analyzer {
	return &Analyzer{
		config: config,
		httpClient: &http.Client{
			Timeout: config.RequestTimeout,
		},
	}
}

// Analyze performs the complete analysis
func (a *Analyzer) Analyze(url string) (*models.AnalysisResult, error) {
    // Implementation
}
```

## Package: internal/handler

### TemplateData
Data structures for template rendering.

```go
package handler

import "website-analyzer/internal/models"

// IndexData is passed to the index.html template
type IndexData struct {
    Error string // Error message to display (if any)
}

// ResultsData is passed to the results.html template
type ResultsData struct {
    Result *models.AnalysisResult
    Error  string // Error message if analysis failed
}
```

### ErrorResponse
HTTP error response structure.

```go
// ErrorResponse represents an HTTP error response
type ErrorResponse struct {
    Message    string `json:"message"`
    StatusCode int    `json:"status_code"`
    Details    string `json:"details,omitempty"` // Technical details
}

// NewErrorResponse creates a new error response
func NewErrorResponse(statusCode int, message string) *ErrorResponse {
    return &ErrorResponse{
        Message:    message,
        StatusCode: statusCode,
    }
}

// WithDetails adds technical details to the error
func (e *ErrorResponse) WithDetails(details string) *ErrorResponse {
    e.Details = details
    return e
}
```

## Package: internal/validator

### ValidationError
Custom error type for validation failures.

```go
package validator

import "fmt"

// ValidationError represents a validation failure
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
    return &ValidationError{
        Field:   field,
        Message: message,
    }
}
```

## Internal Types (not exported)

### linkCheckJob
Used internally for worker pool communication.

```go
package analyzer

// linkCheckJob represents a link to check (internal use only)
type linkCheckJob struct {
    url      string
    linkType models.LinkType
}

// linkCheckResult represents the result of checking a link
type linkCheckResult struct {
    url        string
    statusCode int
    err        error
}
```

## Type Aliases and Constants

### HTTP Status Codes
```go
package handler

const (
    StatusBadRequest          = 400
    StatusInternalServerError = 500
    StatusBadGateway          = 502
    StatusServiceUnavailable  = 503
)
```

### Heading Levels
```go
package analyzer

var headingLevels = []string{"h1", "h2", "h3", "h4", "h5", "h6"}
```

## Usage Examples

### Creating an Analysis Request
```go
req := &models.AnalysisRequest{
    URL: "https://example.com",
}

if err := req.Validate(); err != nil {
    // Handle validation error
}
```

### Working with Analysis Results
```go
result := &models.AnalysisResult{
    URL:         "https://example.com",
    HTMLVersion: "HTML5",
    Title:       "Example Domain",
    Headings: map[string]int{
        "h1": 1,
        "h2": 3,
        "h3": 0,
        "h4": 0,
        "h5": 0,
        "h6": 0,
    },
    InternalLinks: 5,
    ExternalLinks: 2,
    InaccessibleLinks: []models.LinkError{
        {
            URL:        "https://broken.com",
            StatusCode: 404,
            Error:      "Not Found",
        },
    },
    HasLoginForm: false,
}

// Access data
fmt.Printf("Page title: %s\n", result.Title)
fmt.Printf("Total headings: %d\n", sumHeadings(result.Headings))
fmt.Printf("Broken links: %d\n", len(result.InaccessibleLinks))
```

### Template Data Population
```go
// Success case
data := &handler.ResultsData{
    Result: result,
}

// Error case
data := &handler.ResultsData{
    Error: "Could not fetch URL: connection timeout",
}
```

### Link Classification
```go
link := models.Link{
    URL:  "https://example.com/about",
    Type: models.LinkTypeInternal,
    Text: "About Us",
}

fmt.Printf("Link type: %s\n", link.Type.String()) // Output: "internal"
```

## JSON Serialization Examples

### AnalysisResult as JSON
```json
{
  "url": "https://example.com",
  "html_version": "HTML5",
  "title": "Example Domain",
  "headings": {
    "h1": 1,
    "h2": 3,
    "h3": 5,
    "h4": 0,
    "h5": 0,
    "h6": 0
  },
  "internal_links": 12,
  "external_links": 3,
  "inaccessible_links": [
    {
      "url": "https://example.com/broken",
      "status_code": 404,
      "error": "Not Found"
    },
    {
      "url": "https://timeout.example.com",
      "status_code": 0,
      "error": "context deadline exceeded"
    }
  ],
  "has_login_form": true
}
```

### ErrorResponse as JSON
```json
{
  "message": "Invalid URL format",
  "status_code": 400,
  "details": "URL scheme must be http or https"
}
```

## Type Safety Notes

1. **Use pointer receivers** for methods that modify structs
2. **Use value receivers** for read-only methods
3. **Validate** all input data before processing
4. **Export** only types that need to be used across packages
5. **Document** all exported types and fields with GoDoc comments

## Future Extensions

### Potential Additional Models (Out of Scope for MVP)

```go
// AnalysisHistory - for storing past analyses
type AnalysisHistory struct {
    ID        string
    URL       string
    Result    *AnalysisResult
    CreatedAt time.Time
}

// AnalysisMetrics - for performance tracking
type AnalysisMetrics struct {
    Duration      time.Duration
    LinksChecked  int
    BytesFetched  int64
    ErrorCount    int
}

// SEOMetrics - for SEO analysis (future feature)
type SEOMetrics struct {
    MetaDescription string
    MetaKeywords    []string
    OpenGraphTags   map[string]string
    ImageAltTags    int
}
```

These models are not needed for the MVP but show how the system can be extended.
