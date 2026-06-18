package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	srv, err := New()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	return srv
}

func TestIndexPage(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET / status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Blok2ttrpg") {
		t.Error("Index page should contain 'Blok2ttrpg'")
	}
	if !strings.Contains(body, "Character Attributes") {
		t.Error("Index page should contain 'Character Attributes' tab")
	}
	if !strings.Contains(body, "tab_attributes") || !strings.Contains(body, "Name") {
		// The attributes tab should be rendered inline on initial load
		if !strings.Contains(body, "Name") {
			t.Error("Index page should render attributes tab by default")
		}
	}
}

func TestTabEndpoints(t *testing.T) {
	srv := newTestServer(t)

	// First visit index to create session
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	// Get session cookie
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("Expected session cookie to be set")
	}

	tabs := []struct {
		path     string
		contains string
	}{
		{"/tabs/attributes", "Character Attributes"},
		{"/tabs/general-traits", "General Traits"},
		{"/tabs/combative-traits", "Combative Traits"},
		{"/tabs/abilities", "Ability Builder"},
	}

	for _, tt := range tabs {
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		req.AddCookie(cookies[0])
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want 200", tt.path, w.Code)
		}
		if !strings.Contains(w.Body.String(), tt.contains) {
			t.Errorf("GET %s should contain %q", tt.path, tt.contains)
		}
	}
}

func TestAttributeUpdate(t *testing.T) {
	srv := newTestServer(t)

	// Create session
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	cookies := w.Result().Cookies()

	// Update name
	form := url.Values{"field": {"name"}, "value": {"Test Hero"}}
	req = httptest.NewRequest(http.MethodPost, "/attributes/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("POST /attributes/update status = %d, want 200", w.Code)
	}

	// Verify it persisted by loading attributes tab
	req = httptest.NewRequest(http.MethodGet, "/tabs/attributes", nil)
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "Test Hero") {
		t.Error("Attribute update should persist: 'Test Hero' not found")
	}
}

func TestTempAttributeAddRemove(t *testing.T) {
	srv := newTestServer(t)

	// Create session
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	cookies := w.Result().Cookies()

	// Add temp attribute
	form := url.Values{"name": {"Burning"}, "description": {"On fire"}, "duration": {"3 rounds"}}
	req = httptest.NewRequest(http.MethodPost, "/attributes/temp/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("POST /attributes/temp/add status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Burning") {
		t.Error("Response should contain 'Burning'")
	}
	if !strings.Contains(w.Body.String(), "3 rounds") {
		t.Error("Response should contain '3 rounds'")
	}

	// Remove it
	req = httptest.NewRequest(http.MethodDelete, "/attributes/temp/remove?index=0", nil)
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("DELETE /attributes/temp/remove status = %d, want 200", w.Code)
	}
	if strings.Contains(w.Body.String(), "Burning") {
		t.Error("After removal, response should not contain 'Burning'")
	}
}

func TestNotFound(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GET /nonexistent status = %d, want 404", w.Code)
	}
}

func TestGeneralTraitUpdate(t *testing.T) {
	srv := newTestServer(t)

	// Create session
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	cookies := w.Result().Cookies()

	// Set Strength to Trained (proficiency 2, costs 2 points)
	form := url.Values{"trait": {"Strength"}, "proficiency": {"2"}}
	req = httptest.NewRequest(http.MethodPost, "/traits/general/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("POST /traits/general/update status = %d, want 200", w.Code)
	}
	// Should show 6 available (8 total - 2 spent)
	if !strings.Contains(w.Body.String(), "6 / 8") {
		t.Error("Should show 6/8 trait points after spending 2")
	}
}

