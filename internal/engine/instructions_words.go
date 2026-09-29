// This file holds the wording helpers the instruction generator uses: turning
// a roll source, skill id, condition id or count into the phrase that appears
// on a character sheet. They live apart from the generator so the sentence
// structure and the vocabulary can be read independently.
package engine

import (
	"fmt"
	"strings"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
)

// rollText renders a roll source. A plain die (d4..d20) prints literally as
// "1dX"; a namespaced skill prints as "your <Skill> die" so it scales with the
// character's proficiency.
func rollText(v string) string {
	if v == "" {
		return "no roll"
	}
	if i := strings.IndexByte(v, '.'); i >= 0 {
		return "your " + v[i+1:] + " die"
	}
	// A die value, optionally with a flat bonus: "d8" reads as "1d8" and
	// "d12+3" as "1d12+3".
	if len(v) > 1 && v[0] == 'd' {
		return "1" + v
	}
	// A plain number is a flat result with no roll at all (the bottom rung of
	// the ladder), so it prints as the number itself rather than as a die.
	if isNumeric(v) {
		return v
	}
	return "your " + v + " die"
}

// isNumeric reports whether s consists only of digits, i.e. it is a flat
// numeric result rather than a die or a skill name.
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// skillName strips the option-source namespace from a skill value.
func skillName(v string) string {
	if i := strings.IndexByte(v, '.'); i >= 0 {
		return v[i+1:]
	}
	return v
}

// skillNames pulls the "value" column out of multiselect rows, de-namespacing
// each entry and dropping blanks.
func skillNames(rows []map[string]any) []string {
	var out []string
	for _, r := range rows {
		v := skillName(asString(r["value"]))
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

// conditionName resolves a condition id to its display name.
func conditionName(cfg *config.Config, id string) string {
	id = strings.TrimPrefix(strings.TrimPrefix(id, "general."), "specific.")
	if id == "" {
		return "a condition"
	}
	if c, ok := cfg.ConditionByID(id); ok {
		return c.Name
	}
	if c, ok := cfg.GeneralConditionByID(id); ok {
		return c.Name
	}
	if c, ok := cfg.SpecificConditionByID(id); ok {
		return c.Name
	}
	return id
}

// conditionDescription resolves a condition id to its description text, or ""
// when the condition is unknown or has none.
func conditionDescription(cfg *config.Config, id string) string {
	id = strings.TrimPrefix(strings.TrimPrefix(id, "general."), "specific.")
	if id == "" {
		return ""
	}
	if c, ok := cfg.ConditionByID(id); ok {
		return c.Description
	}
	if c, ok := cfg.GeneralConditionByID(id); ok {
		return c.Description
	}
	if c, ok := cfg.SpecificConditionByID(id); ok {
		return c.Description
	}
	return ""
}

// shiftWords turns a signed shift amount into readable direction text.
func shiftWords(n int) string {
	switch {
	case n > 0:
		return fmt.Sprintf("up by %d", n)
	case n < 0:
		return fmt.Sprintf("down by %d", -n)
	default:
		return "by 0"
	}
}

func signed(n int) string {
	if n > 0 {
		return fmt.Sprintf("+%d", n)
	}
	return fmt.Sprintf("%d", n)
}

func rounds2str(n int) string {
	if n == 1 {
		return "1 round"
	}
	return fmt.Sprintf("%d rounds", n)
}

func turns2str(n int) string {
	if n == 1 {
		return "1 turn"
	}
	return fmt.Sprintf("%d turns", n)
}

func joinOr(items []string) string  { return joinWith(items, "or") }
func joinAnd(items []string) string { return joinWith(items, "and") }

func joinWith(items []string, word string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " " + word + " " + items[1]
	default:
		return strings.Join(items[:len(items)-1], ", ") + " " + word + " " + items[len(items)-1]
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
