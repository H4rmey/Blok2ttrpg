// Package export handles converting characters and perks to and from the
// portable YAML representation, plus building print-friendly HTML for PDF.
package export

import (
	"fmt"

	"github.com/harmey/blok2ttrpg-v5/internal/model"
	"gopkg.in/yaml.v3"
)

// CharacterYAML is the portable, human-friendly YAML shape of a character. It
// intentionally mirrors the generic model so any config's traits survive a
// round trip without code changes.
type CharacterYAML struct {
	ID       string                   `yaml:"id,omitempty"`
	Level    int                      `yaml:"level"`
	Traits   map[string]any           `yaml:"traits,omitempty"`
	Skills   map[string]string        `yaml:"skills,omitempty"`
	Perks    []model.Perk             `yaml:"perks,omitempty"`
	Packages []model.InstalledPackage `yaml:"packages,omitempty"`
}

// MarshalCharacter serializes a character to YAML bytes.
func MarshalCharacter(c model.Character) ([]byte, error) {
	out := CharacterYAML{
		ID:       c.ID,
		Level:    c.Level,
		Traits:   c.Traits,
		Skills:   c.Skills,
		Perks:    c.Perks,
		Packages: c.Packages,
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
		ID:       in.ID,
		Level:    in.Level,
		Traits:   in.Traits,
		Skills:   in.Skills,
		Perks:    in.Perks,
		Packages: in.Packages,
	}
	if c.Level < 1 {
		c.Level = 1
	}
	if c.Traits == nil {
		c.Traits = map[string]any{}
	}
	if c.Skills == nil {
		c.Skills = map[string]string{}
	}
	return c, nil
}

// MarshalPerk serializes a single perk to YAML.
func MarshalPerk(a model.Perk) ([]byte, error) {
	return yaml.Marshal(a)
}

// UnmarshalPerk parses YAML bytes into a single perk. The id may be
// overridden by the caller after import.
func UnmarshalPerk(data []byte) (model.Perk, error) {
	var a model.Perk
	if err := yaml.Unmarshal(data, &a); err != nil {
		return model.Perk{}, fmt.Errorf("parsing perk yaml: %w", err)
	}
	return a, nil
}
