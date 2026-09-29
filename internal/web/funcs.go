package web

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/engine"
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
		"skillProf": func(c model.Character, group, skill string) string {
			if c.Skills == nil {
				return ""
			}
			return c.Skills[model.SkillKey(group, skill)]
		},
		// skillProfLabel renders a character's stored proficiency for a skill
		// as a readable label including its die or vital value, e.g.
		// "Trained (d8)". Printed sheets need the die, not the raw tier id.
		// Falls back to the raw stored id when the tier cannot be resolved.
		"skillProfLabel": func(cfg *config.Config, c model.Character, group, skill string) string {
			if c.Skills == nil {
				return ""
			}
			id := c.Skills[model.SkillKey(group, skill)]
			if id == "" || cfg == nil {
				return id
			}
			p, ok := cfg.Proficiency(id)
			if !ok {
				return id
			}
			return profSkillLabelStr(cfg, p, group, skill)
		},
		// proficiencies exposes the ordered proficiency ladder so printable
		// templates can render one tick box per tier.
		"proficiencies": func(cfg *config.Config) []config.Proficiency {
			if cfg == nil {
				return nil
			}
			return cfg.Proficiencies
		},
		// componentName resolves a component id to its display name, falling
		// back to the id so nothing renders blank on a printed sheet.
		"componentName": func(cfg *config.Config, kind, id string) string {
			if cfg == nil || id == "" {
				return id
			}
			if comp, ok := cfg.ComponentByKind(kind, id); ok && comp.Name != "" {
				return comp.Name
			}
			return id
		},
		"resolveOptions": func(cfg *config.Config, f config.Field) []config.Option {
			return cfg.ResolveOptions(f)
		},
		"resolveOptionGroups": func(cfg *config.Config, f config.Field) []config.OptionGroup {
			return cfg.ResolveOptionGroups(f)
		},
		// firstOption returns the first selectable value of a dropdown. Every
		// dropdown must resolve to a real option (there is no empty "none"
		// choice), so this is the fallback when a field carries neither a
		// stored value nor a configured default.
		//
		// This delegates to the engine rather than reimplementing the rule. It
		// used to be a second copy of the same loop, which meant the form could
		// in principle preselect one option while normalization stored another,
		// and the perk would then be priced as something the user never saw.
		// Sharing one implementation makes that disagreement impossible.
		"firstOption": engine.FirstOptionValue,

		// componentByKind resolves a component (enactment/interaction/perk
		// type) by kind and id for the inline builder. Returns nil when not
		// found so the template can guard with `if`.
		"componentByKind": func(cfg *config.Config, kind, id string) *config.Component {
			if comp, ok := cfg.ComponentByKind(kind, id); ok {
				return &comp
			}
			return nil
		},

		// perkTypes/enactments/interactions expose the ordered component
		// lists so templates can range over them.
		"perkTypes":    func(cfg *config.Config) []*config.Component { return cfg.PerkTypes.List() },
		"enactments":   func(cfg *config.Config) []*config.Component { return cfg.Enactments.List() },
		"interactions": func(cfg *config.Config) []*config.Component { return cfg.Interactions.List() },
		"traits":       func(cfg *config.Config) []*config.TraitGroup { return cfg.Traits.List() },
		"skillGroups":  func(cfg *config.Config) []config.SkillGroup { return cfg.Skills.List() },
		// enactmentsFor/interactionsFor/validationFieldsFor apply the UI-only
		// allowed/blocked filtering for a given perk-type or enactment id.
		"enactmentsFor": func(cfg *config.Config, perkTypeID string) []*config.Component {
			return cfg.EnactmentsFor(perkTypeID)
		},
		"interactionsFor": func(cfg *config.Config, enactmentID string) []*config.Component {
			return cfg.InteractionsFor(enactmentID)
		},
		"validationFieldsFor": func(cfg *config.Config, enactmentID string) []config.Field {
			return cfg.ValidationFieldsFor(enactmentID)
		},
		// validationFields exposes the engagement/counter (validation) fields so
		// each enactment can render its own validation region.
		"validationFields": func(cfg *config.Config) []config.Field { return cfg.Validations.Fields },

		// showInteraction/showValidation report whether the Interaction /
		// Validation region should be shown for the enactment at the given
		// index. The first enactment (index 0) always shows both; enactments
		// beyond the first consult additional_enactment.require_interaction /
		// require_validation (default true when unset).
		"showInteraction": func(cfg *config.Config, index int) bool {
			return index == 0 || cfg.AdditionalEnactment.RequiresInteraction()
		},
		"showValidation": func(cfg *config.Config, index int) bool {
			return index == 0 || cfg.AdditionalEnactment.RequiresValidation()
		},

		// hasValues reports whether a stored values map holds anything
		// meaningful. An optional builder region (Interaction / Validation
		// beyond the first enactment) is rendered when the ruleset requires it
		// OR when the perk already carries data there, so opening and saving a
		// perk neither discards that data nor invents defaults the perk never
		// had. Both would change the perk's cost behind the user's back.
		"hasValues": func(values map[string]any) bool {
			for _, v := range values {
				switch t := v.(type) {
				case nil:
				case string:
					if t != "" {
						return true
					}
				default:
					return true
				}
			}
			return false
		},

		// costHint formats a flat cost into a short inline hint such as
		// "(-2 pt, +1 E)". Zero components are omitted; an all-zero cost yields
		// an empty string so nothing is shown.
		"costHint": func(c *config.Cost) string { return costHintStr(c) },
		// costHintVal is costHint for a Cost held by value (e.g. a component's
		// BaseCost), which a template cannot take the address of.
		"costHintVal": func(c config.Cost) string { return costHintStr(&c) },
		// newTargetHint formats the configured cost of giving an enactment its
		// own target, for the builder checkbox label.
		"newTargetHint": func(cfg *config.Config) string {
			c := cfg.AdditionalEnactment.NewTarget.AsCost()
			return costHintStr(&c)
		},

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
		// given skill group, e.g. "Trained (d8)".
		"profLabel": func(p config.Proficiency, groupID string) string {
			if d := p.DieFor(groupID); d != "" {
				return fmt.Sprintf("%s (%s)", p.Name, d)
			}
			// Vital skills show hp/movement/energy rather than dice.
			if v, ok := p.Vitals[groupID]; ok {
				return fmt.Sprintf("%s (%v)", p.Name, v)
			}
			return p.Name
		},
		// profSkillLabel renders a proficiency choice for a specific skill. For
		// vital skills it shows the numeric vital value (keyed by skill name)
		// rather than a die; otherwise it falls back to the dice-based label.
		// The configured vital group id (cfg.VitalGroup) selects which skill
		// group is treated as vitals.
		"profSkillLabel": func(cfg *config.Config, p config.Proficiency, groupID, skill string) string {
			return profSkillLabelStr(cfg, p, groupID, skill)
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
		// atoi parses a string index into an int (0 on failure), so the
		// enactment partial can turn its string ".Index" into a number for the
		// showInteraction/showValidation guards.
		"atoi": func(s string) int {
			n, _ := strconv.Atoi(s)
			return n
		},

		// rowDefault returns the pre-fill value for a given row index and
		// row-field key, from a field's row_defaults config. Empty when none.
		"rowDefault": func(f config.Field, row int, key string) string {
			if row < 0 || row >= len(f.RowDefaults) {
				return ""
			}
			return f.RowDefaults[row][key]
		},

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

		// fvalStr returns the stored value for key from a values map rendered as
		// a string, falling back to def when there is no stored value (or it is
		// empty). Used by the builder to re-populate fields on edit/import.
		"fvalStr": func(values map[string]any, key, def string) string {
			if values != nil {
				if v, ok := values[key]; ok && v != nil {
					if s := fmt.Sprintf("%v", v); s != "" {
						return s
					}
				}
			}
			return def
		},
		// fvalBool returns the stored bool value for key, falling back to def.
		// Accepts native bools and the string forms "true"/"on".
		"fvalBool": func(values map[string]any, key string, def bool) bool {
			if values != nil {
				if v, ok := values[key]; ok {
					switch t := v.(type) {
					case bool:
						return t
					case string:
						return t == "true" || t == "on"
					}
				}
			}
			return def
		},
		// fvalMap returns a nested values map stored under key (used by the
		// inline_builder to re-populate its nested component fields).
		"fvalMap": func(values map[string]any, key string) map[string]any {
			return asValuesMap(mapGet(values, key))
		},
		// resolveRows returns the row value maps a solutions/states field should
		// render: the stored rows when present, otherwise the configured
		// defaults (row_defaults / field defaults up to default_count).
		"resolveRows": func(f config.Field, values map[string]any) []map[string]any {
			return resolveRows(f, values)
		},

		// skillView looks up the post-condition reading of one skill from the
		// map the page envelope carries. A skill no condition touches still has
		// an entry, so the template never has to branch on presence.
		"skillView": func(views map[string]engine.SkillView, group, skill string) engine.SkillView {
			if views == nil {
				return engine.SkillView{}
			}
			return views[model.SkillKey(group, skill)]
		},
		// signedInt renders a shift with an explicit sign, so a badge reads
		// "-2" / "+2" rather than an ambiguous bare number.
		"signedInt": func(n int) string {
			if n > 0 {
				return fmt.Sprintf("+%d", n)
			}
			return strconv.Itoa(n)
		},
		// joinList renders a string slice as a comma-separated sentence, used by
		// the condition tooltip to list the skills it affects.
		"joinList": func(items []string) string {
			return strings.Join(items, ", ")
		},
	}
}

