// Package premade loads the built-in content library that ships with the app:
// importable packages and the perks they reference. A package is a
// collection of proficiency shifts plus a list of perk imports; importing a
// package applies its shifts and copies its perks onto a character.
package premade

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/harmey/blok2ttrpg-v5/internal/export"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
	"gopkg.in/yaml.v3"
)

// PackageYAML is the on-disk shape of a package definition. Shifts map a skill
// key ("group.Skill") to a relative proficiency delta (e.g. +2, -1). Imports
// lists perks to include; each entry is either a short name (resolved to
// ../../perks/<name>/<name>.yaml relative to the package file) or an
// explicit path relative to the package file.
type PackageYAML struct {
	ID          string         `yaml:"id"`
	Name        string         `yaml:"name"`
	Description string         `yaml:"description,omitempty"`
	Shifts      map[string]int `yaml:"shifts,omitempty"`
	Imports     []string       `yaml:"imports,omitempty"`
	// Tags are free-form labels, same contract as model.Perk.Tags: no
	// predefined vocabulary, nothing validates them.
	Tags []string `yaml:"tags,omitempty"`
}

// Package is a loaded package: its metadata, the proficiency shifts it applies,
// and the fully-parsed perks it provides. Category is the library
// subfolder it lives in (e.g. "classes", "races", "backgrounds").
type Package struct {
	ID          string
	Name        string
	Category    string
	Description string
	Shifts      map[string]int
	Perks       []model.Perk
	Tags        []string
}

// Library is the built-in content library rooted at a directory. It exposes the
// packages available for import.
type Library struct {
	Root string
}

// New returns a Library rooted at the given directory (e.g. "library").
func New(root string) *Library {
	return &Library{Root: root}
}

// permanentCategories are the package categories that represent a character's
// core identity. Packages in these categories are always applied and cannot be
// toggled off; every other category (items, etc.) is toggleable.
var permanentCategories = map[string]bool{
	"classes":     true,
	"races":       true,
	"backgrounds": true,
}

// Toggleable reports whether a package in the given category can be enabled and
// disabled by the user. Identity packages (class/race/background) are permanent
// and therefore not toggleable.
func Toggleable(category string) bool {
	return !permanentCategories[category]
}

// packagesDir is the directory holding package definitions.
func (l *Library) packagesDir() string {
	return filepath.Join(l.Root, "packages")
}

// CustomBaseDir returns the perks directory that short-name imports in an
// uploaded (custom) package resolve against.
func (l *Library) CustomBaseDir() string {
	return l.perksDir()
}

// ListPackages scans the packages directory (recursively through category
// subfolders like classes/races/backgrounds) and returns every loadable
// package, sorted by category then name. Packages that fail to parse are
// skipped so one bad file does not break the whole browser.
func (l *Library) ListPackages() ([]Package, error) {
	dir := l.packagesDir()
	cats, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading packages dir: %w", err)
	}
	var out []Package
	for _, cat := range cats {
		if !cat.IsDir() {
			continue
		}
		catDir := filepath.Join(dir, cat.Name())
		entries, err := os.ReadDir(catDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := e.Name()
			path := filepath.Join(catDir, name, name+".yaml")
			if _, err := os.Stat(path); err != nil {
				continue
			}
			pkg, err := l.loadPackage(path)
			if err != nil {
				continue
			}
			pkg.Category = cat.Name()
			out = append(out, *pkg)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Category != out[j].Category {
			return out[i].Category < out[j].Category
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// GetPackage loads a single built-in package by its id by searching every
// category subfolder for a matching directory.
func (l *Library) GetPackage(id string) (*Package, error) {
	dir := l.packagesDir()
	cats, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading packages dir: %w", err)
	}
	for _, cat := range cats {
		if !cat.IsDir() {
			continue
		}
		path := filepath.Join(dir, cat.Name(), id, id+".yaml")
		if _, err := os.Stat(path); err != nil {
			continue
		}
		pkg, err := l.loadPackage(path)
		if err != nil {
			return nil, err
		}
		pkg.Category = cat.Name()
		return pkg, nil
	}
	return nil, fmt.Errorf("package %q not found", id)
}

// loadPackage reads a package file and resolves its imports against the
// library's perks directory.
func (l *Library) loadPackage(path string) (*Package, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading package %q: %w", path, err)
	}
	return ParsePackage(data, l.perksDir())
}

// perksDir is the directory holding the built-in perk definitions.
func (l *Library) perksDir() string {
	return filepath.Join(l.Root, "perks")
}

// ListPerks scans the perks directory and returns every loadable
// built-in perk, sorted by name. Individual files that fail to parse are
// skipped so one bad file does not break the whole browser.
func (l *Library) ListPerks() ([]model.Perk, error) {
	dir := l.perksDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading perks dir: %w", err)
	}
	var out []model.Perk
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		path := filepath.Join(dir, name, name+".yaml")
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		ab, err := export.UnmarshalPerk(data)
		if err != nil {
			continue
		}
		// Use the directory name as a stable library id for lookups.
		ab.ID = name
		out = append(out, ab)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// GetPerk loads a single built-in perk by its library id (the directory
// name).
func (l *Library) GetPerk(id string) (model.Perk, error) {
	path := filepath.Join(l.perksDir(), id, id+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return model.Perk{}, fmt.Errorf("reading perk %q: %w", id, err)
	}
	return export.UnmarshalPerk(data)
}

// ParsePackage parses package YAML bytes and resolves imports relative to
// baseDir (the perks directory). It is exported so custom (uploaded)
// package files can be parsed with a caller-supplied perks directory.

func ParsePackage(data []byte, baseDir string) (*Package, error) {
	var in PackageYAML
	if err := yaml.Unmarshal(data, &in); err != nil {
		return nil, fmt.Errorf("parsing package yaml: %w", err)
	}
	pkg := &Package{
		ID:          in.ID,
		Name:        in.Name,
		Description: in.Description,
		Shifts:      in.Shifts,
		Tags:        in.Tags,
	}
	if pkg.Name == "" {
		pkg.Name = pkg.ID
	}
	for _, imp := range in.Imports {
		abPath := resolveImport(imp, baseDir)
		abData, err := os.ReadFile(abPath)
		if err != nil {
			return nil, fmt.Errorf("reading imported perk %q: %w", imp, err)
		}
		ab, err := export.UnmarshalPerk(abData)
		if err != nil {
			return nil, fmt.Errorf("parsing imported perk %q: %w", imp, err)
		}
		pkg.Perks = append(pkg.Perks, ab)
	}
	return pkg, nil
}

// resolveImport turns an import entry into a path relative to baseDir (the
// perks directory). A short name like "fireball" resolves to
// "<baseDir>/fireball/fireball.yaml"; anything containing a path separator or a
// ".yaml" suffix is treated as an explicit relative path.
func resolveImport(imp, baseDir string) string {
	if filepath.Ext(imp) == ".yaml" || filepath.Ext(imp) == ".yml" ||
		containsSep(imp) {
		return filepath.Join(baseDir, imp)
	}
	return filepath.Join(baseDir, imp, imp+".yaml")
}

func containsSep(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '/' || s[i] == '\\' {
			return true
		}
	}
	return false
}
