package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/xartyomkax/website-analyzer/internal/analyzer"
	"github.com/xartyomkax/website-analyzer/internal/handler"
)

func main() {
	// Configure logging
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// Configuration
	config := &analyzer.Config{
		RequestTimeout:  30 * time.Second,
		LinkTimeout:     5 * time.Second,
		MaxWorkers:      10,
		MaxResponseSize: 10 * 1024 * 1024, // 10MB
	}

	// Create analyzer
	analyzer := analyzer.NewAnalyzer(config)

	// Create handler
	h, err := handler.NewHandler(analyzer, "web/templates")
	if err != nil {
		log.Fatal("Failed to load templates:", err)
	}

	// Routes
	http.HandleFunc("/", h.IndexHandler)
	http.HandleFunc("/analyze", h.AnalyzeHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	slog.Info("server starting", "addr", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
