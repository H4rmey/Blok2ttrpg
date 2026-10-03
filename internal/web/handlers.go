package web

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/engine"
	"github.com/harmey/blok2ttrpg-v5/internal/export"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// pageData is the common data envelope for full-page templates.
type pageData struct {
	Cfg         *config.Config
	Title       string
	Breadcrumbs []crumb
	Character   *model.Character
	Characters  []model.Character

	// charStats carries the budget and vital figures rendered in the character
	// bar. Fields are promoted, so templates keep using .PerkBudget etc.
	charStats

	// Perks carries the character's perks pre-enriched with their computed
	// cost and generated instruction text, so the Perks list can show the rules
	// inline without the user opening each perk in the builder.
	Perks []perkSummary

	// Warning is a non-blocking notice shown at the top of the page, e.g. when
	// a package shift was clamped at the ends of the proficiency ladder.
	Warning string

	// RefreshNotice reports the outcome of a perk recalculation, so the refresh
	// controls visibly confirm what they did.
	RefreshNotice string

	// Conditions are the conditions currently applied, resolved against the
	// config for display. SkillViews is the post-condition reading of every
	// skill, keyed as "<group>.<skill>", which is what lets the skill grid show
	// the shifted proficiency and colour it.
	Conditions []conditionRow
	SkillViews map[string]engine.SkillView

	// Shifts are the hand-applied "Enact Shift" cards, resolved against the
	// config for display. They feed the same SkillViews overlay as Conditions.
	Shifts []shiftRow

	// ReadOnlyStats renders the character bar without editable inputs. The
	// level box and the current-value boxes for vitals are bound to the
	// character form, which only exists on the sheet itself, so every other
	// character-scoped page shows the same numbers as plain text.
	ReadOnlyStats bool
}

// charStats holds the derived budget and vital figures shown in the character
// bar. It is embedded in every envelope that renders the bar so the perks list
// and the perk builder can show the same numbers as the character sheet.
type charStats struct {
	SkillBudget int
	SkillUsed   int
	PerkBudget  int
	PerkUsed    int
	Vitals      []engine.VitalStat

	// InvokeBudget is the maximum invoke points the character's level grants;
	// InvokeCurrent is how many are unspent right now. Unlike the two point
	// pools above, this is not a build budget: it is a live counter the player
	// edits during play, so it is stored and shown like a vital (current/max)
	// rather than computed from what has been bought.
	InvokeBudget  int
	InvokeCurrent int
}

// characterStats computes the character bar figures for a character.
func (a *App) characterStats(c *model.Character) charStats {
	perkUsed := 0
	for _, ab := range c.Perks {
		perkUsed += engine.PerkCost(a.Cfg.Config, ab).Build
	}
	invokeBudget := a.Cfg.InvokePointBudget(c.Level)
	return charStats{
		SkillBudget:   a.Cfg.SkillPointBudget(c.Level),
		SkillUsed:     engine.SkillPointsUsed(a.Cfg.Config, *c),
		PerkBudget:    a.Cfg.PerkPointBudget(c.Level),
		PerkUsed:      perkUsed,
		Vitals:        engine.CharacterVitals(a.Cfg.Config, *c),
		InvokeBudget:  invokeBudget,
		InvokeCurrent: a.invokeCurrent(c, invokeBudget),
	}
}

// invokeCurrentKey is the trait key the unspent invoke point count is stored
// under. It follows the "current_<thing>" convention the editable vitals use so
// it round-trips through the same character sheet form.
const invokeCurrentKey = "current_invoke"

// invokeCurrent reads the stored unspent invoke points, clamped to the range the
// rules allow. A character that has never been edited has no stored value, which
// is read as a full pool: a fresh character starts a session with every point
// available rather than with none.
func (a *App) invokeCurrent(c *model.Character, budget int) int {
	raw, ok := c.Traits[invokeCurrentKey]
	if !ok {
		return budget
	}
	n, ok := atoiAny(raw)
	if !ok {
		return budget
	}
	return a.clampInvoke(n, budget)
}

// clampInvoke constrains an invoke point count to the playable range. The upper
// bound is the level's budget unless the ruleset sets
// invoking.allow_over_maximum, in which case banked points may exceed it.
func (a *App) clampInvoke(n, budget int) int {
	if n < 0 {
		return 0
	}
	if !a.Cfg.Invoking.AllowsOverMaximum() && n > budget {
		return budget
	}
	return n
}

// atoiAny parses an int out of a stored trait value. Values arrive from the form
// as strings and from JSON as float64, so both are accepted.
func atoiAny(v any) (int, bool) {
	switch t := v.(type) {
	case int:
		return t, true
	case float64:
		return int(t), true
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(t))
		if err != nil {
			return 0, false
		}
		return n, true
	}
	return 0, false
}

