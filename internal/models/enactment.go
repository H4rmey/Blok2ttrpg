package models

// EnactmentType represents the type of enactment.
type EnactmentType string

const (
	EnactDamage           EnactmentType = "Enact Damage"
	EnactHealing          EnactmentType = "Enact Healing"
	EnactMovement         EnactmentType = "Enact Movement"
	EnactProficiencyShift EnactmentType = "Enact Proficiency Shift"
	EnactPersistentEffect EnactmentType = "Enact Persistent Effect"
	EnactNegation         EnactmentType = "Enact Negation"
	EnactState            EnactmentType = "Enact State"
)

// AllEnactmentTypes lists all available enactment types.
var AllEnactmentTypes = []EnactmentType{
	EnactDamage,
	EnactHealing,
	EnactMovement,
	EnactProficiencyShift,
	EnactPersistentEffect,
	EnactNegation,
	EnactState,
}

// Enactment represents a single effect within an ability.
type Enactment struct {
	Name        string        `json:"name,omitempty" yaml:"name,omitempty"`
	Description string        `json:"description,omitempty" yaml:"description,omitempty"`
	Type        EnactmentType `json:"type" yaml:"type"`
	BuildCost   int           `json:"build_cost,omitempty" yaml:"build_cost,omitempty"`
	CastCost    int           `json:"cast_cost,omitempty" yaml:"cast_cost,omitempty"`
	Formula     string        `json:"formula,omitempty" yaml:"formula,omitempty"`

	// Fields is the generic field-values map driven by the enactment schema.
	// Values can be string, bool, int, or []string.
	Fields map[string]interface{} `json:"fields,omitempty" yaml:"-"`

	// Common - kept for backward compatibility with hydration logic
	Always bool `json:"always,omitempty" yaml:"always,omitempty"`

	// Damage / Healing fields - kept for backward compatibility
	Source         string        `json:"source,omitempty" yaml:"source,omitempty"` // d4..d12, "trait", "general", "other"
	SourceTrait    string        `json:"source_trait,omitempty" yaml:"source_trait,omitempty"`
	SourceCategory TraitCategory `json:"source_category,omitempty" yaml:"-"`
	OtherRollText  string        `json:"other_roll_text,omitempty" yaml:"other_roll_text,omitempty"`
	FlatBonus      int           `json:"flat_bonus,omitempty" yaml:"flat_bonus,omitempty"`

	// Damage-only
	OffensiveTrait string `json:"offensive_trait,omitempty" yaml:"offensive_trait,omitempty"`

	// Healing-only
	MedicineTrait string `json:"medicine_trait,omitempty" yaml:"medicine_trait,omitempty"`

	// Movement fields
	OriginMode string   `json:"origin_mode,omitempty" yaml:"origin_mode,omitempty"` // engager | other
	OriginText string   `json:"origin_text,omitempty" yaml:"origin_text,omitempty"`
	Distance   int      `json:"distance,omitempty" yaml:"distance,omitempty"`
	Directions []string `json:"directions,omitempty" yaml:"directions,omitempty"`

	// Proficiency Shift fields
	ShiftedTrait string `json:"shifted_trait,omitempty" yaml:"shifted_trait,omitempty"`
	ShiftDir     string `json:"shift_dir,omitempty" yaml:"shift_dir,omitempty"`
	ShiftAmount  int    `json:"shift_amount,omitempty" yaml:"shift_amount,omitempty"`
	ShiftUses    int    `json:"shift_uses,omitempty" yaml:"shift_uses,omitempty"`

	// Persistent Effect fields
	EffectName    string   `json:"effect_name,omitempty" yaml:"effect_name,omitempty"`
	EffectType    string   `json:"effect_type,omitempty" yaml:"effect_type,omitempty"`
	Duration      int      `json:"duration,omitempty" yaml:"duration,omitempty"`
	TriggerTiming string   `json:"trigger_timing,omitempty" yaml:"trigger_timing,omitempty"`
	Solutions     []string `json:"solutions,omitempty" yaml:"solutions,omitempty"`

	Interaction *Interaction `json:"interaction,omitempty" yaml:"interactions,omitempty"`
}

// TotalCost aggregates the enactment's own build/cast plus its interaction and
// validation. The live builder stores these on the card via data-build /
// data-cast attributes, but having a server-side fallback is useful when
// re-displaying a saved ability.
func (e *Enactment) TotalCost() int {
	total := e.BuildCost + e.CastCost
	if e.Interaction != nil {
		total += e.Interaction.TotalCost()
	}
	return total
}
