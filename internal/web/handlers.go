package web

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/engine"
	"github.com/harmey/blok2ttrpg-v5/internal/export"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// pageData is the common data envelope for full-page templates.
type pageData struct {
	Cfg         *config.Config
	Title       string
	Breadcrumbs []crumb
	Character   *model.Character
	Characters  []model.Character

	// Budget summary for the character sheet.
	TraitBudget   int
	TraitUsed     int
	AbilityBudget int
	AbilityUsed   int
}

type crumb struct {
	Label string
	URL   string
}

func (a *App) render(w http.ResponseWriter, name string, data pageData) {
	data.Cfg = a.Cfg.Config
	if data.Title == "" {
		data.Title = a.Cfg.Title
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.Tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	a.render(w, "index.html", pageData{
		Title:       a.Cfg.Title,
		Characters:  a.Store.List(),
		Breadcrumbs: []crumb{{Label: "Home", URL: "/"}},
	})
}

func (a *App) handleNewCharacter(w http.ResponseWriter, r *http.Request) {
	c := a.blankCharacter("")
	a.render(w, "character.html", a.characterPage(&c, true))
}

func (a *App) handleCreateCharacter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := fmt.Sprintf("char-%d", time.Now().UnixNano())
	c := a.blankCharacter(id)
	a.applyCharacterForm(&c, r)
	if err := a.Store.Save(c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/characters/"+id, http.StatusSeeOther)
}

// handleCharacter dispatches all /characters/{id}[/action] routes.
func (a *App) handleCharacter(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/characters/")
	parts := strings.Split(path, "/")
	if parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	id := parts[0]
	c, ok := a.Store.Get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			a.render(w, "character.html", a.characterPage(&c, false))
		case http.MethodPost:
			a.applyCharacterForm(&c, r)
			if err := a.Store.Save(c); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/characters/"+id, http.StatusSeeOther)
		case http.MethodDelete:
			_ = a.Store.Delete(id)
			w.Header().Set("HX-Redirect", "/")
		}
		return
	}

	switch parts[1] {
	case "delete":
		_ = a.Store.Delete(id)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	case "export":
		b, err := export.MarshalCharacter(c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/x-yaml")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", c.Name()+".yaml"))
		w.Write(b)
	case "pdf":
		a.renderCharacterPDF(w, c)
	case "abilities":
		a.handleAbilities(w, r, &c, parts[2:])
	default:
		http.NotFound(w, r)
	}
}

func (a *App) handleImportCharacter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "no file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	c, err := export.UnmarshalCharacter(data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if c.ID == "" {
		c.ID = fmt.Sprintf("char-%d", time.Now().UnixNano())
	}
	if err := a.Store.Save(c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/characters/"+c.ID, http.StatusSeeOther)
}

// blankCharacter builds a character with defaults for every configured trait.
func (a *App) blankCharacter(id string) model.Character {
	c := model.Character{
		ID:         id,
		Level:      1,
		Attributes: map[string]any{},
		Traits:     map[string]string{},
		Abilities:  []model.Ability{},
	}
	def := a.Cfg.DefaultProficiencyID()
	for _, g := range a.Cfg.Traits.List() {
		for _, t := range g.Traits {
			c.Traits[model.TraitKey(g.ID, t)] = def
		}
	}

	return c
}

// applyCharacterForm reads posted form fields into the generic character maps.
func (a *App) applyCharacterForm(c *model.Character, r *http.Request) {
	_ = r.ParseForm()
	if lvl := r.FormValue("level"); lvl != "" {
		if n, err := strconv.Atoi(lvl); err == nil && n >= 1 {
			c.Level = n
		}
	}
	for _, g := range a.Cfg.Attributes.List() {
		for _, f := range g.Fields {
			name := "attr_" + f.Key
			if _, ok := r.Form[name]; ok {
				c.Attributes[f.Key] = r.FormValue(name)
			}
		}
	}
	for _, g := range a.Cfg.Traits.List() {

		for _, t := range g.Traits {
			name := "trait_" + g.ID + "_" + t
			if v := r.FormValue(name); v != "" {
				c.Traits[model.TraitKey(g.ID, t)] = v
			}
		}
	}
}

func (a *App) characterPage(c *model.Character, isNew bool) pageData {
	title := c.Name()
	if isNew {
		title = "New Character"
	}
	// Sum the build cost of every ability the character owns.
	abilityUsed := 0
	for _, ab := range c.Abilities {
		abilityUsed += engine.AbilityCost(a.Cfg.Config, ab).Build
	}
	return pageData{
		Title:     title,
		Character: c,
		Breadcrumbs: []crumb{
			{Label: "Home", URL: "/"},
			{Label: title, URL: "/characters/" + c.ID},
		},
		TraitBudget:   a.Cfg.TraitPointBudget(c.Level),
		TraitUsed:     engine.TraitPointsUsed(a.Cfg.Config, *c),
		AbilityBudget: a.Cfg.AbilityPointBudget(c.Level),
		AbilityUsed:   abilityUsed,
	}
}
