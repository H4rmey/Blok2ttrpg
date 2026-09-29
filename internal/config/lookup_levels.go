// Level budgets: the skill-point and perk-point curves, and clamping a character level to the configured cap.
package config

// MaxLevel returns the highest level a character may reach. It prefers the
// configured leveling.max_level; when that is unset it falls back to the
// highest level present in either budget table, and finally to 1 so the value
// is always usable as a clamp.
func (c *Config) MaxLevel() int {
	if c.Leveling.MaxLevel > 0 {
		return c.Leveling.MaxLevel
	}
	max := 0
	for _, t := range []LevelTable{c.Leveling.SkillPoints, c.Leveling.PerkPoints} {
		for _, e := range t.Levels {
			if e.Level > max {
				max = e.Level
			}
		}
	}
	if max < 1 {
		return 1
	}
	return max
}

// ClampLevel constrains a level to the playable range [1, MaxLevel]. Callers
// normalise user input and imported data through this so a character can never
// sit above the configured cap.
func (c *Config) ClampLevel(level int) int {
	if level < 1 {
		return 1
	}
	if max := c.MaxLevel(); level > max {
		return max
	}
	return level
}

// SkillPointBudget returns the skill-point budget for a given character level.
// The level is clamped to the configured maximum first, so an out-of-range
// level never grants more than the cap.
func (c *Config) SkillPointBudget(level int) int {
	return budgetForLevel(c.Leveling.SkillPoints, c.ClampLevel(level))
}

// PerkPointBudget returns the perk-point budget for a given level.
func (c *Config) PerkPointBudget(level int) int {
	return budgetForLevel(c.Leveling.PerkPoints, c.ClampLevel(level))
}

// budgetForLevel resolves a level table. An explicit row for the requested
// level always wins, which lets a profile override the curve at specific
// levels. Otherwise the budget is derived from the table's formula,
// start + per_level * (level - 1), so the documented progression and the served
// numbers cannot drift apart. When neither is configured it falls back to the
// highest total at or below the requested level.
func budgetForLevel(t LevelTable, level int) int {
	if level < 1 {
		level = 1
	}
	best := 0
	for _, e := range t.Levels {
		if e.Level == level {
			return e.Total
		}
		if e.Level <= level && e.Total >= best {
			best = e.Total
		}
	}
	if t.Start > 0 || t.PerLevel > 0 {
		return t.Start + t.PerLevel*(level-1)
	}
	return best
}
