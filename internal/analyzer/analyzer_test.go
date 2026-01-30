package analyzer

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestAnalyzer_Analyze(t *testing.T) {
	os.Setenv("ALLOW_PRIVATE_IPS", "true")
	defer os.Unsetenv("ALLOW_PRIVATE_IPS")

	// Mock server to serve test HTML
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`
			<!DOCTYPE html>
			<html>
			<head><title>Test Page</title></head>
			<body>
				<h1>Title 1</h1>
				<a href="/internal">Internal</a>
				<a href="https://extern.com">External</a>
				<form><input type="password"></form>
			</body>
			</html>
		`))
	}))
	defer ts.Close()

	config := &Config{
		RequestTimeout:  2 * time.Second,
		LinkTimeout:     1 * time.Second,
		MaxWorkers:      5,
		MaxResponseSize: 1024 * 1024,
	}

	a := NewAnalyzer(config)

	result, err := a.Analyze(ts.URL)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if result.Title != "Test Page" {
		t.Errorf("Expected title 'Test Page', got '%s'", result.Title)
	}

	if result.Headings["h1"] != 1 {
		t.Errorf("Expected 1 h1, got %d", result.Headings["h1"])
	}

	if result.InternalLinks != 1 {
		t.Errorf("Expected 1 internal link, got %d", result.InternalLinks)
	}

	if result.ExternalLinks != 1 {
		t.Errorf("Expected 1 external link, got %d", result.ExternalLinks)
	}

	if !result.HasLoginForm {
		t.Error("Expected login form to be detected")
	}
}
