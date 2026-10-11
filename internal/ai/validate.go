// This file validates a generated character against the loaded ruleset. It
// checks the two build pools the engine owns (skill points and perk points) and
// the referential integrity the cost engine assumes, so an AI reply that names
// an unknown id is reported instead of being silently priced as zero.
package ai

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/engine"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// PerkSpendFloor is the minimum fraction of the perk-point budget that must be
// spent. Perk costs do not divide evenly into the budget, so hitting it exactly
// is often impossible; any spend from this floor up to the full budget (but
// never over) is accepted. Skill points are still spent exactly.
const PerkSpendFloor = 0.8

// Pool is one build budget and how much of it the character spent.
type Pool struct {
	Name   string `json:"name"`
	Used   int    `json:"used"`
	Budget int    `json:"budget"`
}

// Report is the outcome of validating one generated character.
type Report struct {
	Valid  bool     `json:"valid"`
	Level  int      `json:"level"`
	Pools  []Pool   `json:"pools"`
	Issues []string `json:"issues"`
}

// Validate checks a generated character at the requested level. The level is
// forced to the requested (clamped) value so the budgets always match what the
// user asked for, even if the model emitted a different level.
func Validate(cfg *config.Config, c model.Character, level int) Report {
	c.Level = cfg.ClampLevel(level)
	var issues []string

	// Referential integrity. The cost engine silently prices an unknown perk
	// type or enactment as zero, so these are checked up front; otherwise an
	// invalid reply would look like underspending rather than an error.
	issues = append(issues, validateSkills(cfg, c)...)
	issues = append(issues, validatePerks(cfg, c)...)

	// Normalize before costing so the price matches what would be stored.
	for i := range c.Perks {
		c.Perks[i] = engine.NormalizePerk(cfg, c.Perks[i])
	}

	skillUsed := engine.SkillPointsUsed(cfg, c)
	perkUsed := 0
	for _, p := range c.Perks {
		perkUsed += engine.PerkCost(cfg, p).Build
	}

	pools := []Pool{
		{Name: "skill", Used: skillUsed, Budget: cfg.SkillPointBudget(c.Level)},
		{Name: "perk", Used: perkUsed, Budget: cfg.PerkPointBudget(c.Level)},
	}
	for _, p := range pools {
		if issue := poolIssue(p); issue != "" {
			issues = append(issues, issue)
		}
	}

	return Report{Valid: len(issues) == 0, Level: c.Level, Pools: pools, Issues: issues}
}

// poolIssue describes a spending violation in one pool, or "" when the pool is
// within its rules. Skill points must be spent exactly; perk points may be left
// partly unspent down to PerkSpendFloor of the budget, but never overspent.
func poolIssue(p Pool) string {
	if p.Name == "perk" {
		floor := int(math.Ceil(PerkSpendFloor * float64(p.Budget)))
		switch {
		case p.Used > p.Budget:
			return fmt.Sprintf("%d over budget on perk points (%d used of %d).", p.Used-p.Budget, p.Used, p.Budget)
		case p.Used < floor:
			return fmt.Sprintf("only %d of %d perk points spent; at least %d (80%%) is required.", p.Used, p.Budget, floor)
		}
		return ""
	}
	delta := p.Budget - p.Used
	switch {
	case delta > 0:
		return fmt.Sprintf("%d unspent %s points (%d used of %d).", delta, p.Name, p.Used, p.Budget)
	case delta < 0:
		return fmt.Sprintf("%d over budget on %s points (%d used of %d).", -delta, p.Name, p.Used, p.Budget)
	}
	return ""
}

// validateSkills reports unknown skill keys or proficiency ids.
func validateSkills(cfg *config.Config, c model.Character) []string {
	known := map[string]bool{}
	for _, g := range cfg.Skills.List() {
		for _, s := range g.Skills {
			known[model.SkillKey(g.ID, s)] = true
		}
	}
	var issues []string
	keys := make([]string, 0, len(c.Skills))
	for k := range c.Skills {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if !known[k] {
			issues = append(issues, fmt.Sprintf("unknown skill key %q.", k))
			continue
		}
		if id := c.Skills[k]; id != "" {
			if _, ok := cfg.Proficiency(id); !ok {
				issues = append(issues, fmt.Sprintf("skill %q has unknown proficiency %q.", k, id))
			}
		}
	}
	return issues
}

// validatePerks reports unknown perk types, enactment types and interaction
// ids. A passive perk (one carrying a resolvable passive_id) has none of those,
// so it is skipped.
func validatePerks(cfg *config.Config, c model.Character) []string {
	var issues []string
	for i, p := range c.Perks {
		name := p.Name
		if name == "" {
			name = fmt.Sprintf("#%d", i+1)
		}
		if pid := fieldString(p.Fields, "passive_id"); pid != "" {
			if _, ok := cfg.PassiveByID(pid); !ok {
				issues = append(issues, fmt.Sprintf("perk %q names unknown passive %q.", name, pid))
			}
			continue
		}
		if strings.EqualFold(p.Type, "passive") {
			issues = append(issues, fmt.Sprintf("passive perk %q is missing fields.passive_id; set it to one of the passive ids in the reference.", name))
			continue
		}
		if p.Type == "" {
			issues = append(issues, fmt.Sprintf("perk %q has no type.", name))
			continue
		}
		if _, ok := cfg.PerkType(p.Type); !ok {
			issues = append(issues, fmt.Sprintf("perk %q has unknown type %q.", name, p.Type))
		}
		for _, en := range p.Enactments {
			if en.Type == "" {
				continue
			}
			if _, ok := cfg.Enactment(en.Type); !ok {
				issues = append(issues, fmt.Sprintf("perk %q has unknown enactment %q.", name, en.Type))
			}
			if en.Interaction != "" {
				if _, ok := cfg.Interaction(en.Interaction); !ok {
					issues = append(issues, fmt.Sprintf("perk %q has unknown interaction %q.", name, en.Interaction))
				}
			}
		}
	}
	return issues
}

// Correction builds the user-role retry message appended after an invalid
// reply. It states the exact pool errors and asks for the full corrected YAML.
func Correction(r Report) string {
	var b strings.Builder
	b.WriteString("Your character failed validation. ")
	b.WriteString(strings.Join(r.Issues, " "))
	b.WriteString(" Spend ALL skill points exactly, and spend at least 80% of the perk points without exceeding the perk budget.")
	b.WriteString(" Use lowercase proficiency ids (novice, proficient, expert, ...), not tier names, and set fields.passive_id for any passive perk.")
	b.WriteString(" Return the full corrected character YAML only, with no markdown fences or commentary.")
	return b.String()
}

// fieldString reads a string field from a loosely-typed values map.
func fieldString(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	if s, ok := values[key].(string); ok {
		return s
	}
	return ""
}
