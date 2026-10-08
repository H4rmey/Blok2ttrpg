package web

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/engine"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// importCharacterYAML posts the given character YAML to the import endpoint.
func importCharacterYAML(t *testing.T, app *App, yaml string) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("file", "character.yaml")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fw.Write([]byte(yaml)); err != nil {
		t.Fatalf("write yaml: %v", err)
	}
	mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/characters/import", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	app.Router().ServeHTTP(rec, req)
	return rec
}

// TestImportBackfillsSkillsAndRepairsPerkIDs pins the two import repairs: a
// hand-written file that names only some skills backfills the rest with the
// default proficiency (Untrained, not blank/Inept), and perks without ids get
// unique ones so each Configure link opens its own perk.
func TestImportBackfillsSkillsAndRepairsPerkIDs(t *testing.T) {
	app, _ := testAppWithBlankChar(t)

	yaml := `id: imported-1
name: Imported
level: 1
skills:
  offense.Precision: novice
traits:
  name: Imported
perks:
  - name: First
    type: execution
    enactments:
      - type: damage
  - name: Second
    type: execution
    enactments:
      - type: damage
`
	rec := importCharacterYAML(t, app, yaml)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("import returned %d, want 303: %s", rec.Code, rec.Body.String())
	}

	c, ok := app.Store.Get("imported-1")
	if !ok {
		t.Fatal("imported character not stored")
	}
	def := app.Cfg.DefaultProficiencyID()
	for _, g := range app.Cfg.Skills.List() {
		for _, sk := range g.Skills {
			key := model.SkillKey(g.ID, sk)
			if got := c.Skills[key]; got != def && key != model.SkillKey("offense", "Precision") {
				t.Errorf("skill %s = %q, want backfilled %q", key, got, def)
			}
		}
	}
	if len(c.Perks) != 2 {
		t.Fatalf("imported %d perks, want 2", len(c.Perks))
	}
	if c.Perks[0].ID == "" || c.Perks[1].ID == "" {
		t.Fatalf("perks still have empty ids: %q, %q", c.Perks[0].ID, c.Perks[1].ID)
	}
	if c.Perks[0].ID == c.Perks[1].ID {
		t.Errorf("both perks share id %q; Configure links would collide", c.Perks[0].ID)
	}
}

// TestHealingValidationUsesFlatDC pins the use_dc_validation switch: a healing
// enactment rolls its engage source against the configured flat DC instead of
// the target's counter skills.
func TestHealingValidationUsesFlatDC(t *testing.T) {
	cfg := loadCfg(t)
	perk := model.Perk{
		Name: "Mend",
		Type: "execution",
		Enactments: []model.Enactment{{
			Type:           "healing",
			Fields:         map[string]any{"source": "d6"},
			ValidationData: map[string]any{"engage": "d6"},
		}},
	}
	norm := engine.NormalizePerk(cfg, perk)
	lines := engine.PerkInstructions(cfg, norm)
	if len(lines) == 0 {
		t.Fatal("no instructions generated")
	}
	want := fmt.Sprintf("Roll 1d6 vs DC %d.", cfg.DCValidationDC(0))
	if got := lines[0].Validation; got != want {
		t.Errorf("healing validation = %q, want %q", got, want)
	}
}

// TestConditionSolutionShowsBothSkillsAndDC pins the instruction fix for the
// config's solution_1/solution_2/solution_dc shape: both solution skills and
// the DC must appear, rather than the line silently vanishing.
func TestConditionSolutionShowsBothSkillsAndDC(t *testing.T) {
	cfg := loadCfg(t)
	perk := model.Perk{
		Name: "Slip",
		Type: "execution",
		Enactments: []model.Enactment{{
			Type:   "condition",
			Fields: map[string]any{"condition": "hastened", "duration": "1", "solution_1": "defense.Wisdom", "solution_2": "defense.Reflex", "solution_dc": 3},
		}},
	}
	norm := engine.NormalizePerk(cfg, perk)
	lines := engine.PerkInstructions(cfg, norm)
	if len(lines) == 0 {
		t.Fatal("no instructions generated")
	}
	got := lines[0].Solution
	if got == "" {
		t.Fatal("condition has no Solution line")
	}
	for _, want := range []string{"Wisdom", "Reflex", "DC 3"} {
		if !bytes.Contains([]byte(got), []byte(want)) {
			t.Errorf("solution %q missing %q", got, want)
		}
	}
}

