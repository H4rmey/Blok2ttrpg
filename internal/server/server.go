package server

import (
	"html/template"
	"io/fs"
	"net/http"
	"os"

	"github.com/blok2ttrpg/charsheet/internal/server/tmplfuncs"
	"github.com/blok2ttrpg/charsheet/web"
)

// Server holds the HTTP handler, templates, and session state.
type Server struct {
	mux      *http.ServeMux
	tmpl     *template.Template
	sessions *SessionStore
}

// New creates a new Server with routes and templates configured.
func New() (*Server, error) {
	s := &Server{
		mux:      http.NewServeMux(),
		sessions: NewSessionStore(),
	}

	// Use live filesystem in dev mode, embedded FS in production
	var templateFS fs.FS
	var staticFS fs.FS
	if os.Getenv("DEV") != "" {
		templateFS = os.DirFS("web")
		staticFS = os.DirFS("web/static")
	} else {
		templateFS = web.TemplateFS
		sub, err := fs.Sub(web.StaticFS, "static")
		if err != nil {
			return nil, err
		}
		staticFS = sub
	}

	// Parse templates with custom functions
	tmpl, err := template.New("").Funcs(tmplfuncs.FuncMap()).ParseFS(templateFS,
		"templates/layouts/*.html",
		"templates/partials/*.html",
	)
	if err != nil {
		return nil, err
	}
	s.tmpl = tmpl

	// Static file serving
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Page routes
	s.mux.HandleFunc("/", s.handleIndex)

	// Character import/export routes
	s.mux.HandleFunc("/character/export", s.handleCharacterExport)
	s.mux.HandleFunc("/character/import", s.handleCharacterImport)
	s.mux.HandleFunc("/character/new", s.handleCharacterNew)

	// Tab content routes (HTMX partials)
	s.mux.HandleFunc("/tabs/attributes", s.handleTabAttributes)
	s.mux.HandleFunc("/tabs/general-traits", s.handleTabGeneralTraits)
	s.mux.HandleFunc("/tabs/combative-traits", s.handleTabCombativeTraits)
	s.mux.HandleFunc("/tabs/abilities", s.handleTabAbilities)

	// Level management routes
	s.mux.HandleFunc("/level/up", s.handleLevelUp)
	s.mux.HandleFunc("/level/down", s.handleLevelDown)

	// Attribute update routes
	s.mux.HandleFunc("/attributes/update", s.handleAttributeUpdate)
	s.mux.HandleFunc("/attributes/temp/add", s.handleTempAttributeAdd)
	s.mux.HandleFunc("/attributes/temp/remove", s.handleTempAttributeRemove)
	s.mux.HandleFunc("/attributes/custom/add", s.handleCustomFieldAdd)
	s.mux.HandleFunc("/attributes/custom/update", s.handleCustomFieldUpdate)
	s.mux.HandleFunc("/attributes/custom/remove", s.handleCustomFieldRemove)

	// Trait update routes
	s.mux.HandleFunc("/traits/general/update", s.handleGeneralTraitUpdate)
	s.mux.HandleFunc("/traits/combative/update", s.handleCombativeTraitUpdate)

	// Ability management routes
	s.mux.HandleFunc("/abilities/new", s.handleAbilityNew)
	s.mux.HandleFunc("/abilities/delete", s.handleAbilityDelete)
	s.mux.HandleFunc("/abilities/export", s.handleAbilityExport)
	s.mux.HandleFunc("/abilities/import", s.handleAbilityImport)

	// Ability wizard routes
	s.mux.HandleFunc("/abilities/edit", s.handleAbilityEdit)
	s.mux.HandleFunc("/abilities/wizard/add-enactment", s.handleAbilityWizardAddEnactment)
	s.mux.HandleFunc("/abilities/wizard/remove-enactment", s.handleAbilityWizardRemoveEnactment)
	s.mux.HandleFunc("/abilities/wizard/add-perk", s.handleAbilityWizardAddPerk)
	s.mux.HandleFunc("/abilities/wizard/remove-perk", s.handleAbilityWizardRemovePerk)
	s.mux.HandleFunc("/abilities/wizard/set-interaction", s.handleAbilityWizardSetInteraction)
	s.mux.HandleFunc("/abilities/wizard/update-validation", s.handleAbilityWizardUpdateValidation)
	s.mux.HandleFunc("/abilities/wizard/add-interaction-perk", s.handleAbilityWizardAddInteractionPerk)
	s.mux.HandleFunc("/abilities/wizard/remove-interaction-perk", s.handleAbilityWizardRemoveInteractionPerk)
	s.mux.HandleFunc("/abilities/wizard/toggle-optional", s.handleAbilityWizardToggleOptional)
	s.mux.HandleFunc("/abilities/wizard/update-effect-flavor", s.handleAbilityWizardUpdateEffectFlavor)
	s.mux.HandleFunc("/abilities/wizard/back", s.handleAbilityWizardBack)

	return s, nil
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	_, char := s.sessions.GetOrCreate(w, r)
	s.renderTemplate(w, "base.html", char)
}

func (s *Server) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	err := s.tmpl.ExecuteTemplate(w, name, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
