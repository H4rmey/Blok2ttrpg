// Package web wires the HTTP layer: templates, routing and handlers. It is
// intentionally thin; all game rules live in the config and engine.
package web

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/premade"
	"github.com/harmey/blok2ttrpg-v5/internal/store"
)

// App holds the shared dependencies for all handlers.
type App struct {
	Cfg     *config.Loaded
	Store   *store.Store
	Tmpl    *template.Template
	Library *premade.Library
}

// NewApp parses templates and returns a ready App. libraryRoot points at the
// built-in content library (packages and abilities) that ships with the app.
func NewApp(cfg *config.Loaded, st *store.Store, templateDir, libraryRoot string) (*App, error) {
	tmpl, err := template.New("").Funcs(funcMap()).ParseGlob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		return nil, err
	}
	return &App{Cfg: cfg, Store: st, Tmpl: tmpl, Library: premade.New(libraryRoot)}, nil
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
	mux.HandleFunc("/builder/condition-shift", a.handleConditionShift)

	mux.HandleFunc("/builder/cost", a.handleBuilderCost)
	mux.HandleFunc("/builder/autosave", a.handleBuilderAutosave)

	mux.HandleFunc("/builder/ability-type-fields", a.handleAbilityTypeFields)

	// Package and ability library browsers (built-in content).
	mux.HandleFunc("/packages/library", a.handlePackageLibrary)
	mux.HandleFunc("/abilities/library", a.handleAbilityLibrary)

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
