package engine

import (
	"strconv"
	"strings"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// defaultVitalGroupID is the fallback skill group id whose skills (HP,
// Movement, Energy, ...) map to numeric vital values rather than dice, used
// when the config does not set vital_group.
const defaultVitalGroupID = "vital"

// VitalGroupID returns the configured vital skill group id, or the default.
func VitalGroupID(cfg *config.Config) string {
	if cfg != nil && cfg.VitalGroup != "" {
		return cfg.VitalGroup
	}
	return defaultVitalGroupID
}

// VitalStat is a computed vital value for a character. Max is the value granted
// by the selected proficiency tier for that vital skill. For editable vitals
// (HP and Energy) Current holds the character's current value, which may be
// below Max; for non-editable vitals (e.g. Movement) Current equals Max.
type VitalStat struct {
	Skill    string // display name, e.g. "HP"
	Key      string // lowercase key into the proficiency vitals map, e.g. "hp"
	Max      string // proficiency-granted value after conditions, formatted
	Current  string // current value (== Max when not editable)
	Editable bool   // whether Current can be edited independently of Max

	// Base is the value the character's own proficiency grants, before any
	// condition. It differs from Max only while a condition is shifting this
	// vital, which is what lets the card show "was 3" rather than silently
	// replacing the number the player bought.
	Base string

	// Shift is the total condition shift on this vital, and Clamped reports that
	// the proficiency ladder ran out before the whole shift could be applied.
	// Both mirror the SkillView fields of the same name so the vital card and
	// the skill row can be coloured by the same rule.
	Shift   int
	Clamped bool
}

// Shifted reports whether a condition is currently moving this vital.
func (v VitalStat) Shifted() bool { return v.Shift != 0 }

// Down reports whether the vital got worse, which the card colours red.
func (v VitalStat) Down() bool { return v.Shift < 0 }

// Up reports whether the vital got better, which the card colours green.
func (v VitalStat) Up() bool { return v.Shift > 0 }

// editableVitals lists the vital keys that carry a separate current value.
var editableVitals = map[string]bool{"hp": true, "energy": true}

// CharacterVitals returns the computed vital stats for a character, in config
// order. The Max of each vital comes from the proficiency tier the character
// selected for that vital skill, after any condition shifting that vital: a
// character whose Movement is Slowed reads the slowed number at the top of the
// sheet, because the number on the card is the one the player uses in play and
// it must not disagree with the skill row further down.
//
// The current value of an editable vital is read from the character trait
// "current_<key>"; if unset it defaults to Max. Conditions are not permitted to
// shift HP or Energy (the loader rejects it), so an editable vital never has a
// shifted Max and its stored current value is never second-guessed here. A
// hand-applied shift card MAY move a vital during play: only the derived Max
// moves, and the player's own stored current count is never altered.
func CharacterVitals(cfg *config.Config, c model.Character) []VitalStat {
	var out []VitalStat
	vg := VitalGroupID(cfg)
	skills, ok := cfg.Skills.Items[vg]
	if !ok {
		return out
	}
	shifts := SkillShifts(cfg, c)
	for _, skill := range skills {
		key := strings.ToLower(skill)
		baseProf := c.Skills[model.SkillKey(vg, skill)]

		shift := shifts[model.SkillKey(vg, skill)]
		effProf := baseProf
		clamped := false
		if shift != 0 && cfg != nil {
			effProf = cfg.ShiftProficiency(baseProf, shift)
			clamped = cfg.ShiftClamped(baseProf, shift)
		}

		base := vitalValueFor(cfg, baseProf, key)
		max := base
		if effProf != baseProf {
			max = vitalValueFor(cfg, effProf, key)
		}

		editable := editableVitals[key]
		current := max
		if editable {
			if v := c.Attr("current_" + key); v != nil {
				if s := asString(v); s != "" {
					current = s
				}
			}
		}
		out = append(out, VitalStat{
			Skill:    skill,
			Key:      key,
			Max:      max,
			Current:  current,
			Editable: editable,
			Base:     base,
			Shift:    shift,
			Clamped:  clamped,
		})
	}
	return out
}

// vitalValueFor reads the numeric vital a proficiency tier grants for a vital
// key, formatted, or "" when either the tier or the key is unknown.
func vitalValueFor(cfg *config.Config, profID, key string) string {
	if cfg == nil {
		return ""
	}
	p, ok := cfg.Proficiency(profID)
	if !ok {
		return ""
	}
	v, ok := p.Vitals[key]
	if !ok {
		return ""
	}
	return formatVital(v)
}

// formatVital renders a vital value (int or float) without a trailing ".0".
func formatVital(v any) string {
	switch t := v.(type) {
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case int:
		return strconv.Itoa(t)
	default:
		return asString(v)
	}
}
