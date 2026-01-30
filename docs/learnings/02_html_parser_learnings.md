# Task 02: HTML Parser Learnings

## Technical Learnings
- **Goquery DOCTYPE**: `goquery` does not directly expose the `DOCTYPE` declaration. Detecting the HTML version required extracting the raw HTML string and performing case-insensitive substring checks for common DOCTYPE patterns (HTML5, HTML 4.01, XHTML 1.0).
- **Title Cleanup**: Titles often contain leading/trailing whitespace or newlines. Using `strings.TrimSpace` is essential for clean data extraction.
- **Heading Analysis**: A compact way to count all heading levels is to loop from 1 to 6 and use `fmt.Sprintf("h%d", level)` as a selector.

## Challenges & Solutions
- **HTML Detection**: Initially tried to use goquery selectors but quickly realized `DOCTYPE` is outside the standard DOM traversal. Reading `doc.Html()` provided the necessary raw content.

## Best Practices
- **Edge Cases**: Always handle "No title" or empty titles gracefully to avoid showing empty strings in the UI.
- **Modularization**: Keeping parser logic in `internal/analyzer` separate from HTTP handlers makes it easier to test in isolation.
