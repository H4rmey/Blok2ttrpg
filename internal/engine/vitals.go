package engine

import (
	"strconv"
	"strings"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// defaultVitalGroupID is the fallback trait group id whose traits (HP,
// Movement, Energy, ...) map to numeric vital values rather than dice, used
// when the config does not set vital_group.
const defaultVitalGroupID = "vital"

// VitalGroupID returns the configured vital trait group id, or the default.
func VitalGroupID(cfg *config.Config) string {
	if cfg != nil && cfg.VitalGroup != "" {
		return cfg.VitalGroup
	}
	return defaultVitalGroupID
}

// VitalStat is a computed vital value for a character. Max is the value granted
// by the selected proficiency tier for that vital trait. For editable vitals
// (HP and Energy) Current holds the character's current value, which may be
// below Max; for non-editable vitals (e.g. Movement) Current equals Max.
type VitalStat struct {
	Trait    string // display name, e.g. "HP"
	Key      string // lowercase key into the proficiency vitals map, e.g. "hp"
	Max      string // proficiency-granted value, formatted
	Current  string // current value (== Max when not editable)
	Editable bool   // whether Current can be edited independently of Max
}

// editableVitals lists the vital keys that carry a separate current value.
var editableVitals = map[string]bool{"hp": true, "energy": true}

// CharacterVitals returns the computed vital stats for a character, in config
// order. The Max of each vital comes from the proficiency tier the character
// selected for that vital trait. The current value of an editable vital is read
// from the character attribute "current_<key>"; if unset it defaults to Max.
func CharacterVitals(cfg *config.Config, c model.Character) []VitalStat {
	var out []VitalStat
	vg := VitalGroupID(cfg)
	traits, ok := cfg.Traits.Items[vg]
	if !ok {
		return out
	}
	for _, trait := range traits {
		key := strings.ToLower(trait)
		profID := c.Traits[model.TraitKey(vg, trait)]

		max := ""
		if p, ok := cfg.Proficiency(profID); ok {
			if v, ok := p.Vitals[key]; ok {
				max = formatVital(v)
			}
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
			Trait:    trait,
			Key:      key,
			Max:      max,
			Current:  current,
			Editable: editable,
		})
	}
	return out
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
