package handlers

import (
	"bytes"
	"log"
	"net/http"

	"github.com/harmey/blok2ttrpg/ability-builder/internal/docs"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

// DownloadDocsHandler generates a single markdown file from the docs templates
// and the YAML configuration, then returns it as a downloadable attachment.
func (app *App) DownloadDocsHandler(w http.ResponseWriter, r *http.Request) {
	cfg := app.Config
	if cfg == nil {
		http.Error(w, "Configuration not loaded", http.StatusInternalServerError)
		return
	}

	merged, err := docs.RenderFullDocumentation(cfg)
	if err != nil {
		log.Printf("failed to render docs: %v", err)
		http.Error(w, "Failed to render documentation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"blok2ttrpg-docs.md\"")
	if _, err := w.Write([]byte(merged)); err != nil {
		log.Printf("failed to write docs response: %v", err)
	}
}

// PdfDocsHandler renders the full documentation as a print-friendly HTML page
// and automatically opens the browser print dialog so the user can save it as PDF.
func (app *App) PdfDocsHandler(w http.ResponseWriter, r *http.Request) {
	cfg := app.Config
	if cfg == nil {
		http.Error(w, "Configuration not loaded", http.StatusInternalServerError)
		return
	}

	merged, err := docs.RenderFullDocumentation(cfg)
	if err != nil {
		log.Printf("failed to render docs: %v", err)
		http.Error(w, "Failed to render documentation", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	md := goldmark.New(goldmark.WithExtensions(extension.Table))
	if err := md.Convert([]byte(merged), &buf); err != nil {
		log.Printf("failed to convert docs to html: %v", err)
		http.Error(w, "Failed to convert documentation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Blok2 TTRPG Docs</title>
<style>
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #111; max-width: 900px; margin: 2rem auto; padding: 0 1rem; }
h1, h2, h3, h4 { color: #1a202c; margin-top: 1.5em; }
code { background: #f3f4f6; padding: 0.2em 0.4em; border-radius: 4px; font-size: 0.9em; }
pre { background: #f3f4f6; padding: 1em; border-radius: 8px; overflow-x: auto; }
pre code { background: none; padding: 0; }
table { border-collapse: collapse; width: 100%; margin: 1em 0; }
th, td { border: 1px solid #d1d5db; padding: 0.5em; text-align: left; }
th { background: #f3f4f6; }
@media print { .no-print { display: none; } body { margin: 0; } }
</style>
</head>
<body>
<div class="no-print" style="margin-bottom:1rem; padding:1rem; background:#f3f4f6; border-radius:8px;">
<button onclick="window.print()" class="no-print" style="padding:0.5em 1em; font-size:1rem; cursor:pointer;">Print / Save as PDF</button>
<span style="margin-left:1rem; color:#4b5563;">Use your browser's print dialog to save as PDF.</span>
</div>
` + buf.String() + `
</body>
</html>`))
}
