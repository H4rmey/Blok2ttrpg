package web

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/harmey/blok2ttrpg-v5/internal/docs"
	"github.com/harmey/blok2ttrpg-v5/internal/engine"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// handleDocs renders the configuration-driven documentation as an HTML page
// with a Print / Save-as-PDF button (no external PDF dependency).
func (a *App) handleDocs(w http.ResponseWriter, r *http.Request) {
	// The app's content library is passed in so the generated rulebook lists
	// the real classes, races, backgrounds and items it ships with.
	//
	// The outline comes back alongside the HTML rather than being derived in the
	// template, because it is read out of the rendered output: that guarantees
	// the sidebar's links are exactly the anchors the page actually has.
	html, outline, err := docs.RenderHTMLOutline(a.Cfg, a.Library)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := struct {
		Title       string
		Content     template.HTML
		Outline     []docs.Children
		MarkdownURL string
	}{a.Cfg.Title + " - Documentation", template.HTML(html), outline, "/docs/markdown"}
	if err := a.Tmpl.ExecuteTemplate(w, "docs.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleDocsMarkdown downloads the docs as a markdown file.
func (a *App) handleDocsMarkdown(w http.ResponseWriter, r *http.Request) {
	md, err := docs.RenderMarkdown(a.Cfg, a.Library)
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
//
// The envelope carries everything the on-screen sheet shows, not just the raw
// character: the derived budgets and vitals, and every perk paired with its
// computed cost and generated instruction text. A printed sheet is read away
// from the app, so it has to stand on its own.
func (a *App) renderCharacterPDF(w http.ResponseWriter, c model.Character) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	perks := make([]perkSummary, 0, len(c.Perks))
	for _, ab := range c.Perks {
		perks = append(perks, perkSummary{
			Perk:         ab,
			Cost:         engine.PerkCost(a.Cfg.Config, ab),
			Instructions: engine.PerkInstructions(a.Cfg.Config, ab),
		})
	}
	data := pageData{
		Cfg:       a.Cfg.Config,
		Title:     c.Name(),
		Character: &c,
		charStats: a.characterStats(&c),
		Perks:     perks,
		// The printed sheet has no form for inputs to bind to, so every figure
		// renders as plain text.
		ReadOnlyStats: true,
	}
	if err := a.Tmpl.ExecuteTemplate(w, "character_pdf.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var _ = fmt.Sprintf
