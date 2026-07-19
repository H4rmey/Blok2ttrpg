package models

// TraitCategory groups a trait by which "track" it belongs to.
type TraitCategory string

const (
	TraitCategoryGeneral TraitCategory = "general"
	TraitCategoryOffense TraitCategory = "offense"
	TraitCategoryDefense TraitCategory = "defense"
	TraitCategoryVital   TraitCategory = "vital"
)

// EngageMode selects the engagement roll strategy.
type EngageMode string

const (
	EngageModeTrait    EngageMode = "trait"
	EngageModeGeneric  EngageMode = "generic"
	EngageModeOther    EngageMode = "other"
	EngageModePrevious EngageMode = "previous"
)

// CounterType selects how counter rolls are picked.
type CounterType string

const (
	CounterTypeDefenseTrait CounterType = "defense" // default — picks from defense traits
	CounterTypeGeneralTrait CounterType = "general"
	CounterTypeOther        CounterType = "other"
	CounterTypePrevious     CounterType = "previous"
)

// CounterRoll records a single counter-roll selection plus its category, so a
// counter roll can be a defense trait (default), a general trait (costs extra),
// or "use result of previous" (costs extra).
type CounterRoll struct {
	Type  TraitCategory `json:"type" yaml:"-"`
	Trait string        `json:"trait" yaml:"trait"`
	Other string        `json:"other,omitempty" yaml:"-"`
}

// Validation determines whether an enactment successfully affects its targets.
type Validation struct {
	BuildCost           int                    `json:"build_cost,omitempty" yaml:"build_cost,omitempty"`
	CastCost            int                    `json:"cast_cost,omitempty" yaml:"cast_cost,omitempty"`
	Fields              map[string]interface{} `json:"fields,omitempty" yaml:"-"`
	EngageMode          EngageMode             `json:"engage_mode,omitempty" yaml:"engage_mode,omitempty"`
	EngageTrait         string                 `json:"engage_trait,omitempty" yaml:"engage_trait,omitempty"`
	EngageTraitCategory TraitCategory          `json:"engage_trait_category,omitempty" yaml:"-"`
	EngageDie           string                 `json:"engage_die,omitempty" yaml:"engage_die,omitempty"`
	EngageOther         string                 `json:"engage_other,omitempty" yaml:"-"`
	// CounterRolls is kept flat for backward compatibility. New code uses
	// CounterRollEntries for typed counter rolls.
	CounterRolls       []string      `json:"counter_rolls,omitempty" yaml:"counter_rolls,omitempty"`
	CounterRollEntries []CounterRoll `json:"counter_roll_entries,omitempty" yaml:"-"`
	CounterDefaultType CounterType   `json:"counter_default_type,omitempty" yaml:"-"`
}

// TotalCost sums validation build + cast costs.
func (v *Validation) TotalCost() int { return v.BuildCost + v.CastCost }
