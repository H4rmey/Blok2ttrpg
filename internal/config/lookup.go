package config

// AbilityType returns the ability-type component with the given id.
func (c *Config) AbilityType(id string) (Component, bool) {
	if comp, ok := c.AbilityTypes.Get(id); ok {
		return *comp, true
	}
	return Component{}, false
}

// Enactment returns the enactment component with the given id.
func (c *Config) Enactment(id string) (Component, bool) {
	if comp, ok := c.Enactments.Get(id); ok {
		return *comp, true
	}
	return Component{}, false
}

// Interaction returns the interaction component with the given id.
func (c *Config) Interaction(id string) (Component, bool) {
	if comp, ok := c.Interactions.Get(id); ok {
		return *comp, true
	}
	return Component{}, false
}

// ComponentByKind resolves a component id against the map named by kind. It
// backs the generic inline_builder feature so a dropdown can reference any
// enactment, interaction or ability type.
func (c *Config) ComponentByKind(kind, id string) (Component, bool) {
	switch kind {
	case "enactment":
		return c.Enactment(id)
	case "interaction":
		return c.Interaction(id)
	case "ability_type":
		return c.AbilityType(id)
	default:
		return Component{}, false
	}
}

// GeneralStateByID returns the general state with the given id.
func (c *Config) GeneralStateByID(id string) (GeneralState, bool) {
	for _, s := range c.GeneralStates {
		if s.ID == id {
			return s, true
		}
	}
	return GeneralState{}, false
}

// SpecificStateByID returns the specific state with the given id.
func (c *Config) SpecificStateByID(id string) (SpecificState, bool) {
	for _, s := range c.SpecificStates {
		if s.ID == id {
			return s, true
		}
	}
	return SpecificState{}, false
}

// ShiftOptionsFor returns the discrete non-zero shift values a general state
// may take, from its configured min_shift..max_shift range. Zero is skipped
// because applying a state with no shift is meaningless.
func (c *Config) ShiftOptionsFor(generalID string) []int {
	s, ok := c.GeneralStateByID(generalID)
	if !ok {
		return nil
	}
	min, max := s.MinShift, s.MaxShift
	if min == 0 && max == 0 {
		min, max = -6, 6
	}
	var out []int
	for v := min; v <= max; v++ {
		if v != 0 {
			out = append(out, v)
		}
	}
	return out
}

// Proficiency returns the proficiency tier with the given id.
func (c *Config) Proficiency(id string) (Proficiency, bool) {

	for _, p := range c.Proficiencies {
		if p.ID == id {
			return p, true
		}
	}
	return Proficiency{}, false
}

// ProficiencyCost returns the trait-point cost of a proficiency id (0 if none).
func (c *Config) ProficiencyCost(id string) int {
	if p, ok := c.Proficiency(id); ok {
		return p.Cost
	}
	return 0
}

// DefaultProficiencyID returns the id of the first proficiency, used as the
// starting tier for new characters.
func (c *Config) DefaultProficiencyID() string {
	if len(c.Proficiencies) > 0 {
		return c.Proficiencies[0].ID
	}
	return ""
}

// ResolveOptions returns the concrete option list for a field, expanding an
// options_source reference server-side when present.
func (c *Config) ResolveOptions(f Field) []Option {
	if f.OptionsSource != "" {
		return c.OptionsFor(f.OptionsSource)
	}
	return f.Options
}

// OptionGroup is a labelled set of options, used to render <optgroup> blocks.
// A group with an empty Label is rendered as ungrouped options.
type OptionGroup struct {
	Label   string
	Options []Option
}

// ResolveOptionGroups returns options grouped for display. Trait sources that
// span multiple categories (traits_all) are grouped by category so a single
// large dropdown stays readable; every other source becomes one unlabelled
// group.
func (c *Config) ResolveOptionGroups(f Field) []OptionGroup {
	if f.OptionsSource == "states_all" {
		var groups []OptionGroup
		var gen []Option
		for _, s := range c.GeneralStates {
			gen = append(gen, Option{Value: "general." + s.ID, Label: s.Name})
		}
		if len(gen) > 0 {
			groups = append(groups, OptionGroup{Label: "General", Options: gen})
		}
		var spec []Option
		for _, s := range c.SpecificStates {
			cost := &Cost{BuildCost: s.BuildCost, EnergyCost: s.EnergyCost}
			spec = append(spec, Option{Value: "specific." + s.ID, Label: s.Name, Cost: cost})
		}
		if len(spec) > 0 {
			groups = append(groups, OptionGroup{Label: "Specific", Options: spec})
		}
		return groups
	}

	if f.OptionsSource == "traits_all" {
		var groups []OptionGroup

		// Do not deduplicate across categories: a trait such as "Magic" or
		// "Mind" legitimately exists in more than one category (e.g. offense
		// and defense), and each category's entry must remain selectable. The
		// option value is namespaced by category so the two are distinct.
		for _, cat := range c.traitCategories() {
			var opts []Option
			// When the field defines group offsets, surface the group's
			// offset cost on each option so the (-/+x pt, -/+y E) hint shows.
			var groupCost *Cost
			if f.GroupOffsets != nil {
				if oc, ok := f.GroupOffsets.Offsets[cat]; ok {
					groupCost = oc
				}
			}
			for _, t := range c.Traits.Items[cat] {
				opts = append(opts, Option{Value: cat + "." + t, Label: t, Cost: groupCost})
			}
			if len(opts) > 0 {
				groups = append(groups, OptionGroup{Label: titleCase(cat), Options: opts})
			}
		}
		return groups

	}

	return []OptionGroup{{Label: "", Options: c.ResolveOptions(f)}}
}

