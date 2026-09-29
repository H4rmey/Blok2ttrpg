// Proficiency ladder lookups: resolving tiers, the default rung, and moving a skill up or down the ladder with clamping.
package config

// Proficiency returns the proficiency tier with the given id.
func (c *Config) Proficiency(id string) (Proficiency, bool) {

	for _, p := range c.Proficiencies {
		if p.ID == id {
			return p, true
		}
	}
	return Proficiency{}, false
}

// ProficiencyCost returns the skill-point cost of a proficiency id (0 if none).
func (c *Config) ProficiencyCost(id string) int {
	if p, ok := c.Proficiency(id); ok {
		return p.Cost
	}
	return 0
}

// DefaultProficiencyID returns the id of the tier new characters start every
// skill at. When default_proficiency is configured and valid it is used;
// otherwise the first proficiency in the ladder is the default.
func (c *Config) DefaultProficiencyID() string {
	if c.DefaultProficiency != "" {
		if c.proficiencyIndex(c.DefaultProficiency) >= 0 {
			return c.DefaultProficiency
		}
	}
	if len(c.Proficiencies) > 0 {
		return c.Proficiencies[0].ID
	}
	return ""
}

// DefaultProficiencyIndex returns the ladder position of the default tier, or 0
// when it cannot be resolved.
func (c *Config) DefaultProficiencyIndex() int {
	if idx := c.proficiencyIndex(c.DefaultProficiencyID()); idx >= 0 {
		return idx
	}
	return 0
}

// proficiencyIndex returns the position of a proficiency id within the ordered
// ladder, or -1 when it is not found.
func (c *Config) proficiencyIndex(id string) int {
	for i, p := range c.Proficiencies {
		if p.ID == id {
			return i
		}
	}
	return -1
}

// ShiftProficiency moves a proficiency id up or down the ordered ladder by
// delta rungs and returns the resulting id. The result is clamped to the ends
// of the ladder. An unknown current id is treated as the first (default) tier
// so a shift still produces a sensible result.
func (c *Config) ShiftProficiency(current string, delta int) string {
	if len(c.Proficiencies) == 0 {
		return current
	}
	idx := c.proficiencyIndex(current)
	if idx < 0 {
		idx = 0
	}
	idx += delta
	if idx < 0 {
		idx = 0
	}
	if idx > len(c.Proficiencies)-1 {
		idx = len(c.Proficiencies) - 1
	}
	return c.Proficiencies[idx].ID
}

// ShiftClamped reports whether shifting the given proficiency id by delta
// rungs would run off either end of the ladder (i.e. the requested delta could
// not be fully applied). It is used to surface a non-blocking warning when a
// package pushes a skill above or below the possible range.
func (c *Config) ShiftClamped(current string, delta int) bool {
	if len(c.Proficiencies) == 0 || delta == 0 {
		return false
	}
	idx := c.proficiencyIndex(current)
	if idx < 0 {
		idx = 0
	}
	target := idx + delta
	return target < 0 || target > len(c.Proficiencies)-1
}
