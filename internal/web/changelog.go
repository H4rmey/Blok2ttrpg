package web

import (
	"bytes"
	"html/template"
	"net/http"
	"os"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
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
	var buf bytes.Buffer
	gm := goldmark.New(goldmark.WithExtensions(extension.Table))
	if err := gm.Convert(md, &buf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := struct {
		Title   string
		Content template.HTML
	}{a.Cfg.Title + " - Changelog", template.HTML(buf.String())}
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
