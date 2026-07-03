package export

import (
	"fmt"
	"strings"

	"github.com/harmey/blok2ttrpg/ability-builder/internal/models"
)

// ToYAML converts an ability to a YAML string matching the new self-contained
// card schema. Perks are no longer emitted.
func ToYAML(a *models.Ability) string {
	var b strings.Builder

	b.WriteString("ability:\n")
	b.WriteString(fmt.Sprintf("  type: %s\n", a.Type))

	if a.Name != "" {
		b.WriteString(fmt.Sprintf("  name: %s\n", a.Name))
	}
	if a.Description != "" {
		b.WriteString(fmt.Sprintf("  description: %s\n", a.Description))
	}

	if a.HasItemDependency {
		b.WriteString("  has_item_dependency: Yes\n")
		if a.ItemName != "" {
			b.WriteString(fmt.Sprintf("  item_name: %s\n", a.ItemName))
		}
	} else {
		b.WriteString("  has_item_dependency: No\n")
	}

	b.WriteString(fmt.Sprintf("  energy_cost: %d\n", a.EnergyCost))
	b.WriteString(fmt.Sprintf("  action_cost: %d\n", a.ActionCost))

	switch a.Type {
	case models.AbilityExecution:
		b.WriteString(fmt.Sprintf("  energy_adjustment: %+d\n", a.EnergySteps))
		b.WriteString(fmt.Sprintf("  action_adjustment: %+d\n", a.ActionSteps))
	case models.AbilityReaction:
		b.WriteString(fmt.Sprintf("  range: %dm\n", a.ReactionRange))
		b.WriteString(fmt.Sprintf("  uses: %d\n", a.ReactionUses))
		if a.TriggerTrait != "" {
			b.WriteString(fmt.Sprintf("  trigger: %s of type %s\n", a.Trigger, a.TriggerTrait))
		} else if a.Trigger != "" {
			b.WriteString(fmt.Sprintf("  trigger: %s\n", a.Trigger))
		}
	case models.AbilityPhase:
		b.WriteString(fmt.Sprintf("  phase_duration: %d rounds\n", a.PhaseDuration))
		b.WriteString(fmt.Sprintf("  reverse_phase_duration: %d rounds\n", a.ReversePhaseRounds))
		if a.AllKnockoutsReq {
			b.WriteString("  all_knockout_requirements: true\n")
		}
		if a.ReverseKnockoutOK {
			b.WriteString("  knockout_on_reverse_phase: true\n")
		}
		if len(a.Knockouts) > 0 {
			b.WriteString("  knockout_requirements:\n")
			for _, k := range a.Knockouts {
				b.WriteString(fmt.Sprintf("    - %s\n", k))
			}
		}
	case models.AbilityMinion:
		b.WriteString(fmt.Sprintf("  health: %d\n", 10+a.HPBonus*5))
		b.WriteString(fmt.Sprintf("  lifetime: %d rounds\n", 3+a.ExtraLifetime))
	}

	if len(a.Enactments) > 0 {
		b.WriteString("  enactments:\n")
		for i, e := range a.Enactments {
			if i > 0 {
				b.WriteString("    # additional enactment (+1 build)\n")
			}
			writeYAMLEnactment(&b, "    ", e)
		}
	}

	return b.String()
}

