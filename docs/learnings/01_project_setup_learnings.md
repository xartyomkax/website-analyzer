# Task 01: Project Setup Learnings

## Technical Learnings
- **Go Version**: The project uses Go 1.24+ as per `ARCHITECTURE.md`.
- **Docker Multi-stage**: The Dockerfile efficiently separates `builder`, `debug`, and `production` stages. The production stage uses `alpine:latest` and runs as a non-root user (`appuser`).
- **Dependencies**: `goquery` is the primary external dependency for HTML parsing.
- **Project Structure**: Following the standard Go project layout with `cmd/`, `internal/`, and `web/` ensures clear separation of concerns.

## Challenges & Solutions
- **Commit Standards**: Initial commit missed the `COMMITS.md` specification. Fixed by using `<type>: <subject>` format with imperative mood.
- **Missing Directories**: Standard directories like `internal/analyzer` weren't physically present initially (only in docs). Created them with `.gitkeep` to maintain structure.

## Best Practices
- **Early Verification**: Running `make all` early (fmt, tidy, test, build) ensures the foundation is solid before adding features.
- **Containerization First**: Building the Docker image immediately after setup catches environment issues early.
- **Documentation Alignment**: Ensuring `README.md` and `ARCHITECTURE.md` are consistent with the actual implementation is crucial for long-term maintenance.
