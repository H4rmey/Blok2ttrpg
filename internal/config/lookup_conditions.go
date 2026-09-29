// Condition lookups: resolving conditions by id and reporting the discrete shift magnitudes a shiftable condition may take.
package config

// GeneralConditionByID returns the general condition with the given id.
func (c *Config) GeneralConditionByID(id string) (GeneralCondition, bool) {
	for _, s := range c.GeneralConditions {
		if s.ID == id {
			return s, true
		}
	}
	return GeneralCondition{}, false
}

// SpecificConditionByID returns the specific condition with the given id.
func (c *Config) SpecificConditionByID(id string) (SpecificCondition, bool) {
	for _, s := range c.SpecificConditions {
		if s.ID == id {
			return s, true
		}
	}
	return SpecificCondition{}, false
}

// ConditionByID returns the unified condition with the given id.
func (c *Config) ConditionByID(id string) (Condition, bool) {
	for _, s := range c.Conditions {
		if s.ID == id {
			return s, true
		}
	}
	return Condition{}, false
}

// ShiftOptionsFor returns the discrete non-zero shift values a general condition
// may take, from its configured min_shift..max_shift range. Zero is skipped
// because applying a condition with no shift is meaningless.
func (c *Config) ShiftOptionsFor(generalID string) []int {
	var min, max int
	if s, ok := c.GeneralConditionByID(generalID); ok {
		min, max = s.MinShift, s.MaxShift
	} else if u, ok := c.ConditionByID(generalID); ok {
		if !u.Shiftable() {
			// Fixed-cost unified condition: no shift dropdown.
			return nil
		}
		min, max = u.MinShift, u.MaxShift
	} else {
		return nil
	}
	if min == 0 && max == 0 {
		min, max = -6, 6
	}

	var out []int
	for v := min; v <= max; v++ {
		if v != 0 {
			out = append(out, v)
		}
	}
	return out
}
