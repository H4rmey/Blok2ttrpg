package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the top-level ability-builder configuration loaded from YAML.
type Config struct {
	Version        string               `json:"version" yaml:"version"`
	AbilityBuilder AbilityBuilderConfig `json:"ability_builder" yaml:"ability_builder"`
}

// AbilityBuilderConfig holds the main ability-builder rules and costs.
type AbilityBuilderConfig struct {
	AdditionalEnactment CostDefinition              `json:"additional_enactment" yaml:"additional_enactment"`
	AbilityTypes        map[string]AbilityTypeConfig `json:"ability_types" yaml:"ability_types"`
	Enactments          map[string]EnactmentConfig   `json:"enactments" yaml:"enactments"`
	Interactions        map[string]InteractionConfig `json:"interactions" yaml:"interactions"`
	Validations         ValidationConfig             `json:"validations" yaml:"validations"`
	Traits              TraitConfig                  `json:"traits" yaml:"traits"`
	Dice                DiceConfig                   `json:"dice" yaml:"dice"`
}

// CostDefinition is a simple add/energy cost pair.
type CostDefinition struct {
	AddCost     int    `json:"add_cost" yaml:"add_cost"`
	EnergyCost  int    `json:"energy_cost" yaml:"energy_cost"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Step        int    `json:"step,omitempty" yaml:"step,omitempty"`
}

// StepCosts captures the cost of increasing or decreasing a stepped value.
type StepCosts struct {
	Increase CostDefinition `json:"increase" yaml:"increase"`
	Decrease CostDefinition `json:"decrease" yaml:"decrease"`
}

// AbilityTypeConfig defines an ability type (Execution, Reaction, etc.).
type AbilityTypeConfig struct {
	Name                  string               `json:"name" yaml:"name"`
	Description           string               `json:"description" yaml:"description"`
	BaseEnergy            int                  `json:"base_energy" yaml:"base_energy"`
	BaseAction            int                  `json:"base_action,omitempty" yaml:"base_action,omitempty"`
	BaseRange             int                  `json:"base_range,omitempty" yaml:"base_range,omitempty"`
	BaseUses              int                  `json:"base_uses,omitempty" yaml:"base_uses,omitempty"`
	BaseDuration          int                  `json:"base_duration,omitempty" yaml:"base_duration,omitempty"`
	BaseReverseDuration   int                  `json:"base_reverse_duration,omitempty" yaml:"base_reverse_duration,omitempty"`
	BaseHealth            int                  `json:"base_health,omitempty" yaml:"base_health,omitempty"`
	BaseLifetime          int                  `json:"base_lifetime,omitempty" yaml:"base_lifetime,omitempty"`
	RangeCost             CostDefinition       `json:"range_cost,omitempty" yaml:"range_cost,omitempty"`
	UsesCost              CostDefinition       `json:"uses_cost,omitempty" yaml:"uses_cost,omitempty"`
	DurationCost          CostDefinition       `json:"duration_cost,omitempty" yaml:"duration_cost,omitempty"`
	ReverseDurationRefund CostDefinition       `json:"reverse_duration_refund,omitempty" yaml:"reverse_duration_refund,omitempty"`
	HealthBonusCost       CostDefinition       `json:"health_bonus_cost,omitempty" yaml:"health_bonus_cost,omitempty"`
	LifetimeBonusCost     CostDefinition       `json:"lifetime_bonus_cost,omitempty" yaml:"lifetime_bonus_cost,omitempty"`
	StepCosts             map[string]StepCosts `json:"step_costs,omitempty" yaml:"step_costs,omitempty"`
	Perks                 []PerkConfig         `json:"perks" yaml:"perks"`
	Triggers              []TriggerConfig      `json:"triggers,omitempty" yaml:"triggers,omitempty"`
	KnockoutRequirements  []KnockoutConfig     `json:"knockout_requirements,omitempty" yaml:"knockout_requirements,omitempty"`
	CompatibleEnactments  []string             `json:"compatible_enactments" yaml:"compatible_enactments"`
	BadEnactments         []BadEnactmentConfig `json:"bad_enactments,omitempty" yaml:"bad_enactments,omitempty"`
}

// PerkConfig is a single purchasable perk or rule option.
type PerkConfig struct {
	ID          string `json:"id" yaml:"id"`
	Description string `json:"description" yaml:"description"`
	AddCost     int    `json:"add_cost" yaml:"add_cost"`
	EnergyCost  int    `json:"energy_cost" yaml:"energy_cost"`
}

// TriggerConfig is a Reaction trigger condition.
type TriggerConfig struct {
	ID          string `json:"id" yaml:"id"`
	Description string `json:"description" yaml:"description"`
	AddCost     int    `json:"add_cost" yaml:"add_cost"`
	EnergyCost  int    `json:"energy_cost" yaml:"energy_cost"`
}

// KnockoutConfig is a Phase knockout requirement.
type KnockoutConfig struct {
	ID          string `json:"id" yaml:"id"`
	Description string `json:"description" yaml:"description"`
	AddCost     int    `json:"add_cost" yaml:"add_cost"`
	EnergyCost  int    `json:"energy_cost" yaml:"energy_cost"`
}

// BadEnactmentConfig defines a reverse-phase bad enactment option.
type BadEnactmentConfig struct {
	ID          string `json:"id" yaml:"id"`
	Description string `json:"description" yaml:"description"`
	AddCost     int    `json:"add_cost" yaml:"add_cost"`
	EnergyCost  int    `json:"energy_cost" yaml:"energy_cost"`
}

// EnactmentConfig defines an Enactment type and its costs.
type EnactmentConfig struct {
	Type               string         `json:"type" yaml:"type"`
	Description        string         `json:"description" yaml:"description"`
	DefaultDice        string         `json:"default_dice,omitempty" yaml:"default_dice,omitempty"`
	DefaultDistance    int            `json:"default_distance,omitempty" yaml:"default_distance,omitempty"`
	DefaultShiftAmount int            `json:"default_shift_amount,omitempty" yaml:"default_shift_amount,omitempty"`
	DefaultUses        int            `json:"default_uses,omitempty" yaml:"default_uses,omitempty"`
	DefaultDuration    int            `json:"default_duration,omitempty" yaml:"default_duration,omitempty"`
	DiceTiers          map[string]int `json:"dice_tiers,omitempty" yaml:"dice_tiers,omitempty"`
	DiceTierCost       CostDefinition `json:"dice_tier_cost,omitempty" yaml:"dice_tier_cost,omitempty"`
	DistanceCost       CostDefinition `json:"distance_cost,omitempty" yaml:"distance_cost,omitempty"`
	ShiftAmountCost    CostDefinition `json:"shift_amount_cost,omitempty" yaml:"shift_amount_cost,omitempty"`
	ShiftUsesCost      CostDefinition `json:"shift_uses_cost,omitempty" yaml:"shift_uses_cost,omitempty"`
	DurationCost       CostDefinition `json:"duration_cost,omitempty" yaml:"duration_cost,omitempty"`
	BaseCost           CostDefinition `json:"base_cost" yaml:"base_cost"`
	Perks              []PerkConfig   `json:"perks" yaml:"perks"`
	Effects            []EffectConfig `json:"effects,omitempty" yaml:"effects,omitempty"`
}

// EffectConfig is a nested effect type for Persistent Effect.
type EffectConfig struct {
	ID          string `json:"id" yaml:"id"`
	Description string `json:"description" yaml:"description"`
	AddCost     int    `json:"add_cost" yaml:"add_cost"`
	EnergyCost  int    `json:"energy_cost" yaml:"energy_cost"`
}

// InteractionConfig defines an Interaction type and its costs.
type InteractionConfig struct {
	Type               string         `json:"type" yaml:"type"`
	Description        string         `json:"description" yaml:"description"`
	DefaultRange       int            `json:"default_range,omitempty" yaml:"default_range,omitempty"`
	DefaultTargets     int            `json:"default_targets,omitempty" yaml:"default_targets,omitempty"`
	DefaultVisible     bool           `json:"default_visible,omitempty" yaml:"default_visible,omitempty"`
	DefaultObstructed  bool           `json:"default_obstructed,omitempty" yaml:"default_obstructed,omitempty"`
	DefaultRadius      int            `json:"default_radius,omitempty" yaml:"default_radius,omitempty"`
	DefaultDuration    int            `json:"default_duration,omitempty" yaml:"default_duration,omitempty"`
	DefaultCounter     string         `json:"default_counter,omitempty" yaml:"default_counter,omitempty"`
	RangeCost          CostDefinition `json:"range_cost,omitempty" yaml:"range_cost,omitempty"`
	TargetCost         CostDefinition `json:"target_cost,omitempty" yaml:"target_cost,omitempty"`
	RadiusCost         CostDefinition `json:"radius_cost,omitempty" yaml:"radius_cost,omitempty"`
	RangeExtensionCost CostDefinition `json:"range_extension_cost,omitempty" yaml:"range_extension_cost,omitempty"`
	DurationCost       CostDefinition `json:"duration_cost,omitempty" yaml:"duration_cost,omitempty"`
	Perks              []PerkConfig   `json:"perks" yaml:"perks"`
	BaseCost           CostDefinition `json:"base_cost" yaml:"base_cost"`
}

// ValidationConfig defines validation rules and costs.
type ValidationConfig struct {
	DefaultEngagementTrait string                     `json:"default_engagement_trait" yaml:"default_engagement_trait"`
	DefaultCounterCount    int                        `json:"default_counter_count" yaml:"default_counter_count"`
	Engagement             EngagementValidationConfig `json:"engagement" yaml:"engagement"`
	Counter                CounterValidationConfig    `json:"counter" yaml:"counter"`
}

// EngagementValidationConfig defines engagement roll modes.
type EngagementValidationConfig struct {
	Modes []PerkConfig `json:"modes" yaml:"modes"`
}

// CounterValidationConfig defines counter roll rules.
type CounterValidationConfig struct {
	Types            []PerkConfig   `json:"types" yaml:"types"`
	SingleCounterCost CostDefinition `json:"single_counter_cost" yaml:"single_counter_cost"`
	TierShifts       []PerkConfig   `json:"tier_shifts" yaml:"tier_shifts"`
}

// TraitConfig holds the trait lists.
type TraitConfig struct {
	General []string `json:"general" yaml:"general"`
	Offense []string `json:"offense" yaml:"offense"`
	Defense []string `json:"defense" yaml:"defense"`
}

// DiceConfig holds the available dice options.
type DiceConfig struct {
	Damage  []string `json:"damage" yaml:"damage"`
	Generic []string `json:"generic" yaml:"generic"`
}

// Load reads the YAML configuration from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}

// Validate performs basic validation on the loaded configuration.
func (c *Config) Validate() error {
	if c.Version == "" {
		return fmt.Errorf("config version is required")
	}
	if len(c.AbilityBuilder.AbilityTypes) == 0 {
		return fmt.Errorf("at least one ability type is required")
	}
	if len(c.AbilityBuilder.Enactments) == 0 {
		return fmt.Errorf("at least one enactment is required")
	}
	if len(c.AbilityBuilder.Interactions) == 0 {
		return fmt.Errorf("at least one interaction is required")
	}
	return nil
}

// DefaultPath returns the default configuration file path.
func DefaultPath() string {
	return "config/ability-builder.yaml"
}
