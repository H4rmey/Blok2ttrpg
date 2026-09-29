// This file holds the plain-language phrasing helpers the build guide and its
// tables share: cost clauses, pluralisation, number words and sentence
// tidying. They exist so wording stays consistent across every generated
// section rather than being re-spelled at each call site.
package docs

import (
	"fmt"
	"strings"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
)

// costClause renders a trailing cost sentence for a toggle-style field, e.g.
// "Cost: Free." or "Cost: 2 build.".
func costClause(c *config.Cost, _ string) string {
	return "Cost: " + costWords(c) + "."
}

// flatCostClause renders the cost sentence for a dropdown's flat field cost,
// or "" when there is none.
func flatCostClause(c *config.Cost) string {
	if !hasCost(c) {
		return ""
	}
	return "Cost: " + costWords(c) + "."
}

// perStepClause renders the per-step cost sentence for a free_number field.
func perStepClause(f config.Field) string {
	unit := stepUnit(f.Label)
	if f.PerStep != nil && (hasCost(f.PerStep.Increase) || hasCost(f.PerStep.Decrease)) {
		inc := f.PerStep.Increase
		return fmt.Sprintf("Cost: %s per %s.", costWords(inc), unit)
	}
	return fmt.Sprintf("Cost: Free per %s.", unit)
}

// perItemClause renders the per-entry cost sentence for a repeatable list.
func perItemClause(f config.Field) string {
	item := strings.ToLower(singular(f.Label))
	if f.PerItem != nil && (hasCost(f.PerItem.Increase) || hasCost(f.PerItem.Decrease)) {
		return fmt.Sprintf("Cost: %s per %s.", costWords(f.PerItem.Increase), item)
	}
	return fmt.Sprintf("Cost: Free per %s.", item)
}

// stepUnit picks a natural per-step noun from a number field's label.
func stepUnit(label string) string {
	l := strings.ToLower(label)
	switch {
	case strings.Contains(l, "duration") || strings.Contains(l, "round"):
		return "round"
	case strings.Contains(l, "distance") || strings.Contains(l, "range") || strings.Contains(l, "meter"):
		return "meter"
	case strings.Contains(l, "shift"):
		return "step"
	case strings.Contains(l, "bonus"):
		return "+1"
	default:
		return "step"
	}
}

// ensureSentence trims a string and appends a period when it lacks terminal
// punctuation, so descriptions read as complete sentences.
func ensureSentence(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	switch s[len(s)-1] {
	case '.', '!', '?', ':':
		return s
	}
	return s + "."
}

// numberWord spells out small counts (used for "start with two skills").
func numberWord(n int) string {
	words := []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten"}
	if n >= 0 && n < len(words) {
		return words[n]
	}
	return fmt.Sprintf("%d", n)
}

// pluralize picks the singular or plural noun for a count.
func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}

// defaultStr renders a field default value (which decodes as an arbitrary YAML
// scalar) as a plain string for prose.
func defaultStr(v any) string {
	if v == nil {
		return "0"
	}
	switch t := v.(type) {
	case string:
		if t == "" {
			return "0"
		}
		return t
	case int:
		return fmt.Sprintf("%d", t)
	case int64:
		return fmt.Sprintf("%d", t)
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%v", t)
	default:
		return fmt.Sprintf("%v", t)
	}
}
