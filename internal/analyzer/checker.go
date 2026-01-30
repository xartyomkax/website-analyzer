package analyzer

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"website-analyzer/internal/models"
)

// CheckLinksConfig holds configuration for link checking
type CheckLinksConfig struct {
	Timeout      time.Duration
	MaxWorkers   int
	MaxRedirects int
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
	workerCount := config.MaxWorkers
	if workerCount <= 0 {
		workerCount = 10
	}
	wg.Add(workerCount)

	for w := 0; w < workerCount; w++ {
		go worker(jobs, results, config, &wg)
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
func worker(jobs <-chan models.Link, results chan<- checkResult, config CheckLinksConfig, wg *sync.WaitGroup) {
	defer wg.Done()

	client := &http.Client{
		Timeout: config.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= config.MaxRedirects {
				return fmt.Errorf("Too many redirects")
			}
			return nil
		},
	}

	for link := range jobs {
		result := checkLink(client, link.URL)
		results <- result
	}
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
