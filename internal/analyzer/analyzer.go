package analyzer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"website-analyzer/internal/models"
	"website-analyzer/internal/validator"

	"github.com/PuerkitoBio/goquery"
)

type Config struct {
	RequestTimeout  time.Duration
	LinkTimeout     time.Duration
	MaxWorkers      int
	MaxResponseSize int64
	MaxURLLength    int
	MaxRedirects    int
}

type Analyzer struct {
	config     *Config
	httpClient *http.Client
}

func NewAnalyzer(config *Config) *Analyzer {
	return &Analyzer{
		config: config,
		httpClient: &http.Client{
			Timeout: config.RequestTimeout,
		},
	}
}

func (a *Analyzer) Analyze(targetURL string) (*models.AnalysisResult, error) {
	// Validate URL
	if err := validator.ValidateURL(targetURL, a.config.MaxURLLength); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Fetch HTML
	doc, err := a.fetchHTML(targetURL)
	if err != nil {
		return nil, err
	}

	// Extract links
	links, err := ExtractLinks(doc, targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to extract links: %w", err)
	}

	// Count internal/external
	var internal, external int
	for _, link := range links {
		if link.Type == models.LinkTypeInternal {
			internal++
		}

		if link.Type == models.LinkTypeExternal {
			external++
		}
	}

	// Check link accessibility
	checkConfig := CheckLinksConfig{
		Timeout:      a.config.LinkTimeout,
		MaxWorkers:   a.config.MaxWorkers,
		MaxRedirects: a.config.MaxRedirects,
	}
	inaccessible := CheckLinks(links, checkConfig)

	// Build result
	result := &models.AnalysisResult{
		URL:               targetURL,
		HTMLVersion:       DetectHTMLVersion(doc),
		Title:             ExtractTitle(doc),
		Headings:          CountHeadings(doc),
		InternalLinks:     internal,
		ExternalLinks:     external,
		InaccessibleLinks: inaccessible,
		HasLoginForm:      HasLoginForm(doc),
	}

	return result, nil
}

func (a *Analyzer) fetchHTML(url string) (*goquery.Document, error) {
	ctx, cancel := context.WithTimeout(context.Background(), a.config.RequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "WebPageAnalyzer/1.0")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	// Limit response size
	limitedReader := io.LimitReader(resp.Body, a.config.MaxResponseSize)

	doc, err := goquery.NewDocumentFromReader(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	return doc, nil
}
