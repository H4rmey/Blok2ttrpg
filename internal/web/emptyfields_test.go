package web

import (
	"fmt"
	"testing"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/engine"
	"github.com/harmey/blok2ttrpg-v5/internal/premade"
)

// TestNoEmptyFieldsAfterNormalize asserts that a normalized perk never carries
// an empty value: every configured field of the ability type, of each enactment,
// and of each enactment's interaction resolves to something real. A half-filled
// enactment generates rules text with visible gaps, so "unset" is not a legal
// stored state - there is deliberately no empty option in any dropdown.
func TestNoEmptyFieldsAfterNormalize(t *testing.T) {
	cfg, err := config.Load("../../config/Blok2Simplified")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	abs, err := premade.New("../../library").ListAbilities()
	if err != nil {
		t.Fatalf("list abilities: %v", err)
	}
	for _, raw := range abs {
		ab := engine.NormalizeAbility(cfg.Config, raw)

		if at, ok := cfg.AbilityType(ab.Type); ok {
			assertNoEmptyFields(t, ab.Name+" (ability type)", at.Fields, ab.Fields)
		}
		for i, en := range ab.Enactments {
			where := fmt.Sprintf("%s enactment %d", ab.Name, i)
			if en.Type == "" {
				t.Errorf("%s: has no enactment type", where)
			}
			if ec, ok := cfg.Enactment(en.Type); ok {
				assertNoEmptyFields(t, where, ec.Fields, en.Fields)
			}
			if ic, ok := cfg.Interaction(en.Interaction); ok {
				assertNoEmptyFields(t, where+" interaction", ic.Fields, en.InteractionData)
			}
		}
	}
}

// assertNoEmptyFields reports any dropdown or text field left blank, including
// inside the rows of a repeatable field.
func assertNoEmptyFields(t *testing.T, where string, fields []config.Field, values map[string]any) {
	t.Helper()
	for _, f := range fields {
		switch f.Type {
		case "dropdown":
			if s, _ := values[f.Key].(string); s == "" {
				t.Errorf("%s: dropdown %q is empty", where, f.Key)
			}
		case "multiselect", "conditions":
			rows, _ := values[f.Key].([]map[string]any)
			if len(rows) == 0 && f.DefaultCount > 0 {
				t.Errorf("%s: repeatable field %q has no rows", where, f.Key)
			}
			for ri, row := range rows {
				assertNoEmptyFields(t,
					fmt.Sprintf("%s row %d of %q", where, ri, f.Key), f.RowFields, row)
			}
		}
	}
}
