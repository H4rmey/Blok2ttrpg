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

// levelOne starts every fixture at level 1 explicitly, since the model chooses
// the level now and Validate reads it off the character.
func levelOne(c model.Character) model.Character {
	c.Level = 1
	return c
}

func TestValidateExactSpendIsValid(t *testing.T) {
	cfg := loadCfg(t)
	c := withSkillPoints(levelOne(model.Character{}), 13)
	c = withPassives(c, 20) // level 1 perk budget is 20

	report := Validate(cfg, c)
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

	under := withPassives(withSkillPoints(levelOne(model.Character{}), 10), 20)
	r := Validate(cfg, under)
	if r.Valid {
		t.Fatal("character with unspent points reported valid")
	}
	if !hasIssue(r.Issues, "3 unspent skill points") {
		t.Errorf("missing unspent issue: %v", r.Issues)
	}

	over := withPassives(withSkillPoints(levelOne(model.Character{}), 14), 20)
	r = Validate(cfg, over)
	if !hasIssue(r.Issues, "1 over budget on skill points") {
		t.Errorf("missing overspend issue: %v", r.Issues)
	}
}

func TestValidateReportsUnknownIDs(t *testing.T) {
	cfg := loadCfg(t)

	c := withSkillPoints(levelOne(model.Character{}), 13)
	c.Skills["bogus.Skill"] = "novice"
	c = withPassives(c, 20)
	r := Validate(cfg, c)
	if !hasIssue(r.Issues, `unknown skill key "bogus.Skill"`) {
		t.Errorf("missing unknown skill issue: %v", r.Issues)
	}

	c = withSkillPoints(levelOne(model.Character{}), 13)
	c.Perks = append(c.Perks, model.Perk{Name: "Bad", Type: "not_a_type"})
	r = Validate(cfg, c)
	if !hasIssue(r.Issues, `unknown type "not_a_type"`) {
		t.Errorf("missing unknown type issue: %v", r.Issues)
	}
}

// TestValidateRejectsEmptyPerk pins the rule that a non-passive perk must have
// at least one enactment, while a passive (which carries a passive_id) is
// exempt and needs none.
func TestValidateRejectsEmptyPerk(t *testing.T) {
	cfg := loadCfg(t)

	c := withSkillPoints(levelOne(model.Character{}), 13)
	c.Perks = append(c.Perks, model.Perk{Name: "Empty", Type: cfg.PerkTypes.List()[0].ID})
	r := Validate(cfg, c)
	if !hasIssue(r.Issues, `perk "Empty" has no enactments`) {
		t.Errorf("empty non-passive perk was not rejected: %v", r.Issues)
	}

	// The same budgets with passives instead are fine: a passive is exempt.
	ok := withPassives(withSkillPoints(levelOne(model.Character{}), 13), 20)
	if r := Validate(cfg, ok); !r.Valid {
		t.Errorf("passive character should be valid, got: %v", r.Issues)
	}
}

// TestValidatePerkExactSpend pins that the perk pool is back to exact spend now
// that the 1-point dials make an exact total reachable.
func TestValidatePerkExactSpend(t *testing.T) {
	cfg := loadCfg(t)

	atBudget := withPassives(withSkillPoints(levelOne(model.Character{}), 13), 20)
	if r := Validate(cfg, atBudget); !r.Valid {
		t.Errorf("20/20 perk points should be valid, got: %v", r.Issues)
	}

	below := withPassives(withSkillPoints(levelOne(model.Character{}), 13), 19)
	if r := Validate(cfg, below); r.Valid || !hasIssue(r.Issues, "1 unspent perk points") {
		t.Errorf("19/20 perk points should be rejected, got: %v", r.Issues)
	}

	over := withPassives(withSkillPoints(levelOne(model.Character{}), 13), 21)
	if r := Validate(cfg, over); r.Valid || !hasIssue(r.Issues, "over budget on perk points") {
		t.Errorf("21/20 perk points should be rejected, got: %v", r.Issues)
	}
}

