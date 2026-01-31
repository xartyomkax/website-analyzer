package analyzer

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"website-analyzer/internal/models"
)

func TestCheckLinks(t *testing.T) {
	// Create test servers
	server200 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server200.Close()

	server404 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server404.Close()

	links := []models.Link{
		{URL: server200.URL, Type: models.LinkTypeExternal},
		{URL: server404.URL, Type: models.LinkTypeExternal},
	}

	config := CheckLinksConfig{
		Timeout:    5 * time.Second,
		MaxWorkers: 2,
	}

	errors := CheckLinks(links, config)

	// Should have 1 error (404)
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	}

	if len(errors) > 0 && errors[0].StatusCode != 404 {
		t.Errorf("Expected status code 404, got %d", errors[0].StatusCode)
	}
}

func TestCheckLinksTimeout(t *testing.T) {
	// Create slow server
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer slowServer.Close()

	links := []models.Link{
		{URL: slowServer.URL, Type: models.LinkTypeExternal},
	}

	config := CheckLinksConfig{
		Timeout:    100 * time.Millisecond, // Very short timeout
		MaxWorkers: 1,
	}

	errors := CheckLinks(links, config)

	// Should timeout
	if len(errors) != 1 {
		t.Errorf("Expected timeout error, got %d errors", len(errors))
	}
}

func TestCheckLinksMultipleStatuses(t *testing.T) {
	// Create servers with different status codes
	server200 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server200.Close()

	server301 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMovedPermanently)
	}))
	defer server301.Close()

	server500 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server500.Close()

	links := []models.Link{
		{URL: server200.URL, Type: models.LinkTypeExternal},
		{URL: server301.URL, Type: models.LinkTypeExternal},
		{URL: server500.URL, Type: models.LinkTypeExternal},
	}

	config := CheckLinksConfig{
		Timeout:    5 * time.Second,
		MaxWorkers: 3,
	}

	errors := CheckLinks(links, config)

	// Should have 1 error (500)
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	}

	if len(errors) > 0 && errors[0].StatusCode != 500 {
		t.Errorf("Expected status code 500, got %d", errors[0].StatusCode)
	}
}

func TestCheckLinksEmpty(t *testing.T) {
	links := []models.Link{}

	config := CheckLinksConfig{
		Timeout:    5 * time.Second,
		MaxWorkers: 2,
	}

	errors := CheckLinks(links, config)

	if errors != nil {
		t.Errorf("Expected nil for empty links, got %v", errors)
	}
}

func TestCheckLinksGoroutineLeak(t *testing.T) {
	// Sample links
	links := []models.Link{
		{URL: "http://example.com", Type: models.LinkTypeExternal},
	}

	config := CheckLinksConfig{
		Timeout:    100 * time.Millisecond,
		MaxWorkers: 5,
	}

	initialGoroutines := runtime.NumGoroutine()

	// Run multiple times to see if leaks accumulate
	for i := 0; i < 10; i++ {
		_ = CheckLinks(links, config)
	}

	// Small buffer for any runtime-background goroutines that might have started
	// but generally it should be stable. Let's wait a bit for any GC/cleanup.
	time.Sleep(100 * time.Millisecond)

	finalGoroutines := runtime.NumGoroutine()
	if finalGoroutines > initialGoroutines+2 { // +2 for potential background noise, but should be stable
		t.Errorf("Potential goroutine leak: started with %d, ended with %d", initialGoroutines, finalGoroutines)
	}
}

func TestCheckLinksDefaultWorkers(t *testing.T) {
	server200 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server200.Close()

	links := []models.Link{
		{URL: server200.URL, Type: models.LinkTypeExternal},
	}

	// Test with invalid worker count (should default to 10)
	config := CheckLinksConfig{
		Timeout:    5 * time.Second,
		MaxWorkers: 0,
	}

	errors := CheckLinks(links, config)

	if len(errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(errors))
	}
}
