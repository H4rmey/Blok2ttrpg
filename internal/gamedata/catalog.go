// Package gamedata contains the TTRPG rule catalog: perk lists, compatible
// enactments per ability type, triggers, and other reference data used
// by the ability builder wizard.
package gamedata

// PerkDef defines a perk that can be applied to an ability component.
type PerkDef struct {
	Description string
	EnergyCost  int
	AddCost     int
	MaxAmount   int    // 0 = unlimited stacking
	InputType   string // "", "trait", "direction", "text" - type of extra input needed
	InputLabel  string // Label for the input field
}

// EnactmentDef defines an enactment type available to add to an ability.
type EnactmentDef struct {
	Type       string
	EnergyCost int // Base enactment energy cost (0 if first enactment)
	AddCost    int // Add cost to include this enactment
}

// TriggerDef defines a reaction trigger.
type TriggerDef struct {
	Name       string
	EnergyCost int
	AddCost    int
}

// AbilityTypePerks returns the type-level perks for a given ability type.
func AbilityTypePerks(abilityType string) []PerkDef {
	switch abilityType {
	case "Execution":
		return []PerkDef{
			{Description: "Has Item Dependency", EnergyCost: 0, AddCost: -1, MaxAmount: 1, InputType: "text", InputLabel: "Which item?"},
			{Description: "Reduce Energy cost by 1 (min 1)", EnergyCost: -1, AddCost: 3, MaxAmount: 2},
			{Description: "Increase Energy cost by 1", EnergyCost: 1, AddCost: -2, MaxAmount: 3},
			{Description: "Reduce Action Cost by 1", EnergyCost: 1, AddCost: 4, MaxAmount: 1},
			{Description: "Increase Action Cost by 1", EnergyCost: 0, AddCost: -2, MaxAmount: 2},
		}
	case "Reaction":
		return []PerkDef{
			{Description: "Add 1 meter to reaction range", EnergyCost: 0, AddCost: 1, MaxAmount: 0},
			{Description: "Add one more use per round", EnergyCost: 1, AddCost: 4, MaxAmount: 2},
			{Description: "Reduce Energy cost by 1 (min 1)", EnergyCost: -1, AddCost: 3, MaxAmount: 2},
			{Description: "Increase Energy cost by 1", EnergyCost: 1, AddCost: -2, MaxAmount: 3},
			{Description: "Has Item Dependency", EnergyCost: 0, AddCost: -1, MaxAmount: 1, InputType: "text", InputLabel: "Which item?"},
		}
	case "Phase":
		return []PerkDef{
			{Description: "Add another round to Phase and Reverse Phase", EnergyCost: 1, AddCost: 2, MaxAmount: 3},
			{Description: "Remove one round from the Reverse Phase", EnergyCost: 0, AddCost: 4, MaxAmount: 2},
			{Description: "All knockout requirements must be met instead", EnergyCost: 0, AddCost: 3, MaxAmount: 1},
			{Description: "Knockout can be used on the Reverse Phase", EnergyCost: 0, AddCost: 3, MaxAmount: 1},
			{Description: "Reduce Energy cost by 1 (min 1)", EnergyCost: -1, AddCost: 3, MaxAmount: 2},
			{Description: "Increase Energy cost by 1", EnergyCost: 1, AddCost: -2, MaxAmount: 3},
			{Description: "Has Item Dependency", EnergyCost: 0, AddCost: -1, MaxAmount: 1, InputType: "text", InputLabel: "Which item?"},
		}
	case "Minion":
		return []PerkDef{
			{Description: "Increase Health by 5", EnergyCost: 0, AddCost: 1, MaxAmount: 0},
			{Description: "Increase Attack by 1d6", EnergyCost: 1, AddCost: 2, MaxAmount: 3},
			{Description: "Increase Defense by 1d6", EnergyCost: 0, AddCost: 2, MaxAmount: 3},
			{Description: "Increase Speed by 2m", EnergyCost: 0, AddCost: 1, MaxAmount: 0},
			{Description: "Add an additional action per turn", EnergyCost: 2, AddCost: 5, MaxAmount: 1},
			{Description: "Minion can use an additional Ability", EnergyCost: 1, AddCost: 4, MaxAmount: 2},
			{Description: "Increase Lifetime by 1 round", EnergyCost: 1, AddCost: 1, MaxAmount: 0},
			{Description: "Minion has item dependency", EnergyCost: 0, AddCost: -1, MaxAmount: 1, InputType: "text", InputLabel: "Which item?"},
			{Description: "Reduce Energy cost by 1 (min 1)", EnergyCost: -1, AddCost: 3, MaxAmount: 2},
			{Description: "Increase Energy cost by 1", EnergyCost: 1, AddCost: -2, MaxAmount: 3},
		}
	}
	return nil
}

