package web

import (
	"testing"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
)

// Golden-output tests for the inline cost hints. They exist because the
// templates call these through "{{ with costHint ... }}", which means the
// empty-string cases are load-bearing: returning "" suppresses the surrounding
// <span class="cost-hint"> entirely, while returning "()" or " " would emit an
// empty badge on every uncosted field in the builder.
//
// Pass 10 was going to move these behind the engine as "template funcs carrying
// business rules". Reading them closely, they are not rules: they take an
// already-computed config.Cost and format it for display. The rule-bearing
// helper in this file is firstOption, which is covered separately below. So the
// hints are pinned here rather than moved, and the cost engine stays the single
// place that decides what a cost *is*.

func TestCostHintStr(t *testing.T) {
	tests := []struct {
		name string
		cost *config.Cost
		want string
	}{
		{"nil cost renders nothing", nil, ""},
		{"an all-zero cost renders nothing", &config.Cost{}, ""},
		{"build only", &config.Cost{BuildCost: 2}, "(+2 pt)"},
		{"energy only", &config.Cost{EnergyCost: 1}, "(+1 E)"},
		{"both components", &config.Cost{BuildCost: 4, EnergyCost: 2}, "(+4 pt, +2 E)"},
		// Refunds are why the sign is explicit rather than implied: "-2 pt" has
		// to read as giving points back, not as a typo.
		{"a refund keeps its minus sign", &config.Cost{BuildCost: -2}, "(-2 pt)"},
		{"mixed signs", &config.Cost{BuildCost: -2, EnergyCost: 1}, "(-2 pt, +1 E)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := costHintStr(tt.cost); got != tt.want {
				t.Errorf("costHintStr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPerStepHintStr(t *testing.T) {
	inc := &config.Cost{BuildCost: 2, EnergyCost: 1}
	dec := &config.Cost{BuildCost: -1}

	tests := []struct {
		name string
		step *config.PerStep
		want string
	}{
		{"nil renders nothing", nil, ""},
		{"an empty per-step renders nothing", &config.PerStep{}, ""},
		{"a zero increase renders nothing", &config.PerStep{Increase: &config.Cost{}}, ""},
		{"increase only", &config.PerStep{Increase: inc}, "(+2 pt, +1 E / step)"},
		{"decrease only", &config.PerStep{Decrease: dec}, "(-step: -1 pt)"},
		{"both are shown separately", &config.PerStep{Increase: inc, Decrease: dec},
			"(+step: +2 pt, +1 E / -step: -1 pt)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := perStepHintStr(tt.step); got != tt.want {
				t.Errorf("perStepHintStr() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestCostHintsAreEmptyNotBlank is the guard that protects the template. A hint
// that is whitespace or "()" is truthy to "{{ with }}", so it would render an
// empty badge rather than no badge at all. Asserting exact emptiness is the only
// way to keep that from regressing silently.
func TestCostHintsAreEmptyNotBlank(t *testing.T) {
	for _, c := range []*config.Cost{nil, {}, {BuildCost: 0, EnergyCost: 0}} {
		if got := costHintStr(c); got != "" {
			t.Errorf("costHintStr(%+v) = %q, want exactly \"\" so the template omits the span", c, got)
		}
	}
	for _, p := range []*config.PerStep{nil, {}, {Increase: &config.Cost{}, Decrease: &config.Cost{}}} {
		if got := perStepHintStr(p); got != "" {
			t.Errorf("perStepHintStr(%+v) = %q, want exactly \"\" so the template omits the span", p, got)
		}
	}
}

// TestFirstOptionMatchesEngineNormalization pins the one genuinely rule-bearing
// helper here. Its own comment says it "mirrors the engine's normalization", and
// that duplication is the smell pass 10 targeted: the builder picks a dropdown's
// fallback value here, while the engine picks one again when it normalizes the
// stored perk. If the two ever disagree, the form shows one value and the saved
// perk is priced as another.
//
// Rather than move the logic (which would change what the template emits), this
// asserts the contract that makes the duplication safe: firstOption must return
// the first non-empty resolved option, which is exactly what the engine fills in.
func TestFirstOptionMatchesEngineNormalization(t *testing.T) {
	fm := funcMap()
	firstOption, ok := fm["firstOption"].(func(*config.Config, config.Field) string)
	if !ok {
		t.Fatal("firstOption is not registered with the expected signature")
	}

	cfg := &config.Config{}

	tests := []struct {
		name  string
		field config.Field
		want  string
	}{
		{
			name:  "picks the first inline option",
			field: config.Field{Options: []config.Option{{Value: "a"}, {Value: "b"}}},
			want:  "a",
		},
		{
			// There is no empty "none" choice in this UI, so a blank leading
			// value must be skipped rather than selected.
			name:  "skips a leading empty value",
			field: config.Field{Options: []config.Option{{Value: ""}, {Value: "b"}}},
			want:  "b",
		},
		{
			name:  "no options yields no fallback",
			field: config.Field{},
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := firstOption(cfg, tt.field); got != tt.want {
				t.Errorf("firstOption() = %q, want %q", got, tt.want)
			}
		})
	}

	// A nil config must not panic: the printable templates render without one.
	if got := firstOption(nil, config.Field{Options: []config.Option{{Value: "a"}}}); got != "" {
		t.Errorf("firstOption(nil cfg) = %q, want \"\"", got)
	}
}
