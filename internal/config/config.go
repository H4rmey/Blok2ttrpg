package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

var profileIDPattern = regexp.MustCompile(`^[a-z0-9_-]+$`)

// Config is the top-level ability-builder configuration loaded from YAML.
type Config struct {
	Version        string               `json:"version" yaml:"version"`
	ProfileID      string               `json:"profile_id" yaml:"profile_id"`
	AbilityBuilder AbilityBuilderConfig `json:"ability_builder" yaml:"ability_builder"`
}

// AbilityBuilderConfig holds the main ability-builder rules and costs.
type AbilityBuilderConfig struct {
	FileOrder           []string                     `json:"file_order" yaml:"file_order"`
	AdditionalEnactment CostDefinition               `json:"additional_enactment" yaml:"additional_enactment"`
	AbilityTypes        map[string]AbilityTypeConfig `json:"ability_types" yaml:"ability_types"`
	Enactments          map[string]EnactmentConfig   `json:"enactments" yaml:"enactments"`
	Interactions        map[string]InteractionConfig `json:"interactions" yaml:"interactions"`
	Validations         ValidationConfig             `json:"validations" yaml:"validations"`
	Traits              TraitConfig                  `json:"traits" yaml:"traits"`
	Proficiencies       []ProficiencyConfig          `json:"proficiencies" yaml:"proficiencies"`
	Leveling            LevelingConfig               `json:"leveling" yaml:"leveling"`
	Dice                DiceConfig                   `json:"dice" yaml:"dice"`
	Combat              CombatConfig                 `json:"combat" yaml:"combat"`
	States              StatesConfig                 `json:"states" yaml:"states"`
}

// CombatConfig holds global combat rules.
type CombatConfig struct {
	Actions CombatActions `json:"actions" yaml:"actions"`
}

// CombatActions defines the action economy defaults.
type CombatActions struct {
	Amount int `json:"amount" yaml:"amount"`
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
	BaseUpkeepAction      int                  `json:"base_upkeep_action,omitempty" yaml:"base_upkeep_action,omitempty"`
	BaseUpkeepEnergy      int                  `json:"base_upkeep_energy,omitempty" yaml:"base_upkeep_energy,omitempty"`
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
	Fields                []FieldConfig        `json:"fields,omitempty" yaml:"fields,omitempty"`
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
	Fields             []FieldConfig  `json:"fields,omitempty" yaml:"fields,omitempty"`
}

// FieldConfig defines a configurable field on any card.
type FieldConfig struct {
	Key            string          `json:"key" yaml:"key"`
	Label          string          `json:"label" yaml:"label"`
	Type           string          `json:"type" yaml:"type"`
	Cost           *CostDefinition `json:"cost,omitempty" yaml:"cost,omitempty"`
	Options        []FieldOption   `json:"options,omitempty" yaml:"options,omitempty"`
	OptionsSource  string          `json:"options_source,omitempty" yaml:"options_source,omitempty"`
	Default        interface{}     `json:"default,omitempty" yaml:"default,omitempty"`
	Min            int             `json:"min,omitempty" yaml:"min,omitempty"`
	Max            int             `json:"max,omitempty" yaml:"max,omitempty"`
	Step           int             `json:"step,omitempty" yaml:"step,omitempty"`
	Rounding       string          `json:"rounding,omitempty" yaml:"rounding,omitempty"`
	PerStep        *StepCosts      `json:"per_step,omitempty" yaml:"per_step,omitempty"`
	DefaultCount   int             `json:"default_count,omitempty" yaml:"default_count,omitempty"`
	PerItem        *StepCosts      `json:"per_item,omitempty" yaml:"per_item,omitempty"`
	Export         *FieldExport    `json:"export,omitempty" yaml:"export,omitempty"`
	RowFields      []FieldConfig   `json:"row_fields,omitempty" yaml:"row_fields,omitempty"`
	StoresTo       string          `json:"stores_to,omitempty" yaml:"stores_to,omitempty"`
	VisibilityWhen string          `json:"visibility_when,omitempty" yaml:"visibility_when,omitempty"`
	ShowWhen       interface{}     `json:"show_when,omitempty" yaml:"show_when,omitempty"`
}

// FieldOption defines an option for dropdown/cascade field types.
type FieldOption struct {
	Value  string          `json:"value" yaml:"value"`
	Label  string          `json:"label" yaml:"label"`
	Cost   *CostDefinition `json:"cost,omitempty" yaml:"cost,omitempty"`
	Fields []FieldConfig   `json:"fields,omitempty" yaml:"fields,omitempty"`
}

