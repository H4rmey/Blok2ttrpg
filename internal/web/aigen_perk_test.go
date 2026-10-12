package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// validPerkYAML is a single built perk that passes ai.ValidatePerk (one damage
// enactment, known ids).
const validPerkYAML = `name: Power Strike
type: execution
enactments:
  - type: damage
    fields:
      source: offense.Strength
      flat: 0
      damage_type:
        - value: Slashing
    interaction: direct
    interaction_data:
      range: "1"
      targets: 1
    validation_data:
      engage: d6
      counter_skill:
        - value: defense.Reflex
`

// emptyPerkYAML is a non-passive perk with no enactments: it must be rejected.
const emptyPerkYAML = "name: Empty\ntype: execution\n"

// TestSkillSavePushesStatCards pins the regression fix: a skill edit must push
// the point totals out of band, not just swap the skills panel.
func TestSkillSavePushesStatCards(t *testing.T) {
	app, c := testAppWithBlankChar(t)

	form := url.Values{}
	form.Set("_tab", "tab-skills")
	form.Set("level", "1")
	form.Set("skill_general_Stealth", "novice")
	req := httptest.NewRequest(http.MethodPost, "/characters/"+c.ID, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	app.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("skill save returned %d: %s", rec.Code, rec.Body.String())
	}
	if body := rec.Body.String(); !strings.Contains(body, `id="stat-cards" hx-swap-oob`) {
		t.Errorf("skill save did not push the stat cards:\n%s", body)
	}
}

// TestLevelSavePushesStatCards pins the level path: the level input uses
// hx-swap="none", so both the bar and the skills tab must be out of band.
func TestLevelSavePushesStatCards(t *testing.T) {
	app, c := testAppWithBlankChar(t)

	form := url.Values{}
	form.Set("_tab", "level-only")
	form.Set("level", "2")
	req := httptest.NewRequest(http.MethodPost, "/characters/"+c.ID, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	app.Router().ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, `id="stat-cards" hx-swap-oob`) {
		t.Errorf("level save did not push the stat cards:\n%s", body)
	}
	if !strings.Contains(body, `id="tab-skills" hx-swap-oob`) {
		t.Errorf("level save did not push the skills tab:\n%s", body)
	}
}

// TestBuilderRejectsEmptyPerk pins the non-passive-has-enactments rule on the
// manual builder save path.
func TestBuilderRejectsEmptyPerk(t *testing.T) {
	app, c := testAppWithBlankChar(t)

	form := url.Values{}
	form.Set("name", "Empty")
	form.Set("type", app.Cfg.PerkTypes.List()[0].ID)
	form.Set("enactment_count", "0")
	req := httptest.NewRequest(http.MethodPost, "/characters/"+c.ID+"/perks/new", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	app.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("empty perk save returned %d, want 400: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "no enactments") {
		t.Errorf("rejection did not name the rule: %s", rec.Body.String())
	}
}

// TestAIImportPerkRejectsEmpty pins the same rule on the AI perk import path.
func TestAIImportPerkRejectsEmpty(t *testing.T) {
	app, c := testAppWithBlankChar(t)
	req := httptest.NewRequest(http.MethodPost, "/ai/import-perk?character="+c.ID, strings.NewReader(emptyPerkYAML))
	rec := httptest.NewRecorder()
	app.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("empty perk import returned %d, want 400: %s", rec.Code, rec.Body.String())
	}
}

// TestAIImportPerkAppends pins the happy path: a valid perk is added to the
// character.
func TestAIImportPerkAppends(t *testing.T) {
	app, c := testAppWithBlankChar(t)
	before := len(c.Perks)

	req := httptest.NewRequest(http.MethodPost, "/ai/import-perk?character="+c.ID, strings.NewReader(validPerkYAML))
	rec := httptest.NewRecorder()
	app.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("perk import returned %d: %s", rec.Code, rec.Body.String())
	}

	got, ok := app.Store.Get(c.ID)
	if !ok {
		t.Fatalf("character %q missing after import", c.ID)
	}
	if len(got.Perks) != before+1 {
		t.Fatalf("stored %d perks, want %d", len(got.Perks), before+1)
	}
	if got.Perks[len(got.Perks)-1].ID == "" {
		t.Error("imported perk has no id")
	}
}

// generatePerkEvents drives /ai/generate-perk and returns the streamed events.
func generatePerkEvents(t *testing.T, app *App, endpoint string) []streamedEvent {
	t.Helper()
	body := map[string]any{
		"description": "a fiery melee strike",
		"character":   "char-1",
		"endpoint":    endpoint,
		"apiKey":      "test-key",
		"model":       "test-model",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/ai/generate-perk", strings.NewReader(string(b)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("generate-perk returned %d: %s", rec.Code, rec.Body.String())
	}
	var events []streamedEvent
	for _, line := range strings.Split(strings.TrimSpace(rec.Body.String()), "\n") {
		if line == "" {
			continue
		}
		var ev streamedEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			t.Fatalf("bad stream line %q: %v", line, err)
		}
		events = append(events, ev)
	}
	return events
}

// TestAIGeneratePerkRetriesThenSucceeds pins the single-perk retry loop: an
// empty first reply is rejected, a correction is sent, and the valid second
// reply ends the stream.
func TestAIGeneratePerkRetriesThenSucceeds(t *testing.T) {
	app, _ := testAppWithBlankChar(t)
	provider := fakeProvider(t, emptyPerkYAML, validPerkYAML)
	defer provider.Close()

	events := generatePerkEvents(t, app, provider.URL)

	var sawRetry, sawDone bool
	var final streamedEvent
	for _, ev := range events {
		switch ev.Type {
		case "retry":
			sawRetry = true
			if !strings.Contains(ev.Message, "no enactments") {
				t.Errorf("retry message missing the issue: %q", ev.Message)
			}
		case "done":
			sawDone = true
			final = ev
		}
	}
	if !sawRetry {
		t.Fatal("no retry event; the loop did not correct the empty perk")
	}
	if !sawDone || !final.Valid {
		t.Fatalf("generation did not end valid: %+v", final)
	}
	if !strings.Contains(final.YAML, "Power Strike") {
		t.Errorf("final perk YAML missing expected content: %q", final.YAML)
	}
}
