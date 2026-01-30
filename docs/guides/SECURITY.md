# Security Guide

## Overview
This document outlines security considerations and best practices for the Web Page Analyzer application.

## Critical Security Concerns

### 1. SSRF (Server-Side Request Forgery)

**Risk**: Users could make the server request internal resources or cloud metadata.

**Mitigation**:

```go
package validator

import (
    "fmt"
    "net"
)

// Block private IP ranges
func isPrivateIP(ip net.IP) bool {
    privateRanges := []string{
        "10.0.0.0/8",         // Private network
        "172.16.0.0/12",      // Private network
        "192.168.0.0/16",     // Private network
        "127.0.0.0/8",        // Loopback
        "169.254.0.0/16",     // Link-local
        "::1/128",            // IPv6 loopback
        "fe80::/10",          // IPv6 link-local
        "fc00::/7",           // IPv6 private
    }

    for _, cidr := range privateRanges {
        _, network, _ := net.ParseCIDR(cidr)
        if network.Contains(ip) {
            return true
        }
    }

    return false
}

// Validate URL before fetching
func ValidateURL(rawURL string) error {
    parsed, err := url.Parse(rawURL)
    if err != nil {
        return fmt.Errorf("invalid URL: %w", err)
    }

    // Check scheme
    if parsed.Scheme != "http" && parsed.Scheme != "https" {
        return fmt.Errorf("only http and https schemes allowed")
    }

    // Resolve hostname
    ips, err := net.LookupIP(parsed.Hostname())
    if err != nil {
        return fmt.Errorf("could not resolve hostname: %w", err)
    }

    // Check if any IP is private
    for _, ip := range ips {
        if isPrivateIP(ip) {
            return fmt.Errorf("access to private IP addresses not allowed")
        }
    }

    return nil
}
```

**Blocked Examples**:
- `http://localhost/admin`
- `http://127.0.0.1/`
- `http://192.168.1.1/`
- `http://169.254.169.254/latest/meta-data/` (AWS metadata)
- `http://[::1]/`

### 2. Input Validation

**Always validate user input**:

```go
func ValidateAnalysisRequest(req *models.AnalysisRequest) error {
    // URL required
    if req.URL == "" {
        return errors.New("URL is required")
    }

    // Length limit
    if len(req.URL) > 2048 {
        return errors.New("URL too long (max 2048 characters)")
    }

    // Format validation
    if err := ValidateURL(req.URL); err != nil {
        return err
    }

    return nil
}
```

### 3. Resource Limits

**Prevent DoS attacks**:

```go
// Limit response size
const MaxResponseSize = 10 * 1024 * 1024 // 10MB

func fetchHTML(url string) (*goquery.Document, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // Limit reading
    limitedReader := io.LimitReader(resp.Body, MaxResponseSize)
    
    doc, err := goquery.NewDocumentFromReader(limitedReader)
    if err != nil {
        return nil, err
    }

    return doc, nil
}
```

**Timeout Configuration**:
```go
client := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        DialContext: (&net.Dialer{
            Timeout:   10 * time.Second,
        }).DialContext,
        TLSHandshakeTimeout:   10 * time.Second,
        ResponseHeaderTimeout: 10 * time.Second,
    },
}
```

### 4. Template Safety

**Use html/template (auto-escapes)**:

```go
import "html/template"

// Safe - auto-escapes user input
tmpl.Execute(w, data)
```

**Never use text/template for HTML**:
```go
// UNSAFE - no escaping
import "text/template"  // Don't use for HTML!
```

### 5. Header Security

**Set security headers**:

```go
func securityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Content-Security-Policy", "default-src 'self'")
        
        next.ServeHTTP(w, r)
    })
}
```

### 6. Rate Limiting (Future Enhancement)

**Basic implementation**:

```go
import "golang.org/x/time/rate"

var limiter = rate.NewLimiter(rate.Limit(10), 20) // 10 req/s, burst of 20

func rateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

## Security Checklist

### Input Validation
- [ ] URL scheme whitelist (http/https only)
- [ ] URL length limit (2048 chars)
- [ ] Hostname resolution check
- [ ] Private IP blocking
- [ ] Protocol validation

### Resource Protection
- [ ] HTTP client timeout configured
- [ ] Response size limits
- [ ] Connection timeout
- [ ] TLS handshake timeout
- [ ] Maximum redirects limit

### Output Safety
- [ ] Use html/template for rendering
- [ ] No user input in HTTP headers
- [ ] Proper error messages (no stack traces)
- [ ] Content-Type headers set

### HTTP Security
- [ ] Security headers configured
- [ ] HTTPS enforced (in production)
- [ ] No sensitive data in logs
- [ ] Secure cookie settings (if added)

### Dependencies
- [ ] Regular dependency updates
- [ ] Vulnerability scanning
- [ ] Minimal dependencies
- [ ] Trusted sources only

## Attack Scenarios & Defenses

### Scenario 1: SSRF to AWS Metadata
**Attack**: `http://169.254.169.254/latest/meta-data/`  
**Defense**: IP validation blocks 169.254.0.0/16

