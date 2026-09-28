// Config-derived reference table for the passives chapter.
//
// As with every other table in this package, the numbers are read from the
// loaded ruleset rather than written into the markdown: a passive's cost appears
// in the config, the picker and the rulebook, and three copies of a number
// eventually disagree.
package docs

import (
	"fmt"
	"strings"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
)

// passivesTable renders the passive catalogue: name, cost at its defaults, the
// parts a player configures, and the rules text.
//
// The description is rendered at the entry's defaults, because the table
// describes a passive as it is first taken; the Configure column states what can
// be changed and what changing it costs.
func passivesTable(cfg *config.Config) string {
	if cfg == nil || len(cfg.Passives.Entries) == 0 {
		return "_No passives configured._"
	}
	var b strings.Builder
	b.WriteString("| Passive | Cost | Configure | Effect |\n")
	b.WriteString("| --- | --- | --- | --- |\n")
	for _, p := range cfg.Passives.Entries {
		defaults := cfg.PassiveDefaults(p)
		fmt.Fprintf(&b, "| **%s** | %s | %s | %s |\n",
			orDash(p.Name),
			perkPointWords(cfg.PassiveFlatCost(p)),
			passiveFieldWords(cfg, p),
			orDash(oneLine(cfg.PassiveDescription(p, defaults))))
	}
	return strings.TrimSpace(b.String())
}

// passiveFieldWords lists an entry's configurable parts and what each costs, or
// says the passive is fixed. The cost is what the reader needs in order to plan a
// character, so a field that charges nothing is called out as free rather than
// left ambiguous.
func passiveFieldWords(cfg *config.Config, p config.Passive) string {
	if !p.Configurable() {
		return "Nothing to configure"
	}
	parts := make([]string, 0, len(p.Fields))
	for _, f := range p.Fields {
		label := f.Label
		if label == "" {
			label = f.Key
		}
		parts = append(parts, fmt.Sprintf("%s (%s)", label, passiveFieldCostWords(f)))
	}
	return strings.Join(parts, "; ")
}

// passiveFieldCostWords describes what one configurable field costs, in the terms
// that field type is actually priced in.
func passiveFieldCostWords(f config.Field) string {
	switch f.Type {
	case "free_number":
		if f.PerStep != nil && f.PerStep.Increase != nil && f.PerStep.Increase.BuildCost != 0 {
			return fmt.Sprintf("%d to %d, %s per step above %s",
				f.Min, f.Max,
				perkPointWords(f.PerStep.Increase.BuildCost),
				orDash(fmt.Sprintf("%v", f.Default)))
		}
		return fmt.Sprintf("%d to %d, free", f.Min, f.Max)
	case "checkbox":
		if f.Cost != nil && f.Cost.BuildCost != 0 {
			return fmt.Sprintf("optional, %s", perkPointWords(f.Cost.BuildCost))
		}
		return "optional, free"
	case "free_text":
		return "free text, free"
	default:
		if f.Cost != nil && f.Cost.BuildCost != 0 {
			return perkPointWords(f.Cost.BuildCost)
		}
		return "free"
	}
}

// passiveRules states the category-wide rules: what a passive costs to use, and
// whether one may be taken more than once.
func passiveRules(cfg *config.Config) string {
	if cfg == nil || len(cfg.Passives.Entries) == 0 {
		return "_No passives configured._"
	}
	p := cfg.Passives
	var b strings.Builder
	b.WriteString("| Rule | Value |\n")
	b.WriteString("| --- | --- |\n")
	fmt.Fprintf(&b, "| Cost to take | %s, or whatever the entry states, plus whatever its configured parts add |\n", perkPointWords(p.BuildCost))
	fmt.Fprintf(&b, "| Cost to use | %s |\n", passiveEnergyWords(p.EnergyCost))
	fmt.Fprintf(&b, "| Taking one twice | %s |\n", passiveDuplicateWords(p.AllowsDuplicates()))
	fmt.Fprintf(&b, "| Budget | Perk points, the same pool every other perk is bought from. |\n")
	return strings.TrimSpace(b.String())
}

// passiveEnergyWords explains the per-use cost of a passive.
func passiveEnergyWords(energy int) string {
	if energy == 0 {
		return "Nothing. A passive is always on, so there is no moment at which energy is paid."
	}
	return fmt.Sprintf("%d energy", energy)
}

// passiveDuplicateWords explains the uniqueness rule.
func passiveDuplicateWords(allow bool) string {
	if allow {
		return "Permitted: the same passive may be taken more than once and its effects stack."
	}
	return "Not permitted. Each passive may be taken only once, so breadth is the only way to spend more points here."
}

// perkPointWords renders a perk-point cost in plain language.
func perkPointWords(points int) string {
	if points == 0 {
		return "Free"
	}
	return fmt.Sprintf("%d perk point%s", points, plural(points))
}
