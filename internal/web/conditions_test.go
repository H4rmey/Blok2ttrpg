package web

import (
	"testing"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
)

// TestShiftableConditionsHaveShiftCost guards a trap in how conditions are
// priced. A condition is "shiftable" when it declares a min_shift/max_shift
// range, and a shiftable condition is charged shift_cost per unit of shift -
// its build_cost is never read. So a shiftable condition that declares
// build_cost but no shift_cost is silently free, no matter how strong it is.
func TestShiftableConditionsHaveShiftCost(t *testing.T) {
	cfg, err := config.Load("../../config/Blok2Simplified")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if len(cfg.Conditions) == 0 {
		t.Fatal("no conditions configured")
	}
	for _, c := range cfg.Conditions {
		if !c.Shiftable() {
			continue
		}
		if c.ShiftCost.BuildCost == 0 && c.ShiftCost.EnergyCost == 0 {
			t.Errorf("condition %q is shiftable but has no shift_cost, so it is free "+
				"(its build_cost of %d is ignored for shiftable conditions)",
				c.ID, c.BuildCost)
		}
	}
}

// TestConditionPricingOrder pins the relationships where one condition is
// strictly stronger than another, so it must not cost the same or less. These
// are the orderings that can be argued from the rules text alone rather than
// from taste, which makes them safe to assert.
func TestConditionPricingOrder(t *testing.T) {
	cfg, err := config.Load("../../config/Blok2Simplified")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	cost := func(id string) int {
		c, ok := cfg.ConditionByID(id)
		if !ok {
			t.Fatalf("condition %q not found", id)
		}
		return c.BuildCost
	}

	stronger := []struct{ more, less, why string }{
		{"paralyzed", "stunned", "losing your whole turn is worse than losing one action"},
		{"untargetable", "ignored", "cannot be targeted at all beats cannot be singled out"},
		{"invincible", "paralyzed", "total damage immunity outweighs losing a turn"},
		{"encouraged", "frightened", "a buff on an ally is worth more than a debuff on one enemy"},
		{"blessed", "cursed", "same, for die-shift conversion"},
		{"energized", "fatigued", "same, for the energy discount"},
		{"amplified_gear", "broken_gear", "same, for gear die shifts"},
	}
	for _, s := range stronger {
		// Encouraged/Frightened are shiftable, so compare their per-shift cost.
		if s.more == "encouraged" {
			e, _ := cfg.ConditionByID("encouraged")
			f, _ := cfg.ConditionByID("frightened")
			if e.ShiftCost.BuildCost <= f.ShiftCost.BuildCost {
				t.Errorf("encouraged (%d/shift) should cost more than frightened (%d/shift): %s",
					e.ShiftCost.BuildCost, f.ShiftCost.BuildCost, s.why)
			}
			continue
		}
		if cost(s.more) <= cost(s.less) {
			t.Errorf("%s (%d) should cost more than %s (%d): %s",
				s.more, cost(s.more), s.less, cost(s.less), s.why)
		}
	}
}

// TestFixedConditionsHaveCost checks the other half: a non-shiftable condition
// is charged its flat build/energy cost, so a zero cost means it is free.
func TestFixedConditionsHaveCost(t *testing.T) {
	cfg, err := config.Load("../../config/Blok2Simplified")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	for _, c := range cfg.Conditions {
		if c.Shiftable() {
			continue
		}
		// A non-selectable condition cannot be bought at all (Dying is a state
		// you end up in, not an effect you pay for), so a zero cost is fine.
		if !c.IsSelectable() {
			continue
		}
		if c.BuildCost == 0 && c.EnergyCost == 0 {
			t.Errorf("condition %q costs nothing", c.ID)
		}
	}
}

// TestNonSelectableConditionsAreHidden checks that selectable: false keeps a
// condition out of the dropdown while leaving it resolvable by id, so saved
// perks and the generated instructions can still name it.
func TestNonSelectableConditionsAreHidden(t *testing.T) {
	cfg, err := config.Load("../../config/Blok2Simplified")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	offered := map[string]bool{}
	for _, o := range cfg.OptionsFor("conditions") {
		offered[o.Value] = true
	}
	hidden := 0
	for _, c := range cfg.Conditions {
		if c.IsSelectable() {
			if !offered[c.ID] {
				t.Errorf("selectable condition %q is missing from the dropdown", c.ID)
			}
			continue
		}
		hidden++
		if offered[c.ID] {
			t.Errorf("condition %q is selectable: false but still offered in the dropdown", c.ID)
		}
		if _, ok := cfg.ConditionByID(c.ID); !ok {
			t.Errorf("condition %q is no longer resolvable by id", c.ID)
		}
	}
	if hidden == 0 {
		t.Error("no conditions are marked selectable: false, so this guard tests nothing")
	}
}

// TestConditionOptionsCarryDescription checks that every condition option
// exposes its description as hover text in the builder dropdown.
func TestConditionOptionsCarryDescription(t *testing.T) {
	cfg, err := config.Load("../../config/Blok2Simplified")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	byID := map[string]config.Condition{}
	for _, c := range cfg.Conditions {
		byID[c.ID] = c
	}
	for _, o := range cfg.OptionsFor("conditions") {
		c := byID[o.Value]
		if c.Description == "" {
			t.Errorf("condition %q has no description to show on hover", o.Value)
			continue
		}
		if o.Information != c.Description {
			t.Errorf("condition %q option hover text is %q, want the description %q",
				o.Value, o.Information, c.Description)
		}
	}
}
