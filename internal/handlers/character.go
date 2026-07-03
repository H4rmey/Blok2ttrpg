package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/harmey/blok2ttrpg/ability-builder/internal/config"
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
	cfg := app.Config.AbilityBuilder
	traits := cfg.Traits
	c := models.NewCharacter("", traits.General, traits.Offense, traits.Defense)
	c.CurrentHP = vitalValueFromConfig(cfg.Proficiencies, c.VitalHP, "hp")
	c.CurrentEnergy = vitalValueFromConfig(cfg.Proficiencies, c.VitalEnergy, "energy")
	app.render(w, "character.html", map[string]interface{}{
		"Character":            c,
		"IsNew":                true,
		"GeneralTraitNames":    traits.General,
		"OffenseTraitNames":    traits.Offense,
		"DefenseTraitNames":    traits.Defense,
		"ProficiencyOptions":   proficiencyOptionsFromConfig(cfg.Proficiencies),
		"VitalHPOptions":       vitalOptionsFromConfig(cfg.Proficiencies, "hp"),
		"VitalMovementOptions": vitalOptionsFromConfig(cfg.Proficiencies, "movement"),
		"VitalEnergyOptions":   vitalOptionsFromConfig(cfg.Proficiencies, "energy"),
		"TraitPointsBudget":    traitPointsBudget(c.Level, cfg.Leveling),
		"TraitPointsUsed":      traitPointsUsed(c, cfg.Proficiencies),
		"CharacterHP":          vitalValueFromConfig(cfg.Proficiencies, c.VitalHP, "hp"),
		"CharacterMovement":    vitalValueFromConfig(cfg.Proficiencies, c.VitalMovement, "movement"),
		"CharacterEnergy":      vitalValueFromConfig(cfg.Proficiencies, c.VitalEnergy, "energy"),
		"MaxLevel":             cfg.Leveling.MaxLevel,
	})
}

