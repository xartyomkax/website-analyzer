package analyzer

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/xartyomkax/website-analyzer/internal/models"
)

// ExtractLinks finds all <a href> tags and returns their URLs
func ExtractLinks(doc *goquery.Document, baseURL string) ([]models.Link, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	var links []models.Link
	seen := make(map[string]bool) // Deduplicate

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" {
			return
		}

		// Resolve relative URLs
		resolved, err := resolveURL(base, href)
		if err != nil || resolved == "" {
			return
		}

		// Skip duplicates
		if seen[resolved] {
			return
		}
		seen[resolved] = true

		// Classify link
		linkType := classifyLink(resolved, base)

		links = append(links, models.Link{
			URL:  resolved,
			Type: linkType,
		})
	})

	return links, nil
}

// resolveURL converts relative URLs to absolute
func resolveURL(base *url.URL, href string) (string, error) {
	href = strings.TrimSpace(href)

	// Skip invalid schemes
	if strings.HasPrefix(href, "javascript:") ||
		strings.HasPrefix(href, "mailto:") ||
		strings.HasPrefix(href, "tel:") ||
		href == "#" ||
		strings.HasPrefix(href, "#") {
		return "", nil
	}

	// Parse href
	parsed, err := url.Parse(href)
	if err != nil {
		return "", err
	}

	// Resolve against base
	resolved := base.ResolveReference(parsed)

	// Only return http/https URLs
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return "", nil
	}

	return resolved.String(), nil
}

// classifyLink determines if a link is internal or external
func classifyLink(link string, base *url.URL) models.LinkType {
	parsed, err := url.Parse(link)
	if err != nil {
		return models.LinkTypeInvalid
	}

	// Same host (including subdomains) = internal
	if parsed.Host == base.Host {
		return models.LinkTypeInternal
	}

	return models.LinkTypeExternal
}
