package engine

import (
	"testing"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// testCfg builds a minimal ruleset: a four-rung proficiency ladder and three
// conditions covering the three shapes a condition can take (shiftable with
// skills, fixed with skills, and no skills at all).
func testCfg() *config.Config {
	return &config.Config{
		Proficiencies: []config.Proficiency{
			{ID: "inept"}, {ID: "untrained"}, {ID: "novice"}, {ID: "proficient"},
		},
		Conditions: []config.Condition{
			{
				ID: "frightened", Name: "Frightened",
				MinShift: -6, MaxShift: 0,
				AffectsSkills: []string{"offense.Strength", "general.Provoke"},
			},
			{
				ID: "encouraged", Name: "Encouraged",
				MinShift: 1, MaxShift: 6,
				AffectsSkills: []string{"offense.Strength"},
			},
			{
				ID: "slowed", Name: "Slowed",
				FixedShift:    -1,
				AffectsSkills: []string{"vital.Movement"},
			},
			{ID: "silenced", Name: "Silenced"},
		},
	}
}

// TestVitalCardReflectsConditionShift is the guard for the bug where the
// Movement skill row showed a shifted value while the vital card at the top of
// the sheet still showed the unshifted one. The card is the number a player
// actually reads in play, so the two must never disagree.
func TestVitalCardReflectsConditionShift(t *testing.T) {
	cfg := testCfg()
	// Movement is a vital skill, and the ladder grants a distinct value per rung
	// so a shift is visible in the rendered number.
	cfg.Skills.Order = []string{"vital"}
	cfg.Skills.Items = map[string][]string{"vital": {"Movement"}}
	cfg.Proficiencies = []config.Proficiency{
		{ID: "inept", Vitals: map[string]any{"movement": 2}},
		{ID: "untrained", Vitals: map[string]any{"movement": 3}},
		{ID: "novice", Vitals: map[string]any{"movement": 4}},
	}
	c := testChar(model.AppliedCondition{ID: "slowed"})
	c.Skills["vital.Movement"] = "untrained"

	vitals := CharacterVitals(cfg, c)
	if len(vitals) != 1 {
		t.Fatalf("got %d vitals, want 1", len(vitals))
	}
	v := vitals[0]
	if v.Max != "2" {
		t.Errorf("Max = %q, want %q: the card must show the shifted value", v.Max, "2")
	}
	if v.Base != "3" {
		t.Errorf("Base = %q, want %q: the card must still report the unshifted value", v.Base, "3")
	}
	if v.Shift != -1 || !v.Down() {
		t.Errorf("Shift = %d, want -1 and Down() true so the card can be coloured red", v.Shift)
	}
}

// TestUnshiftedVitalIsUnchanged checks the overlay is inert when no condition
// touches a vital: Max equals Base and nothing is marked.
func TestUnshiftedVitalIsUnchanged(t *testing.T) {
	cfg := testCfg()
	cfg.Skills.Order = []string{"vital"}
	cfg.Skills.Items = map[string][]string{"vital": {"Movement"}}
	cfg.Proficiencies = []config.Proficiency{
		{ID: "untrained", Vitals: map[string]any{"movement": 3}},
	}
	c := testChar()
	c.Skills["vital.Movement"] = "untrained"

	v := CharacterVitals(cfg, c)[0]
	if v.Shifted() || v.Max != v.Base || v.Max != "3" {
		t.Errorf("unshifted vital was modified: %+v", v)
	}
}

func testChar(conds ...model.AppliedCondition) model.Character {
	return model.Character{
		Skills: map[string]string{
			"offense.Strength": "novice",
			"general.Provoke":  "untrained",
			"vital.Movement":   "novice",
			"general.Stealth":  "novice",
		},
		Conditions: conds,
	}
}

// TestConditionsDoNotMutateStoredSkills is the load-bearing guarantee of the
// whole feature: the character's own proficiencies must be untouched by applying
// a condition. If this ever fails, removing a condition stops being reversible
// and characters silently rot.
func TestConditionsDoNotMutateStoredSkills(t *testing.T) {
	cfg := testCfg()
	c := testChar(model.AppliedCondition{ID: "frightened", Shift: -2})
	_ = EffectiveSkills(cfg, c)
	if got := c.Skills["offense.Strength"]; got != "novice" {
		t.Errorf("stored proficiency changed to %q; conditions must be a derived "+
			"overlay, never written back into Character.Skills", got)
	}
}

// TestShiftsStack pins the documented stacking rule: two conditions on the same
// skill sum rather than the strongest one winning.
func TestShiftsStack(t *testing.T) {
	cfg := testCfg()
	c := testChar(
		model.AppliedCondition{ID: "frightened", Shift: -1},
		model.AppliedCondition{ID: "frightened", Shift: -1},
	)
	if got := ConditionShifts(cfg, c)["offense.Strength"]; got != -2 {
		t.Errorf("stacked shift = %d, want -2 (shifts must sum)", got)
	}
}

// TestMirroredConditionsCancel checks that a debuff and its mirror at equal
// magnitude leave the skill unmarked, rather than colouring it for a net change
// of zero.
func TestMirroredConditionsCancel(t *testing.T) {
	cfg := testCfg()
	c := testChar(
		model.AppliedCondition{ID: "frightened", Shift: -2},
		model.AppliedCondition{ID: "encouraged", Shift: 2},
	)
	view := EffectiveSkill(cfg, c, "offense", "Strength")
	if view.Shifted() {
		t.Errorf("net shift is zero but the skill still reports Shift=%d", view.Shift)
	}
	if view.Effective != view.Base {
		t.Errorf("effective %q != base %q for a net-zero shift", view.Effective, view.Base)
	}
}

// TestFixedConditionUsesConfiguredShift checks that a fixed condition ignores
// the stored Shift (which the UI never sets for it) and uses its own magnitude.
func TestFixedConditionUsesConfiguredShift(t *testing.T) {
	cfg := testCfg()
	c := testChar(model.AppliedCondition{ID: "slowed", Shift: 99})
	if got := ConditionShifts(cfg, c)["vital.Movement"]; got != -1 {
		t.Errorf("fixed condition shift = %d, want -1 (fixed_shift must win over "+
			"any stored value)", got)
	}
}

// TestConditionWithNoSkillsShiftsNothing checks that a condition which only
// changes what a character may do leaves every number alone.
func TestConditionWithNoSkillsShiftsNothing(t *testing.T) {
	cfg := testCfg()
	c := testChar(model.AppliedCondition{ID: "silenced"})
	if shifts := ConditionShifts(cfg, c); len(shifts) != 0 {
		t.Errorf("condition affecting no skills produced shifts: %v", shifts)
	}
}

// TestClampReportedNotHidden checks that a shift the ladder cannot absorb is
// reported as clamped while still showing the requested magnitude. Reporting the
// achieved amount instead would make a third debuff look like it did nothing.
func TestClampReportedNotHidden(t *testing.T) {
	cfg := testCfg()
	c := testChar(model.AppliedCondition{ID: "frightened", Shift: -6})
	view := EffectiveSkill(cfg, c, "offense", "Strength")
	if !view.Clamped {
		t.Error("a -6 shift on the second rung of a four-rung ladder must report Clamped")
	}
	if view.Shift != -6 {
		t.Errorf("Shift = %d, want the requested -6 so the badge stays honest", view.Shift)
	}
	if view.Effective != "inept" {
		t.Errorf("Effective = %q, want the bottom rung %q", view.Effective, "inept")
	}
}

// TestUnknownConditionIsIgnored checks that dropping a condition from the
// ruleset does not make a saved character unreadable.
func TestUnknownConditionIsIgnored(t *testing.T) {
	cfg := testCfg()
	c := testChar(model.AppliedCondition{ID: "no_such_condition", Shift: -3})
	if shifts := ConditionShifts(cfg, c); len(shifts) != 0 {
		t.Errorf("unknown condition produced shifts: %v", shifts)
	}
}

// TestUnaffectedSkillsStillPresent checks that EffectiveSkills returns an entry
// for every stored skill, so a template can render the grid from one map.
func TestUnaffectedSkillsStillPresent(t *testing.T) {
	cfg := testCfg()
	c := testChar(model.AppliedCondition{ID: "frightened", Shift: -1})
	views := EffectiveSkills(cfg, c)
	v, ok := views["general.Stealth"]
	if !ok {
		t.Fatal("an unaffected skill is missing from the view map")
	}
	if v.Shifted() || v.Effective != "novice" {
		t.Errorf("unaffected skill was modified: %+v", v)
	}
}

// The applied-shift card fixtures. The shift tests need a config that knows its
// skills (SkillShifts refuses a card whose key no ruleset defines), so each
// test that exercises a card adds the "offense" group it targets rather than
// widening the shared testCfg fixture.
func shiftCfg() *config.Config {
	cfg := testCfg()
	cfg.Skills.Order = []string{"offense", "vital"}
	cfg.Skills.Items = map[string][]string{
		"offense": {"Strength"},
		"vital":   {"Movement"},
	}
	return cfg
}

func shiftChar(shifts ...model.AppliedShift) model.Character {
	c := testChar()
	c.Shifts = shifts
	return c
}

// TestAppliedShiftMovesEffectiveSkill checks the core behaviour of the Enact
// Shift play card: one card on one skill moves that skill's effective reading
// while the stored proficiency stays untouched.
func TestAppliedShiftMovesEffectiveSkill(t *testing.T) {
	cfg := shiftCfg()
	c := shiftChar(model.AppliedShift{SkillKey: "offense.Strength", Shift: -1})

	view := EffectiveSkill(cfg, c, "offense", "Strength")
	if !view.Shifted() || view.Shift != -1 || !view.Down() {
		t.Fatalf("view = %+v, want a -1 shift", view)
	}
	if view.Effective != "untrained" {
		t.Errorf("Effective = %q, want the rung one below novice %q", view.Effective, "untrained")
	}
	if got := c.Skills["offense.Strength"]; got != "novice" {
		t.Errorf("stored proficiency changed to %q; shift cards must be a derived "+
			"overlay, never written back into Character.Skills", got)
	}
}

// TestConditionAndShiftCardStack is the bridge between the two overlays: a
// condition and a shift card on the same skill must behave as one summed
// effect, including cancelling each other exactly. Encouraged is used rather
// than Frightened because the fixture's Frightened moves two skills, which
// would make "everything cancelled" an assertion about a different skill.
func TestConditionAndShiftCardStack(t *testing.T) {
	cfg := shiftCfg()
	c := testChar(model.AppliedCondition{ID: "encouraged", Shift: 2})
	c.Shifts = []model.AppliedShift{{SkillKey: "offense.Strength", Shift: -1}}

	if got := SkillShifts(cfg, c)["offense.Strength"]; got != 1 {
		t.Errorf("shifts summed to %d, want 1 (encouraged +2, card -1)", got)
	}

	c.Shifts = []model.AppliedShift{{SkillKey: "offense.Strength", Shift: -2}}
	if shifts := SkillShifts(cfg, c); len(shifts) != 0 {
		t.Errorf("net-zero overlay still produced shifts: %v", shifts)
	}
	view := EffectiveSkill(cfg, c, "offense", "Strength")
	if view.Shifted() || view.Effective != view.Base {
		t.Errorf("a cancelled overlay must leave the skill unmarked: %+v", view)
	}
}

// TestShiftCardsOnSameSkillStack mirrors TestShiftsStack for cards: two cards
// on one skill each count, because each has its own remove button and the
// player is meant to be able to revise one without rewriting the other.
func TestShiftCardsOnSameSkillStack(t *testing.T) {
	cfg := shiftCfg()
	c := shiftChar(
		model.AppliedShift{SkillKey: "offense.Strength", Shift: 1},
		model.AppliedShift{SkillKey: "offense.Strength", Shift: 1},
	)
	if got := SkillShifts(cfg, c)["offense.Strength"]; got != 2 {
		t.Errorf("stacked card shift = %d, want 2", got)
	}
}

// TestShiftCardToUnknownSkillIgnored checks that a card whose skill key the
// ruleset no longer defines shifts nothing and never makes a saved character
// unopenable - the same degradation an unknown condition id gets.
func TestShiftCardToUnknownSkillIgnored(t *testing.T) {
	cfg := shiftCfg()
	c := shiftChar(
		model.AppliedShift{SkillKey: "offense.Gone", Shift: 3},
		model.AppliedShift{SkillKey: "", Shift: 3},
	)
	if shifts := SkillShifts(cfg, c); len(shifts) != 0 {
		t.Errorf("unknown shift card produced shifts: %v", shifts)
	}
}

// TestAppliedShiftFeedsVitals guards the vitals path: Shifts comes straight
// from ConditionShifts, so a shift card must arrive at the vital card the same
// way a Slowed condition does.
func TestAppliedShiftFeedsVitals(t *testing.T) {
	cfg := shiftCfg()
	cfg.Proficiencies = []config.Proficiency{
		{ID: "inept", Vitals: map[string]any{"movement": 2}},
		{ID: "untrained", Vitals: map[string]any{"movement": 3}},
		{ID: "novice", Vitals: map[string]any{"movement": 4}},
	}
	c := shiftChar(model.AppliedShift{SkillKey: "vital.Movement", Shift: -1})
	c.Skills["vital.Movement"] = "novice"

	v := CharacterVitals(cfg, c)[0]
	if v.Max != "3" {
		t.Errorf("Max = %q, want %q: the card must show the shifted value", v.Max, "3")
	}
	if v.Base != "4" {
		t.Errorf("Base = %q, want %q: the card must still report the unshifted value", v.Base, "4")
	}
	if v.Shift != -1 || !v.Down() {
		t.Errorf("Shift = %d, want -1 and Down() true", v.Shift)
	}
}
