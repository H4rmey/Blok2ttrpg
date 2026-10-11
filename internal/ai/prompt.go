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

// SystemPrompt returns the full system message for generating a character at
// the requested level. reference is the config-derived rules summary (see
// docs.AIReference); it is passed in so the caller can build it once and reuse
// it across the retries of a single generation.
func SystemPrompt(loaded *config.Loaded, reference string, level int) string {
	cfg := loaded.Config
	level = cfg.ClampLevel(level)
	var b strings.Builder

	b.WriteString("You are a character generator for the ")
	b.WriteString(cfg.Title)
	b.WriteString(" tabletop RPG system. Build one complete, playable character that matches the user's description at exactly the requested level, then output it as importable YAML.\n\n")

	b.WriteString("## Output rules\n\n")
	b.WriteString("- Output ONLY the raw YAML. No markdown fences, no explanation, no commentary before or after.\n")
	b.WriteString("- The YAML must use the schema below and reference only ids from the reference section.\n")
	b.WriteString("- Spend EXACTLY 100% of every point pool for the level: no unspent points, no overspending. Both the skill-point and perk-point totals must equal their budgets exactly.\n")
	b.WriteString("- Every skill listed in `skills` must use a proficiency id from the ladder; skills not listed default to the free baseline.\n")
	b.WriteString("- Proficiency values are lowercase ids (`novice`, `expert`, ...), NEVER the tier name (`Novice`, `Expert`).\n")
	b.WriteString("- Build cohesive perks that fit the character's described theme and highest skills.\n")
	b.WriteString("- Fill EVERY trait key in the `traits` map (name, age, description, motivation, problem and the traits) with specific, evocative text.\n")
	b.WriteString("- Use a small number of substantial perks (2 to 5). Never pad the budget with many trivial perks; raise a perk's cost through its fields instead.\n")
	b.WriteString("- Keep the YAML compact: include only fields whose value differs from the default.\n\n")

	b.WriteString("## Character YAML schema\n\n")
	b.WriteString(schemaOutline(cfg))
	b.WriteString("\n\n")

	fmt.Fprintf(&b, "## Required budgets for level %d\n\n", level)
	fmt.Fprintf(&b, "- Skill points: you MUST spend exactly %d.\n", cfg.SkillPointBudget(level))
	fmt.Fprintf(&b, "- Perk points: you MUST spend exactly %d (sum of every perk's build cost).\n", cfg.PerkPointBudget(level))

	b.WriteString("\n# Ruleset reference\n\n")
	if reference != "" {
		b.WriteString(reference)
	}
	b.WriteString("\n")

	return b.String()
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
