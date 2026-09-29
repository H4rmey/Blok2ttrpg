// Backwards compatibility for characters saved before the trait/skill rename.
//
// The rename swapped two names at once:
//
//	old "traits"     (dice-backed roster) -> new "skills"
//	old "attributes" (sheet sections)     -> new "traits"
//
// Because "traits" means something different on either side of the rename, a
// stored file cannot be interpreted correctly by looking at one key alone: a
// document containing "traits" is a legacy document only when it also carries
// the legacy "attributes" key (or carries no "skills" key). That test is what
// distinguishes the two layouts, and it is why this is a custom unmarshaller
// rather than a pair of extra struct tags.
//
// Migration happens on read, in memory, so no data file is rewritten until the
// character is saved again through the normal path. That makes it idempotent and
// non-destructive: reading a file twice yields the same result, and a file that
// is never edited is never touched.
package model

import "encoding/json"

// characterJSON mirrors Character for JSON purposes and additionally accepts the
// two legacy key names. The legacy fields are pointers so an absent key can be
// told apart from a present-but-empty one, which is what the layout test needs.
type characterJSON struct {
	ID    string `json:"id"`
	Level int    `json:"level"`

	Traits map[string]any    `json:"traits"`
	Skills map[string]string `json:"skills"`

	// LegacyAttributes is the pre-rename name of what is now Traits.
	LegacyAttributes map[string]any `json:"attributes"`

	Perks    []Perk             `json:"perks"`
	Packages []InstalledPackage `json:"packages,omitempty"`

	// Conditions needs an explicit entry here for the same reason every other
	// field does: this struct fully replaces Character for decoding purposes, so
	// a field missing from it is silently dropped on read. Nothing about it is
	// legacy - it predates no rename - it just has to be listed.
	Conditions []AppliedCondition `json:"conditions,omitempty"`
}

// UnmarshalJSON decodes a character, migrating the pre-rename key layout when it
// is detected.
func (c *Character) UnmarshalJSON(data []byte) error {
	var raw characterJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	c.ID = raw.ID
	c.Level = raw.Level
	c.Perks = raw.Perks
	c.Packages = raw.Packages
	c.Conditions = raw.Conditions

	// A legacy document is one that carries the old "attributes" key. In that
	// layout "traits" holds the dice-backed roster, so it becomes Skills and
	// "attributes" becomes Traits.
	if raw.LegacyAttributes != nil {
		c.Traits = raw.LegacyAttributes
		c.Skills = stringMap(raw.Traits)
	} else {
		c.Traits = raw.Traits
		c.Skills = raw.Skills
	}

	if c.Traits == nil {
		c.Traits = map[string]any{}
	}
	if c.Skills == nil {
		c.Skills = map[string]string{}
	}
	return nil
}

// stringMap narrows the legacy trait map (decoded as map[string]any because it
// shares a key with the new Traits field) to the proficiency ids Skills holds.
// Non-string values are skipped rather than coerced: a proficiency id is always
// a string, so anything else is corrupt and is better dropped to the default
// than silently turned into one.
func stringMap(in map[string]any) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}
	return out
}
