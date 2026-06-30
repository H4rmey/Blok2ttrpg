package docs

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/harmey/blok2ttrpg/ability-builder/internal/config"
)

// TemplateData is the flat data passed to markdown templates.
type TemplateData struct {
	Config *config.AbilityBuilderConfig

	AdditionalEnactmentTable template.HTML

	Execution                    config.AbilityTypeConfig
	ExecutionPerksTable          template.HTML
	ExecutionEnactmentsTable     template.HTML
	Reaction                     config.AbilityTypeConfig
	ReactionPerksTable           template.HTML
	ReactionTriggersTable        template.HTML
	ReactionEnactmentsTable      template.HTML
	Phase                        config.AbilityTypeConfig
	PhasePerksTable              template.HTML
	PhaseKnockoutsTable          template.HTML
	PhaseEnactmentsTable         template.HTML
	Preparation                  config.AbilityTypeConfig
	PreparationPerksTable        template.HTML
	PreparationTriggersTable     template.HTML
	PreparationEnactmentsTable   template.HTML
	Minion                       config.AbilityTypeConfig
	MinionPerksTable             template.HTML
	MinionEnactmentsTable        template.HTML

	Damage                       config.EnactmentConfig
	DamagePerksTable             template.HTML
	Healing                      config.EnactmentConfig
	HealingPerksTable            template.HTML
	Movement                     config.EnactmentConfig
	MovementPerksTable           template.HTML
	ProficiencyShift             config.EnactmentConfig
	ProficiencyShiftPerksTable   template.HTML
	PersistentEffect             config.EnactmentConfig
	PersistentEffectPerksTable   template.HTML
	PersistentEffectEffectsTable template.HTML
	State                        config.EnactmentConfig
	StatePerksTable              template.HTML

	Self                   config.InteractionConfig
	SelfPerksTable         template.HTML
	Direct                 config.InteractionConfig
	DirectPerksTable       template.HTML
	Ranged                 config.InteractionConfig
	RangedPerksTable       template.HTML
	Area                   config.InteractionConfig
	AreaPerksTable         template.HTML
	AreaOfEffect           config.InteractionConfig
	AreaOfEffectPerksTable template.HTML

	EngagementModesTable template.HTML
	CounterTypesTable  template.HTML
	TierShiftsTable    template.HTML
}

// DefaultDir returns the default directory containing markdown templates.
func DefaultDir() string {
	return "docs/ability-builder"
}

// Render generates the merged markdown documentation from the templates and
// the given configuration.
func Render(cfg *config.Config, dir string) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("configuration is nil")
	}
	files, err := CollectFiles(dir)
	if err != nil {
		return "", err
	}
	data := BuildTemplateData(cfg)
	return RenderFiles(files, data)
}

// CollectFiles returns all markdown template files in render order.
func CollectFiles(base string) ([]string, error) {
	files := []string{
		filepath.Join(base, "introduction.md"),
	}

	sections := []string{
		filepath.Join(base, "ability-types"),
		filepath.Join(base, "enactments"),
		filepath.Join(base, "interactions"),
	}
	for _, dir := range sections {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", dir, err)
		}
		var names []string
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			names = append(names, e.Name())
		}
		sort.Strings(names)
		for _, n := range names {
			files = append(files, filepath.Join(dir, n))
		}
	}

	files = append(files,
		filepath.Join(base, "validations.md"),
		filepath.Join(base, "leveling.md"),
	)
	return files, nil
}

// RenderFiles executes each markdown template and concatenates the output.
func RenderFiles(files []string, data TemplateData) (string, error) {
	var out strings.Builder

	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			return "", fmt.Errorf("reading %s: %w", f, err)
		}

		tmpl, err := template.New(filepath.Base(f)).Parse(string(raw))
		if err != nil {
			return "", fmt.Errorf("parsing %s: %w", f, err)
		}

		if err := tmpl.Execute(&out, &data); err != nil {
			return "", fmt.Errorf("executing %s: %w", f, err)
		}
		out.WriteString("\n\n")
	}

	return strings.TrimSpace(out.String()), nil
}

