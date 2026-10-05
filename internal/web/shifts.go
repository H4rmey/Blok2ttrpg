package web

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// shiftRow is one applied shift card as the character sheet renders it: the
// stored instance plus the display names its skill key resolves to, resolved
// once here rather than looked up repeatedly in the template.
type shiftRow struct {
	model.AppliedShift

	// Skill is the display name of the shifted skill and GroupLabel the label
	// of its skill group, so the card can say "Defense: Reflex" rather than a
	// bare key (the same skill name can exist in several groups).
	Skill      string
	GroupLabel string

	// Information is the skill's configured help text, the text behind the
	// card's hover "i". Empty when the ruleset describes none, so the
	// template guards the badge on it.
	Information string

	// Options are the magnitudes a shift card may hold, from the ruleset's
	// Enact Shift range.
	Options []int

	// Known reports whether the card's skill key still resolves against the
	// config. A ruleset may drop a skill, and a saved character must still
	// open; the row is rendered as an unresolved key so the user can delete it.
	Known bool
}

// shiftRows resolves a character's applied shift cards into renderable rows.
func (a *App) shiftRows(c *model.Character) []shiftRow {
	out := make([]shiftRow, 0, len(c.Shifts))
	for _, applied := range c.Shifts {
		row := shiftRow{AppliedShift: applied, Skill: applied.SkillKey}
		group, skill, ok := strings.Cut(applied.SkillKey, ".")
		if ok {
			for _, g := range a.Cfg.Skills.List() {
				if g.ID != group {
					continue
				}
				for _, s := range g.Skills {
					if s == skill {
						row.Known = true
						row.Skill = s
						row.GroupLabel = g.Label
						row.Information = a.Cfg.Skills.Info(group, s)
					}
				}
			}
		}
		row.Options = a.Cfg.Config.EnactShiftOptions()
		out = append(out, row)
	}
	return out
}

// shiftPickerPage is the data envelope for the "apply shift" picker.
type shiftPickerPage struct {
	CharacterID string
	Groups      []shiftPickerGroup

	// Cfg carries the ruleset so each card can show the skill's configured
	// information behind a hover "i", the way the sheet's own list does.
	Cfg *config.Config
}

// shiftPickerGroup is one labelled section of the picker, mirroring how the
// sheet itself groups skills. The range covers every configured group
// including vitals: Movement is exactly the kind of number a play shift moves.
type shiftPickerGroup struct {
	ID     string
	Label  string
	Skills []shiftChoice

	// First marks the group whose <details> starts expanded, so the modal opens
	// showing real choices rather than four collapsed headers.
	First bool
}

// shiftChoice is one selectable skill in the picker.
type shiftChoice struct {
	// Key is the stored "<group>.<skill>" key; Skill is the bare skill name.
	Key   string
	Skill string
}

