package web

import (
	"testing"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/engine"
	"github.com/harmey/blok2ttrpg-v5/internal/premade"
)

// TestEnergyCostMatchesEnactmentCount pins the energy rule: using a perk costs
// its perk type's base energy for the first enactment, plus
// additional_enactment.energy_cost for each one after that. Both numbers are read
// from the config rather than hardcoded, because a perk type may legitimately
// charge more than one energy to use: Reaction charges two, since acting out of
// turn is priced at a premium.
//
// This walks the whole built-in library, so a config change that breaks the rule
// for any perk fails here.
func TestEnergyCostMatchesEnactmentCount(t *testing.T) {
	cfg, err := config.Load("../../config/Blok2Simplified")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	abs, err := premade.New("../../library").ListPerks()
	if err != nil {
		t.Fatalf("list perks: %v", err)
	}
	if len(abs) == 0 {
		t.Fatal("no library perks found")
	}
	for _, raw := range abs {
		ab := engine.NormalizePerk(cfg.Config, raw)

		// Only enactments that actually carry a type are charged for.
		enactments := 0
		refunds := false
		for _, en := range ab.Enactments {
			if en.Type == "" {
				continue
			}
			enactments++
			// Enact Nerf trades a self-inflicted drawback for energy, so a perk
			// using it is legitimately cheaper than the per-enactment rule. Such
			// a perk is only required to stay within [1, want].
			if en.Type == "nerf" {
				refunds = true
			}
		}

		// The first enactment is paid for by the perk type's own base energy;
		// every one after it adds the additional-enactment surcharge. A perk
		// with no enactments still pays the perk type's base, which is the
		// floor for simply using it.
		want := 0
		if at, ok := cfg.Config.PerkType(ab.Type); ok {
			want = at.BaseCost.EnergyCost
		}
		if enactments > 1 {
			want += cfg.Config.AdditionalEnactment.EnergyCost * (enactments - 1)
		}
		if want < 1 {
			want = 1 // the floor: every perk costs at least 1 energy
		}

		got := engine.PerkCost(cfg.Config, ab).Energy
		switch {
		case refunds:
			if got < 1 || got > want {
				t.Errorf("%s (%s, %d enactments, refunds energy): energy = %d, want between 1 and %d",
					ab.Name, ab.Type, enactments, got, want)
			}
		case got != want:
			t.Errorf("%s (%s, %d enactments): energy = %d, want %d",
				ab.Name, ab.Type, enactments, got, want)
		}
	}
}