type crumb struct {
	Label string
	URL   string
}

// perkSummary pairs an perk with the derived values the Perks list shows:
// its advisory cost and its generated play-facing instructions.
type perkSummary struct {
	Perk         model.Perk
	Cost         engine.Cost
	Instructions []engine.Instruction

	// Passive carries the catalogue view of this perk when it is a passive. A
	// passive is a perk, but it is not built out of enactments, interactions and
	// validations, so the list must render it differently: no enactment count,
	// no generated instructions, and no Edit or Export (there is nothing to edit
	// but its value, and nothing portable to export). It is nil for every
	// ordinary perk.
	Passive *passiveSummary
}

// IsPassive reports whether the perk is a predefined passive, so the template
// can branch without reaching into the pointer.
func (s perkSummary) IsPassive() bool { return s.Passive != nil }

// passiveSummary is the in-list view of a selected passive: which catalogue
// entry it is and its rules text at the configured values.
//
// The values themselves are not carried here. Changing them happens in the
// configure modal, which fetches the entry and the stored values itself, so the
// list only needs to render the text and decide whether to offer the button.
type passiveSummary struct {
	ID   string
	Name string

	// Segments is the rules text split so each configured value can be
	// highlighted where it sits in the sentence, which is what makes it obvious
	// which parts of the passive were chosen rather than fixed.
	Segments []config.DescriptionSegment

	// Configurable reports whether the entry has any fields, and therefore
	// whether a Configure button is worth showing. A fixed passive has nothing
	// to open.
	Configurable bool
}

func (a *App) render(w http.ResponseWriter, name string, data pageData) {
	data.Cfg = a.Cfg.Config
	if data.Title == "" {
		data.Title = a.Cfg.Title
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.Tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	a.render(w, "index.html", pageData{
		Title:       a.Cfg.Title,
		Characters:  a.Store.List(),
		Breadcrumbs: []crumb{{Label: "Home", URL: "/"}},
	})
}

func (a *App) handleNewCharacter(w http.ResponseWriter, r *http.Request) {
	c := a.blankCharacter("")
	a.render(w, "character.html", a.characterPage(&c, true))
}

func (a *App) handleCreateCharacter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := fmt.Sprintf("char-%d", time.Now().UnixNano())
	c := a.blankCharacter(id)
	a.applyCharacterForm(&c, r)
	// Ensure the name provided in the creation modal is always stored, even if
	// "name" is not a configured trait field.
	if name := r.FormValue("attr_name"); name != "" {
		c.Traits["name"] = name
	}
	if err := a.Store.Save(c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/characters/"+id, http.StatusSeeOther)
}

// handleCharacter dispatches all /characters/{id}[/action] routes.
func (a *App) handleCharacter(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/characters/")
	parts := strings.Split(path, "/")
	if parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	id := parts[0]
	c, ok := a.Store.Get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			data := a.characterPage(&c, false)
			if warn := r.URL.Query().Get("warn"); warn != "" {
				data.Warning = warn
			}
			a.render(w, "character.html", data)

		case http.MethodPost:
			warn := a.applyCharacterForm(&c, r)
			if err := a.Store.Save(c); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			target := "/characters/" + id
			if warn != "" {
				target += "?warn=" + url.QueryEscape(warn)
			}
			http.Redirect(w, r, target, http.StatusSeeOther)
		case http.MethodDelete:
			_ = a.Store.Delete(id)
			w.Header().Set("HX-Redirect", "/")
		}
		return
	}

	switch parts[1] {
	case "delete":
		_ = a.Store.Delete(id)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	case "export":
		b, err := export.MarshalCharacter(c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/x-yaml")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", c.Name()+".yaml"))
		w.Write(b)
	case "stats":
		// Recompute from the current (unsaved) form values so the bar reflects
		// edits to level and skill dropdowns before saving.
		a.applyCharacterForm(&c, r)
		if lvl := r.URL.Query().Get("level"); lvl != "" {
			if n, err := strconv.Atoi(lvl); err == nil {
				c.Level = a.Cfg.ClampLevel(n)
			}
		}
		data := a.characterPage(&c, false)
		data.Cfg = a.Cfg.Config
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := a.Tmpl.ExecuteTemplate(w, "stat_cards", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case "pdf":
		a.renderCharacterPDF(w, c)
	case "perks":
		a.handlePerks(w, r, &c, parts[2:])
	case "packages":
		a.handlePackages(w, r, &c, parts[2:])
	case "conditions":
		a.handleCharacterConditions(w, r, &c, parts[2:])
	case "shifts":
		a.handleCharacterShifts(w, r, &c, parts[2:])
	default:
		http.NotFound(w, r)
	}
}

