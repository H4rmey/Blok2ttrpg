package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// OptionList is a named option source list. Each entry may be authored either
// as a bare string (value == label, no cost) or as a full object with an
// optional label and cost. This lets a single source mix plain and costed
// entries, replacing the former split between option_sources and
// option_sources_costed.
type OptionList []Option

// UnmarshalYAML accepts a sequence whose items are either scalars (shorthand
// value strings) or mappings (full Option objects).
func (l *OptionList) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.SequenceNode {
		return fmt.Errorf("expected sequence for option list, got kind %d", n.Kind)
	}
	out := make(OptionList, 0, len(n.Content))
	for _, item := range n.Content {
		switch item.Kind {
		case yaml.ScalarNode:
			out = append(out, Option{Value: item.Value})
		case yaml.MappingNode:
			var opt Option
			if err := item.Decode(&opt); err != nil {
				return fmt.Errorf("option entry: %w", err)
			}
			out = append(out, opt)
		default:
			return fmt.Errorf("unexpected option entry kind %d", item.Kind)
		}
	}
	*l = out
	return nil
}

// Options normalizes the list into concrete options: an entry with an empty
// label falls back to its value so display text stays populated.
func (l OptionList) Options() []Option {
	out := make([]Option, 0, len(l))
	for _, o := range l {
		if o.Label == "" {
			o.Label = o.Value
		}
		out = append(out, o)
	}
	return out
}

// OptionGroupDef defines a named grouped dropdown source: an ordered set of
// member groups, each resolving to an underlying option source. It replaces the
// hardcoded traits_all/roll_all/conditions_all grouping in Go.
type OptionGroupDef struct {
	Groups []OptionGroupMember `yaml:"groups" json:"groups"`
}

// OptionGroupMember is one <optgroup> within a grouped source.
type OptionGroupMember struct {
	// Source names the underlying option source to resolve for this group. It
	// may be a plain source name (e.g. "damage_types"), a dotted trait/dice
	// reference (e.g. "traits.general", "dice.generic"), or one of the built-in
	// dynamic sources ("general_conditions", "specific_conditions").
	Source string `yaml:"source" json:"source"`

	// Label is the optgroup heading. When empty a heading is derived from the
	// source (its trait category title, or the source name).
	Label string `yaml:"label,omitempty" json:"label,omitempty"`

	// Namespace, when set, prefixes each option value as "<namespace>.<value>"
	// so options from different groups stay distinct (e.g. "general.Blinded").
	// When empty and the source is a trait category, the category id is used as
	// the namespace so group offsets can key off it.
	Namespace string `yaml:"namespace,omitempty" json:"namespace,omitempty"`

	// OffsetKey names the group_offsets key applied to every option in this
	// group. When empty it falls back to Namespace (or the trait category id).
	OffsetKey string `yaml:"offset_key,omitempty" json:"offset_key,omitempty"`
}
