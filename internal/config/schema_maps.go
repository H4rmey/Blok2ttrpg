// This file implements the ordered, id-keyed YAML collections. Each preserves
// the mapping order from the config file so ruleset authors control the order
// things are presented in, which a plain Go map would discard.
package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// ComponentMap is an ordered, id-keyed collection of components. YAML mapping
// order is preserved so config authors control presentation order.
type ComponentMap struct {
	Order []string
	Items map[string]*Component
	// Information is optional map-level help text. It is set from a reserved
	// top-level "information" key inside the mapping (e.g. directly under
	// perk_types:), and is not treated as a component.
	Information string
	// RenderInformation, when true, renders the map-level Information as plain
	// text between the section header and the first field instead of behind a
	// hover "i" badge next to the header. It is set from a reserved top-level
	// "render_information" key inside the mapping and is not a component.
	RenderInformation bool
}

// UnmarshalYAML decodes a mapping node into an ordered ComponentMap.
func (m *ComponentMap) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping for component map, got kind %d", n.Kind)
	}
	if m.Items == nil {
		m.Items = map[string]*Component{}
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		key := n.Content[i].Value
		// "information" is a reserved map-level key, not a component.
		if key == "information" {
			m.Information = n.Content[i+1].Value
			continue
		}
		// "render_information" is a reserved map-level key, not a component.
		if key == "render_information" {
			m.RenderInformation = n.Content[i+1].Value == "true"
			continue
		}

		var comp Component

		if err := n.Content[i+1].Decode(&comp); err != nil {
			return fmt.Errorf("component %q: %w", key, err)
		}
		comp.ID = key
		if _, seen := m.Items[key]; !seen {
			m.Order = append(m.Order, key)
		}
		m.Items[key] = &comp
	}
	return nil
}

// List returns the components in author order.
func (m ComponentMap) List() []*Component {
	out := make([]*Component, 0, len(m.Order))
	for _, k := range m.Order {
		out = append(out, m.Items[k])
	}
	return out
}

// Get returns a component by id.
func (m ComponentMap) Get(id string) (*Component, bool) {
	c, ok := m.Items[id]
	return c, ok
}

// TraitMap is an ordered, id-keyed collection of trait groups.
type TraitMap struct {
	Order []string
	Items map[string]*TraitGroup
}

// UnmarshalYAML decodes a mapping node into an ordered TraitMap.
func (m *TraitMap) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping for trait map, got kind %d", n.Kind)
	}
	if m.Items == nil {
		m.Items = map[string]*TraitGroup{}
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		key := n.Content[i].Value
		var g TraitGroup
		if err := n.Content[i+1].Decode(&g); err != nil {
			return fmt.Errorf("trait group %q: %w", key, err)
		}
		g.ID = key
		if _, seen := m.Items[key]; !seen {
			m.Order = append(m.Order, key)
		}
		m.Items[key] = &g
	}
	return nil
}

// List returns the trait groups in author order.
func (m TraitMap) List() []*TraitGroup {
	out := make([]*TraitGroup, 0, len(m.Order))
	for _, k := range m.Order {
		out = append(out, m.Items[k])
	}
	return out
}

// SkillMap is an ordered, category-keyed collection of skill lists.
type SkillMap struct {
	Order []string
	Items map[string][]string

	// Information is optional per-skill help text, keyed by bare skill name or
	// by "<group>.<skill>" (the dotted form wins, so the same skill name in
	// several groups can be described differently). It is authored inline on
	// the skill entry and drives the hover "i" badge wherever the skill is
	// listed and the option tooltip in the builder dropdowns.
	Information map[string]string

	// RenderInformation, when true, renders the information as plain text
	// instead of behind a hover "i", mirroring the field schema's flag.
	RenderInformation bool
}

// UnmarshalYAML decodes a mapping node into an ordered SkillMap.
func (m *SkillMap) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping for skill map, got kind %d", n.Kind)
	}
	if m.Items == nil {
		m.Items = map[string][]string{}
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		key := n.Content[i].Value
		// "render_information" is a reserved map-level key, not a skill group,
		// the same way ComponentMap reserves its own.
		if key == "render_information" {
			m.RenderInformation = n.Content[i+1].Value == "true"
			continue
		}

		skills, err := decodeSkillEntries(n.Content[i+1], key, m)
		if err != nil {
			return err
		}
		if _, seen := m.Items[key]; !seen {
			m.Order = append(m.Order, key)
		}
		m.Items[key] = skills
	}
	return nil
}

// decodeSkillEntries decodes one group's skill list. Each entry is either a
// bare scalar (the skill name, no help text) or a mapping with a "name" and an
// optional "information" text, mirroring OptionList's scalar-or-mapping
// authoring. An entry carrying information registers it under
// "<group>.<name>" — and under the bare "name" as well unless that bare key is
// already taken — so the same-named copies in other groups inherit the text
// from the first labelled occurrence.
func decodeSkillEntries(n *yaml.Node, group string, m *SkillMap) ([]string, error) {
	if n.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("expected sequence for skill group %q, got kind %d", group, n.Kind)
	}
	skills := make([]string, 0, len(n.Content))
	for _, item := range n.Content {
		var name, information string
		switch item.Kind {
		case yaml.ScalarNode:
			name = item.Value
		case yaml.MappingNode:
			var entry struct {
				Name        string `yaml:"name"`
				Information string `yaml:"information"`
			}
			if err := item.Decode(&entry); err != nil {
				return nil, fmt.Errorf("skill entry in group %q: %w", group, err)
			}
			if entry.Name == "" {
				return nil, fmt.Errorf("skill entry in group %q: name is required", group)
			}
			name = entry.Name
			information = entry.Information
		default:
			return nil, fmt.Errorf("unexpected skill entry kind %d in group %q", item.Kind, group)
		}
		if information != "" {
			if m.Information == nil {
				m.Information = map[string]string{}
			}
			m.Information[group+"."+name] = information
			if _, ok := m.Information[name]; !ok {
				m.Information[name] = information
			}
		}
		skills = append(skills, name)
	}
	return skills, nil
}

// Info returns the hover help text configured for one skill. A
// "<group>.<skill>" key wins over a bare skill name, so a name that appears in
// several groups can be special-cased per group.
func (m SkillMap) Info(groupID, skill string) string {
	if m.Information == nil {
		return ""
	}
	if s, ok := m.Information[groupID+"."+skill]; ok && s != "" {
		return s
	}
	if s, ok := m.Information[skill]; ok {
		return s
	}
	return ""
}

// SkillGroup is an ordered view of one skill category.
type SkillGroup struct {
	ID     string
	Label  string
	Skills []string
}

// List returns the skill categories as ordered groups.
func (m SkillMap) List() []SkillGroup {
	out := make([]SkillGroup, 0, len(m.Order))
	for _, k := range m.Order {
		out = append(out, SkillGroup{ID: k, Label: titleCase(k), Skills: m.Items[k]})
	}
	return out
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	return string(s[0]-32) + s[1:]
}
