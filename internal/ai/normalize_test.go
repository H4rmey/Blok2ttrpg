package ai

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/harmey/blok2ttrpg-v5/internal/export"
)

func decode(t *testing.T, raw string) map[string]any {
	t.Helper()
	cfg := loadCfg(t)
	out, err := NormalizeYAML(cfg, []byte(raw))
	if err != nil {
		t.Fatalf("NormalizeYAML: %v", err)
	}
	var doc map[string]any
	if err := yaml.Unmarshal(out, &doc); err != nil {
		t.Fatalf("normalized output is not valid YAML: %v\n%s", err, out)
	}
	return doc
}

func TestNormalizeMapsProficiencyNamesToIDs(t *testing.T) {
	doc := decode(t, "skills:\n  general.Crafting: Expert\n  general.Knowledge: novice\n")
	skills, _ := doc["skills"].(map[string]any)
	if skills["general.Crafting"] != "expert" {
		t.Errorf("Crafting = %v, want expert", skills["general.Crafting"])
	}
	if skills["general.Knowledge"] != "novice" {
		t.Errorf("Knowledge = %v, want novice", skills["general.Knowledge"])
	}
}

func TestNormalizePassiveSpellings(t *testing.T) {
	cases := map[string]string{
		"canonical": "perks:\n  - name: R\n    fields:\n      passive_id: resistance\n",
		"name":      "perks:\n  - name: R\n    fields:\n      passive_id: Resistance\n",
		"top level": "perks:\n  - name: PM\n    passive_id: Prepared Mind\n    passive_fields:\n      energy: 2\n",
		"list":      "perks:\n  - name: PM\n    passives:\n      - type: Prepared Mind\n        fields:\n          energy: 2\n",
	}
	for label, raw := range cases {
		doc := decode(t, raw)
		perks, _ := doc["perks"].([]any)
		p, _ := perks[0].(map[string]any)
		fields, _ := p["fields"].(map[string]any)
		if fields["passive_id"] != "prepared_mind" && fields["passive_id"] != "resistance" {
			t.Errorf("%s: passive_id = %v", label, fields["passive_id"])
		}
		if label == "list" || label == "top level" {
			if p["type"] != "passive" {
				t.Errorf("%s: type = %v, want passive", label, p["type"])
			}
			pf, _ := fields["passive_fields"].(map[string]any)
			if pf["energy"] != 2 {
				t.Errorf("%s: passive_fields = %v", label, fields["passive_fields"])
			}
		}
	}
}

// TestNormalizedPassiveValidates ensures the reshaping produces a perk the
// validator accepts (as a passive, via passive_id) rather than "has no type".
func TestNormalizedPassiveValidates(t *testing.T) {
	cfg := loadCfg(t)
	raw := "level: 1\nperks:\n  - name: PM\n    passives:\n      - type: Prepared Mind\n"
	out, err := NormalizeYAML(cfg, []byte(raw))
	if err != nil {
		t.Fatalf("NormalizeYAML: %v", err)
	}
	c, err := export.UnmarshalCharacter(out)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	report := Validate(cfg, c)
	for _, issue := range report.Issues {
		if strings.Contains(issue, "no type") {
			t.Errorf("passive still reported as typeless: %v", report.Issues)
		}
	}
}
