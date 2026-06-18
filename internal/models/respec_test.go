package models

import (
	"errors"
	"testing"
)

func TestSetGeneralTrait(t *testing.T) {
	c := NewCharacter() // 8 trait points available

	// Raise Strength to Trained (costs 2)
	if err := c.SetGeneralTrait("Strength", Trained); err != nil {
		t.Fatalf("SetGeneralTrait(Strength, Trained) failed: %v", err)
	}
	if c.GeneralTraits.Strength != Trained {
		t.Errorf("Strength = %s, want Trained", c.GeneralTraits.Strength)
	}
	if got := c.TraitPointsAvailable(); got != 6 {
		t.Errorf("TraitPointsAvailable = %d, want 6", got)
	}

	// Lower back to Clumsy (refund 2)
	if err := c.SetGeneralTrait("Strength", Clumsy); err != nil {
		t.Fatalf("SetGeneralTrait(Strength, Clumsy) failed: %v", err)
	}
	if got := c.TraitPointsAvailable(); got != 8 {
		t.Errorf("TraitPointsAvailable = %d, want 8", got)
	}
}

func TestSetGeneralTraitInsufficientPoints(t *testing.T) {
	c := NewCharacter() // 8 trait points

	// Spend 8 points
	c.SetGeneralTrait("Strength", Master)    // 4
	c.SetGeneralTrait("Dexterity", Master)   // 4

	// Try to spend 1 more
	err := c.SetGeneralTrait("Stealth", Untrained)
	if !errors.Is(err, ErrInsufficientPoints) {
		t.Errorf("SetGeneralTrait over budget: got %v, want ErrInsufficientPoints", err)
	}
}

func TestSetGeneralTraitLegendaryCap(t *testing.T) {
	c := NewCharacter()

	err := c.SetGeneralTrait("Strength", Legendary)
	if !errors.Is(err, ErrProficiencyCapReached) {
		t.Errorf("SetGeneralTrait to Legendary: got %v, want ErrProficiencyCapReached", err)
	}
}

func TestSetGeneralTraitInvalidName(t *testing.T) {
	c := NewCharacter()

	err := c.SetGeneralTrait("NonExistent", Trained)
	if err == nil {
		t.Error("SetGeneralTrait with invalid name should fail")
	}
}

func TestSetCombativeTrait(t *testing.T) {
	c := NewCharacter() // 8 trait points

	// Raise Precision (offense) to Expert (costs 3)
	if err := c.SetCombativeTrait("offense", "Precision", Expert); err != nil {
		t.Fatalf("SetCombativeTrait failed: %v", err)
	}
	if c.CombativeTraits.Offense.Precision != Expert {
		t.Errorf("Precision = %s, want Expert", c.CombativeTraits.Offense.Precision)
	}
	if got := c.TraitPointsAvailable(); got != 5 {
		t.Errorf("TraitPointsAvailable = %d, want 5", got)
	}

	// Raise HP (vital) to Trained (costs 2)
	if err := c.SetCombativeTrait("vital", "HP", Trained); err != nil {
		t.Fatalf("SetCombativeTrait(vital, HP) failed: %v", err)
	}
	if c.CombativeTraits.Vital.HP != Trained {
		t.Errorf("HP = %s, want Trained", c.CombativeTraits.Vital.HP)
	}
}

func TestSetCombativeTraitSharedPool(t *testing.T) {
	c := NewCharacter() // 8 points shared between general and combative

	// Spend 4 on general
	c.SetGeneralTrait("Strength", Master) // 4

	// Now only 4 left for combative
	err := c.SetCombativeTrait("offense", "Precision", Master) // costs 4 — should work
	if err != nil {
		t.Fatalf("SetCombativeTrait failed: %v", err)
	}

	// No points left — any further spending should fail
	err = c.SetCombativeTrait("defense", "Reflex", Untrained) // costs 1
	if !errors.Is(err, ErrInsufficientPoints) {
		t.Errorf("Expected ErrInsufficientPoints, got %v", err)
	}
}

func TestFreeRespec(t *testing.T) {
	c := NewCharacter() // 8 points

	// Allocate all points
	c.SetGeneralTrait("Strength", Master)    // 4
	c.SetGeneralTrait("Dexterity", Master)   // 4
	// 0 remaining

	// Re-spec: lower Strength, raise Stealth
	if err := c.SetGeneralTrait("Strength", Clumsy); err != nil {
		t.Fatalf("Lower Strength failed: %v", err)
	}
	// Now 4 available
	if err := c.SetGeneralTrait("Stealth", Master); err != nil {
		t.Fatalf("Raise Stealth failed: %v", err)
	}
	if got := c.TraitPointsAvailable(); got != 0 {
		t.Errorf("TraitPointsAvailable = %d, want 0", got)
	}
	if c.GeneralTraits.Strength != Clumsy {
		t.Errorf("Strength = %s, want Clumsy", c.GeneralTraits.Strength)
	}
	if c.GeneralTraits.Stealth != Master {
		t.Errorf("Stealth = %s, want Master", c.GeneralTraits.Stealth)
	}
}

