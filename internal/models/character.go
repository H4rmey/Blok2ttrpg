package models

// LevelSnapshot records what changed when leveling up, enabling undo on level-down.
type LevelSnapshot struct {
	Level              int              `yaml:"level"`
	TraitPointsGained  int              `yaml:"trait_points_gained"`
	AbilityPointsGained int            `yaml:"ability_points_gained"`
	TraitAllocations   map[string]int   `yaml:"trait_allocations,omitempty"`
	AbilityPointsSpent map[string]int   `yaml:"ability_points_spent,omitempty"`
}

// Character represents a full Blok2ttrpg character sheet.
type Character struct {
	// Meta
	Version string `yaml:"version"`
	Level   int    `yaml:"level"`

	// Character info
	Attributes          Attributes           `yaml:"attributes"`
	TemporaryAttributes []TemporaryAttribute `yaml:"temporary_attributes,omitempty"`

	// Traits
	GeneralTraits  GeneralTraits  `yaml:"general_traits"`
	CombativeTraits CombativeTraits `yaml:"combative_traits"`

	// Abilities
	Abilities []Ability `yaml:"abilities,omitempty"`

	// Level history for undo
	LevelHistory []LevelSnapshot `yaml:"level_history,omitempty"`

	// Campaign settings
	Settings CampaignSettings `yaml:"settings"`
}

// CampaignSettings holds configurable values that affect point calculations.
type CampaignSettings struct {
	TraitCount int `yaml:"trait_count"`
}

// NewCharacter creates a new level 1 character with default settings.
func NewCharacter() *Character {
	return &Character{
		Version: "1.0",
		Level:   1,
		Settings: CampaignSettings{
			TraitCount: 22, // Standard setting: 11 general + 11 combative
		},
		LevelHistory: []LevelSnapshot{
			{
				Level:               1,
				TraitPointsGained:   0, // Base points are calculated, not "gained"
				AbilityPointsGained: 0, // Base points are calculated, not "gained"
			},
		},
	}
}

// TotalTraitPoints returns the total trait points available at the current level.
func (c *Character) TotalTraitPoints() int {
	base := (c.Settings.TraitCount + 2) / 3
	for _, snap := range c.LevelHistory {
		base += snap.TraitPointsGained
	}
	return base
}

// TraitPointsSpent returns total trait points currently allocated.
func (c *Character) TraitPointsSpent() int {
	return c.GeneralTraits.PointsSpent() + c.CombativeTraits.PointsSpent()
}

// TraitPointsAvailable returns unspent trait points.
func (c *Character) TraitPointsAvailable() int {
	return c.TotalTraitPoints() - c.TraitPointsSpent()
}

// TotalAbilityPoints returns the total ability points available at the current level.
func (c *Character) TotalAbilityPoints() int {
	base := 10 // Starting ability points at level 1
	for _, snap := range c.LevelHistory {
		base += snap.AbilityPointsGained
	}
	return base
}

// AbilityPointsSpent returns total ability points currently invested in abilities.
func (c *Character) AbilityPointsSpent() int {
	total := 0
	for _, a := range c.Abilities {
		total += a.TotalAddCost()
	}
	return total
}

// AbilityPointsAvailable returns unspent ability points.
func (c *Character) AbilityPointsAvailable() int {
	return c.TotalAbilityPoints() - c.AbilityPointsSpent()
}
