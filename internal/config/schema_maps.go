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
		var skills []string
		if err := n.Content[i+1].Decode(&skills); err != nil {
			return fmt.Errorf("skill group %q: %w", key, err)
		}
		if _, seen := m.Items[key]; !seen {
			m.Order = append(m.Order, key)
		}
		m.Items[key] = skills
	}
	return nil
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
