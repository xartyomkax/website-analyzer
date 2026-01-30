# Architecture Specification

## System Design

### Architecture Pattern
**Monolithic Web Application**
- Single binary deployment
- Embedded templates and static assets
- No microservices or separate frontend
- Stateless request processing

### Technology Stack

#### Core
- **Language**: Go 1.24+
- **Debug lib for local development**: github.com/go-delve/delve/cmd/dlv on port 2345
- **HTTP Server**: `net/http` (standard library)
- **Router**: `net/http.ServeMux` (standard library) or simple pattern matching
- **Templates**: `html/template` (standard library)
- **HTML Parser**: `github.com/PuerkitoBio/goquery` (jQuery-like API)
- **Logging**: `log/slog` (structured logging)

#### Infrastructure
- **Containerization**: Docker
- **Build Tool**: Make
- **Version Control**: Git

#### No unnecessary dependencies
- No Gin, Echo, Fiber, or Chi
- No ORM or database (stateless)
- No JavaScript frameworks
- No external dependencies if they are not simplify the code significantly

## Component Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Browser                             │
└────────────────┬────────────────────────────────────────────┘
                 │ HTTP
                 ▼
┌─────────────────────────────────────────────────────────────┐
│                      HTTP Handler Layer                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │ Form Handler │  │Static Handler│  │Health Check  │       │
│  └──────┬───────┘  └──────────────┘  └──────────────┘       │
└─────────┼───────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────┐
│                    Analyzer Service                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │ HTML Parser  │  │ Link Checker │  │  Validator   │       │
│  └──────────────┘  └──────┬───────┘  └──────────────┘       │
└─────────────────────────────┼───────────────────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │  Worker Pool     │
                    │  (Goroutines +   │
                    │   Channels)      │
                    └──────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │  HTTP Client     │
                    │  (net/http)      │
                    └──────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │  External URLs   │
                    └──────────────────┘
```

## Directory Structure

```
webpage-analyzer/
├── cmd/
│   └── main.go                    # Application entry point, server setup
│
├── internal/
│   ├── analyzer/
│   │   ├── analyzer.go            # Main analysis orchestration
│   │   ├── analyzer_test.go
│   │   ├── html.go                # HTML parsing (version, title, headings)
│   │   ├── html_test.go
│   │   ├── links.go               # Link extraction & classification
│   │   ├── links_test.go
│   │   ├── checker.go             # Concurrent link accessibility checking
│   │   └── checker_test.go
│   │
│   ├── handler/
│   │   ├── handler.go             # HTTP request handlers
│   │   ├── handler_test.go
│   │   └── middleware.go          # Logging, recovery middleware
│   │
│   ├── models/
│   │   └── models.go              # Data structures (Request, Result, etc.)
│   │
│   └── validator/
│       ├── validator.go           # URL validation, SSRF protection
│       └── validator_test.go
│
├── web/
│   ├── templates/
│   │   ├── base.html              # Base layout (optional)
│   │   ├── index.html             # Input form page
│   │   └── results.html           # Analysis results page
│   │
│   └── static/
│       └── style.css              # Minimal CSS styling
│
├── Dockerfile
├── Makefile
├── go.mod
├── go.sum
├── .gitignore
├── README.md
└── docs/                          # This documentation
```

## Component Details

### 1. HTTP Handler Layer (`internal/handler/`)

**Responsibilities**:
- Route HTTP requests
- Parse form data
- Validate input
- Call analyzer service
- Render templates
- Handle errors and return appropriate responses

**Key Functions**:
```go
func IndexHandler(w http.ResponseWriter, r *http.Request)
func AnalyzeHandler(w http.ResponseWriter, r *http.Request)
func HealthHandler(w http.ResponseWriter, r *http.Request)
```

**Request Flow**:
1. User submits form with URL
2. Handler validates URL format
3. Handler calls analyzer service
4. Handler renders results template or error page

### 2. Analyzer Service (`internal/analyzer/`)

**Responsibilities**:
- Fetch target URL
- Parse HTML structure
- Extract and analyze content
- Coordinate link checking
- Aggregate results

**Modules**:

#### `analyzer.go` - Main Orchestration
```go
func Analyze(url string) (*models.Result, error)
```
- Coordinates all analysis steps
- Fetches HTML content
- Calls specialized parsers
- Handles top-level errors

#### `html.go` - HTML Parsing
```go
func DetectHTMLVersion(doc *goquery.Document) string
func ExtractTitle(doc *goquery.Document) string
func CountHeadings(doc *goquery.Document) map[string]int
func HasLoginForm(doc *goquery.Document) bool
```

#### `links.go` - Link Processing
```go
func ExtractLinks(doc *goquery.Document, baseURL string) ([]string, error)
func ClassifyLink(link, baseURL string) LinkType
func ResolveRelativeURL(base, href string) (string, error)
```

#### `checker.go` - Concurrent Link Checking
```go
func CheckLinks(links []string, timeout time.Duration) []models.LinkError
```
- Uses worker pool pattern
- Goroutines + channels for concurrency
- Configurable worker count

### 3. Models (`internal/models/`)

**Data Structures**:
```go
type AnalysisRequest struct {
    URL string
}

type AnalysisResult struct {
    URL               string
    HTMLVersion       string
    Title             string
    Headings          map[string]int
    InternalLinks     int
    ExternalLinks     int
    InaccessibleLinks []LinkError
    HasLoginForm      bool
}

