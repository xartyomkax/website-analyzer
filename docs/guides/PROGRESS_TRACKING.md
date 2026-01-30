# Progress Tracking Guide

## Overview
Track build progress and learnings in separate, focused files to avoid token bloat and improve maintainability.

## File Structure

```
.
├── BUILD_PROGRESS.md      # Current status only (keep lean!)
├── BUILD_LOG.md           # Append-only debug log (last resort)
└── docs/
    └── learnings/
        ├── 01_project_setup_learnings.md
        ├── 02_html_parser_learnings.md
        ├── 03_link_checker_learnings.md
        ├── 04_http_handlers_learnings.md
        └── 05_e2e_tests_learnings.md
```

## BUILD_PROGRESS.md (Keep Current)

**Purpose**: Quick status check - what's done, what's next  
**Size**: Always keep under 100 lines - archive old content to BUILD_LOG.md

```markdown
# Build Progress

Last Updated: 2026-01-30

## Current Task
Task 03: Link Checker - Implementing concurrent workers

## Completed ✅
- [x] Task 01: Project Setup (2026-01-28)
- [x] Task 02: HTML Parser (2026-01-29)

## In Progress 🔧
- [ ] Task 03: Link Checker
  - [x] Link extraction
  - [x] Internal/external classification
  - [ ] Concurrent checking (in progress)

## Next Up 📋
- [ ] Task 04: HTTP Handlers
- [ ] Task 05: Deployment

## Key Decisions
- Changed MaxWorkers from 20→10 (memory constraints)
- Using goquery instead of x/net/html (easier API)

## Blockers 🚧
None currently

## Metrics
- Test Coverage: 82%
- Files Created: 12
- Lines of Code: ~800
```

**Update After Each Session**: Archive old decisions to BUILD_LOG.md

---

## BUILD_LOG.md (Debug History)

**Purpose**: Full chronological log - only read when debugging complex issues  
**Usage**: Append-only, never edit old entries  
**When to Read**: Something broke and you need full context

```markdown
# Build Log

## 2026-01-30 - Session 3
**Task**: Implementing concurrent link checker
**Duration**: 45 min
**Outcome**: Success with adjustments

### What Happened
- Initial worker pool with 20 goroutines caused high memory usage
- Reduced to 10 workers - resolved issue
- Added context timeouts to prevent goroutine leaks

### Changes Made
- `internal/analyzer/checker.go`: Reduced MaxWorkers 20→10
- Added `context.WithTimeout` to all HTTP requests

### Issues Encountered
- Race condition in result collection (fixed with mutex)
- Some links timing out (increased timeout 3s→5s)

### Files Modified
- internal/analyzer/checker.go
- internal/analyzer/checker_test.go

---

## 2026-01-29 - Session 2
**Task**: HTML Parser implementation
...
```

---

## docs/learnings/ (Task-Specific Learnings)

**Purpose**: Reusable knowledge per task - add to documentation  
**When to Create**: After completing each task  
**Format**: Actionable insights for future similar projects

### Example: `docs/learnings/03_link_checker_learnings.md`

