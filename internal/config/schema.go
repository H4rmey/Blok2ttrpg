// This file defines the top-level ruleset schema: the Cost pair every priced
// thing uses, the root Config struct, and the combat/energy/enactment rule
// blocks it embeds. Nothing here is special-cased by id, so a ruleset can add
// content purely in YAML.
package config

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

	// AllowNegativeBuildCost/AllowNegativeEnergyCost control whether an
	// perk's final computed cost may drop below zero. Refund-style options
	// (the energy offset, Enact Nerf, negative-cost knockouts) can otherwise
	// push a total negative. These are pointers so an unset value defaults to
	// false, clamping the corresponding total at zero.
	AllowNegativeBuildCost  *bool `yaml:"allow_negative_build_cost,omitempty" json:"allow_negative_build_cost,omitempty"`
	AllowNegativeEnergyCost *bool `yaml:"allow_negative_energy_cost,omitempty" json:"allow_negative_energy_cost,omitempty"`

	// AllowNegativeSkillPoints controls whether a character may spend more
	// skill (skill) points than its level budget grants. It defaults to false,
	// which makes the app reject any change (manual skill edit or package
	// import) that would push the used total past the budget.
	AllowNegativeSkillPoints *bool `yaml:"allow_negative_skill_points,omitempty" json:"allow_negative_skill_points,omitempty"`

	Combat Combat `yaml:"combat,omitempty" json:"combat,omitempty"`

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
	// replaces the hardcoded skills_all/roll_all/conditions_all grouping.
	OptionGroups map[string]OptionGroupDef `yaml:"option_groups,omitempty" json:"option_groups,omitempty"`

	// SkillCategories lists the skill group ids that make up the "skills_all"
	// option source and its grouped display. When empty the app falls back to
	// the historical general/offense/defense set.
	SkillCategories []string `yaml:"skill_categories,omitempty" json:"skill_categories,omitempty"`

	// VitalGroup names the skill group id whose skills (HP, Movement, Energy)
	// map to numeric vital values rather than dice. Defaults to "vital".
	VitalGroup string `yaml:"vital_group,omitempty" json:"vital_group,omitempty"`

	// Character traits and skills are fully config-driven, keyed by id.
	Traits TraitMap `yaml:"traits,omitempty" json:"traits,omitempty"`
	Skills SkillMap `yaml:"skills,omitempty" json:"skills,omitempty"`

	// Proficiency tiers referenced by skills.
	Proficiencies []Proficiency `yaml:"proficiencies,omitempty" json:"proficiencies,omitempty"`

	// DefaultProficiency names the proficiency tier id that new characters
	// start every skill at (the "free" baseline). When empty the first tier in
	// the Proficiencies list is used. Tiers below the default are free; tiers
	// above the default accrue their cumulative per-tier cost.
	DefaultProficiency string `yaml:"default_proficiency,omitempty" json:"default_proficiency,omitempty"`

	// Leveling budgets, given as per-level tables.
	Leveling Leveling `yaml:"leveling,omitempty" json:"leveling,omitempty"`

	// Invoking is the invoke point economy: what a point buys, how one is
	// earned, and the limits on reactions. The pool size per level lives in
	// Leveling.InvokePoints.
	Invoking Invoking `yaml:"invoking,omitempty" json:"invoking,omitempty"`

	// Negotiation is the structured social encounter: the motivation ladder,
	// the patience clock and the NPC trait-alignment rules. It is unrelated to
	// Interactions, which are the perk-builder's targeting components.
	Negotiation Negotiation `yaml:"negotiation,omitempty" json:"negotiation,omitempty"`

	// Passives is the predefined perk catalogue: perks written by the ruleset
	// author and picked from a list rather than assembled in the builder. See
	// passives.go for why they exist alongside the builder.
	Passives Passives `yaml:"passives,omitempty" json:"passives,omitempty"`

	// Perk building blocks, keyed by id but with author ordering preserved.
	PerkTypes    ComponentMap `yaml:"perk_types,omitempty" json:"perk_types,omitempty"`
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

// AllowsNegativeBuildCost reports whether an perk's final build cost may be
// negative. Defaults to false when unset.
func (c *Config) AllowsNegativeBuildCost() bool {
	return c.AllowNegativeBuildCost != nil && *c.AllowNegativeBuildCost
}

// AllowsNegativeEnergyCost reports whether an perk's final energy cost may
// be negative. Defaults to false when unset.
func (c *Config) AllowsNegativeEnergyCost() bool {
	return c.AllowNegativeEnergyCost != nil && *c.AllowNegativeEnergyCost
}

// AllowsNegativeSkillPoints reports whether a character is permitted to
// overspend its skill (skill) point budget. Defaults to false when unset.
func (c *Config) AllowsNegativeSkillPoints() bool {
	return c.AllowNegativeSkillPoints != nil && *c.AllowNegativeSkillPoints
}

// Combat holds combat-wide settings.

