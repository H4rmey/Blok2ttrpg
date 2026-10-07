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
	// Index is the 1-based position of the enactment within the perk.
	Index int
	// Title is the enactment's display name (e.g. "Enact Damage").
	Title string
	// Trigger states the circumstance that sets a reaction off, and is set only
	// on the first instruction of a perk that selected one. A reaction is not
	// used on your turn, so without this line the generated text describes what
	// the perk does but never says when it happens - which is the single most
	// important thing to know about a reaction at the table.
	//
	// It is perk-level rather than per-enactment: one trigger fires the whole
	// perk, however many enactments follow.
	Trigger string
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

// PerkInstructions generates one Instruction per enactment of an perk.
func PerkInstructions(cfg *config.Config, a model.Perk) []Instruction {
	if cfg == nil {
		return nil
	}
	var out []Instruction
	n := 0
	// inheritedPlural carries the singular/plural wording of the last enactment
	// that owned a target, so an enactment reusing that target words its lines
	// the same way.
	inheritedPlural := false
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
		// The trigger belongs to the perk, so it is stated once on the first
		// enactment rather than repeated on every one.
		if n == 1 {
			ins.Trigger = triggerLine(cfg, a)
		}
		plural := false
		// An enactment owns its target when it is the first one, or when the
		// author ticked "different target than the enactment before it". One
		// that does not simply reuses the previous enactment's target, so it
		// has neither its own Interaction nor its own roll.
		if n == 1 || en.NewTarget {
			ins.Interaction, plural = interactionLine(cfg, en)
			ins.Validation = validationLine(cfg, en, plural)
			inheritedPlural = plural
		} else {
			plural = inheritedPlural
			ins.Interaction = "Same target as the previous enactment."
			if plural {
				ins.Interaction = "The same targets as the previous enactment."
			}
			ins.Validation = "No separate roll; uses the previous enactment's result."
		}

		ins.Success = successLine(cfg, en, plural)
		ins.Solution = solutionLine(cfg, en)
		ins.Note = noteLine(cfg, en)
		out = append(out, ins)
	}
	return out
}

