// Package docs renders the ruleset documentation. Unlike v4, it does not name
// any specific ability type or enactment: it simply passes the whole config to
// each markdown template and lets the template iterate. Docs therefore stay in
// sync with the config automatically.
package docs

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

// funcMap returns the template helpers used by the markdown docs. All helpers
// are config-driven: they look values up from the loaded ruleset so the docs
// stay in sync with the YAML automatically.
func funcMap(cfg *config.Config) template.FuncMap {
	return template.FuncMap{
		// Component lookups. Each returns a *Component (nil when missing) so
		// templates can chain field access without the two-value method form.
		"abilityType": func(id string) *config.Component {
			if c, ok := cfg.AbilityTypes.Get(id); ok {
				return c
			}
			return nil
		},
		"enactment": func(id string) *config.Component {
			if c, ok := cfg.Enactments.Get(id); ok {
				return c
			}
			return nil
		},
		"interaction": func(id string) *config.Component {
			if c, ok := cfg.Interactions.Get(id); ok {
				return c
			}
			return nil
		},
		// fieldDefault returns the configured default of a field by key, or an
		// empty string when the field/default is absent.
		"fieldDefault": func(comp *config.Component, key string) any {
			if comp == nil {
				return ""
			}
			if f, ok := findField(comp.Fields, key); ok && f.Default != nil {
				return f.Default
			}
			return ""
		},
		// perksTable renders a markdown table of every cost-bearing choice on a
		// component: checkboxes, dropdown options, and per-step number fields.
		"perksTable": func(comp *config.Component) string {
			if comp == nil {
				return "_No options configured._"
			}
			return fieldsTable(cfg, comp.Fields)
		},
		// perksFields renders the same table for an explicit field slice, used
		// for sections like validations that live outside a component.
		"perksFields": func(fields []config.Field) string {
			return fieldsTable(cfg, fields)
		},

		// costedOptionsTable renders a standalone reference table for a costed
		// option source (e.g. "trigger_events"), one row per option with its
		// cost in words. It is meant to be printed once and referenced from the
		// ability types that use the source, so the large list is not repeated.
		"costedOptionsTable": func(source string) string {
			opts := cfg.OptionsFor(source)
			if len(opts) == 0 {
				return "_No options configured._"
			}
			var b strings.Builder
			b.WriteString("| Choice | Cost |\n")
			b.WriteString("| --- | --- |\n")
			for _, opt := range opts {
				b.WriteString(fmt.Sprintf("| %s | %s |\n", orDash(optionLabel(opt)), costWords(opt.Cost)))
			}
			return b.String()
		},

		// enactmentSurchargeTable renders the additional-enactment surcharge.
		"enactmentSurchargeTable": func() string {
			var b strings.Builder
			b.WriteString("| Build Cost | Energy Cost | Description |\n")
			b.WriteString("| --- | --- | --- |\n")
			b.WriteString(fmt.Sprintf("| %d | %d | %s |\n",
				cfg.AdditionalEnactment.BuildCost,
				cfg.AdditionalEnactment.EnergyCost,
				orDash(cfg.AdditionalEnactment.Description)))
			return b.String()
		},
	}
}

// findField returns a field by key from a slice.
func findField(fields []config.Field, key string) (config.Field, bool) {
	for _, f := range fields {
		if f.Key == key {
			return f, true
		}
	}
	return config.Field{}, false
}

// sharedListSources names option sources that are large and reused across
// several components. Rather than repeat them inline in every Perks table,
// each is printed once in a shared reference table and skipped here.
var sharedListSources = map[string]bool{
	"trigger_events": true,
}

// fieldsTable builds a reader-friendly markdown table describing the perks a
// component offers. Each row is a plain-language description of a choice plus
// its cost in words (e.g. "4 build, 2 energy" or "Free"), so the docs read as a
// standalone rulebook rather than a dump of the builder's form fields. Purely
// cosmetic selectors that carry no cost are omitted (they belong in the Rules
// prose), and large shared option lists are referenced instead of repeated.
func fieldsTable(cfg *config.Config, fields []config.Field) string {
	if len(fields) == 0 {
		return "_No options configured._"
	}
	var b strings.Builder
	intro := "The following perks can be added when building or upgrading this component. " +
		"Build points are spent when the ability is created or upgraded; energy is paid each time it is used.\n\n"
	b.WriteString(intro)
	b.WriteString("| Perk | Cost |\n")
	b.WriteString("| --- | --- |\n")
	rows := writeFieldRows(&b, cfg, fields, false)
	if rows == 0 {
		return "_This component has no cost-bearing perks; see the Rules above._"
	}
	return b.String()
}