// traitCategories returns the ordered trait group ids that make up "traits_all".
// It is config-driven via trait_categories, falling back to the historical
// general/offense/defense set when unset.
func (c *Config) traitCategories() []string {
	if len(c.TraitCategories) > 0 {
		return c.TraitCategories
	}
	return []string{"general", "offense", "defense"}
}

// OptionsFor resolves a named options_source into a concrete option list. All
// dynamic sources are derived from the config (traits, dice, states); a handful
// of small static lists are defined here.
func (c *Config) OptionsFor(source string) []Option {
	switch source {
	case "traits_general":
		return strOptions(c.Traits.Items["general"])
	case "traits_offense":
		return strOptions(c.Traits.Items["offense"])
	case "traits_defense":
		return strOptions(c.Traits.Items["defense"])
	case "traits_vital":
		return strOptions(c.Traits.Items["vital"])
	case "traits_all":
		// Vitals (HP/Movement/Energy) are not selectable as traits.
		var all []string
		seen := map[string]bool{}
		for _, cat := range c.traitCategories() {
			for _, t := range c.Traits.Items[cat] {
				if !seen[t] {
					seen[t] = true
					all = append(all, t)
				}
			}
		}
		return strOptions(all)

	case "dice_damage":
		return strOptions(c.Dice.Damage)
	case "dice_generic":
		return strOptions(c.Dice.Generic)
	case "states_general":
		out := make([]Option, 0, len(c.GeneralStates))
		for _, s := range c.GeneralStates {
			out = append(out, Option{Value: s.ID, Label: s.Name})
		}
		return out
	case "states_specific":
		out := make([]Option, 0, len(c.SpecificStates))
		for _, s := range c.SpecificStates {
			out = append(out, Option{Value: s.ID, Label: s.Name})
		}
		return out
	case "ability_types":
		return componentOptions(c.AbilityTypes)
	case "enactment_types":
		return componentOptions(c.Enactments)
	case "interaction_types":
		return componentOptions(c.Interactions)
	default:
		// A costed source takes precedence over the plain string list so
		// per-entry (e.g. per-trigger) build costs can be attached in YAML.
		if opts, ok := c.OptionSourcesCosted[source]; ok {
			return costedOptions(opts)
		}
		// Any other name is resolved from the config-driven option_sources
		// map, so static lists (directions, trigger timings, reaction
		// triggers, knockout options, etc.) live in YAML rather than Go.
		if vals, ok := c.OptionSources[source]; ok {
			return strOptions(vals)
		}
		return nil
	}
}

// costedOptions normalizes a costed option list: an entry with an empty Label
// falls back to its Value so the display stays populated.
func costedOptions(opts []Option) []Option {
	out := make([]Option, 0, len(opts))
	for _, o := range opts {
		if o.Label == "" {
			o.Label = o.Value
		}
		out = append(out, o)
	}
	return out
}

// GroupOffsetFor returns the group-offset cost for a selected trait value on a
// field, or nil when the field has no group offsets or the value's group has no
// configured offset. The value is expected to be namespaced as "group.Trait"
// (as produced by the traits_all source); a value without a prefix uses the
// default group.
func (c *Config) GroupOffsetFor(f Field, value string) *Cost {
	if f.GroupOffsets == nil || value == "" {
		return nil
	}
	group := f.GroupOffsets.DefaultGroup
	if i := indexByte(value, '.'); i >= 0 {
		group = value[:i]
	}
	if cost, ok := f.GroupOffsets.Offsets[group]; ok {
		return cost
	}
	return nil
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

func strOptions(vals []string) []Option {

	out := make([]Option, 0, len(vals))
	for _, v := range vals {
		out = append(out, Option{Value: v, Label: v})
	}
	return out
}

func componentOptions(m ComponentMap) []Option {
	out := make([]Option, 0, len(m.Order))
	for _, comp := range m.List() {
		out = append(out, Option{Value: comp.ID, Label: comp.DisplayName()})
	}
	return out
}

// TraitPointBudget returns the trait-point budget for a given character level,
// read from the leveling table.
func (c *Config) TraitPointBudget(level int) int {
	return budgetForLevel(c.Leveling.TraitPoints, level)
}

// AbilityPointBudget returns the ability-point budget for a given level.
func (c *Config) AbilityPointBudget(level int) int {
	return budgetForLevel(c.Leveling.AbilityPoints, level)
}

func budgetForLevel(t LevelTable, level int) int {
	if level < 1 {
		level = 1
	}
	best := 0
	for _, e := range t.Levels {
		if e.Level <= level && e.Total >= best {
			best = e.Total
		}
		if e.Level == level {
			return e.Total
		}
	}
	return best
}