// triggerLine renders the reaction trigger a perk selected, together with the
// watch radius that bounds where the trigger may happen. It returns "" for any
// perk without a trigger, which is every non-reaction perk, so no caller needs
// to test the perk type.
//
// The human-readable text is the configured option's label, so rewording a
// trigger in general.yaml rewords it here with no code change. A stored value
// with no matching option falls back to the raw id rather than vanishing, which
// keeps a hand-edited or outdated perk legible.
func triggerLine(cfg *config.Config, a model.Perk) string {
	id := asString(a.Fields["trigger"])
	if id == "" {
		return ""
	}
	label := id
	for _, o := range cfg.ResolveOptions(config.Field{OptionsSource: "reaction_triggers"}) {
		if o.Value == id {
			if o.Label != "" {
				label = o.Label
			}
			break
		}
	}
	line := strings.TrimSuffix(label, ".")

	// The labels are written with a "within range" / "your reach" placeholder
	// standing in for the radius, because one label has to serve every range a
	// player might pick. Substituting the actual figure into that phrase reads
	// far better than appending a qualifier after it: tacking ", it must happen
	// to you" onto "someone within range is attacked" produces a sentence that
	// contradicts itself.
	//
	// Range 0 is not "no range" - it restricts the trigger to the engager, which
	// is the whole difference between a self-defence reaction and one that
	// guards a neighbour, so it collapses the phrase to "you" instead.
	if raw, ok := a.Fields["trigger_range"]; ok {
		r := asInt(raw)
		var who string
		if r == 0 {
			who = "you"
		} else {
			who = fmt.Sprintf("someone within %dm of you", r)
		}
		switch {
		case strings.Contains(line, "Someone within range"):
			line = strings.Replace(line, "Someone within range", capitalize(who), 1)
		case strings.Contains(line, "someone within range"):
			line = strings.Replace(line, "someone within range", who, 1)
		case r == 0 && strings.Contains(line, "Someone"):
			// A label with no range placeholder ("Someone leaves your reach")
			// already implies proximity, so at range 0 it only needs narrowing.
			line = strings.Replace(line, "Someone", "The creature", 1)
		default:
			// No placeholder to fill and a real radius: state it plainly rather
			// than silently dropping the number.
			if r > 0 {
				line += fmt.Sprintf(" (within %dm)", r)
			}
		}
		// Substituting "you" for a third-person subject leaves the verb
		// disagreeing ("You is attacked"). The labels are third-person singular
		// by convention, so the few verbs that actually appear are corrected
		// here. This is a small closed set, not general-purpose conjugation: if a
		// new label needs a form that is not listed, it shows up immediately as
		// bad grammar in the generated text rather than as a silent error.
		if r == 0 {
			for from, to := range map[string]string{
				"You is ":   "You are ",
				"You has ":  "You have ",
				"You does ": "You do ",
			} {
				line = strings.Replace(line, from, to, 1)
			}
		}
	}
	return line + "."
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
// skills needs no roll. An enactment that opts into flat-DC validation (see
// Component.UseDCValidation) rolls its engage source against a fixed DC taken
// from general.yaml instead of against the target's skills.
func validationLine(cfg *config.Config, en model.Enactment, plural bool) string {
	engage := asString(en.ValidationData["engage"])
	if cfg != nil {
		if ec, ok := cfg.Enactment(en.Type); ok && ec.UsesDCValidation() {
			if engage == "" {
				return "No roll required."
			}
			return fmt.Sprintf("Roll %s vs DC %d.", rollText(engage), cfg.DCValidationDC(asInt(en.ValidationData["validation_dc"])))
		}
	}
	counters := skillNames(asRows(en.ValidationData["counter_skill"]))
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
		dirs := skillNames(asRows(f["direction"]))
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
	case "shift":
		skill := skillName(asString(f["shift-skills"]))
		shift := asInt(f["shift-amount"])
		rounds := asInt(f["shift-duration"])
		return fmt.Sprintf("Shift %s's %s %s for %s.",
			targets, skill, shiftWords(shift), rounds2str(maxInt(rounds, 1)))
	case "phase":
		skills := skillNames(asRows(f["shift-affected-skills"]))
		shift := asInt(f["phase-shift-amount"])
		rounds := maxInt(asInt(f["phase-duration"]), 1)
		if len(skills) == 0 {
			return ""
		}
		return fmt.Sprintf("Shift %s %s for %s, then %s for the same duration.",
			joinAnd(skills), shiftWords(shift), rounds2str(rounds), shiftWords(-shift))
	case "nerf":
		skill := skillName(asString(f["nerf-shifted-skill"]))
		shift := asInt(f["nerf-shift-amount"])
		rounds := maxInt(asInt(f["nerf-duration"]), 1)
		return fmt.Sprintf("Shift your %s %s for %s.", skill, shiftWords(shift), rounds2str(rounds))
	case "negation":
		return "The targeted enactment is nullified and has no effect."
	case "adjustment":
		return fmt.Sprintf("Add %s to that enactment's Source result.", rollText(asString(f["source"])))
	case "resource":
		name := asString(f["resource_name"])
		if name == "" {
			name = "resources"
		}
		switch asString(f["resource_mode"]) {
		case "consume":
			effect := asString(f["consumed_effect"])
			if effect == "" {
				effect = "the linked effect"
			}
			return fmt.Sprintf("Consume stored %s to execute %s once per resource consumed.", name, strings.ToLower(effect))
		default:
			amount := asInt(f["generate_amount"])
			if amount < 1 {
				amount = 1
			}
			return fmt.Sprintf("Gain %d %s.", amount, name)
		}
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
// only produced for enactments that define solution fields. The config names
// them solution_1/solution_2 (two dropdowns plus a solution_dc number); older
// files may carry a legacy "solution" row list, which is still honoured.
func solutionLine(cfg *config.Config, en model.Enactment) string {
	var names []string
	seen := map[string]bool{}
	add := func(v string) {
		name := skillName(v)
		if name == "" || seen[name] {
			return
		}
		seen[name] = true
		names = append(names, name)
	}
	// Current config shape: two plain dropdown values.
	add(asString(en.Fields["solution_1"]))
	add(asString(en.Fields["solution_2"]))
	// Legacy shape: a repeatable row list under "solution".
	for _, n := range skillNames(asRows(en.Fields["solution"])) {
		if !seen[n] {
			seen[n] = true
			names = append(names, n)
		}
	}
	if len(names) == 0 {
		return ""
	}
	dc := asInt(en.Fields["solution_dc"])
	if dc <= 0 {
		dc = defaultSolutionDC(cfg, en.Type)
	}
	what := "the effect"
	if en.Type == "condition" {
		what = "the condition"
	}
	return fmt.Sprintf("1 Action, %s vs DC %d ends %s.", joinOr(names), dc, what)
}

// defaultSolutionDC reads the configured default of the solution_dc field for
// the given enactment type, so an instruction never reports DC 0 when the
// stored value is missing or was clamped away.
func defaultSolutionDC(cfg *config.Config, enactmentID string) int {
	if cfg == nil {
		return 0
	}
	if ec, ok := cfg.Enactment(enactmentID); ok {
		for _, f := range ec.Fields {
			if f.Key == "solution_dc" {
				if n := asInt(f.Default); n > 0 {
					return n
				}
			}
		}
	}
	return 0
}
