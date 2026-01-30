# Commit Message Specification

## Format

```
<type>: <subject>

[optional body]

[optional footer]
```

## Type

Choose one:
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation only
- `style` - Formatting, missing semicolons, etc (no code change)
- `refactor` - Code change that neither fixes a bug nor adds a feature
- `perf` - Performance improvement
- `test` - Adding or updating tests
- `chore` - Maintenance (deps, build, etc)

## Subject

- Use imperative mood ("add" not "added")
- No capitalization
- No period at the end
- Maximum 50 characters

## Body (Optional)

- Explain **what** and **why**, not how
- Wrap at 72 characters
- Separate from subject with blank line

## Examples

### Simple
```
feat: add HTML version detection
```

### With Body
```
fix: prevent SSRF attacks on private IPs

Block requests to private IP ranges (10.0.0.0/8, 192.168.0.0/16, etc)
and localhost to prevent server-side request forgery.
```

### Breaking Change
```
feat: add concurrent link checking

BREAKING CHANGE: Analyzer now requires MaxWorkers config parameter
```

### Multiple Changes
```
chore: update dependencies and fix linting

- Upgrade goquery to v1.9.0
- Fix golangci-lint warnings in validator package
- Update go.mod and go.sum
```

## Quick Reference

| Type | When to Use | Example |
|------|-------------|---------|
| `feat` | New feature | `feat: add login form detection` |
| `fix` | Bug fix | `fix: handle empty title tags` |
| `docs` | Documentation | `docs: update README with examples` |
| `test` | Tests only | `test: add HTML parser unit tests` |
| `refactor` | Code cleanup | `refactor: simplify link classification` |
| `chore` | Tooling/deps | `chore: add Makefile targets` |

## Bad Examples ❌

```
Added new feature
Fix bug
updated readme
WIP
asdf
```

## Good Examples ✅

```
feat: add heading count analysis
fix: resolve relative URLs correctly
docs: add security guide
test: cover SSRF prevention edge cases
refactor: extract HTTP client config
chore: update golangci-lint to v1.55
```

## Tips

1. **One commit = one logical change**
2. **Commit often** - small, focused commits
3. **Test before committing** - `make test`
4. **Format before committing** - `make fmt`

## Pre-commit Checklist

```bash
make fmt              # Format code
make test             # Run tests
git add .             # Stage changes
git commit -m "..."   # Commit with proper message
```

## Git Aliases (Optional)

Add to `~/.gitconfig`:

```ini
[alias]
    cm = commit -m
    feat = "!f() { git commit -m \"feat: $*\"; }; f"
    fix = "!f() { git commit -m \"fix: $*\"; }; f"
    docs = "!f() { git commit -m \"docs: $*\"; }; f"
    test = "!f() { git commit -m \"test: $*\"; }; f"
```

Usage:
```bash
git feat add login form detection
git fix handle empty titles
```