func TestGeneralTraitBudgetEnforcement(t *testing.T) {
	srv := newTestServer(t)

	// Create session
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	cookies := w.Result().Cookies()

	// Spend 4 on Strength (Master)
	form := url.Values{"trait": {"Strength"}, "proficiency": {"4"}}
	req = httptest.NewRequest(http.MethodPost, "/traits/general/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	// Spend 4 on Dexterity (Master)
	form = url.Values{"trait": {"Dexterity"}, "proficiency": {"4"}}
	req = httptest.NewRequest(http.MethodPost, "/traits/general/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	// Now at 0 available — try to spend 1 more on Stealth (should fail)
	form = url.Values{"trait": {"Stealth"}, "proficiency": {"1"}}
	req = httptest.NewRequest(http.MethodPost, "/traits/general/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	// Should still show 0 available and trigger an error
	if !strings.Contains(w.Body.String(), "0 / 8") {
		t.Error("Should still show 0/8 after failed attempt")
	}
	if w.Header().Get("HX-Trigger") == "" {
		t.Error("Should set HX-Trigger with error message")
	}
}

func TestGeneralTraitRespec(t *testing.T) {
	srv := newTestServer(t)

	// Create session
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	cookies := w.Result().Cookies()

	// Spend 4 on Strength (Master)
	form := url.Values{"trait": {"Strength"}, "proficiency": {"4"}}
	req = httptest.NewRequest(http.MethodPost, "/traits/general/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	// Lower Strength back to Clumsy (free 4 points)
	form = url.Values{"trait": {"Strength"}, "proficiency": {"0"}}
	req = httptest.NewRequest(http.MethodPost, "/traits/general/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	// Should show 8 available again
	if !strings.Contains(w.Body.String(), "8 / 8") {
		t.Error("After re-spec, should show 8/8 trait points")
	}
}

func TestCombativeTraitUpdate(t *testing.T) {
	srv := newTestServer(t)

	// Create session
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	cookies := w.Result().Cookies()

	// Set Precision (offense) to Expert (costs 3)
	form := url.Values{"section": {"offense"}, "trait": {"Precision"}, "proficiency": {"3"}}
	req = httptest.NewRequest(http.MethodPost, "/traits/combative/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("POST /traits/combative/update status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "5 / 8") {
		t.Error("Should show 5/8 trait points after spending 3 on Precision")
	}
}

func TestSharedTraitPool(t *testing.T) {
	srv := newTestServer(t)

	// Create session
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	cookies := w.Result().Cookies()

	// Spend 4 on general traits (Strength = Master)
	form := url.Values{"trait": {"Strength"}, "proficiency": {"4"}}
	req = httptest.NewRequest(http.MethodPost, "/traits/general/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	// Check combative tab shows 4/8 available (shared pool)
	req = httptest.NewRequest(http.MethodGet, "/tabs/combative-traits", nil)
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "4 / 8") {
		t.Error("Combative tab should show 4/8 after spending 4 on general traits")
	}

	// Spend 4 on combative (Precision = Master)
	form = url.Values{"section": {"offense"}, "trait": {"Precision"}, "proficiency": {"4"}}
	req = httptest.NewRequest(http.MethodPost, "/traits/combative/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "0 / 8") {
		t.Error("Should show 0/8 after spending all points")
	}

	// Try to spend more on defense — should fail
	form = url.Values{"section": {"defense"}, "trait": {"Reflex"}, "proficiency": {"1"}}
	req = httptest.NewRequest(http.MethodPost, "/traits/combative/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Header().Get("HX-Trigger") == "" {
		t.Error("Should trigger error when budget exceeded")
	}
}

func TestVitalTraitValues(t *testing.T) {
	srv := newTestServer(t)

	// Create session
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	cookies := w.Result().Cookies()

	// Set HP to Trained (should show 16 HP)
	form := url.Values{"section": {"vital"}, "trait": {"HP"}, "proficiency": {"2"}}
	req = httptest.NewRequest(http.MethodPost, "/traits/combative/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "16 HP") {
		t.Error("HP at Trained should show '16 HP'")
	}
}

func TestAbilityCreateAndDelete(t *testing.T) {
	srv := newTestServer(t)

	// Create session
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	cookies := w.Result().Cookies()

	// Create ability
	form := url.Values{"name": {"Fireball"}, "type": {"Execution"}}
	req = httptest.NewRequest(http.MethodPost, "/abilities/new", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("POST /abilities/new status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Fireball") {
		t.Error("Response should contain 'Fireball'")
	}
	if !strings.Contains(w.Body.String(), "Execution") {
		t.Error("Response should contain 'Execution' type")
	}

	// Delete it
	req = httptest.NewRequest(http.MethodDelete, "/abilities/delete?index=0", nil)
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("DELETE /abilities/delete status = %d, want 200", w.Code)
	}
	// Check the ability row is gone (the ability list table should not have Fireball as text content)
	// Note: "Fireball" appears in the placeholder, so we check for the ability row structure
	if !strings.Contains(w.Body.String(), "No abilities created") {
		t.Error("Should show empty state after deletion")
	}
}

func TestAbilityExport(t *testing.T) {
	srv := newTestServer(t)

	// Create session and add ability
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	cookies := w.Result().Cookies()

	form := url.Values{"name": {"Ice Bolt"}, "type": {"Execution"}}
	req = httptest.NewRequest(http.MethodPost, "/abilities/new", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	// Export
	req = httptest.NewRequest(http.MethodGet, "/abilities/export?index=0", nil)
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /abilities/export status = %d, want 200", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/x-yaml" {
		t.Errorf("Content-Type = %q, want application/x-yaml", w.Header().Get("Content-Type"))
	}
	if !strings.Contains(w.Body.String(), "Ice Bolt") {
		t.Error("Exported YAML should contain ability name")
	}
}

func TestAbilityWizard(t *testing.T) {
	srv := newTestServer(t)

	// Create session and ability
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	cookies := w.Result().Cookies()

	form := url.Values{"name": {"Test Spell"}, "type": {"Execution"}}
	req = httptest.NewRequest(http.MethodPost, "/abilities/new", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	// Open wizard
	req = httptest.NewRequest(http.MethodGet, "/abilities/edit?index=0", nil)
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /abilities/edit status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Test Spell") {
		t.Error("Wizard should show ability name")
	}
	if !strings.Contains(w.Body.String(), "Type Perks") {
		t.Error("Wizard should show type perks section")
	}
	if !strings.Contains(w.Body.String(), "Enact Damage") {
		t.Error("Wizard should show compatible enactments")
	}

	// Add enactment
	form = url.Values{"ability_index": {"0"}, "enactment_type": {"Enact Damage"}}
	req = httptest.NewRequest(http.MethodPost, "/abilities/wizard/add-enactment", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("POST add-enactment status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "1d4") {
		t.Error("Added damage enactment should show default 1d4 dice")
	}

	// Add perk to enactment
	form = url.Values{
		"ability_index":    {"0"},
		"perk_target":      {"enactment-0"},
		"perk_description": {"Shift Dice Tier of damage up"},
		"perk_add_cost":    {"2"},
		"perk_energy_cost": {"1"},
		"perk_is_optional": {"true"},
	}
	req = httptest.NewRequest(http.MethodPost, "/abilities/wizard/add-perk", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("POST add-perk status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Shift Dice Tier") {
		t.Error("Perk should appear in applied perks")
	}

	// Verify points deducted - go back to list
	req = httptest.NewRequest(http.MethodGet, "/abilities/wizard/back", nil)
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "8 / 10") {
		t.Error("Ability list should show 8/10 points (2 spent on perk)")
	}
}

func TestLevelUpDown(t *testing.T) {
	srv := newTestServer(t)

	// Create session
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	cookies := w.Result().Cookies()

	// Verify level 1 in initial page
	if !strings.Contains(w.Body.String(), "Level") {
		t.Error("Should show level controls")
	}

	// Level up
	req = httptest.NewRequest(http.MethodPost, "/level/up", nil)
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("POST /level/up status = %d, want 200", w.Code)
	}
	// Should now be level 2 with 9 trait points, 12 ability points
	body := w.Body.String()
	if !strings.Contains(body, "9/9") {
		t.Error("At level 2, should have 9/9 trait points")
	}
	if !strings.Contains(body, "12/12") {
		t.Error("At level 2, should have 12/12 ability points")
	}

	// Level down
	req = httptest.NewRequest(http.MethodPost, "/level/down", nil)
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("POST /level/down status = %d, want 200", w.Code)
	}
	body = w.Body.String()
	if !strings.Contains(body, "8/8") {
		t.Error("Back at level 1, should have 8/8 trait points")
	}
}

func TestLevelDownBlocked(t *testing.T) {
	srv := newTestServer(t)

	// Create session
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	cookies := w.Result().Cookies()

	// Level up to 2
	req = httptest.NewRequest(http.MethodPost, "/level/up", nil)
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	// Spend all 9 trait points
	for _, trait := range []struct{ name, prof string }{
		{"Strength", "4"},
		{"Dexterity", "4"},
		{"Stealth", "1"},
	} {
		form := url.Values{"trait": {trait.name}, "proficiency": {trait.prof}}
		req = httptest.NewRequest(http.MethodPost, "/traits/general/update", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(cookies[0])
		w = httptest.NewRecorder()
		srv.ServeHTTP(w, req)
	}

	// Try to level down — should fail
	req = httptest.NewRequest(http.MethodPost, "/level/down", nil)
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Header().Get("HX-Trigger") == "" {
		t.Error("Level down should be blocked with error when points overspent")
	}
}

func TestCharacterExportImport(t *testing.T) {
	srv := newTestServer(t)

	// Create session and set up character
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	cookies := w.Result().Cookies()

	// Set name
	form := url.Values{"field": {"name"}, "value": {"Round Trip Hero"}}
	req = httptest.NewRequest(http.MethodPost, "/attributes/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	// Set a trait
	form = url.Values{"trait": {"Knowledge"}, "proficiency": {"3"}}
	req = httptest.NewRequest(http.MethodPost, "/traits/general/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	// Level up
	req = httptest.NewRequest(http.MethodPost, "/level/up", nil)
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	// Export
	req = httptest.NewRequest(http.MethodGet, "/character/export", nil)
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Export status = %d, want 200", w.Code)
	}
	yamlData := w.Body.String()
	if !strings.Contains(yamlData, "Round Trip Hero") {
		t.Error("Exported YAML should contain character name")
	}
	if !strings.Contains(yamlData, "Expert") {
		t.Error("Exported YAML should contain Knowledge=Expert")
	}
	if !strings.Contains(yamlData, "level: 2") {
		t.Error("Exported YAML should show level 2")
	}

	// Reset to new character
	req = httptest.NewRequest(http.MethodPost, "/character/new", nil)
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	// Verify it's reset
	req = httptest.NewRequest(http.MethodGet, "/tabs/attributes", nil)
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if strings.Contains(w.Body.String(), "Round Trip Hero") {
		t.Error("After new, should not contain old character name")
	}

	// Import the exported YAML
	boundary := "testboundary123"
	body := fmt.Sprintf("--%s\r\nContent-Disposition: form-data; name=\"character_file\"; filename=\"char.yaml\"\r\nContent-Type: application/x-yaml\r\n\r\n%s\r\n--%s--", boundary, yamlData, boundary)
	req = httptest.NewRequest(http.MethodPost, "/character/import", strings.NewReader(body))
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Import status = %d, want 200", w.Code)
	}
	if w.Header().Get("HX-Redirect") != "/" {
		t.Error("Import should redirect to /")
	}

	// Verify imported data
	req = httptest.NewRequest(http.MethodGet, "/tabs/attributes", nil)
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if !strings.Contains(w.Body.String(), "Round Trip Hero") {
		t.Error("After import, should contain character name")
	}
}
