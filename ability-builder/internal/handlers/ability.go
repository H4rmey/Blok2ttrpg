package handlers

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/harmey/blok2ttrpg/ability-builder/internal/config"
	"github.com/harmey/blok2ttrpg/ability-builder/internal/export"
	"github.com/harmey/blok2ttrpg/ability-builder/internal/models"
)

// BuilderHandler shows the ability builder form for a character.
//
// If a session already holds a draft ability (the edit flow sets it on the
// session), that ability is rendered with its cards pre-populated. Otherwise a
// blank ability is shown.
func (app *App) BuilderHandler(w http.ResponseWriter, r *http.Request) {
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

	sessID, state := app.Sessions.GetOrCreate(w, r)
	if state.Ability.ID == "" {
		state.CharacterID = charID
		state.Ability = models.Ability{}
		app.Sessions.Update(sessID, state)
	}

	initialState := buildInitialState(&state.Ability)

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
		"AllTraits":             models.AllTraitNames(),
		"ReactionTriggers":      models.ReactionTriggers,
		"KnockoutOptions":       models.KnockoutOptions,
		"DirectionOptions":      models.DirectionOptions,
		"ShiftDirectionOptions": models.ShiftDirectionOptions,
		"TriggerTimings":        models.TriggerTimings,
		"AoETimings":            models.AoETriggerTimings,
		"PersistentEffectTypes": models.PersistentEffectTypes,
		"DamageDiceOptions":     models.DamageDiceOptions,
		"GenericDieOptions":     models.GenericDieOptions,
		"InitialStateJSON":      initialState,
		"IsEdit":                state.Ability.ID != "",
		"BuilderConfigJSON":     mustMarshalConfig(app.Config),
	})
}

// compatibleEnactsMapForTemplate returns the compatible-enacts mapping in a
// form that is easy to render as a JS object literal.
func compatibleEnactsMapForTemplate() map[string][]string {
	out := map[string][]string{}
	for _, at := range models.AllAbilityTypes {
		for _, et := range models.CompatibleEnactments[at] {
			out[string(at)] = append(out[string(at)], string(et))
		}
	}
	return out
}