// CompatibleEnactments returns the enactments available for a given ability type.
func CompatibleEnactments(abilityType string) []EnactmentDef {
	switch abilityType {
	case "Execution":
		return []EnactmentDef{
			{Type: "Enact Damage", EnergyCost: 1, AddCost: 2},
			{Type: "Enact Healing", EnergyCost: 1, AddCost: 2},
			{Type: "Enact Movement", EnergyCost: 0, AddCost: 1},
			{Type: "Enact Proficiency Shift", EnergyCost: 1, AddCost: 2},
			{Type: "Enact Persistent Effect", EnergyCost: 2, AddCost: 3},
		}
	case "Reaction":
		return []EnactmentDef{
			{Type: "Enact Damage", EnergyCost: 1, AddCost: 2},
			{Type: "Enact Healing", EnergyCost: 1, AddCost: 2},
			{Type: "Enact Movement", EnergyCost: 0, AddCost: 1},
			{Type: "Enact Proficiency Shift", EnergyCost: 1, AddCost: 2},
			{Type: "Enact Persistent Effect", EnergyCost: 2, AddCost: 3},
			{Type: "Enact other execution", EnergyCost: 2, AddCost: 4},
		}
	case "Phase":
		return []EnactmentDef{
			{Type: "Enact Proficiency Shift", EnergyCost: 0, AddCost: 0},
			{Type: "Enact Damage", EnergyCost: 1, AddCost: 2},
			{Type: "Enact Healing", EnergyCost: 1, AddCost: 2},
		}
	case "Minion":
		return []EnactmentDef{
			{Type: "Enact Damage", EnergyCost: 1, AddCost: 2},
			{Type: "Enact Healing", EnergyCost: 1, AddCost: 2},
			{Type: "Enact Movement", EnergyCost: 0, AddCost: 1},
			{Type: "Enact Proficiency Shift", EnergyCost: 1, AddCost: 2},
			{Type: "Enact Persistent Effect", EnergyCost: 2, AddCost: 3},
		}
	}
	return nil
}

// EnactmentPerks returns the perks for a given enactment type.
func EnactmentPerks(enactmentType string) []PerkDef {
	switch enactmentType {
	case "Enact Damage":
		return []PerkDef{
			{Description: "Shift Dice Tier of damage up", EnergyCost: 1, AddCost: 2, MaxAmount: 4},
			{Description: "Change Damage Dice to one of your traits", EnergyCost: 0, AddCost: 3, MaxAmount: 1, InputType: "trait", InputLabel: "Which trait?"},
			{Description: "Add a flat +1 bonus to the result", EnergyCost: 0, AddCost: 2, MaxAmount: 0},
			{Description: "Add an Offensive Trait Dice to the Damage Dice", EnergyCost: 2, AddCost: 4, MaxAmount: 1, InputType: "offensive_trait", InputLabel: "Which offensive trait?"},
			{Description: "Will Always Resolve", EnergyCost: 3, AddCost: 5, MaxAmount: 1},
			{Description: "Use the result of another roll for this entry", EnergyCost: 1, AddCost: 3, MaxAmount: 1, InputType: "text", InputLabel: "Which roll?"},
		}
	case "Enact Healing":
		return []PerkDef{
			{Description: "Shift Dice Tier of Heal up", EnergyCost: 1, AddCost: 2, MaxAmount: 4},
			{Description: "Change Heal Dice to one of your traits", EnergyCost: 0, AddCost: 3, MaxAmount: 1, InputType: "trait", InputLabel: "Which trait?"},
			{Description: "Add a flat +1 bonus to the result", EnergyCost: 0, AddCost: 2, MaxAmount: 0},
			{Description: "Add Medicine Trait Dice to the heal effect", EnergyCost: 1, AddCost: 3, MaxAmount: 1},
			{Description: "Will Always Resolve", EnergyCost: 2, AddCost: 4, MaxAmount: 1},
			{Description: "Use the result of another roll for this entry", EnergyCost: 1, AddCost: 3, MaxAmount: 1, InputType: "text", InputLabel: "Which roll?"},
		}
	case "Enact Movement":
		return []PerkDef{
			{Description: "Add 1m to the movement", EnergyCost: 0, AddCost: 1, MaxAmount: 0},
			{Description: "Add another direction option", EnergyCost: 0, AddCost: 1, MaxAmount: 5, InputType: "direction", InputLabel: "Direction"},
			{Description: "Change Origin to something else", EnergyCost: 1, AddCost: 2, MaxAmount: 1, InputType: "text", InputLabel: "New origin"},
			{Description: "Change total movement to any other trait", EnergyCost: 0, AddCost: 3, MaxAmount: 1, InputType: "trait", InputLabel: "Which trait?"},
			{Description: "Will Always Resolve", EnergyCost: 1, AddCost: 3, MaxAmount: 1},
			{Description: "Use the result of another roll for this entry", EnergyCost: 1, AddCost: 3, MaxAmount: 1, InputType: "text", InputLabel: "Which roll?"},
		}
	case "Enact Proficiency Shift":
		return []PerkDef{
			{Description: "Add another use to the Shifted Trait", EnergyCost: 1, AddCost: 3, MaxAmount: 3},
			{Description: "Shift the same Trait a second time", EnergyCost: 1, AddCost: 3, MaxAmount: 2},
			{Description: "You may choose if you use the Shifted Proficiency or not", EnergyCost: 0, AddCost: 2, MaxAmount: 1},
			{Description: "Will Always Resolve", EnergyCost: 2, AddCost: 4, MaxAmount: 1},
			{Description: "Use the result of another roll for this entry", EnergyCost: 1, AddCost: 3, MaxAmount: 1},
		}
	case "Enact Persistent Effect":
		return []PerkDef{
			{Description: "Add another round to the Effect", EnergyCost: 1, AddCost: 2, MaxAmount: 3},
			{Description: "Remove one Solution option", EnergyCost: 1, AddCost: 3, MaxAmount: 1},
			{Description: "Add another Effect", EnergyCost: 2, AddCost: 4, MaxAmount: 2},
			{Description: "Will Always Resolve", EnergyCost: 3, AddCost: 5, MaxAmount: 1},
			{Description: "Use the result of another roll for this entry", EnergyCost: 1, AddCost: 3, MaxAmount: 1},
		}
	}
	return nil
}

