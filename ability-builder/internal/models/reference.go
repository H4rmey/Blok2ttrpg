package models

// --- General Traits (ordered) ---

var GeneralTraitNames = []string{
	"Strength",
	"Dexterity",
	"Stealth",
	"Perception",
	"Nature",
	"Crafting",
	"People Skill",
	"Performance",
	"Thievery",
	"Knowledge",
	"Magic",
}

// --- Combative Traits ---

var OffenseTraitNames = []string{
	"Precision",
	"Power",
	"Mind",
	"Magic",
}

var DefenseTraitNames = []string{
	"Reflex",
	"Constitution",
	"Mind",
	"Magic",
}

// AllTraitNames returns all trait names (general + combative).
func AllTraitNames() []string {
	all := make([]string, 0, len(GeneralTraitNames)+len(OffenseTraitNames)+len(DefenseTraitNames))
	all = append(all, GeneralTraitNames...)
	all = append(all, OffenseTraitNames...)
	all = append(all, DefenseTraitNames...)
	return all
}

// --- Ability Type Options ---

// ReactionTriggers are the 21 trigger conditions for Reaction abilities.
var ReactionTriggers = []string{
	"Target moves away from engager",
	"Target moves towards engager",
	"Target moves past engager",
	"Engager gets healed by target",
	"Target damages engager",
	"Target makes a trait check",
	"Target starts casting an ability",
	"Target ends their turn within range",
	"Target enters interaction range",
	"Target leaves interaction range",
	"Target fails a validation",
	"Target succeeds on a validation",
	"Target becomes affected by an enactment",
	"Engager takes damage",
	"Engager gets targeted by an ability",
	"Ally within range takes damage",
	"Ally within range gets healed",
	"A target is moved by an effect",
	"A persistent effect triggers",
	"A minion is summoned within range",
}

// KnockoutOptions are the 15 options for Phase knockout requirements.
var KnockoutOptions = []string{
	"None",
	"Engager takes damage",
	"Engager falls unconscious",
	"Engager dies",
	"Engager gets grabbed or restrained",
	"Engager moves voluntarily",
	"Engager is moved by another effect",
	"Engager fails a validation",
	"Engager uses another phase",
	"Engager loses line of sight to target",
	"Target moves out of range",
	"Target falls unconscious",
	"Target dies",
	"Target succeeds on a counter roll",
	"Phase duration expires",
	"Engager runs out of energy",
}

// --- Enactment Options ---

// DirectionOptions for movement enactments.
var DirectionOptions = []string{
	"Up",
	"Down",
	"Away",
	"Towards",
	"Forward",
	"Left",
	"Right",
	"Free",
}

// ShiftDirectionOptions for proficiency shift.
var ShiftDirectionOptions = []string{
	"UP",
	"DOWN",
}

// TriggerTimings for persistent effects.
var TriggerTimings = []string{
	"Start of Target Turn",
	"End of Engager Turn",
}

// AoETriggerTimings for Area of Effect interactions.
var AoETriggerTimings = []string{
	"Start of Engager Turn",
	"End of Engager Turn",
	"Start of Character Turn",
	"Entering Area",
}

// --- Dice Tier Options ---

// DamageDiceOptions for the main source dice.
var DamageDiceOptions = []string{"d4", "d6", "d8", "d10", "d12"}

// GenericDieOptions for validation engage mode and similar.
var GenericDieOptions = []string{"d6", "d8", "d10", "d12"}

// PersistentEffectTypes are the four options the persistent effect card can
// apply each round.
var PersistentEffectTypes = []string{
	"Enact Damage",
	"Enact Healing",
	"Enact Movement",
	"Enact Proficiency Shift",
}

// --- Compatible Enactments per Ability Type ---

var ExecutionEnactments = []EnactmentType{
	EnactDamage,
	EnactHealing,
	EnactMovement,
	EnactProficiencyShift,
	EnactPersistentEffect,
}

var ReactionEnactments = []EnactmentType{
	EnactDamage,
	EnactHealing,
	EnactMovement,
	EnactProficiencyShift,
	EnactPersistentEffect,
}

var PhaseEnactments = []EnactmentType{
	EnactDamage,
	EnactHealing,
	EnactProficiencyShift,
}

var MinionEnactments = []EnactmentType{
	EnactDamage,
	EnactHealing,
	EnactMovement,
	EnactProficiencyShift,
	EnactPersistentEffect,
}

// CompatibleEnactments maps ability types to their compatible enactment types.
var CompatibleEnactments = map[AbilityType][]EnactmentType{
	AbilityExecution: ExecutionEnactments,
	AbilityReaction:  ReactionEnactments,
	AbilityPhase:     PhaseEnactments,
	AbilityMinion:    MinionEnactments,
}
