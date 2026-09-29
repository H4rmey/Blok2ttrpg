// This file prices individual fields. It walks a field definition and its
// posted value and accumulates build points and energy, which is the primitive
// every higher-level cost is built from.
package engine

import (
	"math"
	"strings"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
)

// Cost is the running total of build points, energy and actions for something.
type Cost struct {
	Build  int `json:"build"`
	Energy int `json:"energy"`
	// Action is the number of actions it takes to use the perk. It is only
	// meaningful for perk types that actually cost actions (base_action > 0);
	// reactions and similar zero-action types leave this at 0. When set it is
	// clamped to a minimum of 1 action.
	Action int `json:"action"`
}

func (c *Cost) plus(x config.Cost) {
	c.Build += x.BuildCost
	c.Energy += x.EnergyCost
}

func (c *Cost) plusN(x config.Cost, n int) {
	c.Build += x.BuildCost * n
	c.Energy += x.EnergyCost * n
}

// FieldsCost computes the cost contribution of a set of field values against
// their field definitions. It handles every field type generically so no
// perk type or enactment is special-cased in Go.
func FieldsCost(cfg *config.Config, fields []config.Field, values map[string]any) Cost {
	var total Cost
	for _, f := range fields {
		// Respect visibility_when: a hidden field contributes nothing.
		if f.VisibilityWhen != "" {
			ctrl := asString(values[f.VisibilityWhen])
			if ctrl == "" {
				// Fall back to the controlling field's default when unsubmitted.
				ctrl = controllingDefault(fields, f.VisibilityWhen)
			}
			if ctrl != f.ShowWhen {
				continue
			}
		}
		switch f.Type {
		case "checkbox":
			if asBool(values[f.Key]) && f.Cost != nil {
				total.plus(*f.Cost)
			}
		case "dropdown":
			val := asString(values[f.Key])
			// Resolve options including any options_source reference so costed
			// sources (e.g. per-trigger costs) contribute, not just inline
			// options defined directly on the field.
			opts := f.Options
			if cfg != nil {
				opts = cfg.ResolveOptions(f)
			}
			for _, opt := range opts {
				if opt.Value == val {
					if opt.Cost != nil {
						total.plus(*opt.Cost)
					}
					// Nested option fields contribute their own cost.
					if len(opt.Fields) > 0 {
						oc := FieldsCost(cfg, opt.Fields, values)
						total.Build += oc.Build
						total.Energy += oc.Energy
					}
				}
			}

			if f.Cost != nil && val != "" {
				total.plus(*f.Cost)
			}
			// Skill group offsets: leaning cost applied per selected group.
			if off := cfg.GroupOffsetFor(f, val); off != nil {
				total.plus(*off)
			}

			// An inline_builder dropdown spawns a nested component builder.
			// Its cost is field-driven only: the referenced component's
			// base_cost is intentionally NOT added, only the cost of the
			// nested field values selected within it. The nested values are
			// stored under "<key>_ib" by the form parser.
			if f.InlineBuilder != nil && val != "" && cfg != nil {
				if comp, ok := cfg.ComponentByKind(f.InlineBuilder.Kind, val); ok {
					if nested, ok := values[f.Key+"_ib"].(map[string]any); ok {
						ic := FieldsCost(cfg, comp.Fields, nested)
						total.Build += ic.Build
						total.Energy += ic.Energy
					}
				}
			}
		case "free_number":
			total = addNumberCost(total, f, values[f.Key])
		case "multiselect":
			total = addRowsCost(cfg, total, f, values[f.Key])

		case "conditions":
			total = addConditionsCost(cfg, total, f, values[f.Key])
		case "condition_select":
			total = addConditionSelectCost(cfg, total, f, values)
		}
	}
	return total
}

// addConditionSelectCost handles a single-condition selector field. The selected value
// is namespaced "general.<id>" or "specific.<id>". A specific condition adds its
// fixed cost; a general condition adds its per-shift cost times the absolute shift
// amount read from the sibling field named by ShiftKey (default
// "shift_amount"). Only one condition can be applied, so there is no additional-
// condition surcharge here.
func addConditionSelectCost(cfg *config.Config, total Cost, f config.Field, values map[string]any) Cost {
	val := asString(values[f.Key])
	if val == "" || cfg == nil {
		return total
	}
	shiftKey := f.ShiftKey
	if shiftKey == "" {
		shiftKey = "shift_amount"
	}
	// A bare (non-namespaced) value refers to a unified condition. Resolve it
	// directly: a shiftable condition pays its per-shift cost, otherwise it
	// pays its flat build/energy cost.
	if strings.IndexByte(val, '.') < 0 {
		if u, ok := cfg.ConditionByID(val); ok {
			if u.Shiftable() {
				shift := abs(asInt(values[shiftKey]))
				total.plusN(u.ShiftCost, shift)
			} else {
				total.Build += u.BuildCost
				total.Energy += u.EnergyCost
			}
		}
		return total
	}
	// Legacy namespaced values: "specific.<id>" / "general.<id>".
	kind, id := val, ""
	if i := strings.IndexByte(val, '.'); i >= 0 {
		kind, id = val[:i], val[i+1:]
	}
	switch kind {
	case "specific":
		if s, ok := cfg.SpecificConditionByID(id); ok {
			total.Build += s.BuildCost
			total.Energy += s.EnergyCost
		}
	case "general":
		if s, ok := cfg.GeneralConditionByID(id); ok {
			shift := abs(asInt(values[shiftKey]))
			total.plusN(s.ShiftCost, shift)
		}
	}
	return total

}

