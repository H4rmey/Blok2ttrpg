package web

import (
	"html/template"
	"net/http"
	"os"

	"github.com/harmey/blok2ttrpg-v5/internal/docs"
)

// changelogPath is the markdown source rendered by the /changelog page. It is
// read from disk on every request rather than embedded, so editing the file
// does not require a rebuild - the same way the config and templates behave.
const changelogPath = "CHANGELOG.md"

// handleChangelog renders CHANGELOG.md as an HTML page. The markdown is the
// single source of truth: the repository file and the in-app page can never
// disagree because there is only one copy.
func (a *App) handleChangelog(w http.ResponseWriter, r *http.Request) {
	md, err := os.ReadFile(changelogPath)
	if err != nil {
		// A missing changelog is not a server fault worth a 500: say so in the
		// page so a deployment that forgot to ship the file is obvious.
		md = []byte("# Changelog\n\nNo changelog is available in this deployment.\n")
	}
	// Converted through the docs package so the changelog's heading anchors are
	// generated exactly like the rulebook's, and so its sidebar outline is read
	// back out of the same rendered HTML the reader sees.
	html, err := docs.MarkdownPage(md)
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
	}{a.Cfg.Title + " - Changelog", template.HTML(html), docs.Tree(docs.OutlineHTML(html)), "/changelog/markdown"}
	if err := a.Tmpl.ExecuteTemplate(w, "changelog.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleChangelogMarkdown serves the raw markdown, for anyone who would rather
// read or diff the source than the rendered page.
func (a *App) handleChangelogMarkdown(w http.ResponseWriter, r *http.Request) {
	md, err := os.ReadFile(changelogPath)
	if err != nil {
		http.Error(w, "changelog not available", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Write(md)
}