// buildInitialState serialises the in-progress ability (or a stored one) into
// a JSON blob that the JS reads on boot to re-hydrate the form.
func buildInitialState(a *models.Ability) string {
	s := map[string]interface{}{
		"name":                 a.Name,
		"description":          a.Description,
		"ability_type":         a.Type,
		"item_dep":             a.HasItemDependency,
		"item_name":            a.ItemName,
		"energy_steps":         a.EnergySteps,
		"action_steps":         a.ActionSteps,
		"reaction_range":       a.ReactionRange,
		"reaction_uses":        a.ReactionUses,
		"trigger":              a.Trigger,
		"trigger_trait":        a.TriggerTrait,
		"phase_duration":       a.PhaseDuration,
		"phase_reverse_rounds": a.ReversePhaseRounds,
		"all_knockouts_req":    a.AllKnockoutsReq,
		"reverse_knockout_ok":  a.ReverseKnockoutOK,
		"no_knockout":          a.NoKnockout,
		"knockouts":            a.Knockouts,
		"hp_bonus":             a.HPBonus,
		"extra_lifetime":       a.ExtraLifetime,
		"enactments":           []map[string]interface{}{},
	}
	for _, e := range a.Enactments {
		em := map[string]interface{}{
			"name":            e.Name,
			"type":            string(e.Type),
			"always":          e.Always,
			"source":          e.Source,
			"source_category": string(e.SourceCategory),
			"source_trait":    e.SourceTrait,
			"other_roll_text": e.OtherRollText,
			"flat_bonus":      e.FlatBonus,
			"offensive_trait": e.OffensiveTrait,
			"medicine_trait":  e.MedicineTrait,
			"origin_mode":     e.OriginMode,
			"origin_text":     e.OriginText,
			"distance":        e.Distance,
			"directions":      e.Directions,
			"shifted_trait":   e.ShiftedTrait,
			"shift_dir":       e.ShiftDir,
			"shift_amount":    e.ShiftAmount,
			"shift_uses":      e.ShiftUses,
			"effect_name":     e.EffectName,
			"effect_type":     e.EffectType,
			"duration":        e.Duration,
			"trigger_timing":  e.TriggerTiming,
			"solutions":       e.Solutions,
		}
		if e.Interaction != nil {
			im := map[string]interface{}{
				"type":           string(e.Interaction.Type),
				"range":          e.Interaction.Range,
				"targets":        e.Interaction.Targets,
				"visible_ok":     e.Interaction.VisibleOK,
				"obstructed_ok":  e.Interaction.ObstructedOK,
				"remove_penalty": e.Interaction.RemovePenalty,
				"radius":         e.Interaction.Radius,
				"origin_mode":    e.Interaction.OriginMode,
				"origin_text":    e.Interaction.OriginText,
				"duration":       e.Interaction.Duration,
				"timing":         e.Interaction.Timing,
				"immune":         e.Interaction.Immune,
				"use_previous":   e.Interaction.UsePrevious,
			}
			if e.Interaction.Validation != nil {
				v := e.Interaction.Validation
				entries := []map[string]interface{}{}
				if len(v.CounterRollEntries) > 0 {
					for _, ce := range v.CounterRollEntries {
						entries = append(entries, map[string]interface{}{
							"type":  string(ce.Type),
							"trait": ce.Trait,
						})
					}
				} else {
					for _, c := range v.CounterRolls {
						entries = append(entries, map[string]interface{}{"type": "defense", "trait": c})
					}
				}
				im["validation"] = map[string]interface{}{
					"engage_mode":           string(v.EngageMode),
					"engage_trait_category": string(v.EngageTraitCategory),
					"engage_trait":          v.EngageTrait,
					"engage_die":            v.EngageDie,
					"engage_other":          v.EngageOther,
					"counter_entries":       entries,
				}
			}
			em["interaction"] = im
		}
		s["enactments"] = append(s["enactments"].([]map[string]interface{}), em)
	}
	buf := new(bytes.Buffer)
	_ = json.NewEncoder(buf).Encode(s)
	// json.Marshal wraps the result in quotes, so we strip them off. We
	// bypass the encoder here to get a clean literal.
	out, err := json.Marshal(s)
	if err != nil {
		return ""
	}
	return string(out)
}

// AbilityTypeConfigHandler returns the HTML partial for the selected ability type card.
// (Legacy route — currently unused since the new builder handles this client-side via JS.)
func (app *App) AbilityTypeConfigHandler(w http.ResponseWriter, r *http.Request) {
	abilityType := models.AbilityType(r.URL.Query().Get("ability_type"))
	if abilityType == "" {
		http.Error(w, "ability_type required", http.StatusBadRequest)
		return
	}

	data := map[string]interface{}{
		"AbilityType":       abilityType,
		"GeneralTraitNames": models.GeneralTraitNames,
		"OffenseTraitNames": models.OffenseTraitNames,
		"DefenseTraitNames": models.DefenseTraitNames,
		"AllTraits":         models.AllTraitNames(),
		"KnockoutOptions":   models.KnockoutOptions,
		"Triggers":          models.ReactionTriggers,
		"IsEdit":            false,
	}

	tmplName := abilityTypePartial(abilityType)
	if tmplName == "" {
		http.Error(w, "unknown ability type", http.StatusBadRequest)
		return
	}
	app.renderPartial(w, tmplName, data)
}

func abilityTypePartial(t models.AbilityType) string {
	switch t {
	case models.AbilityExecution:
		return "ability_type_execution.html"
	case models.AbilityReaction:
		return "ability_type_reaction.html"
	case models.AbilityPhase:
		return "ability_type_phase.html"
	case models.AbilityMinion:
		return "ability_type_minion.html"
	}
	return ""
}

