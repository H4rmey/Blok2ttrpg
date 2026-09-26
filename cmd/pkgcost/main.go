// Command pkgcost prints the perk-point and skill-point cost of every library
// package against a fresh level 1 character. It is a balancing aid for authoring
// packages: it shows at a glance whether a package fits inside the level 1
// budgets reported in the header.
package main

import (
	"fmt"
	"log"
	"sort"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/engine"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
	"github.com/harmey/blok2ttrpg-v5/internal/premade"
)

func main() {
	cfg, err := config.Load("config/Blok2Simplified")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	lib := premade.New("library")
	pkgs, err := lib.ListPackages()
	if err != nil {
		log.Fatalf("list packages: %v", err)
	}

	// A blank level 1 character: every trait sits at the default tier.
	c := model.Character{Level: 1, Traits: map[string]string{}}
	def := cfg.DefaultProficiencyID()
	for _, g := range cfg.Traits.List() {
		for _, t := range g.Traits {
			c.Traits[model.TraitKey(g.ID, t)] = def
		}
	}

	perkBudget := cfg.AbilityPointBudget(1)
	skillBudget := cfg.TraitPointBudget(1)
	fmt.Printf("Level 1 budgets: %d perk pt, %d skill pt\n\n", perkBudget, skillBudget)

	for _, pkg := range pkgs {
		cost := engine.PackageCostFor(cfg.Config, c, pkg.Shifts, pkg.Abilities)
		flag := "ok"
		if cost.Perk > perkBudget || cost.Skill > skillBudget {
			flag = "OVER"
		}
		fmt.Printf("%-12s %-14s perk=%2d skill=%2d  %s\n", pkg.Category, pkg.ID, cost.Perk, cost.Skill, flag)
		// Per-ability breakdown so an over-budget package shows which perk to drop.
		var names []string
		for _, ab := range pkg.Abilities {
			names = append(names, fmt.Sprintf("%s=%d", ab.Name, engine.AbilityCost(cfg.Config, engine.NormalizeAbility(cfg.Config, ab)).Build))
		}
		sort.Strings(names)
		for _, n := range names {
			fmt.Printf("               %s\n", n)
		}
	}
}