// FieldExport defines how a field is exported to YAML.
type FieldExport struct {
	Key             string `json:"key" yaml:"key"`
	Suffix          string `json:"suffix,omitempty" yaml:"suffix,omitempty"`
	OmitWhenDefault bool   `json:"omit_when_default,omitempty" yaml:"omit_when_default,omitempty"`
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
	Fields             []FieldConfig  `json:"fields,omitempty" yaml:"fields,omitempty"`
}

// ValidationConfig defines validation rules and costs.
type ValidationConfig struct {
	DefaultEngagementTrait string                     `json:"default_engagement_trait" yaml:"default_engagement_trait"`
	DefaultCounterCount    int                        `json:"default_counter_count" yaml:"default_counter_count"`
	Engagement             EngagementValidationConfig `json:"engagement" yaml:"engagement"`
	Counter                CounterValidationConfig    `json:"counter" yaml:"counter"`
	Fields                 []FieldConfig              `json:"fields,omitempty" yaml:"fields,omitempty"`
}

// EngagementValidationConfig defines engagement roll modes.
type EngagementValidationConfig struct {
	Modes []PerkConfig `json:"modes" yaml:"modes"`
}

// CounterValidationConfig defines counter roll rules.
type CounterValidationConfig struct {
	Types             []PerkConfig   `json:"types" yaml:"types"`
	SingleCounterCost CostDefinition `json:"single_counter_cost" yaml:"single_counter_cost"`
	TierShifts        []PerkConfig   `json:"tier_shifts" yaml:"tier_shifts"`
}

// TraitConfig holds the trait lists.
type TraitConfig struct {
	General []string `json:"general" yaml:"general"`
	Offense []string `json:"offense" yaml:"offense"`
	Defense []string `json:"defense" yaml:"defense"`
	Vital   []string `json:"vital,omitempty" yaml:"vital,omitempty"`
}

// ProficiencyConfig defines a single proficiency tier and its dice/values.
type ProficiencyConfig struct {
	ID     string          `json:"id" yaml:"id"`
	Name   string          `json:"name" yaml:"name"`
	Cost   int             `json:"cost" yaml:"cost"`
	Note   string          `json:"note,omitempty" yaml:"note,omitempty"`
	Dice   ProficiencyDice `json:"dice" yaml:"dice"`
	Vitals map[string]int  `json:"vitals" yaml:"vitals"`
}

// ProficiencyDice maps category names to the dice they roll at this tier.
type ProficiencyDice struct {
	General string `json:"general" yaml:"general"`
	Offense string `json:"offense" yaml:"offense"`
	Defense string `json:"defense" yaml:"defense"`
}

// LevelingConfig holds the leveling tables.
type LevelingConfig struct {
	MaxLevel      int                  `json:"max_level" yaml:"max_level"`
	TraitPoints   LevelingPointsConfig `json:"trait_points" yaml:"trait_points"`
	AbilityPoints LevelingPointsConfig `json:"ability_points" yaml:"ability_points"`
}

// StatesConfig holds the state definitions for Enact State.
type StatesConfig struct {
	// AdditionalState is the surcharge applied for each state after the first
	// selected state on an Enact State card.
	AdditionalState CostDefinition        `json:"additional_state" yaml:"additional_state"`
	GeneralStates   []GeneralStateConfig  `json:"general_states" yaml:"general_states"`
	SpecificStates  []SpecificStateConfig `json:"specific_states" yaml:"specific_states"`
}

// GeneralStateConfig defines a flexible state with per-shift cost.
type GeneralStateConfig struct {
	ID          string         `json:"id" yaml:"id"`
	Name        string         `json:"name" yaml:"name"`
	Description string         `json:"description" yaml:"description"`
	MinShift    int            `json:"min_shift" yaml:"min_shift"`
	MaxShift    int            `json:"max_shift" yaml:"max_shift"`
	ShiftCost   CostDefinition `json:"shift_cost" yaml:"shift_cost"`
}

