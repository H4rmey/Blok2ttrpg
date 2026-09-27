package web

import (
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/engine"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
	"github.com/harmey/blok2ttrpg-v5/internal/premade"
	"github.com/harmey/blok2ttrpg-v5/internal/store"
)

// testAppWithPerk builds an app plus a character holding the Time Slip library
// perk. Shared fixture for the perk list and builder tests.
func testAppWithPerk(t *testing.T) (*App, *model.Character) {
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
	ab, err := premade.New("../../library").GetPerk("time_slip")
	if err != nil {
		t.Fatalf("get perk: %v", err)
	}
	ab.ID = "perk-1"
	// Library perks are normalized on import, so the fixture mirrors what is
	// actually stored on a character.
	ab = engine.NormalizePerk(cfg.Config, ab)
	c := app.blankCharacter("char-1")
	c.Traits["name"] = "Tester"
	c.Perks = append(c.Perks, ab)
	return app, &c
}

// TestPerkListShowsCostAndInstructions renders the Perks list for a character
// holding an imported library perk and asserts that the page reports a non-zero
// point cost and includes the generated instruction text, so the user does not
// have to open the builder to read the perk.
func TestPerkListShowsCostAndInstructions(t *testing.T) {
	app, c := testAppWithPerk(t)

	rec := httptest.NewRecorder()
	app.renderPerkList(rec, c)
	body := rec.Body.String()

	if !strings.Contains(body, "Time Slip") {
		t.Fatalf("perk name missing from list:\n%s", body)
	}
	// Time Slip costs 8 build points once normalized: the execution base, two
	// condition enactments (the first waives its base cost), the additional-
	// enactment surcharge, the new-target surcharge on the second enactment
	// (it hastens you and then delays an enemy, so it buys its own target), a
	// direct interaction at 5m, and a validation block on each of the two
	// enactments that own a target. A different figure here means field binding
	// regressed (e.g. missing yaml tags on interaction_data / validation_data),
	// an perk-type base_cost stopped being read, or normalization changed
	// what gets stored.
	if !strings.Contains(body, "8 pt") {
		t.Errorf("expected Time Slip to render an 8 pt cost; got:\n%s", body)
	}
	// 9 points at level 1, 8 spent, so one remains.
	if !strings.Contains(body, "1/9") {
		t.Errorf("expected remaining perk points 1/9 in the summary; got:\n%s", body)
	}

	if !strings.Contains(body, "instruction-block") {
		t.Errorf("instructions not rendered inline in the perk list:\n%s", body)
	}
	// The cost is shown as a prominent chip rather than muted inline text.
	if !strings.Contains(body, "perk-cost-chips") {
		t.Errorf("perk costs are not rendered as chips:\n%s", body)
	}
	// The character bar is shown on every character-scoped page.
	if !strings.Contains(body, "stats-bar") {
		t.Errorf("character bar missing from the perk list page:\n%s", body)
	}
}

// TestPerkListRefreshPartial checks that the refresh route returns only the
// perk list region with recomputed costs, so "Refresh All" can swap it in place
// without a page reload, and that it reports what it did.
func TestPerkListRefreshPartial(t *testing.T) {
	app, c := testAppWithPerk(t)

	rec := httptest.NewRecorder()
	app.renderPerkListPartial(rec, c, 0)
	body := rec.Body.String()

	if strings.Contains(body, "<html") {

		t.Errorf("refresh should return a partial, not a full page:\n%s", body)
	}
	if !strings.Contains(body, "8 pt") {
		t.Errorf("refreshed perk list did not recompute the cost:\n%s", body)
	}
	// The refresh reports its outcome, so the control is visibly not a no-op.
	if !strings.Contains(body, "Recalculated") {
		t.Errorf("refresh did not report its outcome:\n%s", body)
	}
}

// TestBuilderShowsCharacterBar checks that opening a perk in the builder shows
// the shared character bar and the manual recalculate control, and that the
// cost is recomputed from the rendered form on load.
func TestBuilderShowsCharacterBar(t *testing.T) {
	app, c := testAppWithPerk(t)

	rec := httptest.NewRecorder()
	app.renderBuilder(rec, c, &c.Perks[0], false)
	body := rec.Body.String()

	if !strings.Contains(body, "stats-bar") {
		t.Errorf("character bar missing from the builder:\n%s", body)
	}
	if !strings.Contains(body, "Skill Points") {
		t.Errorf("builder bar is missing the character budget cards:\n%s", body)
	}
	if !strings.Contains(body, "icon-btn") {
		t.Errorf("recalculate button missing from the builder cost cards:\n%s", body)
	}
	if !strings.Contains(body, `hx-trigger="load,`) {
		t.Errorf("builder does not recompute cost on load:\n%s", body)
	}
}

// TestNormalizeIsIdempotent is the core guarantee of the normalization design:
// normalizing an already normalized perk must not change its cost. If this
// fails, opening and saving a perk could keep shifting its price.
func TestNormalizeIsIdempotent(t *testing.T) {
	app, c := testAppWithPerk(t)
	cfg := app.Cfg.Config

	once := engine.NormalizePerk(cfg, c.Perks[0])
	twice := engine.NormalizePerk(cfg, once)

	if got, want := engine.PerkCost(cfg, twice), engine.PerkCost(cfg, once); got != want {
		t.Errorf("normalization is not idempotent: %+v then %+v", want, got)
	}
}

// TestPerkLibraryShowsCost checks that the built-in perk browser lists each
// perk with its cost, and that the figure matches what the perk will cost once
// imported. Showing a different price in the library than on the character was
// exactly the class of mismatch normalization exists to prevent.
func TestPerkLibraryShowsCost(t *testing.T) {
	app, _ := testAppWithPerk(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/perks/library?character=char-1", nil)
	app.handlePerkLibrary(rec, req)
	body := rec.Body.String()

	if !strings.Contains(body, "Perk Library") {
		t.Errorf("library is not titled Perk Library:\n%s", body)
	}
	if !strings.Contains(body, "perk-cost-chips") {
		t.Errorf("library does not show perk costs:\n%s", body)
	}
	// Time Slip costs 8 perk points on a character, so the library must quote
	// the same figure.
	if !strings.Contains(body, "8 perk pt") {
		t.Errorf("library cost does not match the imported cost:\n%s", body)
	}
	// The pre-rename wording must be gone. This guard used to read "Perk
	// Library", which the Abilities -> Perks rename turned into the same string
	// the assertion above requires, so the test contradicted itself.
	if strings.Contains(body, "Ability Library") {
		t.Errorf("library still uses the old Ability Library wording:\n%s", body)
	}
}

// TestLibraryPerksNormalizeStably checks every built-in perk: importing it
// (which normalizes) and normalizing again must not move its cost. This catches
// config/library mismatches - such as a field whose default falls outside its
// own min/max - at the source rather than in the UI.
func TestLibraryPerksNormalizeStably(t *testing.T) {
	cfg, err := config.Load("../../config/Blok2Simplified")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	abs, err := premade.New("../../library").ListPerks()
	if err != nil {
		t.Fatalf("list perks: %v", err)
	}
	for _, ab := range abs {
		norm := engine.NormalizePerk(cfg.Config, ab)
		again := engine.NormalizePerk(cfg.Config, norm)
		if got, want := engine.PerkCost(cfg.Config, again), engine.PerkCost(cfg.Config, norm); got != want {
			t.Errorf("%s: normalization not stable: %+v then %+v", ab.Name, want, got)
		}
	}
}
