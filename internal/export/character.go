package export

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/harmey/blok2ttrpg/ability-builder/internal/models"
)

// CharacterToYAML serializes a full character (including all abilities) to a
// self-contained YAML document. Each ability is emitted inline using the same
// human-readable schema produced by ToYAML so a single ability file remains
// readable in isolation.
func CharacterToYAML(c *models.Character) string {
	var b strings.Builder

	b.WriteString("character:\n")

	if c.Name != "" {
		b.WriteString(fmt.Sprintf("  name: %q\n", c.Name))
	}
	if c.Level > 0 {
		b.WriteString(fmt.Sprintf("  level: %d\n", c.Level))
	}
	if c.Age != "" {
		b.WriteString(fmt.Sprintf("  age: %q\n", c.Age))
	}
	if c.Size != "" {
		b.WriteString(fmt.Sprintf("  size: %q\n", c.Size))
	}
	if c.Alignment != "" {
		b.WriteString(fmt.Sprintf("  alignment: %q\n", c.Alignment))
	}
	if c.Backstory != "" {
		b.WriteString(fmt.Sprintf("  backstory: %q\n", c.Backstory))
	}
	if c.Personality != "" {
		b.WriteString(fmt.Sprintf("  personality: %q\n", c.Personality))
	}
	if c.Appearance != "" {
		b.WriteString(fmt.Sprintf("  appearance: %q\n", c.Appearance))
	}
	if c.Hobbies != "" {
		b.WriteString(fmt.Sprintf("  hobbies: %q\n", c.Hobbies))
	}
	if c.Occupation != "" {
		b.WriteString(fmt.Sprintf("  occupation: %q\n", c.Occupation))
	}
	if c.Inventory != "" {
		b.WriteString(fmt.Sprintf("  inventory: %q\n", c.Inventory))
	}
	if c.Quirks != "" {
		b.WriteString(fmt.Sprintf("  quirks: %q\n", c.Quirks))
	}

	writeTraitMap(&b, "  general_traits", c.GeneralTraits)
	writeTraitMap(&b, "  offense_traits", c.OffenseTraits)
	writeTraitMap(&b, "  defense_traits", c.DefenseTraits)

	if c.VitalHP != "" {
		b.WriteString(fmt.Sprintf("  vital_hp: %s\n", c.VitalHP))
	}
	if c.VitalMovement != "" {
		b.WriteString(fmt.Sprintf("  vital_movement: %s\n", c.VitalMovement))
	}
	if c.VitalEnergy != "" {
		b.WriteString(fmt.Sprintf("  vital_energy: %s\n", c.VitalEnergy))
	}
	if c.CurrentHP > 0 {
		b.WriteString(fmt.Sprintf("  current_hp: %d\n", c.CurrentHP))
	}
	if c.CurrentEnergy > 0 {
		b.WriteString(fmt.Sprintf("  current_energy: %d\n", c.CurrentEnergy))
	}

	if len(c.Abilities) > 0 {
		b.WriteString("  abilities:\n")
		for i := range c.Abilities {
			abilityYAML := ToYAML(&c.Abilities[i])
			// Indent the ability block under the list item marker and strip
			// its top-level "ability:" header (we already emitted the list
			// marker "- ability:" implicitly via the indented body).
			indented := indentAbilityBlock(abilityYAML, "    ")
			indented = strings.TrimPrefix(indented, "    ability:\n")
			b.WriteString("    -\n")
			b.WriteString(indented)
			if !strings.HasSuffix(indented, "\n") {
				b.WriteString("\n")
			}
		}
	}

	return b.String()
}

func writeTraitMap(b *strings.Builder, key string, m map[string]models.Proficiency) {
	if len(m) == 0 {
		return
	}
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, k)
	}
	sort.Strings(names)
	b.WriteString(key + ":\n")
	for _, n := range names {
		b.WriteString(fmt.Sprintf("    %s: %s\n", n, m[n]))
	}
}

func indentAbilityBlock(block, indent string) string {
	lines := strings.Split(block, "\n")
	for i, l := range lines {
		if l != "" {
			lines[i] = indent + l
		}
	}
	return strings.Join(lines, "\n")
}

