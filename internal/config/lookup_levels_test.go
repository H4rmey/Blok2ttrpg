package config

import "testing"

// The level budget accessors decide how many skill and perk points a character
// of a given level receives, so they are rules math even though they live in the
// config package. They had no tests: the only coverage was indirect, through
// handlers that happened to call them with in-range levels.
//
// These tests pin the three ways a budget can be resolved (explicit row,
// formula, fallback) and the clamping that protects the cap, because each is a
// separate branch and the precedence between them is the part most likely to be
// broken by a well-meaning edit.

// levelCfg builds a Config with both budget tables set from the formula.
func levelCfg(maxLevel int, skill, perk LevelTable) *Config {
	return &Config{Leveling: Leveling{
		MaxLevel:    maxLevel,
		SkillPoints: skill,
		PerkPoints:  perk,
	}}
}

func TestMaxLevel(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
		want int
	}{
		{
			name: "configured max_level wins",
			cfg:  levelCfg(12, LevelTable{Levels: []LevelEntry{{Level: 30, Total: 99}}}, LevelTable{}),
			want: 12,
		},
		{
			name: "falls back to the highest level in either table",
			cfg: levelCfg(0,
				LevelTable{Levels: []LevelEntry{{Level: 3, Total: 9}}},
				LevelTable{Levels: []LevelEntry{{Level: 7, Total: 4}}}),
			want: 7,
		},
		{
			name: "nothing configured still yields a usable clamp",
			cfg:  levelCfg(0, LevelTable{}, LevelTable{}),
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.MaxLevel(); got != tt.want {
				t.Errorf("MaxLevel() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestClampLevel(t *testing.T) {
	cfg := levelCfg(5, LevelTable{}, LevelTable{})
	tests := []struct {
		name  string
		level int
		want  int
	}{
		{"zero clamps up to the first level", 0, 1},
		{"negative clamps up to the first level", -4, 1},
		{"in range is unchanged", 3, 3},
		{"at the cap is unchanged", 5, 5},
		{"above the cap clamps down", 99, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cfg.ClampLevel(tt.level); got != tt.want {
				t.Errorf("ClampLevel(%d) = %d, want %d", tt.level, got, tt.want)
			}
		})
	}
}

// TestBudgetForLevelPrecedence covers the three resolution strategies and, most
// importantly, that an explicit row beats the formula. That precedence is what
// lets a ruleset bend the curve at one level, so reversing it would silently
// change every non-linear profile.
func TestBudgetForLevelPrecedence(t *testing.T) {
	formula := LevelTable{Start: 10, PerLevel: 5}
	override := LevelTable{Start: 10, PerLevel: 5, Levels: []LevelEntry{{Level: 3, Total: 999}}}
	rowsOnly := LevelTable{Levels: []LevelEntry{{Level: 1, Total: 4}, {Level: 5, Total: 20}}}

	tests := []struct {
		name  string
		table LevelTable
		level int
		want  int
	}{
		{"formula at level 1 is Start", formula, 1, 10},
		{"formula adds PerLevel per level after the first", formula, 4, 25},
		{"explicit row overrides the formula", override, 3, 999},
		{"formula still applies to levels without a row", override, 4, 25},
		{"without a formula an exact row is used", rowsOnly, 5, 20},
		{"without a formula it falls back to the highest row at or below", rowsOnly, 4, 4},
		{"level below 1 is treated as level 1", formula, 0, 10},
		{"an empty table yields zero", LevelTable{}, 3, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := budgetForLevel(tt.table, tt.level); got != tt.want {
				t.Errorf("budgetForLevel(%+v, %d) = %d, want %d", tt.table, tt.level, got, tt.want)
			}
		})
	}
}

// TestBudgetsClampBeforeResolving is the guard that matters most here: the
// public accessors must clamp first, so an out-of-range level can never be used
// to mint points beyond the configured cap.
func TestBudgetsClampBeforeResolving(t *testing.T) {
	cfg := levelCfg(3,
		LevelTable{Start: 10, PerLevel: 5},
		LevelTable{Start: 2, PerLevel: 1},
	)

	// Level 3 is the cap, so both level 3 and level 500 must agree.
	atCap := cfg.SkillPointBudget(3)
	beyond := cfg.SkillPointBudget(500)
	if atCap != beyond {
		t.Errorf("SkillPointBudget ignored the level cap: at cap = %d, beyond cap = %d", atCap, beyond)
	}
	if want := 20; atCap != want {
		t.Errorf("SkillPointBudget(3) = %d, want %d", atCap, want)
	}

	if got, want := cfg.PerkPointBudget(500), 4; got != want {
		t.Errorf("PerkPointBudget(500) = %d, want %d (the level-3 cap)", got, want)
	}
	// A level below the floor resolves as level 1 rather than underflowing.
	if got, want := cfg.PerkPointBudget(-10), 2; got != want {
		t.Errorf("PerkPointBudget(-10) = %d, want %d (the level-1 budget)", got, want)
	}
}
