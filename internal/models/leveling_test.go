package models

import (
	"errors"
	"testing"
)

func TestLevelingTable(t *testing.T) {
	// Verify the leveling tables match the docs
	// Trait points: 8 base (22 traits), then +1,+1,+1,+2,+1,+1,+1,+1,+2 = 19 at level 10
	expectedTraitTotals := [11]int{0, 8, 9, 10, 11, 13, 14, 15, 16, 17, 19}
	for level := 1; level <= 10; level++ {
		got := TotalTraitPointsAtLevel(22, level)
		if got != expectedTraitTotals[level] {
			t.Errorf("TotalTraitPointsAtLevel(22, %d) = %d, want %d", level, got, expectedTraitTotals[level])
		}
	}

	// Ability points: 10 base, then +2,+3,+2,+4,+2,+3,+2,+3,+5 = 36 at level 10
	expectedAbilityTotals := [11]int{0, 10, 12, 15, 17, 21, 23, 26, 28, 31, 36}
	for level := 1; level <= 10; level++ {
		got := TotalAbilityPointsAtLevel(level)
		if got != expectedAbilityTotals[level] {
			t.Errorf("TotalAbilityPointsAtLevel(%d) = %d, want %d", level, got, expectedAbilityTotals[level])
		}
	}
}

func TestLevelUp(t *testing.T) {
	c := NewCharacter()

	// Level 1 -> 2
	if err := c.LevelUp(); err != nil {
		t.Fatalf("LevelUp to 2 failed: %v", err)
	}
	if c.Level != 2 {
		t.Errorf("Level = %d, want 2", c.Level)
	}
	if got := c.TotalTraitPoints(); got != 9 {
		t.Errorf("TotalTraitPoints at level 2 = %d, want 9", got)
	}
	if got := c.TotalAbilityPoints(); got != 12 {
		t.Errorf("TotalAbilityPoints at level 2 = %d, want 12", got)
	}

	// Level up to max
	for c.Level < MaxLevel {
		if err := c.LevelUp(); err != nil {
			t.Fatalf("LevelUp to %d failed: %v", c.Level+1, err)
		}
	}
	if c.Level != 10 {
		t.Errorf("Level = %d, want 10", c.Level)
	}
	if got := c.TotalTraitPoints(); got != 19 {
		t.Errorf("TotalTraitPoints at level 10 = %d, want 19", got)
	}
	if got := c.TotalAbilityPoints(); got != 36 {
		t.Errorf("TotalAbilityPoints at level 10 = %d, want 36", got)
	}

	// Cannot exceed max level
	err := c.LevelUp()
	if !errors.Is(err, ErrAlreadyMaxLevel) {
		t.Errorf("LevelUp at max: got %v, want ErrAlreadyMaxLevel", err)
	}
}

func TestLevelDown(t *testing.T) {
	c := NewCharacter()

	// Cannot level down at level 1
	err := c.LevelDown()
	if !errors.Is(err, ErrAlreadyMinLevel) {
		t.Errorf("LevelDown at level 1: got %v, want ErrAlreadyMinLevel", err)
	}

	// Level up to 3
	c.LevelUp()
	c.LevelUp()

	if c.Level != 3 {
		t.Fatalf("Level = %d, want 3", c.Level)
	}

	// Level down: should work (no points spent)
	if err := c.LevelDown(); err != nil {
		t.Fatalf("LevelDown from 3 failed: %v", err)
	}
	if c.Level != 2 {
		t.Errorf("Level = %d, want 2", c.Level)
	}
	if got := c.TotalTraitPoints(); got != 9 {
		t.Errorf("TotalTraitPoints at level 2 = %d, want 9", got)
	}
}

