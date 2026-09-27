// Command lvlcheck is a temporary verification helper: it loads each ruleset
// and prints the clamped level plus the skill/perk budget it grants.
package main

import (
	"fmt"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
)

func main() {
	for _, p := range []string{"config/Blok2Simplified", "config.yaml"} {
		l, err := config.Load(p)
		if err != nil {
			fmt.Println(p, "LOAD ERROR:", err)
			continue
		}
		fmt.Printf("%s max_level=%d\n", p, l.MaxLevel())
		for _, lv := range []int{0, 1, 2, 5, 10, 11, 99} {
			fmt.Printf("  level %-3d -> clamp %-3d skill %-4d perk %-4d invoke %d\n",
				lv, l.ClampLevel(lv), l.SkillPointBudget(lv), l.PerkPointBudget(lv),
				l.InvokePointBudget(lv))

		}
	}
}
