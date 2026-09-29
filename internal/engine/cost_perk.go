// This file prices whole things: a component from its base cost plus fields,
// and a complete perk from its type, enactments and interactions.
package engine

import (
	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// ComponentCost returns a component's base cost plus its field costs.
func ComponentCost(cfg *config.Config, comp config.Component, values map[string]any) Cost {
	c := Cost{Build: comp.BaseCost.BuildCost, Energy: comp.BaseCost.EnergyCost}
	fc := FieldsCost(cfg, comp.Fields, values)
	c.Build += fc.Build
	c.Energy += fc.Energy
	return c
}

// PerkCost computes the full advisory cost of an perk, including the
// additional-enactment surcharge for each enactment beyond the first.
func PerkCost(cfg *config.Config, a model.Perk) Cost {
	var total Cost
	if at, ok := cfg.PerkType(a.Type); ok {
		c := ComponentCost(cfg, at, a.Fields)
		total.Build += c.Build
		total.Energy += c.Energy
		// Action count is only meaningful for perk types that actually cost
		// actions to use (base_action > 0). Reactions and similar zero-action
		// types leave Action at 0 so the builder can hide the badge. The
		// "action_steps" perk field (if present) adjusts the count relative to
		// the base, and the result is clamped to a minimum of 1 action.
		if at.BaseAction > 0 {
			action := at.BaseAction + asInt(a.Fields["action_steps"])
			if action < 1 {
				action = 1
			}
			total.Action = action
		}
	}

	// A passive's cost is its catalogue entry's flat build cost plus whatever its
	// own fields add. Both are charged here because a passive is not a perk type:
	// it is added from the passive picker rather than built, so there is no
	// perk-type component above to carry the entry cost on a dropdown option.
	//
	// The fields are priced with the same generic coster every other component
	// uses, so a passive gets free text, numbers, checkboxes and dropdowns
	// without any passive-specific cost code.
	//
	// A passive has no enactments, so the loop below is a no-op for it.
	if p, ok := cfg.PassiveByID(asString(a.Fields[passiveIDKey])); ok {
		total.Build += cfg.PassiveFlatCost(p)
		fc := FieldsCost(cfg, p.Fields, passiveFieldValues(a))
		total.Build += fc.Build
		total.Energy += fc.Energy
	}

	// Track the first *present* enactment rather than relying on slice index.
	// Enactments can be removed and re-added in the builder, so the first slot
	// is not guaranteed to hold the first real enactment. The additional-
	// enactment surcharge and the first-enactment base-cost waiver both apply
	// based on this running count of enactments that actually have a type.
	present := 0
	for _, en := range a.Enactments {
		if en.Type == "" {
			continue
		}
		if present > 0 {
			total.plus(cfg.AdditionalEnactment.AsCost())
		}
		if ec, ok := cfg.Enactment(en.Type); ok {
			c := ComponentCost(cfg, ec, en.Fields)
			// The first enactment is free to add: its component base_cost
			// is waived (field-driven costs still apply). Subsequent
			// enactments pay their full base cost.
			if present == 0 {
				c.Build -= ec.BaseCost.BuildCost
				c.Energy -= ec.BaseCost.EnergyCost
			}
			total.Build += c.Build
			total.Energy += c.Energy
		}
		// An enactment owns its target when it is the first one, or when the
		// author ticked "different target than the enactment before it". Only
		// an enactment that owns its target has an Interaction and a
		// Validation, so only then do those regions cost anything: the others
		// simply reuse the previous enactment's target for free.
		ownsTarget := present == 0 || en.NewTarget
		if present > 0 && en.NewTarget {
			total.plus(cfg.AdditionalEnactment.NewTarget.AsCost())
		}
		present++

		if ownsTarget && en.Interaction != "" {
			if ic, ok := cfg.Interaction(en.Interaction); ok {
				c := ComponentCost(cfg, ic, en.InteractionData)
				total.Build += c.Build
				total.Energy += c.Energy
			}
		}
		// Validation (engagement/counter) fields also contribute cost.
		if ownsTarget && len(cfg.Validations.Fields) > 0 {
			c := FieldsCost(cfg, cfg.Validations.Fields, en.ValidationData)
			total.Build += c.Build
			total.Energy += c.Energy
		}
	}

	// Apply the configured cost floors. Only the final total is clamped, so
	// refund-style options (energy offsets, Enact Nerf, negative-cost
	// knockouts) still offset other costs internally; they just cannot make an
	// perk cost less than nothing unless the ruleset opts in.
	if !cfg.AllowsNegativeBuildCost() && total.Build < 0 {
		total.Build = 0
	}
	// Using a perk always costs at least 1 energy, so the floor is 1 rather
	// than 0. The per-enactment energy cost comes from the perk type's
	// base_cost plus additional_enactment, so this is a backstop that keeps
	// refund-style options (Enact Nerf, energy offsets) from making a perk free.
	// A ruleset that opts into negative energy cost (via
	// allow_negative_energy_cost) keeps whatever the options computed.
	//
	// Passives are exempt from the energy floor: they are always on, so there is
	// no moment at which energy would be paid, and forcing them to 1 would make
	// every passive look like it had a running cost.
	if !cfg.AllowsNegativeEnergyCost() && total.Energy < 1 && !isPassive(cfg, a) {
		total.Energy = 1
	}

	return total
}

// isPassive reports whether a perk is a predefined passive. It keys off the
// catalogue id stored on the perk rather than off its type, because a passive is
// not a perk type: it is picked from the passive list, so there is no perk-type
// component to inspect. An id that no longer resolves is still treated as a
// passive, so removing an entry from the catalogue cannot retroactively saddle a
// character's passive with an energy cost.
func isPassive(cfg *config.Config, a model.Perk) bool {
	return asString(a.Fields[passiveIDKey]) != ""
}

// passiveIDKey is where a passive's catalogue id is stored on the perk. It is
// the discriminator that marks a perk as a passive at all.
const passiveIDKey = "passive_id"

// passiveFieldsKey is where a passive's configured field values are stored on
// the perk. It is a nested map so the passive's own field keys cannot collide
// with the surrounding perk fields.
const passiveFieldsKey = "passive_fields"

// passiveFieldValues returns the configured values of a passive perk, or nil
// when it has none yet. A nil map is fine: FieldsCost falls back to each field's
// default, which is what an unconfigured passive is priced at.
func passiveFieldValues(a model.Perk) map[string]any {
	if a.Fields == nil {
		return nil
	}
	v, _ := a.Fields[passiveFieldsKey].(map[string]any)
	return v
}
