# Task 04: HTTP Handlers

## Objective
Implement web server with HTTP handlers, templates, and complete the end-to-end analysis flow.

## Prerequisites
- Tasks 01-03 completed
- All analyzer functions implemented and tested

## Implementation Steps

### 1. Create Main Analyzer Service

**File: `internal/analyzer/analyzer.go`**

```go
package analyzer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"webpage-analyzer/internal/models"
	"webpage-analyzer/internal/validator"

	"github.com/PuerkitoBio/goquery"
)

type Config struct {
	RequestTimeout  time.Duration
	LinkTimeout     time.Duration
	MaxWorkers      int
	MaxResponseSize int64
}

type Analyzer struct {
	config     *Config
	httpClient *http.Client
}

func NewAnalyzer(config *Config) *Analyzer {
	return &Analyzer{
		config: config,
		httpClient: &http.Client{
			Timeout: config.RequestTimeout,
		},
	}
}

func (a *Analyzer) Analyze(targetURL string) (*models.AnalysisResult, error) {
	// Validate URL
	if err := validator.ValidateURL(targetURL); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Fetch HTML
	doc, err := a.fetchHTML(targetURL)
	if err != nil {
		return nil, err
	}

	// Extract links
	links, err := ExtractLinks(doc, targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to extract links: %w", err)
	}

	// Count internal/external
	var internal, external int
	for _, link := range links {
		if link.Type == models.LinkTypeInternal {
			internal++
		} else if link.Type == models.LinkTypeExternal {
			external++
		}
	}

	// Check link accessibility
	checkConfig := CheckLinksConfig{
		Timeout:    a.config.LinkTimeout,
		MaxWorkers: a.config.MaxWorkers,
	}
	inaccessible := CheckLinks(links, checkConfig)

	// Build result
	result := &models.AnalysisResult{
		URL:               targetURL,
		HTMLVersion:       DetectHTMLVersion(doc),
		Title:             ExtractTitle(doc),
		Headings:          CountHeadings(doc),
		InternalLinks:     internal,
		ExternalLinks:     external,
		InaccessibleLinks: inaccessible,
		HasLoginForm:      HasLoginForm(doc),
	}

	return result, nil
}

func (a *Analyzer) fetchHTML(url string) (*goquery.Document, error) {
	ctx, cancel := context.WithTimeout(context.Background(), a.config.RequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "WebPageAnalyzer/1.0")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	// Limit response size
	limitedReader := io.LimitReader(resp.Body, a.config.MaxResponseSize)

	doc, err := goquery.NewDocumentFromReader(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	return doc, nil
}
```

### 2. Create Validator

**File: `internal/validator/validator.go`**

```go
package validator

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

func ValidateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("URL is required")
	}

	if len(rawURL) > 2048 {
		return fmt.Errorf("URL too long (max 2048 characters)")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check scheme
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https")
	}

	// Check host
	if parsed.Host == "" {
		return fmt.Errorf("URL must have a host")
	}

	// SSRF protection
	if err := checkSSRF(parsed.Hostname()); err != nil {
		return err
	}

	return nil
}

func checkSSRF(hostname string) error {
	// Resolve hostname
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return fmt.Errorf("could not resolve hostname: %w", err)
	}

	for _, ip := range ips {
		if isPrivateIP(ip) {
			return fmt.Errorf("access to private IP addresses is not allowed")
		}
	}

	return nil
}

func isPrivateIP(ip net.IP) bool {
	// Check for private ranges
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16", // link-local
		"::1/128",        // IPv6 localhost
		"fe80::/10",      // IPv6 link-local
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

### 3. Create HTTP Handlers

**File: `internal/handler/handler.go`**

```go
package handler

import (
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"webpage-analyzer/internal/analyzer"
	"webpage-analyzer/internal/models"
)

type Handler struct {
	analyzer  *analyzer.Analyzer
	templates *template.Template
}

func NewHandler(analyzer *analyzer.Analyzer, templatesPath string) (*Handler, error) {
	tmpl, err := template.ParseGlob(templatesPath + "/*.html")
	if err != nil {
		return nil, err
	}

	return &Handler{
		analyzer:  analyzer,
		templates: tmpl,
	}, nil
}

