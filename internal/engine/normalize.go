package engine

import (
	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// Normalization gives the app a single canonical representation of an ability.
//
// Cost used to be derived from two representations that resolved defaults
// independently: the stored ability (where an absent field meant "nothing
// selected") and the builder form (where the template rendered the field's
// configured default). The same perk therefore had two different prices
// depending on which one you looked at, and merely opening a perk in the
// builder could change its cost.
//
// NormalizeAbility closes that gap by resolving every default once, on the way
// in. After normalization a stored ability is complete and in range, so the
// cost engine and the templates can both read values literally and agree by
// construction.

// NormalizeAbility returns a copy of the ability with every configured field
// present, repeatable fields expanded to their default rows, and numeric fields
// clamped into their configured range. It is idempotent: normalizing an already
// normalized ability changes nothing.
func NormalizeAbility(cfg *config.Config, a model.Ability) model.Ability {
	if cfg == nil {
		return a
	}
	out := a
	if at, ok := cfg.AbilityType(a.Type); ok {
		out.Fields = normalizeFields(cfg, at.Fields, a.Fields)
	}

	if len(a.Enactments) > 0 {
		ens := make([]model.Enactment, 0, len(a.Enactments))
		for _, en := range a.Enactments {
			ens = append(ens, normalizeEnactment(cfg, en))
		}
		out.Enactments = ens
	}
	return out
}

// normalizeEnactment normalizes an enactment's own fields plus its interaction
// and validation data.
func normalizeEnactment(cfg *config.Config, en model.Enactment) model.Enactment {
	out := en
	if ec, ok := cfg.Enactment(en.Type); ok {
		out.Fields = normalizeFields(cfg, ec.Fields, en.Fields)
	}
	if ic, ok := cfg.Interaction(en.Interaction); ok {
		out.InteractionData = normalizeFields(cfg, ic.Fields, en.InteractionData)
	}
	// Validation data is only normalized when the enactment already carries
	// some. An enactment with no validation at all is left alone rather than
	// having a full set of defaults invented for it, which would add cost the
	// perk never had.
	if len(en.ValidationData) > 0 && len(cfg.Validations.Fields) > 0 {
		out.ValidationData = normalizeFields(cfg, cfg.Validations.Fields, en.ValidationData)
	}
	return out
}

// normalizeFields resolves a single set of field values against its definitions.
func normalizeFields(cfg *config.Config, fields []config.Field, values map[string]any) map[string]any {
	out := map[string]any{}
	// Values not backed by a field definition (e.g. an author's "comment", or a
	// sibling shift key) are carried through untouched.
	for k, v := range values {
		out[k] = v
	}
	for _, f := range fields {
		raw, present := values[f.Key]
		switch f.Type {
		case "free_number":
			n := asInt(raw)
			if !present {
				n = asInt(f.Default)
			}
			out[f.Key] = clampNumber(f, n)
		case "multiselect", "conditions":
			out[f.Key] = normalizeRows(cfg, f, raw, present)
		case "dropdown":
			if !present {
				out[f.Key] = asString(f.Default)
			}
			// An inline_builder dropdown carries a nested component builder;
			// recurse so its nested values are normalized too.
			if f.InlineBuilder != nil {
				if val := asString(out[f.Key]); val != "" {
					if comp, ok := cfg.ComponentByKind(f.InlineBuilder.Kind, val); ok {
						nested, _ := out[f.Key+"_ib"].(map[string]any)
						out[f.Key+"_ib"] = normalizeFields(cfg, comp.Fields, nested)
					}
				}
			}
		case "checkbox":
			out[f.Key] = asBool(raw) || (!present && asBool(f.Default))
		default:
			if !present {
				out[f.Key] = asString(f.Default)
			}
		}
	}
	return out
}

// normalizeRows expands a repeatable field to its stored rows, or to the
// configured default rows when the field carries none. Each row's own sub
// fields are normalized in turn.
func normalizeRows(cfg *config.Config, f config.Field, raw any, present bool) []map[string]any {
	rows := asRows(raw)
	if !present && len(rows) == 0 {
		rows = defaultRows(f)
	}
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, normalizeFields(cfg, f.RowFields, row))
	}
	return out
}

// defaultRows builds the initial rows a repeatable field starts with, from
// row_defaults and default_count. This mirrors what the builder renders for a
// field the ability does not yet carry.
func defaultRows(f config.Field) []map[string]any {
	n := f.DefaultCount
	if n < len(f.RowDefaults) {
		n = len(f.RowDefaults)
	}
	rows := make([]map[string]any, 0, n)
	for i := 0; i < n; i++ {
		row := map[string]any{}
		for _, rf := range f.RowFields {
			val := ""
			if i < len(f.RowDefaults) {
				val = f.RowDefaults[i][rf.Key]
			}
			if val == "" {
				val = asString(rf.Default)
			}
			row[rf.Key] = val
		}
		rows = append(rows, row)
	}
	return rows
}

// clampNumber constrains a free_number value to the field's configured range.
// The range is only applied when max is above min, so fields that leave the
// bounds unset are untouched. A config may declare a default outside its own
// min/max; clamping here means the canonical value is always one the builder
// can actually represent.
func clampNumber(f config.Field, n int) int {
	if f.Max <= f.Min {
		return n
	}
	if n < f.Min {
		return f.Min
	}
	if n > f.Max {
		return f.Max
	}
	return n
}

// NormalizeCharacter normalizes every perk a character owns, reporting how many
// perks actually changed. The count drives the user-facing feedback of the
// refresh controls, which exist to repair abilities stored before normalization
// (or by an older version of the config).
func NormalizeCharacter(cfg *config.Config, c *model.Character) int {
	changed := 0
	for i, ab := range c.Abilities {
		before := AbilityCost(cfg, ab)
		norm := NormalizeAbility(cfg, ab)
		if AbilityCost(cfg, norm) != before {
			changed++
		}
		c.Abilities[i] = norm
	}
	return changed
}
