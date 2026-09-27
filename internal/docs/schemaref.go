// This file generates the developer-facing configuration reference from the Go
// types in internal/config, rather than from hand-written markdown.
//
// The previous configuration.md was maintained by hand and had drifted badly
// enough to describe a different codebase: it documented an `add_cost` key that
// is now `build_cost`, a `solutions` field type that is now `multiselect`, a
// `states.yaml` section that is now `conditions.yaml`, and a table of option
// sources that no longer existed. Reading the keys straight off the structs
// removes the possibility of that class of drift, and the accompanying purpose
// map is checked for completeness by LintSchemaCoverage, so adding a key to the
// schema without documenting it fails the test suite.
package docs

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
)

// documentedTypes maps the reference name used in the docs to the Go type whose
// yaml keys are rendered. Adding a struct here makes it documentable via the
// schemaTable helper and subject to the coverage lint.
var documentedTypes = map[string]reflect.Type{
	"Field":         reflect.TypeOf(config.Field{}),
	"Option":        reflect.TypeOf(config.Option{}),
	"Cost":          reflect.TypeOf(config.Cost{}),
	"PerStep":       reflect.TypeOf(config.PerStep{}),
	"GroupOffsets":  reflect.TypeOf(config.GroupOffsets{}),
	"InlineBuilder": reflect.TypeOf(config.InlineBuilder{}),
	"Component":     reflect.TypeOf(config.Component{}),
	"Proficiency":   reflect.TypeOf(config.Proficiency{}),
	"Leveling":      reflect.TypeOf(config.Leveling{}),
	"LevelTable":    reflect.TypeOf(config.LevelTable{}),
	"Condition":     reflect.TypeOf(config.Condition{}),
	"Validations":   reflect.TypeOf(config.Validations{}),
}

// schemaKey is one yaml key discovered on a documented struct.
type schemaKey struct {
	// Name is the yaml key as it is written in the config files.
	Name string
	// GoType is a readable rendering of the value's shape.
	GoType string
	// Purpose is the description from schemaPurposes.
	Purpose string
}

// schemaTable renders the yaml keys of a documented type as a markdown table.
// Keys appear in struct declaration order, which is the order a config author
// naturally reads them in.
func schemaTable(typeName string) string {
	keys, err := schemaKeysFor(typeName)
	if err != nil {
		return fmt.Sprintf("_%v_", err)
	}
	if len(keys) == 0 {
		return fmt.Sprintf("_%s has no configurable keys._", typeName)
	}
	var b strings.Builder
	b.WriteString("| Key | Value | Purpose |\n")
	b.WriteString("| --- | --- | --- |\n")
	for _, k := range keys {
		fmt.Fprintf(&b, "| `%s` | %s | %s |\n", k.Name, k.GoType, orDash(k.Purpose))
	}
	return strings.TrimSpace(b.String())
}

// schemaKeysFor returns the documented yaml keys of a registered type.
func schemaKeysFor(typeName string) ([]schemaKey, error) {
	t, ok := documentedTypes[typeName]
	if !ok {
		return nil, fmt.Errorf("unknown schema type %q", typeName)
	}
	purposes := schemaPurposes[typeName]
	var out []schemaKey
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name, ok := yamlKeyName(f)
		if !ok {
			continue
		}
		out = append(out, schemaKey{
			Name:    name,
			GoType:  describeType(f.Type),
			Purpose: purposes[name],
		})
	}
	return out, nil
}

// yamlKeyName extracts the yaml key from a struct field tag. It reports false
// for unexported fields, fields tagged "-" (never read from YAML), and fields
// with no yaml tag at all.
func yamlKeyName(f reflect.StructField) (string, bool) {
	if f.PkgPath != "" {
		return "", false // unexported
	}
	tag := f.Tag.Get("yaml")
	if tag == "" || tag == "-" {
		return "", false
	}
	name := strings.Split(tag, ",")[0]
	if name == "" || name == "-" {
		return "", false
	}
	return name, true
}

// describeType renders a Go type as the shape a config author would write in
// YAML: a scalar, a list, a mapping, or a named nested block.
func describeType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Ptr:
		// A pointer only signals "optional" in this schema, so the value shape
		// is that of the element, marked optional for the reader.
		return describeType(t.Elem()) + " (optional)"
	case reflect.Slice:
		return "list of " + describeType(t.Elem())
	case reflect.Map:
		return fmt.Sprintf("mapping of %s to %s", describeType(t.Key()), describeType(t.Elem()))
	case reflect.Struct:
		return "`" + t.Name() + "` block"
	case reflect.String:
		return "text"
	case reflect.Bool:
		return "true/false"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "whole number"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Interface:
		return "any scalar"
	default:
		return t.String()
	}
}

// LintSchemaCoverage reports documented types whose yaml keys have no purpose
// text. It is the guard that keeps this reference honest: a new key added to
// internal/config shows up here until it is described, so the docs cannot fall
// behind the schema without the test suite failing.
func LintSchemaCoverage() []string {
	var out []string
	names := make([]string, 0, len(documentedTypes))
	for name := range documentedTypes {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		keys, err := schemaKeysFor(name)
		if err != nil {
			out = append(out, err.Error())
			continue
		}
		for _, k := range keys {
			if strings.TrimSpace(k.Purpose) == "" {
				out = append(out, fmt.Sprintf(
					"%s.%s has no description; add one to schemaPurposes in internal/docs/schemaref.go",
					name, k.Name))
			}
		}
	}
	return out
}