// handleShiftLibrary renders the picker of skills a shift card may be applied
// to. Every configured skill is offered with no exceptions: unlike a condition,
// whose affected skills the ruleset fixed in its config, a shift card is the
// player pointing at one skill and moving it.
func (a *App) handleShiftLibrary(w http.ResponseWriter, r *http.Request) {
	data := shiftPickerPage{CharacterID: r.URL.Query().Get("character")}
	data.Cfg = a.Cfg.Config
	for i, g := range a.Cfg.Skills.List() {
		group := shiftPickerGroup{ID: g.ID, Label: g.Label, First: i == 0}
		for _, s := range g.Skills {
			group.Skills = append(group.Skills, shiftChoice{
				Key:   model.SkillKey(g.ID, s),
				Skill: s,
			})
		}
		data.Groups = append(data.Groups, group)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.Tmpl.ExecuteTemplate(w, "shift_library", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleCharacterShifts dispatches /characters/{id}/shifts[/...]. The routing
// and method/redirect conventions follow the conditions dispatcher, because the
// two flows are the same shape: a plain-form apply that reloads the sheet, and
// HTMX-driven value changes and removals that redirect via header.
func (a *App) handleCharacterShifts(w http.ResponseWriter, r *http.Request, c *model.Character, rest []string) {
	if len(rest) == 0 {
		http.NotFound(w, r)
		return
	}
	switch {
	case rest[0] == "apply":
		a.applyShift(w, r, c)
	case len(rest) >= 2 && rest[1] == "shift":
		a.setShift(w, r, c, rest[0])
	case r.Method == http.MethodDelete:
		a.removeShift(w, r, c, rest[0])
	default:
		http.NotFound(w, r)
	}
}

// applyShift adds one shift card. The same skill may be carded more than once,
// by design: each card is its own effect with its own remove button, and
// engine.SkillShifts sums them, which is also what lets a later card cancel an
// earlier one instead of fighting the removal of a single netting entry.
func (a *App) applyShift(w http.ResponseWriter, r *http.Request, c *model.Character) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	key := r.FormValue("skill_key")
	group, skill, ok := strings.Cut(key, ".")
	if !ok || !skillExists(a.Cfg.Skills, group, skill) {
		// Rejecting an unknown key server-side (not merely hiding it from the
		// picker) keeps a crafted request from adding a card that colours
		// nothing.
		http.Error(w, "unknown skill", http.StatusBadRequest)
		return
	}
	c.Shifts = append(c.Shifts, model.AppliedShift{
		SkillKey: key,
		Shift:    config.SmallestShift(a.Cfg.Config.EnactShiftOptions()),
	})
	if err := a.Store.Save(*c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// A real redirect, not an HX-Redirect header: the Apply button in the
	// picker is an ordinary form submit, so the browser navigates itself and
	// would render an empty page if we only set the HTMX header.
	http.Redirect(w, r, "/characters/"+c.ID+"?t=skills#tab-skills", http.StatusSeeOther)
}

// setShift changes the magnitude of one applied shift card. The index
// identifies which instance, because the same skill may carry several cards.
func (a *App) setShift(w http.ResponseWriter, r *http.Request, c *model.Character, index string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	i, err := strconv.Atoi(index)
	if err != nil || i < 0 || i >= len(c.Shifts) {
		http.NotFound(w, r)
		return
	}
	_ = r.ParseForm()
	n, err := strconv.Atoi(r.FormValue("shift"))
	if err != nil {
		http.Error(w, "shift must be a number", http.StatusBadRequest)
		return
	}
	// The posted magnitude is validated against the ruleset's Enact Shift range
	// rather than stored as sent: a value outside the range would apply a
	// shift the ruleset never allowed.
	allowed := false
	for _, v := range a.Cfg.Config.EnactShiftOptions() {
		if v == n {
			allowed = true
			break
		}
	}
	if !allowed {
		http.Error(w, "shift is outside the Enact Shift range", http.StatusBadRequest)
		return
	}
	c.Shifts[i].Shift = n
	if err := a.Store.Save(*c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Redirect", "/characters/"+c.ID+"?t=skills#tab-skills")
}

// removeShift drops one applied shift card by index. Nothing else has to be
// undone: the skill shifts were never stored, only derived, so deleting the
// entry is the whole of the removal.
func (a *App) removeShift(w http.ResponseWriter, r *http.Request, c *model.Character, index string) {
	i, err := strconv.Atoi(index)
	if err != nil || i < 0 || i >= len(c.Shifts) {
		http.NotFound(w, r)
		return
	}
	c.Shifts = append(c.Shifts[:i], c.Shifts[i+1:]...)
	if err := a.Store.Save(*c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Redirect", "/characters/"+c.ID+"?t=skills#tab-skills")
}

// skillExists reports whether the ruleset defines the given skill in the given
// group.
func skillExists(skills config.SkillMap, group, skill string) bool {
	for _, g := range skills.List() {
		if g.ID != group {
			continue
		}
		for _, s := range g.Skills {
			if s == skill {
				return true
			}
		}
	}
	return false
}
