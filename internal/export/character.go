// Package export handles converting characters and abilities to and from the
// portable YAML representation, plus building print-friendly HTML for PDF.
package export

import (
	"fmt"

	"github.com/harmey/blok2ttrpg-v5/internal/model"
	"gopkg.in/yaml.v3"
)

// CharacterYAML is the portable, human-friendly YAML shape of a character. It
// intentionally mirrors the generic model so any config's attributes survive a
// round trip without code changes.
type CharacterYAML struct {
	ID         string            `yaml:"id,omitempty"`
	Level      int               `yaml:"level"`
	Attributes map[string]any    `yaml:"attributes,omitempty"`
	Traits     map[string]string `yaml:"traits,omitempty"`
	Abilities  []model.Ability   `yaml:"abilities,omitempty"`
}

// MarshalCharacter serializes a character to YAML bytes.
func MarshalCharacter(c model.Character) ([]byte, error) {
	out := CharacterYAML{
		ID:         c.ID,
		Level:      c.Level,
		Attributes: c.Attributes,
		Traits:     c.Traits,
		Abilities:  c.Abilities,
	}
	return yaml.Marshal(out)
}

// UnmarshalCharacter parses YAML bytes into a character. The id may be
// overridden by the caller after import.
func UnmarshalCharacter(data []byte) (model.Character, error) {
	var in CharacterYAML
	if err := yaml.Unmarshal(data, &in); err != nil {
		return model.Character{}, fmt.Errorf("parsing character yaml: %w", err)
	}
	c := model.Character{
		ID:         in.ID,
		Level:      in.Level,
		Attributes: in.Attributes,
		Traits:     in.Traits,
		Abilities:  in.Abilities,
	}
	if c.Level < 1 {
		c.Level = 1
	}
	if c.Attributes == nil {
		c.Attributes = map[string]any{}
	}
	if c.Traits == nil {
		c.Traits = map[string]string{}
	}
	return c, nil
}

// MarshalAbility serializes a single ability to YAML.
func MarshalAbility(a model.Ability) ([]byte, error) {
	return yaml.Marshal(a)
}

// UnmarshalAbility parses YAML bytes into a single ability. The id may be
// overridden by the caller after import.
func UnmarshalAbility(data []byte) (model.Ability, error) {
	var a model.Ability
	if err := yaml.Unmarshal(data, &a); err != nil {
		return model.Ability{}, fmt.Errorf("parsing ability yaml: %w", err)
	}
	return a, nil
}
