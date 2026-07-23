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

// abilityPage is the data envelope for the builder and ability views.
type abilityPage struct {
	Cfg         *config.Config
	Title       string
	Breadcrumbs []crumb
	Character   *model.Character
	Ability     *model.Ability
	Cost        engine.Cost
	Budget      int
	OverBudget  bool
}

// handleAbilities dispatches /characters/{id}/abilities[/...] routes.
func (a *App) handleAbilities(w http.ResponseWriter, r *http.Request, c *model.Character, rest []string) {
	// /abilities        -> list
	if len(rest) == 0 {
		a.renderAbilityList(w, c)
		return
	}

	// /abilities/import -> import an ability from an uploaded YAML file
	if rest[0] == "import" {
		a.importAbility(w, r, c)
		return
	}

	// /abilities/new    -> builder for a new ability
	if rest[0] == "new" {

		if r.Method == http.MethodPost {
			a.saveAbility(w, r, c, "")
			return
		}
		blank := model.Ability{Type: firstAbilityTypeID(a.Cfg.Config)}
		a.renderBuilder(w, c, &blank, true)
		return
	}

	aid := rest[0]
	idx := findAbility(c, aid)
	if idx < 0 {
		http.NotFound(w, r)
		return
	}

	if len(rest) == 1 {
		switch r.Method {
		case http.MethodGet:
			a.renderBuilder(w, c, &c.Abilities[idx], false)
		case http.MethodPost:
			a.saveAbility(w, r, c, aid)
		case http.MethodDelete:
			c.Abilities = append(c.Abilities[:idx], c.Abilities[idx+1:]...)
			_ = a.Store.Save(*c)
			w.Header().Set("HX-Redirect", "/characters/"+c.ID+"/abilities")
		}
		return
	}

	if rest[1] == "export" {
		b, err := export.MarshalAbility(c.Abilities[idx])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/x-yaml")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", c.Abilities[idx].Name+".yaml"))
		w.Write(b)
		return
	}
	http.NotFound(w, r)
}

