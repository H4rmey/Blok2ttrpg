package web

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/engine"
	"github.com/harmey/blok2ttrpg-v5/internal/export"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// perkPage is the data envelope for the builder and perk views.
type perkPage struct {
	Cfg         *config.Config
	Title       string
	Breadcrumbs []crumb
	Character   *model.Character
	Perk        *model.Perk
	Cost        engine.Cost
	Budget      int
	OverBudget  bool

	// charStats carries the character bar figures so the builder shows the same
	// level/points/vitals header as every other character-scoped page.
	charStats

	// ReadOnlyStats renders the bar without editable inputs. It stays true on
	// the library/browser pages that have no character form; the builder sets
	// it false so the bar is fully editable there, like on the sheet.
	ReadOnlyStats bool

	// Instructions is the generated play-facing rules text for the perk,
	// one entry per enactment. It is rendered by the "instructions" partial
	// and refreshed by /builder/instructions as the builder changes.
	Instructions []engine.Instruction
}

// handlePerks dispatches /characters/{id}/perks[/...] routes.
func (a *App) handlePerks(w http.ResponseWriter, r *http.Request, c *model.Character, rest []string) {
	// /perks -> the standalone list page no longer exists; the perk list is a
	// tab on the character page, so send the user straight there.
	if len(rest) == 0 {
		http.Redirect(w, r, "/characters/"+c.ID+"?t=perks#tab-perks", http.StatusSeeOther)
		return
	}

	// /perks/refresh -> re-normalize every perk against the current config,
	// persist the result, and return just the perk list region so the
	// "Refresh All" button can swap it in place. Persisting is the point: cost
	// is a pure function of the stored data, so a display-only refresh could
	// never change anything. This repairs perks stored before normalization or
	// under an older config.
	if rest[0] == "refresh" {
		changed := engine.NormalizeCharacter(a.Cfg.Config, c)
		if err := a.Store.Save(*c); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		quiet := r.URL.Query().Get("quiet") == "1"
		a.renderPerkListPartial(w, c, changed, quiet)
		return
	}

	// /perks/import        -> import a built-in perk by library id
	if rest[0] == "import" {
		a.importBuiltinPerk(w, r, c)
		return
	}

	// /perks/import-custom -> import an perk from an uploaded YAML file
	if rest[0] == "import-custom" {
		a.importPerk(w, r, c)
		return
	}

	// /perks/passives      -> add a predefined passive by catalogue id
	if rest[0] == "passives" {
		a.addPassive(w, r, c)
		return
	}

	// /perks/new    -> builder for a new perk
	if rest[0] == "new" {

		if r.Method == http.MethodPost {
			a.savePerk(w, r, c, "")
			return
		}
		// The name and type are collected up front via the New Perk modal and
		// passed as query parameters so the builder opens with both already
		// set and the rest of the form unlocked.
		name := strings.TrimSpace(r.URL.Query().Get("name"))
		ptype := r.URL.Query().Get("type")
		if ptype == passivePerkType {
			// The New Perk modal intercepts a Passive selection client-side
			// (see app.js) and never lets this request happen with
			// JavaScript enabled. This is the no-JS fallback: rather than
			// silently open a builder that cannot express a passive, send
			// the user back to the perk list.
			http.Redirect(w, r, "/characters/"+c.ID+"?t=perks#tab-perks", http.StatusSeeOther)
			return
		}
		if ptype == "" {
			ptype = firstPerkTypeID(a.Cfg.Config)
		}
		blank := model.Perk{Type: ptype, Name: name}
		a.renderBuilder(w, c, &blank, true)
		return

	}

	aid := rest[0]
	idx := findPerk(c, aid)
	if idx < 0 {
		http.NotFound(w, r)
		return
	}

	// A passive is a perk, but it is not built from enactments, interactions and
	// validations, so the builder cannot express it. Both the GET and POST
	// builder routes are refused here as well as hidden in the UI, so a
	// bookmarked or hand-typed URL cannot reach a builder that would silently
	// strip the passive's fields on save. Export is not refused: the perk as
	// stored (catalogue id, configured values, resolved description) is a
	// normal model.Perk and round-trips through the same YAML shape as a
	// built perk.
	isPassivePerk := passiveIDOf(c.Perks[idx]) != ""

	if len(rest) == 1 {
		switch r.Method {
		case http.MethodGet:
			if isPassivePerk {
				http.Error(w, "A passive is predefined and cannot be edited in the builder. Change its value from the perk list, or delete it and pick another.", http.StatusBadRequest)
				return
			}
			a.renderBuilder(w, c, &c.Perks[idx], false)
		case http.MethodPost:
			if isPassivePerk {
				http.Error(w, "A passive cannot be saved from the builder.", http.StatusBadRequest)
				return
			}
			a.savePerk(w, r, c, aid)
		case http.MethodDelete:
			// Deleting is always allowed: giving a passive back is how its perk
			// points are recovered.
			c.Perks = append(c.Perks[:idx], c.Perks[idx+1:]...)
			_ = a.Store.Save(*c)
			// Return the updated perk list partial so the card disappears in
			// place without a full page reload or tab jump.
			a.renderPerkListPartial(w, c, 0, true)
		}
		return
	}

	// /perks/{id}/configure -> apply the configure modal's field values to a
	// passive already owned. This is the only edit a passive supports, which is
	// why it is its own route rather than a trip through the builder.
	if rest[1] == "configure" {
		a.configurePassive(w, r, c, idx)
		return
	}

	if rest[1] == "export" {
		b, err := export.MarshalPerk(c.Perks[idx])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/x-yaml")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", c.Perks[idx].Name+".yaml"))
		w.Write(b)
		return
	}
	http.NotFound(w, r)
}

