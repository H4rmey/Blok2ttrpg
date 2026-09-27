package web

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/harmey/blok2ttrpg-v5/internal/engine"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
	"github.com/harmey/blok2ttrpg-v5/internal/premade"
)

// packageRow is one package in the browser together with what importing it
// would cost the current character and whether that fits in the remaining
// budget. Affordperk is only meaningful when a character is in context; with
// no character every row is reported as affordable.
type packageRow struct {
	premade.Package
	Cost engine.PackageCost

	// PerkAffordable/SkillAffordable are tracked separately so the template can
	// redden only the budget that actually overflows.
	PerkAffordable  bool
	SkillAffordable bool
	Affordable      bool

	// Clamped lists the skill keys whose shift would run off the top or bottom
	// of the proficiency ladder for this character, so the browser can warn
	// about them before the import happens. It is empty when every shift fits.
	Clamped []string
}

// packageLibraryPage is the data envelope for the built-in package browser.
type packageLibraryPage struct {
	CharacterID string
	Packages    []packageRow

	// HasCharacter reports whether budget figures are available. When false the
	// browser hides the remaining-points header and never blocks an import.
	HasCharacter bool
	charStats
}

// handlePackageLibrary renders the built-in package browser. It expects a
// "character" query parameter so the import buttons post to the right route and
// so each package can be priced against that character's remaining points.
func (a *App) handlePackageLibrary(w http.ResponseWriter, r *http.Request) {
	charID := r.URL.Query().Get("character")
	pkgs, err := a.Library.ListPackages()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := packageLibraryPage{CharacterID: charID}
	var c model.Character
	if charID != "" {
		if found, ok := a.Store.Get(charID); ok {
			c = found
			data.HasCharacter = true
			data.charStats = a.characterStats(&c)
		}
	}
	perkLeft := data.PerkBudget - data.PerkUsed
	skillLeft := data.SkillBudget - data.SkillUsed

	for _, pkg := range pkgs {
		cost := engine.PackageCostFor(a.Cfg.Config, c, pkg.Shifts, pkg.Perks)
		row := packageRow{Package: pkg, Cost: cost, PerkAffordable: true, SkillAffordable: true}
		if data.HasCharacter {
			row.PerkAffordable = cost.Perk <= perkLeft
			// Overspending skill points is only allowed when the ruleset opts in
			// via allow_negative_skill_points.
			row.SkillAffordable = a.Cfg.AllowsNegativeSkillPoints() || cost.Skill <= skillLeft
		}
		row.Affordable = row.PerkAffordable && row.SkillAffordable
		if data.HasCharacter {
			row.Clamped = a.clampedShifts(&c, pkg.Shifts)
		}
		data.Packages = append(data.Packages, row)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.Tmpl.ExecuteTemplate(w, "package_library", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// clampedShifts returns the skill keys whose shift cannot be applied in full
// because the character already sits at an end of the proficiency ladder. It is
// a preview only: it does not modify the character. Keys are returned sorted so
// the warning text is stable between renders.
func (a *App) clampedShifts(c *model.Character, shifts map[string]int) []string {
	var out []string
	for skillKey, delta := range shifts {
		if delta == 0 {
			continue
		}
		current, ok := c.Skills[skillKey]
		if !ok || current == "" {
			current = a.Cfg.DefaultProficiencyID()
		}
		if a.Cfg.ShiftClamped(current, delta) {
			out = append(out, skillKey)
		}
	}
	sort.Strings(out)
	return out
}

// affordPackage reports whether the character can pay for a package, returning a
// human-readable reason when it cannot. Callers turn a false result into a 400
// so an unaffordable import is blocked server-side and not merely hidden in the
// UI.
func (a *App) affordPackage(c *model.Character, pkg *premade.Package) (bool, string) {
	cost := engine.PackageCostFor(a.Cfg.Config, *c, pkg.Shifts, pkg.Perks)
	stats := a.characterStats(c)
	if perkLeft := stats.PerkBudget - stats.PerkUsed; cost.Perk > perkLeft {
		return false, fmt.Sprintf("%q costs %d perk points but only %d remain.", pkg.Name, cost.Perk, perkLeft)
	}
	if skillLeft := stats.SkillBudget - stats.SkillUsed; !a.Cfg.AllowsNegativeSkillPoints() && cost.Skill > skillLeft {
		return false, fmt.Sprintf("%q costs %d skill points but only %d remain.", pkg.Name, cost.Skill, skillLeft)
	}
	return true, ""
}

// handlePackages dispatches /characters/{id}/packages[/...] routes.
func (a *App) handlePackages(w http.ResponseWriter, r *http.Request, c *model.Character, rest []string) {
	if len(rest) == 0 {
		http.NotFound(w, r)
		return
	}

	switch rest[0] {
	case "import":
		// Import a built-in package by id.
		a.importBuiltinPackage(w, r, c)
	case "import-custom":
		// Import a package uploaded from disk.
		a.importCustomPackage(w, r, c)
	default:
		// /packages/{pkgId}/toggle enables or disables a toggleable package.
		if len(rest) >= 2 && rest[1] == "toggle" {
			a.togglePackage(w, r, c, rest[0])
			return
		}
		// /packages/{pkgId} with DELETE removes the package.
		if r.Method == http.MethodDelete {
			a.removePackage(w, r, c, rest[0])
			return
		}
		http.NotFound(w, r)
	}
}

// applyPackage applies a loaded package to a character: it shifts the relevant
// skills, copies the package's perks in (each with a fresh id and a
// PackageID tag), and records an InstalledPackage so removal is exact.
//
// It returns the list of skill keys whose shift was clamped at an end of the
// proficiency ladder (i.e. the requested delta could not be fully applied).
// The caller uses this to surface a non-blocking warning; the stored delta is
// left as requested per the "just warn" policy.
func (a *App) applyPackage(c *model.Character, pkg *premade.Package) []string {
	if c.Skills == nil {
		c.Skills = map[string]string{}
	}
	// Only record shifts we actually applied so removal reverses exactly what
	// was done.
	applied := map[string]int{}
	var clamped []string
	for skillKey, delta := range pkg.Shifts {
		if delta == 0 {
			continue
		}
		current, ok := c.Skills[skillKey]
		if !ok {
			current = a.Cfg.DefaultProficiencyID()
		}
		if a.Cfg.ShiftClamped(current, delta) {
			clamped = append(clamped, skillKey)
		}
		c.Skills[skillKey] = a.Cfg.ShiftProficiency(current, delta)
		applied[skillKey] = delta
	}
	for _, ab := range pkg.Perks {
		ab.ID = fmt.Sprintf("perk-%d", time.Now().UnixNano())
		ab.PackageID = pkg.ID
		c.Perks = append(c.Perks, ab)
		// Ensure unique ids even when copying several perks in the same
		// nanosecond.
		time.Sleep(time.Nanosecond)
	}
	c.Packages = append(c.Packages, model.InstalledPackage{
		ID:         pkg.ID,
		Name:       pkg.Name,
		Shifts:     applied,
		Toggleable: premade.Toggleable(pkg.Category),
		Enabled:    true,
	})
	return clamped
}

// applyPackageEffects re-applies a package's proficiency shifts and perks
// without appending a new InstalledPackage record. It is used when re-enabling
// a package that was previously disabled; the caller owns the existing record
// and updates its recorded shifts from the returned map.
func (a *App) applyPackageEffects(c *model.Character, pkg *premade.Package) map[string]int {
	if c.Skills == nil {
		c.Skills = map[string]string{}
	}
	applied := map[string]int{}
	for skillKey, delta := range pkg.Shifts {
		if delta == 0 {
			continue
		}
		current, ok := c.Skills[skillKey]
		if !ok {
			current = a.Cfg.DefaultProficiencyID()
		}
		c.Skills[skillKey] = a.Cfg.ShiftProficiency(current, delta)
		applied[skillKey] = delta
	}
	for _, ab := range pkg.Perks {
		ab.ID = fmt.Sprintf("perk-%d", time.Now().UnixNano())
		ab.PackageID = pkg.ID
		c.Perks = append(c.Perks, ab)
		time.Sleep(time.Nanosecond)
	}
	return applied
}

// reversePackageEffects reverses the recorded proficiency shifts and removes
// all perks tagged with the package id, leaving the InstalledPackage record
// in place. It is used when disabling a toggleable package.
func (a *App) reversePackageEffects(c *model.Character, rec *model.InstalledPackage) {
	for skillKey, delta := range rec.Shifts {
		current, ok := c.Skills[skillKey]
		if !ok {
			current = a.Cfg.DefaultProficiencyID()
		}
		c.Skills[skillKey] = a.Cfg.ShiftProficiency(current, -delta)
	}
	kept := c.Perks[:0]
	for _, ab := range c.Perks {
		if ab.PackageID == rec.ID {
			continue
		}
		kept = append(kept, ab)
	}
	c.Perks = kept
}

// togglePackage enables or disables a toggleable package. Disabling reverses
// the package's effects (shifts + perks) but keeps the record so it can be
// re-enabled; enabling re-applies the effects from the library definition.
// Class, race, and background packages are not toggleable and are rejected.
func (a *App) togglePackage(w http.ResponseWriter, r *http.Request, c *model.Character, pkgID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var rec *model.InstalledPackage
	for i := range c.Packages {
		if c.Packages[i].ID == pkgID {
			rec = &c.Packages[i]
			break
		}
	}
	if rec == nil {
		http.NotFound(w, r)
		return
	}
	if !rec.Toggleable {
		http.Error(w, "package is not toggleable", http.StatusBadRequest)
		return
	}

	if rec.Enabled {
		// Disable: reverse effects but keep the record.
		a.reversePackageEffects(c, rec)
		rec.Enabled = false
	} else {
		// Enable: re-apply effects from the library definition.
		pkg, err := a.Library.GetPackage(pkgID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Re-enabling spends points again, so it is subject to the same budget
		// check as a fresh import.
		if ok, reason := a.affordPackage(c, pkg); !ok {
			http.Error(w, reason, http.StatusBadRequest)
			return
		}
		rec.Shifts = a.applyPackageEffects(c, pkg)
		rec.Enabled = true
	}

	if err := a.Store.Save(*c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Redirect", "/characters/"+c.ID)
}

// packageRedirect sends the user back to the character sheet after an import,
// attaching a non-blocking warning about any clamped skill shifts.
func packageRedirect(w http.ResponseWriter, r *http.Request, charID string, clamped []string) {
	target := "/characters/" + charID
	if len(clamped) > 0 {
		msg := fmt.Sprintf("Some proficiencies hit the top or bottom of the ladder and could not shift the full amount: %s. Please verify your imported packages and remove the troublemaker(s).", strings.Join(clamped, ", "))
		target += "?warn=" + url.QueryEscape(msg)
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

func (a *App) importBuiltinPackage(w http.ResponseWriter, r *http.Request, c *model.Character) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	id := r.FormValue("package_id")
	if id == "" {
		http.Error(w, "missing package id", http.StatusBadRequest)
		return
	}
	pkg, err := a.Library.GetPackage(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if ok, reason := a.affordPackage(c, pkg); !ok {
		http.Error(w, reason, http.StatusBadRequest)
		return
	}
	clamped := a.applyPackage(c, pkg)
	if err := a.Store.Save(*c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	packageRedirect(w, r, c.ID, clamped)
}

func (a *App) importCustomPackage(w http.ResponseWriter, r *http.Request, c *model.Character) {
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
	// A custom package's imports resolve against the built-in perks
	// directory, so short names keep working for uploaded packages.
	baseDir := a.Library.CustomBaseDir()
	pkg, err := premade.ParsePackage(data, baseDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if ok, reason := a.affordPackage(c, pkg); !ok {
		http.Error(w, reason, http.StatusBadRequest)
		return
	}
	clamped := a.applyPackage(c, pkg)
	if err := a.Store.Save(*c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	packageRedirect(w, r, c.ID, clamped)
}

// removePackage undoes an installed package: it removes every perk tagged
// with the package id and reverses the exact proficiency shifts the package
// applied. Only content originating from this package is touched.
func (a *App) removePackage(w http.ResponseWriter, r *http.Request, c *model.Character, pkgID string) {
	// Find the installed record so we know which shifts to reverse.
	var rec *model.InstalledPackage
	idx := -1
	for i := range c.Packages {
		if c.Packages[i].ID == pkgID {
			rec = &c.Packages[i]
			idx = i
			break
		}
	}
	if rec == nil {
		http.NotFound(w, r)
		return
	}

	// Reverse the recorded shifts. Because we stored the exact deltas applied,
	// subtracting them is safe even when other installed packages also shifted
	// the same skill.
	for skillKey, delta := range rec.Shifts {
		current, ok := c.Skills[skillKey]
		if !ok {
			current = a.Cfg.DefaultProficiencyID()
		}
		c.Skills[skillKey] = a.Cfg.ShiftProficiency(current, -delta)
	}

	// Remove perks tagged with this package id. User-created perks and
	// perks from other packages are left untouched.
	kept := c.Perks[:0]
	for _, ab := range c.Perks {
		if ab.PackageID == pkgID {
			continue
		}
		kept = append(kept, ab)
	}
	c.Perks = kept

	// Drop the installed-package record.
	c.Packages = append(c.Packages[:idx], c.Packages[idx+1:]...)

	if err := a.Store.Save(*c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Redirect", "/characters/"+c.ID)
}