// importAbility reads an uploaded YAML file, parses it into an ability, gives
// it a fresh id, appends it to the character and redirects back to the list.
func (a *App) importAbility(w http.ResponseWriter, r *http.Request, c *model.Character) {
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
	ab, err := export.UnmarshalAbility(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Always assign a fresh id so an imported ability never collides with an
	// existing one on the character.
	ab.ID = fmt.Sprintf("ability-%d", time.Now().UnixNano())
	c.Abilities = append(c.Abilities, ab)
	if err := a.Store.Save(*c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/characters/"+c.ID+"/abilities", http.StatusSeeOther)
}

func (a *App) renderAbilityList(w http.ResponseWriter, c *model.Character) {

	a.render(w, "abilities.html", pageData{
		Title:     c.Name() + " - Abilities",
		Character: c,
		Breadcrumbs: []crumb{
			{Label: "Home", URL: "/"},
			{Label: c.Name(), URL: "/characters/" + c.ID},
			{Label: "Abilities", URL: "/characters/" + c.ID + "/abilities"},
		},
	})
}

func (a *App) renderBuilder(w http.ResponseWriter, c *model.Character, ab *model.Ability, isNew bool) {
	cost := engine.AbilityCost(a.Cfg.Config, *ab)
	budget := a.Cfg.AbilityPointBudget(c.Level)
	title := "New Ability"
	if !isNew {
		title = ab.Name
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := a.Tmpl.ExecuteTemplate(w, "builder.html", abilityPage{
		Cfg:       a.Cfg.Config,
		Title:     title,
		Character: c,
		Ability:   ab,
		Cost:      cost,
		Budget:    budget,
		// Over budget is advisory only: it never blocks saving.
		OverBudget: cost.Build > budget,

		Breadcrumbs: []crumb{
			{Label: "Home", URL: "/"},
			{Label: c.Name(), URL: "/characters/" + c.ID},
			{Label: "Abilities", URL: "/characters/" + c.ID + "/abilities"},
			{Label: title, URL: "#"},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// saveAbility parses the builder form into an ability and stores it. Cost is
// never validated here; over-budget abilities are allowed by design.
func (a *App) saveAbility(w http.ResponseWriter, r *http.Request, c *model.Character, existingID string) {
	_ = r.ParseForm()
	ab := model.Ability{
		ID:          existingID,
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		Type:        r.FormValue("type"),
		Fields:      map[string]any{},
	}
	if ab.ID == "" {
		ab.ID = fmt.Sprintf("ability-%d", time.Now().UnixNano())
	}
	if at, ok := a.Cfg.AbilityType(ab.Type); ok {
		ab.Fields = readFieldValues(a.Cfg.Config, at.Fields, "atype_", r)
	}

	// Enactments are posted as enactment count + per-index type/fields.
	count, _ := strconv.Atoi(r.FormValue("enactment_count"))
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
		en.Interaction = r.FormValue(prefix + "interaction")
		if ic, ok := a.Cfg.Interaction(en.Interaction); ok {
			en.InteractionData = readFieldValues(a.Cfg.Config, ic.Fields, prefix+"i_", r)
		}
		if len(a.Cfg.Validations.Fields) > 0 {
			en.ValidationData = readFieldValues(a.Cfg.Config, a.Cfg.Validations.Fields, prefix+"v_", r)
		}
		ab.Enactments = append(ab.Enactments, en)
	}

	if idx := findAbility(c, existingID); idx >= 0 {
		c.Abilities[idx] = ab
	} else {
		c.Abilities = append(c.Abilities, ab)
	}
	if err := a.Store.Save(*c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/characters/"+c.ID+"/abilities", http.StatusSeeOther)
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
		case "solutions", "states":
			out[f.Key] = readRowValues(f, name, r)
		case "dropdown":
			val := r.FormValue(name)
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

// readRowValues reads a solutions/states repeatable field from the form. Rows
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
	data := map[string]any{
		"Cfg":             a.Cfg.Config,
		"Index":           idx,
		"Type":            etype,
		"Interaction":     interaction,
		"Fields":          map[string]any{},
		"InteractionData": map[string]any{},
		"ValidationData":  map[string]any{},
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
// the component generically by kind (enactment/interaction/ability_type) and
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

// handleAbilityTypeFields renders the config-driven fields for the selected
// ability type, used to swap the type-specific field block in the builder.

func (a *App) handleAbilityTypeFields(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	atype := r.FormValue("type")
	if atype == "" {
		atype = firstAbilityTypeID(a.Cfg.Config)
	}
	comp, ok := a.Cfg.AbilityType(atype)
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

// handleBuilderAutosave saves the ability from the posted builder form without
// redirecting, so the builder can persist edits in the background as the user
// selects options. It returns the (possibly newly created) ability id in the
// HX-Trigger-independent JSON body and an HX-Ability-ID header so the client
// can keep posting to a stable URL. Autosave is skipped until the ability has
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
	existingID := r.FormValue("ability_id")
	ab := a.buildAbilityFromForm(r, existingID)
	if idx := findAbility(&c, existingID); idx >= 0 {
		c.Abilities[idx] = ab
	} else {
		c.Abilities = append(c.Abilities, ab)
	}
	if err := a.Store.Save(c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("X-Ability-ID", ab.ID)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"id":%q}`, ab.ID)
}

// buildAbilityFromForm parses the builder form into an Ability. Shared by the
// manual save and the background autosave paths.
func (a *App) buildAbilityFromForm(r *http.Request, existingID string) model.Ability {
	ab := model.Ability{
		ID:          existingID,
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		Type:        r.FormValue("type"),
		Fields:      map[string]any{},
	}
	if ab.ID == "" {
		ab.ID = fmt.Sprintf("ability-%d", time.Now().UnixNano())
	}
	if at, ok := a.Cfg.AbilityType(ab.Type); ok {
		ab.Fields = readFieldValues(a.Cfg.Config, at.Fields, "atype_", r)
	}
	count, _ := strconv.Atoi(r.FormValue("enactment_count"))
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
		en.Interaction = r.FormValue(prefix + "interaction")
		if ic, ok := a.Cfg.Interaction(en.Interaction); ok {
			en.InteractionData = readFieldValues(a.Cfg.Config, ic.Fields, prefix+"i_", r)
		}
		if len(a.Cfg.Validations.Fields) > 0 {
			en.ValidationData = readFieldValues(a.Cfg.Config, a.Cfg.Validations.Fields, prefix+"v_", r)
		}
		ab.Enactments = append(ab.Enactments, en)
	}
	return ab
}

// handleBuilderCost recomputes advisory cost from posted form values.

func (a *App) handleBuilderCost(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	ab := model.Ability{Type: r.FormValue("type"), Fields: map[string]any{}}
	if at, ok := a.Cfg.AbilityType(ab.Type); ok {
		ab.Fields = readFieldValues(a.Cfg.Config, at.Fields, "atype_", r)
	}
	count, _ := strconv.Atoi(r.FormValue("enactment_count"))
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
		en.Interaction = r.FormValue(prefix + "interaction")
		if ic, ok := a.Cfg.Interaction(en.Interaction); ok {
			en.InteractionData = readFieldValues(a.Cfg.Config, ic.Fields, prefix+"i_", r)
		}
		if len(a.Cfg.Validations.Fields) > 0 {
			en.ValidationData = readFieldValues(a.Cfg.Config, a.Cfg.Validations.Fields, prefix+"v_", r)
		}
		ab.Enactments = append(ab.Enactments, en)
	}
	cost := engine.AbilityCost(a.Cfg.Config, ab)
	// Budget for the over-budget hint; the character id is passed as a form
	// value so this stateless partial can look it up.
	budget := 0
	if c, ok := a.Store.Get(r.FormValue("character_id")); ok {
		budget = a.Cfg.AbilityPointBudget(c.Level)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := abilityPage{
		Cost:       cost,
		Budget:     budget,
		OverBudget: budget > 0 && cost.Build > budget,
	}
	if err := a.Tmpl.ExecuteTemplate(w, "cost_cards", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func findAbility(c *model.Character, id string) int {
	for i, ab := range c.Abilities {
		if ab.ID == id {
			return i
		}
	}
	return -1
}

func firstAbilityTypeID(cfg *config.Config) string {
	if len(cfg.AbilityTypes.Order) > 0 {
		return cfg.AbilityTypes.Order[0]
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

var _ = strings.TrimSpace