// EnactmentConfigHandler is a legacy stub that returns a card partial.
// The new builder renders cards in JS, but we keep this for any
// per-enactment edits that may still use it.
func (app *App) EnactmentConfigHandler(w http.ResponseWriter, r *http.Request) {
	enactType := models.EnactmentType(r.URL.Query().Get("enactment_type"))
	indexStr := r.URL.Query().Get("index")
	index, _ := strconv.Atoi(indexStr)

	data := map[string]interface{}{
		"EnactmentType":     enactType,
		"Index":             index,
		"AllTraits":         models.AllTraitNames(),
		"OffenseTraitNames": models.OffenseTraitNames,
		"DefenseTraitNames": models.DefenseTraitNames,
		"GeneralTraitNames": models.GeneralTraitNames,
		"Directions":        models.DirectionOptions,
		"ShiftDirections":   models.ShiftDirectionOptions,
		"TriggerTimings":    models.TriggerTimings,
		"EffectTypes":       models.PersistentEffectTypes,
	}

	tmplName := enactCardPartial(enactType)
	if tmplName == "" {
		http.Error(w, "Unknown enactment type", http.StatusBadRequest)
		return
	}
	app.renderPartial(w, tmplName, data)
}

func enactCardPartial(t models.EnactmentType) string {
	switch t {
	case models.EnactDamage:
		return "enact_card_damage.html"
	case models.EnactHealing:
		return "enact_card_healing.html"
	case models.EnactMovement:
		return "enact_card_movement.html"
	case models.EnactProficiencyShift:
		return "enact_card_profshift.html"
	case models.EnactPersistentEffect:
		return "enact_card_persistent.html"
	}
	return ""
}

// InteractionConfigHandler is a legacy stub.
func (app *App) InteractionConfigHandler(w http.ResponseWriter, r *http.Request) {
	interType := models.InteractionType(r.URL.Query().Get("interaction_type"))
	indexStr := r.URL.Query().Get("index")
	idx, _ := strconv.Atoi(indexStr)

	data := map[string]interface{}{
		"InteractionType":   interType,
		"EnactmentIndex":    idx,
		"OffenseTraitNames": models.OffenseTraitNames,
		"DefenseTraitNames": models.DefenseTraitNames,
		"GeneralTraitNames": models.GeneralTraitNames,
		"AllTraits":         models.AllTraitNames(),
		"GenericDieOptions": models.GenericDieOptions,
		"AoETimings":        models.AoETriggerTimings,
	}

	tmplName := interCardPartial(interType)
	if tmplName == "" {
		http.Error(w, "Unknown interaction type", http.StatusBadRequest)
		return
	}
	app.renderPartial(w, tmplName, data)
}

func interCardPartial(t models.InteractionType) string {
	switch t {
	case models.InteractionSelf:
		return "inter_card_self.html"
	case models.InteractionDirect:
		return "inter_card_direct.html"
	case models.InteractionRanged:
		return "inter_card_ranged.html"
	case models.InteractionArea:
		return "inter_card_area.html"
	case models.InteractionAreaOfEffect:
		return "inter_card_aoe.html"
	}
	return ""
}

