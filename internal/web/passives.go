// The passive picker and its configure step.
//
// A passive is stored as an ordinary model.Perk of the passive perk type, with
// the catalogue id and the configured field values in its Fields. That is what
// lets the perk list, the budget, the printable sheet and the export handle a
// passive without any of them knowing it is special; the only passive-aware code
// is here and in the cost engine's one call into the generic field coster.
//
// The flow has two steps, deliberately:
//
//	pick  - the picker lists the catalogue at default values and prices.
//	config - an entry with fields opens a modal to set them, then confirms.
//
// An entry with no fields skips the second step entirely, and the same modal is
// reused afterwards to change a passive already on the character. Keeping the
// choice of passive separate from the choice of how far to invest in it means the
// price in the picker is always the price of taking it.
package web

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/engine"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// The field keys a passive perk stores. They are referenced by the config
// (ability_types.yaml) and by the cost engine, so they are named once here.
const (
	passiveIDField = "passive_id"
	// passiveFieldsField is the nested map of the entry's own field values. It
	// is nested so a passive's field keys cannot collide with the perk-type
	// fields around them.
	passiveFieldsField = "passive_fields"
	// passivePerkType is the perk-type id a selected passive is stored under.
	passivePerkType = "passive"
	// passiveFieldPrefix namespaces the entry's fields in the configure form,
	// matching the convention the builder uses for its own field regions.
	passiveFieldPrefix = "pf_"
	// passiveSubmitMarker is a hidden field the configure form always sends. It
	// is how a posted form is told apart from a fresh open even when every field
	// is an unticked checkbox, which browsers omit from the submission entirely.
	passiveSubmitMarker = passiveFieldPrefix + "_submitted"
)

// passiveRow is one catalogue entry as the picker shows it: the entry itself,
// what it costs at its defaults, and why it may not be taken.
type passiveRow struct {
	Passive config.Passive

	// Segments is the rules text split so each configured value can be
	// highlighted where it sits in the sentence.
	Segments []config.DescriptionSegment

	// Cost is the perk-point cost of taking this passive at its defaults.
	Cost int

	// Configurable reports whether selecting this entry opens the configure
	// modal. A fixed passive is added straight away.
	Configurable bool

	// Taken reports that the character already owns this passive. With
	// allow_duplicates false it cannot be taken again.
	Taken bool

	// Affordable reports whether the remaining perk points cover the cost.
	Affordable bool
}

// Selectable reports whether the Select button should be enabled. It is the
// single place the two blocking rules are combined, so the template does not
// have to restate them.
func (r passiveRow) Selectable() bool {
	return !r.Taken && r.Affordable
}

// passiveLibraryPage is the data envelope for the passive picker.
type passiveLibraryPage struct {
	CharacterID string
	Information string
	Rows        []passiveRow

	// Name is the player-typed name carried through from the New Perk modal
	// (or the perk-list rename flow), so whichever entry is picked next opens
	// already carrying it forward as the perk's name.
	Name string

	// AllowDuplicates mirrors the config rule, so the picker can explain why an
	// owned passive is greyed out.
	AllowDuplicates bool

	// HasCharacter reports whether budget figures are available. When false the
	// picker hides the remaining-points header and blocks nothing.
	HasCharacter bool
	charStats
}