// fieldTypesList renders the field types the builder and cost engine actually
// support. The list is derived from the switch statements in buildguide.go and
// render.go, so it cannot list types (such as the long-gone "solutions") that
// nothing implements.
func fieldTypesList() string {
	var b strings.Builder
	b.WriteString("| Type | Behaviour |\n")
	b.WriteString("| --- | --- |\n")
	for _, ft := range supportedFieldTypes {
		fmt.Fprintf(&b, "| `%s` | %s |\n", ft.name, ft.behaviour)
	}
	return strings.TrimSpace(b.String())
}

// supportedFieldTypes is the authoritative list of field types the code handles.
// Every entry here must be handled by writeGuideStep and writeFieldRows.
var supportedFieldTypes = []struct {
	name      string
	behaviour string
}{
	{"checkbox", "A toggle. Its `cost` applies only while it is checked."},
	{"dropdown", "Pick one value, from inline `options` or a named `options_source`. The field `cost` applies when a non-empty value is selected, plus the selected option's own cost and any `group_offsets` for its skill group."},
	{"free_text", "Free-form text. Never carries a cost."},
	{"free_number", "A bounded whole number using `min`, `max`, `step` and `rounding`. `per_step.increase` or `per_step.decrease` is charged per step away from `default`."},
	{"multiselect", "A repeatable set of rows built from `row_fields`, starting at `default_count` and pre-filled from `row_defaults`. `per_item` is charged per row added or removed relative to `default_count`."},
	{"conditions", "A repeatable set of condition rows. Behaves like `multiselect` but each row resolves against the `conditions` list, charging a shiftable condition's `shift_cost` per unit of shift or a fixed condition's `build_cost`/`energy_cost`."},
}

// optionSourcesTable lists the named option sources the loaded ruleset defines,
// with how many entries each holds and whether any of them carry a cost. It
// replaces a hand-written table that had invented source names.
func optionSourcesTable(cfg *config.Config) string {
	if cfg == nil {
		return "_No option sources configured._"
	}
	names := make([]string, 0, len(cfg.OptionSources))
	for name := range cfg.OptionSources {
		names = append(names, name)
	}
	sort.Strings(names)
	if len(names) == 0 {
		return "_No option sources configured._"
	}
	var b strings.Builder
	b.WriteString("| Source | Entries | Costed |\n")
	b.WriteString("| --- | --- | --- |\n")
	for _, name := range names {
		opts := cfg.OptionsFor(name)
		fmt.Fprintf(&b, "| `%s` | %d | %s |\n", name, len(opts), yesNo(anyOptionCosted(opts)))
	}
	return strings.TrimSpace(b.String())
}

// optionGroupsTable lists the grouped dropdown sources and the member sources
// each one concatenates, so an author can see what a grouped source expands to.
func optionGroupsTable(cfg *config.Config) string {
	if cfg == nil || len(cfg.OptionGroups) == 0 {
		return "_No option groups configured._"
	}
	names := make([]string, 0, len(cfg.OptionGroups))
	for name := range cfg.OptionGroups {
		names = append(names, name)
	}
	sort.Strings(names)
	var b strings.Builder
	b.WriteString("| Grouped Source | Member Groups |\n")
	b.WriteString("| --- | --- |\n")
	for _, name := range names {
		def := cfg.OptionGroups[name]
		parts := make([]string, 0, len(def.Groups))
		for _, g := range def.Groups {
			label := g.Label
			if label == "" {
				label = g.Source
			}
			parts = append(parts, fmt.Sprintf("%s (`%s`)", label, g.Source))
		}
		fmt.Fprintf(&b, "| `%s` | %s |\n", name, orDash(strings.Join(parts, ", ")))
	}
	return strings.TrimSpace(b.String())
}

// configFilesList lists the YAML section files the ruleset directory actually
// contains, so the reference names the real files rather than a hardcoded list
// that can fall out of date. The previous hand-written list was missing
// traits.yaml and still named a states.yaml that no longer exists.
func configFilesList(loaded *config.Loaded) string {
	if loaded == nil || loaded.Dir == "" {
		return "_No ruleset directory recorded._"
	}
	entries, err := os.ReadDir(loaded.Dir)
	if err != nil {
		return fmt.Sprintf("_Could not read the ruleset directory: %v_", err)
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if ext := filepath.Ext(e.Name()); ext != ".yaml" && ext != ".yml" {
			continue
		}
		files = append(files, e.Name())
	}
	if len(files) == 0 {
		return "_No section files found._"
	}
	sort.Strings(files)
	var b strings.Builder
	for _, f := range files {
		fmt.Fprintf(&b, "- `%s`\n", f)
	}
	return strings.TrimSpace(b.String())
}

// yesNo renders a boolean for a table cell.
func yesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