// characterYAML is the intermediate struct used to parse a full-character YAML
// document with gopkg.in/yaml.v3. It mirrors the schema emitted by
// CharacterToYAML but is forgiving about missing fields.
type characterYAML struct {
	Character struct {
		Name           string            `yaml:"name"`
		Level          int               `yaml:"level"`
		Age            string            `yaml:"age"`
		Size           string            `yaml:"size"`
		Alignment      string            `yaml:"alignment"`
		Backstory      string            `yaml:"backstory"`
		Personality    string            `yaml:"personality"`
		Appearance     string            `yaml:"appearance"`
		Hobbies        string            `yaml:"hobbies"`
		Occupation     string            `yaml:"occupation"`
		Inventory      string            `yaml:"inventory"`
		Quirks         string            `yaml:"quirks"`
		GeneralTraits  map[string]string `yaml:"general_traits"`
		OffenseTraits  map[string]string `yaml:"offense_traits"`
		DefenseTraits  map[string]string `yaml:"defense_traits"`
		VitalHP        string            `yaml:"vital_hp"`
		VitalMovement  string            `yaml:"vital_movement"`
		VitalEnergy    string            `yaml:"vital_energy"`
		CurrentHP      int               `yaml:"current_hp"`
		CurrentEnergy  int               `yaml:"current_energy"`
		Abilities      []yaml.Node       `yaml:"abilities"`
	} `yaml:"character"`
}

// singleAbilityYAML is used to detect / import a single-ability file (the
// schema produced by ToYAML) so users can drop in an ability file and have it
// attached to a fresh character.
type singleAbilityYAML struct {
	Ability yaml.Node `yaml:"ability"`
}

// ParseCharacterYAML parses a YAML document that may be either a full
// character export or a bare single-ability export. When a bare ability is
// provided, a fresh character with sensible defaults is created and the
// ability is attached.
func ParseCharacterYAML(data []byte) (*models.Character, error) {
	var doc characterYAML
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}

	if doc.Character.Name == "" &&
		doc.Character.Level == 0 &&
		len(doc.Character.GeneralTraits) == 0 &&
		len(doc.Character.Abilities) == 0 {
		var single singleAbilityYAML
		if err := yaml.Unmarshal(data, &single); err != nil {
			return nil, fmt.Errorf("invalid YAML: %w", err)
		}
		if single.Ability.Kind != 0 {
			ability, err := decodeAbilityNode(single.Ability)
			if err != nil {
				return nil, err
			}
			c := models.NewCharacter(genID(), nil, nil, nil)
			c.Name = "Imported Character"
			if ability.Name != "" {
				c.Name = ability.Name
			}
			c.Abilities = []models.Ability{ability}
			return &c, nil
		}
		return nil, fmt.Errorf("YAML document is empty")
	}

	c := models.NewCharacter(genID(), nil, nil, nil)
	c.Name = doc.Character.Name
	c.Level = doc.Character.Level
	if c.Level < 1 {
		c.Level = 1
	}
	c.Age = doc.Character.Age
	c.Size = doc.Character.Size
	c.Alignment = doc.Character.Alignment
	c.Backstory = doc.Character.Backstory
	c.Personality = doc.Character.Personality
	c.Appearance = doc.Character.Appearance
	c.Hobbies = doc.Character.Hobbies
	c.Occupation = doc.Character.Occupation
	c.Inventory = doc.Character.Inventory
	c.Quirks = doc.Character.Quirks
	c.VitalHP = models.Proficiency(doc.Character.VitalHP)
	c.VitalMovement = models.Proficiency(doc.Character.VitalMovement)
	c.VitalEnergy = models.Proficiency(doc.Character.VitalEnergy)
	c.CurrentHP = doc.Character.CurrentHP
	c.CurrentEnergy = doc.Character.CurrentEnergy

	if c.GeneralTraits == nil {
		c.GeneralTraits = map[string]models.Proficiency{}
	}
	if c.OffenseTraits == nil {
		c.OffenseTraits = map[string]models.Proficiency{}
	}
	if c.DefenseTraits == nil {
		c.DefenseTraits = map[string]models.Proficiency{}
	}
	for k, v := range doc.Character.GeneralTraits {
		c.GeneralTraits[k] = models.Proficiency(v)
	}
	for k, v := range doc.Character.OffenseTraits {
		c.OffenseTraits[k] = models.Proficiency(v)
	}
	for k, v := range doc.Character.DefenseTraits {
		c.DefenseTraits[k] = models.Proficiency(v)
	}

	for _, node := range doc.Character.Abilities {
		ability, err := decodeAbilityNode(node)
		if err != nil {
			return nil, err
		}
		c.Abilities = append(c.Abilities, ability)
	}

	return &c, nil
}

