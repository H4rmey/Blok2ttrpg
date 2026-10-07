package web

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

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
			ValidationData: map[string]any{"engage": "d6", "validation_dc": 2},
		}},
	}
	norm := engine.NormalizePerk(cfg, perk)
	lines := engine.PerkInstructions(cfg, norm)
	if len(lines) == 0 {
		t.Fatal("no instructions generated")
	}
	if got := lines[0].Validation; got != "Roll 1d6 vs DC 2." {
		t.Errorf("healing validation = %q, want flat-DC line", got)
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