type Combat struct {
	Actions struct {
		Amount int `yaml:"amount" json:"amount"`
	} `yaml:"actions" json:"actions"`

	// EnergyRecoveryPerRest is how much Energy a character regains on a rest.
	// It is a rules parameter surfaced in the generated documentation; the app
	// does not spend or restore Energy automatically, so nothing in the engine
	// reads it.
	EnergyRecoveryPerRest int `yaml:"energy_recovery_per_rest,omitempty" json:"energy_recovery_per_rest,omitempty"`

	// EnergyOverdraft is what happens when a character uses a perk they cannot
	// pay the Energy for. It is the HP-for-Energy trade, and it is the single
	// most load-bearing number in the resource economy: set it too low and
	// Energy stops being a resource at all, because HP simply becomes a second
	// Energy pool.
	EnergyOverdraft EnergyOverdraft `yaml:"energy_overdraft,omitempty" json:"energy_overdraft,omitempty"`
}

// EnergyOverdraft configures running out of Energy mid-perk.
//
// The rule exists so that being empty is a hard decision rather than a wall: a
// perk may still be used, but it has to be paid for out of the character's own
// body, or paid for by cutting the perk short.
//
// HPPerEnergy is deliberately well above 1. HP and Energy climb at the same rate
// per proficiency rung, so a 1:1 rate makes the whole Energy pool purchasable
// with the whole HP pool at no loss, at every tier. A rate of 3 makes a single
// extra round cost most of a low-tier character's health, which is what keeps
// the overdraft an emergency button instead of a routine optimisation.
type EnergyOverdraft struct {
	// HPPerEnergy is the HP paid for each point of missing Energy.
	HPPerEnergy int `yaml:"hp_per_energy,omitempty" json:"hp_per_energy,omitempty"`

	// EscalationPerUse is added to HPPerEnergy on each subsequent overdraft in
	// the same scene, mirroring how repeated movement gets progressively more
	// expensive. Zero keeps the rate flat, which is simpler at the table.
	EscalationPerUse int `yaml:"escalation_per_use,omitempty" json:"escalation_per_use,omitempty"`

	// ConditionOnOverdraft names a condition applied when a character
	// overdraws, or "" for none. Pointing it at a condition that raises energy
	// costs makes the rule self-limiting without any per-scene bookkeeping.
	ConditionOnOverdraft string `yaml:"condition_on_overdraft,omitempty" json:"condition_on_overdraft,omitempty"`

	// AllowPartialExecution permits paying nothing and instead dropping
	// enactments the character cannot afford. It is a pointer so an unset value
	// defaults to true: it is the more interesting of the two choices, because
	// it degrades the perk rather than the character.
	AllowPartialExecution *bool `yaml:"allow_partial_execution,omitempty" json:"allow_partial_execution,omitempty"`
}

// AllowsPartialExecution reports whether an unaffordable perk may be used at
// reduced effect instead of paying HP. Defaults to true when unset.
func (e EnergyOverdraft) AllowsPartialExecution() bool {
	return e.AllowPartialExecution == nil || *e.AllowPartialExecution
}

// OverdraftRateForUse returns the HP-per-Energy rate for the nth overdraft in a
// scene, where n is 1-based. With EscalationPerUse at zero the rate is flat.
func (e EnergyOverdraft) OverdraftRateForUse(n int) int {
	if n < 1 {
		n = 1
	}
	return e.HPPerEnergy + e.EscalationPerUse*(n-1)
}

// EnergyOverdraftConfigured reports whether an energy_overdraft block was
// declared. It cannot be a struct comparison because the type holds a pointer.
func EnergyOverdraftConfigured(e EnergyOverdraft) bool {
	return e.HPPerEnergy != 0 ||
		e.EscalationPerUse != 0 ||
		e.ConditionOnOverdraft != "" ||
		e.AllowPartialExecution != nil
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

	// NewTarget is the surcharge for letting an enactment beyond the first pick
	// its own target instead of inheriting the previous enactment's target. An
	// enactment that opts in gets its own Interaction and Validation region and
	// pays for them; one that does not inherit both for free.
	NewTarget NewTargetOption `yaml:"new_target,omitempty" json:"new_target,omitempty"`
}

// NewTargetOption configures the "different target than the previous enactment"
// opt-in: what it costs and how it is labelled in the builder.
type NewTargetOption struct {
	BuildCost   int    `yaml:"build_cost,omitempty" json:"build_cost,omitempty"`
	EnergyCost  int    `yaml:"energy_cost,omitempty" json:"energy_cost,omitempty"`
	Label       string `yaml:"label,omitempty" json:"label,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// AsCost converts the surcharge into a plain Cost.
func (n NewTargetOption) AsCost() Cost {
	return Cost{BuildCost: n.BuildCost, EnergyCost: n.EnergyCost}
}

// DisplayLabel returns the checkbox label, falling back to a sensible default
// when the ruleset does not name it.
func (n NewTargetOption) DisplayLabel() string {
	if n.Label != "" {
		return n.Label
	}
	return "Targets something different than the previous enactment"
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
