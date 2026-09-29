package config

import (
	"strings"
	"testing"
)

// filterByList decides which enactments and interactions the builder offers for
// a given component. It is display-only and never enforced on save, but getting
// it wrong either hides valid choices or offers ones a ruleset meant to block,
// so the precedence between the allow and block lists is worth pinning.

func TestFilterByList(t *testing.T) {
	all := []string{"a", "b", "c", "d"}
	tests := []struct {
		name    string
		allowed []string
		blocked []string
		want    string
	}{
		{"neither list returns the input unchanged", nil, nil, "a,b,c,d"},
		{"allow list restricts to its own members", []string{"c", "a"}, nil, "c,a"},
		{"allow list is returned in its own order, not the input order", []string{"d", "b"}, nil, "d,b"},
		{"allow list silently drops ids that do not exist", []string{"a", "zz"}, nil, "a"},
		{"block list removes its members and keeps input order", nil, []string{"b"}, "a,c,d"},
		{"block list ignores ids that do not exist", nil, []string{"zz"}, "a,b,c,d"},
		// The allow list is checked first, so a ruleset that sets both gets the
		// allow list verbatim. Without this the two lists would silently compose
		// and an id could be both permitted and removed.
		{"allow list wins when both are set", []string{"a", "b"}, []string{"a"}, "a,b"},
		{"an empty allow list is treated as unset", []string{}, []string{"a"}, "b,c,d"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strings.Join(filterByList(all, tt.allowed, tt.blocked), ",")
			if got != tt.want {
				t.Errorf("filterByList(%v, allowed=%v, blocked=%v) = %q, want %q",
					all, tt.allowed, tt.blocked, got, tt.want)
			}
		})
	}
}

// TestComponentByKind checks the kind dispatch that backs the inline builder. An
// unknown kind or id must report not-found rather than returning a zero
// Component that would render as an empty form.
func TestComponentByKind(t *testing.T) {
	cfg := &Config{
		PerkTypes:    ComponentMap{Order: []string{"execution"}, Items: map[string]*Component{"execution": {ID: "execution", Name: "Execution"}}},
		Enactments:   ComponentMap{Order: []string{"damage"}, Items: map[string]*Component{"damage": {ID: "damage", Name: "Damage"}}},
		Interactions: ComponentMap{Order: []string{"attack"}, Items: map[string]*Component{"attack": {ID: "attack", Name: "Attack"}}},
	}

	tests := []struct {
		kind, id string
		wantName string
		wantOK   bool
	}{
		{"perk_type", "execution", "Execution", true},
		{"enactment", "damage", "Damage", true},
		{"interaction", "attack", "Attack", true},
		{"enactment", "nonexistent", "", false},
		{"not_a_kind", "damage", "", false},
		{"", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.kind+"/"+tt.id, func(t *testing.T) {
			comp, ok := cfg.ComponentByKind(tt.kind, tt.id)
			if ok != tt.wantOK {
				t.Fatalf("ComponentByKind(%q, %q) ok = %v, want %v", tt.kind, tt.id, ok, tt.wantOK)
			}
			if comp.Name != tt.wantName {
				t.Errorf("ComponentByKind(%q, %q).Name = %q, want %q", tt.kind, tt.id, comp.Name, tt.wantName)
			}
		})
	}
}

// TestEnactmentsForUnknownPerkType covers the fallback: an id the ruleset no
// longer defines must yield the full list rather than an empty one, so a perk
// built against an older config still opens with usable choices.
func TestEnactmentsForUnknownPerkType(t *testing.T) {
	cfg := &Config{
		Enactments: ComponentMap{
			Order: []string{"damage", "heal"},
			Items: map[string]*Component{"damage": {ID: "damage"}, "heal": {ID: "heal"}},
		},
		PerkTypes: ComponentMap{
			Order: []string{"limited"},
			Items: map[string]*Component{"limited": {ID: "limited", AllowedEnactments: []string{"heal"}}},
		},
	}

	if got := cfg.EnactmentsFor("nonexistent"); len(got) != 2 {
		t.Errorf("EnactmentsFor(unknown) returned %d enactments, want all 2", len(got))
	}
	got := cfg.EnactmentsFor("limited")
	if len(got) != 1 || got[0].ID != "heal" {
		t.Errorf("EnactmentsFor(limited) did not apply the allow list, got %d entries", len(got))
	}
}
