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

	// OptionSources holds named static option lists so any field can reference
	// them via options_source without hardcoding the list in Go. Each entry is
	// a plain string list; the value and the label are the same string.
	OptionSources map[string][]string `yaml:"option_sources,omitempty" json:"option_sources,omitempty"`

	// OptionSourcesCosted holds named option lists whose entries each carry
	// their own cost. This backs per-trigger build costs: a source referenced
	// via options_source can have distinct build/energy costs per entry. When a
	// source name exists in both maps, the costed variant takes precedence.
	OptionSourcesCosted map[string][]Option `yaml:"option_sources_costed,omitempty" json:"option_sources_costed,omitempty"`

	// TraitCategories lists the trait group ids that make up the "traits_all"
	// option source and its grouped display. When empty the app falls back to
	// the historical general/offense/defense set.
	TraitCategories []string `yaml:"trait_categories,omitempty" json:"trait_categories,omitempty"`

	// Character attributes and traits are fully config-driven, keyed by id.
	Attributes AttributeMap `yaml:"attributes,omitempty" json:"attributes,omitempty"`
	Traits     TraitMap     `yaml:"traits,omitempty" json:"traits,omitempty"`

	// Proficiency tiers referenced by traits.
	Proficiencies []Proficiency `yaml:"proficiencies,omitempty" json:"proficiencies,omitempty"`

	// Leveling budgets, given as per-level tables.
	Leveling Leveling `yaml:"leveling,omitempty" json:"leveling,omitempty"`

	// Ability building blocks, keyed by id but with author ordering preserved.
	AbilityTypes ComponentMap `yaml:"ability_types,omitempty" json:"ability_types,omitempty"`
	Enactments   ComponentMap `yaml:"enactments,omitempty" json:"enactments,omitempty"`
	Interactions ComponentMap `yaml:"interactions,omitempty" json:"interactions,omitempty"`

	// States for the "Enact State" enactment.
	AdditionalState Cost            `yaml:"additional_state,omitempty" json:"additional_state,omitempty"`
	GeneralStates   []GeneralState  `yaml:"general_states,omitempty" json:"general_states,omitempty"`
	SpecificStates  []SpecificState `yaml:"specific_states,omitempty" json:"specific_states,omitempty"`

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
}

// AsCost converts the surcharge into a plain Cost.
func (a AdditionalEnactment) AsCost() Cost {
	return Cost{BuildCost: a.BuildCost, EnergyCost: a.EnergyCost}
}

// Dice lists the die tiers available for damage and generic rolls.
type Dice struct {
	Damage  []string `yaml:"damage,omitempty" json:"damage,omitempty"`
	Generic []string `yaml:"generic,omitempty" json:"generic,omitempty"`
}

// Validations captures the engagement/counter configuration and its fields.
type Validations struct {
	Fields []Field `yaml:"fields,omitempty" json:"fields,omitempty"`
}

// Proficiency is a single skill tier.
type Proficiency struct {
	ID     string            `yaml:"id" json:"id"`
	Name   string            `yaml:"name" json:"name"`
	Cost   int               `yaml:"cost" json:"cost"`
	Note   string            `yaml:"note,omitempty" json:"note,omitempty"`
	Dice   map[string]string `yaml:"dice,omitempty" json:"dice,omitempty"`
	Vitals map[string]any    `yaml:"vitals,omitempty" json:"vitals,omitempty"`
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

// GeneralState is a shiftable state applied via the state enactment.
type GeneralState struct {
	ID          string `yaml:"id" json:"id"`
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	MinShift    int    `yaml:"min_shift,omitempty" json:"min_shift,omitempty"`
	MaxShift    int    `yaml:"max_shift,omitempty" json:"max_shift,omitempty"`
	ShiftCost   Cost   `yaml:"shift_cost,omitempty" json:"shift_cost,omitempty"`
}

// SpecificState is a fixed-cost named condition.
type SpecificState struct {
	ID          string `yaml:"id" json:"id"`
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	BuildCost   int    `yaml:"build_cost,omitempty" json:"build_cost,omitempty"`
	EnergyCost  int    `yaml:"energy_cost,omitempty" json:"energy_cost,omitempty"`
}

// Component is a generic ability building block: an ability type, enactment or
// interaction. Legacy authoring blocks (step_costs, perks, triggers, etc.) are
// intentionally not modelled here; they are ignored on load.
type Component struct {
	ID          string `yaml:"-" json:"id"`
	Name        string `yaml:"name,omitempty" json:"name,omitempty"`
	Type        string `yaml:"type,omitempty" json:"type,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	BaseCost    Cost   `yaml:"base_cost,omitempty" json:"base_cost,omitempty"`
	BaseEnergy  int    `yaml:"base_energy,omitempty" json:"base_energy,omitempty"`
	BaseAction  int    `yaml:"base_action,omitempty" json:"base_action,omitempty"`

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
	Key         string `yaml:"key" json:"key"`
	Label       string `yaml:"label" json:"label"`
	Type        string `yaml:"type" json:"type"` // checkbox, dropdown, free_text, free_number, solutions, states
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

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

	// solutions/states: a repeatable set of rows built from RowFields. PerItem
	// is the cost delta per row relative to DefaultCount.
	RowFields    []Field  `yaml:"row_fields,omitempty" json:"row_fields,omitempty"`
	DefaultCount int      `yaml:"default_count,omitempty" json:"default_count,omitempty"`
	PerItem      *PerStep `yaml:"per_item,omitempty" json:"per_item,omitempty"`
	// RowDefaults pre-fills the initial rows of a solutions/states field. Each
	// entry is a map of row_field key -> default value for that row, applied in
	// order to the first rows rendered.
	RowDefaults []map[string]string `yaml:"row_defaults,omitempty" json:"row_defaults,omitempty"`

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
	Value  string  `yaml:"value" json:"value"`
	Label  string  `yaml:"label,omitempty" json:"label,omitempty"`
	Cost   *Cost   `yaml:"cost,omitempty" json:"cost,omitempty"`
	Fields []Field `yaml:"fields,omitempty" json:"fields,omitempty"`
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
