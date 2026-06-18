package models

import "fmt"

// SetGeneralTrait sets a general trait to a new proficiency level.
// Returns an error if the change would exceed available trait points or cap.
func (c *Character) SetGeneralTrait(traitName string, newProf Proficiency) error {
	if newProf > Master {
		return ErrProficiencyCapReached
	}

	oldProf, err := c.getGeneralTrait(traitName)
	if err != nil {
		return err
	}

	// Calculate the point difference
	diff := newProf.Cost() - oldProf.Cost()

	// Check if we have enough points (diff > 0 means spending more)
	if diff > 0 && diff > c.TraitPointsAvailable() {
		return ErrInsufficientPoints
	}

	return c.setGeneralTrait(traitName, newProf)
}

// SetCombativeTrait sets a combative trait to a new proficiency level.
// Returns an error if the change would exceed available trait points or cap.
func (c *Character) SetCombativeTrait(section, traitName string, newProf Proficiency) error {
	if newProf > Master {
		return ErrProficiencyCapReached
	}

	oldProf, err := c.getCombativeTrait(section, traitName)
	if err != nil {
		return err
	}

	diff := newProf.Cost() - oldProf.Cost()
	if diff > 0 && diff > c.TraitPointsAvailable() {
		return ErrInsufficientPoints
	}

	return c.setCombativeTrait(section, traitName, newProf)
}

// RemoveAbility removes an ability by index, refunding its add cost to the pool.
// Returns the removed ability for confirmation purposes.
func (c *Character) RemoveAbility(index int) (*Ability, error) {
	if index < 0 || index >= len(c.Abilities) {
		return nil, fmt.Errorf("ability index %d out of range", index)
	}

	removed := c.Abilities[index]
	c.Abilities = append(c.Abilities[:index], c.Abilities[index+1:]...)
	return &removed, nil
}

// AddAbility adds a new ability if there are enough ability points.
func (c *Character) AddAbility(a Ability) error {
	cost := a.TotalAddCost()
	if cost > c.AbilityPointsAvailable() {
		return ErrInsufficientPoints
	}
	c.Abilities = append(c.Abilities, a)
	return nil
}

// UpdateAbility replaces an ability at the given index with a new version.
// Validates that the total cost after replacement doesn't exceed the budget.
func (c *Character) UpdateAbility(index int, updated Ability) error {
	if index < 0 || index >= len(c.Abilities) {
		return fmt.Errorf("ability index %d out of range", index)
	}

	// Calculate what the budget would be after replacing
	oldCost := c.Abilities[index].TotalAddCost()
	newCost := updated.TotalAddCost()
	diff := newCost - oldCost

	if diff > 0 && diff > c.AbilityPointsAvailable() {
		return ErrInsufficientPoints
	}

	c.Abilities[index] = updated
	return nil
}

// Helper: get general trait proficiency by name
func (c *Character) getGeneralTrait(name string) (Proficiency, error) {
	switch name {
	case "Strength":
		return c.GeneralTraits.Strength, nil
	case "Dexterity":
		return c.GeneralTraits.Dexterity, nil
	case "Stealth":
		return c.GeneralTraits.Stealth, nil
	case "Perception":
		return c.GeneralTraits.Perception, nil
	case "Nature":
		return c.GeneralTraits.Nature, nil
	case "Crafting":
		return c.GeneralTraits.Crafting, nil
	case "People Skill":
		return c.GeneralTraits.PeopleSkill, nil
	case "Performance":
		return c.GeneralTraits.Performance, nil
	case "Thievery":
		return c.GeneralTraits.Thievery, nil
	case "Knowledge":
		return c.GeneralTraits.Knowledge, nil
	case "Magic":
		return c.GeneralTraits.Magic, nil
	default:
		return Clumsy, fmt.Errorf("unknown general trait: %q", name)
	}
}

