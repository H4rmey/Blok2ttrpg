package web

import (
	"testing"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/engine"
	"github.com/harmey/blok2ttrpg-v5/internal/premade"
)

// TestEnergyCostMatchesEnactmentCount pins the energy rule: using a perk costs
// 1 energy per enactment, so a two-enactment perk costs 2 energy. The first
// enactment is paid by the ability type's base_cost, each additional one by
// additional_enactment.energy_cost. This walks the whole built-in library, so a
// config change that breaks the rule for any perk fails here.
func TestEnergyCostMatchesEnactmentCount(t *testing.T) {
	cfg, err := config.Load("../../config/Blok2Simplified")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	abs, err := premade.New("../../library").ListAbilities()
	if err != nil {
		t.Fatalf("list abilities: %v", err)
	}
	if len(abs) == 0 {
		t.Fatal("no library perks found")
	}
	for _, raw := range abs {
		ab := engine.NormalizeAbility(cfg.Config, raw)

		// Only enactments that actually carry a type are charged for.
		want := 0
		refunds := false
		for _, en := range ab.Enactments {
			if en.Type == "" {
				continue
			}
			want++
			// Enact Nerf trades a self-inflicted drawback for energy, so a perk
			// using it is legitimately cheaper than the per-enactment rule. Such
			// a perk is only required to stay within [1, enactments].
			if en.Type == "nerf" {
				refunds = true
			}
		}
		if want == 0 {
			want = 1 // the floor: every perk costs at least 1 energy
		}

		got := engine.AbilityCost(cfg.Config, ab).Energy
		switch {
		case refunds:
			if got < 1 || got > want {
				t.Errorf("%s (%s, %d enactments, refunds energy): energy = %d, want between 1 and %d",
					ab.Name, ab.Type, len(ab.Enactments), got, want)
			}
		case got != want:
			t.Errorf("%s (%s, %d enactments): energy = %d, want %d",
				ab.Name, ab.Type, len(ab.Enactments), got, want)
		}
	}
}
