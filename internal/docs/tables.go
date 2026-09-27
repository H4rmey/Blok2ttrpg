// This file holds the config-derived reference tables that the rulebook used to
// hand-maintain or that silently rendered empty.
//
// The leveling tables are the important case. The point budgets stopped being a
// hand-written `levels:` list and became the formula start + per_level*(level-1),
// computed by config.SkillPointBudget / AbilityPointBudget. The docs still
// ranged over the now-absent list, so the table rendered as a bare header with
// no rows. Generating the rows by asking the config for each level's budget
// keeps the documented numbers identical to the numbers the application serves,
// and works whether the budget comes from the formula or from an explicit
// per-level override.
package docs

import (
	"fmt"
	"strings"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
)

// levelingTable renders the point budget for every level as Level / Points
// Gained / Total. The pool argument selects which budget to render: "skill" for
// skill points or "ability" for perk points.
//
// Budgets are read through the config's budget accessors, which are the same
// ones the application uses when it hands a character its points, so the table
// cannot drift from play.
func levelingTable(cfg *config.Config, pool string) string {
	if cfg == nil {
		return "_No leveling configured._"
	}
	budget, label, ok := budgetAccessor(cfg, pool)
	if !ok {
		return fmt.Sprintf("_Unknown point pool %q._", pool)
	}
	max := cfg.MaxLevel()
	if max < 1 {
		return "_No leveling configured._"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "| Level | Points Gained | Total %s |\n", label)
	b.WriteString("| --- | --- | --- |\n")
	prev := 0
	for level := 1; level <= max; level++ {
		total := budget(level)
		gained := total - prev
		// Level 1 is a starting pool rather than a gain, so it is shown as the
		// full amount without a plus sign.
		if level == 1 {
			fmt.Fprintf(&b, "| **%d** | %d (starting) | %d |\n", level, total, total)
		} else {
			fmt.Fprintf(&b, "| **%d** | %s | %d |\n", level, signed(gained), total)
		}
		prev = total
	}
	return strings.TrimSpace(b.String())
}

// budgetAccessor resolves a pool name to the config accessor that computes its
// budget, plus the reader-facing name of the pool.
func budgetAccessor(cfg *config.Config, pool string) (func(int) int, string, bool) {
	switch strings.ToLower(strings.TrimSpace(pool)) {
	// "trait"/"traits" are accepted as legacy aliases: the skill pool was
	// called the trait pool before the rename, and doc templates may still
	// ask for it under the old name.
	case "skill", "skills", "trait", "traits":

		return cfg.SkillPointBudget, "Skill Points", true
	case "ability", "abilities", "perk", "perks":
		return cfg.AbilityPointBudget, "Ability Points", true
	}
	return nil, "", false
}

// proficiencyDiceTable renders the proficiency ladder as Tier / Die. It replaces
// a hand-written table in the dice chapter that had drifted badly: it still
// listed a "Clumsy" tier and a d20 top end, neither of which exists in the
// ladder any more.
func proficiencyDiceTable(cfg *config.Config) string {
	if cfg == nil || len(cfg.Proficiencies) == 0 {
		return "_No proficiencies configured._"
	}
	var b strings.Builder
	b.WriteString("| Proficiency | Die |\n")
	b.WriteString("| --- | --- |\n")
	for _, p := range cfg.Proficiencies {
		fmt.Fprintf(&b, "| %s | %s |\n", orDash(p.Name), orDash(p.Die))
	}
	return strings.TrimSpace(b.String())
}

// lowestDie returns the die of the first rung of the ladder, used so the dice
// chapter's prose ("ranging from X to Y") follows the config.
func lowestDie(cfg *config.Config) string {
	if cfg == nil || len(cfg.Proficiencies) == 0 {
		return "-"
	}
	return orDash(cfg.Proficiencies[0].Die)
}

// highestDie returns the die of the last rung of the ladder.
func highestDie(cfg *config.Config) string {
	if cfg == nil || len(cfg.Proficiencies) == 0 {
		return "-"
	}
	return orDash(cfg.Proficiencies[len(cfg.Proficiencies)-1].Die)
}

// defaultProficiencyName returns the reader-facing name of the free baseline
// rung, so docs can name it without hardcoding "Untrained".
func defaultProficiencyName(cfg *config.Config) string {
	if cfg == nil {
		return "-"
	}
	if p, ok := cfg.Proficiency(cfg.DefaultProficiencyID()); ok {
		return orDash(p.Name)
	}
	return "-"
}

