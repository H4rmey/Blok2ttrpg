// This file defines the rules-domain schema types: dice ladders, validations,
// the proficiency ladder, level budget tables and conditions, together with the
// accessors that give each one its documented defaults.
package config

// Dice lists the die tiers available for damage and generic rolls.
type Dice struct {
	Damage  []string `yaml:"damage,omitempty" json:"damage,omitempty"`
	Generic []string `yaml:"generic,omitempty" json:"generic,omitempty"`
}

// Validations captures the engagement/counter configuration and its fields.
type Validations struct {
	// Information is optional section-level help text. By default it renders as
	// plain text between the Validation header and the first field; when
	// RenderInformation is true it collapses into a hover "i" badge next to the
	// header instead.
	Information       string  `yaml:"information,omitempty" json:"information,omitempty"`
	RenderInformation bool    `yaml:"render_information,omitempty" json:"render_information,omitempty"`
	Fields            []Field `yaml:"fields,omitempty" json:"fields,omitempty"`

	// DCValidation configures the flat-DC validation mode an enactment opts
	// into via use_dc_validation. It defines the DC field the builder shows
	// (label, default, bounds and per-step cost) instead of the counter-roll
	// list. Nil means no enactment can meaningfully opt in.
	DCValidation *DCValidation `yaml:"dc_validation,omitempty" json:"dc_validation,omitempty"`
}

// DCValidation is the configurable shape of flat-DC validation. An enactment
// with use_dc_validation renders one free_number field under this label and
// rolls its engage source against the chosen DC.
type DCValidation struct {
	// Label is the builder label for the DC field.
	Label string `yaml:"label,omitempty" json:"label,omitempty"`
	// DefaultDC is the DC a fresh enactment starts at.
	DefaultDC int `yaml:"default_dc,omitempty" json:"default_dc,omitempty"`
	// MinDC/MaxDC bound the DC the builder offers.
	MinDC int `yaml:"min_dc,omitempty" json:"min_dc,omitempty"`
	MaxDC int `yaml:"max_dc,omitempty" json:"max_dc,omitempty"`
	// PerStep prices each DC step above/below the default, mirroring the
	// per_step of a free_number field.
	PerStep *PerStep `yaml:"per_step,omitempty" json:"per_step,omitempty"`
}

// DCField builds the synthetic validation field for the flat DC value, so the
// builder, normalization and cost engine all share one definition.
func (v *Validations) DCField() *Field {
	d := v.DCValidation
	if d == nil {
		return nil
	}
	label := d.Label
	if label == "" {
		label = "Difficulty (DC)"
	}
	def, min, max := d.DefaultDC, d.MinDC, d.MaxDC
	if def == 0 {
		def = 2
	}
	if min == 0 {
		min = 2
	}
	if max == 0 {
		max = def
	}
	return &Field{
		Key:     "validation_dc",
		Label:   label,
		Type:    "free_number",
		Default: def,
		Min:     min,
		Max:     max,
		Step:    1,
		PerStep: d.PerStep,
	}
}

// DCValidationDC clamps a stored flat-DC value into the configured bounds,
// falling back to the configured default when nothing was stored.
func (c *Config) DCValidationDC(stored int) int {
	d := c.Validations.DCValidation
	def, min, max := 2, 2, 2
	if d != nil {
		if d.DefaultDC != 0 {
			def = d.DefaultDC
		}
		if d.MinDC != 0 {
			min = d.MinDC
		}
		if d.MaxDC != 0 {
			max = d.MaxDC
		}
	}
	if stored <= 0 {
		return def
	}
	if stored < min {
		return min
	}
	if max > min && stored > max {
		return max
	}
	return stored
}

// Proficiency is a single skill tier.
type Proficiency struct {
	ID   string `yaml:"id" json:"id"`
	Name string `yaml:"name" json:"name"`
	Cost int    `yaml:"cost" json:"cost"`
	Note string `yaml:"note,omitempty" json:"note,omitempty"`
	// Die is the fallback die used for every dice-backed skill group at this
	// tier. Per-group overrides in Dice take precedence when present, so a tier
	// only needs the verbose Dice map when a group differs from the rest.
	Die    string            `yaml:"die,omitempty" json:"die,omitempty"`
	Dice   map[string]string `yaml:"dice,omitempty" json:"dice,omitempty"`
	Vitals map[string]any    `yaml:"vitals,omitempty" json:"vitals,omitempty"`
}

// DieFor returns the die this tier grants for a skill group: the per-group
// override in Dice when present, otherwise the shared Die fallback.
func (p Proficiency) DieFor(group string) string {
	if p.Dice != nil {
		if d, ok := p.Dice[group]; ok && d != "" {
			return d
		}
	}
	return p.Die
}

// Leveling describes the point budgets available to a character by level.
type Leveling struct {
	MaxLevel    int        `yaml:"max_level,omitempty" json:"max_level,omitempty"`
	SkillPoints LevelTable `yaml:"skill_points,omitempty" json:"skill_points,omitempty"`
	PerkPoints  LevelTable `yaml:"perk_points,omitempty" json:"perk_points,omitempty"`

	// InvokePoints is the per-session invoke point pool. It uses its own table
	// type because it grows in steps every few levels rather than every level;
	// see InvokeTable in invoking.go.
	InvokePoints InvokeTable `yaml:"invoke_points,omitempty" json:"invoke_points,omitempty"`
}

