package web

import (
	"net/http"
	"slices"
	"strconv"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/engine"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// conditionRow is one applied condition as the character sheet renders it: the
// stored instance plus everything the row needs from the config, resolved once
// here rather than looked up repeatedly in the template.
type conditionRow struct {
	model.AppliedCondition

	// Name and Description come from the config. Description is the text behind
	// the row's hover "i".
	Name        string
	Description string

	// ShiftOptions are the magnitudes this condition may be applied at. It is
	// empty for a condition that does not ask for one, which is what makes the
	// value dropdown disappear rather than render as a single dead choice.
	ShiftOptions []int

	// AffectedSkills lists the skills this condition moves, for the hover text.
	// A condition that moves none shows the description alone.
	AffectedSkills []string

	// EffectiveShift is the magnitude actually applied, which for a fixed
	// condition is its configured value rather than the stored Shift.
	EffectiveShift int

	// Known reports whether the condition id still resolves against the config.
	// A ruleset may drop a condition, and a saved character must still open; the
	// row is rendered as an unresolved id so the user can delete it.
	Known bool
}

// NeedsShiftValue reports whether this row shows a value dropdown.
func (r conditionRow) NeedsShiftValue() bool { return len(r.ShiftOptions) > 0 }

// conditionRows resolves a character's applied conditions into renderable rows.
func (a *App) conditionRows(c *model.Character) []conditionRow {
	out := make([]conditionRow, 0, len(c.Conditions))
	for _, applied := range c.Conditions {
		row := conditionRow{AppliedCondition: applied, Name: applied.ID}
		cond, ok := a.Cfg.ConditionByID(applied.ID)
		if ok {
			row.Known = true
			row.Name = cond.Name
			row.Description = cond.Description
			row.AffectedSkills = cond.AffectsSkills
			row.EffectiveShift = cond.SkillShift(applied.Shift)
			if cond.NeedsShiftValue() {
				row.ShiftOptions = a.Cfg.ShiftOptionsFor(cond.ID)
			}
		}
		out = append(out, row)
	}
	return out
}

// conditionPickerPage is the data envelope for the "apply condition" picker.
type conditionPickerPage struct {
	CharacterID string
	Conditions  []conditionChoice
}

// conditionChoice is one selectable condition in the picker, carrying the detail
// a player needs to choose without opening the rulebook: what it does and which
// skills it will move.
type conditionChoice struct {
	ID             string
	Name           string
	Description    string
	AffectedSkills []string

	// ShiftsSkills separates the two kinds of condition in the picker. One moves
	// numbers; the other changes what you may do. They read very differently at
	// the table, so the picker groups them rather than mixing them into one list.
	ShiftsSkills bool
}

// handleConditionLibrary renders the picker of conditions that may be applied to
// a character. Only selectable conditions are offered: a state the rules impose
// (Dying) or a property of gear is not something a user applies by hand here.
func (a *App) handleConditionLibrary(w http.ResponseWriter, r *http.Request) {
	data := conditionPickerPage{CharacterID: r.URL.Query().Get("character")}
	for _, cond := range a.Cfg.Conditions {
		if !cond.IsSelectable() {
			continue
		}
		data.Conditions = append(data.Conditions, conditionChoice{
			ID:             cond.ID,
			Name:           cond.Name,
			Description:    cond.Description,
			AffectedSkills: cond.AffectsSkills,
			ShiftsSkills:   cond.ShiftsSkills(),
		})
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.Tmpl.ExecuteTemplate(w, "condition_library", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleCharacterConditions dispatches /characters/{id}/conditions[/...].
func (a *App) handleCharacterConditions(w http.ResponseWriter, r *http.Request, c *model.Character, rest []string) {
	if len(rest) == 0 {
		http.NotFound(w, r)
		return
	}
	switch {
	case rest[0] == "apply":
		a.applyCondition(w, r, c)
	case len(rest) >= 2 && rest[1] == "shift":
		a.setConditionShift(w, r, c, rest[0])
	case r.Method == http.MethodDelete:
		a.removeCondition(w, r, c, rest[0])
	default:
		http.NotFound(w, r)
	}
}

// applyCondition adds a condition to a character.
//
// The same condition may be applied more than once. That is intentional: two
// separate sources of Frightened are two effects with their own durations and
// their own solutions, and collapsing them into one row would lose the ability
// to remove just one of them. Their shifts stack, which is the documented rule.
func (a *App) applyCondition(w http.ResponseWriter, r *http.Request, c *model.Character) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	id := r.FormValue("condition_id")
	cond, ok := a.Cfg.ConditionByID(id)
	if !ok {
		http.Error(w, "unknown condition", http.StatusBadRequest)
		return
	}
	// Applying a non-selectable condition by hand is rejected server-side, not
	// merely hidden from the picker, so a crafted request cannot grant
	// Invincible.
	if !cond.IsSelectable() {
		http.Error(w, "that condition cannot be applied directly", http.StatusBadRequest)
		return
	}
	applied := model.AppliedCondition{ID: cond.ID}
	if cond.NeedsShiftValue() {
		applied.Shift = defaultShiftFor(a.Cfg.Config, cond)
	}
	c.Conditions = append(c.Conditions, applied)
	if err := a.Store.Save(*c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// A real redirect, not an HX-Redirect header: the Apply button in the picker
	// is an ordinary form submit (like the package browser's Import), so the
	// browser navigates itself and would render an empty page if we only set the
	// HTMX header. The remove and shift controls below are HTMX-driven and do use
	// the header.
	http.Redirect(w, r, "/characters/"+c.ID+"?t=skills#tab-skills", http.StatusSeeOther)
}

// defaultShiftFor picks the magnitude a freshly applied condition starts at: the
// smallest non-zero step in its range. Starting at the weakest end means
// applying a condition never silently imposes the maximum penalty; the player
// has to choose that.
func defaultShiftFor(cfg *config.Config, cond config.Condition) int {
	return config.SmallestShift(cfg.ShiftOptionsFor(cond.ID))
}

// setConditionShift changes the magnitude of one applied condition. The index
// identifies which instance, because the same condition may appear more than
// once and the id alone would be ambiguous.
func (a *App) setConditionShift(w http.ResponseWriter, r *http.Request, c *model.Character, index string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	i, err := strconv.Atoi(index)
	if err != nil || i < 0 || i >= len(c.Conditions) {
		http.NotFound(w, r)
		return
	}
	_ = r.ParseForm()
	n, err := strconv.Atoi(r.FormValue("shift"))
	if err != nil {
		http.Error(w, "shift must be a number", http.StatusBadRequest)
		return
	}
	// The posted magnitude is validated against the condition's own range rather
	// than stored as sent: the dropdown is the only intended source, and a value
	// outside the range would apply a penalty the ruleset never allowed.
	cond, ok := a.Cfg.ConditionByID(c.Conditions[i].ID)
	if !ok || !cond.NeedsShiftValue() {
		http.Error(w, "that condition has no shift value", http.StatusBadRequest)
		return
	}
	allowed := false
	for _, v := range a.Cfg.ShiftOptionsFor(cond.ID) {
		if v == n {
			allowed = true
			break
		}
	}
	if !allowed {
		http.Error(w, "shift is outside this condition's range", http.StatusBadRequest)
		return
	}
	c.Conditions[i].Shift = n
	if err := a.Store.Save(*c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.renderSkillsTab(w, c)
}

// removeCondition drops one applied condition by index. Nothing else has to be
// undone: the skill shifts were never stored, only derived, so deleting the
// entry is the whole of the removal.
func (a *App) removeCondition(w http.ResponseWriter, r *http.Request, c *model.Character, index string) {
	i, err := strconv.Atoi(index)
	if err != nil || i < 0 || i >= len(c.Conditions) {
		http.NotFound(w, r)
		return
	}
	c.Conditions = append(c.Conditions[:i], c.Conditions[i+1:]...)
	if err := a.Store.Save(*c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.renderSkillsTab(w, c)
}

// conditionWarning reports the skills whose shifts (from applied conditions AND
// hand-applied shift cards) could not be applied in full because the skill
// already sat at an end of the proficiency ladder. It is surfaced as the
// sheet's non-blocking notice so a debuff that appears to do nothing is
// visibly explained rather than argued about.
func (a *App) conditionWarning(c *model.Character) string {
	views := engine.EffectiveSkills(a.Cfg.Config, *c)
	var clamped []string
	for key, v := range views {
		if v.Clamped {
			clamped = append(clamped, key)
		}
	}
	if len(clamped) == 0 {
		return ""
	}
	slices.Sort(clamped)
	return "These skills are already at the end of the proficiency ladder, so the " +
		"full condition or shift could not be applied: " + joinComma(clamped) + "."
}
