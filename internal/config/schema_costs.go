package config

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/harmey/blok2ttrpg/ability-builder/internal/models"
)

func BuildFieldValueMap(values url.Values, fields []FieldConfig) FieldValueMap {
	out := FieldValueMap{}
	for _, f := range fields {
		switch f.Type {
		case "states":
			rows := extractStateRows(values, f.Key)
			filtered := rows[:0]
			for _, r := range rows {
				if !isBlankStateRow(r) {
					filtered = append(filtered, r)
				}
			}
			if len(filtered) > 0 {
				out[f.Key] = filtered
			}
		case "solutions":
			rows, list := extractSolutionRows(values, f.Key)
			if len(rows) > 0 {
				out[f.Key] = rows
			}
			if len(list) > 0 {
				out[f.Key+"__list"] = list
			}
		case "dropdown":
			if v, ok := firstValue(values, f.Key); ok {
				out[f.Key] = v
				childMap := FieldValueMap{}
				for k, vs := range values {
					if !hasPrefix(k, f.Key+"__") {
						continue
					}
					rest := k[len(f.Key+"__"):]
					if len(vs) == 1 {
						childMap[rest] = vs[0]
					} else {
						childMap[rest] = vs
					}
				}
				if len(childMap) > 0 {
					out[f.Key+"__option"] = map[string]any{v: childMap}
				}
			}
		default:
			if v, ok := firstValue(values, f.Key); ok {
				out[f.Key] = v
			}
		}
	}
	return out
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func extractSolutionRows(values url.Values, fieldKey string) ([]FieldValueMap, []string) {
	prefix := fieldKey + "__"
	vals := values[prefix+"value"]
	types := values[prefix+"type"]
	if len(vals) == 0 && len(types) == 0 {
		if v, ok := values[fieldKey]; ok {
			rows := make([]FieldValueMap, 0, len(v))
			for _, s := range v {
				if s != "" {
					rows = append(rows, FieldValueMap{"value": s})
				}
			}
			return rows, v
		}
		return nil, nil
	}
	rows := make([]FieldValueMap, 0, len(vals))
	flat := make([]string, 0, len(vals))
	for i, v := range vals {
		if v == "" {
			continue
		}
		row := FieldValueMap{"value": v}
		if i < len(types) && types[i] != "" {
			row["type"] = types[i]
		}
		rows = append(rows, row)
		flat = append(flat, v)
	}
	return rows, flat
}

func sortStrKeys(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}

func extractStateRows(values url.Values, fieldKey string) []FieldValueMap {
	prefix := fieldKey + "__"
	kinds := values[prefix+"state_kind"]
	specs := values[prefix+"specific_state"]
	gens := values[prefix+"general_state"]
	shifts := values[prefix+"shift_amount"]
	if len(kinds) == 0 && len(specs) == 0 && len(gens) == 0 && len(shifts) == 0 {
		return nil
	}
	n := len(kinds)
	if n < len(specs) {
		n = len(specs)
	}
	if n < len(gens) {
		n = len(gens)
	}
	if n < len(shifts) {
		n = len(shifts)
	}
	if n == 0 {
		n = 1
	}
	rows := make([]FieldValueMap, 0, n)
	for i := 0; i < n; i++ {
		row := FieldValueMap{}
		if i < len(kinds) {
			row["state_kind"] = kinds[i]
		}
		if i < len(specs) {
			row["specific_state"] = specs[i]
		}
		if i < len(gens) {
			row["general_state"] = gens[i]
		}
		if i < len(shifts) {
			if n, err := strconv.Atoi(shifts[i]); err == nil {
				row["shift_amount"] = n
			} else {
				row["shift_amount"] = shifts[i]
			}
		}
		rows = append(rows, row)
	}
	return rows
}

func firstValue(values url.Values, key string) (string, bool) {
	if v, ok := values[key]; ok && len(v) > 0 {
		return v[0], true
	}
	return "", false
}

// ComputeEnactmentCosts returns the authoritative build/cast cost for an
// enactment using the config schema. It includes the enactment base cost,
// evaluates each declared field, and (for Enact State) adds the additional
// state surcharge from StatesConfig.
func ComputeEnactmentCosts(cfg EnactmentConfig, values url.Values, stateCfg StatesConfig) (build, cast int, err error) {
	build = cfg.BaseCost.AddCost
	cast = cfg.BaseCost.EnergyCost
	if len(cfg.Fields) == 0 {
		return build, cast, nil
	}
	fv := BuildFieldValueMap(values, cfg.Fields)
	if err := ValidateSubmittedFields(cfg.Fields, fv); err != nil {
		return 0, 0, err
	}
	b, c, err := EvaluateFieldCosts(cfg.Fields, fv)
	if err != nil {
		return 0, 0, err
	}
	build += b
	cast += c
	if cfg.Type == string(models.EnactState) {
		if rows, ok := fv["states"].([]FieldValueMap); ok {
			enforced, err := enforceStateRowBounds(rows, stateCfg.GeneralStates, stateCfg.SpecificStates)
			if err != nil {
				return 0, 0, err
			}
			for _, row := range enforced {
				rb, rc := EvaluateStateRow(row, stateCfg.GeneralStates, stateCfg.SpecificStates)
				build += rb
				cast += rc
			}
			sb, sc := EvaluateStatesWithSurcharge(stateCfg.AdditionalState, len(enforced))
			build += sb
			cast += sc
		}
	}
	return build, cast, nil
}

func enforceStateRowBounds(rows []FieldValueMap, generals []GeneralStateConfig, specifics []SpecificStateConfig) ([]FieldValueMap, error) {
	out := make([]FieldValueMap, 0, len(rows))
	for i, row := range rows {
		if isBlankStateRow(row) {
			continue
		}
		kind, _ := toString(row["state_kind"])
		if kind == "" {
			return nil, fmt.Errorf("state row %d: kind is required", i)
		}
		switch kind {
		case "specific":
			id, _ := toString(row["specific_state"])
			if id == "" {
				return nil, fmt.Errorf("state row %d: specific state is required", i)
			}
			if !specificStateExists(specifics, id) {
				return nil, fmt.Errorf("state row %d: unknown specific state %q", i, id)
			}
		case "general":
			id, _ := toString(row["general_state"])
			if id == "" {
				return nil, fmt.Errorf("state row %d: general state id is required", i)
			}
			var match *GeneralStateConfig
			for j := range generals {
				if generals[j].ID == id {
					match = &generals[j]
					break
				}
			}
			if match == nil {
				return nil, fmt.Errorf("state row %d: unknown general state %q", i, id)
			}
			shift, _ := toInt(row["shift_amount"])
			if shift < match.MinShift || shift > match.MaxShift {
				return nil, fmt.Errorf("state row %d: shift %d out of range [%d, %d] for %s", i, shift, match.MinShift, match.MaxShift, id)
			}
			row["shift_amount"] = shift
		default:
			return nil, fmt.Errorf("state row %d: invalid state_kind %q", i, kind)
		}
		out = append(out, row)
	}
	return out, nil
}

func specificStateExists(specifics []SpecificStateConfig, id string) bool {
	for _, s := range specifics {
		if s.ID == id {
			return true
		}
	}
	return false
}

// ComputeInteractionCosts returns the authoritative build/cast cost for an
// interaction using the config schema. It includes the interaction base
// cost and evaluates each declared field.
func ComputeInteractionCosts(cfg InteractionConfig, values url.Values) (build, cast int, err error) {
	build = cfg.BaseCost.AddCost
	cast = cfg.BaseCost.EnergyCost
	if len(cfg.Fields) == 0 {
		return build, cast, nil
	}
	fv := BuildFieldValueMap(values, cfg.Fields)
	if err := ValidateSubmittedFields(cfg.Fields, fv); err != nil {
		return 0, 0, err
	}
	b, c, err := EvaluateFieldCosts(cfg.Fields, fv)
	if err != nil {
		return 0, 0, err
	}
	build += b
	cast += c
	return build, cast, nil
}

// ComputeValidationCosts returns the authoritative build/cast cost for the
// validation card using the config schema. The supplied row values come
// from r.Form["enact_<idx>_valid_<key>"] parsed into a url.Values-like map.
func ComputeValidationCosts(cfg ValidationConfig, values url.Values) (build, cast int, err error) {
	if len(cfg.Fields) == 0 {
		return 0, 0, nil
	}
	fv := BuildFieldValueMap(values, cfg.Fields)
	if err := ValidateSubmittedFields(cfg.Fields, fv); err != nil {
		return 0, 0, err
	}
	b, c, err := EvaluateFieldCosts(cfg.Fields, fv)
	if err != nil {
		return 0, 0, err
	}
	build += b
	cast += c
	return build, cast, nil
}

// ComputeAbilityTypeCosts recomputes the ability type build/cast from the
// schema fields when the config defines them. It also writes back derived
// values onto the Ability struct (EnergySteps, ActionSteps, etc.) for export
// compatibility.
func (ab *AbilityBuilderConfig) ComputeAbilityTypeCosts(a *models.Ability, values url.Values) error {
	cfg, ok := ab.AbilityTypes[strings.ToLower(string(a.Type))]
	if !ok {
		return fmt.Errorf("unknown ability type: %s", a.Type)
	}
	energy := cfg.BaseEnergy
	action := cfg.BaseAction
	build := 0
	if len(cfg.Fields) == 0 {
		return ab.computeAbilityCostsLegacy(a, values, cfg, energy, action, build)
	}
	normalized := normalizeAbilityTypeValues(values)
	fv := BuildFieldValueMap(normalized, cfg.Fields)
	if err := ValidateSubmittedFields(cfg.Fields, fv); err != nil {
		return err
	}
	b, c, err := EvaluateFieldCosts(cfg.Fields, fv)
	if err != nil {
		return err
	}
	build += b
	energy += c
	action += computeActionDelta(fv)
	populateAbilityFromFields(a, fv)
	// Item name is a free_text field so it does not contribute cost, but
	// we still want to persist it for export and edit hydration.
	if v, ok := normalized["item_name"]; ok && len(v) > 0 {
		a.ItemName = v[0]
	}
	a.HasItemDependency = toBoolSafe(fv["item_dep"])
	a.BuildCost = build
	a.EnergyCost = energy
	a.ActionCost = action
	a.Fields = fv
	return nil
}

// normalizeAbilityTypeValues copies the form values and renames the
// top-level ability-type fields so the generic schema evaluator can read
// them under their schema keys. The JS submit handler renames
// item_dep -> ability_item_dep and item_name -> ability_item_name on the
// ability-type card only (it skips fields inside enactment blocks).
func normalizeAbilityTypeValues(values url.Values) url.Values {
	out := make(url.Values, len(values))
	for k, v := range values {
		out[k] = v
	}
	if v, ok := out["ability_item_dep"]; ok {
		out["item_dep"] = v
	}
	if v, ok := out["ability_item_name"]; ok {
		out["item_name"] = v
	}
	return out
}

// computeActionDelta returns the change in action cost based on the
// action_steps field. The schema's per_step encodes the build/cast cost of
// each step; the action cost itself shifts by the raw step value.
func computeActionDelta(fv FieldValueMap) int {
	if v, ok := fv["action_steps"]; ok {
		return toIntSafe(v)
	}
	return 0
}

func populateAbilityFromFields(a *models.Ability, fv FieldValueMap) {
	if v, ok := fv["energy_steps"]; ok {
		a.EnergySteps = toIntSafe(v)
	}
	if v, ok := fv["action_steps"]; ok {
		a.ActionSteps = toIntSafe(v)
	}
	if v, ok := fv["range"]; ok {
		a.ReactionRange = toIntSafe(v)
	}
	if v, ok := fv["uses"]; ok {
		a.ReactionUses = toIntSafe(v)
	}
	if v, ok := fv["trigger"]; ok {
		if s, ok := v.(string); ok {
			a.Trigger = s
		}
	}
	if v, ok := fv["trigger_trait"]; ok {
		if s, ok := v.(string); ok {
			a.TriggerTrait = s
		}
	}
	if v, ok := fv["phase_rounds"]; ok {
		a.PhaseDuration = toIntSafe(v)
	}
	if v, ok := fv["reverse_rounds"]; ok {
		a.ReversePhaseRounds = toIntSafe(v)
	}
	if v, ok := fv["hp"]; ok {
		a.HPBonus = toIntSafe(v)
	}
	if v, ok := fv["life"]; ok {
		a.ExtraLifetime = toIntSafe(v)
	}
	if v, ok := fv["all_req"]; ok {
		a.AllKnockoutsReq = toBoolSafe(v)
	}
	if v, ok := fv["reverse_knockout"]; ok {
		a.ReverseKnockoutOK = toBoolSafe(v)
	}
	if v, ok := fv["no_knockout"]; ok {
		a.NoKnockout = toBoolSafe(v)
	}
	if v, ok := fv["knockout"]; ok {
		switch arr := v.(type) {
		case []string:
			a.Knockouts = arr
		}
	}
	if v, ok := fv["effortless"]; ok {
		a.Effortless = toBoolSafe(v)
	}
	if v, ok := fv["iron_will"]; ok {
		a.IronWill = toBoolSafe(v)
	}
	if v, ok := fv["dual_focus"]; ok {
		a.DualFocus = toBoolSafe(v)
	}
}

func toIntSafe(v interface{}) int {
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

func toBoolSafe(v interface{}) bool {
	switch x := v.(type) {
	case bool:
		return x
	case string:
		return x == "on" || x == "true" || x == "1" || x == "yes"
	}
	return false
}
