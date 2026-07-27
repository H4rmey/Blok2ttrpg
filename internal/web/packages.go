package web

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/harmey/blok2ttrpg-v5/internal/model"
	"github.com/harmey/blok2ttrpg-v5/internal/premade"
)

// packageLibraryPage is the data envelope for the built-in package browser.
type packageLibraryPage struct {
	CharacterID string
	Packages    []premade.Package
}

// handlePackageLibrary renders the built-in package browser. It expects a
// "character" query parameter so the import buttons post to the right route.
func (a *App) handlePackageLibrary(w http.ResponseWriter, r *http.Request) {
	charID := r.URL.Query().Get("character")
	pkgs, err := a.Library.ListPackages()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := packageLibraryPage{CharacterID: charID, Packages: pkgs}
	if err := a.Tmpl.ExecuteTemplate(w, "package_library", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
		// /packages/{pkgId} with DELETE removes the package.
		if r.Method == http.MethodDelete {
			a.removePackage(w, r, c, rest[0])
			return
		}
		http.NotFound(w, r)
	}
}

// applyPackage applies a loaded package to a character: it shifts the relevant
// traits, copies the package's abilities in (each with a fresh id and a
// PackageID tag), and records an InstalledPackage so removal is exact.
//
// It returns the list of trait keys whose shift was clamped at an end of the
// proficiency ladder (i.e. the requested delta could not be fully applied).
// The caller uses this to surface a non-blocking warning; the stored delta is
// left as requested per the "just warn" policy.
func (a *App) applyPackage(c *model.Character, pkg *premade.Package) []string {
	if c.Traits == nil {
		c.Traits = map[string]string{}
	}
	// Only record shifts we actually applied so removal reverses exactly what
	// was done.
	applied := map[string]int{}
	var clamped []string
	for traitKey, delta := range pkg.Shifts {
		if delta == 0 {
			continue
		}
		current, ok := c.Traits[traitKey]
		if !ok {
			current = a.Cfg.DefaultProficiencyID()
		}
		if a.Cfg.ShiftClamped(current, delta) {
			clamped = append(clamped, traitKey)
		}
		c.Traits[traitKey] = a.Cfg.ShiftProficiency(current, delta)
		applied[traitKey] = delta
	}
	for _, ab := range pkg.Abilities {
		ab.ID = fmt.Sprintf("ability-%d", time.Now().UnixNano())
		ab.PackageID = pkg.ID
		c.Abilities = append(c.Abilities, ab)
		// Ensure unique ids even when copying several abilities in the same
		// nanosecond.
		time.Sleep(time.Nanosecond)
	}
	c.Packages = append(c.Packages, model.InstalledPackage{
		ID:     pkg.ID,
		Name:   pkg.Name,
		Shifts: applied,
	})
	return clamped
}

// packageRedirect sends the user back to the character sheet after an import,
// attaching a non-blocking warning about any clamped trait shifts.
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
	// A custom package's imports resolve against the built-in abilities
	// directory, so short names keep working for uploaded packages.
	baseDir := a.Library.CustomBaseDir()
	pkg, err := premade.ParsePackage(data, baseDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	clamped := a.applyPackage(c, pkg)
	if err := a.Store.Save(*c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	packageRedirect(w, r, c.ID, clamped)
}

// removePackage undoes an installed package: it removes every ability tagged
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
	// the same trait.
	for traitKey, delta := range rec.Shifts {
		current, ok := c.Traits[traitKey]
		if !ok {
			current = a.Cfg.DefaultProficiencyID()
		}
		c.Traits[traitKey] = a.Cfg.ShiftProficiency(current, -delta)
	}

	// Remove abilities tagged with this package id. User-created abilities and
	// abilities from other packages are left untouched.
	kept := c.Abilities[:0]
	for _, ab := range c.Abilities {
		if ab.PackageID == pkgID {
			continue
		}
		kept = append(kept, ab)
	}
	c.Abilities = kept

	// Drop the installed-package record.
	c.Packages = append(c.Packages[:idx], c.Packages[idx+1:]...)

	if err := a.Store.Save(*c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Redirect", "/characters/"+c.ID)
}
