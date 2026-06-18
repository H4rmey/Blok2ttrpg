package models

import (
	"testing"
)

func TestProficiencyYAMLRoundTrip(t *testing.T) {
	for _, p := range AllProficiencies() {
		data, err := MarshalCharacter(&Character{
			Version: "1.0",
			Level:   1,
			GeneralTraits: GeneralTraits{
				Strength: p,
			},
			Settings: CampaignSettings{TraitCount: 22},
		})
		if err != nil {
			t.Fatalf("MarshalCharacter failed for %s: %v", p, err)
		}

		got, err := UnmarshalCharacter(data)
		if err != nil {
			t.Fatalf("UnmarshalCharacter failed for %s: %v", p, err)
		}
		if got.GeneralTraits.Strength != p {
			t.Errorf("Round-trip failed: got %s, want %s", got.GeneralTraits.Strength, p)
		}
	}
}

func TestFullCharacterYAMLRoundTrip(t *testing.T) {
	original := &Character{
		Version: "1.0",
		Level:   5,
		Attributes: Attributes{
			Name:       "Michael",
			Age:        "2 months",
			Size:       "1.2m",
			Alignment:  "Rebel Chaotic",
			Backstory:  "Forced to work for the Ylten Guild",
			Personality: "Overly Positive and Impulsive",
			Traits:     "Walking Database of Knowledge",
			Appearance: "Wooden Puppet",
			Hobbies:    "Likes Making Furniture of Wood",
			Occupation: "Adventurer",
			Inventory:  "Hidden Compartments in His Body",
			Quirks:     "Likes Bullying Insecure People",
		},
		TemporaryAttributes: []TemporaryAttribute{
			{Name: "Lost His Left Arm", Description: "Must Rebuild It or Find It Back"},
		},
		GeneralTraits: GeneralTraits{
			Strength:    Trained,
			Dexterity:   Expert,
			Stealth:     Untrained,
			Perception:  Trained,
			Nature:      Clumsy,
			Crafting:    Master,
			PeopleSkill: Untrained,
			Performance: Clumsy,
			Thievery:    Clumsy,
			Knowledge:   Expert,
			Magic:       Clumsy,
		},
		CombativeTraits: CombativeTraits{
			Offense: OffenseTraits{
				Precision: Expert,
				Power:     Trained,
				Mind:      Untrained,
				Magic:     Clumsy,
			},
			Defense: DefenseTraits{
				Reflex:       Trained,
				Constitution: Trained,
				Mind:         Untrained,
				Magic:        Clumsy,
			},
			Vital: VitalTraits{
				HP:       Trained,
				Movement: Untrained,
				Energy:   Trained,
			},
		},
		Abilities: []Ability{
			{
				Name:       "Power Shot",
				Type:       AbilityTypeExecution,
				EnergyCost: 3,
				ActionCost: 2,
				Enactments: []Enactment{
					{
						Type:                    EnactmentDamage,
						DamageDice:              "d8 + Power",
						BaseEnactmentEnergyCost: 0,
						Perks: []Perk{
							{Description: "Shift Dice Tier up", AddCost: 2, Amount: 2, TotalAddCost: 4, EnergyCost: 2, IsOptional: true},
							{Description: "Add Offensive Trait Dice", AddCost: 4, Amount: 1, TotalAddCost: 4, EnergyCost: 2, IsOptional: false},
						},
						Interactions: []Interaction{
							{
								Type:         InteractionRanged,
								Engager:      "Self",
								TargetAmount: 1,
								Range:        "14m",
								Visibility:   "Visible",
								Obstruction:  "Not obstructed",
								Perks: []Perk{
									{Description: "Increase reach with 2m", AddCost: 1, Amount: 2, TotalAddCost: 2, EnergyCost: 0},
								},
								Validation: &Validation{
									EngagementRoll: "Precision - 2",
									CounterRoll:    []string{"Reflex", "Constitution"},
								},
							},
						},
					},
				},
			},
			{
				Name:       "Healing Touch",
				Type:       AbilityTypeReaction,
				EnergyCost: 3,
				Range:      1,
				Uses:       1,
				Triggers:   []Trigger{{Name: "Someone runs towards you"}},
				Enactments: []Enactment{
					{
						Type:                    EnactmentHealing,
						HealingDice:             "d10 + Medicine + 4",
						BaseEnactmentEnergyCost: 0,
						Perks: []Perk{
							{Description: "Shift Dice Tier up", AddCost: 2, Amount: 3, TotalAddCost: 6, EnergyCost: 3},
							{Description: "Add Medicine Trait Dice", AddCost: 1, Amount: 1, TotalAddCost: 1, EnergyCost: 1},
							{Description: "Add flat +1 bonus", AddCost: 3, Amount: 4, TotalAddCost: 12, EnergyCost: 0},
						},
						Interactions: []Interaction{
							{
								Type:         InteractionDirect,
								Engager:      "Self",
								TargetAmount: 1,
								Range:        "1m",
								Validation: &Validation{
									EngagementRoll: "n/a",
									CounterRoll:    nil,
								},
							},
						},
					},
				},
			},
		},
		LevelHistory: []LevelSnapshot{
			{Level: 1, TraitPointsGained: 0, AbilityPointsGained: 0},
			{Level: 2, TraitPointsGained: 1, AbilityPointsGained: 2},
			{Level: 3, TraitPointsGained: 1, AbilityPointsGained: 3},
			{Level: 4, TraitPointsGained: 1, AbilityPointsGained: 2},
			{Level: 5, TraitPointsGained: 2, AbilityPointsGained: 4},
		},
		Settings: CampaignSettings{
			TraitCount: 22,
		},
	}

	// Marshal
	data, err := MarshalCharacter(original)
	if err != nil {
		t.Fatalf("MarshalCharacter failed: %v", err)
	}

	// Print for inspection
	t.Logf("YAML output:\n%s", string(data))

	// Unmarshal
	restored, err := UnmarshalCharacter(data)
	if err != nil {
		t.Fatalf("UnmarshalCharacter failed: %v", err)
	}

	// Verify key fields
	if restored.Version != original.Version {
		t.Errorf("Version: got %q, want %q", restored.Version, original.Version)
	}
	if restored.Level != original.Level {
		t.Errorf("Level: got %d, want %d", restored.Level, original.Level)
	}
	if restored.Attributes.Name != original.Attributes.Name {
		t.Errorf("Name: got %q, want %q", restored.Attributes.Name, original.Attributes.Name)
	}
	if restored.GeneralTraits.Crafting != original.GeneralTraits.Crafting {
		t.Errorf("Crafting: got %s, want %s", restored.GeneralTraits.Crafting, original.GeneralTraits.Crafting)
	}
	if restored.CombativeTraits.Offense.Precision != original.CombativeTraits.Offense.Precision {
		t.Errorf("Precision: got %s, want %s", restored.CombativeTraits.Offense.Precision, original.CombativeTraits.Offense.Precision)
	}
	if restored.CombativeTraits.Vital.HP != original.CombativeTraits.Vital.HP {
		t.Errorf("HP: got %s, want %s", restored.CombativeTraits.Vital.HP, original.CombativeTraits.Vital.HP)
	}
	if len(restored.Abilities) != len(original.Abilities) {
		t.Fatalf("Abilities count: got %d, want %d", len(restored.Abilities), len(original.Abilities))
	}
	if restored.Abilities[0].Name != "Power Shot" {
		t.Errorf("Ability[0].Name: got %q, want %q", restored.Abilities[0].Name, "Power Shot")
	}
	if restored.Abilities[0].Type != AbilityTypeExecution {
		t.Errorf("Ability[0].Type: got %q, want %q", restored.Abilities[0].Type, AbilityTypeExecution)
	}
	if restored.Abilities[1].Type != AbilityTypeReaction {
		t.Errorf("Ability[1].Type: got %q, want %q", restored.Abilities[1].Type, AbilityTypeReaction)
	}
	if len(restored.Abilities[1].Triggers) != 1 {
		t.Errorf("Ability[1].Triggers count: got %d, want 1", len(restored.Abilities[1].Triggers))
	}
	if len(restored.LevelHistory) != 5 {
		t.Errorf("LevelHistory count: got %d, want 5", len(restored.LevelHistory))
	}
	if restored.Settings.TraitCount != 22 {
		t.Errorf("TraitCount: got %d, want 22", restored.Settings.TraitCount)
	}

	// Verify point calculations match
	if restored.TotalTraitPoints() != original.TotalTraitPoints() {
		t.Errorf("TotalTraitPoints: got %d, want %d", restored.TotalTraitPoints(), original.TotalTraitPoints())
	}
	if restored.TotalAbilityPoints() != original.TotalAbilityPoints() {
		t.Errorf("TotalAbilityPoints: got %d, want %d", restored.TotalAbilityPoints(), original.TotalAbilityPoints())
	}
	if restored.TraitPointsSpent() != original.TraitPointsSpent() {
		t.Errorf("TraitPointsSpent: got %d, want %d", restored.TraitPointsSpent(), original.TraitPointsSpent())
	}
	if restored.AbilityPointsSpent() != original.AbilityPointsSpent() {
		t.Errorf("AbilityPointsSpent: got %d, want %d", restored.AbilityPointsSpent(), original.AbilityPointsSpent())
	}
}

