// This file builds the system prompt for the character generator. It pairs a
// static directive (role, output rules, YAML shape) with the config-derived
// reference from internal/docs, so the model is told exactly which ids and
// point budgets to obey.
package ai

import (
	"fmt"
	"strings"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/docs"
)

// SystemPrompt returns the full system message for generating a character. The
// model chooses the level itself (there is no requested level); reference is
// the config-derived rules summary (see docs.AIReference), passed in so the
// caller can build it once and reuse it across the retries of a generation.
func SystemPrompt(loaded *config.Loaded, reference string) string {
	cfg := loaded.Config
	var b strings.Builder

	b.WriteString("You are a character generator for the ")
	b.WriteString(cfg.Title)
	b.WriteString(" tabletop RPG system. Build one complete, playable character that matches the user's description, then output it as importable YAML.\n\n")

	b.WriteString("## Output rules\n\n")
	b.WriteString("- Output ONLY the raw YAML. No markdown fences, no explanation, no commentary before or after.\n")
	b.WriteString("- The YAML must use the schema below and reference only ids from the reference section.\n")
	fmt.Fprintf(&b, "- Choose a level from 1 to %d that fits the description and set it in the `level` field. Spend EXACTLY 100%% of that level's point pools: no unspent points, no overspending. Both the skill-point and perk-point totals must equal their budgets exactly.\n", cfg.MaxLevel())
	b.WriteString("- To hit the perk budget exactly, tune ONLY the number dials listed under \"Exact perk spending: the 1-point dials\" (each step is exactly 1 build point). Do not add or remove perks to balance the budget.\n")
	b.WriteString("- A resource generator and its spender are ALWAYS two separate perks sharing one `resource_name`. Never put generate and consume in the same perk; never write `consumed_effect_ib` (the app builds it afterwards).\n")
	b.WriteString("- Every skill listed in `skills` must use a proficiency id from the ladder; skills not listed default to the free baseline.\n")
	b.WriteString("- Proficiency values are lowercase ids (`novice`, `expert`, ...), NEVER the tier name (`Novice`, `Expert`).\n")
	b.WriteString("- Build cohesive perks that fit the character's described theme and highest skills.\n")
	b.WriteString("- Fill EVERY trait key in the `traits` map (name, age, description, motivation, problem and the traits) with specific, evocative text.\n")
	b.WriteString("- Use a small number of substantial perks (2 to 5). Never pad the budget with many trivial perks; raise a perk's cost through its fields instead.\n")
	b.WriteString("- Every non-passive perk needs at least one enactment. A passive-flavoured idea must either use the passive shape below (a `passive_id` from the reference), or be built as an ordinary perk with real enactments. Never emit a perk with no enactments and no passive_id.\n")
	b.WriteString("- Keep the YAML compact: include only fields whose value differs from the default.\n\n")

	b.WriteString("## Character YAML schema\n\n")
	b.WriteString(schemaOutline(cfg))
	b.WriteString("\n\n")

	b.WriteString("# Ruleset reference\n\n")
	if reference != "" {
		b.WriteString(reference)
	}
	b.WriteString("\n")

	return b.String()
}

// PerkSystemPrompt returns the full system message for generating a single
// perk. reference is the same config-derived rules summary the character
// generator uses, so the model sees every valid id and the passive catalogue.
func PerkSystemPrompt(loaded *config.Loaded, reference string) string {
	cfg := loaded.Config
	var b strings.Builder

	b.WriteString("You are a perk generator for the ")
	b.WriteString(cfg.Title)
	b.WriteString(" tabletop RPG system. Build exactly one perk that matches the user's request, then output it as importable YAML.\n\n")

	b.WriteString("## Output rules\n\n")
	b.WriteString("- Output ONLY the raw YAML for a single perk. No markdown fences, no explanation, no commentary.\n")
	b.WriteString("- Reference only ids and field names from the reference section.\n")
	b.WriteString("- A non-passive perk MUST have at least one enactment. If the idea is passive-flavoured, either use the passive shape (type: passive with fields.passive_id from the passive list) or build it as an ordinary perk with real enactments. Never emit a perk with no enactments and no passive_id.\n")
	b.WriteString("- Fill every field the perk type and each enactment need; do not leave required fields blank.\n")
	b.WriteString("- Honor any energy cost, action cost or build point target the user gives, choosing fields that reach it.\n\n")

	b.WriteString("## Perk YAML schema\n\n")
	b.WriteString(perkSchemaOutline())
	b.WriteString("\n\n# Ruleset reference\n\n")
	if reference != "" {
		b.WriteString(reference)
	}
	b.WriteString("\n")

	return b.String()
}

// PerkUserPrompt frames the user's perk request, including the character's
// remaining perk points as guidance when it is known.
func PerkUserPrompt(desc string, remaining int) string {
	var b strings.Builder
	b.WriteString("Perk request:\n")
	b.WriteString(strings.TrimSpace(desc))
	b.WriteString("\n\n")
	if remaining > 0 {
		fmt.Fprintf(&b, "The character has %d perk points free; aim for a build cost at or below that unless the request says otherwise.\n\n", remaining)
	}
	b.WriteString("Generate the single perk YAML now.")
	return b.String()
}

// perkSchemaOutline renders the single-perk YAML shape (built perk or passive).
func perkSchemaOutline() string {
	return `name: <string>
description: <flavour text>
type: <perk type id>
tags: [<string>, ...]
fields: { comment: <string>, ... }   # perk-type fields
enactments:
  - type: <enactment id>
    fields: { comment: <string>, ... }
    interaction: <interaction id>
    interaction_data: { ... }
    validation_data: { ... }
# OR, for a passive from the catalogue:
# type: passive
# fields:
#   passive_id: <passive id>
#   passive_fields: { <field key>: <value>, ... }
`
}

// schemaOutline renders the character YAML shape with comments keyed off the
// loaded ruleset, so the model sees the field names and nesting it must emit.
func schemaOutline(cfg *config.Config) string {
	def := cfg.DefaultProficiencyID()
	if def == "" {
		def = "untrained"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "id: \"<optional slug>\"\nlevel: <integer 1-%d>\ntraits:\n", cfg.MaxLevel())
	for _, f := range docs.TraitFields(cfg) {
		fmt.Fprintf(&b, "  %s: %s\n", f.Key, traitPlaceholder(f.Type))
	}
	fmt.Fprintf(&b, "skills:\n  \"<group>.<Skill>\": <proficiency-id>   # omit skills left at the default (%s)\n", def)
	b.WriteString(`perks:
  - name: <string>
    description: <flavour text>
    type: <perk type id>
    tags: [<string>, ...]
    fields: { comment: <string>, ... }   # perk-type fields
    enactments:
      - type: <enactment id>
        fields: { comment: <string>, ... }
        interaction: <interaction id>
        interaction_data: { ... }
        validation_data: { ... }
  - name: <string>
    type: passive
    fields:
      passive_id: <passive id>
      passive_fields: { <field key>: <value>, ... }   # omit to use defaults
packages: []
`)
	return b.String()
}

// traitPlaceholder renders the YAML value placeholder for a trait field type.
func traitPlaceholder(fieldType string) string {
	switch fieldType {
	case "free_number":
		return "<number>"
	case "checkbox":
		return "<true/false>"
	default:
		return "<text>"
	}
}