func controllingDefault(fields []config.Field, key string) string {
	for _, f := range fields {
		if f.Key == key {
			return asString(f.Default)
		}
	}
	return ""
}

// addNumberCost applies per-step increase/decrease costs relative to the
// field's default value, honoring the step size and rounding mode. The baseline
// is clamped into the field's range because a config may declare a default
// outside its own min/max; normalization stores the clamped value, so the
// baseline has to be clamped the same way for the delta to come out at zero.
func addNumberCost(total Cost, f config.Field, raw any) Cost {
	if f.PerStep == nil {
		return total
	}
	step := f.Step
	if step == 0 {
		step = 1
	}
	delta := asInt(raw) - clampNumber(f, asInt(f.Default))
	if delta == 0 {
		return total
	}
	n := stepsFor(delta, step, f.Rounding)
	if delta > 0 {
		if f.PerStep.Increase != nil {
			total.Build += f.PerStep.Increase.BuildCost * n
			total.Energy += f.PerStep.Increase.EnergyCost * n
		}
	} else {
		if f.PerStep.Decrease != nil {
			total.Build += f.PerStep.Decrease.BuildCost * n
			total.Energy += f.PerStep.Decrease.EnergyCost * n
		}
	}
	return total
}

// stepsFor returns the (positive) number of steps represented by delta at the
// given step size. Rounding controls how a partial step is counted.
func stepsFor(delta, step int, rounding string) int {
	if step <= 0 {
		step = 1
	}
	q := float64(abs(delta)) / float64(step)
	switch rounding {
	case "ceil":
		return int(math.Ceil(q))
	case "floor":
		return int(math.Floor(q))
	default:
		return abs(delta) / step
	}
}

// addRowsCost handles a "multiselect" field: a repeatable set of rows. PerItem

// adjusts cost per row relative to the default count (increase when there are
// more rows than default, decrease when fewer). Each row's fields also cost.
func addRowsCost(cfg *config.Config, total Cost, f config.Field, raw any) Cost {
	rows := asRows(raw)
	if f.PerItem != nil {
		delta := len(rows) - f.DefaultCount
		if delta > 0 && f.PerItem.Increase != nil {
			total.Build += f.PerItem.Increase.BuildCost * delta
			total.Energy += f.PerItem.Increase.EnergyCost * delta
		} else if delta < 0 && f.PerItem.Decrease != nil {
			total.Build += f.PerItem.Decrease.BuildCost * (-delta)
			total.Energy += f.PerItem.Decrease.EnergyCost * (-delta)
		}
	}
	for _, row := range rows {
		rc := FieldsCost(cfg, f.RowFields, row)
		total.Build += rc.Build
		total.Energy += rc.Energy
	}
	return total
}

// addConditionsCost handles a "conditions" field. Each row references either a specific
// condition (fixed cost) or a general condition (per-shift cost). Additional rows beyond
// the first incur the config-wide additional_condition surcharge.
func addConditionsCost(cfg *config.Config, total Cost, f config.Field, raw any) Cost {
	rows := asRows(raw)
	for i, row := range rows {
		if i > 0 {
			total.plus(cfg.AdditionalCondition)
		}
		switch asString(row["condition_kind"]) {
		case "specific":
			id := asString(row["specific_condition"])
			for _, s := range cfg.SpecificConditions {
				if s.ID == id {
					total.Build += s.BuildCost
					total.Energy += s.EnergyCost
				}
			}
		case "general":
			id := asString(row["general_condition"])
			shift := abs(asInt(row["shift_amount"]))
			for _, s := range cfg.GeneralConditions {
				if s.ID == id {
					total.plusN(s.ShiftCost, shift)
				}
			}
		}
		// Row sub-fields (dropdowns/checkboxes/etc.) can carry their own
		// per-entry costs. Reuse the generic field coster so any extra
		// options attached to a condition row contribute their configured cost.
		if len(f.RowFields) > 0 {
			rc := FieldsCost(cfg, f.RowFields, row)
			total.Build += rc.Build
			total.Energy += rc.Energy
		}
	}
	return total
}