func TestAbilityYAMLRoundTrip(t *testing.T) {
	original := &Ability{
		Name:       "Fire Punch",
		Type:       AbilityTypeExecution,
		EnergyCost: 5,
		ActionCost: 2,
		Enactments: []Enactment{
			{
				Type:                    EnactmentDamage,
				DamageDice:              "d6",
				BaseEnactmentEnergyCost: 0,
				Perks: []Perk{
					{Description: "Shift Dice Tier up", AddCost: 2, Amount: 1, TotalAddCost: 2, EnergyCost: 1},
				},
				Interactions: []Interaction{
					{
						Type:         InteractionDirect,
						Engager:      "Self",
						TargetAmount: 1,
						Range:        "1m",
						Validation: &Validation{
							EngagementRoll: "Power",
							CounterRoll:    []string{"Reflex", "Constitution"},
						},
					},
				},
			},
			{
				Type:                    EnactmentPersistentEffect,
				Duration:               "3 rounds",
				TriggerTiming:          "Start of Target's Turn",
				Solutions:              []string{"Dexterity", "Constitution"},
				IsOptional:             true,
				BaseEnactmentEnergyCost: 2,
				Effects: []PersistentEffect{
					{
						Type:                    EnactmentDamage,
						DamageDice:              "1d4",
						IsOptional:              false,
						BaseEnactmentEnergyCost: 0,
					},
				},
				Perks: []Perk{
					{Description: "Add another round", AddCost: 2, Amount: 1, TotalAddCost: 2, EnergyCost: 1, IsOptional: true},
				},
				Interactions: []Interaction{
					{
						Type:    InteractionDirect,
						Engager: "Self",
						Range:   "1m",
						Validation: &Validation{
							EngagementRoll: "Power",
							CounterRoll:    []string{"Constitution", "Reflex"},
						},
					},
				},
			},
		},
	}

	data, err := MarshalAbility(original)
	if err != nil {
		t.Fatalf("MarshalAbility failed: %v", err)
	}

	t.Logf("Ability YAML:\n%s", string(data))

	restored, err := UnmarshalAbility(data)
	if err != nil {
		t.Fatalf("UnmarshalAbility failed: %v", err)
	}

	if restored.Name != original.Name {
		t.Errorf("Name: got %q, want %q", restored.Name, original.Name)
	}
	if restored.Type != original.Type {
		t.Errorf("Type: got %q, want %q", restored.Type, original.Type)
	}
	if len(restored.Enactments) != 2 {
		t.Fatalf("Enactments count: got %d, want 2", len(restored.Enactments))
	}
	if restored.Enactments[1].Type != EnactmentPersistentEffect {
		t.Errorf("Enactment[1].Type: got %q, want %q", restored.Enactments[1].Type, EnactmentPersistentEffect)
	}
	if len(restored.Enactments[1].Effects) != 1 {
		t.Errorf("Effects count: got %d, want 1", len(restored.Enactments[1].Effects))
	}
	if restored.Enactments[1].IsOptional != true {
		t.Errorf("Enactment[1].IsOptional: got %v, want true", restored.Enactments[1].IsOptional)
	}

	// Verify cost calculations
	if restored.TotalAddCost() != original.TotalAddCost() {
		t.Errorf("TotalAddCost: got %d, want %d", restored.TotalAddCost(), original.TotalAddCost())
	}
	if restored.TotalEnergyCost() != original.TotalEnergyCost() {
		t.Errorf("TotalEnergyCost: got %d, want %d", restored.TotalEnergyCost(), original.TotalEnergyCost())
	}
}

