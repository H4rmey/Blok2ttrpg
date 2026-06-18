package models

import (
	"testing"
)

func TestProficiencyString(t *testing.T) {
	tests := []struct {
		prof Proficiency
		want string
	}{
		{Clumsy, "Clumsy"},
		{Untrained, "Untrained"},
		{Trained, "Trained"},
		{Expert, "Expert"},
		{Master, "Master"},
		{Legendary, "Legendary"},
	}
	for _, tt := range tests {
		if got := tt.prof.String(); got != tt.want {
			t.Errorf("Proficiency(%d).String() = %q, want %q", tt.prof, got, tt.want)
		}
	}
}

func TestProficiencyDice(t *testing.T) {
	tests := []struct {
		prof Proficiency
		want string
	}{
		{Clumsy, "d4"},
		{Untrained, "d6"},
		{Trained, "d8"},
		{Expert, "d10"},
		{Master, "d12"},
		{Legendary, "d20"},
	}
	for _, tt := range tests {
		if got := tt.prof.Dice(); got != tt.want {
			t.Errorf("Proficiency(%d).Dice() = %q, want %q", tt.prof, got, tt.want)
		}
	}
}

func TestProficiencyCost(t *testing.T) {
	tests := []struct {
		prof Proficiency
		want int
	}{
		{Clumsy, 0},
		{Untrained, 1},
		{Trained, 2},
		{Expert, 3},
		{Master, 4},
		{Legendary, 5},
	}
	for _, tt := range tests {
		if got := tt.prof.Cost(); got != tt.want {
			t.Errorf("Proficiency(%d).Cost() = %d, want %d", tt.prof, got, tt.want)
		}
	}
}

func TestProficiencyFromString(t *testing.T) {
	tests := []struct {
		input string
		want  Proficiency
	}{
		{"Clumsy", Clumsy},
		{"Untrained", Untrained},
		{"Trained", Trained},
		{"Expert", Expert},
		{"Master", Master},
		{"Legendary", Legendary},
		{"garbage", Clumsy}, // default
	}
	for _, tt := range tests {
		if got := ProficiencyFromString(tt.input); got != tt.want {
			t.Errorf("ProficiencyFromString(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestVitalValues(t *testing.T) {
	if got := Expert.HPValue(); got != 20 {
		t.Errorf("Expert.HPValue() = %d, want 20", got)
	}
	if got := Trained.MovementValue(); got != 5 {
		t.Errorf("Trained.MovementValue() = %d, want 5", got)
	}
	if got := Master.EnergyValue(); got != 7 {
		t.Errorf("Master.EnergyValue() = %d, want 7", got)
	}
	if got := Expert.EnergyPoolValue(); got != 16 {
		t.Errorf("Expert.EnergyPoolValue() = %d, want 16", got)
	}
}

func TestGeneralTraitsPointsSpent(t *testing.T) {
	g := GeneralTraits{
		Strength:  Trained,  // 2
		Dexterity: Expert,   // 3
		Stealth:   Untrained, // 1
		// rest default to Clumsy (0)
	}
	if got := g.PointsSpent(); got != 6 {
		t.Errorf("GeneralTraits.PointsSpent() = %d, want 6", got)
	}
}

func TestCombativeTraitsPointsSpent(t *testing.T) {
	c := CombativeTraits{
		Offense: OffenseTraits{
			Precision: Expert, // 3
			Power:     Trained, // 2
		},
		Defense: DefenseTraits{
			Reflex: Trained, // 2
		},
		Vital: VitalTraits{
			HP:     Untrained, // 1
			Energy: Trained,   // 2
		},
	}
	if got := c.PointsSpent(); got != 10 {
		t.Errorf("CombativeTraits.PointsSpent() = %d, want 10", got)
	}
}

func TestNewCharacter(t *testing.T) {
	c := NewCharacter()
	if c.Level != 1 {
		t.Errorf("NewCharacter().Level = %d, want 1", c.Level)
	}
	if c.Settings.TraitCount != 22 {
		t.Errorf("NewCharacter().Settings.TraitCount = %d, want 22", c.Settings.TraitCount)
	}
	// Base trait points: (22+2)/3 = 8
	if got := c.TotalTraitPoints(); got != 8 {
		t.Errorf("NewCharacter().TotalTraitPoints() = %d, want 8", got)
	}
	// Base ability points: 10
	if got := c.TotalAbilityPoints(); got != 10 {
		t.Errorf("NewCharacter().TotalAbilityPoints() = %d, want 10", got)
	}
	// Nothing spent yet
	if got := c.TraitPointsSpent(); got != 0 {
		t.Errorf("NewCharacter().TraitPointsSpent() = %d, want 0", got)
	}
	if got := c.AbilityPointsSpent(); got != 0 {
		t.Errorf("NewCharacter().AbilityPointsSpent() = %d, want 0", got)
	}
}

func TestAbilityTotalAddCost(t *testing.T) {
	a := Ability{
		Name:       "Fireball",
		Type:       AbilityTypeExecution,
		EnergyCost: 3,
		ActionCost: 2,
		Perks: []Perk{
			{Description: "Reduce Action Cost by 1", AddCost: 4, Amount: 1, TotalAddCost: 4, EnergyCost: 1},
		},
		Enactments: []Enactment{
			{
				Type:                    EnactmentDamage,
				DamageDice:              "d8",
				BaseEnactmentEnergyCost: 0,
				Perks: []Perk{
					{Description: "Shift Dice Tier up", AddCost: 2, Amount: 2, TotalAddCost: 4, EnergyCost: 2},
				},
				Interactions: []Interaction{
					{
						Type:    InteractionRanged,
						Engager: "Self",
						Range:   "14m",
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
	}

	// 4 (ability perk) + 4 (enactment perk) + 2 (interaction perk) = 10
	if got := a.TotalAddCost(); got != 10 {
		t.Errorf("Ability.TotalAddCost() = %d, want 10", got)
	}
}

func TestAbilityTotalEnergyCost(t *testing.T) {
	a := Ability{
		Name:       "Quick Strike",
		Type:       AbilityTypeExecution,
		EnergyCost: 3,
		Perks: []Perk{
			{Description: "Reduce Energy cost by 1", AddCost: 3, Amount: 1, TotalAddCost: 3, EnergyCost: -1},
		},
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
						Type:    InteractionDirect,
						Engager: "Self",
						Validation: &Validation{
							EngagementRoll: "Power",
							CounterRoll:    []string{"Reflex", "Constitution"},
						},
					},
				},
			},
		},
	}

	// 3 (base) + (-1 from perk) + 0 (enactment base) + 1 (enactment perk) = 3
	if got := a.TotalEnergyCost(); got != 3 {
		t.Errorf("Ability.TotalEnergyCost() = %d, want 3", got)
	}
}
