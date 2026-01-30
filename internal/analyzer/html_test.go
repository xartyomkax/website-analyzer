package analyzer

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestDetectHTMLVersion(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "HTML5",
			html:     `<!DOCTYPE html><html><head></head><body></body></html>`,
			expected: "HTML5",
		},
		{
			name:     "HTML 4.01 Strict",
			html:     `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd"><html></html>`,
			expected: "HTML 4.01 Strict",
		},
		{
			name:     "HTML 4.01 Transitional",
			html:     `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN" "http://www.w3.org/TR/html4/loose.dtd"><html></html>`,
			expected: "HTML 4.01 Transitional",
		},
		{
			name:     "XHTML 1.0 Strict",
			html:     `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd"><html></html>`,
			expected: "XHTML 1.0 Strict",
		},
		{
			name:     "XHTML 1.0 Transitional",
			html:     `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd"><html></html>`,
			expected: "XHTML 1.0 Transitional",
		},
		{
			name:     "No DOCTYPE",
			html:     `<html><head></head><body></body></html>`,
			expected: "HTML5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			result := DetectHTMLVersion(doc)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "Normal title",
			html:     `<html><head><title>Test Page</title></head></html>`,
			expected: "Test Page",
		},
		{
			name:     "Title with whitespace",
			html:     `<html><head><title>  Spaced Title  </title></head></html>`,
			expected: "Spaced Title",
		},
		{
			name:     "No title",
			html:     `<html><head></head></html>`,
			expected: "No title",
		},
		{
			name:     "Empty title",
			html:     `<html><head><title></title></head></html>`,
			expected: "No title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			result := ExtractTitle(doc)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestCountHeadings(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected map[string]int
	}{
		{
			name: "Various headings",
			html: `
				<html><body>
					<h1>Title</h1>
					<h2>Section 1</h2>
					<h2>Section 2</h2>
					<h3>Subsection</h3>
				</body></html>
			`,
			expected: map[string]int{
				"h1": 1,
				"h2": 2,
				"h3": 1,
				"h4": 0,
				"h5": 0,
				"h6": 0,
			},
		},
		{
			name: "No headings",
			html: `<html><body><p>No headings here</p></body></html>`,
			expected: map[string]int{
				"h1": 0,
				"h2": 0,
				"h3": 0,
				"h4": 0,
				"h5": 0,
				"h6": 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			result := CountHeadings(doc)

			for level, expected := range tt.expected {
				if result[level] != expected {
					t.Errorf("Heading %s: expected %d, got %d", level, expected, result[level])
				}
			}
		})
	}
}

func TestHasLoginForm(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected bool
	}{
		{
			name: "Has login form",
			html: `
				<html><body>
					<form action="/login" method="post">
						<input type="text" name="username">
						<input type="password" name="password">
						<button type="submit">Login</button>
					</form>
				</body></html>
			`,
			expected: true,
		},
		{
			name: "No password input",
			html: `
				<html><body>
					<form action="/search" method="get">
						<input type="text" name="q">
						<button type="submit">Search</button>
					</form>
				</body></html>
			`,
			expected: false,
		},
		{
			name:     "No forms",
			html:     `<html><body><p>No forms here</p></body></html>`,
			expected: false,
		},
		{
			name: "Multiple forms, one with password",
			html: `
				<html><body>
					<form action="/search"><input type="text" name="q"></form>
					<form action="/login"><input type="password" name="pass"></form>
				</body></html>
			`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			result := HasLoginForm(doc)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
