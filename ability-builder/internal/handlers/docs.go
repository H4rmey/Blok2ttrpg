package handlers

import (
	"log"
	"net/http"

	"github.com/harmey/blok2ttrpg/ability-builder/internal/docs"
)

// DownloadDocsHandler generates a single markdown file from the docs templates
// and the YAML configuration, then returns it as a downloadable attachment.
func (app *App) DownloadDocsHandler(w http.ResponseWriter, r *http.Request) {
	cfg := app.Config
	if cfg == nil {
		http.Error(w, "Configuration not loaded", http.StatusInternalServerError)
		return
	}

	merged, err := docs.Render(cfg, docs.DefaultDir())
	if err != nil {
		log.Printf("failed to render docs: %v", err)
		http.Error(w, "Failed to render documentation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"blok2-ability-builder-docs.md\"")
	if _, err := w.Write([]byte(merged)); err != nil {
		log.Printf("failed to write docs response: %v", err)
	}
}
