package models

// AbilityType represents the type of ability.
type AbilityType string

const (
	AbilityExecution AbilityType = "Execution"
	AbilityReaction  AbilityType = "Reaction"
	AbilityPhase     AbilityType = "Phase"
	AbilityMinion    AbilityType = "Minion"
)

// AllAbilityTypes lists all available ability types.
var AllAbilityTypes = []AbilityType{
	AbilityExecution,
	AbilityReaction,
	AbilityPhase,
	AbilityMinion,
}

// Ability represents a complete ability built from components.
type Ability struct {
	ID          string      `json:"id" yaml:"-"`
	Name        string      `json:"name" yaml:"name,omitempty"`
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
	Type        AbilityType `json:"type" yaml:"type"`
	BuildCost   int         `json:"build_cost,omitempty" yaml:"build_cost,omitempty"`

	// Common
	HasItemDependency bool   `json:"has_item_dependency,omitempty" yaml:"has_item_dependency,omitempty"`
	ItemName          string `json:"item_name,omitempty" yaml:"item_name,omitempty"`

	// Execution (computed display)
	EnergyCost int `json:"energy_cost" yaml:"energy_cost"`
	ActionCost int `json:"action_cost" yaml:"action_cost"`

	// Execution adjustment (form fields)
	EnergySteps int `json:"energy_steps,omitempty" yaml:"energy_steps,omitempty"`
	ActionSteps int `json:"action_steps,omitempty" yaml:"action_steps,omitempty"`

	// Reaction fields
	ReactionRange int    `json:"reaction_range,omitempty" yaml:"-"`
	ReactionUses  int    `json:"reaction_uses,omitempty" yaml:"-"`
	Trigger       string `json:"trigger,omitempty" yaml:"trigger,omitempty"`
	TriggerTrait  string `json:"trigger_trait,omitempty" yaml:"trigger_trait,omitempty"`

	// Phase fields
	PhaseDuration      int      `json:"phase_duration,omitempty" yaml:"-"`
	ReversePhaseRounds int      `json:"reverse_phase_rounds,omitempty" yaml:"-"`
	AllKnockoutsReq    bool     `json:"all_knockouts_req,omitempty" yaml:"-"`
	ReverseKnockoutOK  bool     `json:"reverse_knockout_ok,omitempty" yaml:"-"`
	NoKnockout         bool     `json:"no_knockout,omitempty" yaml:"-"`
	Knockouts          []string `json:"knockouts,omitempty" yaml:"knockouts,omitempty"`

	// Minion fields
	HPBonus       int `json:"hp_bonus,omitempty" yaml:"-"`
	ExtraLifetime int `json:"extra_lifetime,omitempty" yaml:"-"`

	Enactments []Enactment `json:"enactments" yaml:"enactments"`
}

// TotalCost sums build costs and adds +1 for each additional enactment.
func (a *Ability) TotalCost() int {
	return a.TotalBuildCost()
}

func (a *Ability) TotalBuildCost() int {
	total := a.BuildCost
	for i, e := range a.Enactments {
		if i > 0 {
			total++
		}
		total += e.BuildCost
		if e.Interaction != nil {
			total += e.Interaction.BuildCost
			if e.Interaction.Validation != nil {
				total += e.Interaction.Validation.BuildCost
			}
		}
	}
	return total
}

func (a *Ability) TotalEnergyCost(additionalEnergyCost int) int {
	total := a.EnergyCost
	for i, e := range a.Enactments {
		if i > 0 {
			total += additionalEnergyCost
		}
		total += e.CastCost
		if e.Interaction != nil {
			total += e.Interaction.CastCost
			if e.Interaction.Validation != nil {
				total += e.Interaction.Validation.CastCost
			}
		}
	}
	return total
}

// EnactmentSummary returns a short description of the enactments.
func (a *Ability) EnactmentSummary() string {
	if len(a.Enactments) == 0 {
		return "None"
	}
	summary := ""
	for i, e := range a.Enactments {
		if i > 0 {
			summary += " → "
		}
		summary += string(e.Type)
	}
	return summary
}