// TestSelfInteractionUsesFlatDC pins that use_dc_validation on an interaction
// (Self) switches validation to the flat DC, not just the enactment-level flag.
func TestSelfInteractionUsesFlatDC(t *testing.T) {
	cfg := loadCfg(t)
	if ic, ok := cfg.Interaction("self"); !ok || !ic.UsesDCValidation() {
		t.Skip("config/Blok2Simplified self interaction is not flagged")
	}
	perk := model.Perk{
		Name: "Hex",
		Type: "execution",
		Enactments: []model.Enactment{{
			Type:            "damage",
			Interaction:     "self",
			Fields:          map[string]any{"source": "d6"},
			InteractionData: map[string]any{},
			ValidationData:  map[string]any{"engage": "d6"},
		}},
	}
	norm := engine.NormalizePerk(cfg, perk)
	// The flat DC must be normalized in and the counter list dropped.
	if _, ok := norm.Enactments[0].ValidationData["validation_dc"]; !ok {
		t.Fatalf("self interaction did not normalize a validation_dc: %v", norm.Enactments[0].ValidationData)
	}
	lines := engine.PerkInstructions(cfg, norm)
	if len(lines) == 0 {
		t.Fatal("no instructions generated")
	}
	want := fmt.Sprintf("Roll 1d6 vs DC %d.", cfg.DCValidationDC(0))
	if got := lines[0].Validation; got != want {
		t.Errorf("self interaction validation = %q, want %q", got, want)
	}
}

// TestFlatDCValidationFieldsDropCounters pins the builder region for a
// flat-DC enactment: it must show the engage source and the DC, and none of
// the counter-roll options (which the config names counter_option_1/2, not the
// old counter_skill).
func TestFlatDCValidationFieldsDropCounters(t *testing.T) {
	cfg := loadCfg(t)
	fields := cfg.ValidationFieldsFor("healing", "")
	var hasEngage, hasDC, hasCounter bool
	for _, f := range fields {
		switch {
		case f.Key == "engage":
			hasEngage = true
		case f.Key == "validation_dc":
			hasDC = true
		case config.IsCounterValidationField(f.Key):
			hasCounter = true
		}
	}
	if !hasEngage || !hasDC {
		t.Fatalf("flat-DC fields missing engage or DC: %+v", fields)
	}
	if hasCounter {
		t.Errorf("flat-DC still shows counter options: %+v", fields)
	}
}

// TestSelfInteractionBuilderRequestShowsOnlyEngageAndDC pins the interaction
// dropdown's request: it must include the enactment type (hx-include), so the
// handler recognises flat-DC mode and returns engage + DC rather than the
// counter options or nothing.
func TestSelfInteractionBuilderRequestShowsOnlyEngageAndDC(t *testing.T) {
	app, _ := testAppWithBlankChar(t)
	req := httptest.NewRequest(http.MethodGet,
		"/builder/interaction-fields?index=0&en0_interaction=self&en0_type=condition", nil)
	rec := httptest.NewRecorder()
	app.Router().ServeHTTP(rec, req)
	body := rec.Body.String()
	if !strings.Contains(body, "en0_v_validation_dc") {
		t.Errorf("self interaction did not render the DC field:\n%s", body)
	}
	if strings.Contains(body, "en0_v_counter_option") {
		t.Errorf("self interaction still rendered counter options:\n%s", body)
	}
}

// TestContestedValidationShowsCounterSkills pins that a normal contested
// enactment reads its real counter fields (counter_option_1/2) into the
// instruction, instead of the previously-missing line reading "No roll
// required".
func TestContestedValidationShowsCounterSkills(t *testing.T) {
	cfg := loadCfg(t)
	perk := model.Perk{
		Name: "Strike",
		Type: "execution",
		Enactments: []model.Enactment{{
			Type:            "damage",
			Interaction:     "direct",
			Fields:          map[string]any{"source": "d6"},
			InteractionData: map[string]any{},
			ValidationData: map[string]any{
				"engage":           "d6",
				"counter_option_1": "defense.Reflex",
				"counter_option_2": "defense.Wisdom",
			},
		}},
	}
	norm := engine.NormalizePerk(cfg, perk)
	lines := engine.PerkInstructions(cfg, norm)
	if len(lines) == 0 {
		t.Fatal("no instructions generated")
	}
	got := lines[0].Validation
	if got == "No roll required." || got == "" {
		t.Fatalf("contested validation produced %q", got)
	}
	for _, want := range []string{"Roll", "Reflex", "Wisdom"} {
		if !strings.Contains(got, want) {
			t.Errorf("validation %q missing %q", got, want)
		}
	}
}

// TestResourceInstruction pins the previously-missing resource wording: a
// generate enactment states how much of what it grants.
func TestResourceInstruction(t *testing.T) {
	cfg := loadCfg(t)
	perk := model.Perk{
		Name: "Blood Let",
		Type: "execution",
		Enactments: []model.Enactment{{
			Type:   "resource",
			Fields: map[string]any{"resource_mode": "generate", "resource_name": "Blood", "generate_amount": 2},
		}},
	}
	norm := engine.NormalizePerk(cfg, perk)
	lines := engine.PerkInstructions(cfg, norm)
	if len(lines) == 0 {
		t.Fatal("no instructions generated")
	}
	if got := lines[0].Success; got != "Gain 2 Blood." {
		t.Errorf("resource success = %q, want generate wording", got)
	}
}
