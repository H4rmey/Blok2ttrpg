package web

import (
	"fmt"
	"html/template"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// funcMap exposes helpers to templates so views can stay declarative and let
// the config drive what is rendered.
func funcMap() template.FuncMap {
	return template.FuncMap{
		"attr": func(c model.Character, key string) any {
			return c.Attr(key)
		},
		"attrStr": func(c model.Character, key string) string {
			if v := c.Attr(key); v != nil {
				if s, ok := v.(string); ok {
					return s
				}
			}
			return ""
		},
		"traitProf": func(c model.Character, group, trait string) string {
			if c.Traits == nil {
				return ""
			}
			return c.Traits[model.TraitKey(group, trait)]
		},
		"resolveOptions": func(cfg *config.Config, f config.Field) []config.Option {
			return cfg.ResolveOptions(f)
		},
		"resolveOptionGroups": func(cfg *config.Config, f config.Field) []config.OptionGroup {
			return cfg.ResolveOptionGroups(f)
		},

		// abilityTypes/enactments/interactions expose the ordered component
		// lists so templates can range over them.
		"abilityTypes": func(cfg *config.Config) []*config.Component { return cfg.AbilityTypes.List() },
		"enactments":   func(cfg *config.Config) []*config.Component { return cfg.Enactments.List() },
		"interactions": func(cfg *config.Config) []*config.Component { return cfg.Interactions.List() },
		"attributes":   func(cfg *config.Config) []*config.AttributeGroup { return cfg.Attributes.List() },
		"traitGroups":  func(cfg *config.Config) []config.TraitGroup { return cfg.Traits.List() },
		// validationFields exposes the engagement/counter (validation) fields so
		// each enactment can render its own validation region.
		"validationFields": func(cfg *config.Config) []config.Field { return cfg.Validations.Fields },
		// costHint formats a flat cost into a short inline hint such as
		// "(-2 pt, +1 E)". Zero components are omitted; an all-zero cost yields
		// an empty string so nothing is shown.
		"costHint": func(c *config.Cost) string { return costHintStr(c) },
		// perStepHint formats a free_number per-step cost into a hint describing
		// the increase (and decrease, if different) per step.
		"perStepHint": func(p *config.PerStep) string { return perStepHintStr(p) },

		// numberRange returns the discrete values a free_number field may take,
		// so the builder can render it as a dropdown rather than a spinner.
		"numberRange": func(f config.Field) []int {
			step := f.Step
			if step <= 0 {
				step = 1
			}
			var out []int
			for v := f.Min; v <= f.Max; v += step {
				out = append(out, v)
			}
			if len(out) == 0 {
				out = append(out, f.Min)
			}
			return out
		},
		// profLabel renders a proficiency choice with its dice value for the
		// given trait group, e.g. "Trained (d8)".
		"profLabel": func(p config.Proficiency, groupID string) string {
			if d, ok := p.Dice[groupID]; ok && d != "" {
				return fmt.Sprintf("%s (%s)", p.Name, d)
			}
			// Vital traits show hp/movement/energy rather than dice.
			if v, ok := p.Vitals[groupID]; ok {
				return fmt.Sprintf("%s (%v)", p.Name, v)
			}
			return p.Name
		},
		// dict builds a map from alternating key/value pairs, for passing
		// structured data into sub-templates.
		"dict": func(kv ...any) map[string]any {
			m := map[string]any{}
			for i := 0; i+1 < len(kv); i += 2 {
				if k, ok := kv[i].(string); ok {
					m[k] = kv[i+1]
				}
			}
			return m
		},
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		// str renders any value as a string ("" for nil), used to compare a
		// field's default against dropdown option values in templates.
		"str": func(v any) string {
			if v == nil {
				return ""
			}
			return fmt.Sprintf("%v", v)
		},
		// seq returns []int{0,1,...,n-1} so templates can render a fixed number
		// of default rows for solutions/states fields.
		"seq": func(n int) []int {
			if n < 0 {
				n = 0
			}
			out := make([]int, n)
			for i := range out {
				out[i] = i
			}
			return out
		},
	}
}

// costHintStr formats a flat cost into a short inline hint like "(-2 pt, +1 E)".
// Zero components are dropped; an entirely zero (or nil) cost yields "".
func costHintStr(c *config.Cost) string {
	if c == nil {
		return ""
	}
	parts := costParts(c.BuildCost, c.EnergyCost)
	if parts == "" {
		return ""
	}
	return "(" + parts + ")"
}

// perStepHintStr formats a free_number per-step cost. It shows the increase
// per step, and the decrease too when it differs, e.g. "(+2 pt, +1 E / step)".
func perStepHintStr(p *config.PerStep) string {
	if p == nil {
		return ""
	}
	var inc, dec string
	if p.Increase != nil {
		inc = costParts(p.Increase.BuildCost, p.Increase.EnergyCost)
	}
	if p.Decrease != nil {
		dec = costParts(p.Decrease.BuildCost, p.Decrease.EnergyCost)
	}
	switch {
	case inc != "" && dec != "":
		return fmt.Sprintf("(+step: %s / -step: %s)", inc, dec)
	case inc != "":
		return fmt.Sprintf("(%s / step)", inc)
	case dec != "":
		return fmt.Sprintf("(-step: %s)", dec)
	default:
		return ""
	}
}

// costParts renders the non-zero build/energy components with explicit signs,
// e.g. "-2 pt, +1 E". Returns "" when both are zero.
func costParts(build, energy int) string {
	var parts []string
	if build != 0 {
		parts = append(parts, fmt.Sprintf("%+d pt", build))
	}
	if energy != 0 {
		parts = append(parts, fmt.Sprintf("%+d E", energy))
	}
	return joinComma(parts)
}

func joinComma(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += ", "
		}
		out += p
	}
	return out
}
