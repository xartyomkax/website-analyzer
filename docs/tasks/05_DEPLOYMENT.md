# Task 05: Deployment

## Objective
Finalize Docker configuration, documentation, and prepare for production deployment.

## Steps

### 1. Verify Dockerfile

Ensure your Dockerfile is optimized:

```dockerfile
# Multi-stage build
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o analyzer cmd/main.go

# Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/analyzer .
COPY --from=builder /app/web ./web

EXPOSE 8080

CMD ["./analyzer"]
```

### 2. Update README

**File: `README.md`**

Add deployment instructions, usage examples, and screenshots.

### 3. Create .dockerignore

```
.git
.gitignore
README.md
docs/
bin/
*.out
*.test
coverage.html
.DS_Store
```

### 4. Build and Test

```bash
# Build Docker image
make docker-build

# Run container
make docker-run

# Test in browser
open http://localhost:8080

# Test with real URLs
# Try: https://example.com
# Try: https://golang.org
```

### 5. Final Verification

```bash
# Run all tests
make test

# Check coverage
make test-coverage

# Format code
make fmt

# Build binary
make build

# Run binary
./bin/analyzer
```

## Acceptance Criteria

- ✅ Docker image builds successfully
- ✅ Docker container runs without errors
- ✅ Application accessible on http://localhost:8080
- ✅ All tests pass
- ✅ Documentation complete
- ✅ Code properly formatted
- ✅ No security vulnerabilities

## Production Checklist

- [ ] Environment variables documented
- [ ] Error handling comprehensive
- [ ] Logging configured
- [ ] Resource limits set (Docker)
- [ ] Health check endpoint (optional)
- [ ] Monitoring setup (optional)

## Deployment Options

### Local Docker
```bash
docker run -p 8080:8080 webpage-analyzer
```

### Docker Compose
```bash
docker-compose up -d
```

### Cloud Platforms

**Google Cloud Run**:
```bash
gcloud run deploy webpage-analyzer --source .
```

**AWS ECS/Fargate**:
```bash
# Push to ECR
# Create ECS task definition
# Deploy to Fargate
```

**Heroku**:
```bash
heroku container:push web
heroku container:release web
```

## Next Steps

Project complete! 🎉

Consider adding:
- CI/CD pipeline (GitHub Actions)
- Metrics/monitoring (Prometheus)
- Rate limiting
- Caching layer
- API endpoints
