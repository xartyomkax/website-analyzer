package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"website-analyzer/internal/analyzer"
	"website-analyzer/internal/config"
	"website-analyzer/internal/handler"
)

func main() {
	// Configure logging
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// Configuration
	cfg := config.LoadConfig()

	// Analyzer config
	analyzerCfg := &analyzer.Config{
		RequestTimeout:  cfg.RequestTimeout,
		LinkTimeout:     cfg.LinkTimeout,
		MaxWorkers:      cfg.MaxWorkers,
		MaxResponseSize: cfg.MaxResponseSize,
		MaxURLLength:    cfg.MaxURLLength,
		MaxRedirects:    cfg.MaxRedirects,
	}

	// Create analyzer
	analyzer := analyzer.NewAnalyzer(analyzerCfg)

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
	addr := ":" + cfg.Port
	slog.Info("server starting", "addr", addr, "env", cfg.Env)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
