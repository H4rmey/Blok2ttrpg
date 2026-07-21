package model

// Character is fully generic: all identity/vital fields live in Attributes,
// all skills live in Traits, keyed by the ids the config defines. This is what
// lets config authors add or remove character attributes without any code
// change.
type Character struct {
	ID    string `json:"id"`
	Level int    `json:"level"`

	// Attributes maps a config field key to its stored value. Values are
	// strings/numbers/bools depending on the field type.
	Attributes map[string]any `json:"attributes"`

	// Traits maps "<group_id>.<trait_name>" to a proficiency id.
	Traits map[string]string `json:"traits"`

	Abilities []Ability `json:"abilities"`
}

// Name returns a display name, falling back to the id.
func (c *Character) Name() string {
	if c.Attributes != nil {
		if v, ok := c.Attributes["name"]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return c.ID
}

// Attr returns a stored attribute value (or nil).
func (c *Character) Attr(key string) any {
	if c.Attributes == nil {
		return nil
	}
	return c.Attributes[key]
}

// TraitKey builds the composite key used to store a trait proficiency.
func TraitKey(groupID, trait string) string { return groupID + "." + trait }

// Ability is a built ability. Its structured data lives generically in Fields
// and its attached enactments; there are no hardcoded ability-type fields.
type Ability struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	// Type is an ability-type component id from the config.
	Type string `json:"type"`

	// Fields holds the ability-type-level field values.
	Fields map[string]any `json:"fields,omitempty"`

	Enactments []Enactment `json:"enactments,omitempty"`
}

// Enactment is one effect attached to an ability. Type is an enactment
// component id; Interaction is an optional interaction component id.
type Enactment struct {
	Type            string         `json:"type"`
	Fields          map[string]any `json:"fields,omitempty"`
	Interaction     string         `json:"interaction,omitempty"`
	InteractionData map[string]any `json:"interaction_data,omitempty"`
	// ValidationData holds the engagement/counter (validation) field values.
	ValidationData map[string]any `json:"validation_data,omitempty"`
}
