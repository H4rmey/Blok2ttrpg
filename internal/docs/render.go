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
			return fieldsTable(comp.Fields)
		},
		// perksFields renders the same table for an explicit field slice, used
		// for sections like validations that live outside a component.
		"perksFields": func(fields []config.Field) string {
			return fieldsTable(fields)
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

// fieldsTable builds a markdown table describing the cost-bearing options of a
// set of fields. Fields without a direct cost (plain text/number without
// per-step, or dropdowns without option costs) are still listed so the reader
// sees the full option surface.
func fieldsTable(fields []config.Field) string {
	if len(fields) == 0 {
		return "_No options configured._"
	}
	var b strings.Builder
	b.WriteString("| Option | Choice | Build Cost | Energy Cost |\n")
	b.WriteString("| --- | --- | --- | --- |\n")
	rows := 0
	for _, f := range fields {
		switch f.Type {
		case "checkbox":
			b.WriteString(fmt.Sprintf("| %s | Enabled | %s |\n", orDash(f.Label), costCells(f.Cost)))
			rows++
		case "dropdown":
			if len(f.Options) > 0 {
				for _, opt := range f.Options {
					b.WriteString(fmt.Sprintf("| %s | %s | %s |\n", orDash(f.Label), orDash(optionLabel(opt)), costCells(opt.Cost)))
					rows++
				}
			} else {
				// options_source driven: no inline costs, but a flat field cost
				// may still apply.
				b.WriteString(fmt.Sprintf("| %s | Any | %s |\n", orDash(f.Label), costCells(f.Cost)))
				rows++
			}
		case "free_number":
			if f.PerStep != nil {
				if f.PerStep.Increase != nil {
					b.WriteString(fmt.Sprintf("| %s | Per step (increase) | %s |\n", orDash(f.Label), costCells(f.PerStep.Increase)))
					rows++
				}
				if f.PerStep.Decrease != nil {
					b.WriteString(fmt.Sprintf("| %s | Per step (decrease) | %s |\n", orDash(f.Label), costCells(f.PerStep.Decrease)))
					rows++
				}
			}
		}
	}
	if rows == 0 {
		return "_No cost-bearing options configured._"
	}
	return b.String()
}

func optionLabel(o config.Option) string {
	if o.Label != "" {
		return o.Label
	}
	return o.Value
}

// costCells returns the "build | energy" cells for a table row.
func costCells(c *config.Cost) string {
	if c == nil {
		return "0 | 0"
	}
	return fmt.Sprintf("%d | %d", c.BuildCost, c.EnergyCost)
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
