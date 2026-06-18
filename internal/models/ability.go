package models

// AbilityType represents the type of ability.
type AbilityType string

const (
	AbilityTypeExecution AbilityType = "Execution"
	AbilityTypeReaction  AbilityType = "Reaction"
	AbilityTypePhase     AbilityType = "Phase"
	AbilityTypeMinion    AbilityType = "Minion"
)

// AllAbilityTypes returns all valid ability types.
func AllAbilityTypes() []AbilityType {
	return []AbilityType{AbilityTypeExecution, AbilityTypeReaction, AbilityTypePhase, AbilityTypeMinion}
}

// EnactmentType represents the type of enactment.
type EnactmentType string

const (
	EnactmentDamage           EnactmentType = "Enact Damage"
	EnactmentHealing          EnactmentType = "Enact Healing"
	EnactmentMovement         EnactmentType = "Enact Movement"
	EnactmentProficiencyShift EnactmentType = "Enact Proficiency Shift"
	EnactmentPersistentEffect EnactmentType = "Enact Persistent Effect"
)

// AllEnactmentTypes returns all enactment types.
func AllEnactmentTypes() []EnactmentType {
	return []EnactmentType{
		EnactmentDamage, EnactmentHealing, EnactmentMovement,
		EnactmentProficiencyShift, EnactmentPersistentEffect,
	}
}

// InteractionType represents how an enactment is applied.
type InteractionType string

const (
	InteractionSelf         InteractionType = "Self"
	InteractionDirect       InteractionType = "Direct"
	InteractionRanged       InteractionType = "Ranged"
	InteractionArea         InteractionType = "Area"
	InteractionAreaOfEffect InteractionType = "Area of Effect"
)

// AllInteractionTypes returns all interaction types.
func AllInteractionTypes() []InteractionType {
	return []InteractionType{
		InteractionSelf, InteractionDirect, InteractionRanged,
		InteractionArea, InteractionAreaOfEffect,
	}
}

// Perk represents a modifier that can be applied to abilities, enactments, interactions, or validations.
type Perk struct {
	Description  string `yaml:"description"`
	AddCost      int    `yaml:"add_cost"`
	Amount       int    `yaml:"amount"`
	TotalAddCost int    `yaml:"total_add_cost"`
	EnergyCost   int    `yaml:"energy_cost"`
	IsOptional   bool   `yaml:"is_optional"`
}

// Validation defines whether an enactment succeeds against its targets.
type Validation struct {
	EngagementRoll string   `yaml:"engagement_roll"`
	CounterRoll    []string `yaml:"counter_roll"`
	Perks          []Perk   `yaml:"perks,omitempty"`
}

// Interaction defines how an enactment is applied in the game world.
type Interaction struct {
	Type         InteractionType `yaml:"type"`
	Engager      string          `yaml:"engager"`
	TargetAmount int             `yaml:"target_amount,omitempty"`
	Range        string          `yaml:"range,omitempty"`
	Radius       string          `yaml:"radius,omitempty"`
	Origin       string          `yaml:"origin,omitempty"`
	Duration     string          `yaml:"duration,omitempty"`
	Visibility   string          `yaml:"visibility,omitempty"`
	Obstruction  string          `yaml:"obstruction,omitempty"`
	Immunity     bool            `yaml:"immunity,omitempty"`
	Perks        []Perk          `yaml:"perks,omitempty"`
	Validation   *Validation     `yaml:"validation,omitempty"`
}

// PersistentEffect defines the effects within a Persistent Enactment.
type PersistentEffect struct {
	Type                   EnactmentType `yaml:"type"`
	DamageDice             string        `yaml:"damage_dice,omitempty"`
	HealingDice            string        `yaml:"healing_dice,omitempty"`
	IsOptional             bool          `yaml:"is_optional"`
	BaseEnactmentEnergyCost int          `yaml:"base_enactment_energy_cost"`
	Perks                  []Perk        `yaml:"perks,omitempty"`
}