func writeYAMLEnactment(b *strings.Builder, indent string, e models.Enactment) {
	b.WriteString(fmt.Sprintf("%s- type: %s\n", indent, e.Type))
	if e.Always {
		b.WriteString(fmt.Sprintf("%s  always_resolve: true\n", indent))
	}
	if e.BuildCost > 0 {
		b.WriteString(fmt.Sprintf("%s  build_cost: %d\n", indent, e.BuildCost))
	}
	if e.CastCost > 0 {
		b.WriteString(fmt.Sprintf("%s  cast_cost: %d\n", indent, e.CastCost))
	}
	if e.Formula != "" {
		b.WriteString(fmt.Sprintf("%s  formula: %s\n", indent, e.Formula))
	}

	switch e.Type {
	case models.EnactDamage:
		b.WriteString(fmt.Sprintf("%s  source: %s\n", indent, sourceLabel(e)))
		if e.Source == "trait" {
			b.WriteString(fmt.Sprintf("%s  trait: %s\n", indent, orDefault(e.SourceTrait, "(choose a trait)")))
		}
		if e.Source == "other" {
			b.WriteString(fmt.Sprintf("%s  other_text: %s\n", indent, orDefault(e.OtherRollText, "(roll reference)")))
		}
		if e.FlatBonus > 0 {
			b.WriteString(fmt.Sprintf("%s  flat_bonus: +%d\n", indent, e.FlatBonus))
		}
		if e.OffensiveTrait != "" {
			b.WriteString(fmt.Sprintf("%s  offensive_trait: %s\n", indent, e.OffensiveTrait))
		}
	case models.EnactHealing:
		b.WriteString(fmt.Sprintf("%s  source: %s\n", indent, sourceLabel(e)))
		if e.Source == "trait" {
			b.WriteString(fmt.Sprintf("%s  trait: %s\n", indent, orDefault(e.SourceTrait, "(choose a trait)")))
		}
		if e.Source == "other" {
			b.WriteString(fmt.Sprintf("%s  other_text: %s\n", indent, orDefault(e.OtherRollText, "(roll reference)")))
		}
		if e.FlatBonus > 0 {
			b.WriteString(fmt.Sprintf("%s  flat_bonus: +%d\n", indent, e.FlatBonus))
		}
		if e.MedicineTrait != "" {
			b.WriteString(fmt.Sprintf("%s  medicine: true\n", indent))
		}
	case models.EnactMovement:
		b.WriteString(fmt.Sprintf("%s  origin: %s\n", indent, movementOrigin(e)))
		b.WriteString(fmt.Sprintf("%s  distance: %dm\n", indent, e.Distance))
		if len(e.Directions) > 0 {
			b.WriteString(fmt.Sprintf("%s  directions:\n", indent))
			for _, d := range e.Directions {
				b.WriteString(fmt.Sprintf("%s    - %s\n", indent, d))
			}
		}
	case models.EnactProficiencyShift:
		b.WriteString(fmt.Sprintf("%s  trait: %s\n", indent, orDefault(e.ShiftedTrait, "(choose a trait)")))
		b.WriteString(fmt.Sprintf("%s  direction: %s\n", indent, e.ShiftDir))
		b.WriteString(fmt.Sprintf("%s  amount: %d\n", indent, e.ShiftAmount))
		b.WriteString(fmt.Sprintf("%s  uses: %d\n", indent, e.ShiftUses))
	case models.EnactPersistentEffect:
		b.WriteString(fmt.Sprintf("%s  name: %s\n", indent, orDefault(e.EffectName, "Effect")))
		b.WriteString(fmt.Sprintf("%s  applies: %s\n", indent, e.EffectType))
		b.WriteString(fmt.Sprintf("%s  duration: %d rounds\n", indent, e.Duration))
		b.WriteString(fmt.Sprintf("%s  trigger: %s\n", indent, e.TriggerTiming))
		if len(e.Solutions) > 0 {
			b.WriteString(fmt.Sprintf("%s  solutions:\n", indent))
			for _, s := range e.Solutions {
				b.WriteString(fmt.Sprintf("%s    - %s\n", indent, s))
			}
		}
	}

	if e.Interaction != nil {
		writeYAMLInteraction(b, indent+"  ", *e.Interaction)
	}
}

func movementOrigin(e models.Enactment) string {
	if e.OriginMode == "other" && e.OriginText != "" {
		return e.OriginText
	}
	return "Engager"
}

func sourceLabel(e models.Enactment) string {
	switch e.Source {
	case "trait":
		return "1d10 (trait)"
	case "other":
		return "another roll result"
	case "":
		return "1d4"
	default:
		return "1" + e.Source
	}
}

func orDefault(v, d string) string {
	if v == "" {
		return d
	}
	return v
}