// profSkillLabelStr renders a proficiency tier for a specific skill. Vital
// skills show their numeric vital value (keyed by the lowercased skill name);
// every other group shows the die the tier grants. Shared by the
// profSkillLabel and skillProfLabel template helpers.
func profSkillLabelStr(cfg *config.Config, p config.Proficiency, groupID, skill string) string {
	vitalGroup := "vital"
	if cfg != nil && cfg.VitalGroup != "" {
		vitalGroup = cfg.VitalGroup
	}
	if groupID == vitalGroup {
		key := strings.ToLower(skill)
		if v, ok := p.Vitals[key]; ok {
			return fmt.Sprintf("%s (%v)", p.Name, v)
		}
		return p.Name
	}
	if d := p.DieFor(groupID); d != "" {
		return fmt.Sprintf("%s (%s)", p.Name, d)
	}
	return p.Name
}

// mapGet safely reads a key from a values map, returning nil when absent.
func mapGet(values map[string]any, key string) any {
	if values == nil {
		return nil
	}
	return values[key]
}

// asValuesMap coerces a stored value into a map[string]any. It accepts the
// native map[string]any as well as YAML's map[string]interface{} and
// map[interface{}]interface{} shapes.
func asValuesMap(v any) map[string]any {
	switch m := v.(type) {
	case map[string]any:
		return m
	case map[interface{}]interface{}:
		out := map[string]any{}
		for k, val := range m {
			out[fmt.Sprintf("%v", k)] = val
		}
		return out
	}
	return nil
}

