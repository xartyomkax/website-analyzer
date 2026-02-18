package analyzer

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"website-analyzer/internal/models"
)

// CheckLinksConfig holds configuration for link checking
type CheckLinksConfig struct {
	Timeout      time.Duration
	MaxWorkers   int
	MaxRedirects int
	Transport    http.RoundTripper // Optional custom transport for testing
}

// checkResult is used internally for worker communication
type checkResult struct {
	url        string
	statusCode int
	err        error
}

// CheckLinks verifies accessibility of links concurrently
func CheckLinks(links []models.Link, config CheckLinksConfig) []models.LinkError {
	if len(links) == 0 {
		return nil
	}

	// Channels for work distribution
	jobs := make(chan models.Link, len(links))
	results := make(chan checkResult, len(links))

	// Start worker pool
	var wg sync.WaitGroup
	wg.Add(config.MaxWorkers)

	// Circuit breaker
	cb := newCircuitBreaker(5)

	for w := 0; w < config.MaxWorkers; w++ {
		go worker(jobs, results, config, cb, &wg)
	}

	// Send jobs
	for _, link := range links {
		jobs <- link
	}
	close(jobs)

	// Wait for workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect errors
	var errors []models.LinkError
	for result := range results {
		if result.err != nil {
			errors = append(errors, models.LinkError{
				URL:        result.url,
				StatusCode: result.statusCode,
				Error:      result.err.Error(),
			})
		}
	}

	return errors
}

// worker processes link checking jobs
func worker(jobs <-chan models.Link, results chan<- checkResult, config CheckLinksConfig, cb *circuitBreaker, wg *sync.WaitGroup) {
	defer wg.Done()

	client := &http.Client{
		Timeout:   config.Timeout,
		Transport: config.Transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= config.MaxRedirects {
				return fmt.Errorf("Too many redirects")
			}
			return nil
		},
	}

	for link := range jobs {
		domain := getDomain(link.URL)

		// Check circuit breaker
		if domain != "" && !cb.allow(domain) {
			continue
		}

		result := checkLink(client, link.URL)

		// Update circuit breaker based on result
		if domain != "" {
			if result.err != nil {
				cb.recordFailure(domain)
			} else {
				cb.recordSuccess(domain)
			}
		}

		results <- result
	}
}

func getDomain(linkURL string) string {
	u, err := url.Parse(linkURL)
	if err != nil {
		return ""
	}
	return u.Host
}

// checkLink performs a single link check
func checkLink(client *http.Client, url string) checkResult {
	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return checkResult{
			url:        url,
			statusCode: 0,
			err:        err,
		}
	}

	req.Header.Set("User-Agent", "WebPageAnalyzer/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return checkResult{
			url:        url,
			statusCode: 0,
			err:        err,
		}
	}
	defer resp.Body.Close()

	// Consider 2xx and 3xx as success
	if resp.StatusCode >= 400 {
		return checkResult{
			url:        url,
			statusCode: resp.StatusCode,
			err:        fmt.Errorf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode)),
		}
	}

	return checkResult{
		url:        url,
		statusCode: resp.StatusCode,
		err:        nil,
	}
}
