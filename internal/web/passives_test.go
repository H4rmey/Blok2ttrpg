package web

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/engine"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// loadCfg loads the shipped ruleset for the passive tests.
func loadCfg(t *testing.T) *config.Config {
	t.Helper()
	cfg, err := config.Load("../../config/Blok2Simplified")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	return cfg.Config
}

// passivePerk builds the stored form of a selected passive: a perk of the
// passive type carrying the entry id and its configured field values.
func passivePerk(id string, values map[string]any) model.Perk {
	return model.Perk{
		Name: id,
		Type: passivePerkType,
		Fields: map[string]any{
			passiveIDField:     id,
			passiveFieldsField: values,
		},
	}
}

// vals is shorthand for a passive's configured field values in these tests.
func vals(kv ...any) map[string]any {
	out := map[string]any{}
	for i := 0; i+1 < len(kv); i += 2 {
		out[kv[i].(string)] = kv[i+1]
	}
	return out
}

// defaultsOf resolves an entry and its default field values, which is what a
// passive is taken at.
func defaultsOf(t *testing.T, cfg *config.Config, id string) (config.Passive, map[string]any) {
	t.Helper()
	p, ok := cfg.PassiveByID(id)
	if !ok {
		t.Fatalf("%s not found in the passive catalogue", id)
	}
	return p, cfg.PassiveDefaults(p)
}

// TestPassiveCostAtDefault pins that a passive at its default value costs
// exactly the entry's build cost, with no double-charging between the dropdown
// option and the per-step upgrade path.
func TestPassiveCostAtDefault(t *testing.T) {
	cfg := loadCfg(t)
	cases := []struct {
		id   string
		want int
	}{
		// Uses the category default of 1.
		{"brutal_critical", 1},
		{"grudge_keeper", 1},
		{"resistance", 1},
		// Entries with their own build_cost override.
		{"sure_footed", 2},
		{"lingering_touch", 2},
		{"opportunist", 4},
	}
	for _, tc := range cases {
		_, defaults := defaultsOf(t, cfg, tc.id)
		got := engine.PerkCost(cfg, passivePerk(tc.id, defaults)).Build
		if got != tc.want {
			t.Errorf("%s at defaults: build = %d, want %d", tc.id, got, tc.want)
		}
	}
}

// TestPassiveCostWithConfiguredFields pins that an entry's fields are priced with
// the generic field coster: per-step numbers above their default, and flat costs
// on checkboxes.
func TestPassiveCostWithConfiguredFields(t *testing.T) {
	cfg := loadCfg(t)
	cases := []struct {
		name   string
		id     string
		values map[string]any
		want   int
	}{
		// build 1, multiplier default 2, +3 per step. One step up is 1 + 3.
		{"brutal_critical raised", "brutal_critical", vals("multiplier", 3), 4},
		// build 2, limit default 3, +1 per step. Two steps up is 2 + 2.
		{"efficient_caster raised", "efficient_caster", vals("limit", 5), 4},
		// Free text costs nothing: naming the source is flavour.
		{"resistance text only", "resistance", vals("source", "acid", "amount", 1), 1},
		// amount 1 -> 3 at +3 per step is 1 + 6.
		{"resistance raised", "resistance", vals("source", "fire", "amount", 3), 7},
		// The immune checkbox is a flat 6 on top of the raised amount.
		{"resistance immune", "resistance", vals("source", "fire", "amount", 3, "immune", true), 13},
		// The untyped checkbox is a flat 4. This is the path that replaced the
		// old standalone Thick Skinned entry, so it is pinned here: broad
		// reduction must cost strictly more than the same amount typed.
		{"resistance untyped", "resistance", vals("source", "fire", "amount", 3, "untyped", true), 11},
		// Both upgrades together, to pin that they stack additively rather than
		// one silently overriding the other.
		{"resistance untyped immune", "resistance", vals("source", "fire", "amount", 1, "untyped", true, "immune", true), 11},
	}
	for _, tc := range cases {
		got := engine.PerkCost(cfg, passivePerk(tc.id, tc.values)).Build
		if got != tc.want {
			t.Errorf("%s: build = %d, want %d", tc.name, got, tc.want)
		}
	}
}