// importPerk reads an uploaded YAML file, parses it into an perk, gives
// it a fresh id, appends it to the character and redirects back to the list.
func (a *App) importPerk(w http.ResponseWriter, r *http.Request, c *model.Character) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "no file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ab, err := export.UnmarshalPerk(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Always assign a fresh id so an imported perk never collides with an
	// existing one on the character.
	ab.ID = fmt.Sprintf("perk-%d", time.Now().UnixNano())
	ab = engine.NormalizePerk(a.Cfg.Config, ab)
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

// perkSummaries recomputes cost and instruction text for every perk the
// character owns. Costs are always derived here rather than stored, so the
// figures follow the current config even for perks built long ago.
func (a *App) perkSummaries(c *model.Character) []perkSummary {
	// How many perk points are free once every perk (including this one) is paid
	// for. A passive's value change is priced as a delta against its current
	// cost, so the points it already occupies are available to it.
	stats := a.characterStats(c)
	remaining := stats.PerkBudget - stats.PerkUsed

	perks := make([]perkSummary, 0, len(c.Perks))
	for _, ab := range c.Perks {
		s := perkSummary{
			Perk: ab,
			Cost: engine.PerkCost(a.Cfg.Config, ab),
		}
		// A passive is not built from enactments, so it has no generated
		// instruction text: its rules are the catalogue entry's description.
		// Generating instructions for one would produce an empty block.
		if p := a.passiveSummaryFor(ab, remaining); p != nil {
			s.Passive = p
		} else {
			s.Instructions = engine.PerkInstructions(a.Cfg.Config, ab)
		}
		perks = append(perks, s)
	}
	return perks
}

// passiveSummaryFor builds the in-list view of a passive perk, or nil when the
// perk is not a passive. remaining is the character's unspent perk points, used
// to decide which value changes they can currently afford.
func (a *App) passiveSummaryFor(ab model.Perk, remaining int) *passiveSummary {
	id := passiveIDOf(ab)
	if id == "" {
		return nil
	}
	p, ok := a.Cfg.PassiveByID(id)
	if !ok {
		// The entry was removed from the config. The perk is still a passive, so
		// it must not be treated as a buildable one; it simply has nothing left
		// to configure. The stored description is all there is to show.
		return &passiveSummary{
			ID:       id,
			Name:     ab.Name,
			Segments: []config.DescriptionSegment{{Text: ab.Description}},
		}
	}

	// Stored values overlaid on the entry's defaults, so a passive saved before
	// a field was added still renders every field.
	values := mergePassiveFields(a.Cfg.PassiveDefaults(p), storedPassiveFields(ab))

	return &passiveSummary{
		ID:   p.ID,
		Name: p.Name,
		// The rules text is split so the template can highlight each configured
		// value in place rather than burying it in the sentence.
		Segments: a.Cfg.PassiveDescriptionSegments(p, values),
		// Only a passive with fields offers a Configure button; a fixed one has
		// nothing to open a modal for.
		Configurable: p.Configurable(),
	}
}

