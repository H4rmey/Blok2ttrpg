package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
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
	state.CharacterID = charID
	state.Ability = models.Ability{}
	app.Sessions.Update(sessID, state)

	initialState := buildInitialState(&state.Ability)

	app.render(w, "builder.html", map[string]interface{}{
		"Character":              c,
		"Ability":                state.Ability,
		"Breadcrumbs":            newAbilityBreadcrumbs(c),
		"AbilityTypes":           abilityTypeList(app.Config.AbilityBuilder),
		"AllEnactmentTypes":      enactmentTypeList(app.Config.AbilityBuilder),
		"AllInteractionTypes":    interactionTypeList(app.Config.AbilityBuilder),
		"CompatibleEnactsMap":    compatibleEnactsMapForTemplate(app.Config.AbilityBuilder),
		"GeneralTraitNames":      app.Config.AbilityBuilder.Traits.General,
		"OffenseTraitNames":      app.Config.AbilityBuilder.Traits.Offense,
		"DefenseTraitNames":      app.Config.AbilityBuilder.Traits.Defense,
		"AllTraits":              models.AllTraitNames(),
		"ReactionTriggers":       models.ReactionTriggers,
		"KnockoutOptions":        models.KnockoutOptions,
		"DirectionOptions":       models.DirectionOptions,
		"ShiftDirectionOptions":  models.ShiftDirectionOptions,
		"TriggerTimings":         models.TriggerTimings,
		"AoETimings":             models.AoETriggerTimings,
		"PersistentEffectTypes":  models.PersistentEffectTypes,
		"DamageDiceOptions":      models.DamageDiceOptions,
		"GenericDieOptions":      models.GenericDieOptions,
		"InitialStateJSON":       initialState,
		"IsEdit":                 state.Ability.ID != "",
		"BuilderConfigJSON":      mustMarshalConfig(app.Config),
		"AbilityPointsBudget":    abilityPointsBudget(c.Level, app.Config.AbilityBuilder.Leveling),
		"AbilityPointsUsed":      abilityPointsUsed(*c, app.Config.AbilityBuilder),
		"AbilityPointsRemaining": abilityPointsBudget(c.Level, app.Config.AbilityBuilder.Leveling) - abilityPointsUsed(*c, app.Config.AbilityBuilder),
	})
}

// abilityTypeList returns the ordered list of ability type display names.
// When the split config declares ability types, it uses those (sorted by
// their config key) and falls back to the legacy models.AllAbilityTypes
// otherwise.
func abilityTypeList(ab config.AbilityBuilderConfig) []models.AbilityType {
	if len(ab.AbilityTypes) == 0 {
		return models.AllAbilityTypes
	}
	keys := make([]string, 0, len(ab.AbilityTypes))
	for k := range ab.AbilityTypes {
		keys = append(keys, k)
	}
	sortStrings(keys)
	out := make([]models.AbilityType, 0, len(keys))
	for _, k := range keys {
		cfg := ab.AbilityTypes[k]
		if cfg.Name != "" {
			out = append(out, models.AbilityType(cfg.Name))
		} else {
			out = append(out, models.AbilityType(ucfirst(k)))
		}
	}
	return out
}

func enactmentTypeList(ab config.AbilityBuilderConfig) []models.EnactmentType {
	if len(ab.Enactments) == 0 {
		return models.AllEnactmentTypes
	}
	keys := make([]string, 0, len(ab.Enactments))
	for k := range ab.Enactments {
		keys = append(keys, k)
	}
	sortStrings(keys)
	out := make([]models.EnactmentType, 0, len(keys))
	for _, k := range keys {
		cfg := ab.Enactments[k]
		if cfg.Type != "" {
			out = append(out, models.EnactmentType(cfg.Type))
		} else {
			out = append(out, models.EnactmentType("Enact "+ucfirst(k)))
		}
	}
	return out
}

