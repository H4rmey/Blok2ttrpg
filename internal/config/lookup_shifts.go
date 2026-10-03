// Enact Shift lookups: the play-card range for hand-applied skill shifts,
// read from the ruleset's Enact Shift enactment rather than hardcoded.
package config

// EnactShiftOptions returns the discrete non-zero shift values a hand-applied
// shift card may take, taken from the "shift" enactment's "shift-amount" field
// (its min/max/step range). Zero is skipped for the same reason the condition
// dropdowns skip it: a shift card at 0 changes nothing, so offering it would
// only invite dead cards. When the ruleset does not define the enactment or its
// field, the documented Enact Shift range (-6..6, step 1) is used as the
// fallback.
func (c *Config) EnactShiftOptions() []int {
	min, max, step := -6, 6, 1
	if comp, ok := c.ComponentByKind("enactment", "shift"); ok {
		for _, f := range comp.Fields {
			if f.Key != "shift-amount" {
				continue
			}
			// A 0/0 range reads back when the bounds were never written (or were
			// written as unset), which is the same "no range declared" verdict
			// ShiftOptionsFor uses for conditions: keep the documented fallback.
			if f.Min != 0 || f.Max != 0 {
				min, max = f.Min, f.Max
			}
			if f.Step > 0 {
				step = f.Step
			}
			break
		}
	}
	if step <= 0 {
		step = 1
	}

	var out []int
	for v := min; v <= max; v += step {
		if v != 0 {
			out = append(out, v)
		}
	}
	return out
}

// SmallestShift picks the magnitude a freshly applied card starts at: the
// smallest non-zero step in the range. Starting at the weakest end means
// applying a card never silently imposes the maximum penalty; the player has
// to choose that. Ties (a symmetric range holds both -1 and +1) keep the first
// candidate, mirroring the historical condition default.
func SmallestShift(opts []int) int {
	best := 0
	for _, v := range opts {
		if best == 0 || abs(v) < abs(best) {
			best = v
		}
	}
	return best
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