// ReactionTriggers returns all available reaction triggers.
func ReactionTriggers() []TriggerDef {
	return []TriggerDef{
		{Name: "Someone runs away from you", EnergyCost: 0, AddCost: 2},
		{Name: "Someone runs towards you", EnergyCost: 0, AddCost: 2},
		{Name: "Someone has to do a Defense roll", EnergyCost: 0, AddCost: 2},
		{Name: "Someone takes damage", EnergyCost: 0, AddCost: 2},
		{Name: "Someone gets healed", EnergyCost: 0, AddCost: 2},
		{Name: "Someone gets an adjustment", EnergyCost: 0, AddCost: 2},
		{Name: "Someone does a skill check", EnergyCost: 0, AddCost: 2},
		{Name: "Someone summons a minion", EnergyCost: 0, AddCost: 2},
		{Name: "Someone grabs your character", EnergyCost: 0, AddCost: 2},
		{Name: "You walk towards someone", EnergyCost: 0, AddCost: 2},
		{Name: "You walk away from someone", EnergyCost: 0, AddCost: 2},
	}
}

// KnockoutRequirements returns phase knockout options.
func KnockoutRequirements() []string {
	return []string{
		"You take damage",
		"You fall unconscious / die",
		"You get grabbed/restrained",
	}
}

// MovementDirections returns available movement directions.
func MovementDirections() []string {
	return []string{"Up", "Down", "Away", "Towards", "Forward", "Left", "Right"}
}

