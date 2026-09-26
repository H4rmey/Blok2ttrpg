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
		"build_cost":  "Build points charged when the ability is created or upgraded. Negative values refund points.",
		"energy_cost": "Energy charged every time the ability is used. Negative values refund energy.",
	},

	"PerStep": {
		"increase": "Cost charged per step the value moves above its default.",
		"decrease": "Cost charged per step the value moves below its default. A negative cost here refunds points for weakening the ability.",
	},

	"GroupOffsets": {
		"default_group": "The trait group the field leans toward. A value with no group prefix is treated as belonging to this group.",
		"offsets":       "Extra cost added per trait group, keyed by group id. Picking a trait outside the preferred group normally costs more.",
	},

	"InlineBuilder": {
		"kind": "Which component map the selected dropdown value resolves against: enactment, interaction or ability_type. The referenced component's own fields render beneath the dropdown and add their cost.",
	},

	"Option": {
		"value":              "The value stored when this option is selected. Must be stable: it is persisted on saved abilities.",
		"label":              "Text shown in the dropdown. Falls back to the value when omitted.",
		"information":        "Help text shown as a hover tooltip on the option.",
		"render_information": "When true, show this option's information as plain text below the dropdown once selected instead of behind a hover indicator.",
		"cost":               "Cost added when this option is the selected one.",
		"fields":             "Extra fields revealed when this option is selected.",
	},

	"Field": {
		"key":                "Stable submitted field name. It is persisted in saved abilities, so renaming it breaks existing data.",
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
		"group_offsets":      "Per-trait-group cost offsets for a dropdown backed by a multi-group trait source.",
	},

	"Component": {
		"name":                  "Display name, used by ability types.",
		"type":                  "Display name, used by enactments and interactions.",
		"description":           "Reader-facing explanation of what the component does.",
		"information":           "Help text shown as a hover tooltip next to the component.",
		"render_information":    "When true, show the information as plain text under the component header instead of behind a hover indicator.",
		"base_cost":             "Flat cost of including this component. The first enactment of an ability has its base cost waived.",
		"base_energy":           "Starting energy cost of an ability of this type.",
		"base_action":           "Starting action cost of an ability of this type.",
		"base_range":            "Starting range in metres.",
		"base_uses":             "Starting number of uses.",
		"base_duration":         "Starting duration in rounds.",
		"base_reverse_duration": "Starting number of rounds a phase takes to reverse itself.",
		"base_health":           "Starting health, used by summoned minions.",
		"base_lifetime":         "Starting lifetime in rounds, used by summoned minions.",
		"base_upkeep_action":    "Actions required each round to sustain the ability.",
		"base_upkeep_energy":    "Energy required each round to sustain the ability.",
		"default_range":         "Default range of an interaction, in metres.",
		"default_targets":       "Default number of targets an interaction affects.",
		"default_radius":        "Default radius of an area interaction, in metres.",
		"default_duration":      "Default duration of an area interaction, in rounds.",
		"allowed_interactions":  "When set, only these interactions are offered for this enactment, in the order listed.",
		"blocked_interactions":  "When set (and no allowed list is given), every interaction except these is offered.",
		"allowed_validations":   "When set, only these validation fields are shown for this enactment.",
		"blocked_validations":   "When set (and no allowed list is given), every validation field except these is shown.",
		"allowed_enactments":    "When set, only these enactments are offered for this ability type, in the order listed.",
		"blocked_enactments":    "When set (and no allowed list is given), every enactment except these is offered.",
		"fields":                "The choices this component offers, driving both the builder form and the cost engine.",
	},

	"Proficiency": {
		"id":     "Stable identifier for the tier, referenced by traits and by default_proficiency.",
		"name":   "Display name of the tier.",
		"cost":   "Trait points charged to climb onto this rung from the one below it.",
		"note":   "Optional remark about the tier.",
		"die":    "Die rolled by a dice-backed trait at this tier. Used for every trait group unless dice overrides it.",
		"dice":   "Per-trait-group die overrides, keyed by group id. Only needed when a group differs from die.",
		"vitals": "Numeric values this tier grants for the vital traits, keyed by lowercase trait name (hp, movement, energy).",
	},

	"Leveling": {
		"max_level":      "Highest level a character may reach. Levels are clamped to this value on every edit and on import.",
		"trait_points":   "Budget progression for the trait (skill) point pool.",
		"ability_points": "Budget progression for the ability (perk) point pool.",
	},

	"LevelTable": {
		"start":     "Budget granted at level 1.",
		"per_level": "Budget added by each level after the first. The budget for a level is start + per_level x (level - 1).",
		"levels":    "Optional explicit per-level overrides for a non-linear curve. A row for the requested level wins over the formula.",
	},

	"Condition": {
		"id":          "Stable identifier, referenced by saved abilities and generated instructions.",
		"name":        "Display name of the condition.",
		"description": "What the condition does, shown in the rulebook and as a hover tooltip.",
		"build_cost":  "Build points charged for a fixed condition.",
		"energy_cost": "Energy charged for a fixed condition.",
		"min_shift":   "Lowest shift a shiftable condition may apply. A non-zero min or max shift is what makes a condition shiftable.",
		"max_shift":   "Highest shift a shiftable condition may apply.",
		"shift_cost":  "Cost charged per unit of shift applied by a shiftable condition.",
		"selectable":  "Whether the condition can be purchased in the builder. Defaults to true; set false for states the rules impose rather than ones a player buys.",
	},

	"Validations": {
		"information":        "Help text for the validation section as a whole.",
		"render_information": "When true, show the section information as plain text under the header instead of behind a hover indicator.",
		"fields":             "The validation choices offered, such as which trait resolves the engagement roll and what counters it.",
	},
}
