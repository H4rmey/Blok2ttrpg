package config

import "testing"

// The proficiency ladder backs two user-visible behaviours: conditions moving a
// skill up or down, and the non-blocking "this shift was capped" warning on the
// character sheet. Both read from the helpers here, and neither had a test.
//
// The clamping is the interesting part. ShiftProficiency and ShiftClamped have
// to agree: the first decides where a skill lands, the second decides whether
// the sheet says the shift could not be applied in full. If they ever disagree,
// the sheet either warns about a shift that worked or silently swallows one that
// did not, so there is a test below that asserts they stay consistent across the
// whole ladder rather than only at hand-picked points.

// ladderCfg builds a four-rung ladder: none -> trained -> expert -> master.
func ladderCfg(defaultID string) *Config {
	return &Config{
		DefaultProficiency: defaultID,
		Proficiencies: []Proficiency{
			{ID: "none", Name: "Untrained", Cost: 0},
			{ID: "trained", Name: "Trained", Cost: 2},
			{ID: "expert", Name: "Expert", Cost: 5},
			{ID: "master", Name: "Master", Cost: 9},
		},
	}
}

func TestProficiencyLookup(t *testing.T) {
	cfg := ladderCfg("")

	if p, ok := cfg.Proficiency("expert"); !ok {
		t.Error("Proficiency(expert) reported not found")
	} else if p.Cost != 5 {
		t.Errorf("Proficiency(expert).Cost = %d, want 5", p.Cost)
	}

	if _, ok := cfg.Proficiency("nonexistent"); ok {
		t.Error("Proficiency reported an unknown id as found")
	}

	// An unknown id must price as free rather than panicking or guessing, so a
	// character carrying a tier a ruleset has dropped still loads.
	if got := cfg.ProficiencyCost("nonexistent"); got != 0 {
		t.Errorf("ProficiencyCost(unknown) = %d, want 0", got)
	}
	if got := cfg.ProficiencyCost("master"); got != 9 {
		t.Errorf("ProficiencyCost(master) = %d, want 9", got)
	}
}

func TestDefaultProficiency(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *Config
		wantID    string
		wantIndex int
	}{
		{"configured default is used", ladderCfg("trained"), "trained", 1},
		{"unset default falls back to the first rung", ladderCfg(""), "none", 0},
		{"invalid default falls back to the first rung", ladderCfg("bogus"), "none", 0},
		{"an empty ladder yields no default", &Config{}, "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.DefaultProficiencyID(); got != tt.wantID {
				t.Errorf("DefaultProficiencyID() = %q, want %q", got, tt.wantID)
			}
			if got := tt.cfg.DefaultProficiencyIndex(); got != tt.wantIndex {
				t.Errorf("DefaultProficiencyIndex() = %d, want %d", got, tt.wantIndex)
			}
		})
	}
}

func TestShiftProficiency(t *testing.T) {
	cfg := ladderCfg("none")
	tests := []struct {
		name    string
		current string
		delta   int
		want    string
	}{
		{"up one rung", "none", 1, "trained"},
		{"up several rungs", "none", 2, "expert"},
		{"down one rung", "expert", -1, "trained"},
		{"no shift is a no-op", "trained", 0, "trained"},
		{"overshooting the top clamps to the last rung", "expert", 99, "master"},
		{"overshooting the bottom clamps to the first rung", "trained", -99, "none"},
		{"an unknown current id is treated as the first rung", "bogus", 1, "trained"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cfg.ShiftProficiency(tt.current, tt.delta); got != tt.want {
				t.Errorf("ShiftProficiency(%q, %d) = %q, want %q", tt.current, tt.delta, got, tt.want)
			}
		})
	}

	// With no ladder configured there is nothing to shift onto, so the input has
	// to come back untouched instead of becoming "".
	empty := &Config{}
	if got := empty.ShiftProficiency("whatever", 3); got != "whatever" {
		t.Errorf("ShiftProficiency on an empty ladder = %q, want the input back", got)
	}
}

func TestShiftClamped(t *testing.T) {
	cfg := ladderCfg("none")
	tests := []struct {
		name    string
		current string
		delta   int
		want    bool
	}{
		{"a shift that fits is not clamped", "none", 1, false},
		{"landing exactly on the top rung is not clamped", "none", 3, false},
		{"running off the top is clamped", "expert", 2, true},
		{"running off the bottom is clamped", "trained", -2, true},
		{"a zero shift is never clamped", "master", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cfg.ShiftClamped(tt.current, tt.delta); got != tt.want {
				t.Errorf("ShiftClamped(%q, %d) = %v, want %v", tt.current, tt.delta, got, tt.want)
			}
		})
	}

	if cfg := (&Config{}); cfg.ShiftClamped("x", 5) {
		t.Error("ShiftClamped on an empty ladder should report false, not a clamp")
	}
}

// TestShiftClampedAgreesWithShiftProficiency is the consistency guard. For every
// rung and every plausible delta, "the shift was clamped" must mean exactly "the
// skill did not move as far as asked". Asserting the relationship rather than
// individual values means a change to either function that breaks the pairing
// fails here, even if both still look correct in isolation.
func TestShiftClampedAgreesWithShiftProficiency(t *testing.T) {
	cfg := ladderCfg("none")
	last := len(cfg.Proficiencies) - 1

	for from := 0; from <= last; from++ {
		current := cfg.Proficiencies[from].ID
		for delta := -(last + 2); delta <= last+2; delta++ {
			if delta == 0 {
				continue
			}
			landed := cfg.ShiftProficiency(current, delta)
			landedIdx := cfg.proficiencyIndex(landed)

			// What an unclamped shift would have produced.
			wantIdx := from + delta
			fullyApplied := landedIdx == wantIdx

			if got := cfg.ShiftClamped(current, delta); got == fullyApplied {
				t.Errorf("from %q delta %+d: landed on %q (index %d, requested %d), "+
					"but ShiftClamped reported %v",
					current, delta, landed, landedIdx, wantIdx, got)
			}
		}
	}
}
