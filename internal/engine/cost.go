package engine

import (
	"math"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// Cost is the running total of build points and energy for something.
type Cost struct {
	Build  int `json:"build"`
	Energy int `json:"energy"`
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
// ability type or enactment is special-cased in Go.
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
			for _, opt := range f.Options {
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
		case "free_number":
			total = addNumberCost(total, f, values[f.Key])
		case "solutions":
			total = addRowsCost(cfg, total, f, values[f.Key])
		case "states":
			total = addStatesCost(cfg, total, f, values[f.Key])
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
// field's default value, honoring the step size and rounding mode.
func addNumberCost(total Cost, f config.Field, raw any) Cost {
	if f.PerStep == nil {
		return total
	}
	step := f.Step
	if step == 0 {
		step = 1
	}
	delta := asInt(raw) - asInt(f.Default)
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

// addRowsCost handles a "solutions" field: a repeatable set of rows. PerItem
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

// addStatesCost handles a "states" field. Each row references either a specific
// state (fixed cost) or a general state (per-shift cost). Additional rows beyond
// the first incur the config-wide additional_state surcharge.
func addStatesCost(cfg *config.Config, total Cost, f config.Field, raw any) Cost {
	rows := asRows(raw)
	for i, row := range rows {
		if i > 0 {
			total.plus(cfg.AdditionalState)
		}
		switch asString(row["state_kind"]) {
		case "specific":
			id := asString(row["specific_state"])
			for _, s := range cfg.SpecificStates {
				if s.ID == id {
					total.Build += s.BuildCost
					total.Energy += s.EnergyCost
				}
			}
		case "general":
			id := asString(row["general_state"])
			shift := abs(asInt(row["shift_amount"]))
			for _, s := range cfg.GeneralStates {
				if s.ID == id {
					total.plusN(s.ShiftCost, shift)
				}
			}
		}
	}
	return total
}

// ComponentCost returns a component's base cost plus its field costs.
func ComponentCost(cfg *config.Config, comp config.Component, values map[string]any) Cost {
	c := Cost{Build: comp.BaseCost.BuildCost, Energy: comp.BaseCost.EnergyCost}
	fc := FieldsCost(cfg, comp.Fields, values)
	c.Build += fc.Build
	c.Energy += fc.Energy
	return c
}

// AbilityCost computes the full advisory cost of an ability, including the
// additional-enactment surcharge for each enactment beyond the first.
func AbilityCost(cfg *config.Config, a model.Ability) Cost {
	var total Cost
	if at, ok := cfg.AbilityType(a.Type); ok {
		c := ComponentCost(cfg, at, a.Fields)
		total.Build += c.Build
		total.Energy += c.Energy
	}
	for i, en := range a.Enactments {
		if i > 0 {
			total.plus(cfg.AdditionalEnactment.AsCost())
		}
		if ec, ok := cfg.Enactment(en.Type); ok {
			c := ComponentCost(cfg, ec, en.Fields)
			// The first enactment is free to add: its component base_cost
			// is waived (field-driven costs still apply). Subsequent
			// enactments pay their full base cost.
			if i == 0 {
				c.Build -= ec.BaseCost.BuildCost
				c.Energy -= ec.BaseCost.EnergyCost
			}
			total.Build += c.Build
			total.Energy += c.Energy
		}

		if en.Interaction != "" {
			if ic, ok := cfg.Interaction(en.Interaction); ok {
				c := ComponentCost(cfg, ic, en.InteractionData)
				total.Build += c.Build
				total.Energy += c.Energy
			}
		}
		// Validation (engagement/counter) fields also contribute cost.
		if len(cfg.Validations.Fields) > 0 {
			c := FieldsCost(cfg, cfg.Validations.Fields, en.ValidationData)
			total.Build += c.Build
			total.Energy += c.Energy
		}
	}

	return total
}

// TraitPointsUsed sums the proficiency cost of all trait assignments.
func TraitPointsUsed(cfg *config.Config, c model.Character) int {
	used := 0
	for _, g := range cfg.Traits.List() {
		for _, trait := range g.Traits {
			profID := c.Traits[model.TraitKey(g.ID, trait)]
			used += cfg.ProficiencyCost(profID)
		}
	}
	return used
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