// BuildTemplateData creates the flat template data with all generated tables.
func BuildTemplateData(cfg *config.Config) TemplateData {
	ab := cfg.AbilityBuilder
	d := TemplateData{Config: &ab}

	d.AdditionalEnactmentTable = costTable("Additional Enactment", []config.PerkConfig{{Description: ab.AdditionalEnactment.Description, AddCost: ab.AdditionalEnactment.AddCost, EnergyCost: ab.AdditionalEnactment.EnergyCost}})

	d.Execution = ab.AbilityTypes["execution"]
	d.ExecutionPerksTable = perkTable(d.Execution.Perks)
	d.ExecutionEnactmentsTable = enactmentTable(d.Execution.CompatibleEnactments, ab.Enactments)

	d.Reaction = ab.AbilityTypes["reaction"]
	d.ReactionPerksTable = perkTable(d.Reaction.Perks)
	d.ReactionTriggersTable = triggerTable(d.Reaction.Triggers)
	d.ReactionEnactmentsTable = enactmentTable(d.Reaction.CompatibleEnactments, ab.Enactments)

	d.Phase = ab.AbilityTypes["phase"]
	d.PhasePerksTable = perkTable(d.Phase.Perks)
	d.PhaseKnockoutsTable = knockoutTable(d.Phase.KnockoutRequirements)
	d.PhaseEnactmentsTable = enactmentTable(d.Phase.CompatibleEnactments, ab.Enactments)

	d.Preparation = ab.AbilityTypes["preparation"]
	d.PreparationPerksTable = perkTable(d.Preparation.Perks)
	d.PreparationTriggersTable = triggerTable(d.Preparation.Triggers)
	d.PreparationEnactmentsTable = enactmentTable(d.Preparation.CompatibleEnactments, ab.Enactments)

	d.Minion = ab.AbilityTypes["minion"]
	d.MinionPerksTable = perkTable(d.Minion.Perks)
	d.MinionEnactmentsTable = enactmentTable(d.Minion.CompatibleEnactments, ab.Enactments)

	d.Damage = ab.Enactments["damage"]
	d.DamagePerksTable = perkTable(d.Damage.Perks)
	d.Healing = ab.Enactments["healing"]
	d.HealingPerksTable = perkTable(d.Healing.Perks)
	d.Movement = ab.Enactments["movement"]
	d.MovementPerksTable = perkTable(d.Movement.Perks)
	d.ProficiencyShift = ab.Enactments["proficiency_shift"]
	d.ProficiencyShiftPerksTable = perkTable(d.ProficiencyShift.Perks)
	d.PersistentEffect = ab.Enactments["persistent_effect"]
	d.PersistentEffectPerksTable = perkTable(d.PersistentEffect.Perks)
	d.PersistentEffectEffectsTable = effectTable(d.PersistentEffect.Effects)
	d.State = ab.Enactments["state"]
	d.StatePerksTable = perkTable(d.State.Perks)

	d.Self = ab.Interactions["self"]
	d.SelfPerksTable = perkTable(d.Self.Perks)
	d.Direct = ab.Interactions["direct"]
	d.DirectPerksTable = perkTable(d.Direct.Perks)
	d.Ranged = ab.Interactions["ranged"]
	d.RangedPerksTable = perkTable(d.Ranged.Perks)
	d.Area = ab.Interactions["area"]
	d.AreaPerksTable = perkTable(d.Area.Perks)
	d.AreaOfEffect = ab.Interactions["area_of_effect"]
	d.AreaOfEffectPerksTable = perkTable(d.AreaOfEffect.Perks)

	d.EngagementModesTable = engagementModesTable(ab.Validations.Engagement.Modes)
	d.CounterTypesTable = counterTypesTable(ab.Validations.Counter.Types)
	d.TierShiftsTable = tierShiftsTable(ab.Validations.Counter.TierShifts)

	return d
}

// perkTable renders a markdown table for a list of perks.
func perkTable(perks []config.PerkConfig) template.HTML {
	return costTable("Perk", perks)
}

// costTable renders a generic perk-like table with Description, Energy Cost, Add Cost columns.
func costTable(what string, rows []config.PerkConfig) template.HTML {
	if len(rows) == 0 {
		return template.HTML("_No " + strings.ToLower(what) + "s available._")
	}
	var b strings.Builder
	b.WriteString("| Description | Energy Cost | Add Cost |\n")
	b.WriteString("| --- | --- | --- |\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "| %s | %+d | %+d |\n", r.Description, r.EnergyCost, r.AddCost)
	}
	return template.HTML(strings.TrimSpace(b.String()))
}

