// This file prices character-level budgets: the cumulative skill-point cost of
// a proficiency ladder, and what a content package costs to install.
package engine

import (
	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// SkillPointsUsed sums the skill-point cost of all skill assignments. Cost is
// cumulative across proficiency tiers: the default (starting) tier is free, and
// raising a skill to a higher tier costs the sum of the per-tier costs for
// every tier above the default up to and including the selected one. A brand
// new character sits at the default tier on every skill and therefore uses zero
// points; each tier upgrade progressively consumes more.
func SkillPointsUsed(cfg *config.Config, c model.Character) int {
	used := 0
	for _, g := range cfg.Skills.List() {
		for _, skill := range g.Skills {
			profID := c.Skills[model.SkillKey(g.ID, skill)]
			used += cumulativeSkillCost(cfg, profID)
		}
	}
	return used
}

// cumulativeSkillCost returns the total points needed to raise a skill from the
// default tier to the given proficiency tier. Tiers are ordered by their
// position in cfg.Proficiencies; the configured default tier is free. The cost
// is the sum of the per-tier `cost` values for each tier strictly above the
// default up to and including the selected tier. Selecting a tier at or below
// the default costs zero, so tiers below the default (e.g. negative modifiers)
// are free choices rather than discounts.
func cumulativeSkillCost(cfg *config.Config, profID string) int {
	base := cfg.DefaultProficiencyIndex()
	target := -1
	for i, p := range cfg.Proficiencies {
		if p.ID == profID {
			target = i
			break
		}
	}
	if target <= base {
		return 0
	}
	sum := 0
	for i := base + 1; i <= target && i < len(cfg.Proficiencies); i++ {
		sum += cfg.Proficiencies[i].Cost
	}
	return sum
}

// PackageCost is the price of importing a package, expressed in the two budgets
// a character actually spends: perk (perk) points and skill (skill) points.
type PackageCost struct {
	// Perk is the sum of the build cost of every perk the package installs.
	Perk int `json:"perk"`
	// Skill is the number of skill points the package's proficiency shifts
	// consume on top of what the character already spends.
	Skill int `json:"skill"`
}

// PackageCostFor computes what a package would cost the given character. Perk
// points are the summed build cost of the package's perks. Skill points are
// the difference in cumulative skill cost between the character's current tier
// and the tier the shift would move it to, so a shift that is already paid for
// (or that moves a skill downward) does not charge again.
//
// The character is not modified. A skill the character has no entry for is
// treated as sitting at the configured default tier.
func PackageCostFor(cfg *config.Config, c model.Character, shifts map[string]int, perks []model.Perk) PackageCost {
	var out PackageCost
	for _, ab := range perks {
		out.Perk += PerkCost(cfg, ab).Build
	}
	def := cfg.DefaultProficiencyID()
	for skillKey, delta := range shifts {
		if delta == 0 {
			continue
		}
		current, ok := c.Skills[skillKey]
		if !ok || current == "" {
			current = def
		}
		before := cumulativeSkillCost(cfg, current)
		after := cumulativeSkillCost(cfg, cfg.ShiftProficiency(current, delta))
		out.Skill += after - before
	}
	return out
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