func (h *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := struct {
		Error string
	}{}

	if err := h.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		slog.Error("template error", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		h.renderError(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	targetURL := r.FormValue("url")

	// Analyze
	start := time.Now()
	result, err := h.analyzer.Analyze(targetURL)
	duration := time.Since(start)

	slog.Info("analysis completed",
		"url", targetURL,
		"duration", duration,
		"error", err)

	if err != nil {
		h.renderError(w, err.Error(), http.StatusBadGateway)
		return
	}

	// Render results
	h.renderResults(w, result)
}

func (h *Handler) renderResults(w http.ResponseWriter, result *models.AnalysisResult) {
	data := struct {
		Result *models.AnalysisResult
	}{
		Result: result,
	}

	if err := h.templates.ExecuteTemplate(w, "results.html", data); err != nil {
		slog.Error("template error", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) renderError(w http.ResponseWriter, errMsg string, statusCode int) {
	data := struct {
		Error      string
		StatusCode int
	}{
		Error:      errMsg,
		StatusCode: statusCode,
	}

	w.WriteHeader(statusCode)
	if err := h.templates.ExecuteTemplate(w, "error.html", data); err != nil {
		slog.Error("template error", "error", err)
		http.Error(w, errMsg, statusCode)
	}
}
```

### 4. Create Templates

**File: `web/templates/index.html`**

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Web Page Analyzer</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <h1>Web Page Analyzer</h1>
        <p>Enter a URL to analyze its HTML structure and links.</p>
        
        <form method="POST" action="/analyze">
            <div class="form-group">
                <label for="url">Website URL:</label>
                <input 
                    type="url" 
                    id="url" 
                    name="url" 
                    placeholder="https://example.com" 
                    required
                    autofocus
                >
            </div>
            <button type="submit">Analyze</button>
        </form>
    </div>
</body>
</html>
```

**File: `web/templates/results.html`**

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Analysis Results - Web Page Analyzer</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <h1>Analysis Results</h1>
        
        <div class="result-section">
            <h2>Page Information</h2>
            <table>
                <tr>
                    <th>URL:</th>
                    <td>{{.Result.URL}}</td>
                </tr>
                <tr>
                    <th>HTML Version:</th>
                    <td>{{.Result.HTMLVersion}}</td>
                </tr>
                <tr>
                    <th>Title:</th>
                    <td>{{.Result.Title}}</td>
                </tr>
                <tr>
                    <th>Login Form:</th>
                    <td>{{if .Result.HasLoginForm}}Yes{{else}}No{{end}}</td>
                </tr>
            </table>
        </div>

        <div class="result-section">
            <h2>Headings</h2>
            <table>
                <tr><th>H1:</th><td>{{index .Result.Headings "h1"}}</td></tr>
                <tr><th>H2:</th><td>{{index .Result.Headings "h2"}}</td></tr>
                <tr><th>H3:</th><td>{{index .Result.Headings "h3"}}</td></tr>
                <tr><th>H4:</th><td>{{index .Result.Headings "h4"}}</td></tr>
                <tr><th>H5:</th><td>{{index .Result.Headings "h5"}}</td></tr>
                <tr><th>H6:</th><td>{{index .Result.Headings "h6"}}</td></tr>
            </table>
        </div>

        <div class="result-section">
            <h2>Links</h2>
            <table>
                <tr>
                    <th>Internal Links:</th>
                    <td>{{.Result.InternalLinks}}</td>
                </tr>
                <tr>
                    <th>External Links:</th>
                    <td>{{.Result.ExternalLinks}}</td>
                </tr>
                <tr>
                    <th>Inaccessible Links:</th>
                    <td>{{len .Result.InaccessibleLinks}}</td>
                </tr>
            </table>
        </div>

        {{if .Result.InaccessibleLinks}}
        <div class="result-section">
            <h2>Inaccessible Links</h2>
            <table>
                <thead>
                    <tr>
                        <th>URL</th>
                        <th>Status</th>
                        <th>Error</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Result.InaccessibleLinks}}
                    <tr>
                        <td>{{.URL}}</td>
                        <td>{{if .StatusCode}}{{.StatusCode}}{{else}}N/A{{end}}</td>
                        <td>{{.Error}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>
        {{end}}

        <div class="actions">
            <a href="/" class="button">Analyze Another Page</a>
        </div>
    </div>
</body>
</html>
```

**File: `web/templates/error.html`**

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Error - Web Page Analyzer</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <h1>Error</h1>
        <div class="error">
            <p><strong>Status Code:</strong> {{.StatusCode}}</p>
            <p><strong>Message:</strong> {{.Error}}</p>
        </div>
        <div class="actions">
            <a href="/" class="button">Go Back</a>
        </div>
    </div>
</body>
</html>
```

### 5. Update CSS

**File: `web/static/style.css`**

```css
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    line-height: 1.6;
    color: #333;
    background: #f5f5f5;
}

.container {
    max-width: 900px;
    margin: 2rem auto;
    padding: 2rem;
    background: white;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

h1 {
    color: #2c3e50;
    margin-bottom: 1rem;
}

h2 {
    color: #34495e;
    margin-top: 2rem;
    margin-bottom: 1rem;
    border-bottom: 2px solid #3498db;
    padding-bottom: 0.5rem;
}

.form-group {
    margin-bottom: 1.5rem;
}

label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 500;
}

input[type="url"] {
    width: 100%;
    padding: 0.75rem;
    border: 2px solid #ddd;
    border-radius: 4px;
    font-size: 1rem;
}

input[type="url"]:focus {
    outline: none;
    border-color: #3498db;
}

button, .button {
    display: inline-block;
    padding: 0.75rem 1.5rem;
    background: #3498db;
    color: white;
    border: none;
    border-radius: 4px;
    font-size: 1rem;
    cursor: pointer;
    text-decoration: none;
}

button:hover, .button:hover {
    background: #2980b9;
}

.result-section {
    margin-bottom: 2rem;
}

table {
    width: 100%;
    border-collapse: collapse;
    margin-top: 1rem;
}

th, td {
    padding: 0.75rem;
    text-align: left;
    border-bottom: 1px solid #ddd;
}

th {
    font-weight: 600;
    color: #2c3e50;
    width: 30%;
}

.error {
    background: #fee;
    border-left: 4px solid #e74c3c;
    padding: 1rem;
    margin: 1rem 0;
}

.actions {
    margin-top: 2rem;
    text-align: center;
}
```

### 6. Update main.go

**File: `cmd/main.go`**

```go
package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"webpage-analyzer/internal/analyzer"
	"webpage-analyzer/internal/handler"
)

func main() {
	// Configure logging
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// Configuration
	config := &analyzer.Config{
		RequestTimeout:  30 * time.Second,
		LinkTimeout:     5 * time.Second,
		MaxWorkers:      10,
		MaxResponseSize: 10 * 1024 * 1024, // 10MB
	}

	// Create analyzer
	analyzer := analyzer.NewAnalyzer(config)

	// Create handler
	h, err := handler.NewHandler(analyzer, "web/templates")
	if err != nil {
		log.Fatal("Failed to load templates:", err)
	}

	// Routes
	http.HandleFunc("/", h.IndexHandler)
	http.HandleFunc("/analyze", h.AnalyzeHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	slog.Info("server starting", "addr", addr)
	
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
```

## Acceptance Criteria

- ✅ Complete end-to-end flow works
- ✅ Form submits to `/analyze` endpoint
- ✅ Results display all analysis data
- ✅ Error handling shows user-friendly messages
- ✅ Templates render correctly
- ✅ Static files served
- ✅ No crashes on invalid input

## Testing

```bash
# Run application
make run

# Test in browser
open http://localhost:8080

# Test with curl
curl -X POST http://localhost:8080/analyze -d "url=https://example.com"
```

## Next Steps
- [Task 05: E2E Tests](05_E2E_TESTS.md)
