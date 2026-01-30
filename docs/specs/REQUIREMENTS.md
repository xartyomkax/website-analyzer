# Requirements Specification

## Objective
Build a web application that analyzes web pages and provides detailed HTML structure and link analysis.

## Functional Requirements

### 1. User Interface
**Input Form**
- Text field for URL input
- Submit button to trigger analysis
- Clear error messages for invalid input
- Form should process request using POST method

**Results Display**
- Show all analysis results on a single page
- Clean, readable formatting
- No JavaScript required (HTML + CSS only)

### 2. Analysis Features

#### 2.1 HTML Version Detection
- Parse DOCTYPE declaration
- Identify HTML version (HTML5, HTML4.01, XHTML, etc.)
- Default to "HTML5" if DOCTYPE is missing or invalid
- Handle malformed DOCTYPEs gracefully

**Examples**:
- `<!DOCTYPE html>` → HTML5
- `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN">` → HTML4.01
- No DOCTYPE → HTML5 (default)

#### 2.2 Page Title
- Extract content from `<title>` tag
- Display "No title" if missing
- Handle multiple `<title>` tags (use first)
- Trim whitespace

#### 2.3 Heading Analysis
- Count headings by level: H1, H2, H3, H4, H5, H6
- Display counts per level
- Ignore zero counts

**Output Format**:
```
H1: 1
H2: 3
H3: 5
```

#### 2.4 Link Analysis

**Scope** (MVP):
- Only analyze `<a href="">` tags
- Ignore images, CSS, JS files (future enhancement)

**Categorization**:
- **Internal Links**: Same domain/subdomain
  - Examples: `/about`, `./contact.html`, `https://example.com/page`
- **External Links**: Different domain
  - Examples: `https://google.com`, `http://other-site.com`

**Accessibility Check**:
- Verify each link is reachable via HTTP request
- Track inaccessible links with:
  - URL
  - HTTP status code (if available)
  - Error message (timeout, network error, etc.)
- Use HEAD requests to minimize bandwidth
- Timeout: 5 seconds per link
- Concurrent checking with goroutines

**Link Resolution**:
- Resolve relative URLs to absolute
- Handle protocol-relative URLs (`//example.com`)
- Skip invalid URLs (empty, `javascript:`, `mailto:`, `#anchors`)

#### 2.5 Login Form Detection
- Search for `<form>` elements containing `<input type="password">`
- Return boolean: Yes/No
- Handle multiple forms (any form with password = Yes)

### 3. Error Handling

#### Unreachable URLs
When the target URL cannot be fetched, display:
- HTTP status code (if available)
- Descriptive error message
- User-friendly explanation

**Example Error Messages**:
```
Error: Could not reach https://example.com
Status Code: 404
Description: The requested page was not found.

Error: Could not reach https://example.com
Status Code: N/A
Description: Connection timeout after 30 seconds.
```

#### Supported Error Scenarios
- HTTP errors (404, 500, 403, etc.)
- Network timeouts
- DNS resolution failures
- Connection refused
- SSL/TLS errors
- Malformed URLs

### 4. Security Requirements

> **Note**: This section serves as the **Single Source of Truth** for all application limits and configuration defaults.

#### SSRF Prevention
Block requests to:
- Private IP ranges: `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`
- Localhost: `127.0.0.0/8`, `::1`
- Link-local: `169.254.0.0/16`
- Cloud metadata endpoints: `169.254.169.254`

#### Input Validation
- URL scheme whitelist: `http://`, `https://` only
- Maximum URL length: 2048 characters
- URL format validation (valid hostname, etc.)

#### Resource Limits
- Maximum response size: 10MB
- Request timeout: 30 seconds (main page)
- Link check timeout: 5 seconds per link
- Maximum redirects: 10

## Non-Functional Requirements

### Performance
- Concurrent link checking (10-20 workers)
- Efficient HTML parsing
- Responsive user interface (<1s for small pages)

### Reliability
- Graceful error handling (no crashes)
- Partial results on link check failures
- Timeout protection

### Maintainability
- Clean code structure
- 80%+ test coverage
- Comprehensive error logging

### Scalability (Future)
- Currently: Single request processing
- Future: Request queuing, result caching

## Constraints

### Technical Constraints
1. **Language**: Must be written in Golang
2. **Version Control**: Must use Git
3. **Dependencies**: Minimize external dependencies
4. **Frontend**: HTML + CSS only (no JavaScript)
5. **Parsing**: Static HTML only (no JavaScript execution)

### Out of Scope (MVP)
- JavaScript-heavy site rendering (SPA support)
- Authentication/user accounts
- Analysis history/persistence
- API endpoints (only web UI)
- Real-time updates/WebSockets
- Image/CSS/JS file analysis (separate from links)
- Performance metrics
- SEO analysis
- Accessibility scoring

## User Stories

### Story 1: Analyze a Blog Post
```
As a content creator
I want to analyze my blog post HTML
So that I can verify proper heading structure and check for broken links
```

**Acceptance Criteria**:
- Can input blog URL
- See heading count by level
- Identify broken internal/external links
- Verify login form presence (if applicable)

### Story 2: Debug Website Issues
```
As a web developer
I want to check if my website has inaccessible links
So that I can fix them before deployment
```

**Acceptance Criteria**:
- List all inaccessible links with error details
- Differentiate internal vs external links
- See HTTP status codes for failures

### Story 3: Validate HTML Structure
```
As a QA engineer
I want to verify HTML version and structure
So that I can ensure standards compliance
```

**Acceptance Criteria**:
- Detect HTML version from DOCTYPE
- Count all heading levels
- Verify page has a title

## Success Metrics
- ✅ All analysis features working
- ✅ Error handling for all edge cases
- ✅ Security validation prevents SSRF
- ✅ Concurrent link checking completes in reasonable time
- ✅ User-friendly error messages
- ✅ Clean, semantic HTML output
