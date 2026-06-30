package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/harmey/blok2ttrpg/ability-builder/internal/config"
	"github.com/harmey/blok2ttrpg/ability-builder/internal/models"
	"github.com/harmey/blok2ttrpg/ability-builder/internal/session"
	"github.com/harmey/blok2ttrpg/ability-builder/internal/storage"
)

// App holds shared dependencies for all handlers.
type App struct {
	Store       *storage.Store
	Sessions    *session.Manager
	TemplateDir string
	funcMap     template.FuncMap
	Config      *config.Config
}

// templateFuncs returns custom template functions.
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"seq": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i
			}
			return s
		},
		"add": func(a, b int) int {
			return a + b
		},
		"slice": func(args ...int) []int {
			return args
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"contains": func(s []string, v string) bool {
			for _, item := range s {
				if item == v {
					return true
				}
			}
			return false
		},
		"profLabel": func(p models.Proficiency) string {
			dice := models.ProficiencyDice[p]
			return string(p) + " (" + dice + ")"
		},
		"join": func(s []string, sep string) string {
			return strings.Join(s, sep)
		},
		// dict creates a map from key-value pairs for passing to sub-templates
		"dict": func(values ...interface{}) map[string]interface{} {
			if len(values)%2 != 0 {
				return nil
			}
			d := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					continue
				}
				d[key] = values[i+1]
			}
			return d
		},
		"printf": func(format string, args ...interface{}) string {
			return fmt.Sprintf(format, args...)
		},
	}
}

// NewApp creates a new App with parsed templates.
func NewApp(store *storage.Store, sessions *session.Manager, templateDir string, cfg *config.Config) (*App, error) {
	app := &App{
		Store:       store,
		Sessions:    sessions,
		TemplateDir: templateDir,
		funcMap:     templateFuncs(),
		Config:      cfg,
	}

	// Verify templates can be parsed
	_, err := app.parsePageTemplate("index.html")
	if err != nil {
		return nil, err
	}

	return app, nil
}

// parsePageTemplate parses a page template with the layout and all partials.
func (app *App) parsePageTemplate(pageName string) (*template.Template, error) {
	layoutPath := filepath.Join(app.TemplateDir, "layout.html")
	pagePath := filepath.Join(app.TemplateDir, pageName)
	partialsPattern := filepath.Join(app.TemplateDir, "partials", "*.html")

	tmpl, err := template.New("").Funcs(app.funcMap).ParseGlob(partialsPattern)
	if err != nil {
		return nil, err
	}

	tmpl, err = tmpl.ParseFiles(layoutPath, pagePath)
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

// parsePartialTemplate parses a single partial template.
func (app *App) parsePartialTemplate(partialName string) (*template.Template, error) {
	partialsPattern := filepath.Join(app.TemplateDir, "partials", "*.html")

	tmpl, err := template.New("").Funcs(app.funcMap).ParseGlob(partialsPattern)
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

// render renders a full page template with layout.
func (app *App) render(w http.ResponseWriter, pageName string, data interface{}) {
	tmpl, err := app.parsePageTemplate(pageName)
	if err != nil {
		log.Printf("template parse error for %s: %v", pageName, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("template execute error for %s: %v", pageName, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// renderPartial renders a partial template (no layout wrapper).
func (app *App) renderPartial(w http.ResponseWriter, partialName string, data interface{}) {
	tmpl, err := app.parsePartialTemplate(partialName)
	if err != nil {
		log.Printf("partial template parse error for %s: %v", partialName, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, partialName, data); err != nil {
		log.Printf("partial template execute error for %s: %v", partialName, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// extractPathParam extracts a path segment from the URL.
// For a path like /characters/{id}/abilities, extractPathParam(r, "characters", 1)
// returns the segment after "characters".
func extractPathParam(r *http.Request, after string, offset int) string {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	for i, p := range parts {
		if p == after && i+offset < len(parts) {
			return parts[i+offset]
		}
	}
	return ""
}