type LinkError struct {
    URL        string
    StatusCode int
    Error      string
}
```

### 4. Validator (`internal/validator/`)

**Responsibilities**:
- URL format validation
- SSRF protection (block private IPs)
- Scheme validation (http/https only)
- Input sanitization

**Key Functions**:
```go
func ValidateURL(url string) error
func IsPrivateIP(ip net.IP) bool
func IsSafeURL(url string) error
```

## Concurrency Model

### Worker Pool Pattern for Link Checking

```go
// Conceptual flow
func CheckLinks(links []string, timeout time.Duration) []LinkError {
    // Create channels
    jobs := make(chan string, len(links))
    results := make(chan LinkResult, len(links))
    
    // Spawn workers (10 goroutines)
    for w := 0; w < 10; w++ {
        go worker(jobs, results, timeout)
    }
    
    // Send jobs
    for _, link := range links {
        jobs <- link
    }
    close(jobs)
    
    // Collect results
    var errors []LinkError
    for i := 0; i < len(links); i++ {
        result := <-results
        if result.Error != nil {
            errors = append(errors, result.Error)
        }
    }
    
    return errors
}
```

**Configuration**:
- Worker count: 10-20 (configurable via env var)
- Timeout per link: 5 seconds
- Use `context.WithTimeout` for request deadlines
- HEAD requests to minimize bandwidth

## HTTP Client Configuration

```go
var httpClient = &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        DialContext: (&net.Dialer{
            Timeout:   10 * time.Second,
            KeepAlive: 30 * time.Second,
        }).DialContext,
        MaxIdleConns:        100,
        IdleConnTimeout:     90 * time.Second,
        TLSHandshakeTimeout: 10 * time.Second,
    },
    CheckRedirect: func(req *http.Request, via []*http.Request) error {
        if len(via) >= 10 {
            return fmt.Errorf("too many redirects")
        }
        return nil
    },
}
```

**Key Settings**:
- Custom User-Agent header
- No cookie jar (stateless)
- Resource Limits from REQUIREMENTS.md

## Template System

### Template Hierarchy
```
base.html (optional)
├── index.html (extends base)
└── results.html (extends base)
```

### Template Data Passing
```go
// Index page (form)
type IndexData struct {
    Error string // Optional error message
}

// Results page
type ResultsData struct {
    Result *models.AnalysisResult
    Error  string // Optional error
}
```

## Configuration

### Environment Variables
```bash
# Server
PORT=8080
ENV=production # production, development

# Timeouts
REQUEST_TIMEOUT=30s
LINK_CHECK_TIMEOUT=5s

# Concurrency
MAX_WORKERS=10

# Limits
MAX_RESPONSE_SIZE=10485760  # 10MB
MAX_URL_LENGTH=2048
```

### Configuration Loading
```go
type Config struct {
    Port            string
    Env             string
    RequestTimeout  time.Duration
    LinkTimeout     time.Duration
    MaxWorkers      int
    MaxResponseSize int64
    MaxURLLength    int
}

func LoadConfig() *Config {
    // Load from env vars with defaults
}
```

## Error Handling Strategy

### Error Types
1. **Validation Errors**: Invalid input (400)
2. **Fetch Errors**: Cannot reach URL (502)
3. **Parse Errors**: Malformed HTML (500)
4. **System Errors**: Internal failures (500)

### Error Responses
```go
type ErrorResponse struct {
    Message    string
    StatusCode int
    Details    string // Optional technical details
}
```

### Logging Strategy
```go
// Use structured logging
slog.Info("analyzing URL", 
    "url", targetURL,
    "user_ip", remoteAddr)

slog.Error("fetch failed",
    "url", url,
    "error", err,
    "duration", elapsed)
```

## Security Architecture

### Defense Layers

1. **Input Validation**
   - URL format check
   - Scheme whitelist
   - Length limits

2. **SSRF Protection**
   - IP address filtering
   - DNS resolution checks
   - Private range blocking

3. **Resource Limits**
   - Response size caps
   - Timeout enforcement
   - Connection pooling

4. **Output Sanitization**
   - Template auto-escaping (html/template)
   - No user content in headers

## Deployment Architecture

### Docker Container
```
alpine:latest (minimal base)
├── /app/analyzer (single binary)
└── /app/web/ (templates + static files)
```

**Container Properties**:
- Runs as non-root user
- Exposes port 8080
- No persistent storage
- Stateless (no volumes needed)

### Build Process
```
Multi-stage Docker build:
1. Builder stage: Compile Go binary
2. Runtime stage: Copy binary + assets to minimal Alpine image
```

## Performance Considerations

### Optimization Strategies
- Concurrent link checking (10x+ faster)
- HEAD requests for link validation
- Connection pooling
- Response size limits
- Timeout protection

### Expected Performance
- Simple page (<10 links): <2s
- Medium page (50 links): 5-10s
- Large page (200 links): 10-20s

### Scalability Notes
- Current: Single-instance, synchronous request handling
- Future: Add request queue, horizontal scaling, caching

## Testing Architecture

See [TESTING_STRATEGY.md](../guides/TESTING_STRATEGY.md) for detailed testing approach.

**Test Structure**:
- Unit tests per package
- Table-driven tests
- Mock HTTP servers (httptest)
- Integration tests (handler → analyzer)
- No E2E tests (MVP)

## Future Architecture Considerations

### Potential Enhancements (Out of Scope)
1. **Persistence Layer**: PostgreSQL for analysis history
2. **Caching Layer**: Redis for URL results
3. **Queue System**: RabbitMQ/Redis for async processing
4. **API Layer**: REST API alongside web UI
5. **Metrics**: Prometheus instrumentation
6. **Tracing**: OpenTelemetry integration

Current architecture supports these additions without major refactoring.
