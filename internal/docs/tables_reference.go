// This file renders the reference tables that come straight off the ruleset:
// the condition list, the skill roster with its dice and vital ladders, and the
// character trait sheet sections. They are generated so the rulebook cannot
// drift from the config it documents.
package docs

import (
	"fmt"
	"strings"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
)

// conditionsTable renders the full set of conditions defined in the config as
// reader-facing markdown. Shiftable conditions (those that move a collection of
// skills up or down within a shift range) are grouped into one table showing
// their range; every other condition is a fixed effect and is listed in a
// second table with its plain-language effect. The output keeps the rulebook's
// condition list automatically in sync with the config definitions.
func conditionsTable(cfg *config.Config) string {
	if cfg == nil || len(cfg.Conditions) == 0 {
		return "_No conditions configured._"
	}
	var shiftable, fixed []config.Condition
	for _, c := range cfg.Conditions {
		if c.Shiftable() {
			shiftable = append(shiftable, c)
		} else {
			fixed = append(fixed, c)
		}
	}
	var b strings.Builder
	if len(shiftable) > 0 {
		b.WriteString("### Shifting Conditions\n\n")
		b.WriteString("These conditions raise or lower a named set of skills. ")
		b.WriteString("Whoever applies one picks a number of die shifts from the range shown, ")
		b.WriteString("and every skill in the Affects column moves by that amount.\n\n")
		b.WriteString("| Condition | Shift Range | Affects | Effect |\n")
		b.WriteString("| --- | --- | --- | --- |\n")
		for _, c := range shiftable {
			fmt.Fprintf(&b, "| **%s** | %s | %s | %s |\n",
				orDash(c.Name), shiftRange(c), affectedSkills(c), orDash(c.Description))
		}
	}
	if len(fixed) > 0 {
		if len(shiftable) > 0 {
			b.WriteString("\n")
		}
		b.WriteString("### Fixed Conditions\n\n")
		b.WriteString("These conditions apply at a magnitude the rules already fixed, ")
		b.WriteString("so there is nothing to choose when one is applied. ")
		b.WriteString("A dash in the Shift column means the condition changes what a ")
		b.WriteString("character may do rather than any of their numbers.\n\n")
		b.WriteString("| Condition | Shift | Affects | Effect |\n")
		b.WriteString("| --- | --- | --- | --- |\n")
		for _, c := range fixed {
			shift := "-"
			if c.ShiftsSkills() {
				shift = signed(c.FixedShift)
			}
			fmt.Fprintf(&b, "| **%s** | %s | %s | %s |\n",
				orDash(c.Name), shift, affectedSkills(c), orDash(c.Description))
		}
	}
	b.WriteString("\n")
	b.WriteString("Shifts from several conditions on the same skill add together, and the ")
	b.WriteString("result stops at the ends of the proficiency ladder. Stacking two cheap ")
	b.WriteString("penalties is therefore a real tactic, but a skill that has already ")
	b.WriteString("bottomed out cannot be pushed any lower.\n")
	return strings.TrimSpace(b.String())
}

// affectedSkills renders a condition's affected-skill list for the rules table.
// A condition that names none changes no numbers, which the table states rather
// than leaving blank: an empty cell reads as missing data.
func affectedSkills(c config.Condition) string {
	if !c.ShiftsSkills() {
		return "_no skills_"
	}
	return strings.Join(c.AffectsSkills, ", ")
}

// shiftRange renders a shiftable condition's shift range as a readable string,
// e.g. "-6 to 0" or "+1 to +6".
func shiftRange(c config.Condition) string {
	return fmt.Sprintf("%s to %s", signed(c.MinShift), signed(c.MaxShift))
}

// signed renders an integer with an explicit sign for positive values.
func signed(n int) string {
	if n > 0 {
		return fmt.Sprintf("+%d", n)
	}
	return fmt.Sprintf("%d", n)
}

// skillsTable renders the skill roster and proficiency progression from the
// config. Dice-backed skill groups (everything except the vital group) each get
// a table listing the die every proficiency tier grants; the vital group gets a
// table listing the numeric HP/Movement/Energy each tier grants. This keeps the
// skill list and its tier progression in sync with the config definitions.
func skillsTable(cfg *config.Config) string {
	if cfg == nil || len(cfg.Skills.Order) == 0 {
		return "_No skills configured._"
	}
	vitalGroup := cfg.VitalGroup
	if vitalGroup == "" {
		vitalGroup = "vital"
	}
	tiers := cfg.Proficiencies
	var b strings.Builder
	for _, g := range cfg.Skills.List() {
		if g.ID == vitalGroup {
			writeVitalTable(&b, g, tiers)
			continue
		}
		writeDiceSkillTable(&b, g, tiers)
	}
	return strings.TrimSpace(b.String())
}

// writeDiceSkillTable writes a table for a dice-backed skill group: rows are
// skills, columns are proficiency tiers, cells are the die that tier grants.
func writeDiceSkillTable(b *strings.Builder, g config.SkillGroup, tiers []config.Proficiency) {
	fmt.Fprintf(b, "### %s Skills\n\n", g.Label)
	b.WriteString("| Skill |")
	for _, t := range tiers {
		fmt.Fprintf(b, " %s |", t.Name)
	}
	b.WriteString("\n| --- |")
	for range tiers {
		b.WriteString(" --- |")
	}
	b.WriteString("\n| *Cost* |")
	for _, t := range tiers {
		fmt.Fprintf(b, " %d |", t.Cost)
	}
	b.WriteString("\n")
	for _, tr := range g.Skills {
		fmt.Fprintf(b, "| **%s** |", tr)
		for _, t := range tiers {
			fmt.Fprintf(b, " %s |", orDash(t.DieFor(g.ID)))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
}

// writeVitalTable writes the vital group's table: rows are vital skills,
// columns are proficiency tiers, cells are the numeric value that tier grants.
func writeVitalTable(b *strings.Builder, g config.SkillGroup, tiers []config.Proficiency) {
	fmt.Fprintf(b, "### %s Skills\n\n", g.Label)
	b.WriteString("These skills use numeric values rather than dice.\n\n")
	b.WriteString("| Skill |")
	for _, t := range tiers {
		fmt.Fprintf(b, " %s |", t.Name)
	}
	b.WriteString("\n| --- |")
	for range tiers {
		b.WriteString(" --- |")
	}
	b.WriteString("\n")
	for _, tr := range g.Skills {
		key := strings.ToLower(tr)
		fmt.Fprintf(b, "| **%s** |", tr)
		for _, t := range tiers {
			fmt.Fprintf(b, " %s |", vitalValue(t, key))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
}

// vitalValue reads a numeric vital value for a tier by key, returning "-" when
// absent.
func vitalValue(t config.Proficiency, key string) string {
	if t.Vitals == nil {
		return "-"
	}
	v, ok := t.Vitals[key]
	if !ok {
		return "-"
	}
	return defaultStr(v)
}