func TestLevelDownBlockedBySpentPoints(t *testing.T) {
	c := NewCharacter()
	c.LevelUp() // Level 2: 9 trait points, 12 ability points

	// Spend all 9 trait points
	c.GeneralTraits.Strength = Master    // 4
	c.GeneralTraits.Dexterity = Master   // 4
	c.GeneralTraits.Stealth = Untrained  // 1
	// Total spent: 9

	// Level down would reduce trait points to 8, but 9 are spent
	err := c.LevelDown()
	if !errors.Is(err, ErrPointsOverspent) {
		t.Errorf("LevelDown with overspend: got %v, want ErrPointsOverspent", err)
	}

	// Free 1 point
	c.GeneralTraits.Stealth = Clumsy // refund 1, now 8 spent

	// Now level down should work
	if err := c.LevelDown(); err != nil {
		t.Fatalf("LevelDown after respec failed: %v", err)
	}
	if c.Level != 1 {
		t.Errorf("Level = %d, want 1", c.Level)
	}
}

func TestLevelDownBlockedByAbilityPoints(t *testing.T) {
	c := NewCharacter()
	c.LevelUp() // Level 2: 12 ability points

	// Spend 11 ability points via an ability
	c.Abilities = []Ability{
		{
			Name:       "Big Spell",
			Type:       AbilityTypeExecution,
			EnergyCost: 3,
			Perks: []Perk{
				{TotalAddCost: 11},
			},
		},
	}

	// Level down would reduce ability points to 10, but 11 are spent
	err := c.LevelDown()
	if !errors.Is(err, ErrPointsOverspent) {
		t.Errorf("LevelDown with ability overspend: got %v, want ErrPointsOverspent", err)
	}
}

func TestLevelUpDownRoundTrip(t *testing.T) {
	c := NewCharacter()

	// Level up from 1 to 5
	for i := 0; i < 4; i++ {
		c.LevelUp()
	}
	if c.Level != 5 {
		t.Fatalf("Level = %d, want 5", c.Level)
	}

	// Spend some points
	c.GeneralTraits.Strength = Trained  // 2 points
	c.GeneralTraits.Dexterity = Trained // 2 points
	// Total trait spent: 4, available at level 5: 13

	// Level back down to 3
	c.LevelDown() // from 5 to 4 (loses 2 trait, 4 ability)
	c.LevelDown() // from 4 to 3 (loses 1 trait, 2 ability)

	if c.Level != 3 {
		t.Errorf("Level = %d, want 3", c.Level)
	}
	if got := c.TotalTraitPoints(); got != 10 {
		t.Errorf("TotalTraitPoints at level 3 = %d, want 10", got)
	}
	if got := c.TotalAbilityPoints(); got != 15 {
		t.Errorf("TotalAbilityPoints at level 3 = %d, want 15", got)
	}
	if got := c.TraitPointsAvailable(); got != 6 { // 10 - 4 = 6
		t.Errorf("TraitPointsAvailable = %d, want 6", got)
	}
}

func TestCanLevelUpDown(t *testing.T) {
	c := NewCharacter()

	if !c.CanLevelUp() {
		t.Error("CanLevelUp should be true at level 1")
	}
	if c.CanLevelDown() {
		t.Error("CanLevelDown should be false at level 1")
	}

	c.LevelUp()
	if !c.CanLevelUp() {
		t.Error("CanLevelUp should be true at level 2")
	}
	if !c.CanLevelDown() {
		t.Error("CanLevelDown should be true at level 2 with no points spent")
	}
}

func TestPointsNeededToLevelDown(t *testing.T) {
	c := NewCharacter()
	c.LevelUp() // Level 2: 9 trait, 12 ability

	// Spend exactly at the limit
	c.GeneralTraits.Strength = Master   // 4
	c.GeneralTraits.Dexterity = Master  // 4
	c.GeneralTraits.Stealth = Untrained // 1 -> total 9

	traitNeeded, abilityNeeded := c.PointsNeededToLevelDown()
	// Need to free 1 trait point (9 spent, would have 8 after level down)
	if traitNeeded != 1 {
		t.Errorf("traitNeeded = %d, want 1", traitNeeded)
	}
	if abilityNeeded != 0 {
		t.Errorf("abilityNeeded = %d, want 0", abilityNeeded)
	}
}
