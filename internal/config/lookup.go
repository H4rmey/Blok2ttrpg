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

// GeneralConditionByID returns the general condition with the given id.
func (c *Config) GeneralConditionByID(id string) (GeneralCondition, bool) {
	for _, s := range c.GeneralConditions {
		if s.ID == id {
			return s, true
		}
	}
	return GeneralCondition{}, false
}

// SpecificConditionByID returns the specific condition with the given id.
func (c *Config) SpecificConditionByID(id string) (SpecificCondition, bool) {
	for _, s := range c.SpecificConditions {
		if s.ID == id {
			return s, true
		}
	}
	return SpecificCondition{}, false
}

// ConditionByID returns the unified condition with the given id.
func (c *Config) ConditionByID(id string) (Condition, bool) {
	for _, s := range c.Conditions {
		if s.ID == id {
			return s, true
		}
	}
	return Condition{}, false
}

// ShiftOptionsFor returns the discrete non-zero shift values a general condition
// may take, from its configured min_shift..max_shift range. Zero is skipped
// because applying a condition with no shift is meaningless.
func (c *Config) ShiftOptionsFor(generalID string) []int {
	var min, max int
	if s, ok := c.GeneralConditionByID(generalID); ok {
		min, max = s.MinShift, s.MaxShift
	} else if u, ok := c.ConditionByID(generalID); ok {
		if !u.Shiftable() {
			// Fixed-cost unified condition: no shift dropdown.
			return nil
		}
		min, max = u.MinShift, u.MaxShift
	} else {
		return nil
	}
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

// DefaultProficiencyID returns the id of the tier new characters start every
// trait at. When default_proficiency is configured and valid it is used;
// otherwise the first proficiency in the ladder is the default.
func (c *Config) DefaultProficiencyID() string {
	if c.DefaultProficiency != "" {
		if c.proficiencyIndex(c.DefaultProficiency) >= 0 {
			return c.DefaultProficiency
		}
	}
	if len(c.Proficiencies) > 0 {
		return c.Proficiencies[0].ID
	}
	return ""
}

// DefaultProficiencyIndex returns the ladder position of the default tier, or 0
// when it cannot be resolved.
func (c *Config) DefaultProficiencyIndex() int {
	if idx := c.proficiencyIndex(c.DefaultProficiencyID()); idx >= 0 {
		return idx
	}
	return 0
}

// proficiencyIndex returns the position of a proficiency id within the ordered
// ladder, or -1 when it is not found.
func (c *Config) proficiencyIndex(id string) int {
	for i, p := range c.Proficiencies {
		if p.ID == id {
			return i
		}
	}
	return -1
}

// ShiftProficiency moves a proficiency id up or down the ordered ladder by
// delta rungs and returns the resulting id. The result is clamped to the ends
// of the ladder. An unknown current id is treated as the first (default) tier
// so a shift still produces a sensible result.
func (c *Config) ShiftProficiency(current string, delta int) string {
	if len(c.Proficiencies) == 0 {
		return current
	}
	idx := c.proficiencyIndex(current)
	if idx < 0 {
		idx = 0
	}
	idx += delta
	if idx < 0 {
		idx = 0
	}
	if idx > len(c.Proficiencies)-1 {
		idx = len(c.Proficiencies) - 1
	}
	return c.Proficiencies[idx].ID
}

// ShiftClamped reports whether shifting the given proficiency id by delta
// rungs would run off either end of the ladder (i.e. the requested delta could
// not be fully applied). It is used to surface a non-blocking warning when a
// package pushes a trait above or below the possible range.
func (c *Config) ShiftClamped(current string, delta int) bool {
	if len(c.Proficiencies) == 0 || delta == 0 {
		return false
	}
	idx := c.proficiencyIndex(current)
	if idx < 0 {
		idx = 0
	}
	target := idx + delta
	return target < 0 || target > len(c.Proficiencies)-1
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

// ResolveOptionGroups returns options grouped for display. When the field's
// options_source names an entry in cfg.OptionGroups, that config-defined group
// layout is expanded into <optgroup> blocks; otherwise the source resolves to a
// single unlabelled group.
func (c *Config) ResolveOptionGroups(f Field) []OptionGroup {
	if def, ok := c.OptionGroups[f.OptionsSource]; ok {
		return c.expandOptionGroups(f, def)
	}
	return []OptionGroup{{Label: "", Options: c.ResolveOptions(f)}}
}

// expandOptionGroups turns a config-defined grouped source into labelled option
// groups, applying per-group namespacing and group-offset costs.
func (c *Config) expandOptionGroups(f Field, def OptionGroupDef) []OptionGroup {
	var groups []OptionGroup
	for _, m := range def.Groups {
		opts := c.OptionsFor(m.Source)
		if len(opts) == 0 {
			continue
		}
		ns := m.Namespace
		if ns == "" {
			ns = traitCategoryOf(m.Source)
		}
		offsetKey := m.OffsetKey
		if offsetKey == "" {
			offsetKey = ns
		}
		var groupCost *Cost
		if f.GroupOffsets != nil && offsetKey != "" {
			if oc, ok := f.GroupOffsets.Offsets[offsetKey]; ok {
				groupCost = oc
			}
		}
		out := make([]Option, 0, len(opts))
		for _, o := range opts {
			if ns != "" {
				o.Value = ns + "." + o.Value
			}
			// Merge any per-option cost with the group offset cost.
			o.Cost = mergeCost(o.Cost, groupCost)
			out = append(out, o)
		}
		label := m.Label
		if label == "" {
			label = c.groupLabel(m.Source)
		}
		groups = append(groups, OptionGroup{Label: label, Options: out})
	}
	return groups
}

// mergeCost returns the sum of two optional costs, or nil when both are nil.
func mergeCost(a, b *Cost) *Cost {
	if a == nil && b == nil {
		return nil
	}
	out := Cost{}
	if a != nil {
		out.BuildCost += a.BuildCost
		out.EnergyCost += a.EnergyCost
	}
	if b != nil {
		out.BuildCost += b.BuildCost
		out.EnergyCost += b.EnergyCost
	}
	return &out
}

// traitCategoryOf returns the trait category id when source is a dotted
// "traits.<cat>" reference, or "" otherwise. The category id doubles as the
// default namespace/offset key for a trait group.
func traitCategoryOf(source string) string {
	const prefix = "traits."
	if len(source) > len(prefix) && source[:len(prefix)] == prefix {
		return source[len(prefix):]
	}
	return ""
}

// groupLabel derives a default optgroup heading from a source name.
func (c *Config) groupLabel(source string) string {
	if cat := traitCategoryOf(source); cat != "" {
		return titleCase(cat)
	}
	return titleCase(source)
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

// OptionsFor resolves a named options_source into a concrete option list. It
// understands dotted trait/dice references (traits.<cat>, dice.<kind>), the
// built-in condition sources, component sources, config-defined grouped sources
// (flattened for the cost engine), and the config-driven option_sources map.
func (c *Config) OptionsFor(source string) []Option {
	// Dotted references: "traits.<category>" and "dice.<kind>".
	if cat := traitCategoryOf(source); cat != "" {
		return strOptions(c.Traits.Items[cat])
	}
	switch source {
	case "dice.damage":
		return strOptions(c.Dice.Damage)
	case "dice.generic":
		return strOptions(c.Dice.Generic)
	case "general_conditions":
		out := make([]Option, 0, len(c.GeneralConditions))
		for _, s := range c.GeneralConditions {
			out = append(out, Option{Value: s.ID, Label: s.Name})
		}
		return out
	case "specific_conditions":
		out := make([]Option, 0, len(c.SpecificConditions))
		for _, s := range c.SpecificConditions {
			cost := &Cost{BuildCost: s.BuildCost, EnergyCost: s.EnergyCost}
			out = append(out, Option{Value: s.ID, Label: s.Name, Cost: cost})
		}
		return out
	case "conditions":
		out := make([]Option, 0, len(c.Conditions))
		for _, s := range c.Conditions {
			// Shiftable conditions charge per-shift (handled by the cost
			// engine via ConditionByID), so they carry no flat option cost.
			// Fixed-cost conditions attach their build/energy cost so the
			// dropdown option contributes directly.
			var cost *Cost
			if !s.Shiftable() && (s.BuildCost != 0 || s.EnergyCost != 0) {
				cost = &Cost{BuildCost: s.BuildCost, EnergyCost: s.EnergyCost}
			}
			out = append(out, Option{Value: s.ID, Label: s.Name, Cost: cost})
		}
		return out
	case "ability_types":
		return componentOptions(c.AbilityTypes)

	case "enactment_types":
		return componentOptions(c.Enactments)
	case "interaction_types":
		return componentOptions(c.Interactions)
	}

	// A grouped source flattens to the concatenation of its member groups,
	// namespaced the same way as the grouped display, so the cost engine can
	// match posted values. Group-offset costs are applied via GroupOffsetFor.
	if def, ok := c.OptionGroups[source]; ok {
		var out []Option
		for _, m := range def.Groups {
			opts := c.OptionsFor(m.Source)
			ns := m.Namespace
			if ns == "" {
				ns = traitCategoryOf(m.Source)
			}
			for _, o := range opts {
				if ns != "" {
					o.Value = ns + "." + o.Value
				}
				out = append(out, o)
			}
		}
		return out
	}

	// A costed variant (legacy split map) takes precedence over the merged
	// option_sources entry of the same name.
	if opts, ok := c.OptionSourcesCosted[source]; ok {
		return opts.Options()
	}
	// Config-driven named lists (directions, trigger events, reaction triggers,
	// knockout options, etc.), each entry optionally carrying its own cost.
	if opts, ok := c.OptionSources[source]; ok {
		return opts.Options()
	}
	return nil
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
