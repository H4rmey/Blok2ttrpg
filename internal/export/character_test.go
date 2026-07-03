package export

import (
	"strings"
	"testing"

	"github.com/harmey/blok2ttrpg/ability-builder/internal/models"
)

func sampleCharacter() *models.Character {
	c := models.NewCharacter("test-id", []string{"Strength", "Magic"}, []string{"Power"}, []string{"Constitution"}, )
	c.Name = "Aelara the Bold"
	c.Level = 3
	c.Age = "27"
	c.Size = "Medium"
	c.Alignment = "Chaotic Good"
	c.Backstory = "Raised by wolves."
	c.VitalHP = models.ProfTrained
	c.VitalMovement = models.ProfUntrained
	c.VitalEnergy = models.ProfExpert
	c.CurrentHP = 16
	c.CurrentEnergy = 6
	c.GeneralTraits["Strength"] = models.ProfExpert
	c.GeneralTraits["Magic"] = models.ProfTrained
	c.OffenseTraits["Power"] = models.ProfTrained
	c.DefenseTraits["Constitution"] = models.ProfExpert

	ability := models.Ability{
		ID:          "ability-1",
		Name:        "Fire Bolt",
		Description: "A flammable spark ignites.",
		Type:        models.AbilityExecution,
		EnergyCost:  3,
		ActionCost:  2,
		EnergySteps: 0,
		ActionSteps: 0,
		Enactments: []models.Enactment{
			{
				Type:    models.EnactDamage,
				Source:  "d10",
				BuildCost: 1,
				Interaction: &models.Interaction{
					Type:   models.InteractionRanged,
					Range:  30,
					BuildCost: 1,
				},
			},
		},
	}
	c.Abilities = append(c.Abilities, ability)

	ability2 := models.Ability{
		ID:          "ability-2",
		Name:        "Healing Word",
		Type:        models.AbilityExecution,
		EnergyCost:  3,
		ActionCost:  2,
		Enactments: []models.Enactment{
			{
				Type:    models.EnactHealing,
				Source:  "d8",
				Interaction: &models.Interaction{
					Type: models.InteractionDirect,
					Range: 5,
				},
			},
		},
	}
	c.Abilities = append(c.Abilities, ability2)

	return &c
}

func TestCharacterRoundTrip(t *testing.T) {
	c := sampleCharacter()
	yaml := CharacterToYAML(c)
	if !strings.Contains(yaml, "character:") {
		t.Fatalf("missing character root key in:\n%s", yaml)
	}
	if !strings.Contains(yaml, "Fire Bolt") {
		t.Fatalf("missing ability name in:\n%s", yaml)
	}

	parsed, err := ParseCharacterYAML([]byte(yaml))
	if err != nil {
		t.Fatalf("parse error: %v\nyaml:\n%s", err, yaml)
	}
	if parsed.Name != c.Name {
		t.Errorf("name: got %q want %q", parsed.Name, c.Name)
	}
	if parsed.Level != c.Level {
		t.Errorf("level: got %d want %d", parsed.Level, c.Level)
	}
	if parsed.CurrentHP != c.CurrentHP {
		t.Errorf("current_hp: got %d want %d", parsed.CurrentHP, c.CurrentHP)
	}
	if len(parsed.Abilities) != 2 {
		t.Fatalf("expected 2 abilities, got %d", len(parsed.Abilities))
	}
	if parsed.Abilities[0].Name != "Fire Bolt" {
		t.Errorf("ability 0 name: %q", parsed.Abilities[0].Name)
	}
	if parsed.Abilities[0].Type != models.AbilityExecution {
		t.Errorf("ability 0 type: %s", parsed.Abilities[0].Type)
	}
	if len(parsed.Abilities[0].Enactments) != 1 {
		t.Fatalf("expected 1 enactment in ability 0")
	}
	e := parsed.Abilities[0].Enactments[0]
	if e.Type != models.EnactDamage {
		t.Errorf("enactment type: %s", e.Type)
	}
	if e.Source != "d10" {
		t.Errorf("source: %q", e.Source)
	}
	if e.Interaction == nil || e.Interaction.Type != models.InteractionRanged {
		t.Errorf("interaction missing/wrong: %+v", e.Interaction)
	} else if e.Interaction.Range != 30 {
		t.Errorf("interaction range: %d", e.Interaction.Range)
	}
	if parsed.GeneralTraits["Strength"] != models.ProfExpert {
		t.Errorf("general trait strength: %s", parsed.GeneralTraits["Strength"])
	}

	// Re-export and verify it round-trips stably.
	yaml2 := CharacterToYAML(parsed)
	if yaml != yaml2 {
		t.Errorf("re-exported YAML differs from original\n--- orig ---\n%s\n--- new ---\n%s", yaml, yaml2)
	}
}

func TestParseSingleAbility(t *testing.T) {
	yaml := `ability:
  type: Execution
  name: Fire Bolt
  description: A flammable spark ignites.
  enactments:
    - type: Enact Damage
      source: d10
      interactions:
        - type: Ranged
          range: 30m
`
	c, err := ParseCharacterYAML([]byte(yaml))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if c.Name != "Fire Bolt" {
		t.Errorf("expected name derived from ability, got %q", c.Name)
	}
	if len(c.Abilities) != 1 {
		t.Fatalf("expected 1 ability, got %d", len(c.Abilities))
	}
	if c.Abilities[0].Type != models.AbilityExecution {
		t.Errorf("type: %s", c.Abilities[0].Type)
	}
}
