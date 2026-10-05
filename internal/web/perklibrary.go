// This file holds the perk library browser: the rows and tag groups the
// browse page renders, and the handlers that list them and import a chosen
// perk onto a character. It was split out of abilities.go, which had grown to
// carry the builder, the form decoding and this browser at once.

package web

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/harmey/blok2ttrpg-v5/internal/engine"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

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

	// Groups is the same perks bucketed by tag, for browsing a long list. A
	// perk with several tags appears in several groups, so this is a view over
	// Perks rather than a partition of it.
	Groups []perkTagGroup

	// HasCharacter reports whether budget figures are available. When false the
	// browser hides the remaining-points header and never blocks an import.
	HasCharacter bool
	charStats
}

// perkTagGroup is one tag heading and the perks carrying it.
type perkTagGroup struct {
	Tag   string
	Perks []perkLibraryRow
}

// untaggedGroupLabel is the bucket for perks with no tags at all. It is a
// display label, not a tag: no perk file should contain it.
const untaggedGroupLabel = "Untagged"

// groupPerksByTag buckets perks by their free-form tags, alphabetically by tag,
// with the untagged bucket last so a perk that was never tagged is still
// reachable instead of silently disappearing from the browser.
//
// There is no tag whitelist by design (see model.Perk.Tags): whatever tags the
// files carry become the headings. A misspelled tag therefore shows up as its
// own one-item group, which is the intended way to notice it.
func groupPerksByTag(rows []perkLibraryRow) []perkTagGroup {
	byTag := map[string][]perkLibraryRow{}
	var untagged []perkLibraryRow
	for _, r := range rows {
		if len(r.Perk.Tags) == 0 {
			untagged = append(untagged, r)
			continue
		}
		for _, t := range r.Perk.Tags {
			t = strings.TrimSpace(t)
			if t == "" {
				continue
			}
			byTag[t] = append(byTag[t], r)
		}
	}
	tags := make([]string, 0, len(byTag))
	for t := range byTag {
		tags = append(tags, t)
	}
	sort.Strings(tags)

	groups := make([]perkTagGroup, 0, len(tags)+1)
	for _, t := range tags {
		groups = append(groups, perkTagGroup{Tag: t, Perks: byTag[t]})
	}
	if len(untagged) > 0 {
		groups = append(groups, perkTagGroup{Tag: untaggedGroupLabel, Perks: untagged})
	}
	return groups
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
	data.Groups = groupPerksByTag(data.Perks)

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
	http.Redirect(w, r, "/characters/"+c.ID+"?t=perks#tab-perks", http.StatusSeeOther)
}