```markdown
# Learnings: Link Checker Implementation

## What Worked Well ✅

### Worker Pool Pattern
Using buffered channels with fixed worker count:
```go
jobs := make(chan Link, len(links))
results := make(chan Result, len(links))
```
- Prevents memory spikes
- Easy to tune concurrency
- Clean shutdown with WaitGroup

### Context-Based Timeouts
Per-request timeouts prevent hanging goroutines:
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

## What Didn't Work ❌

### Initial Approach: Unlimited Goroutines
```go
// Bad - spawned 200+ goroutines for large sites
for _, link := range links {
    go checkLink(link)
}
```
**Issue**: Memory usage spiked, OS file descriptor limits hit

**Solution**: Worker pool with 10 fixed workers

### HEAD Requests Not Always Reliable
Some servers don't respond to HEAD, had to fallback to GET
```go
// Better approach
resp, err := client.Head(url)
if err != nil || resp.StatusCode >= 400 {
    resp, err = client.Get(url) // Fallback
}
```

## Recommendations for AGENT.md Updates

### Add to Architecture Doc
- Worker pool pattern example with recommended worker count formula:
  `MaxWorkers = min(NumCPUs * 2, 20)`

### Add to Testing Strategy
- Always test concurrent code with `-race` flag
- Mock slow servers to test timeout behavior

### Add to Best Practices
- Use buffered channels sized to workload
- Always use context for HTTP requests
- Defer cancel() immediately after context creation

## Metrics
- Initial implementation: 200+ goroutines, 500MB memory
- Optimized version: 10 goroutines, 50MB memory
- Performance: 50 links checked in ~6 seconds (vs ~250s sequential)

## Would Do Differently Next Time
1. Start with worker pool pattern, not unlimited goroutines
2. Add configurable worker count from day 1
3. Include memory profiling in benchmarks
```

---

## Integration with AGENT.md

Update your main AGENT.md to reference these files:

```markdown
## Progress Tracking

**Check Status**: Read `BUILD_PROGRESS.md` at start of each session
**Log Work**: Append to `BUILD_LOG.md` after each session  
**Capture Learnings**: Update `docs/learnings/NN_task_learnings.md` after completing each task

### AI Agent Instructions
1. At session start: `view BUILD_PROGRESS.md` to understand current state
2. During work: Focus on current task documentation
3. After completing task:
   - Update BUILD_PROGRESS.md (mark task complete, move to next)
   - Append summary to BUILD_LOG.md
   - Create/update relevant learnings file in docs/learnings/
   - If learnings suggest doc updates, note in BUILD_PROGRESS.md → "Documentation Updates Needed"

### When BUILD_PROGRESS.md Exceeds 100 Lines
Move "Key Decisions" and "Completed" tasks to BUILD_LOG.md, keep only:
- Current task
- In progress items
- Next 2-3 tasks
- Recent blockers (last 7 days)
```

---

## Template: Task Learnings File

```markdown
# Learnings: [Task Name]

## What Worked Well ✅
- Specific technique/pattern
- Why it worked
- Code example

## What Didn't Work ❌
- Initial approach
- Why it failed
- What we did instead

## Recommendations for Documentation Updates
- Specific sections to update
- New patterns to add
- Warnings to include

## Metrics
- Performance numbers
- Resource usage
- Time to implement

## Would Do Differently
1. Specific actionable item
2. Another improvement
```

---

## Example Workflow

### Start of Session
```bash
# AI reads current status
view BUILD_PROGRESS.md

# AI reads relevant task doc
view docs/tasks/03_LINK_CHECKER.md

# AI reads relevant learnings from previous similar tasks (if any)
view docs/learnings/
```

### During Work
- AI focuses on implementation
- Only references specific docs as needed

### End of Session
```bash
# Update progress
str_replace BUILD_PROGRESS.md ...

# Log what happened
str_replace BUILD_LOG.md ... (append new entry)

# Capture learnings
create_file docs/learnings/03_link_checker_learnings.md ...
```

---

## Why This Structure?

✅ **BUILD_PROGRESS.md**: Always small, always current, AI reads every session  
✅ **BUILD_LOG.md**: Grows over time but only read when debugging issues  
✅ **learnings/*.md**: Reusable knowledge, can feed back into main docs  

This prevents:
- ❌ Single giant file that wastes tokens
- ❌ Lost context between sessions
- ❌ Repeated mistakes
- ❌ Outdated documentation

## Maintenance Rules

1. **BUILD_PROGRESS.md**: Prune weekly (move old content to BUILD_LOG.md)
2. **BUILD_LOG.md**: Never prune (archive if > 1000 lines)
3. **learnings/*.md**: Merge insights into main docs quarterly
