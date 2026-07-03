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

	energy := cfg.BaseEnergy
	action := cfg.BaseAction

	switch a.Type {
	case models.AbilityExecution:
		energySteps := atoi(values.Get("energy_steps"))
		actionSteps := atoi(values.Get("action_steps"))
		a.EnergySteps = energySteps
		a.ActionSteps = actionSteps

		if energySteps > 0 {
			energy += energySteps * stepEnergyCost(cfg, "energy", 1)
		} else if energySteps < 0 {
			energy += (-energySteps) * stepEnergyCost(cfg, "energy", -1)
		}

		action += actionSteps

		// Reducing action cost adds energy (per YAML step cost).
		if actionSteps < 0 {
			energy += -actionSteps * stepEnergyCost(cfg, "action", -1)
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
			energy += (a.ReactionRange - cfg.BaseRange) * cfg.RangeCost.EnergyCost
		}
		if a.ReactionUses > cfg.BaseUses {
			energy += (a.ReactionUses - cfg.BaseUses) * cfg.UsesCost.EnergyCost
		}

	case models.AbilityPhase:
		a.PhaseDuration = max(atoi(values.Get("phase_rounds")), cfg.BaseDuration)
		a.ReversePhaseRounds = max(atoi(values.Get("reverse_rounds")), 1)
		if a.PhaseDuration > cfg.BaseDuration {
			energy += (a.PhaseDuration - cfg.BaseDuration) * cfg.DurationCost.EnergyCost
		}

	case models.AbilityMinion:
		a.HPBonus = atoi(values.Get("hp"))
		a.ExtraLifetime = atoi(values.Get("life"))
		energy = cfg.BaseEnergy + a.ExtraLifetime*cfg.LifetimeBonusCost.EnergyCost
		action = cfg.BaseAction
	}

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