func (a *App) handleImportCharacter(w http.ResponseWriter, r *http.Request) {
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
	c, err := export.UnmarshalCharacter(data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if c.ID == "" {
		c.ID = fmt.Sprintf("char-%d", time.Now().UnixNano())
	}
	// An imported file may name any level, so normalise it against the
	// ruleset's cap before the character is stored.
	c.Level = a.Cfg.ClampLevel(c.Level)
	if err := a.Store.Save(c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/characters/"+c.ID, http.StatusSeeOther)
}

// blankCharacter builds a character with defaults for every configured skill.
func (a *App) blankCharacter(id string) model.Character {
	c := model.Character{
		ID:     id,
		Level:  1,
		Traits: map[string]any{},
		Skills: map[string]string{},
		Perks:  []model.Perk{},
	}
	def := a.Cfg.DefaultProficiencyID()
	for _, g := range a.Cfg.Skills.List() {
		for _, t := range g.Skills {
			c.Skills[model.SkillKey(g.ID, t)] = def
		}
	}

	return c
}

// applyCharacterForm reads posted form fields into the generic character maps.
// It returns a warning message when the posted skill selections had to be
// rejected for overspending the skill-point budget; an empty string means the
// form was applied as posted.
func (a *App) applyCharacterForm(c *model.Character, r *http.Request) string {
	_ = r.ParseForm()
	// The level is clamped to the configured range, so a hand-edited or
	// scripted request cannot push a character past leveling.max_level.
	if lvl := r.FormValue("level"); lvl != "" {
		if n, err := strconv.Atoi(lvl); err == nil {
			c.Level = a.Cfg.ClampLevel(n)
		}
	}
	for _, g := range a.Cfg.Traits.List() {
		for _, f := range g.Fields {
			name := "attr_" + f.Key
			if _, ok := r.Form[name]; ok {
				c.Traits[f.Key] = r.FormValue(name)
			}
		}
	}
	// Snapshot the skill tiers so an overspending selection can be rolled back
	// wholesale. Skill points are only allowed to go negative when the ruleset
	// sets allow_negative_skill_points.
	before := make(map[string]string, len(c.Skills))
	for k, v := range c.Skills {
		before[k] = v
	}
	for _, g := range a.Cfg.Skills.List() {

		for _, t := range g.Skills {
			name := "skill_" + g.ID + "_" + t
			if v := r.FormValue(name); v != "" {
				c.Skills[model.SkillKey(g.ID, t)] = v
			}
		}
	}
	warning := ""
	if !a.Cfg.AllowsNegativeSkillPoints() {
		budget := a.Cfg.SkillPointBudget(c.Level)
		if used := engine.SkillPointsUsed(a.Cfg.Config, *c); used > budget {
			c.Skills = before
			warning = fmt.Sprintf("That selection would use %d skill points but only %d are available at level %d. Your skill changes were not applied.", used, budget, c.Level)
		}
	}
	// Current values for editable vitals (HP/Energy). Stored as traits
	// keyed "current_<vital>" so they persist alongside the character.
	for _, skill := range a.Cfg.Skills.Items[engine.VitalGroupID(a.Cfg.Config)] {
		key := strings.ToLower(skill)
		name := "current_" + key
		if _, ok := r.Form[name]; ok {
			c.Traits[name] = r.FormValue(name)
		}
	}
	// Unspent invoke points. This is clamped rather than stored verbatim,
	// because the input is a live play counter: a stale form or a hand-edited
	// request must not be able to bank more points than the ruleset allows.
	if _, ok := r.Form[invokeCurrentKey]; ok {
		if n, ok := atoiAny(r.FormValue(invokeCurrentKey)); ok {
			c.Traits[invokeCurrentKey] = strconv.Itoa(a.clampInvoke(n, a.Cfg.InvokePointBudget(c.Level)))
		}
	}
	return warning
}

func (a *App) characterPage(c *model.Character, isNew bool) pageData {
	title := c.Name()
	if isNew {
		title = "New Character"
	}
	return pageData{
		Title:     title,
		Character: c,
		Breadcrumbs: []crumb{
			{Label: "Home", URL: "/"},
			{Label: title, URL: "/characters/" + c.ID},
		},
		charStats:  a.characterStats(c),
		Conditions: a.conditionRows(c),
		Shifts:     a.shiftRows(c),
		SkillViews: engine.EffectiveSkills(a.Cfg.Config, *c),
		// A clamped condition shift is reported here rather than only at apply
		// time, because the clamp can start being true later: lowering a skill by
		// hand can push an already-applied condition off the end of the ladder.
		Warning: a.conditionWarning(c),
	}
}
