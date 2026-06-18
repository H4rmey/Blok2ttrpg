package server

import "github.com/blok2ttrpg/charsheet/internal/models"

// TraitRow represents a single trait row for rendering in templates.
type TraitRow struct {
	Name        string
	Proficiency models.Proficiency
	Dice        string
	Cost        int
}

// GeneralTraitsVM is the view model for the general traits tab.
type GeneralTraitsVM struct {
	Traits          []TraitRow
	PointsSpent     int
	PointsAvailable int
	PointsTotal     int
}

// CombativeTraitsVM is the view model for the combative traits tab.
type CombativeTraitsVM struct {
	Offense         []TraitRow
	Defense         []TraitRow
	Vital           []VitalTraitRow
	PointsSpent     int
	PointsAvailable int
	PointsTotal     int
}

// VitalTraitRow represents a vital trait with its concrete value.
type VitalTraitRow struct {
	Name        string
	Proficiency models.Proficiency
	Value       int
	ValueLabel  string
}

// BuildGeneralTraitsVM builds the view model from a character.
func BuildGeneralTraitsVM(c *models.Character) GeneralTraitsVM {
	traits := []TraitRow{
		{Name: "Strength", Proficiency: c.GeneralTraits.Strength, Dice: c.GeneralTraits.Strength.Dice(), Cost: c.GeneralTraits.Strength.Cost()},
		{Name: "Dexterity", Proficiency: c.GeneralTraits.Dexterity, Dice: c.GeneralTraits.Dexterity.Dice(), Cost: c.GeneralTraits.Dexterity.Cost()},
		{Name: "Stealth", Proficiency: c.GeneralTraits.Stealth, Dice: c.GeneralTraits.Stealth.Dice(), Cost: c.GeneralTraits.Stealth.Cost()},
		{Name: "Perception", Proficiency: c.GeneralTraits.Perception, Dice: c.GeneralTraits.Perception.Dice(), Cost: c.GeneralTraits.Perception.Cost()},
		{Name: "Nature", Proficiency: c.GeneralTraits.Nature, Dice: c.GeneralTraits.Nature.Dice(), Cost: c.GeneralTraits.Nature.Cost()},
		{Name: "Crafting", Proficiency: c.GeneralTraits.Crafting, Dice: c.GeneralTraits.Crafting.Dice(), Cost: c.GeneralTraits.Crafting.Cost()},
		{Name: "People Skill", Proficiency: c.GeneralTraits.PeopleSkill, Dice: c.GeneralTraits.PeopleSkill.Dice(), Cost: c.GeneralTraits.PeopleSkill.Cost()},
		{Name: "Performance", Proficiency: c.GeneralTraits.Performance, Dice: c.GeneralTraits.Performance.Dice(), Cost: c.GeneralTraits.Performance.Cost()},
		{Name: "Thievery", Proficiency: c.GeneralTraits.Thievery, Dice: c.GeneralTraits.Thievery.Dice(), Cost: c.GeneralTraits.Thievery.Cost()},
		{Name: "Knowledge", Proficiency: c.GeneralTraits.Knowledge, Dice: c.GeneralTraits.Knowledge.Dice(), Cost: c.GeneralTraits.Knowledge.Cost()},
		{Name: "Magic", Proficiency: c.GeneralTraits.Magic, Dice: c.GeneralTraits.Magic.Dice(), Cost: c.GeneralTraits.Magic.Cost()},
	}

	return GeneralTraitsVM{
		Traits:          traits,
		PointsSpent:     c.TraitPointsSpent(),
		PointsAvailable: c.TraitPointsAvailable(),
		PointsTotal:     c.TotalTraitPoints(),
	}
}

// BuildCombativeTraitsVM builds the combative traits view model.
func BuildCombativeTraitsVM(c *models.Character) CombativeTraitsVM {
	offense := []TraitRow{
		{Name: "Precision", Proficiency: c.CombativeTraits.Offense.Precision, Dice: c.CombativeTraits.Offense.Precision.Dice(), Cost: c.CombativeTraits.Offense.Precision.Cost()},
		{Name: "Power", Proficiency: c.CombativeTraits.Offense.Power, Dice: c.CombativeTraits.Offense.Power.Dice(), Cost: c.CombativeTraits.Offense.Power.Cost()},
		{Name: "Mind", Proficiency: c.CombativeTraits.Offense.Mind, Dice: c.CombativeTraits.Offense.Mind.Dice(), Cost: c.CombativeTraits.Offense.Mind.Cost()},
		{Name: "Magic", Proficiency: c.CombativeTraits.Offense.Magic, Dice: c.CombativeTraits.Offense.Magic.Dice(), Cost: c.CombativeTraits.Offense.Magic.Cost()},
	}
	defense := []TraitRow{
		{Name: "Reflex", Proficiency: c.CombativeTraits.Defense.Reflex, Dice: c.CombativeTraits.Defense.Reflex.Dice(), Cost: c.CombativeTraits.Defense.Reflex.Cost()},
		{Name: "Constitution", Proficiency: c.CombativeTraits.Defense.Constitution, Dice: c.CombativeTraits.Defense.Constitution.Dice(), Cost: c.CombativeTraits.Defense.Constitution.Cost()},
		{Name: "Mind", Proficiency: c.CombativeTraits.Defense.Mind, Dice: c.CombativeTraits.Defense.Mind.Dice(), Cost: c.CombativeTraits.Defense.Mind.Cost()},
		{Name: "Magic", Proficiency: c.CombativeTraits.Defense.Magic, Dice: c.CombativeTraits.Defense.Magic.Dice(), Cost: c.CombativeTraits.Defense.Magic.Cost()},
	}
	vital := []VitalTraitRow{
		{Name: "HP", Proficiency: c.CombativeTraits.Vital.HP, Value: c.CombativeTraits.Vital.HP.HPValue(), ValueLabel: "HP"},
		{Name: "Movement", Proficiency: c.CombativeTraits.Vital.Movement, Value: c.CombativeTraits.Vital.Movement.MovementValue(), ValueLabel: "squares"},
		{Name: "Energy", Proficiency: c.CombativeTraits.Vital.Energy, Value: c.CombativeTraits.Vital.Energy.EnergyValue(), ValueLabel: "energy"},
	}

	return CombativeTraitsVM{
		Offense:         offense,
		Defense:         defense,
		Vital:           vital,
		PointsSpent:     c.TraitPointsSpent(),
		PointsAvailable: c.TraitPointsAvailable(),
		PointsTotal:     c.TotalTraitPoints(),
	}
}

// AbilityListVM is the view model for the abilities tab.
type AbilityListVM struct {
	Abilities       []AbilityRowVM
	PointsSpent     int
	PointsAvailable int
	PointsTotal     int
}

// AbilityRowVM represents an ability in the list.
type AbilityRowVM struct {
	Index      int
	Name       string
	Type       string
	AddCost    int
	EnergyCost int
}

// BuildAbilityListVM builds the view model for the ability list tab.
func BuildAbilityListVM(c *models.Character) AbilityListVM {
	rows := make([]AbilityRowVM, len(c.Abilities))
	for i, a := range c.Abilities {
		rows[i] = AbilityRowVM{
			Index:      i,
			Name:       a.Name,
			Type:       string(a.Type),
			AddCost:    a.TotalAddCost(),
			EnergyCost: a.TotalEnergyCost(),
		}
	}
	return AbilityListVM{
		Abilities:       rows,
		PointsSpent:     c.AbilityPointsSpent(),
		PointsAvailable: c.AbilityPointsAvailable(),
		PointsTotal:     c.TotalAbilityPoints(),
	}
}