// abilityYAMLSchema mirrors the schema emitted by ToYAML. We declare every
// field with the exact key names used by the exporter so round-trips are
// lossless.
type abilityYAMLSchema struct {
	Type             string                 `yaml:"type"`
	Name             string                 `yaml:"name"`
	Description      string                 `yaml:"description"`
	HasItemDep       string                 `yaml:"has_item_dependency"`
	ItemName         string                 `yaml:"item_name"`
	EnergyCost       int                    `yaml:"energy_cost"`
	ActionCost       int                    `yaml:"action_cost"`
	EnergyAdj        int                    `yaml:"energy_adjustment"`
	ActionAdj        int                    `yaml:"action_adjustment"`
	Range            string                 `yaml:"range"`
	Uses             int                    `yaml:"uses"`
	Trigger          string                 `yaml:"trigger"`
	PhaseDuration    int                    `yaml:"phase_duration"`
	ReverseDuration  int                    `yaml:"reverse_phase_duration"`
	AllKnockouts     bool                   `yaml:"all_knockout_requirements"`
	ReverseKO        bool                   `yaml:"knockout_on_reverse_phase"`
	Knockouts        []string               `yaml:"knockout_requirements"`
	Health           int                    `yaml:"health"`
	Lifetime         int                    `yaml:"lifetime"`
	Enactments       []yaml.Node            `yaml:"enactments"`
	Extra            map[string]interface{} `yaml:",inline"`
}

func decodeAbilityNode(node yaml.Node) (models.Ability, error) {
	// Re-marshal the node so we can decode it through the strict schema
	// alongside a flexible catch-all for fields we don't model directly.
	raw, err := yaml.Marshal(&node)
	if err != nil {
		return models.Ability{}, fmt.Errorf("failed to decode ability: %w", err)
	}

	var schema abilityYAMLSchema
	if err := yaml.Unmarshal(raw, &schema); err != nil {
		return models.Ability{}, fmt.Errorf("failed to decode ability: %w", err)
	}

	a := models.Ability{
		ID:          genID(),
		Name:        schema.Name,
		Description: schema.Description,
		Type:        models.AbilityType(schema.Type),
		ItemName:    schema.ItemName,
		EnergyCost:  schema.EnergyCost,
		ActionCost:  schema.ActionCost,
	}

	switch a.Type {
	case models.AbilityExecution:
		a.EnergySteps = schema.EnergyAdj
		a.ActionSteps = schema.ActionAdj
	case models.AbilityReaction:
		a.ReactionRange = parseRangeMeters(schema.Range)
		a.ReactionUses = schema.Uses
		if trigger, trait, ok := splitTriggerOfType(schema.Trigger); ok {
			a.Trigger = trigger
			a.TriggerTrait = trait
		} else {
			a.Trigger = schema.Trigger
		}
	case models.AbilityPhase:
		a.PhaseDuration = schema.PhaseDuration
		a.ReversePhaseRounds = schema.ReverseDuration
		a.AllKnockoutsReq = schema.AllKnockouts
		a.ReverseKnockoutOK = schema.ReverseKO
		a.Knockouts = schema.Knockouts
	case models.AbilityMinion:
		a.HPBonus = (schema.Health - 10) / 5
		if a.HPBonus < 0 {
			a.HPBonus = 0
		}
		a.ExtraLifetime = schema.Lifetime - 3
		if a.ExtraLifetime < 0 {
			a.ExtraLifetime = 0
		}
	}

	switch normalizeYesNo(schema.HasItemDep) {
	case "yes":
		a.HasItemDependency = true
	case "no":
		a.HasItemDependency = false
	}

	for _, enNode := range schema.Enactments {
		en, err := decodeEnactmentNode(enNode)
		if err != nil {
			return a, err
		}
		a.Enactments = append(a.Enactments, en)
	}

	return a, nil
}