// AddEnactmentHandler is a legacy stub kept so old HTMX paths still work.
func (app *App) AddEnactmentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<div class="enactment-block space-y-3 p-4 border border-gray-700 rounded bg-gray-900 text-xs text-gray-500">[deprecated — builder now uses in-page JS]</div>`))
}

// SaveAbilityHandler parses the new flat-prefixed form fields produced by
// builder.js. Field naming convention:
//
//	hidden ability_type select (top of form)
//	ability_item_dep, ability_item_name
//	ability_<typed fields like phase_rounds, range, uses, hp, life, etc.>
//	enact_<idx>_field         (enact card fields, including "type")
//	enact_<idx>_inter_field   (interaction card fields, including "type")
//	enact_<idx>_valid_field   (validation card fields)
func (app *App) SaveAbilityHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	state := app.Sessions.Get(r)
	if state == nil {
		http.Error(w, "No active builder session", http.StatusBadRequest)
		return
	}

	a := &state.Ability
	a.ID = genID()
	a.Name = strings.TrimSpace(r.FormValue("ability_name"))
	a.Description = strings.TrimSpace(r.FormValue("ability_description"))
	a.Type = models.AbilityType(r.FormValue("ability_type"))
	if a.Type == "" {
		a.Type = models.AbilityType(r.FormValue("hidden-ability-type"))
	}
	a.HasItemDependency = r.FormValue("ability_item_dep") == "on"
	a.ItemName = r.FormValue("ability_item_name")

	if err := app.Config.AbilityBuilder.ComputeAbilityCosts(a, r.Form); err != nil {
		http.Error(w, "Failed to compute ability costs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Phase-specific fields (not cost-related, but part of the ability type form).
	if a.Type == models.AbilityPhase {
		a.AllKnockoutsReq = r.FormValue("all_req") == "on"
		a.ReverseKnockoutOK = r.FormValue("reverse_knockout") == "on"
		a.NoKnockout = r.FormValue("no_knockout") == "on"
		if ks := r.Form["knockout"]; len(ks) > 0 {
			a.Knockouts = ks
		}
	}

	a.Enactments = parseNewEnactments(r)

	if err := app.Store.AddAbility(state.CharacterID, *a); err != nil {
		http.Error(w, "Failed to save ability", http.StatusInternalServerError)
		return
	}

	if cookie, _ := r.Cookie("blok2_session"); cookie != nil {
		app.Sessions.Clear(cookie.Value)
	}

	http.Redirect(w, r, "/characters/"+state.CharacterID+"/abilities", http.StatusSeeOther)
}

// _ keeps atoi alias declared in case handler refactor reuses
var _ = bytes.NewBuffer

// parseNewEnactments walks the form looking for enact_<idx>_type keys.
func parseNewEnactments(r *http.Request) []models.Enactment {
	indices := map[int]bool{}
	for k := range r.Form {
		if rest, ok := stripPrefix(k, "enact_"); ok {
			if idx, ok := parseLeadingInt(rest); ok {
				indices[idx] = true
			}
		}
	}
	var sortedIdx []int
	for k := range indices {
		sortedIdx = append(sortedIdx, k)
	}
	for i := 1; i < len(sortedIdx); i++ {
		for j := i; j > 0 && sortedIdx[j-1] > sortedIdx[j]; j-- {
			sortedIdx[j-1], sortedIdx[j] = sortedIdx[j], sortedIdx[j-1]
		}
	}

	var out []models.Enactment
	for _, idx := range sortedIdx {
		prefix := "enact_" + strconv.Itoa(idx) + "_"
		enactType := r.FormValue(prefix + "type")
		if enactType == "" {
			continue
		}
		src := r.FormValue(prefix + "source")
		cat := models.TraitCategory(r.FormValue(prefix + "source_category"))
		if cat == "" && src == "trait" {
			cat = models.TraitCategoryOffense
		}

		e := models.Enactment{
			Name:           strings.TrimSpace(r.FormValue(prefix + "name")),
			Type:           models.EnactmentType(enactType),
			Always:         r.FormValue(prefix+"always") == "on",
			BuildCost:      atoi(r.FormValue(prefix + "build")),
			CastCost:       atoi(r.FormValue(prefix + "cast")),
			Formula:        r.FormValue(prefix + "formula"),
			Source:         src,
			SourceTrait:    r.FormValue(prefix + "source_trait"),
			SourceCategory: cat,
			OtherRollText:  r.FormValue(prefix + "other"),
			FlatBonus:      atoi(r.FormValue(prefix + "flat")),
			OffensiveTrait: r.FormValue(prefix + "offense"),
			MedicineTrait:  r.FormValue(prefix + "medicine"),
			OriginMode:     r.FormValue(prefix + "origin_mode"),
			OriginText:     r.FormValue(prefix + "origin_text"),
			Distance:       atoi(r.FormValue(prefix + "distance")),
			Directions:     r.Form[prefix+"direction"],
			ShiftedTrait:   r.FormValue(prefix + "shifted_trait"),
			ShiftDir:       r.FormValue(prefix + "shift_dir"),
			ShiftAmount:    atoi(r.FormValue(prefix + "shift_amount")),
			ShiftUses:      atoi(r.FormValue(prefix + "shift_uses")),
			EffectName:     r.FormValue(prefix + "effect_name"),
			EffectType:     r.FormValue(prefix + "effect_type"),
			Duration:       atoi(r.FormValue(prefix + "duration")),
			TriggerTiming:  r.FormValue(prefix + "trigger_timing"),
			Solutions:      r.Form[prefix+"solution"],
		}

		interPrefix := prefix + "inter_"
		interType := r.FormValue(interPrefix + "type")
		if interType != "" {
			inter := models.Interaction{
				Type:          models.InteractionType(interType),
				BuildCost:     atoi(r.FormValue(interPrefix + "build")),
				CastCost:      atoi(r.FormValue(interPrefix + "cast")),
				Range:         atoi(r.FormValue(interPrefix + "range")),
				Targets:       atoi(r.FormValue(interPrefix + "targets")),
				Radius:        atoi(r.FormValue(interPrefix + "radius")),
				OriginMode:    r.FormValue(interPrefix + "origin_mode"),
				OriginText:    r.FormValue(interPrefix + "origin_text"),
				Duration:      atoi(r.FormValue(interPrefix + "duration")),
				Timing:        r.FormValue(interPrefix + "timing"),
				VisibleOK:     r.FormValue(interPrefix+"visible") == "on",
				ObstructedOK:  r.FormValue(interPrefix+"obstructed") == "on",
				RemovePenalty: r.FormValue(interPrefix+"remove_penalty") == "on",
				Immune:        r.FormValue(interPrefix+"immune") == "on",
				UsePrevious:   r.FormValue(interPrefix+"use_previous") == "on",
			}
			validPrefix := prefix + "valid_"
			mode := models.EngageMode(r.FormValue(validPrefix + "engage_mode"))
			engageCat := models.TraitCategory(r.FormValue(validPrefix + "engage_trait_category"))
			inter.Validation = &models.Validation{
				BuildCost:           atoi(r.FormValue(validPrefix + "build")),
				CastCost:            atoi(r.FormValue(validPrefix + "cast")),
				EngageMode:          mode,
				EngageTrait:         r.FormValue(validPrefix + "engage_trait"),
				EngageTraitCategory: engageCat,
				EngageDie:           r.FormValue(validPrefix + "engage_die"),
				EngageOther:         r.FormValue(validPrefix + "engage_other"),
			}
			// Counter roll entries — each row has counter_type, counter_trait
			if counterTypes := r.Form[validPrefix+"counter_type"]; len(counterTypes) > 0 {
				counterTraits := r.Form[validPrefix+"counter_trait"]
				for i, t := range counterTypes {
					entry := models.CounterRoll{
						Type:  models.TraitCategory(t),
						Trait: "",
					}
					if i < len(counterTraits) {
						entry.Trait = counterTraits[i]
					}
					inter.Validation.CounterRollEntries = append(inter.Validation.CounterRollEntries, entry)
					inter.Validation.CounterRolls = append(inter.Validation.CounterRolls, entry.Trait)
				}
			}
			e.Interaction = &inter
		}

		out = append(out, e)
	}
	return out
}

func stripPrefix(s, p string) (string, bool) {
	if len(s) >= len(p) && s[:len(p)] == p {
		return s[len(p):], true
	}
	return "", false
}

func parseLeadingInt(s string) (int, bool) {
	end := 0
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}
	if end == 0 {
		return 0, false
	}
	v, err := strconv.Atoi(s[:end])
	if err != nil {
		return 0, false
	}
	return v, true
}

func atoi(s string) int {
	if s == "" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}

// ReviewHandler shows the ability review page with YAML preview.
func (app *App) ReviewHandler(w http.ResponseWriter, r *http.Request) {
	state := app.Sessions.Get(r)
	if state == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	yamlOutput := export.ToYAML(&state.Ability)

	app.render(w, "review.html", map[string]interface{}{
		"Ability":     state.Ability,
		"YAML":        yamlOutput,
		"CharacterID": state.CharacterID,
		"TotalCost":   state.Ability.TotalCost(),
	})
}

// ResetHandler clears the builder session.
func (app *App) ResetHandler(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("blok2_session")
	if cookie != nil {
		state := app.Sessions.Get(r)
		charID := ""
		if state != nil {
			charID = state.CharacterID
		}
		app.Sessions.Clear(cookie.Value)
		if charID != "" {
			http.Redirect(w, r, "/characters/"+charID+"/abilities/new", http.StatusSeeOther)
			return
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func mustMarshalConfig(cfg *config.Config) template.JS {
	if cfg == nil {
		return "{}"
	}
	b, err := json.Marshal(cfg.AbilityBuilder)
	if err != nil {
		return "{}"
	}
	return template.JS(b)
}
