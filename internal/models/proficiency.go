package models

// Proficiency represents the skill level for a trait.
type Proficiency int

const (
	Clumsy    Proficiency = iota // d4
	Untrained                    // d6
	Trained                      // d8
	Expert                       // d10
	Master                       // d12
	Legendary                    // d20
)

// String returns the human-readable name of the proficiency level.
func (p Proficiency) String() string {
	switch p {
	case Clumsy:
		return "Clumsy"
	case Untrained:
		return "Untrained"
	case Trained:
		return "Trained"
	case Expert:
		return "Expert"
	case Master:
		return "Master"
	case Legendary:
		return "Legendary"
	default:
		return "Unknown"
	}
}

// Dice returns the dice notation for this proficiency level.
func (p Proficiency) Dice() string {
	switch p {
	case Clumsy:
		return "d4"
	case Untrained:
		return "d6"
	case Trained:
		return "d8"
	case Expert:
		return "d10"
	case Master:
		return "d12"
	case Legendary:
		return "d20"
	default:
		return "d4"
	}
}

// Cost returns how many trait points this proficiency level costs from Clumsy.
// Each step costs 1 point. Legendary is not purchasable (n/p).
func (p Proficiency) Cost() int {
	return int(p)
}

// HPValue returns the HP for a given vital proficiency level.
func (p Proficiency) HPValue() int {
	switch p {
	case Clumsy:
		return 8
	case Untrained:
		return 12
	case Trained:
		return 16
	case Expert:
		return 20
	case Master:
		return 24
	case Legendary:
		return 28
	default:
		return 8
	}
}

// MovementValue returns the movement squares for a given vital proficiency level.
func (p Proficiency) MovementValue() int {
	switch p {
	case Clumsy:
		return 3
	case Untrained:
		return 4
	case Trained:
		return 5
	case Expert:
		return 6
	case Master:
		return 7
	case Legendary:
		return 8
	default:
		return 3
	}
}

// EnergyValue returns the energy pool for a given vital proficiency level.
func (p Proficiency) EnergyValue() int {
	switch p {
	case Clumsy:
		return 3
	case Untrained:
		return 4
	case Trained:
		return 5
	case Expert:
		return 6
	case Master:
		return 7
	case Legendary:
		return 8
	default:
		return 3
	}
}

// EnergyPoolValue returns the total energy pool (used for ability costs).
func (p Proficiency) EnergyPoolValue() int {
	switch p {
	case Clumsy:
		return 5
	case Untrained:
		return 8
	case Trained:
		return 12
	case Expert:
		return 16
	case Master:
		return 20
	case Legendary:
		return 25
	default:
		return 5
	}
}

// ProficiencyFromString converts a string name to a Proficiency value.
func ProficiencyFromString(s string) Proficiency {
	switch s {
	case "Clumsy":
		return Clumsy
	case "Untrained":
		return Untrained
	case "Trained":
		return Trained
	case "Expert":
		return Expert
	case "Master":
		return Master
	case "Legendary":
		return Legendary
	default:
		return Clumsy
	}
}

// AllProficiencies returns all proficiency levels in order.
func AllProficiencies() []Proficiency {
	return []Proficiency{Clumsy, Untrained, Trained, Expert, Master, Legendary}
}

// PurchasableProficiencies returns proficiency levels that can be purchased with points.
func PurchasableProficiencies() []Proficiency {
	return []Proficiency{Clumsy, Untrained, Trained, Expert, Master}
}