// CreateCharacterHandler handles POST to create a new character.
func (app *App) CreateCharacterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()

	cfg := app.Config.AbilityBuilder
	traits := cfg.Traits
	c := models.NewCharacter(genID(), traits.General, traits.Offense, traits.Defense)
	populateCharacterFromForm(&c, r, traits.General, traits.Offense, traits.Defense, cfg.Leveling.MaxLevel)
	clampCurrentVitals(&c, cfg.Proficiencies)

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

	cfg := app.Config.AbilityBuilder
	traits := cfg.Traits
	app.render(w, "character.html", map[string]interface{}{
		"Character":            c,
		"IsNew":                false,
		"GeneralTraitNames":    traits.General,
		"OffenseTraitNames":    traits.Offense,
		"DefenseTraitNames":    traits.Defense,
		"ProficiencyOptions":   proficiencyOptionsFromConfig(cfg.Proficiencies),
		"VitalHPOptions":       vitalOptionsFromConfig(cfg.Proficiencies, "hp"),
		"VitalMovementOptions": vitalOptionsFromConfig(cfg.Proficiencies, "movement"),
		"VitalEnergyOptions":   vitalOptionsFromConfig(cfg.Proficiencies, "energy"),
		"TraitPointsBudget":    traitPointsBudget(c.Level, cfg.Leveling),
		"TraitPointsUsed":      traitPointsUsed(*c, cfg.Proficiencies),
		"CharacterHP":          vitalValueFromConfig(cfg.Proficiencies, c.VitalHP, "hp"),
		"CharacterMovement":    vitalValueFromConfig(cfg.Proficiencies, c.VitalMovement, "movement"),
		"CharacterEnergy":      vitalValueFromConfig(cfg.Proficiencies, c.VitalEnergy, "energy"),
		"MaxLevel":             cfg.Leveling.MaxLevel,
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


	traits := app.Config.AbilityBuilder.Traits
	maxLevel := app.Config.AbilityBuilder.Leveling.MaxLevel
	r.ParseForm()
	populateCharacterFromForm(c, r, traits.General, traits.Offense, traits.Defense, maxLevel)
	clampCurrentVitals(c, app.Config.AbilityBuilder.Proficiencies)

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
func populateCharacterFromForm(c *models.Character, r *http.Request, general, offense, defense []string, maxLevel int) {
	if v := r.FormValue("level"); v != "" {
		if lvl, err := strconv.Atoi(v); err == nil && lvl >= 1 {
			c.Level = lvl
		}
	}
	if c.Level < 1 {
		c.Level = 1
	}
	if maxLevel > 0 && c.Level > maxLevel {
		c.Level = maxLevel
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
	for _, name := range general {
		if v := r.FormValue("general_" + name); v != "" {
			c.GeneralTraits[name] = models.Proficiency(v)
		}
	}

	// Offense traits
	for _, name := range offense {
		if v := r.FormValue("offense_" + name); v != "" {
			c.OffenseTraits[name] = models.Proficiency(v)
		}
	}

	// Defense traits
	for _, name := range defense {
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

	// Current HP/Energy
	if v := r.FormValue("current_hp"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			c.CurrentHP = n
		}
	}
	if v := r.FormValue("current_energy"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			c.CurrentEnergy = n
		}
	}
}



// vitalValueFromConfig returns the numeric value for a vital proficiency from YAML config.
func vitalValueFromConfig(profs []config.ProficiencyConfig, prof models.Proficiency, vital string) int {
	for _, p := range profs {
		if models.Proficiency(p.Name) == prof {
			if v, ok := p.Vitals[vital]; ok {
				return v
			}
		}
	}
	return 0
}

// clampCurrentVitals ensures current HP/Energy do not exceed their totals.
func clampCurrentVitals(c *models.Character, profs []config.ProficiencyConfig) {
	if max := vitalValueFromConfig(profs, c.VitalHP, "hp"); c.CurrentHP > max {
		c.CurrentHP = max
	}
	if c.CurrentHP < 0 {
		c.CurrentHP = 0
	}
	if max := vitalValueFromConfig(profs, c.VitalEnergy, "energy"); c.CurrentEnergy > max {
		c.CurrentEnergy = max
	}
	if c.CurrentEnergy < 0 {
		c.CurrentEnergy = 0
	}
}
// proficiencyOptionsFromConfig builds proficiency dropdown options from YAML config.
func proficiencyOptionsFromConfig(profs []config.ProficiencyConfig) []models.ProficiencyOption {
	out := make([]models.ProficiencyOption, 0, len(profs))
	for i, p := range profs {
		out = append(out, models.ProficiencyOption{
			Value:    models.Proficiency(p.Name),
			Label:    p.Name + " (" + p.Dice.General + ")",
			DiceTier: i + 1,
		})
	}
	return out
}

// vitalOptionsFromConfig builds vital stat dropdown options from YAML config.
func vitalOptionsFromConfig(profs []config.ProficiencyConfig, vital string) []models.VitalOption {
	out := make([]models.VitalOption, 0, len(profs))
	for _, p := range profs {
		value, ok := p.Vitals[vital]
		if !ok {
			continue
		}
		var suffix string
		switch vital {
		case "hp":
			suffix = " HP"
		case "movement":
			suffix = "m"
		case "energy":
			suffix = ""
		}
		out = append(out, models.VitalOption{
			Value: models.Proficiency(p.Name),
			Label: p.Name + " (" + strconv.Itoa(value) + suffix + ")",
		})
	}
	return out
}

// traitPointsBudget returns the total trait points for a level based on YAML config.
func traitPointsBudget(level int, cfg config.LevelingConfig) int {
	if level < 1 {
		level = 1
	}
	for i := len(cfg.TraitPoints.Levels) - 1; i >= 0; i-- {
		if cfg.TraitPoints.Levels[i].Level <= level {
			return cfg.TraitPoints.Levels[i].Total
		}
	}
	return 0
}

// traitPointsUsed returns the total trait points spent using YAML config costs.
func traitPointsUsed(c models.Character, profs []config.ProficiencyConfig) int {
	costs := make(map[models.Proficiency]int)
	for _, p := range profs {
		costs[models.Proficiency(p.Name)] = p.Cost
	}
	total := 0
	for _, p := range c.GeneralTraits {
		total += costs[p]
	}
	for _, p := range c.OffenseTraits {
		total += costs[p]
	}
	for _, p := range c.DefenseTraits {
		total += costs[p]
	}
	total += costs[c.VitalHP]
	total += costs[c.VitalMovement]
	total += costs[c.VitalEnergy]
	return total
}