func TestCharacterPointCalculations(t *testing.T) {
	c := &Character{
		Version: "1.0",
		Level:   5,
		GeneralTraits: GeneralTraits{
			Strength:  Trained, // 2
			Dexterity: Expert,  // 3
			Crafting:  Trained, // 2
		},
		CombativeTraits: CombativeTraits{
			Offense: OffenseTraits{Precision: Trained}, // 2
			Vital:   VitalTraits{HP: Untrained},        // 1
		},
		Abilities: []Ability{
			{
				Name:       "Test",
				Type:       AbilityTypeExecution,
				EnergyCost: 3,
				Perks: []Perk{
					{TotalAddCost: 5, EnergyCost: 0},
				},
			},
		},
		LevelHistory: []LevelSnapshot{
			{Level: 1, TraitPointsGained: 0, AbilityPointsGained: 0},
			{Level: 2, TraitPointsGained: 1, AbilityPointsGained: 2},
			{Level: 3, TraitPointsGained: 1, AbilityPointsGained: 3},
			{Level: 4, TraitPointsGained: 1, AbilityPointsGained: 2},
			{Level: 5, TraitPointsGained: 2, AbilityPointsGained: 4},
		},
		Settings: CampaignSettings{TraitCount: 22},
	}

	// Base: (22+2)/3 = 8, + 1+1+1+2 = 13
	if got := c.TotalTraitPoints(); got != 13 {
		t.Errorf("TotalTraitPoints = %d, want 13", got)
	}

	// Spent: 2+3+2+2+1 = 10
	if got := c.TraitPointsSpent(); got != 10 {
		t.Errorf("TraitPointsSpent = %d, want 10", got)
	}

	// Available: 13 - 10 = 3
	if got := c.TraitPointsAvailable(); got != 3 {
		t.Errorf("TraitPointsAvailable = %d, want 3", got)
	}

	// Ability points base: 10 + 2+3+2+4 = 21
	if got := c.TotalAbilityPoints(); got != 21 {
		t.Errorf("TotalAbilityPoints = %d, want 21", got)
	}

	// Ability points spent: 5 (from the test ability perk)
	if got := c.AbilityPointsSpent(); got != 5 {
		t.Errorf("AbilityPointsSpent = %d, want 5", got)
	}

	// Available: 21 - 5 = 16
	if got := c.AbilityPointsAvailable(); got != 16 {
		t.Errorf("AbilityPointsAvailable = %d, want 16", got)
	}
}
