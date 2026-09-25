package engine

import (
	"fmt"
	"strings"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// This file implements the play-facing "instruction" generator. Where the cost
// engine turns builder choices into points, this turns the same choices into
// the short rules text a player reads at the table:
//
//   Interaction: 2 targets within 25m.
//   Validation:  Roll Precision vs the target's Reflex or Constitution.
//   On success:  Deal your Precision die in damage to each target that failed.
//   Solution:    1 Action, Reflex or Constitution vs DC 6 ends the condition.
//
// Every line is derived mechanically from config field values, so no enactment
// or interaction is special-cased beyond reading the keys it defines. Unknown
// ids simply yield fewer lines rather than an error.

// Instruction is the generated rules text for a single enactment.
type Instruction struct {
	// Index is the 1-based position of the enactment within the ability.
	Index int
	// Title is the enactment's display name (e.g. "Enact Damage").
	Title string
	// Interaction, Validation, Success and Solution are the generated lines.
	// An empty string means the line does not apply and is not rendered.
	Interaction string
	Validation  string
	Success     string
	Solution    string
	// Note is optional supporting detail rendered under the generated lines,
	// such as the description of the condition an enactment applies. It is a
	// reminder of what the chosen option does, not a separate rule.
	Note string
}

// AbilityInstructions generates one Instruction per enactment of an ability.
func AbilityInstructions(cfg *config.Config, a model.Ability) []Instruction {
	if cfg == nil {
		return nil
	}
	var out []Instruction
	n := 0
	for _, en := range a.Enactments {
		if en.Type == "" {
			continue
		}
		n++
		ins := Instruction{Index: n}
		ec, haveEnactment := cfg.Enactment(en.Type)
		if haveEnactment {
			ins.Title = ec.DisplayName()
		} else {
			ins.Title = en.Type
		}
		plural := false
		ins.Interaction, plural = interactionLine(cfg, en)
		ins.Validation = validationLine(cfg, en, plural)
		ins.Success = successLine(cfg, en, plural)
		ins.Solution = solutionLine(cfg, en)
		ins.Note = noteLine(cfg, en)
		out = append(out, ins)
	}
	return out
}

// interactionLine describes who the enactment hits. It also reports whether the
// enactment affects more than one target, which the other lines use to pick
// singular or plural wording.
func interactionLine(cfg *config.Config, en model.Enactment) (string, bool) {
	if en.Interaction == "" {
		return "", false
	}
	d := en.InteractionData
	switch en.Interaction {
	case "self":
		return "Self.", false
	case "direct":
		rng := asInt(d["range"])
		targets := asInt(d["targets"])
		if targets < 1 {
			targets = 1
		}
		if rng < 1 {
			rng = 1
		}
		who := "1 target"
		if targets > 1 {
			who = fmt.Sprintf("up to %d targets", targets)
		}
		where := fmt.Sprintf("within %dm", rng)
		if rng <= 1 {
			where = "within 1m (melee)"
		}
		return fmt.Sprintf("%s %s.", capitalize(who), where), targets > 1
	case "zone":
		rng := asInt(d["zone-range"])
		radius := asInt(d["zone-radius"])
		rounds := asInt(d["zone-duration"])
		if rng < 1 {
			rng = 1
		}
		if radius < 1 {
			radius = 1
		}
		line := fmt.Sprintf("%dm radius centred on a point within %dm", radius, rng)
		if rounds > 0 {
			line += fmt.Sprintf(", for %s", rounds2str(rounds))
		}
		return line + ".", true
	}
	// Unknown interaction: fall back to its display name.
	if ic, ok := cfg.Interaction(en.Interaction); ok {
		return ic.DisplayName() + ".", false
	}
	return "", false
}

// validationLine describes the contested roll. An engage source with no counter
// traits (or a self interaction) needs no roll.
func validationLine(cfg *config.Config, en model.Enactment, plural bool) string {
	engage := asString(en.ValidationData["engage"])
	counters := traitNames(asRows(en.ValidationData["counter_trait"]))
	if engage == "" || len(counters) == 0 {
		return "No roll required."
	}
	subject := "the target's"
	if plural {
		subject = "each target's"
	}
	return fmt.Sprintf("Roll %s vs %s %s.", rollText(engage), subject, joinOr(counters))
}

// successLine describes what happens when the validation succeeds. Each branch
// reads only the field keys its own enactment defines.
func successLine(cfg *config.Config, en model.Enactment, plural bool) string {
	f := en.Fields
	targets := "the target"
	failed := "the target"
	if plural {
		targets = "each target"
		failed = "each target that failed"
	}
	switch en.Type {
	case "damage":
		return fmt.Sprintf("Deal %s damage to %s.", rollText(asString(f["source"])), failed)
	case "healing":
		return fmt.Sprintf("Restore %s health.", rollText(asString(f["source"])))
	case "motion":
		dist := asInt(f["distance"])
		dirs := traitNames(asRows(f["direction"]))
		dir := ""
		if len(dirs) > 0 {
			dir = " " + strings.ToLower(joinOr(dirs))
		}
		return fmt.Sprintf("Move %s %dm%s.", targets, dist, dir)
	case "condition":
		name := conditionName(cfg, asString(f["condition"]))
		turns := asInt(f["duration"])
		line := fmt.Sprintf("%s gains %s", capitalize(targets), name)
		if shift := asInt(f["shift_amount"]); shift != 0 {
			line += fmt.Sprintf(" (%s)", signed(shift))
		}
		if turns > 0 {
			line += fmt.Sprintf(" for %s", turns2str(turns))
		}
		return line + "."
	case "effect":
		kind := asString(f["effect_type"])
		rounds := asInt(f["effect-duration"])
		src := ""
		if nested, ok := f["effect_type_ib"].(map[string]any); ok {
			src = asString(nested["source"])
		}
		what := strings.ToLower(kind)
		if what == "" {
			what = "the effect"
		}
		line := fmt.Sprintf("For %s, at the start of each of %s's turns apply %s",
			rounds2str(maxInt(rounds, 1)), targets, what)
		if src != "" {
			line += fmt.Sprintf(" of %s", rollText(src))
		}
		return line + "."
	case "modification":
		trait := traitName(asString(f["modification-traits"]))
		shift := asInt(f["modification-shift-amount"])
		rounds := asInt(f["modificaiton-shift-duration"])
		return fmt.Sprintf("Shift %s's %s %s for %s.",
			targets, trait, shiftWords(shift), rounds2str(maxInt(rounds, 1)))
	case "phase":
		traits := traitNames(asRows(f["shift-affected-traits"]))
		shift := asInt(f["phase-shift-amount"])
		rounds := maxInt(asInt(f["phase-duration"]), 1)
		if len(traits) == 0 {
			return ""
		}
		return fmt.Sprintf("Shift %s %s for %s, then %s for the same duration.",
			joinAnd(traits), shiftWords(shift), rounds2str(rounds), shiftWords(-shift))
	case "nerf":
		trait := traitName(asString(f["nerf-shifted-trait"]))
		shift := asInt(f["nerf-shift-amount"])
		rounds := maxInt(asInt(f["nerf-duration"]), 1)
		return fmt.Sprintf("Shift your %s %s for %s.", trait, shiftWords(shift), rounds2str(rounds))
	case "negation":
		return "The targeted enactment is nullified and has no effect."
	case "adjustment":
		return fmt.Sprintf("Add %s to that enactment's Source result.", rollText(asString(f["source"])))
	}
	return ""
}

// noteLine returns optional supporting detail for an enactment. Condition
// enactments quote the selected condition's description so the player does not
// have to look it up.
func noteLine(cfg *config.Config, en model.Enactment) string {
	if en.Type != "condition" {
		return ""
	}
	return conditionDescription(cfg, asString(en.Fields["condition"]))
}

// solutionLine describes how a target ends a condition or effect early. It is
// only produced for enactments that define solution fields.
func solutionLine(cfg *config.Config, en model.Enactment) string {
	rows := asRows(en.Fields["solution"])
	names := traitNames(rows)
	if len(names) == 0 {
		return ""
	}
	dc := asInt(en.Fields["solution_dc"])
	what := "the effect"
	if en.Type == "condition" {
		what = "the condition"
	}
	return fmt.Sprintf("1 Action, %s vs DC %d ends %s.", joinOr(names), dc, what)
}

// rollText renders a roll source. A plain die (d4..d20) prints literally as
// "1dX"; a namespaced trait prints as "your <Trait> die" so it scales with the
// character's proficiency.
func rollText(v string) string {
	if v == "" {
		return "no roll"
	}
	if i := strings.IndexByte(v, '.'); i >= 0 {
		return "your " + v[i+1:] + " die"
	}
	if len(v) > 1 && v[0] == 'd' {
		return "1" + v
	}
	return "your " + v + " die"
}

// traitName strips the option-source namespace from a trait value.
func traitName(v string) string {
	if i := strings.IndexByte(v, '.'); i >= 0 {
		return v[i+1:]
	}
	return v
}

// traitNames pulls the "value" column out of multiselect rows, de-namespacing
// each entry and dropping blanks.
func traitNames(rows []map[string]any) []string {
	var out []string
	for _, r := range rows {
		v := traitName(asString(r["value"]))
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
