package models

// Proficiency represents a character's proficiency level in a trait.
type Proficiency string

const (
	ProfClumsy    Proficiency = "Clumsy"
	ProfUntrained Proficiency = "Untrained"
	ProfTrained   Proficiency = "Trained"
	ProfExpert    Proficiency = "Expert"
	ProfMaster    Proficiency = "Master"
	ProfLegendary Proficiency = "Legendary"
)

// ProficiencyDice maps proficiency levels to their dice.
var ProficiencyDice = map[Proficiency]string{
	ProfClumsy:    "d4",
	ProfUntrained: "d6",
	ProfTrained:   "d8",
	ProfExpert:    "d10",
	ProfMaster:    "d12",
	ProfLegendary: "d20",
}

// AllProficiencies in order from lowest to highest.
var AllProficiencies = []Proficiency{
	ProfClumsy,
	ProfUntrained,
	ProfTrained,
	ProfExpert,
	ProfMaster,
	ProfLegendary,
}

// ProficiencyOption is used for template rendering of proficiency dropdowns.
type ProficiencyOption struct {
	Value    Proficiency
	Label    string // e.g. "Trained (d8)"
	DiceTier int
}

// GetProficiencyOptions returns all proficiency options for template dropdowns.
func GetProficiencyOptions() []ProficiencyOption {
	return []ProficiencyOption{
		{ProfClumsy, "Clumsy (d4)", 1},
		{ProfUntrained, "Untrained (d6)", 2},
		{ProfTrained, "Trained (d8)", 3},
		{ProfExpert, "Expert (d10)", 4},
		{ProfMaster, "Master (d12)", 5},
		{ProfLegendary, "Legendary (d20)", 6},
	}
}

// TraitPointsForLevel returns the number of trait points available at a given level.
// Starts at 10 points at level 1.
func TraitPointsForLevel(level int) int {
	if level < 1 {
		level = 1
	}
	return 10 + (level - 1)
}

// Character represents a player character with traits and abilities.
type Character struct {
	ID          string `json:"id"`
	Level       int    `json:"level"`
	Name        string `json:"name"`
	Age         string `json:"age"`
	Size        string `json:"size"`
	Alignment   string `json:"alignment"`
	Backstory   string `json:"backstory"`
	Personality string `json:"personality"`
	Appearance  string `json:"appearance"`
	Hobbies     string `json:"hobbies"`
	Occupation  string `json:"occupation"`
	Inventory   string `json:"inventory"`
	Quirks      string `json:"quirks"`

	GeneralTraits map[string]Proficiency `json:"general_traits"`
	OffenseTraits map[string]Proficiency `json:"offense_traits"`
	DefenseTraits map[string]Proficiency `json:"defense_traits"`

	VitalHP       Proficiency `json:"vital_hp"`
	VitalMovement Proficiency `json:"vital_movement"`
	VitalEnergy   Proficiency `json:"vital_energy"`

	CurrentHP     int `json:"current_hp"`
	CurrentEnergy int `json:"current_energy"`

	Abilities []Ability `json:"abilities"`
}

// NewCharacter creates a character with default trait values.
func NewCharacter(id string, general, offense, defense []string) Character {
	c := Character{
		ID:            id,
		Level:         1,
		GeneralTraits: make(map[string]Proficiency),
		OffenseTraits: make(map[string]Proficiency),
		DefenseTraits: make(map[string]Proficiency),
		VitalHP:       ProfUntrained,
		VitalMovement: ProfUntrained,
		VitalEnergy:   ProfUntrained,
		Abilities:     []Ability{},
	}
	for _, t := range general {
		c.GeneralTraits[t] = ProfUntrained
	}
	for _, t := range offense {
		c.OffenseTraits[t] = ProfUntrained
	}
	for _, t := range defense {
		c.DefenseTraits[t] = ProfUntrained
	}
	return c
}

// VitalOption is used for template rendering of vital stat dropdowns.
type VitalOption struct {
	Value Proficiency
	Label string
}

