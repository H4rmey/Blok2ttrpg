package models

import "errors"

// MaxLevel is the highest level a character can reach.
const MaxLevel = 10

// MinLevel is the lowest level.
const MinLevel = 1

// TraitPointsPerLevel defines how many trait points are gained at each level.
// Index 0 is unused; index 1 = level 1 (base, no gain), index 2 = level 2, etc.
var TraitPointsPerLevel = [11]int{
	0, // unused
	0, // Level 1: base (calculated from formula)
	1, // Level 2: +1
	1, // Level 3: +1
	1, // Level 4: +1
	2, // Level 5: +2
	1, // Level 6: +1
	1, // Level 7: +1
	1, // Level 8: +1
	1, // Level 9: +1
	2, // Level 10: +2
}

// AbilityPointsPerLevel defines how many ability points are gained at each level.
// Index 0 is unused; index 1 = level 1 (base 10, no gain), index 2 = level 2, etc.
var AbilityPointsPerLevel = [11]int{
	0, // unused
	0, // Level 1: base 10 (not a "gain")
	2, // Level 2: +2
	3, // Level 3: +3
	2, // Level 4: +2
	4, // Level 5: +4
	2, // Level 6: +2
	3, // Level 7: +3
	2, // Level 8: +2
	3, // Level 9: +3
	5, // Level 10: +5
}

// Leveling errors.
var (
	ErrAlreadyMaxLevel     = errors.New("character is already at maximum level")
	ErrAlreadyMinLevel     = errors.New("character is already at minimum level")
	ErrPointsOverspent     = errors.New("cannot level down: points from this level are still spent, re-spec first")
	ErrInsufficientPoints  = errors.New("insufficient points for this operation")
	ErrProficiencyCapReached = errors.New("cannot raise proficiency above Master with points (Legendary is not purchasable)")
)

// LevelUp advances the character one level, granting trait and ability points.
// Returns an error if already at max level.
func (c *Character) LevelUp() error {
	if c.Level >= MaxLevel {
		return ErrAlreadyMaxLevel
	}

	newLevel := c.Level + 1
	traitGain := TraitPointsPerLevel[newLevel]
	abilityGain := AbilityPointsPerLevel[newLevel]

	snapshot := LevelSnapshot{
		Level:               newLevel,
		TraitPointsGained:   traitGain,
		AbilityPointsGained: abilityGain,
	}

	c.Level = newLevel
	c.LevelHistory = append(c.LevelHistory, snapshot)

	return nil
}

// LevelDown reduces the character one level, revoking the points granted at that level.
// Returns an error if at min level or if the points from that level are still spent.
func (c *Character) LevelDown() error {
	if c.Level <= MinLevel {
		return ErrAlreadyMinLevel
	}

	// Check if we can safely remove the points
	lastSnapshot := c.LevelHistory[len(c.LevelHistory)-1]

	// After removing these points, would we be overspent on traits?
	futureTraitTotal := c.TotalTraitPoints() - lastSnapshot.TraitPointsGained
	if c.TraitPointsSpent() > futureTraitTotal {
		return ErrPointsOverspent
	}

	// After removing these points, would we be overspent on abilities?
	futureAbilityTotal := c.TotalAbilityPoints() - lastSnapshot.AbilityPointsGained
	if c.AbilityPointsSpent() > futureAbilityTotal {
		return ErrPointsOverspent
	}

	c.Level--
	c.LevelHistory = c.LevelHistory[:len(c.LevelHistory)-1]

	return nil
}

// CanLevelUp returns true if the character can level up.
func (c *Character) CanLevelUp() bool {
	return c.Level < MaxLevel
}

// CanLevelDown returns true if the character can level down without issues.
func (c *Character) CanLevelDown() bool {
	if c.Level <= MinLevel {
		return false
	}
	lastSnapshot := c.LevelHistory[len(c.LevelHistory)-1]
	futureTraitTotal := c.TotalTraitPoints() - lastSnapshot.TraitPointsGained
	futureAbilityTotal := c.TotalAbilityPoints() - lastSnapshot.AbilityPointsGained
	return c.TraitPointsSpent() <= futureTraitTotal && c.AbilityPointsSpent() <= futureAbilityTotal
}

// PointsNeededToLevelDown returns how many trait and ability points need to be freed
// before leveling down is possible. Returns (0, 0) if level down is already possible.
func (c *Character) PointsNeededToLevelDown() (traitPoints int, abilityPoints int) {
	if c.Level <= MinLevel {
		return 0, 0
	}
	lastSnapshot := c.LevelHistory[len(c.LevelHistory)-1]

	futureTraitTotal := c.TotalTraitPoints() - lastSnapshot.TraitPointsGained
	traitOverspend := c.TraitPointsSpent() - futureTraitTotal
	if traitOverspend < 0 {
		traitOverspend = 0
	}

	futureAbilityTotal := c.TotalAbilityPoints() - lastSnapshot.AbilityPointsGained
	abilityOverspend := c.AbilityPointsSpent() - futureAbilityTotal
	if abilityOverspend < 0 {
		abilityOverspend = 0
	}

	return traitOverspend, abilityOverspend
}

// TotalTraitPointsAtLevel calculates the total trait points at a specific level.
func TotalTraitPointsAtLevel(traitCount int, level int) int {
	base := (traitCount + 2) / 3
	for l := 2; l <= level; l++ {
		base += TraitPointsPerLevel[l]
	}
	return base
}

// TotalAbilityPointsAtLevel calculates the total ability points at a specific level.
func TotalAbilityPointsAtLevel(level int) int {
	base := 10
	for l := 2; l <= level; l++ {
		base += AbilityPointsPerLevel[l]
	}
	return base
}
