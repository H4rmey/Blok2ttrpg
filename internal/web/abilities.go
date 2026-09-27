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

	// ReadOnlyStats is always true on the builder: the character bar's inputs
	// belong to the character sheet form, which does not exist here.
	ReadOnlyStats bool

	// Instructions is the generated play-facing rules text for the perk,
	// one entry per enactment. It is rendered by the "instructions" partial
	// and refreshed by /builder/instructions as the builder changes.
	Instructions []engine.Instruction
}

// handlePerks dispatches /characters/{id}/perks[/...] routes.
func (a *App) handlePerks(w http.ResponseWriter, r *http.Request, c *model.Character, rest []string) {
	// /perks        -> list
	if len(rest) == 0 {
		a.renderPerkList(w, c)
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
		a.renderPerkListPartial(w, c, changed)
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

	// /perks/new    -> builder for a new perk
	if rest[0] == "new" {

		if r.Method == http.MethodPost {
			a.savePerk(w, r, c, "")
			return
		}
		// The name is collected up front via a modal (mirroring the new
		// character flow) and passed as a query parameter so the builder opens
		// with the name already set and the rest of the form unlocked.
		blank := model.Perk{Type: firstPerkTypeID(a.Cfg.Config), Name: strings.TrimSpace(r.URL.Query().Get("name"))}
		a.renderBuilder(w, c, &blank, true)
		return

	}

	aid := rest[0]
	idx := findPerk(c, aid)
	if idx < 0 {
		http.NotFound(w, r)
		return
	}

	if len(rest) == 1 {
		switch r.Method {
		case http.MethodGet:
			a.renderBuilder(w, c, &c.Perks[idx], false)
		case http.MethodPost:
			a.savePerk(w, r, c, aid)
		case http.MethodDelete:
			c.Perks = append(c.Perks[:idx], c.Perks[idx+1:]...)
			_ = a.Store.Save(*c)
			w.Header().Set("HX-Redirect", "/characters/"+c.ID+"/perks")
		}
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
	http.Redirect(w, r, "/characters/"+c.ID+"/perks", http.StatusSeeOther)
}

// perkLibraryRow is one perk in the browser plus whether the character can still
// pay for it. With no character in context it is always affordable.
type perkLibraryRow struct {
	perkSummary
	Affordable bool
}

// perkLibraryPage is the data envelope for the built-in perk browser. Perks
// carry their computed cost so the library can show the price of a perk before
// it is imported, along with the character's remaining budgets so an
// unaffordable perk can be flagged and blocked.
type perkLibraryPage struct {
	CharacterID string
	Perks       []perkLibraryRow

	// HasCharacter reports whether budget figures are available. When false the
	// browser hides the remaining-points header and never blocks an import.
	HasCharacter bool
	charStats
}

// handlePerkLibrary renders the built-in perk browser. It expects a
// "character" query parameter so the import buttons post to the right route and
// so each perk can be checked against that character's remaining perk points.
func (a *App) handlePerkLibrary(w http.ResponseWriter, r *http.Request) {
	charID := r.URL.Query().Get("character")
	abs, err := a.Library.ListPerks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := perkLibraryPage{CharacterID: charID}
	if charID != "" {
		if c, ok := a.Store.Get(charID); ok {
			data.HasCharacter = true
			data.charStats = a.characterStats(&c)
		}
	}
	perkLeft := data.PerkBudget - data.PerkUsed

	// Each perk is normalized before its cost is computed, exactly as it will be
	// on import. That keeps the price shown in the library identical to the
	// price the perk ends up with once it is on a character.
	for _, ab := range abs {
		norm := engine.NormalizePerk(a.Cfg.Config, ab)
		cost := engine.PerkCost(a.Cfg.Config, norm)
		row := perkLibraryRow{
			perkSummary: perkSummary{Perk: norm, Cost: cost},
			Affordable:  true,
		}
		if data.HasCharacter {
			row.Affordable = cost.Build <= perkLeft
		}
		data.Perks = append(data.Perks, row)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.Tmpl.ExecuteTemplate(w, "perk_library", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// affordPerk reports whether the character has enough perk points left to
// take on the given (already normalized) perk, with a human-readable reason
// when it does not. Callers turn a false result into a 400 so an unaffordable
// import is blocked server-side and not merely flagged in the UI.
func (a *App) affordPerk(c *model.Character, ab model.Perk) (bool, string) {
	cost := engine.PerkCost(a.Cfg.Config, ab)
	stats := a.characterStats(c)
	if left := stats.PerkBudget - stats.PerkUsed; cost.Build > left {
		return false, fmt.Sprintf("%q costs %d perk points but only %d remain.", ab.Name, cost.Build, left)
	}
	return true, ""
}

// importBuiltinPerk copies a built-in perk (by library id) onto the
// character with a fresh id and redirects back to the perk list.
func (a *App) importBuiltinPerk(w http.ResponseWriter, r *http.Request, c *model.Character) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	id := r.FormValue("perk_id")
	if id == "" {
		http.Error(w, "missing perk id", http.StatusBadRequest)
		return
	}
	ab, err := a.Library.GetPerk(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
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
	http.Redirect(w, r, "/characters/"+c.ID+"/perks", http.StatusSeeOther)
}

// perkSummaries recomputes cost and instruction text for every perk the
// character owns. Costs are always derived here rather than stored, so the
// figures follow the current config even for perks built long ago.
func (a *App) perkSummaries(c *model.Character) []perkSummary {
	perks := make([]perkSummary, 0, len(c.Perks))
	for _, ab := range c.Perks {
		perks = append(perks, perkSummary{
			Perk:         ab,
			Cost:         engine.PerkCost(a.Cfg.Config, ab),
			Instructions: engine.PerkInstructions(a.Cfg.Config, ab),
		})
	}
	return perks
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
			{Label: "Perks", URL: "/characters/" + c.ID + "/perks"},
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
func (a *App) renderPerkListPartial(w http.ResponseWriter, c *model.Character, changed int) {

	data := a.perkListPage(c)
	data.Cfg = a.Cfg.Config
	data.RefreshNotice = refreshNotice(changed)
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

		charStats:     a.characterStats(c),
		ReadOnlyStats: true,

		Instructions: engine.PerkInstructions(a.Cfg.Config, *ab),

		Breadcrumbs: []crumb{
			{Label: "Home", URL: "/"},
			{Label: c.Name(), URL: "/characters/" + c.ID},
			{Label: "Perks", URL: "/characters/" + c.ID + "/perks"},
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
	http.Redirect(w, r, "/characters/"+c.ID+"/perks", http.StatusSeeOther)
}

// readFieldValues extracts values for a set of fields from a form using a key
// prefix. It handles each field type generically.
func readFieldValues(cfg *config.Config, fields []config.Field, prefix string, r *http.Request) map[string]any {
	out := map[string]any{}
	for _, f := range fields {
		name := prefix + f.Key
		switch f.Type {
		case "checkbox":
			out[f.Key] = r.FormValue(name) == "on" || r.FormValue(name) == "true"
		case "free_number":
			n, _ := strconv.Atoi(r.FormValue(name))
			out[f.Key] = n
		case "multiselect", "conditions":
			out[f.Key] = readRowValues(f, name, r)

		case "condition_select":
			out[f.Key] = r.FormValue(name)
			// The per-condition shift dropdown is injected by HTMX only for general
			// conditions, so it is not a standalone config field. Capture its posted
			// value under the sibling shift key so the cost engine can read it.
			shiftKey := f.ShiftKey
			if shiftKey == "" {
				shiftKey = "shift_amount"
			}
			if v := r.FormValue(prefix + shiftKey); v != "" {
				n, _ := strconv.Atoi(v)
				out[shiftKey] = n
			}
		case "dropdown":

			val := r.FormValue(name)
			// A dropdown must always carry a real value. If the post is missing
			// one (e.g. a region that was not rendered), fall back to the
			// configured default and then to the first option, matching what
			// normalization stores and what the builder renders.
			if val == "" {
				val = asStringValue(f.Default)
			}
			if val == "" && cfg != nil {
				for _, opt := range cfg.ResolveOptions(f) {
					if opt.Value != "" {
						val = opt.Value
						break
					}
				}
			}
			out[f.Key] = val
			// An inline_builder dropdown carries a nested component builder.
			// Parse the referenced component's fields under "<name>_ib_" and
			// store them under a nested key so the cost engine can recurse.
			if f.InlineBuilder != nil && val != "" && cfg != nil {
				if comp, ok := cfg.ComponentByKind(f.InlineBuilder.Kind, val); ok {
					out[f.Key+"_ib"] = readFieldValues(cfg, comp.Fields, name+"_ib_", r)
				}
			}
		default:
			out[f.Key] = r.FormValue(name)
		}

	}
	return out
}

// asStringValue renders a config default as a string ("" for nil).
func asStringValue(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// readRowValues reads a multiselect/conditions repeatable field from the form. Rows

// are posted with an index in the name, e.g. "<name>_0_type", "<name>_0_value".
// The posted "<name>_count" holds the number of rows.
func readRowValues(f config.Field, name string, r *http.Request) []map[string]any {
	count, _ := strconv.Atoi(r.FormValue(name + "_count"))
	rows := make([]map[string]any, 0, count)
	for i := 0; i < count; i++ {
		rowPrefix := fmt.Sprintf("%s_%d_", name, i)
		row := map[string]any{}
		empty := true
		for _, rf := range f.RowFields {
			v := r.FormValue(rowPrefix + rf.Key)
			row[rf.Key] = v
			if v != "" {
				empty = false
			}
		}
		if !empty {
			rows = append(rows, row)
		}
	}
	return rows
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
	fmt.Fprintf(w, `{"id":%q}`, ab.ID)
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
		if len(a.Cfg.Validations.Fields) > 0 {
			en.ValidationData = readFieldValues(a.Cfg.Config, a.Cfg.Validations.Fields, prefix+"v_", r)
		}
		out = append(out, en)
	}
	return out
}

// handleBuilderCost recomputes advisory cost from posted form values.

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
	if c, ok := a.Store.Get(r.FormValue("character_id")); ok {
		budget = a.Cfg.PerkPointBudget(c.Level)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := perkPage{
		Cost:       cost,
		Budget:     budget,
		OverBudget: budget > 0 && cost.Build > budget,
	}
	if err := a.Tmpl.ExecuteTemplate(w, "cost_cards", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