// handlePassiveLibrary renders the passive picker. It expects a "character"
// query parameter so the buttons post to the right route and so each entry can be
// checked against that character's remaining perk points and owned passives. An
// optional "name" query parameter carries a player-typed name (from the New
// Perk modal) forward to whichever entry is picked next.
func (a *App) handlePassiveLibrary(w http.ResponseWriter, r *http.Request) {
	charID := r.URL.Query().Get("character")

	name := r.URL.Query().Get("name")
	if name == "" {
		// "Back to the list" from the configure modal includes that form's
		// own passive_name field via hx-include, so a name typed there before
		// backing out is not lost.
		name = r.URL.Query().Get("passive_name")
	}

	data := passiveLibraryPage{
		CharacterID:     charID,
		Information:     a.Cfg.Passives.Information,
		Name:            name,
		AllowDuplicates: a.Cfg.Passives.AllowsDuplicates(),
	}

	var char *model.Character
	if charID != "" {
		if c, ok := a.Store.Get(charID); ok {
			char = &c
			data.HasCharacter = true
			data.charStats = a.characterStats(&c)
		}
	}
	perkLeft := data.PerkBudget - data.PerkUsed

	for _, p := range a.Cfg.Passives.Entries {
		// Every entry is listed at its defaults: configuring happens in the
		// modal, so the price here is always the price of taking it.
		defaults := a.Cfg.PassiveDefaults(p)
		row := passiveRow{
			Passive:      p,
			Segments:     a.Cfg.PassiveDescriptionSegments(p, defaults),
			Cost:         a.passiveCost(p, defaults),
			Configurable: p.Configurable(),
			Affordable:   true,
		}
		if char != nil {
			row.Taken = !data.AllowDuplicates && characterHasPassive(char, p.ID)
			row.Affordable = row.Cost <= perkLeft
		}
		data.Rows = append(data.Rows, row)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.Tmpl.ExecuteTemplate(w, "passive_library", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// passiveCost prices a passive at the given field values: the entry's flat cost
// plus whatever its fields add. It is the same sum the cost engine performs on a
// stored perk, kept in one place so the picker, the modal and the perk list can
// never quote different figures.
func (a *App) passiveCost(p config.Passive, values map[string]any) int {
	return a.Cfg.PassiveFlatCost(p) + engine.FieldsCost(a.Cfg.Config, p.Fields, values).Build
}

// passiveConfigPage is the data envelope for the configure modal, used both when
// taking a passive and when changing one already owned.
type passiveConfigPage struct {
	// Cfg is needed because the modal renders the entry's fields with the
	// builder's own "field" partial, which resolves option sources from it.
	Cfg         *config.Config
	CharacterID string
	Passive     config.Passive

	// Name is the perk name the form's Name field renders from. Priority
	// order: the posted passive_name (a live re-render as the player types),
	// the perk's own stored Name (editing an owned passive), the carried-
	// through query name (a fresh pick, name typed in the New Perk modal),
	// falling back to the catalogue entry's own Name.
	Name string

	// Values are the current field values the form renders from.
	Values map[string]any

	// Segments is the rules text at those values, so the modal shows what the
	// passive will actually say before it is confirmed.
	Segments []config.DescriptionSegment

	// Cost is the total perk-point cost at those values. Delta is what
	// confirming would change the character's spend by: the full cost when
	// taking a new passive, the difference when editing one.
	Cost  int
	Delta int

	// PerkID is set when editing an owned passive, empty when taking a new one.
	// It decides which route the form posts to.
	PerkID string

	// Affordable reports whether the character can cover Delta.
	Affordable bool
	// Remaining is the unspent perk points, shown alongside the cost.
	Remaining int
}

// Editing reports whether the modal is changing an owned passive rather than
// taking a new one.
func (p passiveConfigPage) Editing() bool { return p.PerkID != "" }

// handlePassiveConfig renders the configure modal. It serves three cases with one
// handler because they differ only in where the values come from:
//
//	passive=<id>              - taking a new passive, at its defaults
//	passive=<id> + posted form - live re-render as fields change
//	perk=<perk id>            - editing an owned passive, at its stored values
//
// The live re-render is what keeps the cost and the rules text honest while the
// player is still deciding.
func (a *App) handlePassiveConfig(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	charID := r.FormValue("character")
	if charID == "" {
		charID = r.URL.Query().Get("character")
	}
	c, ok := a.Store.Get(charID)
	if !ok {
		http.Error(w, "unknown character", http.StatusBadRequest)
		return
	}

	data := passiveConfigPage{Cfg: a.Cfg.Config, CharacterID: charID}

	// Editing an owned passive: resolve the entry and the stored values from the
	// perk itself.
	perkID := r.FormValue("perk")
	if perkID == "" {
		perkID = r.URL.Query().Get("perk")
	}
	var current int
	var storedName string
	if perkID != "" {
		idx := findPerk(&c, perkID)
		if idx < 0 {
			http.NotFound(w, r)
			return
		}
		p, ok := a.Cfg.PassiveByID(passiveIDOf(c.Perks[idx]))
		if !ok {
			http.Error(w, "that perk is not a configurable passive.", http.StatusBadRequest)
			return
		}
		data.Passive = p
		data.PerkID = perkID
		storedName = c.Perks[idx].Name
		// What it costs today, so the modal can price the change rather than the
		// whole passive.
		current = a.passiveCost(p, storedPassiveFields(c.Perks[idx]))
	} else {
		p, ok := a.Cfg.PassiveByID(r.FormValue("passive"))
		if !ok {
			p, ok = a.Cfg.PassiveByID(r.URL.Query().Get("passive"))
			if !ok {
				http.Error(w, "unknown passive", http.StatusBadRequest)
				return
			}
		}
		data.Passive = p
	}

	// The name the form renders from. Priority: a posted passive_name (typed
	// this round, live re-render), the perk's own stored name (editing an
	// owned passive), the carried-through query name (a fresh pick, typed in
	// the New Perk modal), then the catalogue entry's own name.
	if v := r.FormValue("passive_name"); v != "" {
		data.Name = v
	} else if storedName != "" {
		data.Name = storedName
	} else if v := r.URL.Query().Get("name"); v != "" {
		data.Name = v
	} else {
		data.Name = data.Passive.Name
	}

	// Values come from the posted form when there is one (a live re-render or a
	// reopened modal), otherwise from the stored perk, otherwise the defaults.
	data.Values = a.readPassiveFields(data.Passive, r)
	if !passiveFieldsPosted(data.Passive, r) {
		if perkID != "" {
			if idx := findPerk(&c, perkID); idx >= 0 {
				data.Values = mergePassiveFields(a.Cfg.PassiveDefaults(data.Passive),
					storedPassiveFields(c.Perks[idx]))
			}
		} else {
			data.Values = a.Cfg.PassiveDefaults(data.Passive)
		}
	}

	data.Segments = a.Cfg.PassiveDescriptionSegments(data.Passive, data.Values)
	data.Cost = a.passiveCost(data.Passive, data.Values)
	data.Delta = data.Cost - current

	stats := a.characterStats(&c)
	data.Remaining = stats.PerkBudget - stats.PerkUsed
	// Only an increase has to fit the budget. A change that costs the same or
	// less is always allowed, which matters when a character is already over
	// budget: lowering a value is the way back under it.
	data.Affordable = data.Delta <= 0 || data.Delta <= data.Remaining

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.Tmpl.ExecuteTemplate(w, "passive_config", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// readPassiveFields parses an entry's fields out of a posted configure form. It
// delegates to the builder's own field parser, so every field type behaves
// exactly as it does in the perk builder, then clamps numbers to their declared
// range so a hand-made post cannot buy an out-of-range value.
func (a *App) readPassiveFields(p config.Passive, r *http.Request) map[string]any {
	out := readFieldValues(a.Cfg.Config, p.Fields, passiveFieldPrefix, r)
	for _, f := range p.Fields {
		if f.Type == "free_number" {
			out[f.Key] = config.ClampFieldNumber(f, asInt(out[f.Key]))
		}
	}
	return out
}

// passiveFieldsPosted reports whether the request actually carried this entry's
// fields. A checkbox is absent from the form when unticked, so presence is
// judged on any non-checkbox field being present; an entry made only of
// checkboxes falls back to a marker the form always sends.
func passiveFieldsPosted(p config.Passive, r *http.Request) bool {
	if r.Form == nil {
		return false
	}
	if _, ok := r.Form[passiveSubmitMarker]; ok {
		return true
	}
	for _, f := range p.Fields {
		if f.Type == "checkbox" {
			continue
		}
		if _, ok := r.Form[passiveFieldPrefix+f.Key]; ok {
			return true
		}
	}
	return false
}

// storedPassiveFields returns the configured field values of a passive perk, or
// nil when it has none.
func storedPassiveFields(ab model.Perk) map[string]any {
	if ab.Fields == nil {
		return nil
	}
	v, _ := ab.Fields[passiveFieldsField].(map[string]any)
	return v
}

// mergePassiveFields overlays stored values onto the entry's defaults, so a
// passive stored before a field was added still renders every field.
func mergePassiveFields(defaults, stored map[string]any) map[string]any {
	out := make(map[string]any, len(defaults))
	for k, v := range defaults {
		out[k] = v
	}
	for k, v := range stored {
		if _, declared := out[k]; declared {
			out[k] = v
		}
	}
	return out
}

// characterHasPassive reports whether the character already owns the passive
// with the given catalogue id.
func characterHasPassive(c *model.Character, id string) bool {
	for _, ab := range c.Perks {
		if passiveIDOf(ab) == id {
			return true
		}
	}
	return false
}

// passiveIDOf returns the catalogue id a perk refers to, or "" when the perk is
// not a passive.
func passiveIDOf(ab model.Perk) string {
	if ab.Fields == nil {
		return ""
	}
	id, _ := ab.Fields[passiveIDField].(string)
	return id
}

// addPassive appends a selected passive to the character as a normal perk.
//
// An entry with fields arrives here from the configure modal with those fields
// posted; a fixed entry arrives straight from the picker and is taken at its
// (empty) defaults. Both blocking rules are enforced here rather than only in the
// UI, because a hand-made post must be rejected too.
func (a *App) addPassive(w http.ResponseWriter, r *http.Request, c *model.Character) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()

	id := r.FormValue(passiveIDField)
	if id == "" {
		http.Error(w, "missing passive id", http.StatusBadRequest)
		return
	}
	p, ok := a.Cfg.PassiveByID(id)
	if !ok {
		http.Error(w, fmt.Sprintf("unknown passive %q", id), http.StatusBadRequest)
		return
	}
	if !a.Cfg.Passives.AllowsDuplicates() && characterHasPassive(c, id) {
		http.Error(w, fmt.Sprintf("%q is already taken and cannot be taken twice.", p.Name), http.StatusBadRequest)
		return
	}

	values := a.Cfg.PassiveDefaults(p)
	if p.Configurable() && passiveFieldsPosted(p, r) {
		values = a.readPassiveFields(p, r)
	}

	name := strings.TrimSpace(r.FormValue("passive_name"))
	ab := a.buildPassivePerk(fmt.Sprintf("perk-%d", time.Now().UnixNano()), p, values, name)
	if ok, reason := a.affordPerk(c, ab); !ok {
		http.Error(w, reason, http.StatusBadRequest)
		return
	}
	c.Perks = append(c.Perks, ab)
	if err := a.Store.Save(*c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/characters/"+c.ID+"?t=perks#tab-perks", http.StatusSeeOther)
}

// buildPassivePerk assembles the stored form of a configured passive. name is
// the player-typed name; when empty the catalogue entry's own name is used,
// matching how every passive worked before naming was added.
func (a *App) buildPassivePerk(id string, p config.Passive, values map[string]any, name string) model.Perk {
	if name == "" {
		name = p.Name
	}
	ab := model.Perk{
		ID:   id,
		Name: name,
		// The stored description has every placeholder already substituted, so
		// the perk list, the printed sheet and the export read correctly without
		// resolving the catalogue again.
		Description: a.Cfg.PassiveDescription(p, values),
		Type:        passivePerkType,
		Fields: map[string]any{
			passiveIDField:     p.ID,
			passiveFieldsField: values,
		},
	}
	return engine.NormalizePerk(a.Cfg.Config, ab)
}

// configurePassive applies posted field values to a passive already on the
// character.
//
// The change is priced as a delta against what the passive costs today, so the
// points it already occupies are its own to keep; only an increase has to fit in
// what is left. A decrease always succeeds, which is what lets a character who
// is over budget get back under it.
func (a *App) configurePassive(w http.ResponseWriter, r *http.Request, c *model.Character, idx int) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()

	ab := c.Perks[idx]
	p, ok := a.Cfg.PassiveByID(passiveIDOf(ab))
	if !ok {
		http.Error(w, "that perk is not a passive, so it has nothing to configure.", http.StatusBadRequest)
		return
	}
	if !p.Configurable() {
		http.Error(w, fmt.Sprintf("%q has nothing to configure.", p.Name), http.StatusBadRequest)
		return
	}

	values := a.readPassiveFields(p, r)
	delta := a.passiveCost(p, values) - a.passiveCost(p, storedPassiveFields(ab))

	stats := a.characterStats(c)
	if left := stats.PerkBudget - stats.PerkUsed; delta > 0 && delta > left {
		http.Error(w, fmt.Sprintf(
			"That change to %s costs %d more perk point(s) but only %d remain.",
			p.Name, delta, left), http.StatusBadRequest)
		return
	}

	// A posted name renames the passive; leaving it blank keeps the name it
	// already has, rather than reverting to the catalogue's own name.
	name := strings.TrimSpace(r.FormValue("passive_name"))
	if name == "" {
		name = ab.Name
	}
	c.Perks[idx] = a.buildPassivePerk(ab.ID, p, values, name)
	if err := a.Store.Save(*c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// The whole list is returned rather than the single card: reconfiguring moves
	// the perk-point total, which every other card is measured against.
	a.renderPerkListPartial(w, c, 0)
}
