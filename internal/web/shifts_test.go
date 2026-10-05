package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/engine"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
	"github.com/harmey/blok2ttrpg-v5/internal/store"
)

// testAppWithBlankChar builds an app plus a persisted blank character. The
// Enact Shift handlers always run against a character loaded from the store
// (the dispatcher does the loading), so the fixture is saved before use.
func testAppWithBlankChar(t *testing.T) (*App, *model.Character) {
	t.Helper()
	cfg, err := config.Load("../../config/Blok2Simplified")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	st, err := store.New(filepath.Join(t.TempDir(), "characters.json"))
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	app, err := NewApp(cfg, st, "../../templates", "../../library")
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	c := app.blankCharacter("char-1")
	c.Traits["name"] = "Tester"
	if err := st.Save(c); err != nil {
		t.Fatalf("save character: %v", err)
	}
	return app, &c
}

// TestEnactShiftOptionsCoverShippedProfile pins the shipped ruleset's Enact
// Shift range: the dropdown reaches both ends of the config and never offers a
// shift of 0.
func TestEnactShiftOptionsCoverShippedProfile(t *testing.T) {
	cfg, err := config.Load("../../config/Blok2Simplified")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	opts := cfg.Config.EnactShiftOptions()
	if opts[0] != -6 || opts[len(opts)-1] != 6 {
		t.Errorf("Enact Shift range = %v, want -6..6 skipping 0", opts)
	}
	for _, v := range opts {
		if v == 0 {
			t.Errorf("0 must not be offered as a shift card value: %v", opts)
		}
	}
}