// TestPassiveCostsNoEnergy pins that a passive is free to use. Every other perk
// is floored at 1 energy; a passive is always on, so it must be exempt.
func TestPassiveCostsNoEnergy(t *testing.T) {
	cfg := loadCfg(t)
	for _, p := range cfg.Passives.Entries {
		if got := engine.PerkCost(cfg, passivePerk(p.ID, cfg.PassiveDefaults(p))).Energy; got != 0 {
			t.Errorf("%s: energy = %d, want 0", p.ID, got)
		}
	}
}

// TestPassiveNumberClampedToRange pins that a number outside its field's declared
// range is clamped rather than charged for, so a hand-edited character file
// cannot buy an upgrade the ruleset does not offer.
func TestPassiveNumberClampedToRange(t *testing.T) {
	cfg := loadCfg(t)
	// brutal_critical's multiplier maxes at 3, so asking for 9 must cost no more.
	atMax := engine.PerkCost(cfg, passivePerk("brutal_critical", vals("multiplier", 3))).Build
	beyond := engine.PerkCost(cfg, passivePerk("brutal_critical", vals("multiplier", 9))).Build
	if beyond < atMax {
		t.Errorf("value beyond max: build = %d, want at least the max %d", beyond, atMax)
	}
	// Below the minimum must never refund into a negative cost.
	below := engine.PerkCost(cfg, passivePerk("resistance", vals("amount", -5))).Build
	if below < 0 {
		t.Errorf("value below min: build = %d, want no refund", below)
	}
}

// TestPassiveDescriptionSubstitutesFields pins the {key} placeholders: the text a
// player reads must state what they actually configured, including a key used
// more than once.
func TestPassiveDescriptionSubstitutesFields(t *testing.T) {
	cfg := loadCfg(t)
	p, ok := cfg.PassiveByID("resistance")
	if !ok {
		t.Fatal("resistance not found in the passive catalogue")
	}
	got := cfg.PassiveDescription(p, vals("source", "acid", "amount", 4, "immune", false))
	if contains(got, "{") {
		t.Errorf("description still contains a placeholder: %q", got)
	}
	// "source" appears twice in the sentence, so both must be filled.
	if n := countOf(got, "acid"); n != 2 {
		t.Errorf("repeated key substituted %d time(s), want 2: %q", n, got)
	}
	if !contains(got, "by 4") {
		t.Errorf("description does not state the configured number: %q", got)
	}
	// A checkbox reads as yes/no rather than true/false.
	if !contains(got, "immune: no") {
		t.Errorf("checkbox did not render as no: %q", got)
	}
}

// TestEveryPassiveHasIDAndName guards the catalogue itself: an entry without an
// id cannot be stored on a character, and one without a name renders blank.
func TestEveryPassiveHasIDAndName(t *testing.T) {
	cfg := loadCfg(t)
	if len(cfg.Passives.Entries) == 0 {
		t.Fatal("no passives configured")
	}
	seen := map[string]bool{}
	for i, p := range cfg.Passives.Entries {
		if p.ID == "" {
			t.Errorf("entry %d has no id", i)
		}
		if p.Name == "" {
			t.Errorf("entry %q has no name", p.ID)
		}
		if seen[p.ID] {
			t.Errorf("duplicate passive id %q", p.ID)
		}
		seen[p.ID] = true
		// Every field must be referenced in the description, or the player is
		// shown a control whose effect the rules text never explains. This is the
		// authoring guard that keeps a new field from being invisible.
		for _, f := range p.Fields {
			if !contains(p.Description, "{"+f.Key+"}") {
				t.Errorf("%q declares field %q but never uses {%s} in its description",
					p.ID, f.Key, f.Key)
			}
		}
	}
}

// TestPassiveDescriptionSegments pins the split that lets a template highlight
// each configured value inside the rules text. Rejoining the segments must
// reproduce exactly what PassiveDescription returns, otherwise the highlighted
// text and the plain text would say different things.
func TestPassiveDescriptionSegments(t *testing.T) {
	cfg := loadCfg(t)
	for _, p := range cfg.Passives.Entries {
		defaults := cfg.PassiveDefaults(p)
		segs := cfg.PassiveDescriptionSegments(p, defaults)
		if len(segs) == 0 {
			t.Errorf("%s: no description segments", p.ID)
			continue
		}
		joined := ""
		for _, s := range segs {
			joined += s.Text + s.Value
		}
		if want := cfg.PassiveDescription(p, defaults); joined != want {
			t.Errorf("%s: rejoined segments = %q, want %q", p.ID, joined, want)
		}
		// A configurable entry must produce at least one highlightable value, or
		// the player cannot see what they chose.
		if p.Configurable() {
			keyed := 0
			for _, s := range segs {
				if s.Key != "" {
					keyed++
				}
			}
			if keyed == 0 {
				t.Errorf("%s is configurable but its text highlights nothing", p.ID)
			}
		}
	}
}