### Scenario 2: SSRF to Internal Services
**Attack**: `http://192.168.1.100:8080/admin`  
**Defense**: Private IP ranges blocked

### Scenario 3: DNS Rebinding
**Attack**: Domain resolves to public IP initially, then to private IP  
**Defense**: Validate IPs at request time, not just domain lookup

### Scenario 4: Large Response DoS
**Attack**: URL returns 1GB response  
**Defense**: io.LimitReader with 10MB limit

### Scenario 5: Slow Loris Attack
**Attack**: Slow response to tie up connections  
**Defense**: HTTP client timeouts

### Scenario 6: Redirect Loop
**Attack**: URL redirects infinitely  
**Defense**: Limit redirects in HTTP client

## Logging Security

**Safe logging**:
```go
// Good - no sensitive data
slog.Info("analysis started",
    "url", sanitizeURL(targetURL),
    "client_ip", anonymizeIP(clientIP))

// Bad - sensitive data exposed
slog.Info("analysis started",
    "url", targetURL,
    "session_token", token)  // Don't log!
```

**Sanitize URLs**:
```go
func sanitizeURL(rawURL string) string {
    parsed, err := url.Parse(rawURL)
    if err != nil {
        return "[invalid]"
    }

    // Remove query parameters and fragments
    parsed.RawQuery = ""
    parsed.Fragment = ""
    
    return parsed.String()
}
```

## Error Handling

**Don't expose internals**:

```go
// Bad - reveals internal structure
if err != nil {
    return fmt.Errorf("database query failed: %v", err)
}

// Good - user-friendly message
if err != nil {
    slog.Error("database error", "error", err)
    return errors.New("an internal error occurred")
}
```

## Production Hardening

### Environment Variables
```bash
# Never commit these
export API_KEY=secret123

# Use .env file locally (gitignored)
# Use secrets manager in production
```

### Docker Security
```dockerfile
# Run as non-root user
FROM alpine:latest
RUN addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser appuser
USER appuser

# Read-only filesystem
COPY --chown=appuser:appuser app /app
```

### TLS Configuration
```go
// Force HTTPS in production
if os.Getenv("ENV") == "production" {
    cfg := &tls.Config{
        MinVersion: tls.VersionTLS12,
        CipherSuites: []uint16{
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
        },
    }
    
    server := &http.Server{
        Addr:      ":443",
        Handler:   handler,
        TLSConfig: cfg,
    }
    
    server.ListenAndServeTLS("cert.pem", "key.pem")
}
```

## Testing Security

### SSRF Test Cases
```go
func TestSSRFPrevention(t *testing.T) {
    tests := []string{
        "http://127.0.0.1",
        "http://localhost",
        "http://192.168.1.1",
        "http://10.0.0.1",
        "http://169.254.169.254",
        "http://[::1]",
    }

    for _, url := range tests {
        err := ValidateURL(url)
        if err == nil {
            t.Errorf("Expected error for %s", url)
        }
    }
}
```

## Security Monitoring

**Metrics to track**:
- Failed validation attempts
- Blocked IP addresses
- Response size violations
- Timeout occurrences
- Error rates

**Alerting**:
```go
if blockedAttempts > threshold {
    slog.Warn("possible attack detected",
        "blocked_count", blockedAttempts,
        "source_ip", clientIP)
}
```

## Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CWE-918: SSRF](https://cwe.mitre.org/data/definitions/918.html)
- [Go Security Checklist](https://github.com/Checkmarx/Go-SCP)
- [Cloud Metadata Endpoints](https://gist.github.com/jhaddix/78cece26c91c6263653f31ba453e273b)

## Security Updates

Stay informed:
- Subscribe to Go security announcements
- Monitor dependency vulnerabilities
- Regular security audits
- Penetration testing (for production)