// triggerTable renders a markdown table for triggers.
func triggerTable(triggers []config.TriggerConfig) template.HTML {
	if len(triggers) == 0 {
		return template.HTML("_No triggers available._")
	}
	var b strings.Builder
	b.WriteString("| Description | Add Cost |\n")
	b.WriteString("| --- | --- |\n")
	for _, t := range triggers {
		fmt.Fprintf(&b, "| %s | %+d |\n", t.Description, t.AddCost)
	}
	return template.HTML(strings.TrimSpace(b.String()))
}

// knockoutTable renders a markdown table for knockout requirements.
func knockoutTable(kos []config.KnockoutConfig) template.HTML {
	if len(kos) == 0 {
		return template.HTML("_No knockout requirements available._")
	}
	var b strings.Builder
	b.WriteString("| Description | Add Cost |\n")
	b.WriteString("| --- | --- |\n")
	for _, k := range kos {
		fmt.Fprintf(&b, "| %s | %+d |\n", k.Description, k.AddCost)
	}
	return template.HTML(strings.TrimSpace(b.String()))
}

// enactmentTable renders a markdown table for compatible enactment types.
func enactmentTable(types []string, enactments map[string]config.EnactmentConfig) template.HTML {
	if len(types) == 0 {
		return template.HTML("_No compatible enactments._")
	}
	var b strings.Builder
	b.WriteString("| Description | Energy Cost | Add Cost |\n")
	b.WriteString("| --- | --- | --- |\n")
	for _, t := range types {
		cfg := lookupEnactment(t, enactments)
		fmt.Fprintf(&b, "| %s | %+d | %+d |\n", t, cfg.BaseCost.EnergyCost, cfg.BaseCost.AddCost)
	}
	return template.HTML(strings.TrimSpace(b.String()))
}

// effectTable renders a markdown table for persistent effect nested options.
func effectTable(effects []config.EffectConfig) template.HTML {
	if len(effects) == 0 {
		return template.HTML("_No effect options available._")
	}
	var b strings.Builder
	b.WriteString("| Description | Energy Cost | Add Cost |\n")
	b.WriteString("| --- | --- | --- |\n")
	for _, e := range effects {
		fmt.Fprintf(&b, "| %s | %+d | %+d |\n", e.Description, e.EnergyCost, e.AddCost)
	}
	return template.HTML(strings.TrimSpace(b.String()))
}

// engagementModesTable renders a markdown table for validation engagement modes.
func engagementModesTable(modes []config.PerkConfig) template.HTML {
	if len(modes) == 0 {
		return template.HTML("_No engagement modes available._")
	}
	var b strings.Builder
	b.WriteString("| Description | Energy Cost | Add Cost |\n")
	b.WriteString("| --- | --- | --- |\n")
	for _, m := range modes {
		fmt.Fprintf(&b, "| %s | %+d | %+d |\n", m.Description, m.EnergyCost, m.AddCost)
	}
	return template.HTML(strings.TrimSpace(b.String()))
}

// counterTypesTable renders a markdown table for validation counter types.
func counterTypesTable(types []config.PerkConfig) template.HTML {
	if len(types) == 0 {
		return template.HTML("_No counter types available._")
	}
	var b strings.Builder
	b.WriteString("| Description | Energy Cost | Add Cost |\n")
	b.WriteString("| --- | --- | --- |\n")
	for _, t := range types {
		fmt.Fprintf(&b, "| %s | %+d | %+d |\n", t.Description, t.EnergyCost, t.AddCost)
	}
	return template.HTML(strings.TrimSpace(b.String()))
}

// tierShiftsTable renders a markdown table for validation tier shifts.
func tierShiftsTable(shifts []config.PerkConfig) template.HTML {
	if len(shifts) == 0 {
		return template.HTML("_No tier shifts available._")
	}
	var b strings.Builder
	b.WriteString("| Description | Energy Cost | Add Cost |\n")
	b.WriteString("| --- | --- | --- |\n")
	for _, s := range shifts {
		fmt.Fprintf(&b, "| %s | %+d | %+d |\n", s.Description, s.EnergyCost, s.AddCost)
	}
	return template.HTML(strings.TrimSpace(b.String()))
}

// lookupEnactment finds an EnactmentConfig by its display type name.
func lookupEnactment(name string, enactments map[string]config.EnactmentConfig) config.EnactmentConfig {
	key := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(name, "Enact ", ""), " ", "_"))
	if cfg, ok := enactments[key]; ok {
		return cfg
	}
	for _, cfg := range enactments {
		if cfg.Type == name {
			return cfg
		}
	}
	return config.EnactmentConfig{}
}