// TestValidateResourceSplit pins the resource pair rule: generate and consume
// must live in two separate perks sharing one resource_name, and every side
// needs its counterpart.
func TestValidateResourceSplit(t *testing.T) {
	cfg := loadCfg(t)
	fields := func(pairs ...string) map[string]any {
		m := map[string]any{}
		for i := 0; i+1 < len(pairs); i += 2 {
			m[pairs[i]] = pairs[i+1]
		}
		return m
	}
	gen := func(name string, extras ...string) model.Enactment {
		f := append([]string{"resource_mode", "generate", "resource_name", name}, extras...)
		return model.Enactment{Type: "resource", Fields: fields(f...)}
	}
	use := func(name, effect string) model.Enactment {
		return model.Enactment{Type: "resource", Fields: fields(
			"resource_mode", "consume", "resource_name", name, "consumed_effect", effect)}
	}

	// One perk that generates and names an effect in a single enactment must fail.
	combined := levelOne(model.Character{Perks: []model.Perk{{
		Name: "Combo", Type: "execution",
		Enactments: []model.Enactment{gen("Rage", "consumed_effect", "damage")},
	}}})
	if r := Validate(cfg, combined); !hasIssue(r.Issues, "separate perks") {
		t.Errorf("combined generate+consume should be rejected, got: %v", r.Issues)
	}

	// A consumer with no generator must fail.
	orphan := levelOne(model.Character{Perks: []model.Perk{{
		Name: "Spender", Type: "execution", Enactments: []model.Enactment{use("Rage", "damage")},
	}}})
	if r := Validate(cfg, orphan); !hasIssue(r.Issues, "no perk generates it") {
		t.Errorf("orphan consumer should be rejected, got: %v", r.Issues)
	}

	// A generator with no consumer must fail.
	lonely := levelOne(model.Character{Perks: []model.Perk{{
		Name: "Builder", Type: "execution", Enactments: []model.Enactment{gen("Rage")},
	}}})
	if r := Validate(cfg, lonely); !hasIssue(r.Issues, "no perk consumes it") {
		t.Errorf("orphan generator should be rejected, got: %v", r.Issues)
	}

	// The matching pair must pass the resource checks.
	pair := levelOne(model.Character{Perks: []model.Perk{
		{Name: "Builder", Type: "execution", Enactments: []model.Enactment{gen("Rage")}},
		{Name: "Spender", Type: "execution", Enactments: []model.Enactment{use("Rage", "damage")}},
	}})
	for _, issue := range Validate(cfg, pair).Issues {
		if strings.Contains(issue, "resource") || strings.Contains(issue, "generates") || strings.Contains(issue, "consumes") {
			t.Errorf("valid pair flagged: %q", issue)
		}
	}

	// A hand-written consumed_effect_ib is flagged even on the pair.
	withIB := pair
	withIB.Perks[1].Enactments[0].Fields["consumed_effect_ib"] = map[string]any{"source": "1d6"}
	if r := Validate(cfg, withIB); !hasIssue(r.Issues, "consumed_effect_ib") {
		t.Errorf("hand-written nested key should be rejected, got: %v", r.Issues)
	}
}

// TestValidatePerk checks the single-perk validator used by the perk generator.
func TestValidatePerk(t *testing.T) {
	cfg := loadCfg(t)

	if issues := ValidatePerk(cfg, model.Perk{Name: "Empty", Type: cfg.PerkTypes.List()[0].ID}); len(issues) == 0 {
		t.Error("ValidatePerk accepted a non-passive perk with no enactments")
	}
	if issues := ValidatePerk(cfg, model.Perk{Name: "PM", Type: "passive", Fields: map[string]any{"passive_id": "mobile_attacker"}}); len(issues) != 0 {
		t.Errorf("ValidatePerk rejected a valid passive: %v", issues)
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
