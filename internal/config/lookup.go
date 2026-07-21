package config

// AbilityType returns the ability-type component with the given id.
func (c *Config) AbilityType(id string) (Component, bool) {
	if comp, ok := c.AbilityTypes.Get(id); ok {
		return *comp, true
	}
	return Component{}, false
}

// Enactment returns the enactment component with the given id.
func (c *Config) Enactment(id string) (Component, bool) {
	if comp, ok := c.Enactments.Get(id); ok {
		return *comp, true
	}
	return Component{}, false
}

// Interaction returns the interaction component with the given id.
func (c *Config) Interaction(id string) (Component, bool) {
	if comp, ok := c.Interactions.Get(id); ok {
		return *comp, true
	}
	return Component{}, false
}

// Proficiency returns the proficiency tier with the given id.
func (c *Config) Proficiency(id string) (Proficiency, bool) {
	for _, p := range c.Proficiencies {
		if p.ID == id {
			return p, true
		}
	}
	return Proficiency{}, false
}

// ProficiencyCost returns the trait-point cost of a proficiency id (0 if none).
func (c *Config) ProficiencyCost(id string) int {
	if p, ok := c.Proficiency(id); ok {
		return p.Cost
	}
	return 0
}

// DefaultProficiencyID returns the id of the first proficiency, used as the
// starting tier for new characters.
func (c *Config) DefaultProficiencyID() string {
	if len(c.Proficiencies) > 0 {
		return c.Proficiencies[0].ID
	}
	return ""
}

// ResolveOptions returns the concrete option list for a field, expanding an
// options_source reference server-side when present.
func (c *Config) ResolveOptions(f Field) []Option {
	if f.OptionsSource != "" {
		return c.OptionsFor(f.OptionsSource)
	}
	return f.Options
}

// OptionGroup is a labelled set of options, used to render <optgroup> blocks.
// A group with an empty Label is rendered as ungrouped options.
type OptionGroup struct {
	Label   string
	Options []Option
}

// ResolveOptionGroups returns options grouped for display. Trait sources that
// span multiple categories (traits_all) are grouped by category so a single
// large dropdown stays readable; every other source becomes one unlabelled
// group.
func (c *Config) ResolveOptionGroups(f Field) []OptionGroup {
	if f.OptionsSource == "traits_all" {
		var groups []OptionGroup
		seen := map[string]bool{}
		for _, cat := range []string{"general", "offense", "defense"} {
			var opts []Option
			for _, t := range c.Traits.Items[cat] {
				if seen[t] {
					continue
				}
				seen[t] = true
				opts = append(opts, Option{Value: t, Label: t})
			}
			if len(opts) > 0 {
				groups = append(groups, OptionGroup{Label: titleCase(cat), Options: opts})
			}
		}
		return groups
	}
	return []OptionGroup{{Label: "", Options: c.ResolveOptions(f)}}
}

// OptionsFor resolves a named options_source into a concrete option list. All
// dynamic sources are derived from the config (traits, dice, states); a handful
// of small static lists are defined here.
func (c *Config) OptionsFor(source string) []Option {
	switch source {
	case "traits_general":
		return strOptions(c.Traits.Items["general"])
	case "traits_offense":
		return strOptions(c.Traits.Items["offense"])
	case "traits_defense":
		return strOptions(c.Traits.Items["defense"])
	case "traits_vital":
		return strOptions(c.Traits.Items["vital"])
	case "traits_all":
		// Vitals (HP/Movement/Energy) are not selectable as traits.
		var all []string
		seen := map[string]bool{}
		for _, cat := range []string{"general", "offense", "defense"} {
			for _, t := range c.Traits.Items[cat] {
				if !seen[t] {
					seen[t] = true
					all = append(all, t)
				}
			}
		}
		return strOptions(all)

	case "dice_damage":
		return strOptions(c.Dice.Damage)
	case "dice_generic":
		return strOptions(c.Dice.Generic)
	case "states_general":
		out := make([]Option, 0, len(c.GeneralStates))
		for _, s := range c.GeneralStates {
			out = append(out, Option{Value: s.ID, Label: s.Name})
		}
		return out
	case "states_specific":
		out := make([]Option, 0, len(c.SpecificStates))
		for _, s := range c.SpecificStates {
			out = append(out, Option{Value: s.ID, Label: s.Name})
		}
		return out
	case "directions_all", "directions":
		return strOptions([]string{"North", "East", "South", "West", "Up", "Down", "Any"})
	case "shift_directions":
		return strOptions([]string{"UP", "DOWN"})
	case "ability_types":
		return componentOptions(c.AbilityTypes)
	case "enactment_types":
		return componentOptions(c.Enactments)
	case "interaction_types":
		return componentOptions(c.Interactions)
	case "aoe_trigger_timings", "trigger_timings":
		return strOptions([]string{"Start of Target Turn", "End of Engager Turn"})
	case "reaction_triggers":
		return strOptions(reactionTriggers)
	case "knockout_options":
		return strOptions(knockoutOptions)
	default:
		return nil
	}
}

// reactionTriggers is the static list of reaction trigger events. These were
// previously derived from a legacy "triggers" block that has been removed.
var reactionTriggers = []string{
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

// knockoutOptions is the static list of phase knockout requirements. These were
// previously derived from a legacy "knockout_requirements" block.
var knockoutOptions = []string{
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

func strOptions(vals []string) []Option {
	out := make([]Option, 0, len(vals))
	for _, v := range vals {
		out = append(out, Option{Value: v, Label: v})
	}
	return out
}

func componentOptions(m ComponentMap) []Option {
	out := make([]Option, 0, len(m.Order))
	for _, comp := range m.List() {
		out = append(out, Option{Value: comp.ID, Label: comp.DisplayName()})
	}
	return out
}

// TraitPointBudget returns the trait-point budget for a given character level,
// read from the leveling table.
func (c *Config) TraitPointBudget(level int) int {
	return budgetForLevel(c.Leveling.TraitPoints, level)
}

// AbilityPointBudget returns the ability-point budget for a given level.
func (c *Config) AbilityPointBudget(level int) int {
	return budgetForLevel(c.Leveling.AbilityPoints, level)
}

func budgetForLevel(t LevelTable, level int) int {
	if level < 1 {
		level = 1
	}
	best := 0
	for _, e := range t.Levels {
		if e.Level <= level && e.Total >= best {
			best = e.Total
		}
		if e.Level == level {
			return e.Total
		}
	}
	return best
}
