// This file renders the compact, config-derived reference handed to the AI
// character generator. It is deliberately smaller than the full rulebook
// (RenderMarkdown) so the prompt fits the context window of a small local model
// (Ollama / LM Studio) while still enumerating every valid id and field choice
// the generator must stay within.
//
// It reuses the same helper functions the rulebook uses, so the reference stays
// in sync with the loaded ruleset automatically.
package docs

import (
	"fmt"
	"strings"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
)

// AIReference builds the machine-readable rules summary for the generator.
func AIReference(loaded *config.Loaded, lib PackageLister) string {
	if loaded == nil || loaded.Config == nil {
		return ""
	}
	cfg := loaded.Config
	var b strings.Builder

	b.WriteString("## Skill points and proficiency ladder\n\n")
	b.WriteString(skillsTable(cfg))
	b.WriteString("\n\n")
	b.WriteString("**Proficiency values are ids, not tier names.** In the `skills` map, the value MUST be one of these lowercase ids. The number is its skill-point cost relative to the free default:\n\n")
	b.WriteString(proficiencyIDs(cfg))
	b.WriteString("\n\n## Point budgets by level\n\nSkill points:\n\n")
	b.WriteString(levelingTable(cfg, "skill"))
	b.WriteString("\n\nPerk points:\n\n")
	b.WriteString(levelingTable(cfg, "perk"))

	b.WriteString("\n\n## Exact perk spending: the 1-point dials\n\n")
	b.WriteString("The perk total must equal the budget EXACTLY. First pick perks and values that look right, then tune ONLY these number dials (each step is exactly 1 build point) until the total hits the budget. Do not add or remove perks to balance the budget:\n\n")
	b.WriteString(spendDials(cfg))

	b.WriteString("\n\n## Resource enactment: generator and spender must be separate perks\n\n")
	b.WriteString(resourcePattern())

	b.WriteString("\n\n## Valid skill keys\n\nThese are the only allowed `skills` map keys, written as `\"<group>.<Skill>\"`:\n\n")
	for _, g := range cfg.Skills.List() {
		for _, s := range g.Skills {
			fmt.Fprintf(&b, "- `%s.%s`\n", g.ID, s)
		}
	}

	b.WriteString("\n## Character traits\n\nPut these EXACT keys in the `traits` map (fill every one):\n\n")
	for _, f := range TraitFields(cfg) {
		label := f.Label
		if label == "" {
			label = f.Key
		}
		fmt.Fprintf(&b, "- `%s` — %s\n", f.Key, label)
	}

	b.WriteString("\n\n## Conditions\n\n")
	b.WriteString(conditionsTable(cfg))

	b.WriteString("\n\n## Perk types\n\nValid `type` values: ")
	b.WriteString(idList(componentIDs(cfg.PerkTypes.List())))
	b.WriteString("\n")
	for _, comp := range cfg.PerkTypes.List() {
		fmt.Fprintf(&b, "\n### Perk type `%s`\n\n%s\n", comp.ID, buildGuide(cfg, comp))
	}

	b.WriteString("\n\n## Enactments\n\nValid enactment `type` values: ")
	b.WriteString(idList(componentIDs(cfg.Enactments.List())))
	b.WriteString("\n")
	for _, comp := range cfg.Enactments.List() {
		fmt.Fprintf(&b, "\n### Enactment `%s`\n\n%s\n", comp.ID, buildGuide(cfg, comp))
	}

	b.WriteString("\n\n## Interactions\n\nValid `interaction` values: ")
	b.WriteString(idList(componentIDs(cfg.Interactions.List())))
	b.WriteString("\n")
	for _, comp := range cfg.Interactions.List() {
		fmt.Fprintf(&b, "\n### Interaction `%s`\n\n%s\n", comp.ID, buildGuide(cfg, comp))
	}

	if len(cfg.Validations.Fields) > 0 {
		b.WriteString("\n\n## Validation fields\n\n")
		b.WriteString(buildGuideFields(cfg, cfg.Validations.Fields))
	}

	if pt := passivesTable(cfg); strings.TrimSpace(pt) != "" {
		b.WriteString("\n\n## Passives\n\n")
		b.WriteString("Valid `passive_id` values (use the id, not the name): ")
		b.WriteString(idList(passiveIDs(cfg)))
		b.WriteString("\n\nA passive perk sets `type: passive` and nestles its catalogue id and configured values like this:\n\n")
		b.WriteString("```yaml\n- name: <perk name>\n  type: passive\n  fields:\n    passive_id: <passive id from the list above>\n    passive_fields:\n      <field key>: <value>   # optional; omit a field to use its default\n```\n\n")

		b.WriteString("Passives are worth 1 build point (see each entry's Cost) and CANNOT use the 1-point dials: their only field is `passive_fields`, and changing passive entries changes the perk's meaning. Use passives as-is for flavor and bulk, and tune the total with the built perks' dials.\n\n")
		b.WriteString("Costs and configurable parts per passive (names shown for reference):\n\n")
		b.WriteString(pt)
	}

	b.WriteString("\n\n## Resource enactment: generator and spender must be separate perks\n\n")
	b.WriteString(resourcePattern())

	return strings.TrimSpace(b.String())
}

