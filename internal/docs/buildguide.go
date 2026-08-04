// This file implements the config-driven "build guide" renderer. Instead of the
// former Rules/Perks split, a chapter simply asks for buildGuide of a component
// and gets a self-contained, numbered walkthrough that a reader can follow to
// build the ability entirely by hand: every field, its range/default, and every
// concrete choice (traits, dice, damage types, etc.) is listed with its cost.
//
// The output deliberately avoids any application or configuration terminology.
// It reads as a standalone rulebook; the builder application is a convenience
// layer over the same definitions, not a prerequisite.
package docs

import (
	"fmt"
	"strings"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
)

// conditionsTable renders the full set of conditions defined in the config as
// reader-facing markdown. Shiftable conditions (those that move a collection of
// traits up or down within a shift range) are grouped into one table showing
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
		b.WriteString("These conditions raise or lower a collection of traits. ")
		b.WriteString("The value is a number of die shifts within the range shown; ")
		b.WriteString("which traits are affected is decided at the table.\n\n")
		b.WriteString("| Condition | Shift Range | Effect |\n")
		b.WriteString("| --- | --- | --- |\n")
		for _, c := range shiftable {
			fmt.Fprintf(&b, "| **%s** | %s | %s |\n",
				orDash(c.Name), shiftRange(c), orDash(c.Description))
		}
	}
	if len(fixed) > 0 {
		if len(shiftable) > 0 {
			b.WriteString("\n")
		}
		b.WriteString("### Fixed Conditions\n\n")
		b.WriteString("These conditions apply a set effect rather than a trait shift.\n\n")
		b.WriteString("| Condition | Effect |\n")
		b.WriteString("| --- | --- |\n")
		for _, c := range fixed {
			fmt.Fprintf(&b, "| **%s** | %s |\n",
				orDash(c.Name), orDash(c.Description))
		}
	}
	return strings.TrimSpace(b.String())
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

// traitsTable renders the trait roster and proficiency progression from the
// config. Dice-backed trait groups (everything except the vital group) each get
// a table listing the die every proficiency tier grants; the vital group gets a
// table listing the numeric HP/Movement/Energy each tier grants. This keeps the
// trait list and its tier progression in sync with the config definitions.
func traitsTable(cfg *config.Config) string {
	if cfg == nil || len(cfg.Traits.Order) == 0 {
		return "_No traits configured._"
	}
	vitalGroup := cfg.VitalGroup
	if vitalGroup == "" {
		vitalGroup = "vital"
	}
	tiers := cfg.Proficiencies
	var b strings.Builder
	for _, g := range cfg.Traits.List() {
		if g.ID == vitalGroup {
			writeVitalTable(&b, g, tiers)
			continue
		}
		writeDiceTraitTable(&b, g, tiers)
	}
	return strings.TrimSpace(b.String())
}

