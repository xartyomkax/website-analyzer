# Web Page Analyzer - AI Agent Guide

## Project Overview
A lightweight Go web application that analyzes web pages for HTML structure, headings, links, and login forms. Built with Go standard library, focusing on simplicity, testability, and concurrent processing.

## Quick Start

```bash
# Development
make test          # Run all tests
make run           # Start dev server
make test-coverage # Generate coverage report

# Production
make docker-build  # Build Docker image
make docker-run    # Run in container
```

## Progress Tracking

### Files
- `docs/BUILD_PROGRESS.md` - Current status (read every session, keep under 100 lines)
- `docs/BUILD_LOG.md` - Full debug history (append-only, read when debugging)
- `docs/learnings/` - Task-specific learnings (create after each task)

### Workflow for AI Agents
**Session Start**:
1. Read [BUILD_PROGRESS.md](docs/BUILD_PROGRESS.md) to understand current state
2. Read relevant task doc from `docs/tasks/`
3. Check `docs/learnings/` for related insights

**During Work**:
- Focus on implementation
- Reference specific docs as needed

**Session End**:
1. Update `docs/BUILD_PROGRESS.md` (mark completed, move to next task)
2. Append summary to `docs/BUILD_LOG.md`
3. Create `docs/learnings/NN_task_name_learnings.md` if task complete
4. Note any documentation updates needed
5. Create commit according to [COMMITS.md](guides/COMMITS.md) documentation

See [PROGRESS_TRACKING.md](guides/PROGRESS_TRACKING.md) for detailed guide.

## Documentation Structure

### 📋 Specifications (What to Build)
- [REQUIREMENTS.md](docs/specs/REQUIREMENTS.md) - Feature list and constraints
- [ARCHITECTURE.md](docs/specs/ARCHITECTURE.md) - System design and tech stack
- [DATA_MODELS.md](docs/specs/DATA_MODELS.md) - Go structs and interfaces

### 🔨 Tasks (How to Build)
1. [Project Setup](docs/tasks/01_PROJECT_SETUP.md) - Scaffolding, dependencies, Makefile
2. [HTML Parser](docs/tasks/02_HTML_PARSER.md) - Parse HTML version, title, headings, forms
3. [Link Checker](docs/tasks/03_LINK_CHECKER.md) - Extract and verify links concurrently
4. [HTTP Handlers](docs/tasks/04_HTTP_HANDLERS.md) - Web server, routing, templates
5. [E2E Tests](docs/tasks/05_E2E_TESTS.md) - 

### 📚 Guides (Best Practices)
- [GO_BEST_PRACTICES.md](docs/guides/GO_BEST_PRACTICES.md) - Go idioms for this project
- [TESTING_STRATEGY.md](docs/guides/TESTING_STRATEGY.md) - Unit test patterns
- [SECURITY.md](docs/guides/SECURITY.md) - SSRF prevention, input validation

## Tech Stack
- **Language**: Go 1.21+
- **Router**: `net/http` (standard library)
- **Parser**: `github.com/PuerkitoBio/goquery`
- **Templates**: `html/template`
- **Deployment**: Docker

## Project Status

### Completed
- [ ] Project scaffolding
- [ ] HTML version detection
- [ ] Title extraction
- [ ] Heading analysis
- [ ] Login form detection
- [ ] Link extraction
- [ ] Internal/external link classification
- [ ] Concurrent link checking
- [ ] HTTP handlers
- [ ] Templates (form + results)
- [ ] Error handling
- [ ] Unit tests (80%+ coverage)
- [ ] Dockerfile
- [ ] Makefile
- [ ] Documentation

### Current Task
Start with [01_PROJECT_SETUP.md](tasks/01_PROJECT_SETUP.md)

## Key Principles
1. **Simplicity First** - Use standard library where possible
2. **Test Coverage** - 80%+ unit test coverage required
3. **Security** - SSRF protection, input validation
4. **Concurrency** - Goroutines + channels for link checking
5. **No JavaScript** - HTML + CSS only for frontend

## Repository Structure
```
.
├── cmd/
│   └── main.go
├── internal/
│   ├── analyzer/
│   ├── handler/
│   └── models/
├── web/
│   ├── templates/
│   └── static/
├── Dockerfile
├── Makefile
└── docs/
```

## Success Criteria
✅ All features implemented per REQUIREMENTS.md  
✅ Unit tests pass with ≥80% coverage  
✅ Docker image builds and runs  
✅ Application works as single binary  
✅ Clean, idiomatic Go code  
✅ SSRF protection implemented  

---

**For AI Agents**: Read this file first, then dive into specific task files. Each task is self-contained with clear acceptance criteria. Reference guides for best practices.