// resourcePattern states the only correct way to use the resource enactment.
func resourcePattern() string {
	return "A resource pair is ALWAYS two perks using the same `resource_name`:\n\n" +
		"1. GENERATOR perk: an enactment of `type: resource` with `resource_mode: generate`, e.g.\n\n" +
		"```yaml\n- name: <generator name>\n  type: execution\n  enactments:\n    - type: resource\n      fields:\n        resource_mode: generate\n        resource_name: <Name>\n        generate_amount: 2\n```\n\n" +
		"2. SPENDER perk: a DIFFERENT perk with an enactment of `type: resource` and `resource_mode: consume` using that same `resource_name`.\n\n" +
		"NEVER put generate and consume in the same perk, and NEVER write `consumed_effect_ib` by hand: it contains the nested enactment's prefixed field keys and is built by the app, not the author. " +
		"A consume entry needs only `resource_mode: consume`, `resource_name` (matching the generator exactly) and `consumed_effect` (one of `damage`, `healing`, `motion`, `condition`, `shift`, `adjustment`, `phase`); " +
		"the nested enactment's own fields are configured afterwards in the character sheet UI.\n"
}

// spendDials lists the 1-build-point-per-step number dials the generator may
// turn to land the perk total exactly on budget.
func spendDials(cfg *config.Config) string {
	var b strings.Builder
	addDials(&b, cfg, componentPtrs(cfg.PerkTypes.List()), "perk type")
	addDials(&b, cfg, componentPtrs(cfg.Enactments.List()), "enactment")
	addDials(&b, cfg, componentPtrs(cfg.Interactions.List()), "interaction")
	out := strings.TrimSpace(b.String())
	if out == "" {
		return "_(no 1-point dials configured)_"
	}
	return out + "\n"
}

// componentPtrs widens a component slice for addDials.
func componentPtrs(comps []*config.Component) []config.Component {
	out := make([]config.Component, 0, len(comps))
	for _, c := range comps {
		if c != nil {
			out = append(out, *c)
		}
	}
	return out
}

// addDials appends one line per free_number field whose steps cost exactly 1
// build point, rendered as its full field path so the model can find it.
func addDials(b *strings.Builder, cfg *config.Config, comps []config.Component, kind string) {
	for _, comp := range comps {
		for _, f := range comp.Fields {
			if f.Type != "free_number" || f.PerStep == nil || f.PerStep.Increase == nil {
				continue
			}
			if f.PerStep.Increase.BuildCost != 1 {
				continue
			}
			fmt.Fprintf(b, "- %s `%s`, field `%s` (%s; steps %d..%d of %d)\n",
				kind, comp.ID, f.Key, orDash(f.Label), f.Min, f.Max, f.Step)
		}
	}
}

// TraitField is one character trait field: its storage key and its label.
type TraitField struct {
	Key   string
	Label string
	Type  string
}

// TraitFields returns every configured trait field in author order, so the AI
// generator can name the exact keys rather than just the section labels.
func TraitFields(cfg *config.Config) []TraitField {
	if cfg == nil {
		return nil
	}
	var out []TraitField
	for _, sec := range cfg.Traits.List() {
		for _, f := range sec.Fields {
			if f.Key == "" {
				continue
			}
			out = append(out, TraitField{Key: f.Key, Label: f.Label, Type: f.Type})
		}
	}
	return out
}

// proficiencyIDs lists the proficiency ladder as "id (cumulative cost)".
func proficiencyIDs(cfg *config.Config) string {
	if len(cfg.Proficiencies) == 0 {
		return "_(no proficiencies configured)_"
	}
	defaultIdx := cfg.DefaultProficiencyIndex()
	var b strings.Builder
	running := 0
	for i, p := range cfg.Proficiencies {
		if i <= defaultIdx {
			// At or below the baseline: cost is the refund owed back to the
			// default, so it walks backwards from zero.
			running = 0
			for j := i; j < defaultIdx; j++ {
				running -= cfg.Proficiencies[j+1].Cost
			}
		} else {
			running += p.Cost
		}
		label := "cost"
		if i == defaultIdx {
			label = "default, cost"
		}
		fmt.Fprintf(&b, "- `%s` (%s %d)\n", p.ID, label, running)
	}
	return strings.TrimSpace(b.String())
}

// passiveIDs extracts the ordered catalogue ids of the passives.
func passiveIDs(cfg *config.Config) []string {
	out := make([]string, 0, len(cfg.Passives.Entries))
	for _, p := range cfg.Passives.Entries {
		out = append(out, p.ID)
	}
	return out
}

// componentIDs extracts the ordered ids of a component list.
func componentIDs(comps []*config.Component) []string {
	out := make([]string, 0, len(comps))
	for _, c := range comps {
		if c != nil {
			out = append(out, c.ID)
		}
	}
	return out
}

// idList renders ids as a comma-separated list of backticked values.
func idList(ids []string) string {
	if len(ids) == 0 {
		return "_(none configured)_"
	}
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, "`"+id+"`")
	}
	return strings.Join(parts, ", ")
}
