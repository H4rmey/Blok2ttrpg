// This file holds the human descriptions for every yaml key of the config types
// documented by schemaref.go.
//
// It is deliberately separate from the reflection logic so that the only manual
// step in producing the configuration reference is writing a sentence per key.
// LintSchemaCoverage walks the same structs and reports any key missing an entry
// here, and the docs render test fails on that report, so a new config key
// cannot be added without also being documented.
package docs

// schemaPurposes maps a documented type name to its yaml key descriptions. The
// keys of the inner map are the yaml names as written in the config files.
var schemaPurposes = map[string]map[string]string{
	"Cost": {
		"build_cost":  "Build points charged when the perk is created or upgraded. Negative values refund points.",
		"energy_cost": "Energy charged every time the perk is used. Negative values refund energy.",
	},

	"PerStep": {
		"increase": "Cost charged per step the value moves above its default.",
		"decrease": "Cost charged per step the value moves below its default. A negative cost here refunds points for weakening the perk.",
	},

	"GroupOffsets": {
		"default_group": "The skill group the field leans toward. A value with no group prefix is treated as belonging to this group.",
		"offsets":       "Extra cost added per skill group, keyed by group id. Picking a skill outside the preferred group normally costs more.",
	},

	"InlineBuilder": {
		"kind": "Which component map the selected dropdown value resolves against: enactment, interaction or perk_type. The referenced component's own fields render beneath the dropdown and add their cost.",
	},

	"Option": {
		"value":              "The value stored when this option is selected. Must be stable: it is persisted on saved perks.",
		"label":              "Text shown in the dropdown. Falls back to the value when omitted.",
		"information":        "Help text shown as a hover tooltip on the option.",
		"render_information": "When true, show this option's information as plain text below the dropdown once selected instead of behind a hover indicator.",
		"cost":               "Cost added when this option is the selected one.",
		"fields":             "Extra fields revealed when this option is selected.",
	},

	"Field": {
		"key":                "Stable submitted field name. It is persisted in saved perks, so renaming it breaks existing data.",
		"label":              "Text shown next to the field in the builder and used as its name in the documentation.",
		"type":               "Which kind of input this is. See the supported field types below.",
		"description":        "Reader-facing explanation of the field, used as the lead sentence in the generated build guide.",
		"information":        "Help text shown as a hover tooltip next to the field label.",
		"render_information": "When true, show the information as plain text below the field instead of behind a hover indicator.",
		"default":            "The starting value of the field. For numbers it is also the zero point that per_step costs are measured from.",
		"cost":               "Flat cost applied whenever the field is active: a checked checkbox or a dropdown with a non-empty value.",
		"min":                "Lowest value a free_number accepts.",
		"max":                "Highest value a free_number accepts.",
		"step":               "Size of one increment of a free_number. Defaults to 1 when omitted.",
		"rounding":           "How partial steps are handled: ceil rounds a partial step up, floor rounds it down.",
		"per_step":           "Cost charged per step a free_number moves away from its default.",
		"options":            "Inline list of dropdown choices. Do not combine with options_source on the same field.",
		"options_source":     "Name of a shared option source or grouped source to populate the dropdown from.",
		"shift_key":          "On a condition field, the sibling field holding the shift amount. The condition's shift_cost is multiplied by the absolute value read from it. Defaults to shift_amount.",
		"row_fields":         "The fields that make up one row of a multiselect or conditions field.",
		"default_count":      "How many rows a repeatable field starts with. per_item costs are measured relative to this count.",
		"per_item":           "Cost charged per row added beyond default_count, or refunded per row removed below it.",
		"row_defaults":       "Values pre-filled into the first rows of a repeatable field, one map of row field key to value per row.",
		"conjunction":        "The joining word shown to the left of each repeatable row after the first, either and or or. Display only; defaults to or.",
		"visibility_when":    "Name of the sibling field that controls whether this field is shown.",
		"show_when":          "The value the controlling field must have for this field to be shown. A hidden field contributes no cost.",
		"inline_builder":     "Turns a dropdown into a nested builder for the component the selected value names.",
		"group_offsets":      "Per-skill-group cost offsets for a dropdown backed by a multi-group skill source.",
	},

	"Component": {
		"name":                  "Display name, used by perk types.",
		"type":                  "Display name, used by enactments and interactions.",
		"description":           "Reader-facing explanation of what the component does.",
		"information":           "Help text shown as a hover tooltip next to the component.",
		"render_information":    "When true, show the information as plain text under the component header instead of behind a hover indicator.",
		"base_cost":             "Flat cost of including this component. The first enactment of an perk has its base cost waived.",
		"base_energy":           "Starting energy cost of an perk of this type.",
		"base_action":           "Starting action cost of an perk of this type.",
		"base_range":            "Starting range in metres.",
		"base_uses":             "Starting number of uses.",
		"base_duration":         "Starting duration in rounds.",
		"base_reverse_duration": "Starting number of rounds a phase takes to reverse itself.",
		"base_health":           "Starting health, used by summoned minions.",
		"base_lifetime":         "Starting lifetime in rounds, used by summoned minions.",
		"base_upkeep_action":    "Actions required each round to sustain the perk.",
		"base_upkeep_energy":    "Energy required each round to sustain the perk.",
		"skip_invoke_cost":      "When true, using this perk type is exempt from the invoke point cost that the equivalent improvised action pays. Set on Reaction: an improvised invoke action costs an invoke point, but a Reaction perk bought with build points does not.",
		"default_range":         "Default range of an interaction, in metres.",
		"default_targets":       "Default number of targets an interaction affects.",
		"default_radius":        "Default radius of an area interaction, in metres.",
		"default_duration":      "Default duration of an area interaction, in rounds.",
		"allowed_interactions":  "When set, only these interactions are offered for this enactment, in the order listed.",
		"blocked_interactions":  "When set (and no allowed list is given), every interaction except these is offered.",
		"allowed_validations":   "When set, only these validation fields are shown for this enactment.",
		"blocked_validations":   "When set (and no allowed list is given), every validation field except these is shown.",
		"allowed_enactments":    "When set, only these enactments are offered for this perk type, in the order listed.",
		"blocked_enactments":    "When set (and no allowed list is given), every enactment except these is offered.",
		"fields":                "The choices this component offers, driving both the builder form and the cost engine.",
	},

	"Proficiency": {
		"id":     "Stable identifier for the tier, referenced by skills and by default_proficiency.",
		"name":   "Display name of the tier.",
		"cost":   "Skill points charged to climb onto this rung from the one below it.",
		"note":   "Optional remark about the tier.",
		"die":    "Die rolled by a dice-backed skill at this tier. Used for every skill group unless dice overrides it.",
		"dice":   "Per-skill-group die overrides, keyed by group id. Only needed when a group differs from die.",
		"vitals": "Numeric values this tier grants for the vital skills, keyed by lowercase skill name (hp, movement, energy).",
	},

	"Leveling": {
		"max_level":     "Highest level a character may reach. Levels are clamped to this value on every edit and on import.",
		"skill_points":  "Budget progression for the skill point pool.",
		"perk_points":   "Budget progression for the perk (perk) point pool.",
		"invoke_points": "Pool progression for the per-session invoke point currency.",
	},

	"InvokeTable": {
		"start":           "Invoke points available at level 1.",
		"per_step":        "Invoke points added by each step of the curve.",
		"levels_per_step": "How many levels apart the steps are. With a value of 2, per_step points are granted at levels 3, 5, 7 and so on.",
		"levels":          "Optional explicit per-level overrides for an irregular curve. A row for the requested level wins over the formula.",
	},

	"Invoking": {
		"refresh":            "When every character's invoke points return to their maximum, written as a phrase for the rulebook.",
		"allow_over_maximum": "Whether points earned in play may be banked above the level maximum. Defaults to false, so a full pool must be spent before more can be banked.",
		"spends":             "The ways an invoke point can be spent.",
		"gains":              "The ways an invoke point is earned, in and out of combat.",
		"combat_gains":       "Extra earning triggers that apply only during combat, plus the per-combat cap on them.",
		"invoke_actions":     "Cost and timing of the improvised out-of-turn action bought with an invoke point.",
		"reaction_limit":     "How often a character may act out of turn at all, counting invoke actions and Reaction perks together.",
	},

	"InvokeSpend": {
		"id":          "Stable identifier for the spend.",
		"name":        "Display name shown in the rulebook.",
		"invoke_cost": "Invoke points charged for this spend.",
		"energy_cost": "Energy charged for this spend, on top of the invoke points.",
		"description": "What the spend does and any conditions on using it.",
	},

	"InvokeGain": {
		"id":          "Stable identifier for the gain.",
		"name":        "Display name shown in the rulebook.",
		"points":      "Invoke points earned.",
		"description": "What has to happen to earn the points.",
	},

	"CombatGains": {
		"max_per_combat": "Ceiling on invoke points a character may earn from a single combat, no matter how many triggers fire.",
		"triggers":       "The in-combat events that earn invoke points.",
	},

	"CombatGainTrigger": {
		"id":              "Stable identifier for the trigger.",
		"name":            "Display name shown in the rulebook.",
		"points":          "Invoke points earned when the trigger fires.",
		"once_per_combat": "Whether the trigger may only fire a single time per fight.",
		"every_rounds":    "When set, the trigger fires at the start of every Nth round instead of on an event.",
		"description":     "What has to happen for the trigger to fire.",
	},

	"InvokeActions": {
		"name":        "What the rulebook calls this action. Defaults to \"Invoke Action\". It is named so the improvised out-of-turn action is never confused with the Reaction perk type.",
		"invoke_cost": "Invoke points charged to improvise an out-of-turn action at the table.",
		"energy_cost": "Energy charged to improvise an out-of-turn action.",
		"timing":      "When an invoke action may interrupt: between_actions resolves it before or after a whole action, anytime allows it mid-action.",
		"description": "Reader-facing explanation of what an invoke action is, rendered into the rulebook.",
	},

	"ReactionLimit": {
		"max_per_round":    "How many times a character may act out of turn in one round, by either route.",
		"shared":           "Whether max_per_round is a single budget covering invoke actions and Reaction perks together. Defaults to true, so owning several Reaction perks does not allow acting out of turn more than once per round.",
		"perk_invoke_cost": "Invoke points charged to fire a Reaction perk. Normally zero, because the point is considered pre-paid by the perk point spent to build it.",
	},

	"EnergyOverdraft": {
		"hp_per_energy":           "HP paid for each point of missing Energy when using a perk you cannot afford. Keep this well above 1: HP and Energy grow at the same rate per proficiency rung, so a 1:1 rate turns the HP pool into a second Energy pool and Energy stops being a resource.",
		"escalation_per_use":      "Added to hp_per_energy on each later overdraft in the same scene, the way repeated movement gets progressively more expensive. Zero keeps the rate flat.",
		"condition_on_overdraft":  "Condition id applied when a character overdraws, or empty for none. Pointing it at a condition that raises energy costs makes the rule self-limiting.",
		"allow_partial_execution": "Whether a character may instead drop the enactments they cannot pay for. Defaults to true; it degrades the perk rather than the character.",
	},

	"Negotiation": {
		"motivation":      "The ladder of outcomes a negotiation can end on.",
		"patience":        "The countdown that bounds how long an NPC keeps listening.",
		"argument":        "How a single attempt to persuade resolves.",
		"trait_alignment": "How the NPC's own traits clamp movement on the motivation ladder.",
	},

	"Motivation": {
		"min":   "Lowest rung of the ladder.",
		"max":   "Highest rung of the ladder.",
		"rungs": "The rungs themselves, each with the outcome it produces. The starting rung is set by the GM per NPC, not here.",
	},

	"MotivationRung": {
		"id":      "Stable identifier for the rung.",
		"value":   "The numeric motivation this rung sits at.",
		"name":    "Display name of the rung.",
		"outcome": "What happens when a negotiation ends here. Guidance for the GM rather than a result applied mechanically.",
	},

	"Patience": {
		"min":               "Value at which the negotiation ends, normally zero.",
		"max":               "Highest patience an NPC may have. The starting value is set by the GM per NPC.",
		"loss_per_argument": "Patience spent by each argument, whether it succeeds or fails.",
	},

	"Argument": {
		"on_success":    "Motivation change when the skill check succeeds.",
		"on_failure":    "Motivation change when the skill check fails.",
		"allow_repeats": "Whether the same argument or trait may be raised more than once in one negotiation. Defaults to false.",
	},

	"TraitAlignment": {
		"against": "Clamp applied when an argument runs against one of the NPC's traits. cannot_increase stops motivation rising.",
		"with":    "Clamp applied when an argument runs with one of the NPC's traits. cannot_decrease stops motivation falling.",
	},

	"LevelTable": {
		"start":     "Budget granted at level 1.",
		"per_level": "Budget added by each level after the first. The budget for a level is start + per_level x (level - 1).",
		"levels":    "Optional explicit per-level overrides for a non-linear curve. A row for the requested level wins over the formula.",
	},

	"Condition": {
		"id":          "Stable identifier, referenced by saved perks and generated instructions.",
		"name":        "Display name of the condition.",
		"description": "What the condition does, shown in the rulebook and as a hover tooltip.",
		"build_cost":  "Build points charged for a fixed condition.",
		"energy_cost": "Energy charged for a fixed condition.",
		"min_shift":   "Lowest shift a shiftable condition may apply. A non-zero min or max shift is what makes a condition shiftable.",
		"max_shift":   "Highest shift a shiftable condition may apply.",
		"shift_cost":  "Cost charged per unit of shift applied by a shiftable condition.",
		"affects_skills": "The skills this condition moves, as \"<group>.<skill>\" keys from the skills list. " +
			"Naming them is what lets a character sheet recolour those skills and read them at their shifted value. " +
			"Leave it empty for a condition that changes what a character may do rather than what they roll; such a " +
			"condition asks for no value when applied. Movement may be shifted, HP and Energy may not.",
		"fixed_shift": "The shift applied to every skill in affects_skills by a non-shiftable condition, for a " +
			"condition whose own text already names its magnitude (\"shifted one down\"). It is ignored for shiftable " +
			"conditions, which take their magnitude from the value chosen when they are applied.",
		"selectable": "Whether the condition can be purchased in the builder. Defaults to true; set false for states the rules impose rather than ones a player buys.",
	},

	"Validations": {
		"information":        "Help text for the validation section as a whole.",
		"render_information": "When true, show the section information as plain text under the header instead of behind a hover indicator.",
		"fields":             "The validation choices offered, such as which skill resolves the engagement roll and what counters it.",
	},
}