func decodeEnactmentNode(node yaml.Node) (models.Enactment, error) {
	raw, err := yaml.Marshal(&node)
	if err != nil {
		return models.Enactment{}, fmt.Errorf("failed to decode enactment: %w", err)
	}

	var schema struct {
		Type           string       `yaml:"type"`
		Always         bool         `yaml:"always_resolve"`
		BuildCost      int          `yaml:"build_cost"`
		CastCost       int          `yaml:"cast_cost"`
		Formula        string       `yaml:"formula"`
		Source         string       `yaml:"source"`
		Trait          string       `yaml:"trait"`
		OtherText      string       `yaml:"other_text"`
		FlatBonus      int          `yaml:"flat_bonus"`
		OffensiveTrait string       `yaml:"offensive_trait"`
		Medicine       bool         `yaml:"medicine"`
		Origin         string       `yaml:"origin"`
		Distance       int          `yaml:"distance"`
		Directions     []string     `yaml:"directions"`
		ShiftedTrait   string       `yaml:"shifted_trait"`
		Direction      string       `yaml:"direction"`
		Amount         int          `yaml:"amount"`
		ShiftUses      int          `yaml:"uses"`
		EffectName     string       `yaml:"name"`
		Applies        string       `yaml:"applies"`
		Duration       int          `yaml:"duration"`
		Trigger        string       `yaml:"trigger"`
		Solutions      []string     `yaml:"solutions"`
		Interaction    yaml.Node    `yaml:"interaction"`
		Interactions   []yaml.Node  `yaml:"interactions"`
	}
	if err := yaml.Unmarshal(raw, &schema); err != nil {
		return models.Enactment{}, fmt.Errorf("failed to decode enactment: %w", err)
	}

	e := models.Enactment{
		Type:            models.EnactmentType(schema.Type),
		BuildCost:       schema.BuildCost,
		CastCost:        schema.CastCost,
		Formula:         schema.Formula,
		Always:          schema.Always,
		FlatBonus:       schema.FlatBonus,
		OffensiveTrait:  schema.OffensiveTrait,
		MedicineTrait:   schema.ShiftedTrait, // unused; kept for symmetry
		Distance:        schema.Distance,
		Directions:      schema.Directions,
		ShiftedTrait:    schema.ShiftedTrait,
		ShiftDir:        schema.Direction,
		ShiftAmount:     schema.Amount,
		ShiftUses:       schema.ShiftUses,
		EffectName:      schema.EffectName,
		EffectType:      schema.Applies,
		Duration:        schema.Duration,
		TriggerTiming:   schema.Trigger,
		Solutions:       schema.Solutions,
	}

	if schema.Source != "" {
		e.Source, e.SourceTrait, e.OtherRollText = decodeSource(schema.Source, schema.Trait, schema.OtherText)
	}
	if schema.Origin != "" && schema.Origin != "Engager" {
		e.OriginMode = "other"
		e.OriginText = schema.Origin
	}
	if schema.Medicine {
		e.MedicineTrait = "Medicine"
	}

	if schema.Interaction.Kind != 0 {
		inter, err := decodeInteractionNode(schema.Interaction)
		if err != nil {
			return e, err
		}
		e.Interaction = &inter
	}

	return e, nil
}

func decodeInteractionNode(node yaml.Node) (models.Interaction, error) {
	raw, err := yaml.Marshal(&node)
	if err != nil {
		return models.Interaction{}, fmt.Errorf("failed to decode interaction: %w", err)
	}
	var schema struct {
		Type         string      `yaml:"type"`
		UsePrevious  bool        `yaml:"use_previous"`
		BuildCost    int         `yaml:"build_cost"`
		CastCost     int         `yaml:"cast_cost"`
		Targets      int         `yaml:"targets"`
		Range        string      `yaml:"range"`
		Radius       int         `yaml:"radius"`
		Origin       string      `yaml:"origin"`
		Duration     int         `yaml:"duration"`
		Timing       string      `yaml:"timing"`
		Immune       bool        `yaml:"engager_immune"`
		VisibleOK    interface{} `yaml:"target_may_be_not_visible"`
		ObstructedOK interface{} `yaml:"target_may_be_obstructed"`
		Remove       interface{} `yaml:"remove_engagement_penalty"`
		Validation   yaml.Node   `yaml:"validation"`
	}
	if err := yaml.Unmarshal(raw, &schema); err != nil {
		return models.Interaction{}, fmt.Errorf("failed to decode interaction: %w", err)
	}

	i := models.Interaction{
		Type:        models.InteractionType(schema.Type),
		BuildCost:   schema.BuildCost,
		CastCost:    schema.CastCost,
		Targets:     schema.Targets,
		Range:       parseRangeMeters(schema.Range),
		Radius:      schema.Radius,
		Duration:    schema.Duration,
		Timing:      schema.Timing,
		Immune:      schema.Immune,
		UsePrevious: schema.UsePrevious,
	}
	if schema.Origin != "" && schema.Origin != "Engager" {
		i.OriginMode = "other"
		i.OriginText = schema.Origin
	}
	i.VisibleOK = coerceBool(schema.VisibleOK)
	i.ObstructedOK = coerceBool(schema.ObstructedOK)
	i.RemovePenalty = coerceBool(schema.Remove)

	if schema.Validation.Kind == yaml.MappingNode {
		v, err := decodeValidationNode(schema.Validation)
		if err != nil {
			return i, err
		}
		i.Validation = &v
	}

	return i, nil
}