// asInt reads an int out of a stored field value. Values arrive as int from the
// form parser and as float64 from JSON, so both are accepted.
func asInt(v any) int {
	n, _ := atoiAny(v)
	return n
}

// perkListPage builds the envelope shared by the full Perks page and the
// perk-list partial returned by the refresh route.
func (a *App) perkListPage(c *model.Character) pageData {
	return pageData{
		Title:         c.Name() + " - Perks",
		Character:     c,
		Perks:         a.perkSummaries(c),
		charStats:     a.characterStats(c),
		ReadOnlyStats: true,
		Breadcrumbs: []crumb{
			{Label: "Home", URL: "/"},
			{Label: c.Name(), URL: "/characters/" + c.ID},
			{Label: "Perks", URL: "/characters/" + c.ID + "#tab-perks"},
		},
	}
}

func (a *App) renderPerkList(w http.ResponseWriter, c *model.Character) {
	a.render(w, "perks.html", a.perkListPage(c))
}

// refreshNotice phrases the outcome of a refresh for the user.
func refreshNotice(changed int) string {
	switch changed {
	case 0:
		return "Recalculated: all perk costs were already up to date."
	case 1:
		return "Recalculated: 1 perk was updated."
	default:
		return fmt.Sprintf("Recalculated: %d perks were updated.", changed)
	}
}