// writeFieldRows appends perk rows for a field slice to b and returns how many
// rows it wrote. When nested is true the fields are the row sub-fields of a
// solutions/states block, in which case conditional follow-ups and purely
// cosmetic selectors are skipped.
func writeFieldRows(b *strings.Builder, cfg *config.Config, fields []config.Field, nested bool) int {
	rows := 0
	seenSources := map[string]bool{}
	for _, f := range fields {
		// Conditional follow-up fields only appear in the builder after a
		// specific parent choice; they are documented via their parent, so
		// they never belong in a flat perks table.
		if f.VisibilityWhen != "" {
			continue
		}
		switch f.Type {
		case "checkbox":
			// A checkbox is only a perk when toggling it changes the cost.
			if hasCost(f.Cost) {
				b.WriteString(row("Enable: "+orDash(f.Label), costWords(f.Cost)))
				rows++
			}
		case "dropdown":
			rows += writeDropdownRows(b, cfg, f, nested, seenSources)
		case "free_number":
			if f.PerStep != nil {
				if f.PerStep.Increase != nil && hasCost(f.PerStep.Increase) {
					b.WriteString(row("Each +1 to "+orDash(f.Label), costWords(f.PerStep.Increase)))
					rows++
				}
				if f.PerStep.Decrease != nil && hasCost(f.PerStep.Decrease) {
					b.WriteString(row("Each -1 to "+orDash(f.Label), costWords(f.PerStep.Decrease)))
					rows++
				}
			}
		case "solutions", "states":
			single := singular(f.Label)
			if f.PerItem != nil {
				if f.PerItem.Increase != nil && hasCost(f.PerItem.Increase) {
					b.WriteString(row("Each additional "+single, costWords(f.PerItem.Increase)))
					rows++
				}
				if f.PerItem.Decrease != nil && hasCost(f.PerItem.Decrease) {
					b.WriteString(row("Each removed "+single, costWords(f.PerItem.Decrease)))
					rows++
				}
			}
			rows += writeFieldRows(b, cfg, f.RowFields, true)
		}
	}
	return rows
}

// writeDropdownRows renders the perk rows for a single dropdown field.
func writeDropdownRows(b *strings.Builder, cfg *config.Config, f config.Field, nested bool, seenSources map[string]bool) int {
	rows := 0
	// Inline options with their own costs: list them all (including the free
	// baseline) so the reader sees the full choice, each with its cost.
	if len(f.Options) > 0 {
		if !anyOptionCosted(f.Options) && !hasCost(f.Cost) {
			// A pure "pick one, no cost difference" selector: not a perk.
			return 0
		}
		for _, opt := range f.Options {
			b.WriteString(row(f.Label+": "+orDash(optionLabel(opt)), costWords(opt.Cost)))
			rows++
		}
		return rows
	}

	// options_source driven. Skip the big shared lists (documented once
	// elsewhere) but still surface the group offsets below.
	if !sharedListSources[f.OptionsSource] {
		if opts := resolveCostedOptions(cfg, f); len(opts) > 0 {
			if !seenSources[f.OptionsSource] {
				seenSources[f.OptionsSource] = true
				for _, opt := range opts {
					b.WriteString(row(f.Label+": "+orDash(optionLabel(opt)), costWords(opt.Cost)))
					rows++
				}
			}
		} else if !nested && hasCost(f.Cost) {
			// A flat field cost applies whenever this selector is used.
			b.WriteString(row(f.Label, costWords(f.Cost)))
			rows++
		}
	}

	// Group offsets attach a cost to picking a trait from a given group.
	if f.GroupOffsets != nil {
		for _, grp := range orderedGroups(f.GroupOffsets) {
			c := f.GroupOffsets.Offsets[grp]
			if !hasCost(c) {
				continue // the preferred group is free; not worth a row.
			}
			b.WriteString(row("Use "+groupPhrase(grp)+" trait for "+orDash(f.Label), costWords(c)))
			rows++
		}
	}
	return rows
}

// row formats one "| perk | cost |" markdown line.
func row(perk, cost string) string {
	return fmt.Sprintf("| %s | %s |\n", perk, cost)
}

// hasCost reports whether a cost is present and non-zero.
func hasCost(c *config.Cost) bool {
	return c != nil && (c.BuildCost != 0 || c.EnergyCost != 0)
}

// anyOptionCosted reports whether any option in the list carries a cost.
func anyOptionCosted(opts []config.Option) bool {
	for _, o := range opts {
		if hasCost(o.Cost) {
			return true
		}
	}
	return false
}

