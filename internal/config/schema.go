package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Cost is a simple additive cost pair used everywhere in the ruleset.
// Both values are additive: a positive BuildCost makes an option cost more
// build points, a positive EnergyCost makes it cost more energy.
type Cost struct {
	BuildCost  int `yaml:"build_cost,omitempty" json:"build_cost,omitempty"`
	EnergyCost int `yaml:"energy_cost,omitempty" json:"energy_cost,omitempty"`
}

// PerStep describes the per-step cost of a free_number field. Increase applies
// when the value moves above its default; Decrease applies when it moves below.
type PerStep struct {
	Increase *Cost `yaml:"increase,omitempty" json:"increase,omitempty"`
	Decrease *Cost `yaml:"decrease,omitempty" json:"decrease,omitempty"`
}

// Config is the top-level ruleset. Everything the app renders and costs is
// derived from this structure, loaded from a directory of YAML files.
type Config struct {
	Version   int    `yaml:"version" json:"version"`
	ProfileID string `yaml:"profile_id" json:"profile_id"`
	Title     string `yaml:"title,omitempty" json:"title,omitempty"`

	Combat              Combat              `yaml:"combat,omitempty" json:"combat,omitempty"`
	AdditionalEnactment AdditionalEnactment `yaml:"additional_enactment,omitempty" json:"additional_enactment,omitempty"`
	Dice                Dice                `yaml:"dice,omitempty" json:"dice,omitempty"`
	Validations         Validations         `yaml:"validations,omitempty" json:"validations,omitempty"`

	// OptionSources holds named option lists so any field can reference them via
	// options_source without hardcoding the list in Go. Each entry may be a
	// plain string (value == label) or an object with an optional cost, so a
	// single source can mix free and costed entries. This replaces the former
	// split between option_sources and option_sources_costed.
	OptionSources map[string]OptionList `yaml:"option_sources,omitempty" json:"option_sources,omitempty"`

	// OptionSourcesCosted is retained only for backwards compatibility with
	// profiles that still split costed entries into their own map. New configs
	// should attach costs inline in option_sources instead. When a source name
	// exists in both maps, this variant takes precedence.
	OptionSourcesCosted map[string]OptionList `yaml:"option_sources_costed,omitempty" json:"option_sources_costed,omitempty"`

	// OptionGroups defines named grouped dropdown sources. A field referencing
	// one of these names via options_source is rendered as <optgroup> blocks in
	// author order, and its flattened option list backs the cost engine. This
	// replaces the hardcoded traits_all/roll_all/conditions_all grouping.
	OptionGroups map[string]OptionGroupDef `yaml:"option_groups,omitempty" json:"option_groups,omitempty"`

	// TraitCategories lists the trait group ids that make up the "traits_all"
	// option source and its grouped display. When empty the app falls back to
	// the historical general/offense/defense set.
	TraitCategories []string `yaml:"trait_categories,omitempty" json:"trait_categories,omitempty"`

	// VitalGroup names the trait group id whose traits (HP, Movement, Energy)
	// map to numeric vital values rather than dice. Defaults to "vital".
	VitalGroup string `yaml:"vital_group,omitempty" json:"vital_group,omitempty"`

	// Character attributes and traits are fully config-driven, keyed by id.
	Attributes AttributeMap `yaml:"attributes,omitempty" json:"attributes,omitempty"`
	Traits     TraitMap     `yaml:"traits,omitempty" json:"traits,omitempty"`

	// Proficiency tiers referenced by traits.
	Proficiencies []Proficiency `yaml:"proficiencies,omitempty" json:"proficiencies,omitempty"`

	// DefaultProficiency names the proficiency tier id that new characters
	// start every trait at (the "free" baseline). When empty the first tier in
	// the Proficiencies list is used. Tiers below the default are free; tiers
	// above the default accrue their cumulative per-tier cost.
	DefaultProficiency string `yaml:"default_proficiency,omitempty" json:"default_proficiency,omitempty"`

	// Leveling budgets, given as per-level tables.
	Leveling Leveling `yaml:"leveling,omitempty" json:"leveling,omitempty"`

	// Ability building blocks, keyed by id but with author ordering preserved.
	AbilityTypes ComponentMap `yaml:"ability_types,omitempty" json:"ability_types,omitempty"`
	Enactments   ComponentMap `yaml:"enactments,omitempty" json:"enactments,omitempty"`
	Interactions ComponentMap `yaml:"interactions,omitempty" json:"interactions,omitempty"`

	// Conditions for the "Enact Condition" enactment.
	AdditionalCondition Cost `yaml:"additional_condition,omitempty" json:"additional_condition,omitempty"`

	// Conditions is the unified condition list. An entry is "shiftable" when it
	// declares a shift range (min_shift/max_shift) and pays shift_cost per unit
	// of shift; otherwise it is a fixed-cost condition paying build_cost/
	// energy_cost. This replaces the former general/specific split.
	Conditions []Condition `yaml:"conditions,omitempty" json:"conditions,omitempty"`

	// GeneralConditions/SpecificConditions are retained for backwards
	// compatibility with profiles that still split conditions into two lists.
	// New configs should use the unified Conditions list instead.
	GeneralConditions  []GeneralCondition  `yaml:"general_conditions,omitempty" json:"general_conditions,omitempty"`
	SpecificConditions []SpecificCondition `yaml:"specific_conditions,omitempty" json:"specific_conditions,omitempty"`

	// FileOrder lists the ordered markdown files for documentation, relative
	// to the module root.
	FileOrder []string `yaml:"file_order,omitempty" json:"file_order,omitempty"`
}

