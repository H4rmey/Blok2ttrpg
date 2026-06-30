package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/harmey/blok2ttrpg/ability-builder/internal/models"
)

func genID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// IndexHandler shows the landing page with character list.
func (app *App) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	characters := app.Store.ListCharacters()
	app.render(w, "index.html", map[string]interface{}{
		"Characters": characters,
	})
}

// NewCharacterHandler shows an empty character sheet form.
func (app *App) NewCharacterHandler(w http.ResponseWriter, r *http.Request) {
	c := models.NewCharacter("")
	app.render(w, "character.html", map[string]interface{}{
		"Character":            c,
		"IsNew":                true,
		"GeneralTraitNames":    models.GeneralTraitNames,
		"OffenseTraitNames":    models.OffenseTraitNames,
		"DefenseTraitNames":    models.DefenseTraitNames,
		"ProficiencyOptions":   models.GetProficiencyOptions(),
		"VitalHPOptions":       models.GetVitalHPOptions(),
		"VitalMovementOptions": models.GetVitalMovementOptions(),
		"VitalEnergyOptions":   models.GetVitalEnergyOptions(),
		"TraitPointsBudget":    c.TraitPointsBudget(),
		"TraitPointsUsed":      c.TraitPointsUsed(),
	})
}

// CreateCharacterHandler handles POST to create a new character.
func (app *App) CreateCharacterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()

	c := models.NewCharacter(genID())
	populateCharacterFromForm(&c, r)

	if err := app.Store.SaveCharacter(c); err != nil {
		http.Error(w, "Failed to save character", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/characters/"+c.ID, http.StatusSeeOther)
}

// ViewCharacterHandler shows a character sheet.
func (app *App) ViewCharacterHandler(w http.ResponseWriter, r *http.Request) {
	id := extractPathParam(r, "characters", 1)
	if id == "" {
		http.NotFound(w, r)
		return
	}

	c, err := app.Store.GetCharacter(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	app.render(w, "character.html", map[string]interface{}{
		"Character":            c,
		"IsNew":                false,
		"GeneralTraitNames":    models.GeneralTraitNames,
		"OffenseTraitNames":    models.OffenseTraitNames,
		"DefenseTraitNames":    models.DefenseTraitNames,
		"ProficiencyOptions":   models.GetProficiencyOptions(),
		"VitalHPOptions":       models.GetVitalHPOptions(),
		"VitalMovementOptions": models.GetVitalMovementOptions(),
		"VitalEnergyOptions":   models.GetVitalEnergyOptions(),
		"TraitPointsBudget":    c.TraitPointsBudget(),
		"TraitPointsUsed":      c.TraitPointsUsed(),
	})
}

// UpdateCharacterHandler handles POST to update an existing character.
func (app *App) UpdateCharacterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := extractPathParam(r, "characters", 1)
	if id == "" {
		http.NotFound(w, r)
		return
	}

	c, err := app.Store.GetCharacter(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	r.ParseForm()
	populateCharacterFromForm(c, r)

	if err := app.Store.SaveCharacter(*c); err != nil {
		http.Error(w, "Failed to save character", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/characters/"+c.ID, http.StatusSeeOther)
}

// DeleteCharacterHandler handles DELETE to remove a character.
func (app *App) DeleteCharacterHandler(w http.ResponseWriter, r *http.Request) {
	id := extractPathParam(r, "characters", 1)
	if id == "" {
		http.NotFound(w, r)
		return
	}

	if err := app.Store.DeleteCharacter(id); err != nil {
		http.Error(w, "Failed to delete character", http.StatusInternalServerError)
		return
	}

	// For HTMX requests, return empty content
	if r.Header.Get("HX-Request") == "true" {
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// populateCharacterFromForm fills character fields from form data.
func populateCharacterFromForm(c *models.Character, r *http.Request) {
	if v := r.FormValue("level"); v != "" {
		if lvl, err := strconv.Atoi(v); err == nil && lvl >= 1 {
			c.Level = lvl
		}
	}
	if c.Level < 1 {
		c.Level = 1
	}
	c.Name = r.FormValue("name")
	c.Age = r.FormValue("age")
	c.Size = r.FormValue("size")
	c.Alignment = r.FormValue("alignment")
	c.Backstory = r.FormValue("backstory")
	c.Personality = r.FormValue("personality")
	c.Appearance = r.FormValue("appearance")
	c.Hobbies = r.FormValue("hobbies")
	c.Occupation = r.FormValue("occupation")
	c.Inventory = r.FormValue("inventory")
	c.Quirks = r.FormValue("quirks")

	// General traits
	for _, name := range models.GeneralTraitNames {
		if v := r.FormValue("general_" + name); v != "" {
			c.GeneralTraits[name] = models.Proficiency(v)
		}
	}

	// Offense traits
	for _, name := range models.OffenseTraitNames {
		if v := r.FormValue("offense_" + name); v != "" {
			c.OffenseTraits[name] = models.Proficiency(v)
		}
	}

	// Defense traits
	for _, name := range models.DefenseTraitNames {
		if v := r.FormValue("defense_" + name); v != "" {
			c.DefenseTraits[name] = models.Proficiency(v)
		}
	}

	// Vitals
	if v := r.FormValue("vital_hp"); v != "" {
		c.VitalHP = models.Proficiency(v)
	}
	if v := r.FormValue("vital_movement"); v != "" {
		c.VitalMovement = models.Proficiency(v)
	}
	if v := r.FormValue("vital_energy"); v != "" {
		c.VitalEnergy = models.Proficiency(v)
	}
}
