// Package web wires the HTTP layer: templates, routing and handlers. It is
// intentionally thin; all game rules live in the config and engine.
package web

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/store"
)

// App holds the shared dependencies for all handlers.
type App struct {
	Cfg   *config.Loaded
	Store *store.Store
	Tmpl  *template.Template
}

// NewApp parses templates and returns a ready App.
func NewApp(cfg *config.Loaded, st *store.Store, templateDir string) (*App, error) {
	tmpl, err := template.New("").Funcs(funcMap()).ParseGlob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		return nil, err
	}
	return &App{Cfg: cfg, Store: st, Tmpl: tmpl}, nil
}

// Router builds the HTTP mux for the app.
func (a *App) Router() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/static/", noCache(http.StripPrefix("/static/", http.FileServer(http.Dir("static")))))

	mux.HandleFunc("/", a.handleIndex)
	mux.HandleFunc("/characters/new", a.handleNewCharacter)
	mux.HandleFunc("/characters/create", a.handleCreateCharacter)
	mux.HandleFunc("/characters/import", a.handleImportCharacter)

	// /characters/{id}/...  dispatched in handleCharacter.
	mux.HandleFunc("/characters/", a.handleCharacter)

	// Builder partials (HTMX).
	mux.HandleFunc("/builder/enactment", a.handleBuilderEnactment)
	mux.HandleFunc("/builder/enactment-fields", a.handleEnactmentFields)
	mux.HandleFunc("/builder/interaction-fields", a.handleInteractionFields)
	mux.HandleFunc("/builder/inline-fields", a.handleInlineFields)

	mux.HandleFunc("/builder/cost", a.handleBuilderCost)
	mux.HandleFunc("/builder/autosave", a.handleBuilderAutosave)

	mux.HandleFunc("/builder/ability-type-fields", a.handleAbilityTypeFields)

	// Docs.
	mux.HandleFunc("/docs", a.handleDocs)
	mux.HandleFunc("/docs/markdown", a.handleDocsMarkdown)

	return mux
}

func noCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, must-revalidate")
		h.ServeHTTP(w, r)
	})
}