// Helper: set general trait proficiency by name
func (c *Character) setGeneralTrait(name string, p Proficiency) error {
	switch name {
	case "Strength":
		c.GeneralTraits.Strength = p
	case "Dexterity":
		c.GeneralTraits.Dexterity = p
	case "Stealth":
		c.GeneralTraits.Stealth = p
	case "Perception":
		c.GeneralTraits.Perception = p
	case "Nature":
		c.GeneralTraits.Nature = p
	case "Crafting":
		c.GeneralTraits.Crafting = p
	case "People Skill":
		c.GeneralTraits.PeopleSkill = p
	case "Performance":
		c.GeneralTraits.Performance = p
	case "Thievery":
		c.GeneralTraits.Thievery = p
	case "Knowledge":
		c.GeneralTraits.Knowledge = p
	case "Magic":
		c.GeneralTraits.Magic = p
	default:
		return fmt.Errorf("unknown general trait: %q", name)
	}
	return nil
}

// Helper: get combative trait proficiency by section and name
func (c *Character) getCombativeTrait(section, name string) (Proficiency, error) {
	switch section {
	case "offense":
		switch name {
		case "Precision":
			return c.CombativeTraits.Offense.Precision, nil
		case "Power":
			return c.CombativeTraits.Offense.Power, nil
		case "Mind":
			return c.CombativeTraits.Offense.Mind, nil
		case "Magic":
			return c.CombativeTraits.Offense.Magic, nil
		default:
			return Clumsy, fmt.Errorf("unknown offense trait: %q", name)
		}
	case "defense":
		switch name {
		case "Reflex":
			return c.CombativeTraits.Defense.Reflex, nil
		case "Constitution":
			return c.CombativeTraits.Defense.Constitution, nil
		case "Mind":
			return c.CombativeTraits.Defense.Mind, nil
		case "Magic":
			return c.CombativeTraits.Defense.Magic, nil
		default:
			return Clumsy, fmt.Errorf("unknown defense trait: %q", name)
		}
	case "vital":
		switch name {
		case "HP":
			return c.CombativeTraits.Vital.HP, nil
		case "Movement":
			return c.CombativeTraits.Vital.Movement, nil
		case "Energy":
			return c.CombativeTraits.Vital.Energy, nil
		default:
			return Clumsy, fmt.Errorf("unknown vital trait: %q", name)
		}
	default:
		return Clumsy, fmt.Errorf("unknown combative section: %q", section)
	}
}

// Helper: set combative trait proficiency by section and name
func (c *Character) setCombativeTrait(section, name string, p Proficiency) error {
	switch section {
	case "offense":
		switch name {
		case "Precision":
			c.CombativeTraits.Offense.Precision = p
		case "Power":
			c.CombativeTraits.Offense.Power = p
		case "Mind":
			c.CombativeTraits.Offense.Mind = p
		case "Magic":
			c.CombativeTraits.Offense.Magic = p
		default:
			return fmt.Errorf("unknown offense trait: %q", name)
		}
	case "defense":
		switch name {
		case "Reflex":
			c.CombativeTraits.Defense.Reflex = p
		case "Constitution":
			c.CombativeTraits.Defense.Constitution = p
		case "Mind":
			c.CombativeTraits.Defense.Mind = p
		case "Magic":
			c.CombativeTraits.Defense.Magic = p
		default:
			return fmt.Errorf("unknown defense trait: %q", name)
		}
	case "vital":
		switch name {
		case "HP":
			c.CombativeTraits.Vital.HP = p
		case "Movement":
			c.CombativeTraits.Vital.Movement = p
		case "Energy":
			c.CombativeTraits.Vital.Energy = p
		default:
			return fmt.Errorf("unknown vital trait: %q", name)
		}
	default:
		return fmt.Errorf("unknown combative section: %q", section)
	}
	return nil
}

// GetGeneralTrait retrieves a general trait's proficiency by name (exported).
func (c *Character) GetGeneralTrait(name string) (Proficiency, error) {
	return c.getGeneralTrait(name)
}

// GetCombativeTrait retrieves a combative trait's proficiency by section and name (exported).
func (c *Character) GetCombativeTrait(section, name string) (Proficiency, error) {
	return c.getCombativeTrait(section, name)
}