func interactionTypeList(ab config.AbilityBuilderConfig) []models.InteractionType {
	if len(ab.Interactions) == 0 {
		return models.AllInteractionTypes
	}
	keys := make([]string, 0, len(ab.Interactions))
	for k := range ab.Interactions {
		keys = append(keys, k)
	}
	sortStrings(keys)
	out := make([]models.InteractionType, 0, len(keys))
	for _, k := range keys {
		cfg := ab.Interactions[k]
		if cfg.Type != "" {
			out = append(out, models.InteractionType(cfg.Type))
		} else {
			out = append(out, models.InteractionType(ucfirst(strings.ReplaceAll(k, "_", " "))))
		}
	}
	return out
}

func ucfirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}

// compatibleEnactsMapForTemplate returns the compatible-enacts mapping in a
// form that is easy to render as a JS object literal. It prefers the split
// config's CompatibleEnactments per ability type, falling back to the
// legacy models.CompatibleEnactments map.
func compatibleEnactsMapForTemplate(ab config.AbilityBuilderConfig) map[string][]string {
	out := map[string][]string{}
	if len(ab.AbilityTypes) == 0 {
		for _, at := range models.AllAbilityTypes {
			for _, et := range models.CompatibleEnactments[at] {
				out[string(at)] = append(out[string(at)], string(et))
			}
		}
		return out
	}
	keys := make([]string, 0, len(ab.AbilityTypes))
	for k := range ab.AbilityTypes {
		keys = append(keys, k)
	}
	sortStrings(keys)
	for _, k := range keys {
		cfg := ab.AbilityTypes[k]
		name := cfg.Name
		if name == "" {
			name = ucfirst(k)
		}
		out[name] = append(out[name], cfg.CompatibleEnactments...)
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
		"effortless":           a.Effortless,
		"iron_will":            a.IronWill,
		"dual_focus":           a.DualFocus,
		"enactments":           []map[string]interface{}{},
		"fields":               a.Fields,
	}
	for _, e := range a.Enactments {
		em := map[string]interface{}{
			"name":            e.Name,
			"description":     e.Description,
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
			"fields":          e.Fields,
		}
		if e.Type == models.EnactState && e.Fields == nil {
			em["state_type"] = "specific"
			if e.Source != "" {
				em["specific_state"] = e.Source
			}
			em["shift_amount"] = e.ShiftAmount
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
				"fields":         e.Interaction.Fields,
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
					"fields":                v.Fields,
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

	// Concentration-specific fields
	if a.Type == models.AbilityConcentration {
		a.Effortless = r.FormValue("effortless") == "on"
		a.IronWill = r.FormValue("iron_will") == "on"
		a.DualFocus = r.FormValue("dual_focus") == "on"
	}

	enactments, err := parseNewEnactments(r, &app.Config.AbilityBuilder)
	if err != nil {
		http.Error(w, "Invalid enactment submission: "+err.Error(), http.StatusBadRequest)
		return
	}
	a.Enactments = enactments
	if !hasInteraction(a.Enactments) {
		http.Error(w, "At least one interaction is required", http.StatusBadRequest)
		return
	}

	if c, err := app.Store.GetCharacter(state.CharacterID); err == nil {
		budget := abilityPointsBudget(c.Level, app.Config.AbilityBuilder.Leveling)
		used := abilityPointsUsed(*c, app.Config.AbilityBuilder)
		cost := abilityBuildCost(*a, app.Config.AbilityBuilder)
		if used+cost > budget {
			http.Error(w, "Not enough ability points to save this ability", http.StatusBadRequest)
			return
		}
	}

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
// Uses the config schema to compute costs authoritatively. Invalid
// submissions (out-of-bounds numbers, unknown dropdown options, malformed
// state rows) are returned as errors so the handler can respond with a
// proper 400 status. Legacy configs without a fields schema fall back to
// the submitted build/cast values.
func parseNewEnactments(r *http.Request, cfg *config.AbilityBuilderConfig) ([]models.Enactment, error) {
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

		enactCfgKey := enactTypeKey(enactType)
		enactFields := url.Values{}
		interFields := url.Values{}
		validFields := url.Values{}
		for k, v := range r.Form {
			if rest, ok := stripPrefix(k, prefix); ok {
				if rest == "type" || rest == "name" || rest == "description" || rest == "build" || rest == "cast" || rest == "formula" {
					continue
				}
				if interRest, ok := stripPrefix(rest, "inter_"); ok {
					interFields[interRest] = v
					continue
				}
				if validRest, ok := stripPrefix(rest, "valid_"); ok {
					validFields[validRest] = v
					continue
				}
				enactFields[rest] = v
			}
		}

		var ecfg config.EnactmentConfig
		var hasEnactFields bool
		if e, ok := cfg.Enactments[enactCfgKey]; ok {
			ecfg = e
			hasEnactFields = len(e.Fields) > 0
		}

		var build, cast int
		if hasEnactFields {
			b, c, err := config.ComputeEnactmentCosts(ecfg, enactFields, cfg.States)
			if err != nil {
				return nil, fmt.Errorf("enactment %d (%s): %w", idx, enactType, err)
			}
			build, cast = b, c
		} else {
			build = atoi(r.FormValue(prefix + "build"))
			cast = atoi(r.FormValue(prefix + "cast"))
		}

		e := models.Enactment{
			Name:           strings.TrimSpace(r.FormValue(prefix + "name")),
			Description:    strings.TrimSpace(r.FormValue(prefix + "description")),
			Type:           models.EnactmentType(enactType),
			BuildCost:      build,
			CastCost:       cast,
			Formula:        r.FormValue(prefix + "formula"),
			Source:         r.FormValue(prefix + "source"),
			SourceTrait:    r.FormValue(prefix + "source_trait"),
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
		if hasEnactFields {
			e.Fields = config.BuildFieldValueMap(enactFields, ecfg.Fields)
			populateEnactmentTypedFromFields(&e, e.Fields)
		}

		interType := r.FormValue(prefix + "inter_type")
		if interType != "" {
			interCfgKey := interTypeKey(interType)
			var icfg config.InteractionConfig
			var hasInterFields bool
			if i2, ok := cfg.Interactions[interCfgKey]; ok {
				icfg = i2
				hasInterFields = len(i2.Fields) > 0
			}
			inter := models.Interaction{
				Type:          models.InteractionType(interType),
				Range:         atoi(interFields.Get("range")),
				Targets:       atoi(interFields.Get("targets")),
				Radius:        atoi(interFields.Get("radius")),
				OriginMode:    interFields.Get("origin_mode"),
				OriginText:    interFields.Get("origin_text"),
				Duration:      atoi(interFields.Get("duration")),
				Timing:        interFields.Get("timing"),
				VisibleOK:     interFields.Get("visible") == "on",
				ObstructedOK:  interFields.Get("obstructed") == "on",
				RemovePenalty: interFields.Get("remove_penalty") == "on",
				Immune:        interFields.Get("immune") == "on",
				UsePrevious:   interFields.Get("use_previous") == "on",
			}
			if hasInterFields {
				b, c, err := config.ComputeInteractionCosts(icfg, interFields)
				if err != nil {
					return nil, fmt.Errorf("interaction %d (%s) in enactment %d: %w", idx, interType, idx, err)
				}
				inter.BuildCost = b
				inter.CastCost = c
				inter.Fields = config.BuildFieldValueMap(interFields, icfg.Fields)
				populateInteractionTypedFromFields(&inter, inter.Fields)
			} else {
				inter.BuildCost = atoi(r.FormValue(prefix + "inter_build"))
				inter.CastCost = atoi(r.FormValue(prefix + "inter_cast"))
			}

			mode := models.EngageMode(validFields.Get("engage_mode"))
			inter.Validation = &models.Validation{
				EngageMode:          mode,
				EngageTrait:         validFields.Get("engage_trait"),
				EngageTraitCategory: models.TraitCategory(validFields.Get("engage_trait_category")),
				EngageDie:           validFields.Get("engage_die"),
				EngageOther:         validFields.Get("engage_other"),
			}
			if len(cfg.Validations.Fields) > 0 {
				b, c, err := config.ComputeValidationCosts(cfg.Validations, validFields)
				if err != nil {
					return nil, fmt.Errorf("validation in enactment %d: %w", idx, err)
				}
				inter.Validation.BuildCost = b
				inter.Validation.CastCost = c
				inter.Validation.Fields = config.BuildFieldValueMap(validFields, cfg.Validations.Fields)
				populateValidationTypedFromFields(inter.Validation, inter.Validation.Fields)
			} else {
				inter.Validation.BuildCost = atoi(r.FormValue(prefix + "valid_build"))
				inter.Validation.CastCost = atoi(r.FormValue(prefix + "valid_cast"))
			}
			if counterTypes := validFields["counter_type"]; len(counterTypes) > 0 {
				counterTraits := validFields["counter_trait"]
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
	return out, nil
}

func populateEnactmentTypedFromFields(e *models.Enactment, fv config.FieldValueMap) {
	if v, ok := fv["source"]; ok {
		if s, ok := v.(string); ok {
			e.Source = s
		}
	}
	if v, ok := fv["source_trait"]; ok {
		if s, ok := v.(string); ok {
			e.SourceTrait = s
		}
	}
	if v, ok := fv["other"]; ok {
		if s, ok := v.(string); ok {
			e.OtherRollText = s
		}
	}
	if v, ok := fv["flat"]; ok {
		e.FlatBonus = configToInt(v)
	}
	if v, ok := fv["offense"]; ok {
		if s, ok := v.(string); ok {
			e.OffensiveTrait = s
		}
	}
	if v, ok := fv["medicine"]; ok {
		if s, ok := v.(string); ok {
			e.MedicineTrait = s
		}
	}
	if v, ok := fv["origin_mode"]; ok {
		if s, ok := v.(string); ok {
			e.OriginMode = s
		}
	}
	if v, ok := fv["origin_text"]; ok {
		if s, ok := v.(string); ok {
			e.OriginText = s
		}
	}
	if v, ok := fv["distance"]; ok {
		e.Distance = configToInt(v)
	}
	if v, ok := fv["direction"]; ok {
		if arr, ok := v.([]string); ok {
			e.Directions = arr
		}
	}
	if v, ok := fv["shifted_trait"]; ok {
		if s, ok := v.(string); ok {
			e.ShiftedTrait = s
		}
	}
	if v, ok := fv["shift_dir"]; ok {
		if s, ok := v.(string); ok {
			e.ShiftDir = s
		}
	}
	if v, ok := fv["shift_amount"]; ok {
		e.ShiftAmount = configToInt(v)
	}
	if v, ok := fv["shift_uses"]; ok {
		e.ShiftUses = configToInt(v)
	}
	if v, ok := fv["effect_name"]; ok {
		if s, ok := v.(string); ok {
			e.EffectName = s
		}
	}
	if v, ok := fv["effect_type"]; ok {
		if s, ok := v.(string); ok {
			e.EffectType = s
		}
	}
	if v, ok := fv["duration"]; ok {
		e.Duration = configToInt(v)
	}
	if v, ok := fv["trigger_timing"]; ok {
		if s, ok := v.(string); ok {
			e.TriggerTiming = s
		}
	}
	if v, ok := fv["solution"]; ok {
		if arr, ok := v.([]string); ok {
			e.Solutions = arr
		}
	}
	if v, ok := fv["always"]; ok {
		if s, ok := v.(string); ok {
			e.Always = s == "on" || s == "true" || s == "1"
		}
	}
}

func populateInteractionTypedFromFields(i *models.Interaction, fv config.FieldValueMap) {
	if v, ok := fv["range"]; ok {
		i.Range = configToInt(v)
	}
	if v, ok := fv["targets"]; ok {
		i.Targets = configToInt(v)
	}
	if v, ok := fv["radius"]; ok {
		i.Radius = configToInt(v)
	}
	if v, ok := fv["origin_mode"]; ok {
		if s, ok := v.(string); ok {
			i.OriginMode = s
		}
	}
	if v, ok := fv["origin_text"]; ok {
		if s, ok := v.(string); ok {
			i.OriginText = s
		}
	}
	if v, ok := fv["duration"]; ok {
		i.Duration = configToInt(v)
	}
	if v, ok := fv["timing"]; ok {
		if s, ok := v.(string); ok {
			i.Timing = s
		}
	}
	if v, ok := fv["visible"]; ok {
		if s, ok := v.(string); ok {
			i.VisibleOK = s == "on" || s == "true"
		}
	}
	if v, ok := fv["obstructed"]; ok {
		if s, ok := v.(string); ok {
			i.ObstructedOK = s == "on" || s == "true"
		}
	}
	if v, ok := fv["remove_penalty"]; ok {
		if s, ok := v.(string); ok {
			i.RemovePenalty = s == "on" || s == "true"
		}
	}
	if v, ok := fv["immune"]; ok {
		if s, ok := v.(string); ok {
			i.Immune = s == "on" || s == "true"
		}
	}
	if v, ok := fv["use_previous"]; ok {
		if s, ok := v.(string); ok {
			i.UsePrevious = s == "on" || s == "true"
		}
	}
}

func populateValidationTypedFromFields(v *models.Validation, fv config.FieldValueMap) {
	if x, ok := fv["engage_mode"]; ok {
		if s, ok := x.(string); ok {
			v.EngageMode = models.EngageMode(s)
		}
	}
	if x, ok := fv["engage_trait"]; ok {
		if s, ok := x.(string); ok {
			v.EngageTrait = s
		}
	}
	if x, ok := fv["engage_die"]; ok {
		tier := configToInt(x)
		v.EngageDie = dieForTier(tier)
	}
	if x, ok := fv["engage_other"]; ok {
		if s, ok := x.(string); ok {
			v.EngageOther = s
		}
	}
	v.CounterDefaultType = models.CounterTypeDefenseTrait
	if x, ok := fv["counter_trait"]; ok {
		if rows, ok := x.([]config.FieldValueMap); ok {
			v.CounterRolls = nil
			v.CounterRollEntries = nil
			for _, row := range rows {
				val := configToString(row["value"])
				if val == "" {
					continue
				}
				t := models.TraitCategory(configToString(row["type"]))
				if t == "" {
					t = models.TraitCategoryDefense
				}
				v.CounterRolls = append(v.CounterRolls, val)
				v.CounterRollEntries = append(v.CounterRollEntries, models.CounterRoll{Type: t, Trait: val})
			}
		}
	}
}

func dieForTier(tier int) string {
	switch tier {
	case 0:
		return "d6"
	case 1:
		return "d8"
	case 2:
		return "d10"
	case 3:
		return "d12"
	}
	return "d6"
}

func configToString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func configToInt(v interface{}) int {
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	case string:
		n, _ := strconv.Atoi(x)
		return n
	}
	return 0
}

// interTypeKey converts an interaction type to a config key.
func interTypeKey(it string) string {
	switch it {
	case "Self":
		return "self"
	case "Direct":
		return "direct"
	case "Ranged":
		return "ranged"
	case "Area":
		return "area"
	case "Area of Effect":
		return "area_of_effect"
	}
	return strings.ToLower(strings.ReplaceAll(it, " ", "_"))
}

// enactTypeKey converts an enactment type to config key.
func enactTypeKey(et string) string {
	switch et {
	case "Enact Damage":
		return "damage"
	case "Enact Healing":
		return "healing"
	case "Enact Movement":
		return "movement"
	case "Enact Proficiency Shift":
		return "proficiency_shift"
	case "Enact Persistent Effect":
		return "persistent_effect"
	case "Enact Negation":
		return "negation"
	case "Enact State":
		return "state"
	default:
		return strings.ToLower(strings.ReplaceAll(et, " ", "_"))
	}
}

func hasInteraction(enactments []models.Enactment) bool {
	for _, e := range enactments {
		if e.Interaction != nil && e.Interaction.Type != "" {
			return true
		}
	}
	return false
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
	var c *models.Character
	if state.CharacterID != "" {
		if character, err := app.Store.GetCharacter(state.CharacterID); err == nil {
			c = character
		}
	}

	app.render(w, "review.html", map[string]interface{}{
		"Ability":     state.Ability,
		"Breadcrumbs": reviewBreadcrumbs(c),
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
