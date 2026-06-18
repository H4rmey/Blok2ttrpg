package models

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// MarshalYAML implements yaml.Marshaler for Proficiency.
func (p Proficiency) MarshalYAML() (interface{}, error) {
	return p.String(), nil
}

// UnmarshalYAML implements yaml.Unmarshaler for Proficiency.
func (p *Proficiency) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	*p = ProficiencyFromString(s)
	if s != "" && p.String() != s {
		return fmt.Errorf("unknown proficiency level: %q", s)
	}
	return nil
}

// MarshalCharacter serializes a Character to YAML bytes.
func MarshalCharacter(c *Character) ([]byte, error) {
	return yaml.Marshal(c)
}

// UnmarshalCharacter deserializes YAML bytes into a Character.
func UnmarshalCharacter(data []byte) (*Character, error) {
	var c Character
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// MarshalAbility serializes a single Ability to YAML bytes.
func MarshalAbility(a *Ability) ([]byte, error) {
	return yaml.Marshal(a)
}

// UnmarshalAbility deserializes YAML bytes into an Ability.
func UnmarshalAbility(data []byte) (*Ability, error) {
	var a Ability
	if err := yaml.Unmarshal(data, &a); err != nil {
		return nil, err
	}
	return &a, nil
}