func TestRemoveAbility(t *testing.T) {
	c := NewCharacter()

	a1 := Ability{Name: "Fireball", Type: AbilityTypeExecution, EnergyCost: 3, Perks: []Perk{{TotalAddCost: 5}}}
	a2 := Ability{Name: "Shield", Type: AbilityTypeReaction, EnergyCost: 3, Perks: []Perk{{TotalAddCost: 3}}}
	c.Abilities = []Ability{a1, a2}

	// 10 - (5+3) = 2 available
	if got := c.AbilityPointsAvailable(); got != 2 {
		t.Errorf("AbilityPointsAvailable = %d, want 2", got)
	}

	// Remove Fireball (index 0)
	removed, err := c.RemoveAbility(0)
	if err != nil {
		t.Fatalf("RemoveAbility failed: %v", err)
	}
	if removed.Name != "Fireball" {
		t.Errorf("Removed ability = %q, want Fireball", removed.Name)
	}

	// Now 10 - 3 = 7 available
	if got := c.AbilityPointsAvailable(); got != 7 {
		t.Errorf("AbilityPointsAvailable after remove = %d, want 7", got)
	}
	if len(c.Abilities) != 1 {
		t.Errorf("Abilities count = %d, want 1", len(c.Abilities))
	}
}

func TestAddAbilityBudgetCheck(t *testing.T) {
	c := NewCharacter() // 10 ability points

	// Add ability costing exactly 10
	a := Ability{Name: "Big Spell", Type: AbilityTypeExecution, EnergyCost: 3, Perks: []Perk{{TotalAddCost: 10}}}
	if err := c.AddAbility(a); err != nil {
		t.Fatalf("AddAbility(10 points) failed: %v", err)
	}

	// Try to add another
	a2 := Ability{Name: "Small Spell", Type: AbilityTypeExecution, EnergyCost: 3, Perks: []Perk{{TotalAddCost: 1}}}
	err := c.AddAbility(a2)
	if !errors.Is(err, ErrInsufficientPoints) {
		t.Errorf("AddAbility over budget: got %v, want ErrInsufficientPoints", err)
	}
}

func TestUpdateAbility(t *testing.T) {
	c := NewCharacter() // 10 ability points

	a := Ability{Name: "Fireball", Type: AbilityTypeExecution, EnergyCost: 3, Perks: []Perk{{TotalAddCost: 5}}}
	c.Abilities = []Ability{a}

	// Update to cost 8 (diff = +3, available = 5) — should work
	updated := Ability{Name: "Fireball+", Type: AbilityTypeExecution, EnergyCost: 5, Perks: []Perk{{TotalAddCost: 8}}}
	if err := c.UpdateAbility(0, updated); err != nil {
		t.Fatalf("UpdateAbility failed: %v", err)
	}
	if c.Abilities[0].Name != "Fireball+" {
		t.Errorf("Updated name = %q, want Fireball+", c.Abilities[0].Name)
	}

	// Try to update to cost 11 (diff = +3, available = 2) — should fail
	tooExpensive := Ability{Name: "Fireball++", Type: AbilityTypeExecution, EnergyCost: 5, Perks: []Perk{{TotalAddCost: 11}}}
	err := c.UpdateAbility(0, tooExpensive)
	if !errors.Is(err, ErrInsufficientPoints) {
		t.Errorf("UpdateAbility over budget: got %v, want ErrInsufficientPoints", err)
	}
}

func TestRemoveAbilityInvalidIndex(t *testing.T) {
	c := NewCharacter()

	_, err := c.RemoveAbility(0)
	if err == nil {
		t.Error("RemoveAbility on empty list should fail")
	}

	_, err = c.RemoveAbility(-1)
	if err == nil {
		t.Error("RemoveAbility with negative index should fail")
	}
}

func TestGetTraitAccessors(t *testing.T) {
	c := NewCharacter()
	c.GeneralTraits.Knowledge = Expert
	c.CombativeTraits.Defense.Constitution = Trained

	got, err := c.GetGeneralTrait("Knowledge")
	if err != nil {
		t.Fatalf("GetGeneralTrait failed: %v", err)
	}
	if got != Expert {
		t.Errorf("Knowledge = %s, want Expert", got)
	}

	got, err = c.GetCombativeTrait("defense", "Constitution")
	if err != nil {
		t.Fatalf("GetCombativeTrait failed: %v", err)
	}
	if got != Trained {
		t.Errorf("Constitution = %s, want Trained", got)
	}
}