// rulesFlagsTable documents the cost floors and budget enforcement switches.
// These change what a player is allowed to do (whether a perk can end up
// refunding more than it costs, and whether a character may overspend its skill
// points), but they appeared nowhere in the rulebook.
func rulesFlagsTable(cfg *config.Config) string {
	if cfg == nil {
		return "_No rules flags configured._"
	}
	var b strings.Builder
	b.WriteString("| Rule | In effect |\n")
	b.WriteString("| --- | --- |\n")
	// The accessors apply the documented defaults for these optional flags, so
	// the table reports the behaviour that is actually in force rather than
	// whether the key happens to be present in the YAML.
	fmt.Fprintf(&b, "| Build point refunds | %s |\n", floorPhrase(cfg.AllowsNegativeBuildCost(), "build"))
	fmt.Fprintf(&b, "| Energy cost refunds | %s |\n", floorPhrase(cfg.AllowsNegativeEnergyCost(), "energy"))
	fmt.Fprintf(&b, "| Skill point budget | %s |\n", skillBudgetPhrase(cfg.AllowsNegativeSkillPoints()))
	fmt.Fprintf(&b, "| Maximum level | A character's level is capped at %d. |\n", cfg.MaxLevel())
	return strings.TrimSpace(b.String())
}

// floorPhrase explains whether a total may go below zero.
func floorPhrase(allowNegative bool, unit string) string {
	if allowNegative {
		return fmt.Sprintf("Drawbacks may refund more %s than the rest of the ability costs, so a total can go below zero.", unit)
	}
	return fmt.Sprintf("Drawbacks refund %s points, but never past zero: an ability's total %s cost stops at nothing.", unit, unit)
}

// skillBudgetPhrase explains whether the skill point budget is enforced.
func skillBudgetPhrase(allowNegative bool) string {
	if allowNegative {
		return "A character may spend more skill points than its level grants."
	}
	return "A character may not spend more skill points than its level grants; edits and package imports that would overspend are refused."
}

// enactmentSurchargeTable documents everything about attaching more than one
// enactment to an ability: the flat surcharge, whether the extra enactments must
// declare their own interaction and validation, and the opt-in surcharge for
// giving an extra enactment its own target. The new-target rule and the two
// requirement flags were previously undocumented.
func enactmentSurchargeTable(cfg *config.Config) string {
	if cfg == nil {
		return "_No enactment rules configured._"
	}
	ae := cfg.AdditionalEnactment
	var b strings.Builder
	b.WriteString("| Rule | Build Cost | Energy Cost | Description |\n")
	b.WriteString("| --- | --- | --- | --- |\n")
	fmt.Fprintf(&b, "| **Each enactment beyond the first** | %d | %d | %s |\n",
		ae.BuildCost, ae.EnergyCost, orDash(ae.Description))
	// NewTarget is a value, so it is considered configured when it carries a
	// cost or any descriptive text.
	nt := ae.NewTarget
	hasNewTarget := nt.BuildCost != 0 || nt.EnergyCost != 0 ||
		strings.TrimSpace(nt.Label) != "" || strings.TrimSpace(nt.Description) != ""
	if hasNewTarget {
		fmt.Fprintf(&b, "| **%s** | %d | %d | %s |\n",
			orDash(nt.DisplayLabel()), nt.BuildCost, nt.EnergyCost, orDash(nt.Description))
	}
	b.WriteString("\n")
	b.WriteString("The first enactment of an ability always declares who it affects and what roll resolves it. ")
	b.WriteString(requirementPhrase(ae.RequiresInteraction(), ae.RequiresValidation()))
	if hasNewTarget {
		b.WriteString(" Taking the separate-target option on a later enactment gives that enactment its own interaction and validation, ")
		b.WriteString("so their costs apply on top of the surcharge above.")
	}
	return strings.TrimSpace(b.String())
}

// requirementPhrase explains whether enactments after the first must declare
// their own interaction and validation, or inherit them from the one before.
func requirementPhrase(requireInteraction, requireValidation bool) string {
	switch {
	case requireInteraction && requireValidation:
		return "Every enactment after the first must also declare its own interaction and validation."
	case requireInteraction:
		return "Every enactment after the first must declare its own interaction, but inherits the validation of the enactment before it."
	case requireValidation:
		return "Every enactment after the first must declare its own validation, but inherits the target of the enactment before it."
	default:
		return "Enactments after the first inherit both the target and the resolving roll of the enactment before them."
	}
}
