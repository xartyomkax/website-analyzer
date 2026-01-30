package handler

import (
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"website-analyzer/internal/analyzer"
	"website-analyzer/internal/models"
)

type Handler struct {
	analyzer  *analyzer.Analyzer
	templates *template.Template
}

func NewHandler(analyzer *analyzer.Analyzer, templatesPath string) (*Handler, error) {
	tmpl, err := template.ParseGlob(templatesPath + "/*.html")
	if err != nil {
		return nil, err
	}

	return &Handler{
		analyzer:  analyzer,
		templates: tmpl,
	}, nil
}

func (h *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := struct {
		Error string
	}{}

	if err := h.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		slog.Error("template error", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		h.renderError(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	targetURL := r.FormValue("url")

	// Analyze
	start := time.Now()
	result, err := h.analyzer.Analyze(targetURL)
	duration := time.Since(start)

	slog.Info("analysis completed",
		"url", targetURL,
		"duration", duration,
		"error", err)

	if err != nil {
		h.renderError(w, err.Error(), http.StatusBadGateway)
		return
	}

	// Render results
	h.renderResults(w, result)
}

func (h *Handler) renderResults(w http.ResponseWriter, result *models.AnalysisResult) {
	data := struct {
		Result *models.AnalysisResult
	}{
		Result: result,
	}

	if err := h.templates.ExecuteTemplate(w, "results.html", data); err != nil {
		slog.Error("template error", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) renderError(w http.ResponseWriter, errMsg string, statusCode int) {
	data := struct {
		Error      string
		StatusCode int
	}{
		Error:      errMsg,
		StatusCode: statusCode,
	}

	w.WriteHeader(statusCode)
	if err := h.templates.ExecuteTemplate(w, "error.html", data); err != nil {
		slog.Error("template error", "error", err)
		http.Error(w, errMsg, statusCode)
	}
}
