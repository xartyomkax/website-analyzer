# Task 01: Project Setup

## Objective
Initialize the Go project structure with all necessary files, dependencies, and build tooling.

## Prerequisites
- Git installed
- Docker installed (for containerization)
- Make installed (for build automation)

## Steps

### 1. Install Dependencies
Such as goquery

### 2. Create .gitignore & .editorconfig files
Standard for simple golang project

### 3. Create Makefile
see ARCHITECTURE.md, include targets for build, test, run, clean, fmt, tidy, docker-build, docker-run, docker-clean

### 4. Create Dockerfile & .dockerignore
see ARCHITECTURE.md, multi-stage build for dev with debug and minimal production image

### 5. Create Basic README.md
Describe the project, its features, and how to run it.

## Configuration
see ARCHITECTURE.md, set via environment variables

## Testing
```bash
# Run all tests
make test

# With coverage
make test-coverage
```

## License
MIT
```

### 9. Create Placeholder Files

**cmd/main.go**
```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Web Page Analyzer - Coming Soon")
	})

	addr := ":" + port
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
```

**internal/models/models.go**
```go
package models

// Placeholder - will be populated in later tasks
```

**web/templates/index.html**
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Web Page Analyzer</title>
</head>
<body>
    <h1>Web Page Analyzer</h1>
    <p>Coming soon...</p>
</body>
</html>
```

**web/static/style.css**
```css
/* Placeholder styles */
body {
    font-family: system-ui, -apple-system, sans-serif;
    max-width: 800px;
    margin: 0 auto;
    padding: 2rem;
}
```

### 10. Verify Setup
```bash
# Format code
make fmt

# Tidy dependencies
make tidy

# Run tests (should pass with no tests yet)
make test

# Build application
make build

# Run application
make run
```

Visit http://localhost:8080 - should see "Web Page Analyzer - Coming Soon"

### 11. Initial Git Commit
```bash
git add .
git commit -m "Initial project setup"
```

## Acceptance Criteria
- ✅ Go module initialized
- ✅ Directory structure created
- ✅ Dependencies installed (goquery)
- ✅ Makefile with all targets working
- ✅ Dockerfile builds successfully
- ✅ Basic HTTP server runs on port 8080
- ✅ .gitignore configured
- ✅ README.md created
- ✅ All files committed to Git

## Verification Commands
```bash
# Check Go version
go version

# List dependencies
go list -m all

# Test Makefile targets
make help
make build
make test
make clean

# Test Docker build
make docker-build

# Verify file structure
tree -L 3 -I 'bin|*.out|*.html'
```

## Next Steps
Once this task is complete, proceed to:
- [Task 02: HTML Parser](02_HTML_PARSER.md)

## Troubleshooting

### Port already in use
```bash
# Find process using port 8080
lsof -i :8080

# Kill process
kill -9 <PID>
```

### Docker build fails
```bash
# Clean Docker cache
docker builder prune

# Rebuild without cache
docker build --no-cache -t webpage-analyzer .
```

### Go module issues
```bash
# Clear module cache
go clean -modcache

# Re-download dependencies
go mod download
```

## Related Documentation
- [ARCHITECTURE.md](../specs/ARCHITECTURE.md) - System architecture
- [Go Best Practices](../guides/GO_BEST_PRACTICES.md) - Coding standards
