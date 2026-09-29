package config

import "fmt"

// Validate performs light structural checks. The philosophy is "config leads":
// we verify referential integrity and required identity, but we do not impose
// game rules. Cost is advisory and never validated here.
func (c *Config) Validate() error {
	if c.Version == 0 {
		return fmt.Errorf("version is required")
	}
	if c.ProfileID == "" {
		return fmt.Errorf("profile_id is required")
	}
	if !profileIDPattern.MatchString(c.ProfileID) {
		return fmt.Errorf("profile_id %q must be lowercase letters, numbers, _ or -", c.ProfileID)
	}
	if len(c.PerkTypes.Order) == 0 {
		return fmt.Errorf("at least one perk type is required")
	}

	check := func(kind string, m ComponentMap) error {
		for _, comp := range m.List() {
			if comp.ID == "" {
				return fmt.Errorf("%s: component id is required", kind)
			}
			if err := validateFields(comp.ID, comp.Fields); err != nil {
				return fmt.Errorf("%s %q: %w", kind, comp.ID, err)
			}
		}
		return nil
	}
	if err := check("perk_type", c.PerkTypes); err != nil {
		return err
	}
	if err := check("enactment", c.Enactments); err != nil {
		return err
	}
	if err := check("interaction", c.Interactions); err != nil {
		return err
	}

	for _, g := range c.Traits.List() {
		if err := validateFields("traits."+g.ID, g.Fields); err != nil {
			return err
		}
	}
	if err := validateFields("validations", c.Validations.Fields); err != nil {
		return err
	}
	if err := c.validateConditions(); err != nil {
		return err
	}
	return nil
}

// validateConditions checks the referential integrity of every condition's
// affects_skills list. A typo there would silently produce a condition that
// colours nothing on the character sheet and shifts nothing, which is the kind
// of failure a player only notices mid-session, so it is a load error instead.
//
// The vital group is restricted to a single allowed skill (movement): letting a
// condition shift HP or Energy would change a character's maximum pool as a
// side effect of a temporary state, which is a different rule than "your rolls
// get worse" and is not what conditions are for.
func (c *Config) validateConditions() error {
	if len(c.Conditions) == 0 {
		return nil
	}
	known := map[string]bool{}
	for _, g := range c.Skills.List() {
		for _, s := range g.Skills {
			known[g.ID+"."+s] = true
		}
	}
	vitalGroup := c.VitalGroup
	if vitalGroup == "" {
		vitalGroup = "vital"
	}
	for _, cond := range c.Conditions {
		for _, key := range cond.AffectsSkills {
			if !known[key] {
				return fmt.Errorf("condition %q: affects_skills references unknown skill %q "+
					"(expected \"<group>.<skill>\" from the skills list)", cond.ID, key)
			}
			group, skill := splitSkillKey(key)
			if group == vitalGroup && !allowedVitalShift[skill] {
				return fmt.Errorf("condition %q: affects_skills may not shift the vital %q; "+
					"conditions only shift %q in the vital group", cond.ID, skill, vitalMovementSkill)
			}
		}
		// A fixed condition that names skills but no magnitude would render a
		// coloured, unshifted skill, which reads as a display bug.
		if cond.ShiftsSkills() && !cond.Shiftable() && cond.FixedShift == 0 {
			return fmt.Errorf("condition %q: affects_skills is set but the condition is neither "+
				"shiftable (min_shift/max_shift) nor has a fixed_shift, so it would shift nothing", cond.ID)
		}
	}
	return nil
}

// vitalMovementSkill is the one vital a condition is permitted to shift. HP and
// Energy are pools a character buys, not numbers a temporary state moves.
const vitalMovementSkill = "Movement"

var allowedVitalShift = map[string]bool{vitalMovementSkill: true}

// splitSkillKey splits a "<group>.<skill>" key. A key without a separator is
// treated as having no group, which the caller reports as unknown anyway.
func splitSkillKey(key string) (group, skill string) {
	for i := 0; i < len(key); i++ {
		if key[i] == '.' {
			return key[:i], key[i+1:]
		}
	}
	return "", key
}

var validFieldTypes = map[string]bool{
	"checkbox":         true,
	"dropdown":         true,
	"free_text":        true,
	"free_number":      true,
	"multiselect":      true,
	"conditions":       true,
	"condition_select": true,
}

func validateFields(scope string, fields []Field) error {
	seen := map[string]bool{}
	for _, f := range fields {
		if f.Key == "" {
			return fmt.Errorf("%s: field key is required", scope)
		}
		if seen[f.Key] {
			return fmt.Errorf("%s: duplicate field key %q", scope, f.Key)
		}
		seen[f.Key] = true
		if !validFieldTypes[f.Type] {
			return fmt.Errorf("%s: field %q has unknown type %q", scope, f.Key, f.Type)
		}
		if len(f.Options) > 0 && f.OptionsSource != "" {
			return fmt.Errorf("%s: field %q mixes options and options_source", scope, f.Key)
		}
		if f.Type == "multiselect" || f.Type == "conditions" {

			if len(f.RowFields) == 0 {
				return fmt.Errorf("%s: %s field %q requires row_fields", scope, f.Type, f.Key)
			}
			if err := validateFields(scope+"."+f.Key, f.RowFields); err != nil {
				return err
			}
		}
		// Nested option fields (an option that reveals child fields).
		for _, opt := range f.Options {
			if len(opt.Fields) > 0 {
				if err := validateFields(scope+"."+f.Key+"."+opt.Value, opt.Fields); err != nil {
					return err
				}
			}
		}
		if f.Type == "free_number" && f.Min > f.Max && f.Max != 0 {
			return fmt.Errorf("%s: field %q min > max", scope, f.Key)
		}
	}
	return nil
}
