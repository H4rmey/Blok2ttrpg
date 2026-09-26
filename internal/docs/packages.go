// This file renders the built-in content library (classes, races, backgrounds
// and items) as reader-facing markdown. The library lives on disk under
// library/packages and is loaded by internal/premade; previously none of it
// appeared in the documentation at all, so the rulebook listed no classes,
// races, backgrounds or items even though the application offered them.
//
// The tables are generated from the same loader the application uses, so a new
// package file shows up in the docs as soon as it is added.
package docs

import (
	"fmt"
	"sort"
	"strings"

	"github.com/harmey/blok2ttrpg-v5/internal/model"
	"github.com/harmey/blok2ttrpg-v5/internal/premade"
)

// PackageLister is the subset of the content library the docs need. It is an
// interface so the renderer does not depend on how the library is loaded and so
// tests can supply a fixed set of packages.
type PackageLister interface {
	ListPackages() ([]premade.Package, error)
	ListAbilities() ([]model.Ability, error)
}

// abilitiesTable renders the pre-built abilities that ship in the content
// library. The rulebook had headings for these lists but nothing underneath
// them, so the ready-made abilities the application offers were invisible to a
// reader even though they were sitting in the library directory.
func abilitiesTable(lib PackageLister) string {
	if lib == nil {
		return "_No abilities configured._"
	}
	abs, err := lib.ListAbilities()
	if err != nil {
		return fmt.Sprintf("_Could not read the ability library: %v_", err)
	}
	if len(abs) == 0 {
		return "_No abilities configured._"
	}
	sort.Slice(abs, func(i, j int) bool { return abs[i].Name < abs[j].Name })
	var b strings.Builder
	b.WriteString("| Ability | Type |\n")
	b.WriteString("| --- | --- |\n")
	for _, a := range abs {
		fmt.Fprintf(&b, "| **%s** | %s |\n", orDash(a.Name), orDash(a.Type))
	}
	return strings.TrimSpace(b.String())
}

// categoryLabels gives each library category a reader-facing heading. Categories
// not listed here fall back to a title-cased form of the directory name, so a
// new category folder still renders.
var categoryLabels = map[string]string{
	"classes":     "Classes",
	"races":       "Races",
	"backgrounds": "Backgrounds",
	"items":       "Items",
}

// packagesTable renders every package in the given categories as one section per
// category. Each package lists its description, the proficiency shifts it
// applies, the abilities it grants, and whether it can be toggled off.
//
// When categories is empty every category found in the library is rendered, in
// the order given by categoryOrder.
func packagesTable(lib PackageLister, categories ...string) string {
	if lib == nil {
		return "_No packages configured._"
	}
	pkgs, err := lib.ListPackages()
	if err != nil {
		return fmt.Sprintf("_Could not read the content library: %v_", err)
	}
	byCategory := groupPackages(pkgs)
	order := categories
	if len(order) == 0 {
		order = categoryOrder(byCategory)
	}

	var b strings.Builder
	for _, cat := range order {
		list := byCategory[cat]
		if len(list) == 0 {
			continue
		}
		writePackageCategory(&b, cat, list)
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return "_No packages configured._"
	}
	return out
}

// groupPackages buckets packages by their category, sorting each bucket by name
// so the rendered order is stable regardless of directory walk order.
func groupPackages(pkgs []premade.Package) map[string][]premade.Package {
	out := map[string][]premade.Package{}
	for _, p := range pkgs {
		out[p.Category] = append(out[p.Category], p)
	}
	for _, list := range out {
		sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	}
	return out
}

// preferredCategoryOrder is the order categories are presented in when the
// caller does not name them explicitly. Identity packages come first, then
// equipment; anything unrecognised is appended alphabetically.
var preferredCategoryOrder = []string{"classes", "races", "backgrounds", "items"}

// categoryOrder returns the categories present in the library, preferring the
// canonical order and appending any unknown categories alphabetically.
func categoryOrder(byCategory map[string][]premade.Package) []string {
	seen := map[string]bool{}
	var out []string
	for _, c := range preferredCategoryOrder {
		if len(byCategory[c]) > 0 {
			out = append(out, c)
			seen[c] = true
		}
	}
	var rest []string
	for c := range byCategory {
		if !seen[c] {
			rest = append(rest, c)
		}
	}
	sort.Strings(rest)
	return append(out, rest...)
}

// writePackageCategory writes the heading and table for a single category.
func writePackageCategory(b *strings.Builder, category string, list []premade.Package) {
	fmt.Fprintf(b, "### %s\n\n", categoryLabel(category))
	if premade.Toggleable(category) {
		b.WriteString("These may be picked up and put down freely: the trait shifts apply only while the package is active.\n\n")
	} else {
		b.WriteString("This is part of a character's core identity: once chosen it is always applied and cannot be switched off.\n\n")
	}
	b.WriteString("| Name | Description | Trait Shifts | Abilities Granted |\n")
	b.WriteString("| --- | --- | --- | --- |\n")
	for _, p := range list {
		fmt.Fprintf(b, "| **%s** | %s | %s | %s |\n",
			orDash(p.Name),
			orDash(oneLine(p.Description)),
			orDash(shiftsPhrase(p.Shifts)),
			orDash(abilityNames(p)))
	}
	b.WriteString("\n")
}

// categoryLabel returns the reader-facing heading for a library category.
func categoryLabel(category string) string {
	if l, ok := categoryLabels[category]; ok {
		return l
	}
	return titleCaseWord(category)
}

// shiftsPhrase renders a package's trait shifts as a readable list, e.g.
// "Power +2, Stealth -1". Trait keys are namespaced as "group.Trait"; the group
// prefix is dropped because the trait names are unique to the reader.
func shiftsPhrase(shifts map[string]int) string {
	if len(shifts) == 0 {
		return ""
	}
	keys := make([]string, 0, len(shifts))
	for k := range shifts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s %s", traitDisplayName(k), signed(shifts[k])))
	}
	return strings.Join(parts, ", ")
}

// traitDisplayName strips the group namespace from a "group.Trait" key.
func traitDisplayName(key string) string {
	if i := strings.IndexByte(key, '.'); i >= 0 {
		return key[i+1:]
	}
	return key
}

// abilityNames renders the names of the abilities a package grants.
func abilityNames(p premade.Package) string {
	if len(p.Abilities) == 0 {
		return ""
	}
	names := make([]string, 0, len(p.Abilities))
	for _, a := range p.Abilities {
		if n := strings.TrimSpace(a.Name); n != "" {
			names = append(names, n)
		}
	}
	return strings.Join(names, ", ")
}

// oneLine collapses newlines and pipe characters so a description is safe to
// place inside a markdown table cell.
func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "|", "/")
	return strings.Join(strings.Fields(s), " ")
}
