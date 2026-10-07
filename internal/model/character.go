package model

// Character is fully generic: all identity/vital fields live in Traits,
// all skills live in Skills, keyed by the ids the config defines. This is what
// lets config authors add or remove character traits without any code
// change.
type Character struct {
	ID    string `json:"id"`
	Level int    `json:"level"`

	// Traits maps a config field key to its stored value. Values are
	// strings/numbers/bools depending on the field type.
	Traits map[string]any `json:"traits"`

	// Skills maps "<group_id>.<skill_name>" to a proficiency id.
	Skills map[string]string `json:"skills"`

	Perks []Perk `json:"perks"`

	// Conditions are the conditions currently applied to the character. They are
	// play state, not build state: nothing here costs points and nothing here is
	// written into Skills. The stored proficiency in Skills always remains the
	// character's own, and a condition's effect is computed as an overlay at
	// render time (see engine.EffectiveSkills).
	//
	// Keeping them separate is what makes removal exact. Writing shifts into
	// Skills instead would mean a condition applied before a package toggle and
	// removed after it would subtract from a different baseline than it added
	// to, silently corrupting the character - the same trap InstalledPackage.
	// Shifts exists to avoid.
	Conditions []AppliedCondition `json:"conditions,omitempty"`

	// Shifts are the hand-applied "Enact Shift" play cards: one card moves one
	// skill, at a magnitude the player picks. Like Conditions they are play
	// state stored as a derived overlay (see engine.SkillShifts), so removing a
	// card is exact by construction and several cards on the same skill simply
	// stack.
	Shifts []AppliedShift `json:"shifts,omitempty"`

	// Packages lists the currently installed content packages. Each records
	// exactly what it applied (proficiency shifts) so removal is precise and
	// reversible even when multiple packages stack shifts on the same skill.
	Packages []InstalledPackage `json:"packages,omitempty"`
}

// InstalledPackage is the record of a package imported onto a character. It is
// the source of truth for undoing a package: Shifts holds the proficiency
// deltas that were applied, and perks added by the package carry a matching
// PackageID so they can be removed together.
type InstalledPackage struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Shifts map[string]int `json:"shifts,omitempty"`

	// Toggleable marks a package whose effects can be turned on and off by the
	// user (e.g. items). Class, race, and background packages are permanent and
	// are not toggleable.
	Toggleable bool `json:"toggleable,omitempty"`

	// Enabled reports whether a toggleable package's effects are currently
	// applied. Non-toggleable packages are always enabled. When a toggleable
	// package is disabled its proficiency shifts are reversed and its perks
	// are removed, but the record is kept so it can be re-enabled later.
	Enabled bool `json:"enabled"`
}

// AppliedCondition is one condition currently affecting a character.
type AppliedCondition struct {
	// ID is the config condition id.
	ID string `json:"id"`

	// Shift is the magnitude the player chose, for a condition whose config
	// declares a shift range. It is ignored for every other condition: a fixed
	// condition's magnitude comes from its own config, and a condition that
	// affects no skills has no magnitude at all.
	Shift int `json:"shift,omitempty"`

	// Note is free text for whatever the rules attach to this instance but the
	// app does not model: the remaining duration, the solution the perk set, or
	// which target a Taunt points at. It exists because the alternative is a
	// table forgetting to clear a condition that never had anywhere to record
	// when it ends.
	Note string `json:"note,omitempty"`
}

// AppliedShift is one hand-applied shift card: a temporary adjustment to a
// single skill, created from the character sheet's Apply Shift button. Unlike a
// condition it names its skill directly rather than going through a config
// entry, and its magnitude comes from the Enact Shift rules (a player choice of
// the configured range).
type AppliedShift struct {
	// SkillKey is the "<group>.<skill>" key it moves, picking the ruleset's
	// composite skill key for the picker (see model.SkillKey). A key the config
	// no longer defines is tolerated on read and rendered as removable.
	SkillKey string `json:"skill_key"`

	// Shift is the magnitude the player chose (signed; the config's Enact Shift
	// range currently allows -6..6 without 0).
	Shift int `json:"shift,omitempty"`
}

// Name returns a display name, falling back to the id.
func (c *Character) Name() string {
	if c.Traits != nil {
		if v, ok := c.Traits["name"]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return c.ID
}

// Attr returns a stored trait value (or nil).
func (c *Character) Attr(key string) any {
	if c.Traits == nil {
		return nil
	}
	return c.Traits[key]
}

// SkillKey builds the composite key used to store a skill proficiency.
func SkillKey(groupID, skill string) string { return groupID + "." + skill }

// Perk is a built perk. Its structured data lives generically in Fields
// and its attached enactments; there are no hardcoded perk-type fields.
type Perk struct {
	ID          string `json:"id" yaml:"id,omitempty"`
	Name        string `json:"name" yaml:"name,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Type is an perk-type component id from the config.
	Type string `json:"type" yaml:"type,omitempty"`

	// Tags are free-form labels set by the perk author. Deliberately not
	// validated against any list: there is no tag vocabulary in the config, so a
	// tag is whatever a perk file says it is and the library groups by whatever
	// tags it finds. Adding a label to a file makes it appear; removing its last
	// use makes it disappear. Nothing in the code reads a specific tag value.
	//
	// The tradeoff is that a misspelled tag silently becomes a new group rather
	// than an error. cmd/libaudit reports per-tag counts so a typo shows up as a
	// group of one.
	Tags []string `json:"tags,omitempty" yaml:"tags,omitempty"`

	// Fields holds the perk-type-level field values.
	Fields map[string]any `json:"fields,omitempty" yaml:"fields,omitempty"`

	Enactments []Enactment `json:"enactments,omitempty" yaml:"enactments,omitempty"`

	// PackageID, when set, records the package this perk was imported from.
	// It is used only for package removal: deleting a package removes every
	// perk tagged with its id. Editing the perk never touches the
	// package definition, so the tag stays purely for ownership tracking.
	PackageID string `json:"package_id,omitempty" yaml:"package_id,omitempty"`
}

// Enactment is one effect attached to an perk. Type is an enactment
// component id; Interaction is an optional interaction component id.
type Enactment struct {
	Type        string         `json:"type" yaml:"type,omitempty"`
	Fields      map[string]any `json:"fields,omitempty" yaml:"fields,omitempty"`
	Interaction string         `json:"interaction,omitempty" yaml:"interaction,omitempty"`
	// Explicit yaml tags are required on the multi-word keys: yaml.v3 lowercases
	// the Go field name by default ("interactiondata") and would silently ignore
	// the "interaction_data" key used in the library and export files, which
	// made interaction and validation costs vanish on import.
	InteractionData map[string]any `json:"interaction_data,omitempty" yaml:"interaction_data,omitempty"`
	// ValidationData holds the engagement/counter (validation) field values.
	ValidationData map[string]any `json:"validation_data,omitempty" yaml:"validation_data,omitempty"`

	// NewTarget marks an enactment beyond the first that picks its own target
	// instead of inheriting the target of the enactment before it. The first
	// enactment always has its own target, so the flag is ignored there. When
	// set, the enactment shows and pays for its own Interaction and Validation
	// plus the configured additional_enactment.new_target surcharge; when
	// unset, it reuses the previous enactment's target and pays for neither.
	NewTarget bool `json:"new_target,omitempty" yaml:"new_target,omitempty"`
}
