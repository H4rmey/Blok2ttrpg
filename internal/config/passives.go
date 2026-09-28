// Schema and accessors for passives: the predefined half of the perk system.
//
// Every other perk is assembled by the player out of enactments, interactions
// and validations, and the cost engine prices whatever they built. A passive is
// the opposite: it is written by the ruleset author, picked from a list, and
// costs a flat number of perk points plus whatever its own fields add. That is
// what lets a passive do things the builder cannot express at all, such as
// bending the dice ladder, changing a vital, or altering the invoke economy.
//
// A passive is still a perk. It is stored as a model.Perk of the "passive" perk
// type with its id in the fields, so the perk list, the budget, the printable
// sheet and the export all handle it without knowing it is special.
//
// An entry's configurable parts are ordinary Field values - the same type the
// perk builder uses - rather than a bespoke mechanism. That is deliberate: it
// means a passive gets free text, numbers, checkboxes, dropdowns and their
// entire cost model without this package implementing any of it, and a passive
// like "you are resistant to <text> and reduce that damage by <number>" needs no
// new code at all.
package config

import "strconv"

// Passives is the predefined perk catalogue plus the rules that apply to the
// whole category.
type Passives struct {
	// Information is optional section-level help text, rendered into the
	// rulebook and shown above the passive picker.
	Information string `yaml:"information,omitempty" json:"information,omitempty"`

	// BuildCost is the default perk-point cost of a passive. An entry may
	// override it with its own build_cost, and its fields add on top.
	BuildCost int `yaml:"build_cost,omitempty" json:"build_cost,omitempty"`

	// EnergyCost is what a passive costs to use. It is normally zero: a passive
	// is always on, so there is no moment at which energy would be paid. It is
	// configurable anyway so a ruleset can charge for one if it wants to.
	EnergyCost int `yaml:"energy_cost,omitempty" json:"energy_cost,omitempty"`

	// AllowDuplicates permits the same passive to be taken more than once. It is
	// a pointer so an unset value defaults to false, which is the rule that
	// keeps a character from stacking one strong passive instead of broadening.
	AllowDuplicates *bool `yaml:"allow_duplicates,omitempty" json:"allow_duplicates,omitempty"`

	// Entries is the catalogue itself, in the order it is presented.
	Entries []Passive `yaml:"entries,omitempty" json:"entries,omitempty"`
}

// AllowsDuplicates reports whether the same passive may be selected more than
// once. Defaults to false when unset.
func (p Passives) AllowsDuplicates() bool {
	return p.AllowDuplicates != nil && *p.AllowDuplicates
}

// Passive is one predefined perk.
type Passive struct {
	// ID is the stable identifier stored on the character. Renaming it breaks
	// saved characters; renaming Name does not.
	ID   string `yaml:"id" json:"id"`
	Name string `yaml:"name" json:"name"`

	// Description is the rules text. Any "{key}" in it is replaced by the value
	// of the field with that key, so the text a player reads always states what
	// they actually configured. A key may appear more than once.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// BuildCost overrides the category default for this entry. Zero means "use
	// the default", so an entry only needs it when it is priced differently.
	// Field costs are added on top of whichever applies.
	BuildCost int `yaml:"build_cost,omitempty" json:"build_cost,omitempty"`

	// Fields are the configurable parts of the passive, using the same Field
	// type (and therefore the same renderer and cost engine) as the perk
	// builder. An entry with no fields is a fixed passive: there is nothing to
	// configure and no configure step when taking it.
	Fields []Field `yaml:"fields,omitempty" json:"fields,omitempty"`
}

// Configurable reports whether the passive has anything to configure. A fixed
// passive is added straight to the character without a configure step.
func (p Passive) Configurable() bool { return len(p.Fields) > 0 }

// PassiveByID returns a passive entry by id.
func (c *Config) PassiveByID(id string) (Passive, bool) {
	for _, p := range c.Passives.Entries {
		if p.ID == id {
			return p, true
		}
	}
	return Passive{}, false
}

// PassiveFlatCost returns the entry's own build cost before its fields are
// priced: its build_cost when set, otherwise the category default. The field
// costs are added by the engine, which owns the generic field cost model.
func (c *Config) PassiveFlatCost(p Passive) int {
	if p.BuildCost != 0 {
		return p.BuildCost
	}
	return c.Passives.BuildCost
}

// PassiveEnergyCost returns what using a passive costs in energy. It is a
// category-wide value because a passive is always on: there is no per-entry
// moment at which a different amount would be paid.
func (c *Config) PassiveEnergyCost() int {
	return c.Passives.EnergyCost
}

// PassiveDefaults returns the starting value of every field on a passive, which
// is what it is taken at before any configuration.
func (c *Config) PassiveDefaults(p Passive) map[string]any {
	out := make(map[string]any, len(p.Fields))
	for _, f := range p.Fields {
		switch f.Type {
		case "checkbox":
			out[f.Key] = asBoolValue(f.Default)
		case "free_number":
			out[f.Key] = ClampFieldNumber(f, asIntValue(f.Default))
		default:
			v := fieldDefaultString(f)
			if v == "" && f.Type == "dropdown" {
				// A dropdown must resolve to a real option, matching what the
				// builder renders and what normalization stores.
				for _, opt := range c.ResolveOptions(f) {
					if opt.Value != "" {
						v = opt.Value
						break
					}
				}
			}
			out[f.Key] = v
		}
	}
	return out
}

