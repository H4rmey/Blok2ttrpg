package config

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/harmey/blok2ttrpg/ability-builder/internal/models"
)

// AbilityCostResult holds the computed costs for an ability type.
type AbilityCostResult struct {
	EnergyCost int
	ActionCost int
}

// ComputeAbilityCosts returns the energy and action cost for an ability based on
// the YAML configuration and submitted form values. It uses the same rules as the
// frontend live calculator so the server stays in sync with the displayed values.
func (ab *AbilityBuilderConfig) ComputeAbilityCosts(a *models.Ability, values url.Values) error {
	cfg, ok := ab.AbilityTypes[strings.ToLower(string(a.Type))]
	if !ok {
		return fmt.Errorf("unknown ability type: %s", a.Type)
	}

	if len(cfg.Fields) > 0 {
		return ab.ComputeAbilityTypeCosts(a, values)
	}

	energy := cfg.BaseEnergy
	action := cfg.BaseAction
	build := 0
	return ab.computeAbilityCostsLegacy(a, values, cfg, energy, action, build)
}

// computeAbilityCostsLegacy contains the original perk-based cost logic used
// when the ability type config does not declare generic fields.
func (ab *AbilityBuilderConfig) computeAbilityCostsLegacy(a *models.Ability, values url.Values, cfg AbilityTypeConfig, energy int, action int, build int) error {
	if values.Get("ability_item_dep") == "on" {
		build += perkAddCost(cfg.Perks, "item_dependency")
	}

	switch a.Type {
	case models.AbilityExecution:
		energySteps := atoi(values.Get("energy_steps"))
		actionSteps := atoi(values.Get("action_steps"))
		a.EnergySteps = energySteps
		a.ActionSteps = actionSteps

		if energySteps > 0 {
			energy += energySteps * stepEnergyCost(cfg, "energy", 1)
			build += energySteps * stepAddCost(cfg, "energy", 1)
		} else if energySteps < 0 {
			energy += (-energySteps) * stepEnergyCost(cfg, "energy", -1)
			build += (-energySteps) * stepAddCost(cfg, "energy", -1)
		}

		action += actionSteps

		if actionSteps > 0 {
			energy += actionSteps * stepEnergyCost(cfg, "action", 1)
			build += actionSteps * stepAddCost(cfg, "action", 1)
		}
		if actionSteps < 0 {
			energy += -actionSteps * stepEnergyCost(cfg, "action", -1)
			build += -actionSteps * stepAddCost(cfg, "action", -1)
		}
		if action < 0 {
			action = 0
		}

	case models.AbilityReaction:
		a.ReactionRange = atoi(values.Get("range"))
		a.ReactionUses = atoi(values.Get("uses"))
		a.Trigger = values.Get("trigger")
		if a.Trigger == "Target makes a trait check" {
			a.TriggerTrait = values.Get("trigger_trait")
		}
		if a.ReactionRange > cfg.BaseRange {
			build += (a.ReactionRange - cfg.BaseRange) * cfg.RangeCost.AddCost
			energy += (a.ReactionRange - cfg.BaseRange) * cfg.RangeCost.EnergyCost
		}
		if a.ReactionUses > cfg.BaseUses {
			build += (a.ReactionUses - cfg.BaseUses) * cfg.UsesCost.AddCost
			energy += (a.ReactionUses - cfg.BaseUses) * cfg.UsesCost.EnergyCost
		}
		build += triggerAddCost(cfg.Triggers, a.Trigger)

	case models.AbilityPhase:
		a.PhaseDuration = max(atoi(values.Get("phase_rounds")), cfg.BaseDuration)
		a.ReversePhaseRounds = max(atoi(values.Get("reverse_rounds")), 1)
		if a.PhaseDuration > cfg.BaseDuration {
			build += (a.PhaseDuration - cfg.BaseDuration) * cfg.DurationCost.AddCost
			energy += (a.PhaseDuration - cfg.BaseDuration) * cfg.DurationCost.EnergyCost
		}
		if a.ReversePhaseRounds < a.PhaseDuration {
			build += (a.PhaseDuration - a.ReversePhaseRounds) * cfg.ReverseDurationRefund.AddCost
			energy += (a.PhaseDuration - a.ReversePhaseRounds) * cfg.ReverseDurationRefund.EnergyCost
		}
		if values.Get("all_req") == "on" {
			build += perkAddCost(cfg.Perks, "all_knockouts_req")
			energy += perkEnergyCost(cfg.Perks, "all_knockouts_req")
		}
		if values.Get("reverse_knockout") == "on" {
			build += perkAddCost(cfg.Perks, "reverse_knockout")
			energy += perkEnergyCost(cfg.Perks, "reverse_knockout")
		}
		if values.Get("no_knockout") == "on" {
			build += perkAddCost(cfg.Perks, "no_knockout")
			energy += perkEnergyCost(cfg.Perks, "no_knockout")
		}

	case models.AbilityMinion:
		a.HPBonus = atoi(values.Get("hp"))
		a.ExtraLifetime = atoi(values.Get("life"))
		build += a.HPBonus*cfg.HealthBonusCost.AddCost + a.ExtraLifetime*cfg.LifetimeBonusCost.AddCost
		energy = cfg.BaseEnergy + a.HPBonus*cfg.HealthBonusCost.EnergyCost + a.ExtraLifetime*cfg.LifetimeBonusCost.EnergyCost
		action = cfg.BaseAction

	case models.AbilityPreparation:
		a.ReactionRange = atoi(values.Get("range"))
		a.ReactionUses = atoi(values.Get("uses"))
		a.Trigger = values.Get("trigger")
		if a.Trigger == "Target makes a trait check" {
			a.TriggerTrait = values.Get("trigger_trait")
		}
		if a.ReactionRange > cfg.BaseRange {
			build += (a.ReactionRange - cfg.BaseRange) * cfg.RangeCost.AddCost
			energy += (a.ReactionRange - cfg.BaseRange) * cfg.RangeCost.EnergyCost
		}
		if a.ReactionUses > cfg.BaseUses {
			build += (a.ReactionUses - cfg.BaseUses) * cfg.UsesCost.AddCost
			energy += (a.ReactionUses - cfg.BaseUses) * cfg.UsesCost.EnergyCost
		}
		build += triggerAddCost(cfg.Triggers, a.Trigger)

		actionSteps := atoi(values.Get("action_steps"))
		energySteps := atoi(values.Get("energy_steps"))
		a.ActionSteps = actionSteps
		a.EnergySteps = energySteps

		if actionSteps > 0 {
			energy += actionSteps * stepEnergyCost(cfg, "action", 1)
			build += actionSteps * stepAddCost(cfg, "action", 1)
		} else if actionSteps < 0 {
			energy += (-actionSteps) * stepEnergyCost(cfg, "action", -1)
			build += (-actionSteps) * stepAddCost(cfg, "action", -1)
		}
		if action < 0 {
			action = 0
		}

		if energySteps > 0 {
			energy += energySteps * stepEnergyCost(cfg, "energy", 1)
			build += energySteps * stepAddCost(cfg, "energy", 1)
		} else if energySteps < 0 {
			energy += (-energySteps) * stepEnergyCost(cfg, "energy", -1)
			build += (-energySteps) * stepAddCost(cfg, "energy", -1)
		}

	case models.AbilityConcentration:
		a.Effortless = values.Get("effortless") == "on"
		a.IronWill = values.Get("iron_will") == "on"
		a.DualFocus = values.Get("dual_focus") == "on"
		if a.Effortless {
			build += perkAddCost(cfg.Perks, "effortless")
			energy += perkEnergyCost(cfg.Perks, "effortless")
		}
		if a.IronWill {
			build += perkAddCost(cfg.Perks, "iron_will")
			energy += perkEnergyCost(cfg.Perks, "iron_will")
		}
		if a.DualFocus {
			build += perkAddCost(cfg.Perks, "dual_focus")
			energy += perkEnergyCost(cfg.Perks, "dual_focus")
		}
		energySteps := atoi(values.Get("energy_steps"))
		a.EnergySteps = energySteps
		if energySteps > 0 {
			energy += energySteps * stepEnergyCost(cfg, "energy", 1)
			build += energySteps * stepAddCost(cfg, "energy", 1)
		} else if energySteps < 0 {
			energy += (-energySteps) * stepEnergyCost(cfg, "energy", -1)
			build += (-energySteps) * stepAddCost(cfg, "energy", -1)
		}
	}

	a.BuildCost = build
	a.EnergyCost = energy
	a.ActionCost = action
	return nil
}

// stepEnergyCost returns the energy cost per step for the given direction and step type.
func stepEnergyCost(cfg AbilityTypeConfig, stepType string, direction int) int {
	steps, ok := cfg.StepCosts[stepType]
	if !ok {
		return 0
	}
	if direction >= 0 {
		return steps.Increase.EnergyCost
	}
	return steps.Decrease.EnergyCost
}

func stepAddCost(cfg AbilityTypeConfig, stepType string, direction int) int {
	steps, ok := cfg.StepCosts[stepType]
	if !ok {
		return 0
	}
	if direction >= 0 {
		return steps.Increase.AddCost
	}
	return steps.Decrease.AddCost
}

func perkAddCost(perks []PerkConfig, id string) int {
	for _, p := range perks {
		if p.ID == id {
			return p.AddCost
		}
	}
	return 0
}

func perkEnergyCost(perks []PerkConfig, id string) int {
	for _, p := range perks {
		if p.ID == id {
			return p.EnergyCost
		}
	}
	return 0
}

func triggerAddCost(triggers []TriggerConfig, id string) int {
	for _, t := range triggers {
		if t.ID == id {
			return t.AddCost
		}
	}
	return 0
}

func atoi(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
