package handlers

import (
	"net/http"

	"github.com/harmey/blok2ttrpg/ability-builder/internal/export"
	"github.com/harmey/blok2ttrpg/ability-builder/internal/models"
)

// AbilityListHandler shows all abilities for a character.
func (app *App) AbilityListHandler(w http.ResponseWriter, r *http.Request) {
	charID := extractPathParam(r, "characters", 1)
	if charID == "" {
		http.NotFound(w, r)
		return
	}

	c, err := app.Store.GetCharacter(charID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	app.render(w, "abilities.html", map[string]interface{}{
		"Character": c,
		"Abilities": c.Abilities,
	})
}

// AbilityDetailHandler shows a single ability with YAML output.
func (app *App) AbilityDetailHandler(w http.ResponseWriter, r *http.Request) {
	charID := extractPathParam(r, "characters", 1)
	abilityID := extractPathParam(r, "abilities", 1)

	if charID == "" || abilityID == "" {
		http.NotFound(w, r)
		return
	}

	c, err := app.Store.GetCharacter(charID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	ability, err := app.Store.GetAbility(charID, abilityID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	yamlOutput := export.ToYAML(ability)

	app.render(w, "review.html", map[string]interface{}{
		"Character":   c,
		"Ability":     ability,
		"YAML":        yamlOutput,
		"CharacterID": charID,
		"TotalCost":   ability.TotalCost(),
		"IsDetail":    true,
	})
}

// DeleteAbilityHandler removes an ability from a character.
func (app *App) DeleteAbilityHandler(w http.ResponseWriter, r *http.Request) {
	charID := extractPathParam(r, "characters", 1)
	abilityID := extractPathParam(r, "abilities", 1)

	if charID == "" || abilityID == "" {
		http.NotFound(w, r)
		return
	}

	if err := app.Store.DeleteAbility(charID, abilityID); err != nil {
		http.Error(w, "Failed to delete ability", http.StatusInternalServerError)
		return
	}

	// For HTMX requests, return empty content to remove the card
	if r.Header.Get("HX-Request") == "true" {
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/characters/"+charID+"/abilities", http.StatusSeeOther)
}

// EditAbilityHandler loads an existing ability into the session and redirects
// to the builder. The BuilderHandler picks up the in-progress ability and
// hydrates the form via the InitialStateJSON payload.
func (app *App) EditAbilityHandler(w http.ResponseWriter, r *http.Request) {
	charID := extractPathParam(r, "characters", 1)
	abilityID := extractPathParam(r, "abilities", 1)

	if charID == "" || abilityID == "" {
		http.NotFound(w, r)
		return
	}

	c, err := app.Store.GetCharacter(charID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	ability, err := app.Store.GetAbility(charID, abilityID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Remove the old ability from storage; we will re-save it when the
	// builder submits. This avoids stale duplicates when the user saves the
	// edited copy.
	if err := app.Store.DeleteAbility(charID, abilityID); err != nil {
		http.Error(w, "Failed to start edit", http.StatusInternalServerError)
		return
	}

	sessID, state := app.Sessions.GetOrCreate(w, r)
	state.CharacterID = charID
	state.Ability = *ability
	app.Sessions.Update(sessID, state)

	app.render(w, "builder.html", map[string]interface{}{
		"Character":             c,
		"Ability":               state.Ability,
		"AbilityTypes":          models.AllAbilityTypes,
		"AllEnactmentTypes":     models.AllEnactmentTypes,
		"AllInteractionTypes":   models.AllInteractionTypes,
		"CompatibleEnactsMap":   compatibleEnactsMapForTemplate(),
		"GeneralTraitNames":     models.GeneralTraitNames,
		"OffenseTraitNames":     models.OffenseTraitNames,
		"DefenseTraitNames":     models.DefenseTraitNames,
		"AllTraits":             models.AllTraitNames,
		"ReactionTriggers":      models.ReactionTriggers,
		"KnockoutOptions":       models.KnockoutOptions,
		"DirectionOptions":      models.DirectionOptions,
		"ShiftDirectionOptions": models.ShiftDirectionOptions,
		"TriggerTimings":        models.TriggerTimings,
		"AoETimings":            models.AoETriggerTimings,
		"PersistentEffectTypes": models.PersistentEffectTypes,
		"DamageDiceOptions":     models.DamageDiceOptions,
		"GenericDieOptions":     models.GenericDieOptions,
		"InitialStateJSON":      buildInitialState(ability),
		"IsEdit":                true,
	})
}

// ExportAbilityYAMLHandler returns a YAML file download for an ability.
func (app *App) ExportAbilityYAMLHandler(w http.ResponseWriter, r *http.Request) {
	charID := extractPathParam(r, "characters", 1)
	abilityID := extractPathParam(r, "abilities", 1)

	if charID == "" || abilityID == "" {
		http.NotFound(w, r)
		return
	}

	ability, err := app.Store.GetAbility(charID, abilityID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	yamlOutput := export.ToYAML(ability)

	filename := "ability"
	if ability.Name != "" {
		filename = ability.Name
	}

	w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+".yaml\"")
	w.Write([]byte(yamlOutput))
}