// costWords renders a cost in plain language, e.g. "Free", "2 build",
// "4 build, 2 energy", or "-1 build (refund)".
func costWords(c *config.Cost) string {
	if c == nil || (c.BuildCost == 0 && c.EnergyCost == 0) {
		return "Free"
	}
	var parts []string
	if c.BuildCost != 0 {
		s := fmt.Sprintf("%d build", c.BuildCost)
		if c.BuildCost < 0 {
			s += " (refund)"
		}
		parts = append(parts, s)
	}
	if c.EnergyCost != 0 {
		parts = append(parts, fmt.Sprintf("%d energy", c.EnergyCost))
	}
	return strings.Join(parts, ", ")
}

// groupPhrase turns a trait group id into a readable article + adjective, e.g.
// "offense" -> "an Offensive", "defense" -> "a Defensive".
func groupPhrase(group string) string {
	switch group {
	case "offense":
		return "an Offensive"
	case "defense":
		return "a Defensive"
	case "general":
		return "a General"
	default:
		return "a " + titleCaseWord(group)
	}
}

// singular strips a trailing plural "s" from a label for "Each additional X"
// phrasing. It is deliberately simple; irregular plurals are rare here.
func singular(label string) string {
	l := strings.TrimSpace(label)
	if strings.HasSuffix(l, "ies") {
		return l[:len(l)-3] + "y"
	}
	if strings.HasSuffix(l, "s") && !strings.HasSuffix(l, "ss") {
		return l[:len(l)-1]
	}
	return l
}

func titleCaseWord(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// resolveCostedOptions returns the resolved options for an options_source-driven
// field only when at least one option carries a non-zero cost. This keeps plain
// (uncosted) sources from rendering while costed sources such as reaction
// triggers expand into a row per option.
func resolveCostedOptions(cfg *config.Config, f config.Field) []config.Option {
	if cfg == nil || f.OptionsSource == "" {
		return nil
	}
	opts := cfg.OptionsFor(f.OptionsSource)
	for _, o := range opts {
		if o.Cost != nil && (o.Cost.BuildCost != 0 || o.Cost.EnergyCost != 0) {
			return opts
		}
	}
	return nil
}

// orderedGroups returns the group keys of a GroupOffsets in a stable order:
// the default group first (when set), then the remaining keys sorted.
func orderedGroups(g *config.GroupOffsets) []string {
	seen := map[string]bool{}
	var out []string
	if g.DefaultGroup != "" {
		if _, ok := g.Offsets[g.DefaultGroup]; ok {
			out = append(out, g.DefaultGroup)
			seen[g.DefaultGroup] = true
		}
	}
	var rest []string
	for k := range g.Offsets {
		if !seen[k] {
			rest = append(rest, k)
		}
	}
	sort.Strings(rest)
	return append(out, rest...)
}

func optionLabel(o config.Option) string {
	if o.Label != "" {
		return o.Label
	}
	return o.Value
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

// RenderMarkdown builds the full markdown documentation by rendering each file
// listed in cfg.Docs.Order (relative to dir) against the config as template
// data, then concatenating the results.
func RenderMarkdown(loaded *config.Loaded) (string, error) {
	if loaded == nil || loaded.Config == nil {
		return "", fmt.Errorf("config is nil")
	}
	order := loaded.FileOrder
	if len(order) == 0 {
		return "", fmt.Errorf("no file_order configured")
	}

	fns := funcMap(loaded.Config)

	var sections []string
	for _, rel := range order {
		// file_order entries are authored relative to the config directory.
		// Fall back to a path relative to the current working directory (the
		// project root) when the config-relative path does not exist, so docs
		// stored under ./docs/ resolve regardless of where the config lives.
		path := filepath.Join(loaded.Dir, filepath.FromSlash(rel))
		raw, err := os.ReadFile(path)
		if err != nil {
			alt := filepath.FromSlash(rel)
			if altRaw, altErr := os.ReadFile(alt); altErr == nil {
				raw, err = altRaw, nil
			}
		}
		if err != nil {
			return "", fmt.Errorf("reading doc %q: %w", rel, err)
		}

		tmpl, err := template.New(filepath.Base(path)).Funcs(fns).Parse(string(raw))
		if err != nil {
			return "", fmt.Errorf("parsing doc %q: %w", rel, err)
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, loaded.Config); err != nil {
			return "", fmt.Errorf("executing doc %q: %w", rel, err)
		}
		sections = append(sections, strings.TrimSpace(buf.String()))
	}
	return strings.Join(sections, "\n\n"), nil
}

// RenderHTML converts the markdown documentation to an HTML fragment.
func RenderHTML(loaded *config.Loaded) (string, error) {
	md, err := RenderMarkdown(loaded)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	gm := goldmark.New(goldmark.WithExtensions(extension.Table))
	if err := gm.Convert([]byte(md), &buf); err != nil {
		return "", fmt.Errorf("converting markdown: %w", err)
	}
	return buf.String(), nil
}
