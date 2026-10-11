package ai

import (
	"strings"
	"testing"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

func loadCfg(t *testing.T) *config.Config {
	t.Helper()
	loaded, err := config.Load("../../config/Blok2Simplified")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	return loaded.Config
}

// generalSkills lists far more than 13 general skill keys, so tests can raise
// an exact number of them.
func generalSkills() []string {
	return []string{
		"Stealth", "Perception", "Nature", "Crafting", "People Skill",
		"Performance", "Thievery", "Knowledge", "Magic", "Medicine",
		"Provoke", "Intimidate", "Athletics", "Deceive", "Resources",
	}
}

// withSkillPoints raises the first n general skills to novice, which costs one
// point each, so n points are spent.
func withSkillPoints(c model.Character, n int) model.Character {
	if c.Skills == nil {
		c.Skills = map[string]string{}
	}
	for i, s := range generalSkills() {
		if i >= n {
			break
		}
		c.Skills[model.SkillKey("general", s)] = "novice"
	}
	return c
}

// withPassives adds n "Mobile Attacker" passives, each of which costs one perk
// point, so n perk points are spent.
func withPassives(c model.Character, n int) model.Character {
	for i := 0; i < n; i++ {
		c.Perks = append(c.Perks, model.Perk{
			Name:   "Mobile Attacker",
			Fields: map[string]any{"passive_id": "mobile_attacker"},
		})
	}
	return c
}

func TestValidateExactSpendIsValid(t *testing.T) {
	cfg := loadCfg(t)
	c := withSkillPoints(model.Character{}, 13)
	c = withPassives(c, 20) // level 1 perk budget is 20

	report := Validate(cfg, c, 1)
	if !report.Valid {
		t.Fatalf("expected valid, got issues: %v", report.Issues)
	}
	if len(report.Pools) != 2 {
		t.Fatalf("expected 2 pools, got %d", len(report.Pools))
	}
	for _, p := range report.Pools {
		if p.Used != p.Budget {
			t.Errorf("%s pool = %d used of %d", p.Name, p.Used, p.Budget)
		}
	}
}

func TestValidateReportsUnspentAndOverspentSkillPoints(t *testing.T) {
	cfg := loadCfg(t)

	under := withPassives(withSkillPoints(model.Character{}, 10), 20)
	r := Validate(cfg, under, 1)
	if r.Valid {
		t.Fatal("character with unspent points reported valid")
	}
	if !hasIssue(r.Issues, "3 unspent skill points") {
		t.Errorf("missing unspent issue: %v", r.Issues)
	}

	over := withPassives(withSkillPoints(model.Character{}, 14), 20)
	r = Validate(cfg, over, 1)
	if !hasIssue(r.Issues, "1 over budget on skill points") {
		t.Errorf("missing overspend issue: %v", r.Issues)
	}
}

func TestValidateReportsUnknownIDs(t *testing.T) {
	cfg := loadCfg(t)

	c := withSkillPoints(model.Character{}, 13)
	c.Skills["bogus.Skill"] = "novice"
	c = withPassives(c, 20)
	r := Validate(cfg, c, 1)
	if !hasIssue(r.Issues, `unknown skill key "bogus.Skill"`) {
		t.Errorf("missing unknown skill issue: %v", r.Issues)
	}

	c = withSkillPoints(model.Character{}, 13)
	c.Perks = append(c.Perks, model.Perk{Name: "Bad", Type: "not_a_type"})
	r = Validate(cfg, c, 1)
	if !hasIssue(r.Issues, `unknown type "not_a_type"`) {
		t.Errorf("missing unknown type issue: %v", r.Issues)
	}
}

func TestValidatePerkSpendFloor(t *testing.T) {
	cfg := loadCfg(t)
	// Level 1 perk budget is 20, so the 80% floor is 16.
	atFloor := withPassives(withSkillPoints(model.Character{}, 13), 16)
	if r := Validate(cfg, atFloor, 1); !r.Valid {
		t.Errorf("16/20 perk points should be valid, got: %v", r.Issues)
	}

	below := withPassives(withSkillPoints(model.Character{}, 13), 15)
	r := Validate(cfg, below, 1)
	if r.Valid || !hasIssue(r.Issues, "only 15 of 20 perk points spent") {
		t.Errorf("15/20 perk points should be rejected, got: %v", r.Issues)
	}

	over := withPassives(withSkillPoints(model.Character{}, 13), 21)
	if r := Validate(cfg, over, 1); r.Valid || !hasIssue(r.Issues, "over budget on perk points") {
		t.Errorf("21/20 perk points should be rejected, got: %v", r.Issues)
	}
}

func TestCorrectionNamesTheIssues(t *testing.T) {
	r := Report{Issues: []string{"3 unspent skill points (13 used of 16)."}}
	got := Correction(r)
	if !strings.Contains(got, "3 unspent skill points") || !strings.Contains(got, "YAML only") {
		t.Errorf("correction text missing detail: %q", got)
	}
}

func hasIssue(issues []string, substr string) bool {
	for _, i := range issues {
		if strings.Contains(i, substr) {
			return true
		}
	}
	return false
}
