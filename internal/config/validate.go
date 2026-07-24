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
	if len(c.AbilityTypes.Order) == 0 {
		return fmt.Errorf("at least one ability type is required")
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
	if err := check("ability_type", c.AbilityTypes); err != nil {
		return err
	}
	if err := check("enactment", c.Enactments); err != nil {
		return err
	}
	if err := check("interaction", c.Interactions); err != nil {
		return err
	}

	for _, g := range c.Attributes.List() {
		if err := validateFields("attributes."+g.ID, g.Fields); err != nil {
			return err
		}
	}
	if err := validateFields("validations", c.Validations.Fields); err != nil {
		return err
	}
	return nil
}

var validFieldTypes = map[string]bool{
	"checkbox":     true,
	"dropdown":     true,
	"free_text":    true,
	"free_number":  true,
	"solutions":    true,
	"states":       true,
	"state_select": true,
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
		if f.Type == "solutions" || f.Type == "states" {
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