// renderPerkListPartial returns only the perk list region with freshly
// recomputed costs; used by the "Refresh All" button. changed is the number of
// perks whose cost moved during normalization, reported back so the user can see
// the refresh did something (or confirm everything was already correct).
//
// This is the HTMX fragment counterpart of renderPerkList above, which renders
// the whole page. The two were distinct before the Abilities -> Perks rename
// collapsed their names together.
func (a *App) renderPerkListPartial(w http.ResponseWriter, c *model.Character, changed int, quiet bool) {

	data := a.perkListPage(c)
	data.Cfg = a.Cfg.Config
	if !quiet {
		data.RefreshNotice = refreshNotice(changed)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.Tmpl.ExecuteTemplate(w, "perk_list", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) renderBuilder(w http.ResponseWriter, c *model.Character, ab *model.Perk, isNew bool) {
	cost := engine.PerkCost(a.Cfg.Config, *ab)
	budget := a.Cfg.PerkPointBudget(c.Level)
	title := "New Perk"
	if !isNew {
		title = ab.Name
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := a.Tmpl.ExecuteTemplate(w, "builder.html", perkPage{
		Cfg:       a.Cfg.Config,
		Title:     title,
		Character: c,
		Perk:      ab,
		Cost:      cost,
		Budget:    budget,
		// Over budget is advisory only: it never blocks saving.
		OverBudget: cost.Build > budget,

		charStats: a.characterStats(c),
		// The builder bar is fully editable, like the sheet: the same level
		// and current-vital inputs post to the same character endpoint, and
		// perk edits refresh the bar live without resetting Current to Max.
		ReadOnlyStats: false,

		Instructions: engine.PerkInstructions(a.Cfg.Config, *ab),

		Breadcrumbs: []crumb{
			{Label: "Home", URL: "/"},
			{Label: c.Name(), URL: "/characters/" + c.ID},
			{Label: "Perks", URL: "/characters/" + c.ID + "#tab-perks"},
			{Label: title, URL: "#"},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// savePerk parses the builder form into an perk and stores it. Cost is
// never validated here; over-budget perks are allowed by design.
func (a *App) savePerk(w http.ResponseWriter, r *http.Request, c *model.Character, existingID string) {
	_ = r.ParseForm()
	if existingID == "" {
		existingID = r.FormValue("perk_id")
	}
	// One shared parse path for save, autosave and the cost preview so the
	// three can never disagree about what the form said.
	ab := a.buildPerkFromForm(r, existingID)

	// Normalize on the way in so the stored perk is canonical: every configured
	// field present, repeatable fields expanded, numbers in range. The cost
	// engine and the builder then read the same values and cannot disagree.
	ab = engine.NormalizePerk(a.Cfg.Config, ab)
	if idx := findPerk(c, existingID); idx >= 0 {
		c.Perks[idx] = ab
	} else {
		c.Perks = append(c.Perks, ab)
	}
	if err := a.Store.Save(*c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/characters/"+c.ID+"?t=perks#tab-perks", http.StatusSeeOther)
}

// handleBuilderEnactment returns an enactment form partial for a given index.

func (a *App) handleBuilderEnactment(w http.ResponseWriter, r *http.Request) {
	idx := r.URL.Query().Get("index")
	etype := r.URL.Query().Get("type")
	if etype == "" {
		// When re-rendering after an interaction change we keep the posted
		// enactment type from the form; fall back to the first type otherwise.
		etype = r.FormValue(fmt.Sprintf("en%s_type", idx))
	}
	if etype == "" {
		etype = firstEnactmentID(a.Cfg.Config)
	}
	interaction := r.URL.Query().Get("interaction")
	if interaction == "" {
		interaction = r.FormValue(fmt.Sprintf("en%s_interaction", idx))
	}
	if interaction == "" {
		interaction = firstInteractionID(a.Cfg.Config)
	}
	// The perk type drives which enactments are offered (allowed/blocked
	// filtering) and, for enactments beyond the first, whether the Interaction
	// and Validation regions are shown.
	atype := r.URL.Query().Get("atype")
	if atype == "" {
		atype = r.FormValue("type")
	}
	data := map[string]any{
		"Cfg":             a.Cfg.Config,
		"Index":           idx,
		"PerkType":        atype,
		"Type":            etype,
		"Interaction":     interaction,
		"Fields":          map[string]any{},
		"InteractionData": map[string]any{},
		"ValidationData":  map[string]any{},
		// A freshly added enactment inherits the previous enactment's target,
		// so the "different target" box starts unchecked unless the form
		// already had it ticked for this index.
		"NewTarget": r.FormValue(fmt.Sprintf("en%s_new_target", idx)) == "on",
	}

	var buf bytes.Buffer

	if err := a.Tmpl.ExecuteTemplate(&buf, "enactment_partial.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = buf.WriteTo(w)
}

// handleEnactmentFields renders just the config-driven fields for the selected
// enactment type. Only the enactment field sub-region is swapped, so changing
// the enactment type leaves the Validation and Interaction regions untouched.
func (a *App) handleEnactmentFields(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	idx := r.URL.Query().Get("index")
	etype := r.FormValue(fmt.Sprintf("en%s_type", idx))
	if etype == "" {
		etype = r.URL.Query().Get("type")
	}
	comp, ok := a.Cfg.Enactment(etype)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if !ok {
		return
	}
	prefix := fmt.Sprintf("en%s_f_", idx)
	for _, f := range comp.Fields {
		data := map[string]any{"Cfg": a.Cfg.Config, "Field": f, "Prefix": prefix}
		if err := a.Tmpl.ExecuteTemplate(w, "field", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// handleInteractionFields renders just the config-driven fields for the
// selected interaction type. Only the interaction field sub-region is swapped,
// so changing the interaction type leaves the Validation region untouched.
func (a *App) handleInteractionFields(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	idx := r.URL.Query().Get("index")
	itype := r.FormValue(fmt.Sprintf("en%s_interaction", idx))
	if itype == "" {
		itype = r.URL.Query().Get("interaction")
	}
	comp, ok := a.Cfg.Interaction(itype)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if !ok {
		return
	}
	prefix := fmt.Sprintf("en%s_i_", idx)
	for _, f := range comp.Fields {
		data := map[string]any{"Cfg": a.Cfg.Config, "Field": f, "Prefix": prefix}
		if err := a.Tmpl.ExecuteTemplate(w, "field", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// handleInlineFields renders the nested fields of the component referenced by
// an inline_builder dropdown. It parallels handleEnactmentFields but resolves
// the component generically by kind (enactment/interaction/perk_type) and
// renders its fields under the "<name>_ib_" prefix so they stay namespaced.
func (a *App) handleInlineFields(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	name := r.URL.Query().Get("name")
	kind := r.URL.Query().Get("kind")
	// The selected value is the current value of the dropdown, posted under
	// its own name by htmx.
	value := r.FormValue(name)
	if value == "" {
		value = r.URL.Query().Get("value")
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if value == "" {
		return
	}
	comp, ok := a.Cfg.ComponentByKind(kind, value)
	if !ok {
		return
	}
	prefix := name + "_ib_"
	for _, f := range comp.Fields {
		data := map[string]any{"Cfg": a.Cfg.Config, "Field": f, "Prefix": prefix}
		if err := a.Tmpl.ExecuteTemplate(w, "field", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// handleConditionShift renders the shift-amount dropdown for a condition_select field.
// The selected condition (posted under the field's own name, namespaced as
// "general.<id>" or "specific.<id>") drives what is returned: for a general
// condition the field shows a dropdown of that condition's configured non-zero shift
// range; for a specific condition (or none) nothing is rendered, so the shift row
// disappears. The dropdown is named "<name>_shift" so it posts as the sibling
// shift field the cost engine reads (ShiftKey defaults to "shift_amount", and
// the field name is "<prefix>shift_amount").
func (a *App) handleConditionShift(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	name := r.URL.Query().Get("name")
	shiftName := r.URL.Query().Get("shift_name")
	value := r.FormValue(name)
	if value == "" {
		value = r.URL.Query().Get("value")
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if value == "" {
		return
	}
	// Legacy namespaced values use a "general."/"specific." prefix; unified
	// values are bare condition ids. Specific (fixed-cost) conditions have no
	// shift dropdown, so strip a "general." prefix and reject "specific.".
	id := value
	if strings.HasPrefix(value, "specific.") {
		return
	}
	id = strings.TrimPrefix(id, "general.")
	shifts := a.Cfg.ShiftOptionsFor(id)
	if len(shifts) == 0 {
		// Fixed-cost condition (or unknown): no shift dropdown.
		return
	}
	cost := ""
	if s, ok := a.Cfg.GeneralConditionByID(id); ok {
		cost = costHintStr(&s.ShiftCost)
	} else if u, ok := a.Cfg.ConditionByID(id); ok {
		cost = costHintStr(&u.ShiftCost)
	}

	var buf bytes.Buffer
	buf.WriteString(`<label>Shift Amount `)
	if cost != "" {
		buf.WriteString(`<span class="cost-hint">` + cost + ` / shift</span>`)
	}
	buf.WriteString(`</label>`)
	buf.WriteString(`<select name="` + name2attr(shiftName) + `">`)
	for _, n := range shifts {
		sign := ""
		if n > 0 {
			sign = "+"
		}
		buf.WriteString(fmt.Sprintf(`<option value="%d">%s%d</option>`, n, sign, n))
	}
	buf.WriteString(`</select>`)
	_, _ = buf.WriteTo(w)
}

// name2attr is a tiny guard that keeps a form-field name safe for direct
// embedding in an HTML trait. Builder field names are already limited to
// [A-Za-z0-9_], so this only strips the double-quote character defensively.
func name2attr(s string) string {
	return strings.ReplaceAll(s, `"`, "")
}

// handlePerkTypeFields renders the config-driven fields for the selected
// perk type, used to swap the type-specific field block in the builder.

func (a *App) handlePerkTypeFields(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	atype := r.FormValue("type")
	if atype == "" {
		atype = firstPerkTypeID(a.Cfg.Config)
	}
	comp, ok := a.Cfg.PerkType(atype)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if !ok {
		return
	}
	for _, f := range comp.Fields {
		data := map[string]any{"Cfg": a.Cfg.Config, "Field": f, "Prefix": "atype_"}
		if err := a.Tmpl.ExecuteTemplate(w, "field", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// handleBuilderAutosave saves the perk from the posted builder form without
// redirecting, so the builder can persist edits in the background as the user
// selects options. It returns the (possibly newly created) perk id in the
// HX-Trigger-independent JSON body and an HX-Perk-ID header so the client
// can keep posting to a stable URL. Autosave is skipped until the perk has
// a name, mirroring the manual save gate.
func (a *App) handleBuilderAutosave(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	charID := r.FormValue("character_id")
	c, ok := a.Store.Get(charID)
	if !ok {
		http.Error(w, "unknown character", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		// Nothing to save yet; the name gate still applies.
		w.WriteHeader(http.StatusNoContent)
		return
	}
	existingID := r.FormValue("perk_id")
	ab := engine.NormalizePerk(a.Cfg.Config, a.buildPerkFromForm(r, existingID))
	if idx := findPerk(&c, existingID); idx >= 0 {
		c.Perks[idx] = ab
	} else {
		c.Perks = append(c.Perks, ab)
	}
	if err := a.Store.Save(c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("X-Perk-ID", ab.ID)
	w.Header().Set("Content-Type", "application/json")
	// Return the refreshed character bar alongside the perk id: perk edits
	// change PerkUsed, so the bar goes stale unless it is re-rendered. The
	// current-vital inputs are preserved client-side (see app.js), so a live
	// update never resets Current back to Max.
	bar, err := a.renderStatCards(c, a.characterStats(&c))
	if err != nil {
		fmt.Fprintf(w, `{"id":%q}`, ab.ID)
		return
	}
	fmt.Fprintf(w, `{"id":%q,"stat_cards":%q}`, ab.ID, bar)
}

// renderStatCards renders the stat_cards partial for a character to a string.
// Used by the builder autosave to push live bar updates without a reload.
func (a *App) renderStatCards(c model.Character, stats charStats) (string, error) {
	data := a.characterPage(&c, false)
	var buf strings.Builder
	if err := a.Tmpl.ExecuteTemplate(&buf, "stat_cards", data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// buildPerkFromForm parses the builder form into an Perk. Shared by the
// manual save and the background autosave paths.
func (a *App) buildPerkFromForm(r *http.Request, existingID string) model.Perk {
	ab := model.Perk{
		ID:          existingID,
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		Type:        r.FormValue("type"),
		Fields:      map[string]any{},
	}
	if ab.ID == "" {
		ab.ID = fmt.Sprintf("perk-%d", time.Now().UnixNano())
	}
	if at, ok := a.Cfg.PerkType(ab.Type); ok {
		ab.Fields = readFieldValues(a.Cfg.Config, at.Fields, "atype_", r)
	}
	ab.Enactments = a.readEnactments(r)
	return ab
}

// readEnactments parses the posted enactment blocks of the builder form. They
// arrive as an "enactment_count" plus per-index "en<i>_" prefixed values. It is
// the single parse path shared by save, autosave and the cost/instruction
// previews so those can never disagree about what the form said.
func (a *App) readEnactments(r *http.Request) []model.Enactment {
	count, _ := strconv.Atoi(r.FormValue("enactment_count"))
	var out []model.Enactment
	for i := 0; i < count; i++ {
		prefix := fmt.Sprintf("en%d_", i)
		etype := r.FormValue(prefix + "type")
		if etype == "" {
			continue
		}
		en := model.Enactment{Type: etype}
		if ec, ok := a.Cfg.Enactment(etype); ok {
			en.Fields = readFieldValues(a.Cfg.Config, ec.Fields, prefix+"f_", r)
		}
		// The first enactment always owns its target, so its checkbox is not
		// rendered and the flag stays false there; later enactments opt in.
		if v := r.FormValue(prefix + "new_target"); v == "on" || v == "true" {
			en.NewTarget = true
		}
		en.Interaction = r.FormValue(prefix + "interaction")
		if ic, ok := a.Cfg.Interaction(en.Interaction); ok {
			en.InteractionData = readFieldValues(a.Cfg.Config, ic.Fields, prefix+"i_", r)
		}
		// The validation region renders ValidationFieldsFor(etype), which for
		// a flat-DC enactment is engage plus the synthetic DC field. Parse
		// exactly that set so the form, cost and instructions agree.
		if fields := a.Cfg.ValidationFieldsFor(etype); len(fields) > 0 {
			en.ValidationData = readFieldValues(a.Cfg.Config, fields, prefix+"v_", r)
		}
		out = append(out, en)
	}
	return out
}

// handleBuilderCost recomputes advisory cost from posted form values. It also
// pushes the refreshed character bar out of band, so the builder's top bar
// stays live (perk-points-used follows the edit) instead of freezing at the
// values the page opened with.

func (a *App) handleBuilderCost(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	ab := model.Perk{Type: r.FormValue("type"), Fields: map[string]any{}}
	if at, ok := a.Cfg.PerkType(ab.Type); ok {
		ab.Fields = readFieldValues(a.Cfg.Config, at.Fields, "atype_", r)
	}
	ab.Enactments = a.readEnactments(r)
	cost := engine.PerkCost(a.Cfg.Config, ab)

	// Budget for the over-budget hint; the character id is passed as a form
	// value so this conditionless partial can look it up.
	budget := 0
	var c model.Character
	if found, ok := a.Store.Get(r.FormValue("character_id")); ok {
		c = found
		budget = a.Cfg.PerkPointBudget(c.Level)
	}
	// Preview the perk as if the current form were saved, so the bar reflects
	// the edit in progress rather than the last autosave.
	preview := c
	if preview.ID != "" {
		norm := engine.NormalizePerk(a.Cfg.Config, ab)
		if ab.ID != "" {
			if idx := findPerk(&preview, ab.ID); idx >= 0 {
				preview.Perks[idx] = norm
			} else {
				preview.Perks = append(preview.Perks, norm)
			}
		} else {
			preview.Perks = append(preview.Perks, norm)
		}
		if lvl := r.FormValue("level"); lvl != "" {
			if n, err := strconv.Atoi(lvl); err == nil {
				preview.Level = a.Cfg.ClampLevel(n)
			}
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := perkPage{
		Cfg:           a.Cfg.Config,
		Character:     &preview,
		Cost:          cost,
		Budget:        budget,
		OverBudget:    budget > 0 && cost.Build > budget,
		charStats:     a.characterStats(&preview),
		ReadOnlyStats: false,
	}
	if err := a.Tmpl.ExecuteTemplate(w, "cost_cards", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Live top bar: push the recomputed stat cards alongside the cost badge.
	// Current-vital inputs are preserved client-side (see app.js), so this
	// never resets Current back to Max while typing.
	if preview.ID != "" {
		_, _ = w.Write([]byte(`<div id="stat-cards" hx-swap-oob="innerHTML">`))
		_ = a.Tmpl.ExecuteTemplate(w, "stat_cards", data)
		_, _ = w.Write([]byte(`</div>`))
	}
}

// handleBuilderInstructions regenerates the play-facing instruction text from
// the posted builder form. It mirrors handleBuilderCost: the form is parsed
// into a throwaway perk, the generator runs over it, and only the
// "instructions" partial is returned so htmx can swap that region.
func (a *App) handleBuilderInstructions(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	ab := a.buildPerkFromForm(r, r.FormValue("perk_id"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := perkPage{
		Cfg:          a.Cfg.Config,
		Instructions: engine.PerkInstructions(a.Cfg.Config, ab),
	}
	if err := a.Tmpl.ExecuteTemplate(w, "instructions", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func findPerk(c *model.Character, id string) int {

	for i, ab := range c.Perks {
		if ab.ID == id {
			return i
		}
	}
	return -1
}

func firstPerkTypeID(cfg *config.Config) string {
	if len(cfg.PerkTypes.Order) > 0 {
		return cfg.PerkTypes.Order[0]
	}
	return ""
}

func firstEnactmentID(cfg *config.Config) string {
	if len(cfg.Enactments.Order) > 0 {
		return cfg.Enactments.Order[0]
	}
	return ""
}

func firstInteractionID(cfg *config.Config) string {
	if len(cfg.Interactions.Order) > 0 {
		return cfg.Interactions.Order[0]
	}
	return ""
}
