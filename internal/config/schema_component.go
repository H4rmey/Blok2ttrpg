// This file defines the component and field schema: the generic building block
// shared by perk types, enactments and interactions, and the field definitions
// that drive both the builder UI and the cost engine.
package config

// Component is a generic perk building block: an perk type, enactment or
// interaction. Fields drive the builder UI and the cost engine; BaseCost is the
// flat component cost. The Base*/Default* values are advisory rule parameters
// (starting energy, action, range, etc.) surfaced by the documentation and
// character sheet. Nothing is special-cased by component id in Go, so new types
// can be added purely in YAML.
type Component struct {
	ID          string `yaml:"-" json:"id"`
	Name        string `yaml:"name,omitempty" json:"name,omitempty"`
	Type        string `yaml:"type,omitempty" json:"type,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Information is optional help text surfaced as a hover tooltip via a small
	// "i" indicator next to the component in the builder UI.
	Information string `yaml:"information,omitempty" json:"information,omitempty"`
	// RenderInformation, when true, renders Information as plain text between
	// the component header and its dropdown instead of behind a hover "i".
	RenderInformation bool `yaml:"render_information,omitempty" json:"render_information,omitempty"`
	BaseCost          Cost `yaml:"base_cost,omitempty" json:"base_cost,omitempty"`

	BaseEnergy int `yaml:"base_energy,omitempty" json:"base_energy,omitempty"`
	BaseAction int `yaml:"base_action,omitempty" json:"base_action,omitempty"`

	// Perk-type base parameters. Not every component sets all of these;
	// unset values decode as zero.
	BaseRange           int `yaml:"base_range,omitempty" json:"base_range,omitempty"`
	BaseUses            int `yaml:"base_uses,omitempty" json:"base_uses,omitempty"`
	BaseDuration        int `yaml:"base_duration,omitempty" json:"base_duration,omitempty"`
	BaseReverseDuration int `yaml:"base_reverse_duration,omitempty" json:"base_reverse_duration,omitempty"`
	BaseHealth          int `yaml:"base_health,omitempty" json:"base_health,omitempty"`
	BaseLifetime        int `yaml:"base_lifetime,omitempty" json:"base_lifetime,omitempty"`
	BaseUpkeepAction    int `yaml:"base_upkeep_action,omitempty" json:"base_upkeep_action,omitempty"`
	BaseUpkeepEnergy    int `yaml:"base_upkeep_energy,omitempty" json:"base_upkeep_energy,omitempty"`

	// SkipInvokeCost marks a perk type whose use is exempt from an invoke point
	// cost that the equivalent improvised action would pay. The Reaction perk
	// type sets it: a freeform out-of-turn action costs an invoke point, but a
	// reaction bought with build points does not.
	//
	// Nothing in the cost engine reads this, because invoke points are not a
	// build currency and are never summed by the engine. It exists so the
	// generated documentation and the character sheet can state the exemption
	// instead of a config author having to write it out in prose.
	SkipInvokeCost *bool `yaml:"skip_invoke_cost,omitempty" json:"skip_invoke_cost,omitempty"`

	// DefaultRange/DefaultTargets etc. are used by interaction components.
	DefaultRange    int `yaml:"default_range,omitempty" json:"default_range,omitempty"`
	DefaultTargets  int `yaml:"default_targets,omitempty" json:"default_targets,omitempty"`
	DefaultRadius   int `yaml:"default_radius,omitempty" json:"default_radius,omitempty"`
	DefaultDuration int `yaml:"default_duration,omitempty" json:"default_duration,omitempty"`

	// UseDCValidation opts an enactment into flat-DC validation: instead of a
	// contested roll against the target's counter skills, the enactment rolls
	// its engage source against a fixed DC configured under validations in
	// general.yaml. Unset (false) keeps the historical contested roll. This is
	// enactment-only by design; interactions never force it.
	UseDCValidation bool `yaml:"use_dc_validation,omitempty" json:"use_dc_validation,omitempty"`

	// Allowed/blocked lists drive UI filtering only; they are never enforced
	// on save. The rule is: when the allowed list is non-empty only those ids
	// are shown (in config order); otherwise when the blocked list is
	// non-empty everything except those ids is shown; otherwise everything is
	// shown. AllowedInteractions/BlockedInteractions and AllowedValidations/
	// BlockedValidations apply to enactment components (filtering the
	// interaction dropdown and validation fields shown for that enactment).
	// AllowedEnactments/BlockedEnactments apply to perk-type components
	// (filtering the enactment dropdown shown for that perk type).
	AllowedInteractions []string `yaml:"allowed_interactions,omitempty" json:"allowed_interactions,omitempty"`
	BlockedInteractions []string `yaml:"blocked_interactions,omitempty" json:"blocked_interactions,omitempty"`
	AllowedValidations  []string `yaml:"allowed_validations,omitempty" json:"allowed_validations,omitempty"`
	BlockedValidations  []string `yaml:"blocked_validations,omitempty" json:"blocked_validations,omitempty"`
	AllowedEnactments   []string `yaml:"allowed_enactments,omitempty" json:"allowed_enactments,omitempty"`
	BlockedEnactments   []string `yaml:"blocked_enactments,omitempty" json:"blocked_enactments,omitempty"`

	Fields []Field `yaml:"fields,omitempty" json:"fields,omitempty"`
}

// SkipsInvokeCost reports whether using this perk type is exempt from the
// invoke point cost of an improvised equivalent. Defaults to false when unset.
func (c Component) SkipsInvokeCost() bool {
	return c.SkipInvokeCost != nil && *c.SkipInvokeCost
}

// DisplayName returns the human-facing label for a component. Perk types use
// "name"; enactments and interactions use "type" as their display name.
func (c Component) DisplayName() string {
	if c.Name != "" {
		return c.Name
	}
	if c.Type != "" {
		return c.Type
	}
	return c.ID
}

// UsesDCValidation reports whether the enactment validates against a flat DC
// rather than a contested roll.
func (c Component) UsesDCValidation() bool { return c.UseDCValidation }

// Field drives both the builder UI and the cost engine.
type Field struct {
	Key   string `yaml:"key" json:"key"`
	Label string `yaml:"label" json:"label"`
	Type  string `yaml:"type" json:"type"` // checkbox, dropdown, free_text, free_number, multiselect, conditions

	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Information is optional help text surfaced as a hover tooltip via a small
	// "i" indicator next to the field label in the builder UI.
	Information string `yaml:"information,omitempty" json:"information,omitempty"`
	// RenderInformation, when true, renders Information as plain text below the
	// field instead of behind a hover "i".
	RenderInformation bool `yaml:"render_information,omitempty" json:"render_information,omitempty"`

	Default any `yaml:"default,omitempty" json:"default,omitempty"`

	// Flat cost applied when the field is "on" (checkbox true, dropdown value
	// selected, etc.).
	Cost *Cost `yaml:"cost,omitempty" json:"cost,omitempty"`

	// free_number bounds, step and rounding, plus per-step increase/decrease.
	Min      int      `yaml:"min,omitempty" json:"min,omitempty"`
	Max      int      `yaml:"max,omitempty" json:"max,omitempty"`
	Step     int      `yaml:"step,omitempty" json:"step,omitempty"`
	Rounding string   `yaml:"rounding,omitempty" json:"rounding,omitempty"` // ceil or floor
	PerStep  *PerStep `yaml:"per_step,omitempty" json:"per_step,omitempty"`

	// dropdown options (inline) or a reference to a named option source.
	Options       []Option `yaml:"options,omitempty" json:"options,omitempty"`
	OptionsSource string   `yaml:"options_source,omitempty" json:"options_source,omitempty"`

	// ShiftKey, on a condition_select field, names the sibling field that holds
	// the per-condition shift amount for general conditions. Defaults to
	// "shift_amount" when unset. The general condition's shift_cost is multiplied
	// by the absolute shift value read from that sibling field.
	ShiftKey string `yaml:"shift_key,omitempty" json:"shift_key,omitempty"`

	// multiselect/conditions: a repeatable set of rows built from RowFields.
	// PerItem is the cost delta per row relative to DefaultCount.
	RowFields    []Field  `yaml:"row_fields,omitempty" json:"row_fields,omitempty"`
	DefaultCount int      `yaml:"default_count,omitempty" json:"default_count,omitempty"`
	PerItem      *PerStep `yaml:"per_item,omitempty" json:"per_item,omitempty"`
	// RowDefaults pre-fills the initial rows of a multiselect/conditions field.
	// Each entry is a map of row_field key -> default value for that row,
	// applied in order to the first rows rendered.
	RowDefaults []map[string]string `yaml:"row_defaults,omitempty" json:"row_defaults,omitempty"`

	// Conjunction is the small joining word rendered to the left of each
	// multiselect row after the first ("and" or "or"). It is display-only and
	// defaults to "or" when unset.
	Conjunction string `yaml:"conjunction,omitempty" json:"conjunction,omitempty"`

	// Conditional visibility: show this field only when the field named
	// VisibilityWhen currently equals ShowWhen.
	VisibilityWhen string `yaml:"visibility_when,omitempty" json:"visibility_when,omitempty"`
	ShowWhen       string `yaml:"show_when,omitempty" json:"show_when,omitempty"`

	// InlineBuilder, when set on a dropdown field, spawns a nested inline
	// builder for the component the selected value refers to. The referenced
	// component's own fields render underneath the dropdown and contribute
	// their (field-driven) cost to the total.
	InlineBuilder *InlineBuilder `yaml:"inline_builder,omitempty" json:"inline_builder,omitempty"`

	// GroupOffsets applies a per-skill-group cost offset on a dropdown backed
	// by a multi-group skill source (skills_all). The selected option value is
	// namespaced as "group.Skill"; the group prefix selects which offset to
	// add. This lets a field "lean" toward a preferred skill group: picking a
	// skill outside the leaning group can cost extra (or a preferred group can
	// cost less).
	GroupOffsets *GroupOffsets `yaml:"group_offsets,omitempty" json:"group_offsets,omitempty"`
}

// GroupOffsets configures per-skill-group cost offsets for a skill dropdown.
// DefaultGroup names the preferred (leaning) group; Offsets maps each skill
// group id to the cost added when a skill from that group is selected. Groups
// not present in Offsets contribute no offset.
type GroupOffsets struct {
	DefaultGroup string           `yaml:"default_group,omitempty" json:"default_group,omitempty"`
	Offsets      map[string]*Cost `yaml:"offsets,omitempty" json:"offsets,omitempty"`
}

// InlineBuilder configures a dropdown field to render a nested component
// builder for whatever option value is selected. It is fully generic so any
// dropdown in any component can opt in.
type InlineBuilder struct {
	// Kind selects which component map the selected value resolves against:
	// "enactment", "interaction" or "perk_type".
	Kind string `yaml:"kind" json:"kind"`
}

// Option is a dropdown choice which may carry its own cost and nested fields.
type Option struct {
	Value string `yaml:"value" json:"value"`
	Label string `yaml:"label,omitempty" json:"label,omitempty"`
	// Information is optional help text surfaced as a native hover tooltip on
	// the dropdown option (rendered via the option's title trait).
	Information string `yaml:"information,omitempty" json:"information,omitempty"`
	// RenderInformation, when true, renders this option's Information as plain
	// text below the dropdown once selected instead of behind the trailing
	// "i" indicator.
	RenderInformation bool    `yaml:"render_information,omitempty" json:"render_information,omitempty"`
	Cost              *Cost   `yaml:"cost,omitempty" json:"cost,omitempty"`
	Fields            []Field `yaml:"fields,omitempty" json:"fields,omitempty"`
}

// TraitGroup is a titled section of character fields.
type TraitGroup struct {
	ID     string  `yaml:"-" json:"id"`
	Label  string  `yaml:"label" json:"label"`
	Fields []Field `yaml:"fields" json:"fields"`
}
