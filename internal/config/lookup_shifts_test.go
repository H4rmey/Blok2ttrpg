package config

import "testing"

// TestEnactShiftOptionsFallback checks the range a shift card offers when the
// ruleset does not define an Enact Shift enactment (or its shift-amount field):
// the documented -6..6 range minus zero. Zero is skipped for the same reason
// the condition dropdowns skip it - a shift card at 0 does nothing.
func TestEnactShiftOptionsFallback(t *testing.T) {
	opts := (&Config{}).EnactShiftOptions()
	if len(opts) != 12 {
		t.Fatalf("got %d options, want 12 (-6..6 minus 0): %v", len(opts), opts)
	}
	for _, v := range opts {
		if v == 0 {
			t.Errorf("0 must not be offered: %v", opts)
		}
	}
	if len(opts) > 0 && (opts[0] != -6 || opts[len(opts)-1] != 6) {
		t.Errorf("range = %v, want -6..6 (skipping 0)", opts)
	}
}

// TestEnactShiftOptionsFromField checks the range is read from the shift
// enactment's shift-amount field (min/max/step), so a ruleset can re-tune the
// play card in YAML rather than in code.
func TestEnactShiftOptionsFromField(t *testing.T) {
	cfg := &Config{
		Enactments: ComponentMap{
			Order: []string{"shift"},
			Items: map[string]*Component{
				"shift": {
					ID: "shift",
					Fields: []Field{
						{Key: "shift-amount", Type: "free_number", Min: -4, Max: 4, Step: 2},
					},
				},
			},
		},
	}
	opts := cfg.EnactShiftOptions()
	if len(opts) != 4 {
		t.Fatalf("got %v, want [-4 -2 2 4]", opts)
	}
	for i, want := range []int{-4, -2, 2, 4} {
		if opts[i] != want {
			t.Errorf("opts[%d] = %d, want %d", i, opts[i], want)
		}
	}
}

// TestSmallestShift pins the apply default: the weakest magnitude, so applying
// a card never silently imposes the largest possible penalty. A symmetric
// range holds both -1 and +1; the first candidate wins the tie, mirroring the
// historical condition default.
func TestSmallestShift(t *testing.T) {
	cases := []struct {
		name string
		opts []int
		want int
	}{
		{"symmetric range ties first candidate", []int{-6, -5, -4, -3, -2, -1, 1, 2, 3, 4, 5, 6}, -1},
		{"negative-only range", []int{-6, -5, -4, -3, -2, -1}, -1},
		{"positive-only range", []int{1, 2, 3}, 1},
		{"empty range", nil, 0},
	}
	for _, tc := range cases {
		if got := SmallestShift(tc.opts); got != tc.want {
			t.Errorf("%s: SmallestShift(%v) = %d, want %d", tc.name, tc.opts, got, tc.want)
		}
	}
}
