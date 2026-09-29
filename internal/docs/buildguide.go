// This file renders the config-driven `build guide`: a numbered, self-contained
// walkthrough of a component that a reader can follow to build a perk entirely
// by hand. Every field, its range and default, and every concrete choice is
// listed with its cost.
//
// The output deliberately avoids application or configuration terminology. It
// reads as a standalone rulebook; the builder application is a convenience
// layer over the same definitions, not a prerequisite.
package docs

import (
	"fmt"
	"strings"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
)

// traitSections renders the character trait sheet sections from the
// config: one bulleted list per section, each listing its field labels.
func traitSections(cfg *config.Config) string {
	if cfg == nil || len(cfg.Traits.Order) == 0 {
		return "_No trait sections configured._"
	}
	var b strings.Builder
	for _, sec := range cfg.Traits.List() {
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
	body := buildGuideFields(cfg, comp.Fields)
	if body == "" {
		body = strings.TrimSpace(comp.Description)
	}
	// The allowed/blocked lists decide which enactments, interactions and
	// validations may legally be combined with this component. That is a rule a
	// reader needs in order to build an perk by hand, so it is appended to
	// the walkthrough rather than left implicit in the builder's dropdowns.
	if combos := allowedCombinations(cfg, comp); combos != "" {
		if body == "" {
			return combos
		}
		return body + "\n\n" + combos
	}
	return body
}

// allowedCombinations describes which other components may be used with this
// one, based on the component's allowed/blocked lists resolved through the same
// config lookups the builder uses. It returns "" when the component places no
// restrictions, so unrestricted components read no differently than before.
func allowedCombinations(cfg *config.Config, comp *config.Component) string {
	if cfg == nil || comp == nil {
		return ""
	}
	var lines []string
	// Perk types filter which enactments they accept.
	if len(comp.AllowedEnactments) > 0 || len(comp.BlockedEnactments) > 0 {
		if names := componentNames(cfg.EnactmentsFor(comp.ID)); names != "" {
			lines = append(lines, "*   Enactments: "+names)
		}
	}
	// Enactments filter which interactions and validations they accept.
	if len(comp.AllowedInteractions) > 0 || len(comp.BlockedInteractions) > 0 {
		if names := componentNames(cfg.InteractionsFor(comp.ID)); names != "" {
			lines = append(lines, "*   Interactions: "+names)
		}
	}
	if len(comp.AllowedValidations) > 0 || len(comp.BlockedValidations) > 0 {
		if names := fieldLabels(cfg.ValidationFieldsFor(comp.ID)); names != "" {
			lines = append(lines, "*   Validation rolls: "+names)
		}
	}
	if len(lines) == 0 {
		return ""
	}
	return "**Usable with**\n\n" + strings.Join(lines, "\n")
}

// componentNames renders the display names of a component list.
func componentNames(comps []*config.Component) string {
	names := make([]string, 0, len(comps))
	for _, c := range comps {
		if c == nil {
			continue
		}
		names = append(names, c.DisplayName())
	}
	return strings.Join(names, ", ")
}

// fieldLabels renders the labels of a field list.
func fieldLabels(fields []config.Field) string {
	names := make([]string, 0, len(fields))
	for _, f := range fields {
		if l := strings.TrimSpace(f.Label); l != "" {
			names = append(names, l)
		}
	}
	return strings.Join(names, ", ")
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
	return "A note you can write on the perk."
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