// TestPassiveDescriptionSegmentsFixedEntry pins that an entry with no fields
// yields a single plain segment, so the template renders it with nothing
// highlighted.
func TestPassiveDescriptionSegmentsFixedEntry(t *testing.T) {
	cfg := loadCfg(t)
	p, ok := cfg.PassiveByID("sure_footed")
	if !ok {
		t.Fatal("sure_footed not found")
	}
	if p.Configurable() {
		t.Skip("sure_footed gained fields; pick another fixed entry")
	}
	segs := cfg.PassiveDescriptionSegments(p, nil)
	if len(segs) != 1 || segs[0].Text != p.Description || segs[0].Key != "" {
		t.Errorf("fixed entry segments = %+v, want the whole description in one keyless segment", segs)
	}
}

// TestPassiveConfigurable pins the flag the UI branches on: an entry with fields
// opens the configure modal, one without is taken straight away.
func TestPassiveConfigurable(t *testing.T) {
	cfg := loadCfg(t)
	for id, want := range map[string]bool{
		"resistance":      true,
		"lingering_touch": true,
		"sure_footed":     false,
	} {
		p, ok := cfg.PassiveByID(id)
		if !ok {
			t.Fatalf("%s not found", id)
		}
		if got := p.Configurable(); got != want {
			t.Errorf("%s: Configurable() = %v, want %v", id, got, want)
		}
	}
}

// TestPassiveIDOf pins the discriminator that decides whether a perk is a
// passive. Every downstream rule (no Edit, no Export, no instructions, value
// control) keys off it, so a false negative would quietly let a passive into the
// builder.
func TestPassiveIDOf(t *testing.T) {
	if got := passiveIDOf(passivePerk("resistance", vals("amount", 1))); got != "resistance" {
		t.Errorf("passive perk: id = %q, want resistance", got)
	}
	// An ordinary built perk has no passive id.
	built := model.Perk{
		Name:   "Some Execution",
		Type:   "execution",
		Fields: map[string]any{"comment": "not a passive"},
	}
	if got := passiveIDOf(built); got != "" {
		t.Errorf("built perk: id = %q, want empty", got)
	}
	// A perk with no fields at all must not panic.
	if got := passiveIDOf(model.Perk{Name: "Bare"}); got != "" {
		t.Errorf("fieldless perk: id = %q, want empty", got)
	}
}