// resolveRows produces the row value maps to render for a repeatable field.
// Stored rows (from a saved/imported perk) take precedence; otherwise the
// configured row defaults are used, one map per default row.
func resolveRows(f config.Field, values map[string]any) []map[string]any {
	if raw := mapGet(values, f.Key); raw != nil {
		if rows := normalizeRows(raw); len(rows) > 0 {
			return rows
		}
	}
	out := []map[string]any{}
	for i := 0; i < f.DefaultCount; i++ {
		row := map[string]any{}
		for _, rf := range f.RowFields {
			val := ""
			if i < len(f.RowDefaults) {
				val = f.RowDefaults[i][rf.Key]
			}
			if val == "" && rf.Default != nil {
				val = fmt.Sprintf("%v", rf.Default)
			}
			row[rf.Key] = val
		}
		out = append(out, row)
	}
	return out
}

// normalizeRows coerces a stored rows value into []map[string]any. It accepts
// the native []map[string]any as well as the []interface{} of maps that YAML
// unmarshalling produces.
func normalizeRows(v any) []map[string]any {
	switch rows := v.(type) {
	case []map[string]any:
		return rows
	case []interface{}:
		out := make([]map[string]any, 0, len(rows))
		for _, r := range rows {
			if m := asValuesMap(r); m != nil {
				out = append(out, m)
			}
		}
		return out
	}
	return nil
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