// LevelTable holds the budget progression for one point pool. The budget is
// normally derived from the formula Start + PerLevel * (level - 1), which keeps
// the documented curve and the served numbers from drifting apart. Levels is an
// optional explicit override used for non-linear curves: when it contains a row
// for the requested level that row's Total wins.
type LevelTable struct {
	// Start is the budget granted at level 1.
	Start int `yaml:"start,omitempty" json:"start,omitempty"`
	// PerLevel is the budget added by each level after the first.
	PerLevel int `yaml:"per_level,omitempty" json:"per_level,omitempty"`
	// Levels optionally overrides the formula on a per-level basis.
	Levels []LevelEntry `yaml:"levels,omitempty" json:"levels,omitempty"`
}

// LevelEntry is one row in a level table.
type LevelEntry struct {
	Level        int `yaml:"level" json:"level"`
	PointsGained int `yaml:"points_gained" json:"points_gained"`
	Total        int `yaml:"total" json:"total"`
}

// GeneralCondition is a shiftable condition applied via the condition enactment.
type GeneralCondition struct {
	ID          string `yaml:"id" json:"id"`
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	MinShift    int    `yaml:"min_shift,omitempty" json:"min_shift,omitempty"`
	MaxShift    int    `yaml:"max_shift,omitempty" json:"max_shift,omitempty"`
	ShiftCost   Cost   `yaml:"shift_cost,omitempty" json:"shift_cost,omitempty"`
}

// SpecificCondition is a fixed-cost named condition.
type SpecificCondition struct {
	ID          string `yaml:"id" json:"id"`
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	BuildCost   int    `yaml:"build_cost,omitempty" json:"build_cost,omitempty"`
	EnergyCost  int    `yaml:"energy_cost,omitempty" json:"energy_cost,omitempty"`
}

// Condition is a unified condition entry. It is "shiftable" when it declares a
// non-empty shift range (min_shift/max_shift), in which case it pays ShiftCost
// per unit of applied shift; otherwise it is a fixed-cost condition paying
// BuildCost/EnergyCost. This single type replaces the former general/specific
// split.
type Condition struct {
	ID          string `yaml:"id" json:"id"`
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Fixed-cost fields (non-shiftable conditions).
	BuildCost  int `yaml:"build_cost,omitempty" json:"build_cost,omitempty"`
	EnergyCost int `yaml:"energy_cost,omitempty" json:"energy_cost,omitempty"`

	// Shiftable fields. When MinShift or MaxShift is non-zero the condition is
	// treated as shiftable and ShiftCost is charged per unit of shift.
	MinShift  int  `yaml:"min_shift,omitempty" json:"min_shift,omitempty"`
	MaxShift  int  `yaml:"max_shift,omitempty" json:"max_shift,omitempty"`
	ShiftCost Cost `yaml:"shift_cost,omitempty" json:"shift_cost,omitempty"`

	// AffectsSkills lists the skills this condition moves, as "<group>.<skill>"
	// keys matching the character's skill map. A shiftable condition moves every
	// listed skill by the shift the player picked; a fixed condition moves them
	// by FixedShift. An empty list means the condition changes no skill numbers
	// at all - it only changes what the character may do - which is what leaves
	// its row on the character sheet without a value picker.
	AffectsSkills []string `yaml:"affects_skills,omitempty" json:"affects_skills,omitempty"`

	// FixedShift is the shift applied to every skill in AffectsSkills by a
	// non-shiftable condition. It exists so a condition whose own text already
	// names its magnitude ("movement speed shifted one down") can colour the
	// sheet without turning into a player-chosen range. It is ignored for
	// shiftable conditions, which take their magnitude from the applied value.
	FixedShift int `yaml:"fixed_shift,omitempty" json:"fixed_shift,omitempty"`

	// Selectable controls whether the condition appears in the builder's
	// condition dropdown. Some conditions are states the rules impose (Dying),
	// gear states, or GM-only effects, and must not be purchasable as an
	// enactment. It is a pointer so an unset value defaults to true, leaving
	// existing profiles unchanged. Non-selectable conditions are still
	// resolvable by id, so saved perks and generated instructions keep
	// working.
	Selectable *bool `yaml:"selectable,omitempty" json:"selectable,omitempty"`
}

// IsSelectable reports whether the condition may be picked in the builder.
// Defaults to true when unset.
func (c Condition) IsSelectable() bool {
	return c.Selectable == nil || *c.Selectable
}

// Shiftable reports whether the condition applies a skill shift (and therefore
// pays a per-shift cost) rather than a flat build/energy cost.
func (c Condition) Shiftable() bool {
	return c.MinShift != 0 || c.MaxShift != 0
}

// ShiftsSkills reports whether applying this condition changes any skill
// number. It is false for a condition that only changes what a character may
// do (Silenced, Charmed), which is why such a condition needs no value picker.
func (c Condition) ShiftsSkills() bool {
	return len(c.AffectsSkills) > 0
}

// NeedsShiftValue reports whether the player must choose a magnitude when
// applying this condition. Only a shiftable condition that actually moves
// skills asks for one; every other condition applies at a magnitude the
// ruleset already decided.
func (c Condition) NeedsShiftValue() bool {
	return c.Shiftable() && c.ShiftsSkills()
}

// SkillShift returns the per-skill shift this condition applies when it was
// applied at the given value. A shiftable condition uses the chosen value; a
// fixed condition ignores it and uses its configured FixedShift.
func (c Condition) SkillShift(applied int) int {
	if !c.ShiftsSkills() {
		return 0
	}
	if c.Shiftable() {
		return applied
	}
	return c.FixedShift
}
