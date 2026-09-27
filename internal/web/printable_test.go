package web

import (
	"net/http/httptest"
	"strings"
	"testing"
)

// TestCharacterPDFIsComplete guards the three things a printed sheet used to be
// missing: the dedicated print stylesheet, proficiencies with the die they
// grant (rather than the bare tier id), and the perks with their cost and
// generated rules text. A sheet is read away from the app, so an incomplete one
// is unusable.
func TestCharacterPDFIsComplete(t *testing.T) {
	app, c := testAppWithPerk(t)

	rec := httptest.NewRecorder()
	app.renderCharacterPDF(rec, *c)
	body := rec.Body.String()

	if !strings.Contains(body, "print.css") {
		t.Errorf("PDF page does not load the print stylesheet:\n%s", body)
	}
	// The default proficiency of a blank character is a named tier with a die,
	// so the rendered label must carry a die rather than the raw stored id.
	if !strings.Contains(body, "(d") {
		t.Errorf("skill proficiencies render without their die:\n%s", body)
	}
	if strings.Contains(body, "<dd>untrained</dd>") {
		t.Errorf("skill proficiencies still render as raw tier ids:\n%s", body)
	}
	if !strings.Contains(body, "Time Slip") {
		t.Errorf("perks missing from the PDF:\n%s", body)
	}
	// The same 8 pt figure the on-screen perk list quotes.
	if !strings.Contains(body, "8 pt") {
		t.Errorf("perk cost missing from the PDF:\n%s", body)
	}
	if !strings.Contains(body, "print-instruction") {
		t.Errorf("generated perk instructions missing from the PDF:\n%s", body)
	}
	// Derived totals belong on the sheet too: they are the figures looked up
	// most often during play.
	if !strings.Contains(body, "Perk Points") {
		t.Errorf("derived point totals missing from the PDF:\n%s", body)
	}
}

// TestBlankCharacterSheet checks the printable empty sheet is driven by the
// config: every skill group appears, and each skill offers one tick box per
// proficiency tier labelled with what that tier grants.
func TestBlankCharacterSheet(t *testing.T) {
	app, _ := testAppWithPerk(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/print/character", nil)
	app.handleBlankCharacterSheet(rec, req)
	body := rec.Body.String()

	if rec.Code != 200 {
		t.Fatalf("blank sheet returned %d", rec.Code)
	}
	for _, g := range app.Cfg.Skills.List() {
		if !strings.Contains(body, g.Label) {
			t.Errorf("skill group %q missing from the blank sheet", g.Label)
		}
	}
	if !strings.Contains(body, "print-box") {
		t.Errorf("blank sheet has no proficiency tick boxes:\n%s", body)
	}
	if !strings.Contains(body, "print-line") {
		t.Errorf("blank sheet has no write-in lines:\n%s", body)
	}
}

// TestBlankCharacterSheetRowsClamped checks the ?rows= count is honoured and
// bounded, so a hand-edited request cannot ask for an unbounded page count.
func TestBlankCharacterSheetRowsClamped(t *testing.T) {
	app, _ := testAppWithPerk(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/print/character?rows=500", nil)
	app.handleBlankCharacterSheet(rec, req)
	body := rec.Body.String()

	if strings.Contains(body, "Perk 21") {
		t.Errorf("blank sheet did not clamp the requested row count:\n%s", body)
	}
	if !strings.Contains(body, "Perk 20") {
		t.Errorf("blank sheet did not honour a raised row count:\n%s", body)
	}
}

// TestPerkTemplate checks the printable perk worksheet lists every perk type and
// enactment the config defines, so a perk drafted on paper can be priced and
// then typed into the builder without guessing.
func TestPerkTemplate(t *testing.T) {
	app, _ := testAppWithPerk(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/print/perk", nil)
	app.handlePerkTemplate(rec, req)
	body := rec.Body.String()

	if rec.Code != 200 {
		t.Fatalf("perk template returned %d", rec.Code)
	}
	for _, pt := range app.Cfg.PerkTypes.List() {
		if !strings.Contains(body, pt.Name) {
			t.Errorf("perk type %q missing from the perk template", pt.Name)
		}
	}
	for _, en := range app.Cfg.Enactments.List() {
		if !strings.Contains(body, en.Name) {
			t.Errorf("enactment %q missing from the perk template", en.Name)
		}
	}
	// Three enactment blocks by default, each with its own write-in region.
	if !strings.Contains(body, "Enactment 3") {
		t.Errorf("perk template does not render the default enactment blocks:\n%s", body)
	}
	if !strings.Contains(body, "Cost Tally") {
		t.Errorf("perk template has no cost tally box:\n%s", body)
	}
}
