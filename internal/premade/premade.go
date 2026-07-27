// Package premade loads the built-in content library that ships with the app:
// importable packages and the abilities they reference. A package is a
// collection of proficiency shifts plus a list of ability imports; importing a
// package applies its shifts and copies its abilities onto a character.
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

// PackageYAML is the on-disk shape of a package definition. Shifts map a trait
// key ("group.Trait") to a relative proficiency delta (e.g. +2, -1). Imports
// lists abilities to include; each entry is either a short name (resolved to
// ../../abilities/<name>/<name>.yaml relative to the package file) or an
// explicit path relative to the package file.
type PackageYAML struct {
	ID          string         `yaml:"id"`
	Name        string         `yaml:"name"`
	Description string         `yaml:"description,omitempty"`
	Shifts      map[string]int `yaml:"shifts,omitempty"`
	Imports     []string       `yaml:"imports,omitempty"`
}

// Package is a loaded package: its metadata, the proficiency shifts it applies,
// and the fully-parsed abilities it provides.
type Package struct {
	ID          string
	Name        string
	Description string
	Shifts      map[string]int
	Abilities   []model.Ability
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

// packagesDir is the directory holding package definitions.
func (l *Library) packagesDir() string {
	return filepath.Join(l.Root, "packages")
}

// CustomBaseDir returns the base directory used to resolve imports for an
// uploaded (custom) package. It mirrors the on-disk layout of a built-in
// package (library/packages/<name>) so short-name imports still resolve to
// library/abilities/<name>/<name>.yaml.
func (l *Library) CustomBaseDir() string {
	return filepath.Join(l.packagesDir(), "_custom")
}

// ListPackages scans the packages directory and returns every loadable package,
// sorted by name. Individual packages that fail to parse are skipped so one bad
// file does not break the whole browser.
func (l *Library) ListPackages() ([]Package, error) {
	dir := l.packagesDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading packages dir: %w", err)
	}
	var out []Package
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		path := filepath.Join(dir, name, name+".yaml")
		if _, err := os.Stat(path); err != nil {
			continue
		}
		pkg, err := LoadPackage(path)
		if err != nil {
			continue
		}
		out = append(out, *pkg)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// GetPackage loads a single built-in package by its id (which matches the
// directory name).
func (l *Library) GetPackage(id string) (*Package, error) {
	path := filepath.Join(l.packagesDir(), id, id+".yaml")
	return LoadPackage(path)
}

// abilitiesDir is the directory holding the built-in ability definitions.
func (l *Library) abilitiesDir() string {
	return filepath.Join(l.Root, "abilities")
}

// ListAbilities scans the abilities directory and returns every loadable
// built-in ability, sorted by name. Individual files that fail to parse are
// skipped so one bad file does not break the whole browser.
func (l *Library) ListAbilities() ([]model.Ability, error) {
	dir := l.abilitiesDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading abilities dir: %w", err)
	}
	var out []model.Ability
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
		ab, err := export.UnmarshalAbility(data)
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

// GetAbility loads a single built-in ability by its library id (the directory
// name).
func (l *Library) GetAbility(id string) (model.Ability, error) {
	path := filepath.Join(l.abilitiesDir(), id, id+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return model.Ability{}, fmt.Errorf("reading ability %q: %w", id, err)
	}
	return export.UnmarshalAbility(data)
}

// LoadPackage parses a package YAML file and resolves its ability imports into
// fully-parsed abilities. Import entries may be short names or explicit paths,
// both resolved relative to the package file's directory.
func LoadPackage(path string) (*Package, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading package %q: %w", path, err)
	}
	return ParsePackage(data, filepath.Dir(path))
}

// ParsePackage parses package YAML bytes and resolves imports relative to
// baseDir. It is separated from LoadPackage so custom (uploaded) package files
// can be parsed with a caller-supplied base directory.
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
	}
	if pkg.Name == "" {
		pkg.Name = pkg.ID
	}
	for _, imp := range in.Imports {
		abPath := resolveImport(imp, baseDir)
		abData, err := os.ReadFile(abPath)
		if err != nil {
			return nil, fmt.Errorf("reading imported ability %q: %w", imp, err)
		}
		ab, err := export.UnmarshalAbility(abData)
		if err != nil {
			return nil, fmt.Errorf("parsing imported ability %q: %w", imp, err)
		}
		pkg.Abilities = append(pkg.Abilities, ab)
	}
	return pkg, nil
}

// resolveImport turns an import entry into an absolute-ish path relative to the
// package's directory. A short name like "fireball" resolves to
// "../../abilities/fireball/fireball.yaml"; anything containing a path
// separator or a ".yaml" suffix is treated as an explicit relative path.
func resolveImport(imp, baseDir string) string {
	if filepath.Ext(imp) == ".yaml" || filepath.Ext(imp) == ".yml" ||
		containsSep(imp) {
		return filepath.Join(baseDir, imp)
	}
	return filepath.Join(baseDir, "..", "..", "abilities", imp, imp+".yaml")
}

func containsSep(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '/' || s[i] == '\\' {
			return true
		}
	}
	return false
}