// DescriptionSegment is one piece of a passive's rendered rules text: a run of
// literal text followed by the value of a field.
//
// Splitting the text this way lets a template highlight each configured value
// where it sits in the sentence without this package emitting HTML. The final
// segment carries only Text, with an empty Key.
type DescriptionSegment struct {
	// Text is the literal run before the value.
	Text string
	// Key names the field whose value follows, or "" on the trailing segment.
	Key string
	// Value is that field's current value, rendered for display.
	Value string
}

// PassiveDescriptionSegments splits a passive's rules text at every "{key}"
// placeholder, pairing each literal run with the value that follows it.
//
// Values come from the supplied map, falling back to the field's default when a
// key is absent, so a partially configured passive still reads correctly. A
// placeholder naming a field the entry does not declare is left verbatim: that
// is an authoring error, and silently blanking it would hide it.
func (c *Config) PassiveDescriptionSegments(p Passive, values map[string]any) []DescriptionSegment {
	if len(p.Fields) == 0 {
		return []DescriptionSegment{{Text: p.Description}}
	}
	resolved := c.PassiveDefaults(p)
	for k, v := range values {
		if _, declared := resolved[k]; declared {
			resolved[k] = v
		}
	}

	var out []DescriptionSegment
	rest := p.Description
	// Text accumulated since the last real placeholder. An unrecognised
	// placeholder is folded into this run verbatim rather than dropped.
	pending := ""
	for {
		open := indexOf(rest, "{")
		if open < 0 {
			break
		}
		closeAt := indexOfFrom(rest, "}", open)
		if closeAt < 0 {
			break
		}
		key := rest[open+1 : closeAt]
		if _, declared := resolved[key]; !declared {
			// Not one of this entry's fields: keep the braces in the text so the
			// authoring mistake is visible instead of silently vanishing.
			pending += rest[:closeAt+1]
			rest = rest[closeAt+1:]
			continue
		}
		out = append(out, DescriptionSegment{
			Text:  pending + rest[:open],
			Key:   key,
			Value: displayValue(resolved[key]),
		})
		pending = ""
		rest = rest[closeAt+1:]
	}
	return append(out, DescriptionSegment{Text: pending + rest})
}

// PassiveDescription returns a passive's rules text with every placeholder
// replaced, as a flat string. It is the plain-text counterpart of
// PassiveDescriptionSegments and is what gets stored on the perk.
func (c *Config) PassiveDescription(p Passive, values map[string]any) string {
	out := ""
	for _, seg := range c.PassiveDescriptionSegments(p, values) {
		out += seg.Text + seg.Value
	}
	return out
}

// ClampFieldNumber constrains a free_number value to its field's declared range.
// It is exported because the passive configure path needs the same clamping the
// builder's normalization applies, so a hand-made post cannot buy a value the
// ruleset does not offer.
func ClampFieldNumber(f Field, n int) int {
	if f.Min != 0 || f.Max != 0 {
		if n < f.Min {
			return f.Min
		}
		if f.Max != 0 && n > f.Max {
			return f.Max
		}
	}
	return n
}

// displayValue renders a stored field value for the rules text. A checkbox reads
// as yes/no because a bare "true" in a sentence is jarring.
func displayValue(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case bool:
		if t {
			return "yes"
		}
		return "no"
	case int:
		return strconv.Itoa(t)
	case float64:
		return strconv.Itoa(int(t))
	default:
		return ""
	}
}

// fieldDefaultString renders a field's configured default as a string.
func fieldDefaultString(f Field) string {
	if f.Default == nil {
		return ""
	}
	if s, ok := f.Default.(string); ok {
		return s
	}
	return displayValue(f.Default)
}

func asIntValue(v any) int {
	switch t := v.(type) {
	case int:
		return t
	case float64:
		return int(t)
	case string:
		n, err := strconv.Atoi(t)
		if err != nil {
			return 0
		}
		return n
	}
	return 0
}

func asBoolValue(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t == "true" || t == "on"
	}
	return false
}

// indexOf returns the first index of sub in s, or -1.
func indexOf(s, sub string) int {
	return indexOfFrom(s, sub, 0)
}

// indexOfFrom returns the first index of sub in s at or after start, or -1.
func indexOfFrom(s, sub string, start int) int {
	if start < 0 || len(sub) > len(s) {
		return -1
	}
	for i := start; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// passivesConfigured reports whether a section file actually declared a passives
// block. It cannot be a struct comparison because Passives holds a slice and a
// pointer, neither of which is comparable.
func passivesConfigured(p Passives) bool {
	return p.Information != "" ||
		p.BuildCost != 0 ||
		p.EnergyCost != 0 ||
		p.AllowDuplicates != nil ||
		len(p.Entries) > 0
}