func writeYAMLInteraction(b *strings.Builder, indent string, i models.Interaction) {
	b.WriteString(fmt.Sprintf("%sinteraction:\n", indent))
	b.WriteString(fmt.Sprintf("%s  type: %s\n", indent, i.Type))
	if i.UsePrevious {
		b.WriteString(fmt.Sprintf("%s  use_previous: true\n", indent))
	}
	if i.BuildCost > 0 {
		b.WriteString(fmt.Sprintf("%s  build_cost: %d\n", indent, i.BuildCost))
	}
	if i.CastCost > 0 {
		b.WriteString(fmt.Sprintf("%s  cast_cost: %d\n", indent, i.CastCost))
	}

	switch i.Type {
	case models.InteractionSelf:
		b.WriteString(fmt.Sprintf("%s  target: Self\n", indent))
		b.WriteString(fmt.Sprintf("%s  counter: d8\n", indent))
	case models.InteractionDirect:
		b.WriteString(fmt.Sprintf("%s  targets: %d\n", indent, i.Targets))
		b.WriteString(fmt.Sprintf("%s  range: %dm\n", indent, i.Range))
	case models.InteractionRanged:
		b.WriteString(fmt.Sprintf("%s  targets: %d\n", indent, i.Targets))
		b.WriteString(fmt.Sprintf("%s  range: %dm\n", indent, i.Range))
		b.WriteString(fmt.Sprintf("%s  target_may_be_not_visible: %t\n", indent, i.VisibleOK))
		b.WriteString(fmt.Sprintf("%s  target_may_be_obstructed: %t\n", indent, i.ObstructedOK))
		b.WriteString(fmt.Sprintf("%s  remove_engagement_penalty: %t\n", indent, i.RemovePenalty))
	case models.InteractionArea:
		b.WriteString(fmt.Sprintf("%s  radius: %dm\n", indent, i.Radius))
		b.WriteString(fmt.Sprintf("%s  range: %dm\n", indent, i.Range))
		b.WriteString(fmt.Sprintf("%s  origin: %s\n", indent, interactionOrigin(i)))
	case models.InteractionAreaOfEffect:
		b.WriteString(fmt.Sprintf("%s  radius: %dm\n", indent, i.Radius))
		b.WriteString(fmt.Sprintf("%s  range: %dm\n", indent, i.Range))
		b.WriteString(fmt.Sprintf("%s  origin: %s\n", indent, interactionOrigin(i)))
		b.WriteString(fmt.Sprintf("%s  duration: %d rounds\n", indent, i.Duration))
		b.WriteString(fmt.Sprintf("%s  timing: %s\n", indent, i.Timing))
		if i.Immune {
			b.WriteString(fmt.Sprintf("%s  engager_immune: true\n", indent))
		}
	}

	if i.Validation != nil {
		writeYAMLValidation(b, indent+"  ", *i.Validation)
	} else {
		b.WriteString(fmt.Sprintf("%s  validation: n/a\n", indent))
	}
}

func interactionOrigin(i models.Interaction) string {
	if i.OriginMode == "other" && i.OriginText != "" {
		return i.OriginText
	}
	return "Engager"
}

func writeYAMLValidation(b *strings.Builder, indent string, v models.Validation) {
	b.WriteString(fmt.Sprintf("%svalidation:\n", indent))
	if v.BuildCost > 0 {
		b.WriteString(fmt.Sprintf("%s  build_cost: %d\n", indent, v.BuildCost))
	}
	if v.CastCost > 0 {
		b.WriteString(fmt.Sprintf("%s  cast_cost: %d\n", indent, v.CastCost))
	}
	switch v.EngageMode {
	case "", "trait":
		catLabel := "offensive trait"
		switch v.EngageTraitCategory {
		case models.TraitCategoryGeneral:
			catLabel = "general trait"
		case models.TraitCategoryDefense:
			catLabel = "defensive trait"
		}
		b.WriteString(fmt.Sprintf("%s  engagement_roll: %s (%s)\n", indent, orDefault(v.EngageTrait, "Precision"), catLabel))
	case "generic":
		b.WriteString(fmt.Sprintf("%s  engagement_roll: %s (generic)\n", indent, v.EngageDie))
	case "other":
		b.WriteString(fmt.Sprintf("%s  engagement_roll: %s\n", indent, orDefault(v.EngageOther, "(roll reference)")))
	case "previous":
		b.WriteString(fmt.Sprintf("%s  engagement_roll: result of previous\n", indent))
	}
	if len(v.CounterRolls) == 1 {
		b.WriteString(fmt.Sprintf("%s  counter_roll: %s\n", indent, v.CounterRolls[0]))
	} else if len(v.CounterRolls) > 1 {
		b.WriteString(fmt.Sprintf("%s  counter_roll:\n", indent))
		for _, c := range v.CounterRolls {
			b.WriteString(fmt.Sprintf("%s    - %s\n", indent, c))
		}
	}
}