// writeDiceTraitTable writes a table for a dice-backed trait group: rows are
// traits, columns are proficiency tiers, cells are the die that tier grants.
func writeDiceTraitTable(b *strings.Builder, g config.TraitGroup, tiers []config.Proficiency) {
	fmt.Fprintf(b, "### %s Traits\n\n", g.Label)
	b.WriteString("| Trait |")
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
	for _, tr := range g.Traits {
		fmt.Fprintf(b, "| **%s** |", tr)
		for _, t := range tiers {
			fmt.Fprintf(b, " %s |", orDash(t.DieFor(g.ID)))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
}

// writeVitalTable writes the vital group's table: rows are vital traits,
// columns are proficiency tiers, cells are the numeric value that tier grants.
func writeVitalTable(b *strings.Builder, g config.TraitGroup, tiers []config.Proficiency) {
	fmt.Fprintf(b, "### %s Traits\n\n", g.Label)
	b.WriteString("These traits use numeric values rather than dice.\n\n")
	b.WriteString("| Trait |")
	for _, t := range tiers {
		fmt.Fprintf(b, " %s |", t.Name)
	}
	b.WriteString("\n| --- |")
	for range tiers {
		b.WriteString(" --- |")
	}
	b.WriteString("\n")
	for _, tr := range g.Traits {
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

// attributeSections renders the character attribute sheet sections from the
// config: one bulleted list per section, each listing its field labels.
func attributeSections(cfg *config.Config) string {
	if cfg == nil || len(cfg.Attributes.Order) == 0 {
		return "_No attribute sections configured._"
	}
	var b strings.Builder
	for _, sec := range cfg.Attributes.List() {
		fmt.Fprintf(&b, "**%s**\n\n", orDash(sec.Label))
		for _, f := range sec.Fields {
			fmt.Fprintf(&b, "*   %s\n", orDash(f.Label))
		}
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}

// buildGuide renders a component as a numbered "How to build it" walkthrough.

// It walks the component's fields in author order and describes each one in
// plain language, listing concrete choices for dropdowns and explicit ranges
// and starting values for numbers and repeatable lists.
func buildGuide(cfg *config.Config, comp *config.Component) string {
	if comp == nil {
		return ""
	}
	if guide := buildGuideFields(cfg, comp.Fields); guide != "" {
		return guide
	}
	return strings.TrimSpace(comp.Description)
}

// buildGuideFields renders a numbered "How to build it" walkthrough over an
// explicit field slice, used for sections such as validations that live outside
// a component. It returns "" when the slice yields no steps so callers can fall
// back to their own prose.
func buildGuideFields(cfg *config.Config, fields []config.Field) string {
	var b strings.Builder
	b.WriteString("**How to build it**\n\n")
	n := 0
	for _, f := range fields {
		// Conditional follow-up fields are described under their controlling
		// choice, so they are not given their own top-level step.
		if f.VisibilityWhen != "" {
			continue
		}
		n++
		writeGuideStep(&b, cfg, f, n)
	}
	if n == 0 {
		return ""
	}
	return strings.TrimSpace(b.String())
}

// writeGuideStep writes a single numbered step for a top-level field.
func writeGuideStep(b *strings.Builder, cfg *config.Config, f config.Field, n int) {
	switch f.Type {
	case "free_text":
		fmt.Fprintf(b, "%d. **%s** *(optional)* - %s No cost.\n", n, orDash(f.Label), textDetail(f))
	case "checkbox":
		fmt.Fprintf(b, "%d. **%s** - %s %s\n", n, orDash(f.Label), checkboxDetail(f), costClause(f.Cost, ""))
	case "free_number":
		fmt.Fprintf(b, "%d. **%s** - %s %s\n", n, orDash(f.Label), numberDetail(f), perStepClause(f))
	case "dropdown":
		fmt.Fprintf(b, "%d. **%s** - %s\n", n, orDash(f.Label), dropdownIntro(f))
		writeChoiceList(b, cfg, f, "   ")
		if c := flatCostClause(f.Cost); c != "" {
			fmt.Fprintf(b, "\n   %s\n", c)
		}
	case "multiselect", "conditions":
		fmt.Fprintf(b, "%d. **%s** - %s %s\n", n, orDash(f.Label), listIntro(cfg, f), perItemClause(f))
		writeRowFields(b, cfg, f, "   ")
	default:
		fmt.Fprintf(b, "%d. **%s** - %s\n", n, orDash(f.Label), orDash(f.Description))
	}
}

// textDetail returns the description for a free-text field, or a sensible
// default sentence when none is configured.
func textDetail(f config.Field) string {
	if strings.TrimSpace(f.Description) != "" {
		return ensureSentence(f.Description)
	}
	return "A note you can write on the ability."
}

// checkboxDetail describes a checkbox toggle.
func checkboxDetail(f config.Field) string {
	if strings.TrimSpace(f.Description) != "" {
		return ensureSentence(f.Description)
	}
	if strings.TrimSpace(f.Information) != "" {
		return ensureSentence(f.Information)
	}
	return ensureSentence("Enable " + f.Label)
}

// numberDetail describes a free_number field's range and starting value.
func numberDetail(f config.Field) string {
	var lead string
	if strings.TrimSpace(f.Description) != "" {
		lead = ensureSentence(f.Description) + " "
	} else if strings.TrimSpace(f.Information) != "" {
		lead = ensureSentence(f.Information) + " "
	}
	return fmt.Sprintf("%sAny whole number from **%d to %d** (starts at %s).", lead, f.Min, f.Max, defaultStr(f.Default))
}

// dropdownIntro returns the descriptive lead-in for a dropdown, ending with
// "Choose one of:" so the concrete option list follows.
func dropdownIntro(f config.Field) string {
	var lead string
	if strings.TrimSpace(f.Description) != "" {
		lead = ensureSentence(f.Description) + " "
	} else if strings.TrimSpace(f.Information) != "" {
		lead = ensureSentence(f.Information) + " "
	}
	return lead + "Choose one of:"
}

// listIntro describes a repeatable multiselect field: its starting count and
// any pre-filled entries.
func listIntro(cfg *config.Config, f config.Field) string {
	var lead string
	if strings.TrimSpace(f.Description) != "" {
		lead = ensureSentence(f.Description) + " "
	} else if strings.TrimSpace(f.Information) != "" {
		lead = ensureSentence(f.Information) + " "
	}
	count := f.DefaultCount
	countWord := numberWord(count)
	item := strings.ToLower(singular(f.Label))
	plural := item + "s"
	start := fmt.Sprintf("You start with **%s** %s", countWord, pluralize(count, item, plural))
	if len(f.RowDefaults) > 0 {
		names := prefilledNames(f.RowDefaults)
		if names != "" {
			start += " (" + names + ")"
		}
	}
	start += " and may add or remove " + plural + "."
	return lead + start
}

// prefilledNames renders the human-readable names of pre-filled list rows,
// stripping any group namespace prefix (e.g. "offense.Power" -> "Power").
func prefilledNames(rows []map[string]string) string {
	var names []string
	for _, r := range rows {
		v := r["value"]
		if v == "" {
			continue
		}
		if i := strings.IndexByte(v, '.'); i >= 0 {
			v = v[i+1:]
		}
		names = append(names, titleCaseWord(v))
	}
	switch len(names) {
	case 0:
		return ""
	case 1:
		return names[0]
	case 2:
		return names[0] + " and " + names[1]
	default:
		return strings.Join(names[:len(names)-1], ", ") + " and " + names[len(names)-1]
	}
}

// writeChoiceList prints the concrete options of a dropdown field, grouped by
// their optgroup label when the source is a grouped one. Per-option costs are
// appended when non-zero.
func writeChoiceList(b *strings.Builder, cfg *config.Config, f config.Field, indent string) {
	groups := cfg.ResolveOptionGroups(f)
	// A single unlabelled group renders as one inline bullet line.
	if len(groups) == 1 && groups[0].Label == "" {
		fmt.Fprintf(b, "%s- %s\n", indent, joinOptions(groups[0].Options))
		return
	}
	for _, g := range groups {
		if len(g.Options) == 0 {
			continue
		}
		fmt.Fprintf(b, "%s- *%s:* %s\n", indent, g.Label, joinOptions(g.Options))
	}
}

// writeRowFields describes the per-entry choices of a repeatable list field.
func writeRowFields(b *strings.Builder, cfg *config.Config, f config.Field, indent string) {
	// A multiselect may itself carry an options_source (each row is that
	// source) or define row_fields. Prefer row_fields when present.
	if len(f.RowFields) > 0 {
		for _, rf := range f.RowFields {
			if rf.Type == "dropdown" {
				fmt.Fprintf(b, "\n%sFor each %s, choose one of:\n", indent, strings.ToLower(singular(f.Label)))
				writeChoiceList(b, cfg, rf, indent)
			}
		}
		return
	}
	if f.OptionsSource != "" {
		fmt.Fprintf(b, "\n%sFor each %s, choose one of:\n", indent, strings.ToLower(singular(f.Label)))
		writeChoiceList(b, cfg, f, indent)
	}
}

// joinOptions renders a comma-separated list of option labels, each with its
// cost appended when non-zero.
func joinOptions(opts []config.Option) string {
	parts := make([]string, 0, len(opts))
	for _, o := range opts {
		label := optionLabel(o)
		if hasCost(o.Cost) {
			label += " (" + costWords(o.Cost) + ")"
		}
		parts = append(parts, label)
	}
	return strings.Join(parts, ", ")
}

// costClause renders a trailing cost sentence for a toggle-style field, e.g.
// "Cost: Free." or "Cost: 2 build.".
func costClause(c *config.Cost, _ string) string {
	return "Cost: " + costWords(c) + "."
}

// flatCostClause renders the cost sentence for a dropdown's flat field cost,
// or "" when there is none.
func flatCostClause(c *config.Cost) string {
	if !hasCost(c) {
		return ""
	}
	return "Cost: " + costWords(c) + "."
}

// perStepClause renders the per-step cost sentence for a free_number field.
func perStepClause(f config.Field) string {
	unit := stepUnit(f.Label)
	if f.PerStep != nil && (hasCost(f.PerStep.Increase) || hasCost(f.PerStep.Decrease)) {
		inc := f.PerStep.Increase
		return fmt.Sprintf("Cost: %s per %s.", costWords(inc), unit)
	}
	return fmt.Sprintf("Cost: Free per %s.", unit)
}

// perItemClause renders the per-entry cost sentence for a repeatable list.
func perItemClause(f config.Field) string {
	item := strings.ToLower(singular(f.Label))
	if f.PerItem != nil && (hasCost(f.PerItem.Increase) || hasCost(f.PerItem.Decrease)) {
		return fmt.Sprintf("Cost: %s per %s.", costWords(f.PerItem.Increase), item)
	}
	return fmt.Sprintf("Cost: Free per %s.", item)
}

// stepUnit picks a natural per-step noun from a number field's label.
func stepUnit(label string) string {
	l := strings.ToLower(label)
	switch {
	case strings.Contains(l, "duration") || strings.Contains(l, "round"):
		return "round"
	case strings.Contains(l, "distance") || strings.Contains(l, "range") || strings.Contains(l, "meter"):
		return "meter"
	case strings.Contains(l, "shift"):
		return "step"
	case strings.Contains(l, "bonus"):
		return "+1"
	default:
		return "step"
	}
}

// ensureSentence trims a string and appends a period when it lacks terminal
// punctuation, so descriptions read as complete sentences.
func ensureSentence(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	switch s[len(s)-1] {
	case '.', '!', '?', ':':
		return s
	}
	return s + "."
}

// numberWord spells out small counts (used for "start with two traits").
func numberWord(n int) string {
	words := []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten"}
	if n >= 0 && n < len(words) {
		return words[n]
	}
	return fmt.Sprintf("%d", n)
}

// pluralize picks the singular or plural noun for a count.
func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}

// defaultStr renders a field default value (which decodes as an arbitrary YAML
// scalar) as a plain string for prose.
func defaultStr(v any) string {
	if v == nil {
		return "0"
	}
	switch t := v.(type) {
	case string:
		if t == "" {
			return "0"
		}
		return t
	case int:
		return fmt.Sprintf("%d", t)
	case int64:
		return fmt.Sprintf("%d", t)
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%v", t)
	default:
		return fmt.Sprintf("%v", t)
	}
}
