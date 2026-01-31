package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
	"website-analyzer/internal/analyzer"
)

func TestE2E_FullFlow(t *testing.T) {
	// 1. Setup mock target server (the site being analyzed)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`
			<!DOCTYPE html>
			<html>
			<head><title>E2E Test Site</title></head>
			<body>
				<h1>Welcome</h1>
				<a href="/about">Internal Link</a>
				<a href="https://google.com">External Link</a>
				<form action="/login" method="POST">
					<input type="password" name="pwd">
				</form>
			</body>
			</html>
		`))
	}))
	defer ts.Close()

	// 2. Allow private IPs for local testing
	os.Setenv("ALLOW_PRIVATE_IPS", "true")
	defer os.Unsetenv("ALLOW_PRIVATE_IPS")

	// 3. Setup Analyzer
	analyzerCfg := &analyzer.Config{
		RequestTimeout:  5 * time.Second,
		LinkTimeout:     2 * time.Second,
		MaxWorkers:      5,
		MaxResponseSize: 1024 * 1024,
		MaxURLLength:    2048,
		MaxRedirects:    5,
	}
	a := analyzer.NewAnalyzer(analyzerCfg)

	// 4. Setup Handler
	// Note: Path is relative to the test file location (internal/handler)
	h, err := NewHandler(a, "../../web/templates")
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// 5. Test Index Page (GET /)
	t.Run("IndexPage", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()
		h.IndexHandler(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", rr.Code)
		}

		body := rr.Body.String()
		if !strings.Contains(body, "Web Page Analyzer") {
			t.Error("Index page doesn't contain expected content")
		}
	})

	// 6. Test Analysis (POST /analyze)
	t.Run("AnalyzeFlow", func(t *testing.T) {
		form := url.Values{}
		form.Add("url", ts.URL)

		req := httptest.NewRequest("POST", "/analyze", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		h.AnalyzeHandler(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v. Body: %s", rr.Code, rr.Body.String())
		}

		body := rr.Body.String()
		expectedSnippets := []string{
			"E2E Test Site",
			"HTML5",
			"Internal Links",
			"External Links",
			"Yes", // Login Form: Yes
		}

		for _, snippet := range expectedSnippets {
			if !strings.Contains(body, snippet) {
				t.Errorf("Results page missing expected snippet: %s", snippet)
			}
		}
	})

	// 7. Test Error Handling (Invalid URL)
	t.Run("InvalidURL", func(t *testing.T) {
		form := url.Values{}
		form.Add("url", "not-a-url")

		req := httptest.NewRequest("POST", "/analyze", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		h.AnalyzeHandler(rr, req)

		if rr.Code != http.StatusBadGateway {
			t.Errorf("Expected status Bad Gateway, got %v", rr.Code)
		}

		body := rr.Body.String()
		if !strings.Contains(body, "URL scheme must be http or https") {
			t.Errorf("Error page missing expected error message. Got: %s", body)
		}
	})
}
