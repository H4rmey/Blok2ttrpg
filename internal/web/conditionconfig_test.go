package web

import (
	"testing"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
)

// TestShippedConditionsResolveTheirSkills guards the shipped profile against the
// failure mode that a typo in affects_skills produces: a condition that looks
// applied but colours nothing. The loader rejects it, so this is really a check
// that the profile still loads at all - but it fails with a message naming the
// offending key, which a bare load test would not.
func TestShippedConditionsResolveTheirSkills(t *testing.T) {
	cfg, err := config.Load("../../config/Blok2Simplified")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	known := map[string]bool{}
	for _, g := range cfg.Skills.List() {
		for _, s := range g.Skills {
			known[g.ID+"."+s] = true
		}
	}
	for _, c := range cfg.Conditions {
		for _, key := range c.AffectsSkills {
			if !known[key] {
				t.Errorf("condition %q affects unknown skill %q", c.ID, key)
			}
		}
	}
}

// TestConditionsThatReadAsShiftsDeclareSkills catches a condition whose text
// promises a die shift while its config declares none. Such a condition silently
// does nothing on the character sheet: the player reads "shifted one down", sees
// no skill change, and the two disagree. The check is a heuristic on the wording
// rather than a rule, which is exactly why it is a test and not loader
// validation - a false positive here is a prompt to reword, not a broken config.
func TestConditionsThatReadAsShiftsDeclareSkills(t *testing.T) {
	cfg, err := config.Load("../../config/Blok2Simplified")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	// Conditions legitimately described in terms of shifts without owning any
	// themselves. Each needs a reason, not just an entry.
	exempt := map[string]string{
		// The item decides which rolls it covers, so no fixed skill set exists.
		"broken_gear":    "which gear is affected is a table decision",
		"amplified_gear": "which gear is affected is a table decision",
		// These rewrite how other shifts resolve; they are not a shift.
		"cursed":  "converts other shifts rather than applying one",
		"blessed": "converts other shifts rather than applying one",
		// Conditional on a target change, so it cannot be a standing shift.
		"distracted": "applies only against a newly chosen target",
		// Applies to whoever attacks the target, not to the target's own skills.
		"fragile": "shifts the attacker's die, not the target's skills",
	}
	for _, c := range cfg.Conditions {
		if c.ShiftsSkills() {
			continue
		}
		if _, ok := exempt[c.ID]; ok {
			continue
		}
		if containsFold(c.Description, "shift") {
			t.Errorf("condition %q describes a shift (%q) but declares no "+
				"affects_skills, so applying it changes nothing on the sheet. "+
				"Either give it affects_skills, or reword it, or add it to the "+
				"exempt list with a reason.", c.ID, c.Description)
		}
	}
}

// TestFixedShiftConditionsNameSkills is the mirror guard: a fixed_shift with no
// skills to apply it to is dead config.
func TestFixedShiftConditionsNameSkills(t *testing.T) {
	cfg, err := config.Load("../../config/Blok2Simplified")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	for _, c := range cfg.Conditions {
		if c.FixedShift != 0 && !c.ShiftsSkills() {
			t.Errorf("condition %q sets fixed_shift %d but names no skills to apply it to",
				c.ID, c.FixedShift)
		}
	}
}

// TestConditionsDoNotShiftHPOrEnergy pins the design decision that a temporary
// state may slow a character but must never change the size of a pool they
// bought with skill points.
func TestConditionsDoNotShiftHPOrEnergy(t *testing.T) {
	cfg, err := config.Load("../../config/Blok2Simplified")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	for _, c := range cfg.Conditions {
		for _, key := range c.AffectsSkills {
			if key == "vital.HP" || key == "vital.Energy" {
				t.Errorf("condition %q shifts %q; conditions may only shift vital.Movement",
					c.ID, key)
			}
		}
	}
}

// containsFold is a small case-insensitive substring test, kept local so the
// heuristic above does not pull a dependency into the package for one call.
func containsFold(haystack, needle string) bool {
	h, n := lower(haystack), lower(needle)
	if len(n) == 0 || len(n) > len(h) {
		return len(n) == 0
	}
	for i := 0; i+len(n) <= len(h); i++ {
		if h[i:i+len(n)] == n {
			return true
		}
	}
	return false
}

func lower(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 'a' - 'A'
		}
	}
	return string(b)
}
