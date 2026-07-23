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
			base.OptionSources = map[string][]string{}
		}
		for k, v := range in.OptionSources {
			base.OptionSources[k] = v
		}
	}
	if len(in.OptionSourcesCosted) > 0 {
		if base.OptionSourcesCosted == nil {
			base.OptionSourcesCosted = map[string][]Option{}
		}
		for k, v := range in.OptionSourcesCosted {
			base.OptionSourcesCosted[k] = v
		}
	}
	if len(in.TraitCategories) > 0 {
		base.TraitCategories = in.TraitCategories
	}

	mergeAttributeMap(&base.Attributes, in.Attributes)
	mergeTraitMap(&base.Traits, in.Traits)
	mergeComponentMap(&base.AbilityTypes, in.AbilityTypes)
	mergeComponentMap(&base.Enactments, in.Enactments)
	mergeComponentMap(&base.Interactions, in.Interactions)

	base.Proficiencies = append(base.Proficiencies, in.Proficiencies...)
	base.FileOrder = append(base.FileOrder, in.FileOrder...)

	if in.Leveling.MaxLevel != 0 || len(in.Leveling.TraitPoints.Levels) > 0 || len(in.Leveling.AbilityPoints.Levels) > 0 {
		base.Leveling = in.Leveling
	}

	if (in.AdditionalState != Cost{}) {
		base.AdditionalState = in.AdditionalState
	}
	base.GeneralStates = append(base.GeneralStates, in.GeneralStates...)
	base.SpecificStates = append(base.SpecificStates, in.SpecificStates...)
}

func mergeComponentMap(base *ComponentMap, in ComponentMap) {
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

func mergeAttributeMap(base *AttributeMap, in AttributeMap) {
	if len(in.Order) == 0 {
		return
	}
	if base.Items == nil {
		base.Items = map[string]*AttributeGroup{}
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
		base.Items = map[string][]string{}
	}
	for _, k := range in.Order {
		if _, seen := base.Items[k]; !seen {
			base.Order = append(base.Order, k)
		}
		base.Items[k] = in.Items[k]
	}
}