// Enactment represents a specific action within an ability.
type Enactment struct {
	Type                    EnactmentType `yaml:"type"`
	IsOptional              bool          `yaml:"is_optional"`
	BaseEnactmentEnergyCost int           `yaml:"base_enactment_energy_cost"`

	// Enact Damage fields
	DamageDice string `yaml:"damage_dice,omitempty"`

	// Enact Healing fields
	HealingDice string `yaml:"healing_dice,omitempty"`

	// Enact Movement fields
	MinimalDistance  string   `yaml:"minimal_distance,omitempty"`
	Origin          string   `yaml:"origin,omitempty"`
	DirectionOptions []string `yaml:"direction_options,omitempty"`

	// Enact Proficiency Shift fields
	ShiftedTrait   string `yaml:"shifted_trait,omitempty"`
	ShiftDirection string `yaml:"shift_direction,omitempty"`
	ShiftAmount    int    `yaml:"shift_amount,omitempty"`
	ShiftUses      int    `yaml:"shift_uses,omitempty"`

	// Enact Persistent Effect fields
	Duration        string             `yaml:"duration,omitempty"`
	TriggerTiming   string             `yaml:"trigger_timing,omitempty"`
	Solutions       []string           `yaml:"solutions,omitempty"`
	EffectFlavor    string             `yaml:"effect_flavor,omitempty"` // e.g. "Fire", "Ice", "Poison"
	Effects         []PersistentEffect `yaml:"effects,omitempty"`

	Perks        []Perk        `yaml:"perks,omitempty"`
	Interactions []Interaction `yaml:"interactions,omitempty"`
}

// Trigger represents a reaction trigger condition.
type Trigger struct {
	Name string `yaml:"name"`
}

// KnockoutRequirement represents a phase knockout condition.
type KnockoutRequirement struct {
	Name string `yaml:"name"`
}

// MinionStats holds the default stats for a minion ability.
type MinionStats struct {
	Health   int    `yaml:"health"`
	Attack   string `yaml:"attack"`
	Defense  string `yaml:"defense"`
	Speed    string `yaml:"speed"`
	Lifetime int    `yaml:"lifetime"`
}

// Ability represents a complete ability built using the Ability Builder.
type Ability struct {
	Name       string      `yaml:"name"`
	Type       AbilityType `yaml:"type"`
	EnergyCost int         `yaml:"energy_cost"`
	ActionCost int         `yaml:"action_cost,omitempty"`

	// Execution/Reaction/Phase common
	HasItemDependency bool   `yaml:"has_item_dependency"`
	ItemDependency    string `yaml:"item_dependency,omitempty"`

	// Reaction-specific
	Range    int       `yaml:"range,omitempty"`
	Uses     int       `yaml:"uses,omitempty"`
	Triggers []Trigger `yaml:"triggers,omitempty"`

	// Phase-specific
	PhaseDuration        int                   `yaml:"phase_duration,omitempty"`
	KnockoutRequirements []KnockoutRequirement `yaml:"knockout_requirements,omitempty"`
	BadEnactments        []Enactment           `yaml:"bad_enactments,omitempty"`

	// Minion-specific
	MinionStats *MinionStats `yaml:"minion_stats,omitempty"`

	Enactments []Enactment `yaml:"enactments"`
	Perks      []Perk      `yaml:"perks,omitempty"`
}

// TotalAddCost calculates the total ability points spent on this ability.
func (a *Ability) TotalAddCost() int {
	total := 0

	// Ability-level perks
	for _, p := range a.Perks {
		total += p.TotalAddCost
	}

	// Enactment costs
	for _, e := range a.Enactments {
		for _, p := range e.Perks {
			total += p.TotalAddCost
		}
		// Interaction costs
		for _, i := range e.Interactions {
			for _, p := range i.Perks {
				total += p.TotalAddCost
			}
			// Validation costs
			if i.Validation != nil {
				for _, p := range i.Validation.Perks {
					total += p.TotalAddCost
				}
			}
		}
		// Persistent effect costs
		for _, eff := range e.Effects {
			for _, p := range eff.Perks {
				total += p.TotalAddCost
			}
		}
	}

	// Bad enactment costs (Phase - these are usually negative/refunds)
	for _, e := range a.BadEnactments {
		for _, p := range e.Perks {
			total += p.TotalAddCost
		}
	}

	// Trigger costs (Reaction - first is free, additional cost 2 each)
	if len(a.Triggers) > 1 {
		total += (len(a.Triggers) - 1) * 2
	}

	return total
}

// TotalEnergyCost calculates the total energy cost to use this ability.
func (a *Ability) TotalEnergyCost() int {
	total := a.EnergyCost

	// Ability-level perks
	for _, p := range a.Perks {
		total += p.EnergyCost
	}

	// Enactment costs
	for _, e := range a.Enactments {
		total += e.BaseEnactmentEnergyCost
		for _, p := range e.Perks {
			total += p.EnergyCost
		}
		for _, i := range e.Interactions {
			for _, p := range i.Perks {
				total += p.EnergyCost
			}
			if i.Validation != nil {
				for _, p := range i.Validation.Perks {
					total += p.EnergyCost
				}
			}
		}
		for _, eff := range e.Effects {
			total += eff.BaseEnactmentEnergyCost
			for _, p := range eff.Perks {
				total += p.EnergyCost
			}
		}
	}

	return total
}
