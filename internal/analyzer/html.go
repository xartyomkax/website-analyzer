package analyzer

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// DetectHTMLVersion parses the DOCTYPE and returns the HTML version
func DetectHTMLVersion(doc *goquery.Document) string {
	// Get the HTML node
	htmlNode := doc.Find("html").First()
	if htmlNode.Length() == 0 {
		return "HTML5" // Default
	}

	// Try to get DOCTYPE from document
	// goquery doesn't directly expose DOCTYPE, so we check common patterns
	html, _ := doc.Html()
	htmlLower := strings.ToLower(html)

	// HTML5
	if strings.Contains(htmlLower, "<!doctype html>") {
		return "HTML5"
	}

	// HTML 4.01 Strict
	if strings.Contains(htmlLower, "html 4.01") && strings.Contains(htmlLower, "strict") {
		return "HTML 4.01 Strict"
	}

	// HTML 4.01 Transitional
	if strings.Contains(htmlLower, "html 4.01") && strings.Contains(htmlLower, "transitional") {
		return "HTML 4.01 Transitional"
	}

	// XHTML 1.0 Strict
	if strings.Contains(htmlLower, "xhtml 1.0") && strings.Contains(htmlLower, "strict") {
		return "XHTML 1.0 Strict"
	}

	// XHTML 1.0 Transitional
	if strings.Contains(htmlLower, "xhtml 1.0") && strings.Contains(htmlLower, "transitional") {
		return "XHTML 1.0 Transitional"
	}

	// Default to HTML5 for modern pages
	return "HTML5"
}

// ExtractTitle returns the page title, or "No title" if not found
func ExtractTitle(doc *goquery.Document) string {
	title := doc.Find("title").First().Text()
	title = strings.TrimSpace(title)

	if title == "" {
		return "No title"
	}

	return title
}

// CountHeadings counts headings by level (h1-h6)
func CountHeadings(doc *goquery.Document) map[string]int {
	headings := map[string]int{
		"h1": 0,
		"h2": 0,
		"h3": 0,
		"h4": 0,
		"h5": 0,
		"h6": 0,
	}

	// Count each heading level
	for level := 1; level <= 6; level++ {
		selector := fmt.Sprintf("h%d", level)
		count := doc.Find(selector).Length()
		headings[selector] = count
	}

	return headings
}

// HasLoginForm checks if the page contains a login form
// (a form with a password input field)
func HasLoginForm(doc *goquery.Document) bool {
	// Find all forms
	hasPasswordInput := false

	doc.Find("form").Each(func(i int, form *goquery.Selection) {
		// Check if this form has a password input
		if form.Find("input[type='password']").Length() > 0 {
			hasPasswordInput = true
		}
	})

	return hasPasswordInput
}