// Combat holds combat-wide settings.
type Combat struct {
	Actions struct {
		Amount int `yaml:"amount" json:"amount"`
	} `yaml:"actions" json:"actions"`
}

// AdditionalEnactment is the surcharge for each enactment beyond the first.
type AdditionalEnactment struct {
	BuildCost   int    `yaml:"build_cost,omitempty" json:"build_cost,omitempty"`
	EnergyCost  int    `yaml:"energy_cost,omitempty" json:"energy_cost,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// RequireInteraction/RequireValidation control whether the second and
	// following enactments must display an Interaction / Validation region.
	// The first enactment always shows both. These are pointers so an unset
	// value defaults to true (preserving the historical mandatory behaviour);
	// setting either to false hides that region for enactments beyond the
	// first.
	RequireInteraction *bool `yaml:"require_interaction,omitempty" json:"require_interaction,omitempty"`
	RequireValidation  *bool `yaml:"require_validation,omitempty" json:"require_validation,omitempty"`
}

// AsCost converts the surcharge into a plain Cost.
func (a AdditionalEnactment) AsCost() Cost {
	return Cost{BuildCost: a.BuildCost, EnergyCost: a.EnergyCost}
}

// RequiresInteraction reports whether enactments beyond the first must show an
// Interaction region. Defaults to true when unset.
func (a AdditionalEnactment) RequiresInteraction() bool {
	return a.RequireInteraction == nil || *a.RequireInteraction
}

// RequiresValidation reports whether enactments beyond the first must show a
// Validation region. Defaults to true when unset.
func (a AdditionalEnactment) RequiresValidation() bool {
	return a.RequireValidation == nil || *a.RequireValidation
}

// Dice lists the die tiers available for damage and generic rolls.
type Dice struct {
	Damage  []string `yaml:"damage,omitempty" json:"damage,omitempty"`
	Generic []string `yaml:"generic,omitempty" json:"generic,omitempty"`
}

// Validations captures the engagement/counter configuration and its fields.
type Validations struct {
	// Information is optional section-level help text. By default it renders as
	// plain text between the Validation header and the first field; when
	// RenderInformation is true it collapses into a hover "i" badge next to the
	// header instead.
	Information       string  `yaml:"information,omitempty" json:"information,omitempty"`
	RenderInformation bool    `yaml:"render_information,omitempty" json:"render_information,omitempty"`
	Fields            []Field `yaml:"fields,omitempty" json:"fields,omitempty"`
}

// Proficiency is a single skill tier.
type Proficiency struct {
	ID   string `yaml:"id" json:"id"`
	Name string `yaml:"name" json:"name"`
	Cost int    `yaml:"cost" json:"cost"`
	Note string `yaml:"note,omitempty" json:"note,omitempty"`
	// Die is the fallback die used for every dice-backed trait group at this
	// tier. Per-group overrides in Dice take precedence when present, so a tier
	// only needs the verbose Dice map when a group differs from the rest.
	Die    string            `yaml:"die,omitempty" json:"die,omitempty"`
	Dice   map[string]string `yaml:"dice,omitempty" json:"dice,omitempty"`
	Vitals map[string]any    `yaml:"vitals,omitempty" json:"vitals,omitempty"`
}

// DieFor returns the die this tier grants for a trait group: the per-group
// override in Dice when present, otherwise the shared Die fallback.
func (p Proficiency) DieFor(group string) string {
	if p.Dice != nil {
		if d, ok := p.Dice[group]; ok && d != "" {
			return d
		}
	}
	return p.Die
}

// Leveling describes the point budgets available to a character by level.
type Leveling struct {
	MaxLevel      int        `yaml:"max_level,omitempty" json:"max_level,omitempty"`
	TraitPoints   LevelTable `yaml:"trait_points,omitempty" json:"trait_points,omitempty"`
	AbilityPoints LevelTable `yaml:"ability_points,omitempty" json:"ability_points,omitempty"`
}

// LevelTable holds a per-level budget table.
type LevelTable struct {
	StandardTraitCount int          `yaml:"standard_trait_count,omitempty" json:"standard_trait_count,omitempty"`
	StartingFormula    string       `yaml:"starting_formula,omitempty" json:"starting_formula,omitempty"`
	Levels             []LevelEntry `yaml:"levels,omitempty" json:"levels,omitempty"`
}

// LevelEntry is one row in a level table.
type LevelEntry struct {
	Level        int `yaml:"level" json:"level"`
	PointsGained int `yaml:"points_gained" json:"points_gained"`
	Total        int `yaml:"total" json:"total"`
}

// GeneralCondition is a shiftable condition applied via the condition enactment.
type GeneralCondition struct {
	ID          string `yaml:"id" json:"id"`
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	MinShift    int    `yaml:"min_shift,omitempty" json:"min_shift,omitempty"`
	MaxShift    int    `yaml:"max_shift,omitempty" json:"max_shift,omitempty"`
	ShiftCost   Cost   `yaml:"shift_cost,omitempty" json:"shift_cost,omitempty"`
}

// SpecificCondition is a fixed-cost named condition.
type SpecificCondition struct {
	ID          string `yaml:"id" json:"id"`
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	BuildCost   int    `yaml:"build_cost,omitempty" json:"build_cost,omitempty"`
	EnergyCost  int    `yaml:"energy_cost,omitempty" json:"energy_cost,omitempty"`
}

// Condition is a unified condition entry. It is "shiftable" when it declares a
// non-empty shift range (min_shift/max_shift), in which case it pays ShiftCost
// per unit of applied shift; otherwise it is a fixed-cost condition paying
// BuildCost/EnergyCost. This single type replaces the former general/specific
// split.
type Condition struct {
	ID          string `yaml:"id" json:"id"`
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Fixed-cost fields (non-shiftable conditions).
	BuildCost  int `yaml:"build_cost,omitempty" json:"build_cost,omitempty"`
	EnergyCost int `yaml:"energy_cost,omitempty" json:"energy_cost,omitempty"`

	// Shiftable fields. When MinShift or MaxShift is non-zero the condition is
	// treated as shiftable and ShiftCost is charged per unit of shift.
	MinShift  int  `yaml:"min_shift,omitempty" json:"min_shift,omitempty"`
	MaxShift  int  `yaml:"max_shift,omitempty" json:"max_shift,omitempty"`
	ShiftCost Cost `yaml:"shift_cost,omitempty" json:"shift_cost,omitempty"`
}

// Shiftable reports whether the condition applies a trait shift (and therefore
// pays a per-shift cost) rather than a flat build/energy cost.
func (c Condition) Shiftable() bool {
	return c.MinShift != 0 || c.MaxShift != 0
}

// Component is a generic ability building block: an ability type, enactment or
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

	// Ability-type base parameters. Not every component sets all of these;
	// unset values decode as zero.
	BaseRange           int `yaml:"base_range,omitempty" json:"base_range,omitempty"`
	BaseUses            int `yaml:"base_uses,omitempty" json:"base_uses,omitempty"`
	BaseDuration        int `yaml:"base_duration,omitempty" json:"base_duration,omitempty"`
	BaseReverseDuration int `yaml:"base_reverse_duration,omitempty" json:"base_reverse_duration,omitempty"`
	BaseHealth          int `yaml:"base_health,omitempty" json:"base_health,omitempty"`
	BaseLifetime        int `yaml:"base_lifetime,omitempty" json:"base_lifetime,omitempty"`
	BaseUpkeepAction    int `yaml:"base_upkeep_action,omitempty" json:"base_upkeep_action,omitempty"`
	BaseUpkeepEnergy    int `yaml:"base_upkeep_energy,omitempty" json:"base_upkeep_energy,omitempty"`

	// DefaultRange/DefaultTargets etc. are used by interaction components.
	DefaultRange    int `yaml:"default_range,omitempty" json:"default_range,omitempty"`
	DefaultTargets  int `yaml:"default_targets,omitempty" json:"default_targets,omitempty"`
	DefaultRadius   int `yaml:"default_radius,omitempty" json:"default_radius,omitempty"`
	DefaultDuration int `yaml:"default_duration,omitempty" json:"default_duration,omitempty"`

	// Allowed/blocked lists drive UI filtering only; they are never enforced
	// on save. The rule is: when the allowed list is non-empty only those ids
	// are shown (in config order); otherwise when the blocked list is
	// non-empty everything except those ids is shown; otherwise everything is
	// shown. AllowedInteractions/BlockedInteractions and AllowedValidations/
	// BlockedValidations apply to enactment components (filtering the
	// interaction dropdown and validation fields shown for that enactment).
	// AllowedEnactments/BlockedEnactments apply to ability-type components
	// (filtering the enactment dropdown shown for that ability type).
	AllowedInteractions []string `yaml:"allowed_interactions,omitempty" json:"allowed_interactions,omitempty"`
	BlockedInteractions []string `yaml:"blocked_interactions,omitempty" json:"blocked_interactions,omitempty"`
	AllowedValidations  []string `yaml:"allowed_validations,omitempty" json:"allowed_validations,omitempty"`
	BlockedValidations  []string `yaml:"blocked_validations,omitempty" json:"blocked_validations,omitempty"`
	AllowedEnactments   []string `yaml:"allowed_enactments,omitempty" json:"allowed_enactments,omitempty"`
	BlockedEnactments   []string `yaml:"blocked_enactments,omitempty" json:"blocked_enactments,omitempty"`

	Fields []Field `yaml:"fields,omitempty" json:"fields,omitempty"`
}

// DisplayName returns the human-facing label for a component. Ability types use
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

	// GroupOffsets applies a per-trait-group cost offset on a dropdown backed
	// by a multi-group trait source (traits_all). The selected option value is
	// namespaced as "group.Trait"; the group prefix selects which offset to
	// add. This lets a field "lean" toward a preferred trait group: picking a
	// trait outside the leaning group can cost extra (or a preferred group can
	// cost less).
	GroupOffsets *GroupOffsets `yaml:"group_offsets,omitempty" json:"group_offsets,omitempty"`
}

// GroupOffsets configures per-trait-group cost offsets for a trait dropdown.
// DefaultGroup names the preferred (leaning) group; Offsets maps each trait
// group id to the cost added when a trait from that group is selected. Groups
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
	// "enactment", "interaction" or "ability_type".
	Kind string `yaml:"kind" json:"kind"`
}

// Option is a dropdown choice which may carry its own cost and nested fields.
type Option struct {
	Value string `yaml:"value" json:"value"`
	Label string `yaml:"label,omitempty" json:"label,omitempty"`
	// Information is optional help text surfaced as a native hover tooltip on
	// the dropdown option (rendered via the option's title attribute).
	Information string `yaml:"information,omitempty" json:"information,omitempty"`
	// RenderInformation, when true, renders this option's Information as plain
	// text below the dropdown once selected instead of behind the trailing
	// "i" indicator.
	RenderInformation bool    `yaml:"render_information,omitempty" json:"render_information,omitempty"`
	Cost              *Cost   `yaml:"cost,omitempty" json:"cost,omitempty"`
	Fields            []Field `yaml:"fields,omitempty" json:"fields,omitempty"`
}

// AttributeGroup is a titled section of character fields.
type AttributeGroup struct {
	ID     string  `yaml:"-" json:"id"`
	Label  string  `yaml:"label" json:"label"`
	Fields []Field `yaml:"fields" json:"fields"`
}

// ComponentMap is an ordered, id-keyed collection of components. YAML mapping
// order is preserved so config authors control presentation order.
type ComponentMap struct {
	Order []string
	Items map[string]*Component
	// Information is optional map-level help text. It is set from a reserved
	// top-level "information" key inside the mapping (e.g. directly under
	// ability_types:), and is not treated as a component.
	Information string
	// RenderInformation, when true, renders the map-level Information as plain
	// text between the section header and the first field instead of behind a
	// hover "i" badge next to the header. It is set from a reserved top-level
	// "render_information" key inside the mapping and is not a component.
	RenderInformation bool
}

// UnmarshalYAML decodes a mapping node into an ordered ComponentMap.
func (m *ComponentMap) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping for component map, got kind %d", n.Kind)
	}
	if m.Items == nil {
		m.Items = map[string]*Component{}
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		key := n.Content[i].Value
		// "information" is a reserved map-level key, not a component.
		if key == "information" {
			m.Information = n.Content[i+1].Value
			continue
		}
		// "render_information" is a reserved map-level key, not a component.
		if key == "render_information" {
			m.RenderInformation = n.Content[i+1].Value == "true"
			continue
		}

		var comp Component

		if err := n.Content[i+1].Decode(&comp); err != nil {
			return fmt.Errorf("component %q: %w", key, err)
		}
		comp.ID = key
		if _, seen := m.Items[key]; !seen {
			m.Order = append(m.Order, key)
		}
		m.Items[key] = &comp
	}
	return nil
}

// List returns the components in author order.
func (m ComponentMap) List() []*Component {
	out := make([]*Component, 0, len(m.Order))
	for _, k := range m.Order {
		out = append(out, m.Items[k])
	}
	return out
}

// Get returns a component by id.
func (m ComponentMap) Get(id string) (*Component, bool) {
	c, ok := m.Items[id]
	return c, ok
}

// AttributeMap is an ordered, id-keyed collection of attribute groups.
type AttributeMap struct {
	Order []string
	Items map[string]*AttributeGroup
}

// UnmarshalYAML decodes a mapping node into an ordered AttributeMap.
func (m *AttributeMap) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping for attribute map, got kind %d", n.Kind)
	}
	if m.Items == nil {
		m.Items = map[string]*AttributeGroup{}
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		key := n.Content[i].Value
		var g AttributeGroup
		if err := n.Content[i+1].Decode(&g); err != nil {
			return fmt.Errorf("attribute group %q: %w", key, err)
		}
		g.ID = key
		if _, seen := m.Items[key]; !seen {
			m.Order = append(m.Order, key)
		}
		m.Items[key] = &g
	}
	return nil
}

// List returns the attribute groups in author order.
func (m AttributeMap) List() []*AttributeGroup {
	out := make([]*AttributeGroup, 0, len(m.Order))
	for _, k := range m.Order {
		out = append(out, m.Items[k])
	}
	return out
}

// TraitMap is an ordered, category-keyed collection of trait lists.
type TraitMap struct {
	Order []string
	Items map[string][]string
}

// UnmarshalYAML decodes a mapping node into an ordered TraitMap.
func (m *TraitMap) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping for trait map, got kind %d", n.Kind)
	}
	if m.Items == nil {
		m.Items = map[string][]string{}
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		key := n.Content[i].Value
		var traits []string
		if err := n.Content[i+1].Decode(&traits); err != nil {
			return fmt.Errorf("trait group %q: %w", key, err)
		}
		if _, seen := m.Items[key]; !seen {
			m.Order = append(m.Order, key)
		}
		m.Items[key] = traits
	}
	return nil
}

// TraitGroup is an ordered view of one trait category.
type TraitGroup struct {
	ID     string
	Label  string
	Traits []string
}

// List returns the trait categories as ordered groups.
func (m TraitMap) List() []TraitGroup {
	out := make([]TraitGroup, 0, len(m.Order))
	for _, k := range m.Order {
		out = append(out, TraitGroup{ID: k, Label: titleCase(k), Traits: m.Items[k]})
	}
	return out
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	return string(s[0]-32) + s[1:]
}
