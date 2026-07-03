package models

// InteractionType represents how an enactment is applied.
type InteractionType string

const (
	InteractionSelf         InteractionType = "Self"
	InteractionDirect       InteractionType = "Direct"
	InteractionRanged       InteractionType = "Ranged"
	InteractionArea         InteractionType = "Area"
	InteractionAreaOfEffect InteractionType = "Area of Effect"
)

// AllInteractionTypes lists all available interaction types.
var AllInteractionTypes = []InteractionType{
	InteractionSelf,
	InteractionDirect,
	InteractionRanged,
	InteractionArea,
	InteractionAreaOfEffect,
}

// Interaction determines how an enactment is applied in the game world.
type Interaction struct {
	Type      InteractionType `json:"type" yaml:"type"`
	BuildCost int             `json:"build_cost,omitempty" yaml:"build_cost,omitempty"`
	CastCost  int             `json:"cast_cost,omitempty" yaml:"cast_cost,omitempty"`

	// Range / Targets (Direct, Ranged)
	Range   int `json:"range,omitempty" yaml:"range,omitempty"`
	Targets int `json:"targets,omitempty" yaml:"targets,omitempty"`

	// Ranged-specific
	VisibleOK     bool `json:"visible_ok,omitempty" yaml:"target_may_be_visible,omitempty"`
	ObstructedOK  bool `json:"obstructed_ok,omitempty" yaml:"target_may_be_obstructed,omitempty"`
	RemovePenalty bool `json:"remove_penalty,omitempty" yaml:"remove_penalty,omitempty"`

	// Area / AoE-specific
	Radius     int    `json:"radius,omitempty" yaml:"radius,omitempty"`
	OriginMode string `json:"origin_mode,omitempty" yaml:"origin_mode,omitempty"`
	OriginText string `json:"origin_text,omitempty" yaml:"origin,omitempty"`

	// AoE-specific
	Duration int    `json:"duration,omitempty" yaml:"duration,omitempty"`
	Timing   string `json:"timing,omitempty" yaml:"timing,omitempty"`
	Immune   bool   `json:"immune,omitempty" yaml:"immune,omitempty"`

	// Use result of previous interaction/validation
	UsePrevious bool `json:"use_previous,omitempty" yaml:"use_previous,omitempty"`

	Validation *Validation `json:"validation,omitempty" yaml:"validation,omitempty"`
}

// TotalCost aggregates interaction build/cast plus its validation.
func (i *Interaction) TotalCost() int {
	total := i.BuildCost + i.CastCost
	if i.Validation != nil {
		total += i.Validation.TotalCost()
	}
	return total
}
