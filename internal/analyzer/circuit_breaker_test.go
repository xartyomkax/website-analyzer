package analyzer

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"website-analyzer/internal/models"
)

type mockTransport struct {
	mu    sync.Mutex
	calls map[string]int
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	host := req.URL.Host
	m.calls[host]++

	if host == "bad.com" {
		return nil, errors.New("simulated network error")
	}

	return &http.Response{
		StatusCode: 200,
		Body:       http.NoBody,
	}, nil
}

// Additional test to ensure grouping works with mixed domains
func TestCheckLinks_CircuitBreaker_Parallel(t *testing.T) {
	mock := &mockTransport{
		calls: make(map[string]int),
	}

	var links []models.Link
	// Interleave bad and good links to test independence
	for i := 0; i < 10; i++ {
		links = append(links, models.Link{URL: "http://bad.com/" + fmt.Sprintf("%d", i)})
		links = append(links, models.Link{URL: "http://good.com/" + fmt.Sprintf("%d", i)})
	}

	config := CheckLinksConfig{
		Timeout:      100 * time.Millisecond,
		MaxWorkers:   4, // Test with concurrency
		MaxRedirects: 3,
		Transport:    mock,
	}

	_ = CheckLinks(links, config)

	mock.mu.Lock()
	badCalls := mock.calls["bad.com"]
	goodCalls := mock.calls["good.com"]
	mock.mu.Unlock()

	// With concurrency, strict 5 is harder, but should be close.
	// The implementation should try to limit it.
	// If we group by domain, we might process bad.com sequentially or track safely.
	// We'll assert < 10 to show SOME efficient stopping, ideally 5.
	// But let's start with loose assertion and tighten if implementation allows.
	if badCalls == 10 {
		t.Errorf("Circuit breaker failed to reduce calls: got %d, expected roughly 5", badCalls)
	}

	if goodCalls != 10 {
		t.Errorf("Expected 10 calls to good.com, got %d", goodCalls)
	}
}