// InteractionPerks returns the perks available for a given interaction type.
func InteractionPerks(interactionType string) []PerkDef {
	switch interactionType {
	case "Self":
		return []PerkDef{
			{Description: "Validation counter roll replaced by Generic Dice: d8", EnergyCost: 0, AddCost: 2, MaxAmount: 1},
			{Description: "Will always resolve", EnergyCost: 3, AddCost: 5, MaxAmount: 1},
			{Description: "Use the result of another roll for this entry", EnergyCost: 1, AddCost: 3, MaxAmount: 1},
		}
	case "Direct":
		return []PerkDef{
			{Description: "Increase range with 1m", EnergyCost: 0, AddCost: 1, MaxAmount: 0},
			{Description: "Add another target", EnergyCost: 2, AddCost: 3, MaxAmount: 3},
			{Description: "Will always resolve", EnergyCost: 3, AddCost: 5, MaxAmount: 1},
			{Description: "Use the result of another roll for this entry", EnergyCost: 1, AddCost: 3, MaxAmount: 1},
		}
	case "Ranged":
		return []PerkDef{
			{Description: "Increase range with 2m", EnergyCost: 0, AddCost: 1, MaxAmount: 0},
			{Description: "Decrease range with 2m", EnergyCost: 0, AddCost: -1, MaxAmount: 4},
			{Description: "Add another target", EnergyCost: 2, AddCost: 3, MaxAmount: 3},
			{Description: "Target does not have to be visible", EnergyCost: 1, AddCost: 3, MaxAmount: 1},
			{Description: "Target may be obstructed", EnergyCost: 1, AddCost: 3, MaxAmount: 1},
			{Description: "Remove the Engagement Roll Penalty", EnergyCost: 1, AddCost: 3, MaxAmount: 1},
			{Description: "Will always resolve", EnergyCost: 3, AddCost: 5, MaxAmount: 1},
			{Description: "Use the result of another roll for this entry", EnergyCost: 1, AddCost: 3, MaxAmount: 1},
		}
	case "Area":
		return []PerkDef{
			{Description: "Increase radius with 1m", EnergyCost: 1, AddCost: 2, MaxAmount: 0},
			{Description: "Increase Range by 2m", EnergyCost: 0, AddCost: 1, MaxAmount: 0},
			{Description: "Change Origin to something else", EnergyCost: 1, AddCost: 2, MaxAmount: 1, InputType: "text", InputLabel: "New origin"},
			{Description: "Will always resolve", EnergyCost: 3, AddCost: 5, MaxAmount: 1},
			{Description: "Use the result of another roll for this entry", EnergyCost: 1, AddCost: 3, MaxAmount: 1, InputType: "text", InputLabel: "Which roll?"},
		}
	case "Area of Effect":
		return []PerkDef{
			{Description: "Increase radius with 1m", EnergyCost: 1, AddCost: 2, MaxAmount: 0},
			{Description: "Increase Range by 2m", EnergyCost: 0, AddCost: 1, MaxAmount: 0},
			{Description: "Change Origin to something else", EnergyCost: 1, AddCost: 2, MaxAmount: 1, InputType: "text", InputLabel: "New origin"},
			{Description: "Increase the amount of rounds by 1", EnergyCost: 1, AddCost: 2, MaxAmount: 3},
			{Description: "Engager is immune to the effect", EnergyCost: 0, AddCost: 2, MaxAmount: 1},
			{Description: "Will always resolve", EnergyCost: 3, AddCost: 5, MaxAmount: 1},
			{Description: "Use the result of another roll for this entry", EnergyCost: 1, AddCost: 3, MaxAmount: 1, InputType: "text", InputLabel: "Which roll?"},
		}
	}
	return nil
}

// ValidationPerks returns the perks available for validation configuration.
func ValidationPerks() []PerkDef {
	return []PerkDef{
		{Description: "Engage Roll replaced by Generic Roll (default 1d6)", EnergyCost: 0, AddCost: -2, MaxAmount: 1},
		{Description: "Counter Roll replaced by Generic Roll (default 1d12)", EnergyCost: 0, AddCost: -2, MaxAmount: 1},
		{Description: "Replace one Counter Roll to any other Trait", EnergyCost: 0, AddCost: 2, MaxAmount: 1},
		{Description: "Remove one Counter Roll option", EnergyCost: 1, AddCost: 3, MaxAmount: 1},
		{Description: "Replace Engagement Roll to any other Trait", EnergyCost: 0, AddCost: 3, MaxAmount: 1},
		{Description: "Add another Engagement Roll option", EnergyCost: 0, AddCost: 2, MaxAmount: 2},
		{Description: "Shift Generic Counter Roll UP", EnergyCost: 0, AddCost: -2, MaxAmount: 3},
		{Description: "Shift Generic Counter Roll DOWN", EnergyCost: 1, AddCost: 3, MaxAmount: 3},
		{Description: "Shift Generic Engagement Roll UP", EnergyCost: 1, AddCost: 3, MaxAmount: 3},
		{Description: "Shift Generic Engagement Roll DOWN", EnergyCost: 0, AddCost: -2, MaxAmount: 3},
		{Description: "Use the result of another roll for this entry", EnergyCost: 1, AddCost: 3, MaxAmount: 1},
	}
}

// AllInteractionTypes returns all available interaction types.
func AllInteractionTypes() []string {
	return []string{"Self", "Direct", "Ranged", "Area", "Area of Effect"}
}

// OffensiveTraits returns the traits usable for engagement rolls.
func OffensiveTraits() []string {
	return []string{"Precision", "Power", "Mind", "Magic"}
}

// DefensiveTraits returns the traits usable for counter rolls.
func DefensiveTraits() []string {
	return []string{"Reflex", "Constitution", "Mind", "Magic"}
}

// AllTraits returns all general traits for trait selection dropdowns.
func AllTraits() []string {
	return []string{
		"Strength", "Dexterity", "Stealth", "Perception", "Nature",
		"Crafting", "People Skill", "Performance", "Thievery", "Knowledge", "Magic",
	}
}
