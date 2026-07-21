package web

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/harmey/blok2ttrpg-v5/internal/docs"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// handleDocs renders the configuration-driven documentation as an HTML page
// with a Print / Save-as-PDF button (no external PDF dependency).
func (a *App) handleDocs(w http.ResponseWriter, r *http.Request) {
	html, err := docs.RenderHTML(a.Cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := struct {
		Title   string
		Content template.HTML
	}{a.Cfg.Title + " - Documentation", template.HTML(html)}
	if err := a.Tmpl.ExecuteTemplate(w, "docs.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleDocsMarkdown downloads the docs as a markdown file.
func (a *App) handleDocsMarkdown(w http.ResponseWriter, r *http.Request) {
	md, err := docs.RenderMarkdown(a.Cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="documentation.md"`)
	w.Write([]byte(md))
}

// renderCharacterPDF renders a print-friendly character sheet page which the
// browser can save as PDF via window.print(). This keeps the app dependency
// free (no Node/puppeteer).
func (a *App) renderCharacterPDF(w http.ResponseWriter, c model.Character) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := pageData{
		Cfg:       a.Cfg.Config,
		Title:     c.Name(),
		Character: &c,
	}
	if err := a.Tmpl.ExecuteTemplate(w, "character_pdf.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var _ = fmt.Sprintf