func decodeValidationNode(node yaml.Node) (models.Validation, error) {
	raw, err := yaml.Marshal(&node)
	if err != nil {
		return models.Validation{}, fmt.Errorf("failed to decode validation: %w", err)
	}
	var schema struct {
		BuildCost    int      `yaml:"build_cost"`
		CastCost     int      `yaml:"cast_cost"`
		EngageRoll   string   `yaml:"engagement_roll"`
		CounterRoll  string   `yaml:"counter_roll"`
		CounterList  []string `yaml:"counter_roll_list"`
	}
	if err := yaml.Unmarshal(raw, &schema); err != nil {
		return models.Validation{}, fmt.Errorf("failed to decode validation: %w", err)
	}

	v := models.Validation{
		BuildCost: schema.BuildCost,
		CastCost:  schema.CastCost,
	}
	v.EngageMode, v.EngageTrait, v.EngageDie, v.EngageOther, v.EngageTraitCategory = decodeEngageRoll(schema.EngageRoll)
	if schema.CounterRoll != "" {
		v.CounterRolls = []string{schema.CounterRoll}
	} else if len(schema.CounterList) > 0 {
		v.CounterRolls = schema.CounterList
	}
	return v, nil
}

func decodeEngageRoll(s string) (mode models.EngageMode, trait, die, other string, cat models.TraitCategory) {
	lower := strings.ToLower(strings.TrimSpace(s))
	switch {
	case strings.HasPrefix(lower, "result of previous"):
		return models.EngageModePrevious, "", "", "", ""
	case strings.HasSuffix(lower, "(generic)"):
		return models.EngageModeGeneric, "", strings.TrimSpace(strings.TrimSuffix(s, "(generic)")), "", ""
	case strings.HasSuffix(lower, "(offensive trait)"),
		strings.HasSuffix(lower, "(general trait)"),
		strings.HasSuffix(lower, "(defensive trait)"):
		open := strings.LastIndex(s, "(")
		if open < 0 {
			return models.EngageModeTrait, s, "", "", ""
		}
		trait = strings.TrimSpace(s[:open])
		category := strings.TrimSuffix(strings.TrimPrefix(s[open:], "("), ")")
		switch category {
		case "general trait":
			cat = models.TraitCategoryGeneral
		case "defensive trait":
			cat = models.TraitCategoryDefense
		default:
			cat = models.TraitCategoryOffense
		}
		return models.EngageModeTrait, trait, "", "", cat
	}
	if lower == "" {
		return "", "", "", "", ""
	}
	if strings.HasPrefix(lower, "1d") || strings.HasPrefix(lower, "d") {
		return models.EngageModeGeneric, "", s, "", ""
	}
	return models.EngageModeOther, "", "", s, ""
}

func decodeSource(source, trait, otherText string) (string, string, string) {
	trimmed := strings.TrimSpace(source)
	switch trimmed {
	case "1d4":
		return "d4", "", ""
	case "1d10 (trait)":
		return "trait", trait, ""
	case "another roll result":
		return "other", "", otherText
	}
	if strings.HasPrefix(trimmed, "1d") {
		return strings.TrimPrefix(trimmed, "1"), "", ""
	}
	return trimmed, "", ""
}

func coerceBool(v interface{}) bool {
	switch x := v.(type) {
	case bool:
		return x
	case string:
		b, _ := yaml.Marshal(x)
		return strings.TrimSpace(strings.ToLower(strings.TrimSpace(string(b)))) == "true"
	case nil:
		return false
	}
	return false
}

func parseRangeMeters(s string) int {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.TrimSuffix(s, "m")
	s = strings.TrimSuffix(s, " meters")
	s = strings.TrimSuffix(s, " meter")
	if s == "" {
		return 0
	}
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

func splitTriggerOfType(s string) (string, string, bool) {
	idx := strings.Index(s, " of type ")
	if idx < 0 {
		return "", "", false
	}
	return strings.TrimSpace(s[:idx]), strings.TrimSpace(s[idx+len(" of type "):]), true
}

func normalizeYesNo(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "yes", "true", "y", "1":
		return "yes"
	case "no", "false", "n", "0":
		return "no"
	}
	return ""
}

func genID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
