// Conditions as a derived overlay on a character's skills.
//
// The rule this file implements is deliberately narrow: an applied condition
// never changes what is stored on a character. Character.Skills always holds
// the proficiency the player bought; the effect of every applied condition is
// recomputed here, from scratch, every time the sheet is rendered.
//
// The alternative - writing the shift into Character.Skills on apply and
// subtracting it on remove - looks simpler and is a trap. Shifts clamp at the
// ends of the proficiency ladder, so a shift that was clamped on the way down
// does not restore the original rung on the way back up, and any other change
// in between (a package toggle, a manual edit) moves the baseline the removal
// subtracts from. Both silently corrupt a character. Deriving the value instead
// makes removal exact by construction: delete the entry and the overlay is gone.
//
// Stacking: shifts from several conditions on the same skill are summed, then
// the sum is applied once. Summing rather than taking the worst is what makes
// piling conditions on a single target a real tactic, but it also means two
// cheap debuffs can add up to a large penalty - see the note in conditions.yaml.
package engine

import (
	"slices"
	"strings"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// SkillView is one skill as it should be read right now: the proficiency to use
// after conditions and applied shift cards, and how far it moved from the
// character's own value.
type SkillView struct {
	// Base is the proficiency id stored on the character, before conditions.
	Base string

	// Effective is the proficiency id to roll with, after every applied
	// condition and shift card. It equals Base when no overlay touches this
	// skill.
	Effective string

	// Shift is the total requested shift, summed across conditions and shift
	// cards. It is the number shown to the player, and it is deliberately the
	// *requested* amount rather than the achieved one, so a skill sitting at the
	// bottom of the ladder still reads as "-3" and does not look like the
	// condition failed to apply. Clamped reports the discrepancy instead.
	Shift int

	// Clamped reports that the ladder ran out before the full shift could be
	// applied. Without this a third debuff on an already-bottomed-out skill
	// appears to do nothing, which reliably starts an argument at the table.
	Clamped bool
}

// Shifted reports whether conditions moved this skill at all.
func (s SkillView) Shifted() bool { return s.Shift != 0 }

// Down reports whether the skill got worse, which is what the sheet colours red.
func (s SkillView) Down() bool { return s.Shift < 0 }

// Up reports whether the skill got better, which is what the sheet colours green.
func (s SkillView) Up() bool { return s.Shift > 0 }

// ConditionShifts returns the total shift each skill key receives from the
// character's applied conditions. Only keys that actually move are present, so
// the caller can treat a missing key as "unaffected" without checking for zero.
//
// An applied condition naming an id that no longer exists in the config is
// skipped rather than treated as an error: a ruleset can drop a condition, and
// that must not make a saved character unopenable.
func ConditionShifts(cfg *config.Config, c model.Character) map[string]int {
	if cfg == nil || len(c.Conditions) == 0 {
		return nil
	}
	out := map[string]int{}
	for _, applied := range c.Conditions {
		cond, ok := cfg.ConditionByID(applied.ID)
		if !ok {
			continue
		}
		shift := cond.SkillShift(applied.Shift)
		if shift == 0 {
			continue
		}
		for _, key := range cond.AffectsSkills {
			out[key] += shift
		}
	}
	// A condition and its mirror can cancel out exactly (Frightened -2 with
	// Encouraged +2). Dropping the zero entries keeps such a skill from being
	// coloured as if something happened to it.
	for key, total := range out {
		if total == 0 {
			delete(out, key)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// SkillShifts returns the total shift each skill key receives from every play
// overlay: the applied conditions AND the hand-applied "Enact Shift" cards
// (Character.Shifts). This is the map every consumer of the effective reading
// (skill grid, vital cards, clamp warning) must use, so a shift card and a
// condition on the same skill behave as one summed effect.
//
// A shift card naming a key the config no longer defines is skipped rather than
// treated as an error - a ruleset can drop a skill, and that must not make a
// saved character unopenable. The exact-zero cleanup runs after both sources
// are summed, so a condition and a shift card cancelling each other leave the
// skill unmarked rather than coloured for nothing.
func SkillShifts(cfg *config.Config, c model.Character) map[string]int {
	out := ConditionShifts(cfg, c)
	if len(c.Shifts) == 0 {
		return out
	}
	if out == nil {
		out = map[string]int{}
	}
	for _, card := range c.Shifts {
		if card.SkillKey == "" || card.Shift == 0 {
			continue
		}
		if cfg != nil && !skillKeyConfigured(cfg, card.SkillKey) {
			continue
		}
		out[card.SkillKey] += card.Shift
	}
	for key, total := range out {
		if total == 0 {
			delete(out, key)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// skillKeyConfigured reports whether key names a "<group>.<skill>" pair the
// ruleset still defines. A card whose skill was dropped by a ruleset change
// must not colour anything, but the strip still renders it as removable.
func skillKeyConfigured(cfg *config.Config, key string) bool {
	if cfg == nil {
		return false
	}
	group, skill, ok := strings.Cut(key, ".")
	if !ok || skill == "" {
		return false
	}
	for _, g := range cfg.Skills.List() {
		if g.ID == group && slices.Contains(g.Skills, skill) {
			return true
		}
	}
	return false
}

// EffectiveSkills returns the post-condition view of every skill the character
// has stored, keyed the same way as Character.Skills. Skills no condition
// touches are still present, with Effective equal to Base and a zero Shift, so
// a template can render the whole grid from one map.
func EffectiveSkills(cfg *config.Config, c model.Character) map[string]SkillView {
	shifts := SkillShifts(cfg, c)
	out := make(map[string]SkillView, len(c.Skills))
	for key, base := range c.Skills {
		view := SkillView{Base: base, Effective: base}
		if shift, ok := shifts[key]; ok && cfg != nil {
			view.Shift = shift
			view.Effective = cfg.ShiftProficiency(base, shift)
			view.Clamped = cfg.ShiftClamped(base, shift)
		}
		out[key] = view
	}
	return out
}

// EffectiveSkill returns the post-condition view of a single skill. It is the
// convenience form for a template that already knows the group and skill name.
func EffectiveSkill(cfg *config.Config, c model.Character, group, skill string) SkillView {
	key := model.SkillKey(group, skill)
	base := ""
	if c.Skills != nil {
		base = c.Skills[key]
	}
	view := SkillView{Base: base, Effective: base}
	if cfg == nil {
		return view
	}
	shift := SkillShifts(cfg, c)[key]
	if shift == 0 {
		return view
	}
	view.Shift = shift
	view.Effective = cfg.ShiftProficiency(base, shift)
	view.Clamped = cfg.ShiftClamped(base, shift)
	return view
}
