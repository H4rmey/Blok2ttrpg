package models

// GeneralTraits holds proficiency levels for all general (non-combat) traits.
type GeneralTraits struct {
	Strength    Proficiency `yaml:"strength"`
	Dexterity   Proficiency `yaml:"dexterity"`
	Stealth     Proficiency `yaml:"stealth"`
	Perception  Proficiency `yaml:"perception"`
	Nature      Proficiency `yaml:"nature"`
	Crafting    Proficiency `yaml:"crafting"`
	PeopleSkill Proficiency `yaml:"people_skill"`
	Performance Proficiency `yaml:"performance"`
	Thievery    Proficiency `yaml:"thievery"`
	Knowledge   Proficiency `yaml:"knowledge"`
	Magic       Proficiency `yaml:"magic"`
}

// OffenseTraits holds proficiency levels for offensive combat traits.
type OffenseTraits struct {
	Precision Proficiency `yaml:"precision"`
	Power     Proficiency `yaml:"power"`
	Mind      Proficiency `yaml:"mind"`
	Magic     Proficiency `yaml:"magic"`
}

// DefenseTraits holds proficiency levels for defensive combat traits.
type DefenseTraits struct {
	Reflex       Proficiency `yaml:"reflex"`
	Constitution Proficiency `yaml:"constitution"`
	Mind         Proficiency `yaml:"mind"`
	Magic        Proficiency `yaml:"magic"`
}

// VitalTraits holds proficiency levels for vital stats.
type VitalTraits struct {
	HP       Proficiency `yaml:"hp"`
	Movement Proficiency `yaml:"movement"`
	Energy   Proficiency `yaml:"energy"`
}

// CombativeTraits groups all combat-related traits.
type CombativeTraits struct {
	Offense OffenseTraits `yaml:"offense"`
	Defense DefenseTraits `yaml:"defense"`
	Vital   VitalTraits   `yaml:"vital"`
}

// PointsSpent calculates how many trait points are spent on general traits.
// Each proficiency level above Clumsy costs 1 point per step.
func (g *GeneralTraits) PointsSpent() int {
	return g.Strength.Cost() +
		g.Dexterity.Cost() +
		g.Stealth.Cost() +
		g.Perception.Cost() +
		g.Nature.Cost() +
		g.Crafting.Cost() +
		g.PeopleSkill.Cost() +
		g.Performance.Cost() +
		g.Thievery.Cost() +
		g.Knowledge.Cost() +
		g.Magic.Cost()
}

// PointsSpent calculates how many trait points are spent on combative traits.
func (c *CombativeTraits) PointsSpent() int {
	return c.Offense.Precision.Cost() +
		c.Offense.Power.Cost() +
		c.Offense.Mind.Cost() +
		c.Offense.Magic.Cost() +
		c.Defense.Reflex.Cost() +
		c.Defense.Constitution.Cost() +
		c.Defense.Mind.Cost() +
		c.Defense.Magic.Cost() +
		c.Vital.HP.Cost() +
		c.Vital.Movement.Cost() +
		c.Vital.Energy.Cost()
}

// GeneralTraitNames returns the list of general trait names in order.
func GeneralTraitNames() []string {
	return []string{
		"Strength", "Dexterity", "Stealth", "Perception", "Nature",
		"Crafting", "People Skill", "Performance", "Thievery", "Knowledge", "Magic",
	}
}

// OffenseTraitNames returns offensive trait names.
func OffenseTraitNames() []string {
	return []string{"Precision", "Power", "Mind", "Magic"}
}

// DefenseTraitNames returns defensive trait names.
func DefenseTraitNames() []string {
	return []string{"Reflex", "Constitution", "Mind", "Magic"}
}

// VitalTraitNames returns vital trait names.
func VitalTraitNames() []string {
	return []string{"HP", "Movement", "Energy"}
}
