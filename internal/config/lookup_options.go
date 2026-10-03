// Option resolution: turning a field's inline options or named options_source into a concrete list, including grouped sources and group-offset costs.
package config

import "strings"

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
			ns = skillCategoryOf(m.Source)
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

// skillCategoryOf returns the skill category id when source is a dotted
// "skills.<cat>" reference, or "" otherwise. The category id doubles as the
// default namespace/offset key for a skill group.
func skillCategoryOf(source string) string {
	const prefix = "skills."
	if len(source) > len(prefix) && source[:len(prefix)] == prefix {
		return source[len(prefix):]
	}
	return ""
}

// groupLabel derives a default optgroup heading from a source name.
func (c *Config) groupLabel(source string) string {
	if cat := skillCategoryOf(source); cat != "" {
		return titleCase(cat)
	}
	return titleCase(source)
}

// OptionsFor resolves a named options_source into a concrete option list. It

// understands dotted skill/dice references (skills.<cat>, dice.<kind>), the
// built-in condition sources, component sources, config-defined grouped sources
// (flattened for the cost engine), and the config-driven option_sources map.
func (c *Config) OptionsFor(source string) []Option {
	// Dotted references: "skills.<category>" and "dice.<kind>".
	if cat := skillCategoryOf(source); cat != "" {
		return skillOptions(c.Skills, cat)
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
			// Conditions marked selectable: false are states the rules impose
			// rather than effects an perk can buy, so they never appear in
			// the dropdown. They stay resolvable via ConditionByID.
			if !s.IsSelectable() {
				continue
			}
			// Shiftable conditions charge per-shift (handled by the cost
			// engine via ConditionByID), so they carry no flat option cost.
			// Fixed-cost conditions attach their build/energy cost so the
			// dropdown option contributes directly.
			var cost *Cost
			if !s.Shiftable() && (s.BuildCost != 0 || s.EnergyCost != 0) {
				cost = &Cost{BuildCost: s.BuildCost, EnergyCost: s.EnergyCost}
			}
			// The description doubles as the option's hover tooltip, so a
			// player can read what a condition does before picking it.
			out = append(out, Option{Value: s.ID, Label: s.Name, Information: s.Description, Cost: cost})
		}
		return out
	case "passives":
		// The passive catalogue as dropdown options. Each entry carries its flat
		// build cost, so the generic cost engine prices the selected passive
		// without knowing anything about passives. The cost of the entry's own
		// fields is added by the engine from those fields directly.
		out := make([]Option, 0, len(c.Passives.Entries))
		for _, p := range c.Passives.Entries {
			out = append(out, Option{
				Value: p.ID,
				Label: p.Name,
				// The rules text doubles as the option tooltip so a player can
				// read what a passive does before picking it. Placeholders are
				// left unsubstituted here: nothing has been configured yet.
				Information: p.Description,
				Cost: &Cost{
					BuildCost:  c.PassiveFlatCost(p),
					EnergyCost: c.Passives.EnergyCost,
				},
			})
		}
		return out
	case "perk_types":
		return componentOptions(c.PerkTypes)

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
				ns = skillCategoryOf(m.Source)
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

// GroupOffsetFor returns the group-offset cost for a selected skill value on a
// field, or nil when the field has no group offsets or the value's group has no
// configured offset. The value is expected to be namespaced as "group.Skill"
// (as produced by the skills_all source); a value without a prefix uses the
// default group.
func (c *Config) GroupOffsetFor(f Field, value string) *Cost {
	if f.GroupOffsets == nil || value == "" {
		return nil
	}
	group := f.GroupOffsets.DefaultGroup
	if i := strings.IndexByte(value, '.'); i >= 0 {

		group = value[:i]
	}
	if cost, ok := f.GroupOffsets.Offsets[group]; ok {
		return cost
	}
	return nil
}

func strOptions(vals []string) []Option {

	out := make([]Option, 0, len(vals))
	for _, v := range vals {
		out = append(out, Option{Value: v, Label: v})
	}
	return out
}

// skillOptions turns one skill category into dropdown options carrying each
// skill's configured information, so the builder's skill dropdowns (and any
// grouped source built from them) show the same hover text the sheet does.
func skillOptions(m SkillMap, cat string) []Option {
	vals := m.Items[cat]
	out := make([]Option, 0, len(vals))
	for _, v := range vals {
		out = append(out, Option{Value: v, Label: v, Information: m.Info(cat, v), RenderInformation: m.RenderInformation})
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
