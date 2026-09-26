package web

import (
	"testing"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/engine"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// twoEnactmentPerk builds a perk whose second enactment either inherits the
// first enactment's target or buys its own, depending on newTarget. Both
// variants carry identical interaction and validation data, so the only
// difference between their costs is what the flag makes the engine charge for.
func twoEnactmentPerk(newTarget bool) model.Ability {
	direct := map[string]any{"range": "5", "targets": 1}
	validation := map[string]any{
		"engage":        "d6",
		"counter_trait": []map[string]any{{"value": "defense.Reflex"}},
	}
	return model.Ability{
		Name: "Test Perk",
		Type: "execution",
		Enactments: []model.Enactment{
			{
				Type:            "damage",
				Fields:          map[string]any{"source": "offense.Strength"},
				Interaction:     "direct",
				InteractionData: direct,
				ValidationData:  validation,
			},
			{
				Type:            "condition",
				Fields:          map[string]any{"condition": "prone", "duration": "1"},
				Interaction:     "direct",
				InteractionData: direct,
				ValidationData:  validation,
				NewTarget:       newTarget,
			},
		},
	}
}

// TestNewTargetCostsMoreThanInheriting is the whole point of the new_target
// opt-in: an enactment that inherits the previous enactment's target pays for
// neither an Interaction nor a Validation, while one that picks its own target
// pays the configured surcharge plus both regions. The inheriting variant must
// therefore be strictly cheaper.
func TestNewTargetCostsMoreThanInheriting(t *testing.T) {
	cfg, err := config.Load("../../config/Blok2Simplified")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	inherit := engine.AbilityCost(cfg.Config, engine.NormalizeAbility(cfg.Config, twoEnactmentPerk(false)))
	own := engine.AbilityCost(cfg.Config, engine.NormalizeAbility(cfg.Config, twoEnactmentPerk(true)))

	if own.Build <= inherit.Build {
		t.Errorf("a second enactment with its own target should cost more build: inherit=%d own=%d",
			inherit.Build, own.Build)
	}
	// The surcharge itself must be charged on top of the revealed regions, so
	// the gap is at least the configured new_target build cost.
	if min := cfg.AdditionalEnactment.NewTarget.BuildCost; own.Build-inherit.Build < min {
		t.Errorf("new_target surcharge of %d not applied: inherit=%d own=%d",
			min, inherit.Build, own.Build)
	}
}

// TestInheritedTargetInstructions checks the generated rules text: an enactment
// that reuses the previous target says so and asks for no separate roll, so a
// player cannot mistake it for a second independent attack.
func TestInheritedTargetInstructions(t *testing.T) {
	cfg, err := config.Load("../../config/Blok2Simplified")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	ins := engine.AbilityInstructions(cfg.Config, engine.NormalizeAbility(cfg.Config, twoEnactmentPerk(false)))
	if len(ins) != 2 {
		t.Fatalf("expected 2 instructions, got %d", len(ins))
	}
	if ins[1].Interaction == ins[0].Interaction {
		t.Errorf("second enactment should describe an inherited target, got %q", ins[1].Interaction)
	}
	if ins[1].Validation == ins[0].Validation {
		t.Errorf("second enactment should not repeat the first roll, got %q", ins[1].Validation)
	}
}