// SpecificStateConfig defines a specific state with fixed costs.
type SpecificStateConfig struct {
	ID          string `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	AddCost     int    `json:"add_cost" yaml:"add_cost"`
	EnergyCost  int    `json:"energy_cost" yaml:"energy_cost"`
}

// LevelingPointsConfig holds one resource's leveling table.
type LevelingPointsConfig struct {
	StandardTraitCount int          `json:"standard_trait_count,omitempty" yaml:"standard_trait_count,omitempty"`
	StartingFormula    string       `json:"starting_formula,omitempty" yaml:"starting_formula,omitempty"`
	Levels             []LevelEntry `json:"levels" yaml:"levels"`
}

// LevelEntry is a single row in a leveling table.
type LevelEntry struct {
	Level        int `json:"level" yaml:"level"`
	PointsGained int `json:"points_gained" yaml:"points_gained"`
	Total        int `json:"total" yaml:"total"`
}

// DiceConfig holds the available dice options.
type DiceConfig struct {
	Damage  []string `json:"damage" yaml:"damage"`
	Generic []string `json:"generic" yaml:"generic"`
}

// Load reads the YAML configuration from the given path.
// If the path is a directory, it loads section files from that directory.
// If the path is a file, it uses the legacy single-file loading behavior.
func Load(path string) (*Config, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat config path: %w", err)
	}

	var cfg Config
	if info.IsDir() {
		cfg, err = loadFromDirectory(path)
		if err != nil {
			return nil, fmt.Errorf("loading config from directory: %w", err)
		}
	} else {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parsing config file: %w", err)
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}

// strictUnmarshal parses YAML data with KnownFields(true) so unknown keys
// surface as errors instead of being silently dropped. The yaml.v3 decoder
// returns the offending field name in the error message.
func strictUnmarshal(data []byte, out interface{}) error {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(out); err != nil {
		return err
	}
	return nil
}

func loadFromDirectory(dir string) (Config, error) {
	cfg := Config{}

	// Load general.yaml for top-level fields
	generalPath := filepath.Join(dir, "general.yaml")
	if data, err := os.ReadFile(generalPath); err == nil {
		var general struct {
			Version             string           `yaml:"version"`
			ProfileID           string           `yaml:"profile_id"`
			Combat              CombatConfig     `yaml:"combat"`
			AdditionalEnactment CostDefinition   `yaml:"additional_enactment"`
			Dice                DiceConfig       `yaml:"dice"`
			Validations         ValidationConfig `yaml:"validations"`
		}
		if err := strictUnmarshal(data, &general); err != nil {
			return Config{}, fmt.Errorf("parsing general.yaml: %w", err)
		}
		cfg.Version = general.Version
		cfg.ProfileID = general.ProfileID
		cfg.AbilityBuilder.Combat = general.Combat
		cfg.AbilityBuilder.AdditionalEnactment = general.AdditionalEnactment
		cfg.AbilityBuilder.Dice = general.Dice
		cfg.AbilityBuilder.Validations = general.Validations
	}

	// Load file_order.yaml
	fileOrderPath := filepath.Join(dir, "file_order.yaml")
	if data, err := os.ReadFile(fileOrderPath); err == nil {
		var fileOrder struct {
			FileOrder []string `yaml:"file_order"`
		}
		if err := strictUnmarshal(data, &fileOrder); err != nil {
			return Config{}, fmt.Errorf("parsing file_order.yaml: %w", err)
		}
		cfg.AbilityBuilder.FileOrder = fileOrder.FileOrder
	}

	// Load ability_types.yaml
	abilityTypesPath := filepath.Join(dir, "ability_types.yaml")
	if data, err := os.ReadFile(abilityTypesPath); err == nil {
		var types struct {
			AbilityTypes map[string]AbilityTypeConfig `yaml:"ability_types"`
		}
		if err := strictUnmarshal(data, &types); err != nil {
			return Config{}, fmt.Errorf("parsing ability_types.yaml: %w", err)
		}
		cfg.AbilityBuilder.AbilityTypes = types.AbilityTypes
	}

	// Load enactments.yaml
	enactmentsPath := filepath.Join(dir, "enactments.yaml")
	if data, err := os.ReadFile(enactmentsPath); err == nil {
		var enacts struct {
			Enactments map[string]EnactmentConfig `yaml:"enactments"`
		}
		if err := strictUnmarshal(data, &enacts); err != nil {
			return Config{}, fmt.Errorf("parsing enactments.yaml: %w", err)
		}
		cfg.AbilityBuilder.Enactments = enacts.Enactments
	}

	// Load interactions.yaml
	interactionsPath := filepath.Join(dir, "interactions.yaml")
	if data, err := os.ReadFile(interactionsPath); err == nil {
		var inters struct {
			Interactions map[string]InteractionConfig `yaml:"interactions"`
		}
		if err := strictUnmarshal(data, &inters); err != nil {
			return Config{}, fmt.Errorf("parsing interactions.yaml: %w", err)
		}
		cfg.AbilityBuilder.Interactions = inters.Interactions
	}

	// Load proficiencies.yaml
	proficienciesPath := filepath.Join(dir, "proficiencies.yaml")
	if data, err := os.ReadFile(proficienciesPath); err == nil {
		var profs struct {
			Proficiencies []ProficiencyConfig `yaml:"proficiencies"`
		}
		if err := strictUnmarshal(data, &profs); err != nil {
			return Config{}, fmt.Errorf("parsing proficiencies.yaml: %w", err)
		}
		cfg.AbilityBuilder.Proficiencies = profs.Proficiencies
	}

	// Load leveling.yaml
	levelingPath := filepath.Join(dir, "leveling.yaml")
	if data, err := os.ReadFile(levelingPath); err == nil {
		var leveling struct {
			Leveling LevelingConfig `yaml:"leveling"`
		}
		if err := strictUnmarshal(data, &leveling); err != nil {
			return Config{}, fmt.Errorf("parsing leveling.yaml: %w", err)
		}
		cfg.AbilityBuilder.Leveling = leveling.Leveling
	}

	// Load traits.yaml
	traitsPath := filepath.Join(dir, "traits.yaml")
	if data, err := os.ReadFile(traitsPath); err == nil {
		var traits struct {
			Traits TraitConfig `yaml:"traits"`
		}
		if err := strictUnmarshal(data, &traits); err != nil {
			return Config{}, fmt.Errorf("parsing traits.yaml: %w", err)
		}
		cfg.AbilityBuilder.Traits = traits.Traits
	}

	// Load states.yaml (states configuration for Enact State)
	statesPath := filepath.Join(dir, "states.yaml")
	if data, err := os.ReadFile(statesPath); err == nil {
		var states StatesConfig
		if err := strictUnmarshal(data, &states); err != nil {
			return Config{}, fmt.Errorf("parsing states.yaml: %w", err)
		}
		cfg.AbilityBuilder.States = states
	}

	return cfg, nil
}

// Validate performs basic validation on the loaded configuration.
func (c *Config) Validate() error {
	if c.Version == "" {
		return fmt.Errorf("config version is required")
	}
	if c.ProfileID == "" {
		return fmt.Errorf("profile_id is required")
	}
	if !profileIDPattern.MatchString(c.ProfileID) {
		return fmt.Errorf("profile_id must contain only lowercase letters, numbers, underscores, and hyphens")
	}
	if len(c.AbilityBuilder.FileOrder) == 0 {
		return fmt.Errorf("ability_builder.file_order is required")
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

	// Validate field schemas
	for typeName, at := range c.AbilityBuilder.AbilityTypes {
		if err := validateFields(typeName, at.Fields); err != nil {
			return fmt.Errorf("ability_type %q: %w", typeName, err)
		}
	}
	for enName, en := range c.AbilityBuilder.Enactments {
		if err := validateFields(enName, en.Fields); err != nil {
			return fmt.Errorf("enactment %q: %w", enName, err)
		}
	}
	for intName, in := range c.AbilityBuilder.Interactions {
		if err := validateFields(intName, in.Fields); err != nil {
			return fmt.Errorf("interaction %q: %w", intName, err)
		}
	}
	if err := validateFields("validations", c.AbilityBuilder.Validations.Fields); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

var validFieldTypes = map[string]bool{
	"":            true, // allow empty default
	"checkbox":    true,
	"dropdown":    true,
	"free_text":   true,
	"free_number": true,
	"solutions":   true,
	"states":      true,
}

func validateFields(scope string, fields []FieldConfig) error {
	seen := map[string]bool{}
	for _, f := range fields {
		if f.Key == "" {
			return fmt.Errorf("field key is required")
		}
		if seen[f.Key] {
			return fmt.Errorf("duplicate field key %q", f.Key)
		}
		seen[f.Key] = true
		if f.Type == "" {
			return fmt.Errorf("field %q: type is required", f.Key)
		}
		if !validFieldTypes[f.Type] {
			return fmt.Errorf("field %q: unknown type %q", f.Key, f.Type)
		}
		if f.Type == "free_number" {
			if f.Min > f.Max {
				return fmt.Errorf("field %q: min (%d) > max (%d)", f.Key, f.Min, f.Max)
			}
			if f.Step < 0 {
				return fmt.Errorf("field %q: step must be non-negative", f.Key)
			}
		}
		if f.Type == "states" || f.Type == "solutions" {
			if len(f.RowFields) == 0 {
				return fmt.Errorf("field %q: %s requires row_fields", f.Key, f.Type)
			}
			if err := validateFields(f.Key+".row", f.RowFields); err != nil {
				return err
			}
		}
		if len(f.Options) > 0 && f.OptionsSource != "" {
			return fmt.Errorf("field %q: cannot mix options and options_source", f.Key)
		}
		for _, opt := range f.Options {
			if opt.Value == "" {
				return fmt.Errorf("field %q: option value is required", f.Key)
			}
			if err := validateFields(f.Key+"."+opt.Value, opt.Fields); err != nil {
				return err
			}
		}
	}
	return nil
}

// DefaultPath returns the default configuration file path.
func DefaultPath() string {
	return "config/ability-builder"
}
