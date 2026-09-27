package web

import (
	"net/http"
	"strconv"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
)

// printData is the envelope for the blank, fill-in-by-hand print templates.
// They are driven entirely by the config, so no character is involved: the
// template ranges over the trait groups, skill groups and perk components and
// renders write-in lines and tick boxes for them.
type printData struct {
	Cfg   *config.Config
	Title string

	// Rows is how many blank repeat blocks to render: perk cards on the blank
	// character sheet, enactment blocks on the perk builder template. It is
	// taken from the "rows" query parameter so a user can print a sheet with
	// as much space as they need.
	Rows []int
}

// blankRows parses the requested number of blank repeat blocks from the query
// string, clamped to a sane printable range so a hand-edited request cannot
// ask the server to render thousands of pages.
func blankRows(r *http.Request, def int) []int {
	n := def
	if v := r.URL.Query().Get("rows"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			n = parsed
		}
	}
	if n < 1 {
		n = 1
	}
	if n > 20 {
		n = 20
	}
	out := make([]int, n)
	for i := range out {
		out[i] = i + 1
	}
	return out
}

// handleBlankCharacterSheet renders a printable, empty character sheet. Every
// skill row carries one tick box per proficiency tier, labelled with the die
// that tier grants for that skill group, so a player can fill the sheet in by
// hand without consulting the rulebook.
func (a *App) handleBlankCharacterSheet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := printData{
		Cfg:   a.Cfg.Config,
		Title: a.Cfg.Title + " - Blank Character Sheet",
		Rows:  blankRows(r, 6),
	}
	if err := a.Tmpl.ExecuteTemplate(w, "character_blank.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handlePerkTemplate renders a printable, empty perk-builder worksheet: the
// perk-type and enactment/interaction choices as tick boxes and their fields as
// write-in lines, mirroring the on-screen builder so a perk can be drafted in
// pencil and typed up later.
func (a *App) handlePerkTemplate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := printData{
		Cfg:   a.Cfg.Config,
		Title: a.Cfg.Title + " - Perk Builder Template",
		Rows:  blankRows(r, 3),
	}
	if err := a.Tmpl.ExecuteTemplate(w, "perk_template.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
