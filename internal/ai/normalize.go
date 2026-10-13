// This file makes the application "lead" the model: it rewrites a generated
// character into the canonical import shape before it is parsed and validated.
//
// Models reliably choose the right concept but express a few things in human
// terms rather than the exact ids and nesting the importer expects: a
// proficiency as its tier name ("Expert"), a passive as a `passives:` list
// instead of `fields.passive_id`, or a hand-built `consumed_effect_ib` nested
// map with the wrong prefixed keys. Normalising those here means a character
// that is correct in spirit imports on the first try instead of bouncing
// through retries for a cosmetic mismatch.
package ai

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
)

// NormalizeYAML rewrites a generated character document into the canonical
// import shape. It never invents data: it only maps names to the ids they
// already refer to, and reshapes a passive into the fields the importer reads.
func NormalizeYAML(cfg *config.Config, raw []byte) ([]byte, error) {
	var doc map[string]any
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return raw, err
	}
	if doc == nil {
		return raw, fmt.Errorf("empty character document")
	}
	normalizeSkills(cfg, doc)
	normalizePerks(cfg, doc)
	out, err := yaml.Marshal(doc)
	if err != nil {
		return raw, err
	}
	return out, nil
}

// normalizeSkills maps each skill proficiency from a tier name (or id, case
// insensitively) to its canonical lowercase id.
func normalizeSkills(cfg *config.Config, doc map[string]any) {
	skills, _ := doc["skills"].(map[string]any)
	if skills == nil {
		return
	}
	idx := proficiencyIndex(cfg)
	for k, v := range skills {
		s, ok := v.(string)
		if !ok {
			continue
		}
		if id, ok := idx[strings.ToLower(strings.TrimSpace(s))]; ok {
			skills[k] = id
		}
	}
}

func normalizePerks(cfg *config.Config, doc map[string]any) {
	perks, _ := doc["perks"].([]any)
	if perks == nil {
		return
	}
	idx := passiveIndex(cfg)
	for _, raw := range perks {
		if p, ok := raw.(map[string]any); ok {
			normalizePassive(p, idx)
			normalizeResource(p)
		}
	}
}

// normalizeResource drops a hand-written consumed_effect_ib nested map from a
// resource enactment. The app builds that value itself (NormalizePerk), and the
// model's hand-written copy always names the wrong prefixed keys. The intent
// (resource_mode, resource_name, consumed_effect) is preserved, so validation
// can still judge the generate/spender split.
func normalizeResource(p map[string]any) {
	enactments, _ := p["enactments"].([]any)
	for _, raw := range enactments {
		en, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if t, _ := en["type"].(string); !strings.EqualFold(strings.TrimSpace(t), "resource") {
			continue
		}
		if fields, ok := en["fields"].(map[string]any); ok {
			delete(fields, "consumed_effect_ib")
		}
	}
}

// normalizePassive reshapes every accepted spelling of a passive into
// fields.passive_id (+ fields.passive_fields):
//
//	fields: { passive_id: resistance }        # already canonical
//	passive_id: resistance                     # top level
//	passives: [ { type: Resistance, fields: {...} } ]   # catalogue list
//
// A perk with none of these is left untouched (an ordinary built perk).
func normalizePassive(p map[string]any, idx map[string]string) {
	fields, _ := p["fields"].(map[string]any)
	if id := mapPassiveID(idx, stringField(fields, "passive_id")); id != "" {
		ensureFields(p)["passive_id"] = id
		return
	}
	if id := mapPassiveID(idx, firstString(p, "passive_id", "passive")); id != "" {
		f := ensureFields(p)
		f["passive_id"] = id
		if pf, ok := p["passive_fields"].(map[string]any); ok {
			f["passive_fields"] = pf
		}
		if p["type"] == nil {
			p["type"] = "passive"
		}
		return
	}
	list, _ := p["passives"].([]any)
	if len(list) == 0 {
		return
	}
	entry, _ := list[0].(map[string]any)
	if entry == nil {
		return
	}
	id := mapPassiveID(idx, firstString(entry, "type", "id", "passive_id", "name"))
	if id == "" {
		return
	}
	f := ensureFields(p)
	f["passive_id"] = id
	if ef, ok := entry["fields"].(map[string]any); ok {
		f["passive_fields"] = ef
	}
	p["type"] = "passive"
	delete(p, "passives")
}

func ensureFields(p map[string]any) map[string]any {
	if f, ok := p["fields"].(map[string]any); ok {
		return f
	}
	f := map[string]any{}
	p["fields"] = f
	return f
}

func stringField(fields map[string]any, key string) string {
	if fields == nil {
		return ""
	}
	s, _ := fields[key].(string)
	return s
}

func firstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if s, ok := m[k].(string); ok && strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}

func mapPassiveID(idx map[string]string, s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	return idx[strings.ToLower(strings.TrimSpace(s))]
}

// proficiencyIndex maps tier names and ids (lowercased) to the canonical id.
func proficiencyIndex(cfg *config.Config) map[string]string {
	m := map[string]string{}
	for _, p := range cfg.Proficiencies {
		m[strings.ToLower(strings.TrimSpace(p.Name))] = p.ID
		m[strings.ToLower(p.ID)] = p.ID
	}
	return m
}

// passiveIndex maps passive names and ids (lowercased) to the canonical id.
func passiveIndex(cfg *config.Config) map[string]string {
	m := map[string]string{}
	for _, p := range cfg.Passives.Entries {
		m[strings.ToLower(strings.TrimSpace(p.Name))] = p.ID
		m[strings.ToLower(p.ID)] = p.ID
	}
	return m
}
