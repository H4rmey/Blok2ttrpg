package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// characterYAML builds a character YAML with n novice skills and nPassive
// "Mobile Attacker" passives. At level 1 the skill budget is 13 and the perk
// budget is 20, so (13, 20) is exactly valid and (10, 20) leaves 3 skill
// points unspent.
func characterYAML(skills, passives int) string {
	general := []string{
		"Stealth", "Perception", "Nature", "Crafting", "People Skill",
		"Performance", "Thievery", "Knowledge", "Magic", "Medicine",
		"Provoke", "Intimidate", "Athletics", "Deceive", "Resources",
	}
	var b strings.Builder
	b.WriteString("level: 1\nskills:\n")
	for i, s := range general {
		if i >= skills {
			break
		}
		fmt.Fprintf(&b, "  general.%s: novice\n", s)
	}
	b.WriteString("perks:\n")
	for i := 0; i < passives; i++ {
		b.WriteString("  - name: Mobile Attacker\n    fields:\n      passive_id: mobile_attacker\n")
	}
	return b.String()
}

// fakeProvider is an OpenAI-compatible test server that returns the given
// contents in order, one per call.
func fakeProvider(t *testing.T, contents ...string) *httptest.Server {
	t.Helper()
	var calls int32
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		idx := int(n) - 1
		if idx >= len(contents) {
			idx = len(contents) - 1
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": contents[idx]}},
			},
		})
	}))
}

type streamedEvent struct {
	Type    string          `json:"type"`
	Message string          `json:"message"`
	Valid   bool            `json:"valid"`
	YAML    string          `json:"yaml"`
	Report  json.RawMessage `json:"report"`
}

func generateEvents(t *testing.T, app *App, endpoint string) []streamedEvent {
	t.Helper()
	body := map[string]any{
		"description": "a stealthy rogue",
		"level":       1,
		"endpoint":    endpoint,
		"apiKey":      "test-key",
		"model":       "test-model",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/ai/generate", strings.NewReader(string(b)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("generate returned %d: %s", rec.Code, rec.Body.String())
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

// TestAIGenerateRetriesThenSucceeds pins the conversational retry loop: an
// invalid first reply (unspent skill points) is reported, a correction is sent,
// and the valid second reply ends the stream.
func TestAIGenerateRetriesThenSucceeds(t *testing.T) {
	app, _ := testAppWithBlankChar(t)
	provider := fakeProvider(t, characterYAML(10, 20), characterYAML(13, 20))
	defer provider.Close()

	events := generateEvents(t, app, provider.URL)

	var sawRetry, sawDone bool
	var final streamedEvent
	for _, ev := range events {
		switch ev.Type {
		case "retry":
			sawRetry = true
			if !strings.Contains(ev.Message, "unspent skill points") {
				t.Errorf("retry message missing the issue: %q", ev.Message)
			}
		case "done":
			sawDone = true
			final = ev
		}
	}
	if !sawRetry {
		t.Fatal("no retry event; the loop did not correct the invalid reply")
	}
	if !sawDone || !final.Valid {
		t.Fatalf("generation did not end valid: %+v", final)
	}
	if !strings.Contains(final.YAML, "mobile_attacker") {
		t.Errorf("final YAML missing expected content: %q", final.YAML)
	}
}

// TestAIGenerateStopsAfterSixAttempts pins the retry cap and the keep/discard
// payload: after six invalid replies the stream ends not-valid with the last
// draft and a summary.
func TestAIGenerateStopsAfterSixAttempts(t *testing.T) {
	app, _ := testAppWithBlankChar(t)
	// Always return the under-spent character.
	provider := fakeProvider(t, characterYAML(10, 20))
	defer provider.Close()

	events := generateEvents(t, app, provider.URL)

	attempts, done := 0, streamedEvent{}
	for _, ev := range events {
		if ev.Type == "status" {
			attempts++
		}
		if ev.Type == "done" {
			done = ev
		}
	}
	if attempts != maxAttempts {
		t.Errorf("made %d attempts, want %d", attempts, maxAttempts)
	}
	if done.Valid || done.YAML == "" || !strings.Contains(done.Message, "unspent skill points") {
		t.Errorf("unexpected final event: %+v", done)
	}
}

// TestIndexHasGenerateCharacterEntryPoint pins the UI entry point and that the
// level picker renders the configured range.
func TestIndexHasGenerateCharacterEntryPoint(t *testing.T) {
	app, _ := testAppWithBlankChar(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	app.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("index returned %d", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"Generate Character", "generate-character-modal", "ai-generate-btn", `<option value="10">`} {
		if !strings.Contains(body, want) {
			t.Errorf("index page missing %q", want)
		}
	}
}

// TestAIImportSavesDraft pins that the keep path stores the draft with the
// usual import repairs.
func TestAIImportSavesDraft(t *testing.T) {
	app, _ := testAppWithBlankChar(t)
	req := httptest.NewRequest(http.MethodPost, "/ai/import", strings.NewReader(characterYAML(13, 20)))
	rec := httptest.NewRecorder()
	app.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("import returned %d: %s", rec.Code, rec.Body.String())
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode import response: %v", err)
	}
	if out.ID == "" {
		t.Fatal("import returned no id")
	}
	c, ok := app.Store.Get(out.ID)
	if !ok {
		t.Fatalf("character %q not stored", out.ID)
	}
	if len(c.Perks) != 20 {
		t.Errorf("stored %d perks, want 20", len(c.Perks))
	}
	for i, p := range c.Perks {
		if p.ID == "" {
			t.Fatalf("perk %d has no repaired id", i)
		}
	}
}
