# Web Page Analyzer

A lightweight Go web application that analyzes web pages for HTML structure, headings, links, and login forms. Built with Go standard library, focusing on simplicity, testability, and concurrent processing.

## Features

- **HTML Version Detection** - Identifies HTML version (HTML5, XHTML, HTML 4.01, etc.)
- **Title Extraction** - Extracts page title
- **Heading Analysis** - Counts all heading levels (H1-H6)
- **Login Form Detection** - Identifies password input fields
- **Link Extraction** - Extracts all links with internal/external classification
- **Concurrent Link Checking** - Validates link accessibility using goroutines
- **SSRF Protection** - Blocks requests to private IP ranges

## Tech Stack

- **Language**: Go 1.24+
- **HTTP Server**: `net/http` (standard library)
- **HTML Parser**: `github.com/PuerkitoBio/goquery`
- **Templates**: `html/template` (standard library)
- **Logging**: `log/slog` (structured logging)
- **Deployment**: Docker

## Quick Start

### Prerequisites

- Go 1.24 or higher
- Docker (optional, for containerized deployment)
- Make (optional, for build automation)

### Local Development

```bash
# Clone the repository
git clone https://website-analyzer
cd webpage-analyzer

# Install dependencies
go mod download

# Run tests
make test

# Build the application
make build

# Run the application
make run
```

The application will start on `http://localhost:8080`

### Using Docker

```bash
# Build Docker image
make docker-build

# Run container
make docker-run
```

### Debug Mode

```bash
# Run with delve debugger (local)
make debug

# Or using Docker
make docker-build-debug
make docker-run-debug
```

Debugger will be available on port 2345 for remote debugging.

## Configuration

Configuration is managed through environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `ENV` | `production` | Environment (production/development) |
| `REQUEST_TIMEOUT` | `30s` | Timeout for fetching target URLs |
| `LINK_CHECK_TIMEOUT` | `5s` | Timeout for checking individual links |
| `MAX_WORKERS` | `10` | Number of concurrent workers for link checking |
| `MAX_RESPONSE_SIZE` | `10485760` | Maximum response size (10MB) |
| `MAX_URL_LENGTH` | `2048` | Maximum URL length |

### Example

```bash
# Set custom port
export PORT=3000

# Run with custom configuration
make run
```

## Usage

1. Open your browser and navigate to `http://localhost:8080`
2. Enter a URL to analyze
3. Click "Analyze"
4. View the analysis results including:
   - HTML version
   - Page title
   - Heading counts (H1-H6)
   - Internal/external link counts
   - Inaccessible links
   - Login form detection

## Project Structure

```
webpage-analyzer/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── analyzer/              # HTML parsing and analysis logic
│   ├── handler/               # HTTP request handlers
│   ├── models/                # Data structures
│   └── validator/             # URL validation and SSRF protection
├── web/
│   ├── templates/             # HTML templates
│   └── static/                # CSS and static assets
├── Dockerfile                 # Multi-stage Docker build
├── Makefile                   # Build automation
├── go.mod                     # Go module definition
└── README.md                  # This file
```

## Development

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run go vet
make vet
```

### Code Formatting

```bash
# Format code
make fmt

# Tidy dependencies
make tidy
```

### Build Commands

```bash
# Build binary
make build

# Clean build artifacts
make clean

# Run all checks (fmt, tidy, test, build)
make all
```

## Security

This application implements several security measures:

- **SSRF Protection**: Blocks requests to private IP ranges (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, 127.0.0.0/8)
- **Input Validation**: URL format and scheme validation (http/https only)
- **Resource Limits**: Response size caps and timeout enforcement
- **Output Sanitization**: Automatic HTML escaping via `html/template`

## Performance

- **Concurrent Link Checking**: Uses goroutines and channels for 10x+ faster link validation
- **Connection Pooling**: Reuses HTTP connections for better performance
- **Timeouts**: Prevents hanging on slow or unresponsive URLs

Expected performance:
- Simple page (<10 links): <2s
- Medium page (50 links): 5-10s
- Large page (200 links): 10-20s

## Testing

The project maintains 80%+ test coverage with:
- Unit tests for all packages
- Table-driven tests
- Mock HTTP servers using `httptest`
- Integration tests

## Docker

### Production Build

```bash
docker build -t webpage-analyzer:latest .
docker run -p 8080:8080 webpage-analyzer:latest
```

### Debug Build

```bash
docker build --target debug -t webpage-analyzer:debug .
docker run -p 8080:8080 -p 2345:2345 webpage-analyzer:debug
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Format code: `make fmt`
6. Submit a pull request

## License

MIT License - see LICENSE file for details

## Documentation

For more detailed documentation, see:
- [Architecture](docs/specs/ARCHITECTURE.md) - System design and tech stack
- [Requirements](docs/specs/REQUIREMENTS.md) - Feature specifications
- [Testing Strategy](docs/guides/TESTING_STRATEGY.md) - Testing approach
- [Security](docs/guides/SECURITY.md) - Security considerations
- [Go Best Practices](docs/guides/GO_BEST_PRACTICES.md) - Coding standards

## Support

For issues, questions, or contributions, please open an issue on GitHub.
