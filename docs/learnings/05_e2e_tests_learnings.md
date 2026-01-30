# Task 05: E2E Tests Learnings

## Technical Learnings
- **Httptest for E2E**: Using `httptest.NewServer` to mock the *target* website and `httptest.NewRecorder` to mock the *client* allows for fully isolated E2E tests without external dependencies or actual browser automation.
- **Template Path Resolution**: When running tests from a package directory (like `internal/handler`), relative paths to templates (like `../../web/templates`) must be carefully managed.
- **Environment Isolation**: Setting `ALLOW_PRIVATE_IPS=true` via `os.Setenv` is essential for E2E tests that analyze `127.0.0.1` servers, preventing the SSRF protector from blocking the test.

## Challenges & Solutions
- **String Assertions**: HTML templates use dynamic logic (e.g., `{{if .HasLoginForm}}Yes{{else}}No{{end}}`). Tests must assert the *rendered* output rather than the raw data structures.
- **Error Cascading**: Validating how the application handles invalid URLs across layers (Handler -> Validator -> Analyzer) ensured that error messages are properly propagated and displayed to the user.

## Best Practices
- **Full Flow Verification**: E2E tests should cover the "happy path" (successful analysis), edge cases (login forms, various link types), and failure modes (invalid URLs, network errors).
- **Cleanup**: Always unset environment variables and close test servers using `defer` to ensure a clean state for subsequent tests.