// TestApplyShiftAddsCard walks the Apply Shift flow end to end: the picker's
// plain form POST lands, a card is added at the weakest magnitude (the shipped
// range is symmetric, so the default is -1), and the redirect re-opens the
// sheet.
func TestApplyShiftAddsCard(t *testing.T) {
	app, c := testAppWithBlankChar(t)
	req := httptest.NewRequest(http.MethodPost, "/characters/char-1/shifts/apply",
		strings.NewReader(url.Values{"skill_key": {"offense.Strength"}}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	app.handleCharacterShifts(rec, req, c, []string{"apply"})

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("apply returned %d, want 303: %s", rec.Code, rec.Body.String())
	}
	// The shift strip lives on the Skills tab, so the redirect carries the tab
	// hash to come back to it rather than resetting the sheet to Traits.
	if rec.Header().Get("Location") != "/characters/char-1?t=skills#tab-skills" {
		t.Errorf("Location = %q, want the sheet's skills tab", rec.Header().Get("Location"))
	}
	if len(c.Shifts) != 1 || c.Shifts[0].SkillKey != "offense.Strength" {
		t.Fatalf("shifts = %+v, want one offense.Strength card", c.Shifts)
	}
	if want := config.SmallestShift(app.Cfg.Config.EnactShiftOptions()); c.Shifts[0].Shift != want {
		t.Errorf("default shift = %d, want the weakest offered magnitude %d", c.Shifts[0].Shift, want)
	}
}

// TestApplyShiftRejectsUnknownSkill checks the server-side guard: a crafted
// form post naming a key no rule defines must be refused, not merely absent
// from the picker.
func TestApplyShiftRejectsUnknownSkill(t *testing.T) {
	app, c := testAppWithBlankChar(t)
	for _, key := range []string{"offense.Blasted", "not-a-key", ""} {
		req := httptest.NewRequest(http.MethodPost, "/characters/char-1/shifts/apply",
			strings.NewReader(url.Values{"skill_key": {key}}.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		app.handleCharacterShifts(rec, req, c, []string{"apply"})
		if rec.Code != http.StatusBadRequest {
			t.Errorf("apply with key %q returned %d, want 400", key, rec.Code)
		}
	}
	if len(c.Shifts) != 0 {
		t.Errorf("a rejected apply must not add a card, got %+v", c.Shifts)
	}
}

// TestSetShiftValidatesAgainstConfigRange checks both halves of value editing:
// a configured magnitude is persisted, and one outside the Enact Shift range is
// refused rather than stored.
func TestSetShiftValidatesAgainstConfigRange(t *testing.T) {
	app, c := testAppWithBlankChar(t)
	c.Shifts = []model.AppliedShift{{SkillKey: "offense.Strength", Shift: -1}}

	post := func(shift string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/characters/char-1/shifts/0/shift",
			strings.NewReader(url.Values{"shift": {shift}}.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		app.handleCharacterShifts(rec, req, c, []string{"0", "shift"})
		return rec
	}

	if rec := post("3"); rec.Code != http.StatusOK || rec.Header().Get("HX-Redirect") != "/characters/char-1?t=skills#tab-skills" {
		t.Fatalf("setting 3 returned %d, want 200 with HX-Redirect: %s", rec.Code, rec.Body.String())
	}
	if c.Shifts[0].Shift != 3 {
		t.Errorf("stored shift = %d, want 3", c.Shifts[0].Shift)
	}
	// 0 is skipped from the offered range by design, so it must be refused as a
	// stored value.
	for _, bad := range []string{"0", "7", "-7"} {
		if rec := post(bad); rec.Code != http.StatusBadRequest {
			t.Errorf("setting %q returned %d, want 400 (outside the offered range)", bad, rec.Code)
		}
	}
	if c.Shifts[0].Shift != 3 {
		t.Errorf("a rejected set must not change the stored value, got %d", c.Shifts[0].Shift)
	}
}

// TestRemoveShiftByIndex checks removal is exact by position: with two cards on
// the same skill (the legal duplicate case there is no id to disambiguate),
// removing one must leave the other intact.
func TestRemoveShiftByIndex(t *testing.T) {
	app, c := testAppWithBlankChar(t)
	c.Shifts = []model.AppliedShift{
		{SkillKey: "offense.Strength", Shift: 1},
		{SkillKey: "offense.Strength", Shift: 2},
		{SkillKey: "vital.Movement", Shift: -1},
	}
	req := httptest.NewRequest(http.MethodDelete, "/characters/char-1/shifts/1", nil)
	rec := httptest.NewRecorder()
	app.handleCharacterShifts(rec, req, c, []string{"1"})

	if rec.Header().Get("HX-Redirect") != "/characters/char-1?t=skills#tab-skills" {
		t.Errorf("remove must HX-Redirect to the sheet's skills tab, got %q", rec.Header().Get("HX-Redirect"))
	}
	if len(c.Shifts) != 2 || c.Shifts[0].Shift != 1 || c.Shifts[1].SkillKey != "vital.Movement" {
		t.Fatalf("removal removed the wrong card(s): %+v", c.Shifts)
	}
}

// TestShiftLibraryOffersEverySkill checks the picker lists every configured
// skill with no exceptions - including the vital group - grouped by the same
// group labels the sheet uses, so "Magic" arrives unambiguous.
func TestShiftLibraryOffersEverySkill(t *testing.T) {
	app, _ := testAppWithBlankChar(t)
	req := httptest.NewRequest(http.MethodGet, "/shifts/library?character=char-1", nil)
	rec := httptest.NewRecorder()
	app.handleShiftLibrary(rec, req)
	body := rec.Body.String()

	for _, group := range app.Cfg.Skills.List() {
		if !strings.Contains(body, group.Label) {
			t.Errorf("picker is missing the %q group label", group.Label)
		}
		for _, s := range group.Skills {
			key := model.SkillKey(group.ID, s)
			if !strings.Contains(body, `value="`+key+`"`) {
				t.Errorf("picker is missing skill option %q", key)
			}
		}
	}
}

// TestSheetRendersShiftCard pins what a shift card is for: the strip lists the
// card and, on the skill itself, the sheet shows the shifted "Now:" reading -
// so a player picking a shift sees the number change immediately.
func TestSheetRendersShiftCard(t *testing.T) {
	app, c := testAppWithBlankChar(t)
	c.Shifts = []model.AppliedShift{{SkillKey: "offense.Strength", Shift: 1}}
	// The sheet handler loads its own copy from the store, so the card must be
	// persisted before the request, the way every real flow does.
	if err := app.Store.Save(*c); err != nil {
		t.Fatalf("save: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/characters/char-1", nil)
	rec := httptest.NewRecorder()
	app.Router().ServeHTTP(rec, req)
	body := rec.Body.String()

	if !strings.Contains(body, "Apply Shift") || !strings.Contains(body, `id="shift-modal"`) {
		t.Fatal("the sheet is missing the shifts strip or its modal")
	}
	// The strip's remove control is labelled with the skill, so its presence
	// distinguishes the card from an ordinary sheet mention of the skill name.
	if !strings.Contains(body, `aria-label="Remove Strength shift"`) {
		t.Error("the shift card is rendered, but the strip is missing its remove control")
	}
	if !strings.Contains(body, "Now:") {
		t.Error("the shifted skill must show its Now reading")
	}

	// And the skill grid must agree with the engine for the card.
	views := engine.EffectiveSkills(app.Cfg.Config, *c)
	if v := views[model.SkillKey("offense", "Strength")]; v.Shift != 1 || !v.Up() {
		t.Errorf("skill view = %+v, want a +1 shift", v)
	}
}

// TestAppliedShiftWarningNamesBothSources checks the clamp notice triggers for
// shift cards too - a card pushing a skill past the ladder end must be
// explained, not silently swallowed.
func TestAppliedShiftWarningNamesBothSources(t *testing.T) {
	app, c := testAppWithBlankChar(t)
	// The shipped ladder starts at "inept"; the blank character's low rungs run
	// out quickly, so a -6 card on any skill is clamped.
	c.Shifts = []model.AppliedShift{{SkillKey: "defense.Reflex", Shift: -6}}

	if warn := app.conditionWarning(c); warn == "" {
		t.Fatal("a clamped shift card produced no warning")
	} else if !strings.Contains(warn, "condition or shift") {
		t.Errorf("warning %q must name shift cards, not just conditions", warn)
	}
}
