// This file implements a post-render lint over the generated documentation.
//
// The docs are produced by executing markdown templates against the loaded
// config. When the config shape drifts away from what a template expects, Go's
// text/template fails silently: a range over a missing or empty slice renders
// nothing at all, so a table keeps its header and separator but loses every
// body row. Nothing in the pipeline used to notice, which is how the ability
// point leveling table shipped empty.
//
// Lint therefore inspects the rendered markdown and reports the shapes that
// indicate a template lost its data. cmd/gendocs and the docs render test both
// treat a non-empty result as a failure, so drift between the code and the
// documentation becomes a build/test error rather than a silent hole in the
// rulebook.
package docs

import (
	"fmt"
	"strings"
)

// Finding is a single lint problem found in the rendered documentation.
type Finding struct {
	// Line is the 1-based line number the problem was found on.
	Line int
	// Message describes the problem in terms a maintainer can act on.
	Message string
}

// String renders a finding as "line 42: message".
func (f Finding) String() string {
	return fmt.Sprintf("line %d: %s", f.Line, f.Message)
}

// FindingsText renders a list of findings as an indented, newline-separated
// block suitable for a log or test failure message.
func FindingsText(fs []Finding) string {
	parts := make([]string, 0, len(fs))
	for _, f := range fs {
		parts = append(parts, "  "+f.String())
	}
	return strings.Join(parts, "\n")
}

// Lint reports problems in the rendered markdown documentation. An empty
// result means the docs rendered with all of their data present.
func Lint(md string) []Finding {
	lines := strings.Split(md, "\n")
	var out []Finding
	out = append(out, lintEmptyTables(lines)...)
	out = append(out, lintPlaceholders(lines)...)
	out = append(out, lintTemplateLeftovers(lines)...)
	out = append(out, lintEmptySections(lines)...)
	return out
}

// lintEmptyTables reports markdown tables that have a header row and a
// separator row but no body rows. This is the signature of a template whose
// range produced no items.
func lintEmptyTables(lines []string) []Finding {
	var out []Finding
	for i, line := range lines {
		if !isTableSeparator(line) {
			continue
		}
		// A separator must be preceded by a header row and followed by at
		// least one body row to be a populated table.
		if i == 0 || !isTableRow(lines[i-1]) {
			continue
		}
		if i+1 < len(lines) && isTableRow(lines[i+1]) {
			continue
		}
		out = append(out, Finding{
			Line: i + 1,
			Message: fmt.Sprintf("table has a header and separator but no rows (header: %q); "+
				"a template range most likely produced no items", strings.TrimSpace(lines[i-1])),
		})
	}
	return out
}

// placeholderMarkers are the "nothing configured" fallbacks the renderer emits.
// They are legitimate defensive output, but reaching one in the real ruleset
// means a doc is asking for data the config does not supply.
var placeholderMarkers = []string{
	"_No conditions configured._",
	"_No traits configured._",
	"_No attribute sections configured._",
	"_No options configured._",
	"_No packages configured._",
	"_This component has no cost-bearing perks",
}

// lintPlaceholders reports rendered "nothing configured" fallback text.
func lintPlaceholders(lines []string) []Finding {
	var out []Finding
	for i, line := range lines {
		for _, m := range placeholderMarkers {
			if strings.Contains(line, m) {
				out = append(out, Finding{
					Line:    i + 1,
					Message: fmt.Sprintf("rendered an empty-data placeholder: %s", strings.TrimSpace(line)),
				})
				break
			}
		}
	}
	return out
}

// lintTemplateLeftovers reports unexecuted template syntax. Because the docs
// are executed with text/template, a surviving delimiter means the template
// author escaped it or the action was inside a block that never ran.
func lintTemplateLeftovers(lines []string) []Finding {
	var out []Finding
	for i, line := range lines {
		if strings.Contains(line, "{{") || strings.Contains(line, "}}") {
			out = append(out, Finding{
				Line:    i + 1,
				Message: fmt.Sprintf("unexecuted template syntax: %s", strings.TrimSpace(line)),
			})
		}
	}
	return out
}

// lintEmptySections reports headings whose section is genuinely empty.
//
// Two shapes are deliberately not reported, because neither indicates lost data:
//
//   - A heading followed by a deeper heading is ordinary nesting ("# Chapter"
//     then "## Section"); the subsection is the content.
//   - A heading followed by a sibling heading is the house style used throughout
//     these docs, where a file opens with its slug as a title and repeats it as
//     the first section ("# items" then "## Items").
//
// What is left, and what this reports, is a heading followed by a *shallower*
// heading or by the end of the document. Such a heading is dangling: the section
// it introduces was closed out by its parent without ever being filled, which is
// exactly what happens when the helper meant to fill it returned nothing.
func lintEmptySections(lines []string) []Finding {
	var out []Finding
	for i, line := range lines {
		if !isHeading(line) {
			continue
		}
		if sectionHasContent(lines, i) {
			continue
		}
		out = append(out, Finding{
			Line:    i + 1,
			Message: fmt.Sprintf("heading has no content: %s", strings.TrimSpace(line)),
		})
	}
	return out
}

// sectionHasContent reports whether the heading at index i is followed by body
// text, a nested subheading, or a sibling heading. Only a shallower heading or
// the end of the document leaves it genuinely dangling.
func sectionHasContent(lines []string, i int) bool {
	level := headingLevel(lines[i])
	for j := i + 1; j < len(lines); j++ {
		s := strings.TrimSpace(lines[j])
		if s == "" {
			continue
		}
		if isHeading(lines[j]) {
			return headingLevel(lines[j]) >= level
		}
		return true
	}
	return false
}

// isHeading reports whether a line is an ATX markdown heading.
func isHeading(line string) bool {
	s := strings.TrimSpace(line)
	if !strings.HasPrefix(s, "#") {
		return false
	}
	// An ATX heading requires whitespace after its run of hashes, which keeps a
	// comment or an id-style "#foo" from being mistaken for a heading.
	rest := strings.TrimLeft(s, "#")
	return rest == "" || strings.HasPrefix(rest, " ") || strings.HasPrefix(rest, "\t")
}

// headingLevel returns the number of leading hashes of an ATX heading.
func headingLevel(line string) int {
	s := strings.TrimSpace(line)
	n := 0
	for n < len(s) && s[n] == '#' {
		n++
	}
	return n
}

// isTableRow reports whether a line looks like a markdown table row.
func isTableRow(line string) bool {
	s := strings.TrimSpace(line)
	return strings.HasPrefix(s, "|") && strings.HasSuffix(s, "|") && len(s) > 1
}

// isTableSeparator reports whether a line is a markdown table separator row,
// i.e. every cell consists only of dashes and alignment colons.
func isTableSeparator(line string) bool {
	if !isTableRow(line) {
		return false
	}
	s := strings.Trim(strings.TrimSpace(line), "|")
	cells := strings.Split(s, "|")
	if len(cells) == 0 {
		return false
	}
	for _, c := range cells {
		c = strings.TrimSpace(c)
		if c == "" {
			return false
		}
		if strings.Trim(c, "-:") != "" {
			return false
		}
		if !strings.Contains(c, "-") {
			return false
		}
	}
	return true
}