// GetVitalHPOptions returns HP options for template dropdowns.
func GetVitalHPOptions() []VitalOption {
	return []VitalOption{
		{ProfClumsy, "Clumsy (8 HP)"},
		{ProfUntrained, "Untrained (12 HP)"},
		{ProfTrained, "Trained (16 HP)"},
		{ProfExpert, "Expert (20 HP)"},
		{ProfMaster, "Master (24 HP)"},
		{ProfLegendary, "Legendary (28 HP)"},
	}
}

// GetVitalMovementOptions returns Movement options for template dropdowns.
func GetVitalMovementOptions() []VitalOption {
	return []VitalOption{
		{ProfClumsy, "Clumsy (3m)"},
		{ProfUntrained, "Untrained (4m)"},
		{ProfTrained, "Trained (5m)"},
		{ProfExpert, "Expert (6m)"},
		{ProfMaster, "Master (7m)"},
		{ProfLegendary, "Legendary (8m)"},
	}
}

// GetVitalEnergyOptions returns Energy options for template dropdowns.
func GetVitalEnergyOptions() []VitalOption {
	return []VitalOption{
		{ProfClumsy, "Clumsy (3)"},
		{ProfUntrained, "Untrained (4)"},
		{ProfTrained, "Trained (5)"},
		{ProfExpert, "Expert (6)"},
		{ProfMaster, "Master (7)"},
		{ProfLegendary, "Legendary (8)"},
	}
}

// VitalHPValues maps proficiency to HP values.
var VitalHPValues = map[Proficiency]int{
	ProfClumsy:    8,
	ProfUntrained: 12,
	ProfTrained:   16,
	ProfExpert:    20,
	ProfMaster:    24,
	ProfLegendary: 28,
}

// VitalMovementValues maps proficiency to movement in meters.
var VitalMovementValues = map[Proficiency]int{
	ProfClumsy:    3,
	ProfUntrained: 4,
	ProfTrained:   5,
	ProfExpert:    6,
	ProfMaster:    7,
	ProfLegendary: 8,
}

// VitalEnergyValues maps proficiency to energy points.
var VitalEnergyValues = map[Proficiency]int{
	ProfClumsy:    3,
	ProfUntrained: 4,
	ProfTrained:   5,
	ProfExpert:    6,
	ProfMaster:    7,
	ProfLegendary: 8,
}

// TraitPointsBudget returns the total trait points available for this character.
func (c *Character) TraitPointsBudget() int {
	return TraitPointsForLevel(c.Level)
}

// profCost returns the cost of a proficiency level (0 for Untrained, which is the baseline).
func profCost(p Proficiency) int {
	switch p {
	case ProfClumsy:
		return -1
	case ProfUntrained:
		return 0
	case ProfTrained:
		return 1
	case ProfExpert:
		return 2
	case ProfMaster:
		return 3
	case ProfLegendary:
		return 4
	}
	return 0
}

// TraitPointsUsed returns the total trait points spent on all traits.
func (c *Character) TraitPointsUsed() int {
	total := 0
	for _, p := range c.GeneralTraits {
		total += profCost(p)
	}
	for _, p := range c.OffenseTraits {
		total += profCost(p)
	}
	for _, p := range c.DefenseTraits {
		total += profCost(p)
	}
	total += profCost(c.VitalHP)
	total += profCost(c.VitalMovement)
	total += profCost(c.VitalEnergy)
	return total
}

// GetHP returns the character's HP based on their vital HP proficiency.
func (c *Character) GetHP() int {
	return VitalHPValues[c.VitalHP]
}

// GetMovement returns the character's movement based on their vital movement proficiency.
func (c *Character) GetMovement() int {
	return VitalMovementValues[c.VitalMovement]
}

// GetEnergy returns the character's energy based on their vital energy proficiency.
func (c *Character) GetEnergy() int {
	return VitalEnergyValues[c.VitalEnergy]
}
