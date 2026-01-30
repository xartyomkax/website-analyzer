package analyzer

import (
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/xartyomkax/website-analyzer/internal/models"
)

func TestExtractLinks(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		baseURL  string
		expected int
		internal int
		external int
	}{
		{
			name: "Internal and external links",
			html: `
				<html><body>
					<a href="/about">About</a>
					<a href="https://example.com/contact">Contact</a>
					<a href="https://google.com">Google</a>
				</body></html>
			`,
			baseURL:  "https://example.com",
			expected: 3,
			internal: 2,
			external: 1,
		},
		{
			name: "Skip invalid links",
			html: `
				<html><body>
					<a href="javascript:void(0)">JS</a>
					<a href="mailto:test@example.com">Email</a>
					<a href="#">Anchor</a>
					<a href="/valid">Valid</a>
				</body></html>
			`,
			baseURL:  "https://example.com",
			expected: 1,
			internal: 1,
			external: 0,
		},
		{
			name: "Deduplicate links",
			html: `
				<html><body>
					<a href="/page">Page 1</a>
					<a href="/page">Page 2</a>
					<a href="https://example.com/page">Page 3</a>
				</body></html>
			`,
			baseURL:  "https://example.com",
			expected: 1,
			internal: 1,
			external: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			links, err := ExtractLinks(doc, tt.baseURL)

			if err != nil {
				t.Fatalf("ExtractLinks failed: %v", err)
			}

			if len(links) != tt.expected {
				t.Errorf("Expected %d links, got %d", tt.expected, len(links))
			}

			internal := 0
			external := 0
			for _, link := range links {
				if link.Type == models.LinkTypeInternal {
					internal++
				} else if link.Type == models.LinkTypeExternal {
					external++
				}
			}

			if internal != tt.internal {
				t.Errorf("Expected %d internal links, got %d", tt.internal, internal)
			}
			if external != tt.external {
				t.Errorf("Expected %d external links, got %d", tt.external, external)
			}
		})
	}
}

func TestResolveURL(t *testing.T) {
	baseURL := mustParseURL("https://example.com/path/page.html")

	tests := []struct {
		name     string
		href     string
		expected string
		hasError bool
	}{
		{"Absolute URL", "https://google.com", "https://google.com", false},
		{"Relative path", "/about", "https://example.com/about", false},
		{"Relative to current", "contact", "https://example.com/path/contact", false},
		{"Skip javascript", "javascript:void(0)", "", false},
		{"Skip mailto", "mailto:test@example.com", "", false},
		{"Skip anchor", "#section", "", false},
		{"Skip tel", "tel:+1234567890", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolveURL(baseURL, tt.href)

			if tt.hasError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestClassifyLink(t *testing.T) {
	baseURL := mustParseURL("https://example.com")

	tests := []struct {
		name     string
		link     string
		expected models.LinkType
	}{
		{"Internal same path", "https://example.com/about", models.LinkTypeInternal},
		{"Internal root", "https://example.com/", models.LinkTypeInternal},
		{"External", "https://google.com", models.LinkTypeExternal},
		{"External subdomain", "https://blog.example.com", models.LinkTypeExternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifyLink(tt.link, baseURL)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// Helper
func mustParseURL(s string) *url.URL {
	u, _ := url.Parse(s)
	return u
}
