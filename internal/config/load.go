package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"gopkg.in/yaml.v3"
)

var profileIDPattern = regexp.MustCompile(`^[a-z0-9_-]+$`)

// Loaded is a Config together with the directory it was loaded from. The
// directory is used to resolve relative paths such as documentation files.
type Loaded struct {
	*Config
	Dir string
}

// Load reads a ruleset. If path is a directory, every *.yaml / *.yml file in it
// is merged into a single Config (in filename order). If path is a single file,
// that file is parsed directly.
func Load(path string) (*Loaded, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat config path %q: %w", path, err)
	}

	var cfg Config
	dir := path
	if info.IsDir() {
		if err := loadDir(path, &cfg); err != nil {
			return nil, err
		}
	} else {
		if err := loadFile(path, &cfg); err != nil {
			return nil, err
		}
		dir = filepath.Dir(path)
	}

	if cfg.Title == "" {
		cfg.Title = "Blok2 TTRPG"
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return &Loaded{Config: &cfg, Dir: dir}, nil
}

func loadDir(dir string, cfg *Config) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading config dir: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := filepath.Ext(e.Name())
		if ext == ".yaml" || ext == ".yml" {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		return fmt.Errorf("no yaml files found in %q", dir)
	}
	for _, name := range names {
		if err := loadFile(filepath.Join(dir, name), cfg); err != nil {
			return err
		}
	}
	return nil
}

// loadFile decodes one YAML file into cfg. Each file contributes the sections
// it defines. KnownFields is on, so any key not modelled by the schema is a
// hard error; this keeps the config from accumulating dead/misspelled keys.
func loadFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %q: %w", path, err)
	}
	var incoming Config
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)

	if err := dec.Decode(&incoming); err != nil {
		return fmt.Errorf("parsing %q: %w", path, err)
	}
	merge(cfg, &incoming)
	return nil
}

// merge folds incoming into base. Scalars overwrite when non-zero; ordered maps
// and slices accumulate so each section file owns its part of the ruleset.
func merge(base, in *Config) {
	if in.Version != 0 {
		base.Version = in.Version
	}
	if in.ProfileID != "" {
		base.ProfileID = in.ProfileID
	}
	if in.Title != "" {
		base.Title = in.Title
	}
	if in.Combat.Actions.Amount != 0 {
		base.Combat = in.Combat
	}
	// The allow_negative_* flags are tri-state pointers: only a section file
	// that actually sets one overwrites the base value, so an unset flag keeps
	// whatever an earlier file declared (and ultimately defaults to false).
	if in.AllowNegativeBuildCost != nil {
		base.AllowNegativeBuildCost = in.AllowNegativeBuildCost
	}
	if in.AllowNegativeEnergyCost != nil {
		base.AllowNegativeEnergyCost = in.AllowNegativeEnergyCost
	}
	if in.AllowNegativeSkillPoints != nil {
		base.AllowNegativeSkillPoints = in.AllowNegativeSkillPoints
	}
	if (in.AdditionalEnactment != AdditionalEnactment{}) {
		base.AdditionalEnactment = in.AdditionalEnactment
	}
	if len(in.Dice.Damage) > 0 || len(in.Dice.Generic) > 0 {
		base.Dice = in.Dice
	}
	if len(in.Validations.Fields) > 0 {
		base.Validations = in.Validations
	}
	if len(in.OptionSources) > 0 {
		if base.OptionSources == nil {
			base.OptionSources = map[string]OptionList{}
		}
		for k, v := range in.OptionSources {
			base.OptionSources[k] = v
		}
	}
	if len(in.OptionSourcesCosted) > 0 {
		if base.OptionSourcesCosted == nil {
			base.OptionSourcesCosted = map[string]OptionList{}
		}
		for k, v := range in.OptionSourcesCosted {
			base.OptionSourcesCosted[k] = v
		}
	}
	if len(in.OptionGroups) > 0 {
		if base.OptionGroups == nil {
			base.OptionGroups = map[string]OptionGroupDef{}
		}
		for k, v := range in.OptionGroups {
			base.OptionGroups[k] = v
		}
	}
	if len(in.SkillCategories) > 0 {
		base.SkillCategories = in.SkillCategories
	}
	if in.VitalGroup != "" {
		base.VitalGroup = in.VitalGroup
	}

	if in.DefaultProficiency != "" {
		base.DefaultProficiency = in.DefaultProficiency
	}

	mergeTraitMap(&base.Traits, in.Traits)
	mergeSkillMap(&base.Skills, in.Skills)
	mergeComponentMap(&base.AbilityTypes, in.AbilityTypes)
	mergeComponentMap(&base.Enactments, in.Enactments)
	mergeComponentMap(&base.Interactions, in.Interactions)

	base.Proficiencies = append(base.Proficiencies, in.Proficiencies...)
	base.FileOrder = append(base.FileOrder, in.FileOrder...)

	if levelingConfigured(in.Leveling) {
		base.Leveling = in.Leveling
	}

	// The invoking and negotiation sections each live in a single file, so a
	// whole-block overwrite is enough; the guards keep a file that does not
	// mention them from blanking what another file declared.
	if invokingConfigured(in.Invoking) {
		base.Invoking = in.Invoking
	}
	if negotiationConfigured(in.Negotiation) {
		base.Negotiation = in.Negotiation
	}

	if (in.AdditionalCondition != Cost{}) {
		base.AdditionalCondition = in.AdditionalCondition
	}
	base.Conditions = append(base.Conditions, in.Conditions...)
	base.GeneralConditions = append(base.GeneralConditions, in.GeneralConditions...)
	base.SpecificConditions = append(base.SpecificConditions, in.SpecificConditions...)

}

// levelingConfigured reports whether a section file actually declared a
// leveling block. A block counts as configured when it sets max_level or when
// either pool supplies a formula (start/per_level) or an explicit level table.
func levelingConfigured(l Leveling) bool {
	if l.MaxLevel != 0 {
		return true
	}
	for _, t := range []LevelTable{l.SkillPoints, l.AbilityPoints} {
		if t.Start != 0 || t.PerLevel != 0 || len(t.Levels) > 0 {
			return true
		}
	}
	return false
}

func mergeComponentMap(base *ComponentMap, in ComponentMap) {
	// Map-level help text is carried on the map itself, not on a component, so
	// it must be copied across even when the incoming map declares no ordered
	// components. Non-empty incoming values overwrite the base.
	if in.Information != "" {
		base.Information = in.Information
	}
	if in.RenderInformation {
		base.RenderInformation = in.RenderInformation
	}
	if len(in.Order) == 0 {
		return
	}
	if base.Items == nil {

		base.Items = map[string]*Component{}
	}
	for _, k := range in.Order {
		if _, seen := base.Items[k]; !seen {
			base.Order = append(base.Order, k)
		}
		base.Items[k] = in.Items[k]
	}
}

func mergeTraitMap(base *TraitMap, in TraitMap) {
	if len(in.Order) == 0 {
		return
	}
	if base.Items == nil {
		base.Items = map[string]*TraitGroup{}
	}
	for _, k := range in.Order {
		if _, seen := base.Items[k]; !seen {
			base.Order = append(base.Order, k)
		}
		base.Items[k] = in.Items[k]
	}
}

func mergeSkillMap(base *SkillMap, in SkillMap) {
	if len(in.Order) == 0 {
		return
	}
	if base.Items == nil {
		base.Items = map[string][]string{}
	}
	for _, k := range in.Order {
		if _, seen := base.Items[k]; !seen {
			base.Order = append(base.Order, k)
		}
		base.Items[k] = in.Items[k]
	}
}