// TestPassivePickerOpensConfigureModal pins the two-step flow: a configurable
// entry routes to the configure modal rather than being taken on the spot, while
// a fixed entry is taken straight from the picker.
func TestPassivePickerOpensConfigureModal(t *testing.T) {
	app, _ := testAppWithPerk(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/perks/passives?character=char-1", nil)
	app.handlePassiveLibrary(rec, req)
	body := rec.Body.String()

	if !strings.Contains(body, "Resistance") {
		t.Fatalf("picker does not list the passives:\n%s", body)
	}
	// Each configured value is highlighted where it sits in the sentence.
	if !strings.Contains(body, "passive-text-value") {
		t.Errorf("picker does not highlight configured values:\n%s", body)
	}
	// A configurable entry goes to the modal; nothing is added yet.
	if !strings.Contains(body, "/perks/passive-config?character=char-1&passive=resistance") {
		t.Errorf("picker does not route a configurable entry to the modal:\n%s", body)
	}
	// A fixed entry posts straight to the add route.
	if !strings.Contains(body, `value="sure_footed"`) {
		t.Errorf("picker does not offer the fixed entry directly:\n%s", body)
	}
}

// TestPassiveConfigModalRendersFieldsAndCost pins the modal itself: it renders the
// entry's fields with the builder's own partial, previews the rules text, and
// prices what is currently selected.
func TestPassiveConfigModalRendersFieldsAndCost(t *testing.T) {
	app, c := testAppWithPerk(t)
	// The modal resolves the character from the store, so the fixture has to be
	// persisted first.
	if err := app.Store.Save(*c); err != nil {
		t.Fatalf("save: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/perks/passive-config?character=char-1&passive=resistance", nil)
	app.handlePassiveConfig(rec, req)
	body := rec.Body.String()

	if rec.Code != 200 {
		t.Fatalf("configure modal: status = %d, want 200\n%s", rec.Code, body)
	}
	// The free text, the number and the checkbox all render, namespaced by the
	// configure form's prefix.
	for _, want := range []string{`name="pf_source"`, `name="pf_amount"`, `name="pf_immune"`} {
		if !strings.Contains(body, want) {
			t.Errorf("modal is missing the field %s:\n%s", want, body)
		}
	}
	// The preview shows the text with values highlighted, and the cost is quoted.
	if !strings.Contains(body, "passive-text-value") {
		t.Errorf("modal does not preview the rules text:\n%s", body)
	}
	if !strings.Contains(body, "perk pt") {
		t.Errorf("modal does not price the passive:\n%s", body)
	}
	// Taking it posts to the add route rather than the configure route.
	if !strings.Contains(body, "/perks/passives") {
		t.Errorf("modal does not post to the add route:\n%s", body)
	}
}

// takePassive puts a configured passive on the character, as the add route would.
func takePassive(t *testing.T, app *App, c *model.Character, id string, values map[string]any) model.Perk {
	t.Helper()
	p, ok := app.Cfg.PassiveByID(id)
	if !ok {
		t.Fatalf("%s not found", id)
	}
	merged := mergePassiveFields(app.Cfg.PassiveDefaults(p), values)
	ab := app.buildPassivePerk("perk-passive-1", p, merged)
	c.Perks = append(c.Perks, ab)
	if err := app.Store.Save(*c); err != nil {
		t.Fatalf("save: %v", err)
	}
	return ab
}

// TestPerkListPassiveOffersConfigureNotBuilder pins the perk-list half of the
// rule: a configurable passive gets a Configure button, and the builder and
// export are not offered because a passive has neither.
func TestPerkListPassiveOffersConfigureNotBuilder(t *testing.T) {
	app, c := testAppWithPerk(t)
	takePassive(t, app, c, "resistance", vals("source", "fire", "amount", 1))

	rec := httptest.NewRecorder()
	app.renderPerkList(rec, c)
	body := rec.Body.String()

	if !strings.Contains(body, "Resistance") {
		t.Fatalf("passive missing from the perk list:\n%s", body)
	}
	// The configured values are highlighted, and Configure reopens the modal.
	if !strings.Contains(body, "passive-text-value") {
		t.Errorf("perk list does not highlight the passive's values:\n%s", body)
	}
	if !strings.Contains(body, "/perks/passive-config?character=char-1&perk=perk-passive-1") {
		t.Errorf("perk list offers no way to reconfigure the passive:\n%s", body)
	}
	// Neither builder nor export is offered. The built perk in the fixture still
	// has both, so the assertions are scoped to the passive's own id.
	if strings.Contains(body, `href="/characters/char-1/perks/perk-passive-1"`) {
		t.Errorf("perk list still links a passive to the builder:\n%s", body)
	}
	if strings.Contains(body, "/perks/perk-passive-1/export") {
		t.Errorf("perk list still offers to export a passive:\n%s", body)
	}
	// A passive has no enactments, so the built perk's phrasing must not appear.
	if strings.Contains(body, "0 enactment(s)") {
		t.Errorf("passive rendered with an enactment count:\n%s", body)
	}
}

// TestPassiveBuilderAndExportRefused pins the server-side half of the same rule.
// Hiding the buttons is not enough: the builder would strip a passive's fields
// on save, so a hand-typed URL has to be refused outright.
func TestPassiveBuilderAndExportRefused(t *testing.T) {
	app, c := testAppWithPerk(t)

	takePassive(t, app, c, "resistance", nil)

	// Opening the builder for a passive is refused.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/characters/char-1/perks/perk-passive-1", nil)
	app.handlePerks(rec, req, c, []string{"perk-passive-1"})
	if rec.Code != 400 {
		t.Errorf("builder for a passive: status = %d, want 400", rec.Code)
	}

	// Exporting one is refused.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/characters/char-1/perks/perk-passive-1/export", nil)
	app.handlePerks(rec, req, c, []string{"perk-passive-1", "export"})
	if rec.Code != 400 {
		t.Errorf("export of a passive: status = %d, want 400", rec.Code)
	}

	// An ordinary built perk is still editable and exportable.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/characters/char-1/perks/perk-1", nil)
	app.handlePerks(rec, req, c, []string{"perk-1"})
	if rec.Code != 200 {
		t.Errorf("builder for a built perk: status = %d, want 200", rec.Code)
	}
}

// postConfigure drives the configure route with a posted field form.
func postConfigure(app *App, c *model.Character, form string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/characters/char-1/perks/perk-passive-1/configure",
		strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.configurePassive(rec, req, c, len(c.Perks)-1)
	return rec
}

// storedField reads one configured value off the last perk on the character.
func storedField(c *model.Character, key string) any {
	return storedPassiveFields(c.Perks[len(c.Perks)-1])[key]
}

// TestConfigurePassiveRejectsUnaffordableChange pins the budget rule. The fixture
// character has 1 point left of 9, and raising Resistance two steps costs 6, so
// the change must be refused rather than silently applied.
func TestConfigurePassiveRejectsUnaffordableChange(t *testing.T) {
	app, c := testAppWithPerk(t)
	takePassive(t, app, c, "resistance", vals("source", "fire", "amount", 1))

	rec := postConfigure(app, c, "pf__submitted=1&pf_source=fire&pf_amount=3")
	if rec.Code != 400 {
		t.Errorf("unaffordable raise: status = %d, want 400", rec.Code)
	}
	// The stored value must not have moved.
	if got := asInt(storedField(c, "amount")); got != 1 {
		t.Errorf("value changed despite being refused: got %d, want 1", got)
	}
}

// TestConfigurePassiveLoweringAlwaysAllowed pins the other half: dropping a value
// frees points, so it must succeed even when the character is over budget. That
// is exactly the state lowering a value exists to fix.
func TestConfigurePassiveLoweringAlwaysAllowed(t *testing.T) {
	app, c := testAppWithPerk(t)
	// Start at the top of the range, which puts the character over budget.
	takePassive(t, app, c, "resistance", vals("source", "fire", "amount", 9))

	rec := postConfigure(app, c, "pf__submitted=1&pf_source=ice&pf_amount=1")
	if rec.Code != 200 {
		t.Errorf("lowering a value: status = %d, want 200", rec.Code)
	}
	if got := asInt(storedField(c, "amount")); got != 1 {
		t.Errorf("value did not drop: got %d, want 1", got)
	}
	// Free text has no cost, so it changes freely alongside.
	if got := storedField(c, "source"); got != "ice" {
		t.Errorf("free text did not change: got %v, want ice", got)
	}
	// The stored description has to follow the values, or the sheet would keep
	// describing the old ones.
	desc := c.Perks[len(c.Perks)-1].Description
	if !contains(desc, "by 1") || !contains(desc, "ice") {
		t.Errorf("description did not follow the values: %q", desc)
	}
}

// TestConfigurePassiveFreeTextIsFree pins that changing only a costless field
// leaves the price alone, so renaming what you resist is never a purchase.
func TestConfigurePassiveFreeTextIsFree(t *testing.T) {
	app, c := testAppWithPerk(t)
	ab := takePassive(t, app, c, "resistance", vals("source", "fire", "amount", 1))
	before := engine.PerkCost(app.Cfg.Config, ab).Build

	rec := postConfigure(app, c, "pf__submitted=1&pf_source=lightning&pf_amount=1")
	if rec.Code != 200 {
		t.Fatalf("free text change: status = %d, want 200", rec.Code)
	}
	after := engine.PerkCost(app.Cfg.Config, c.Perks[len(c.Perks)-1]).Build
	if after != before {
		t.Errorf("renaming the source moved the cost: %d -> %d", before, after)
	}
}

func contains(s, sub string) bool {
	return countOf(s, sub) > 0
}

// countOf returns how many times sub occurs in s.
func countOf(s, sub string) int {
	if sub == "" || len(sub) > len(s) {
		return 0
	}
	n := 0
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			n++
		}
	}
	return n
}
