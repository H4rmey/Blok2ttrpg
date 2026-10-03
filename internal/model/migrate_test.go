package model

import (
	"encoding/json"
	"testing"
)

// TestUnmarshalLegacyLayout checks that a character saved before the rename is
// reinterpreted correctly: the old "traits" roster must land in Skills and the
// old "attributes" section must land in Traits. Getting this backwards would
// silently blank every character's skills, so it is worth pinning down.
func TestUnmarshalLegacyLayout(t *testing.T) {
	const legacy = `{
		"id": "char-1",
		"level": 3,
		"attributes": {"name": "micheal", "current_hp": "10"},
		"traits": {"general.Perception": "novice", "vital.HP": "untrained"}
	}`

	var c Character
	if err := json.Unmarshal([]byte(legacy), &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := c.Traits["name"]; got != "micheal" {
		t.Errorf("Traits[name] = %v, want micheal (old attributes should become Traits)", got)
	}
	if got := c.Skills["general.Perception"]; got != "novice" {
		t.Errorf("Skills[general.Perception] = %q, want novice (old traits should become Skills)", got)
	}
	if c.Level != 3 {
		t.Errorf("Level = %d, want 3", c.Level)
	}
}

// TestUnmarshalCurrentLayout checks the post-rename layout is read as-is, and in
// particular that a document whose "traits" key holds sheet fields is not
// mistaken for the legacy roster.
func TestUnmarshalCurrentLayout(t *testing.T) {
	const current = `{
		"id": "char-2",
		"level": 1,
		"traits": {"name": "aster"},
		"skills": {"general.Stealth": "master"}
	}`

	var c Character
	if err := json.Unmarshal([]byte(current), &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := c.Traits["name"]; got != "aster" {
		t.Errorf("Traits[name] = %v, want aster", got)
	}
	if got := c.Skills["general.Stealth"]; got != "master" {
		t.Errorf("Skills[general.Stealth] = %q, want master", got)
	}
}

// TestUnmarshalRoundTrip checks that re-reading a character the app has written
// produces the same value, so the migration cannot flip a document back and
// forth between layouts.
func TestUnmarshalRoundTrip(t *testing.T) {
	in := Character{
		ID:     "char-3",
		Level:  2,
		Traits: map[string]any{"name": "rook"},
		Skills: map[string]string{"offense.Strength": "proficient"},
	}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out Character
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Traits["name"] != "rook" || out.Skills["offense.Strength"] != "proficient" {
		t.Errorf("round trip changed the character: %+v", out)
	}
}

// TestUnmarshalKeepsShifts guards the characterJSON mirror: the mirror struct
// fully replaces Character for decoding, so a field missing from it is silently
// dropped on READ - the code compiles and every saved shift card quietly
// disappears the first time the character is opened. Adding "shifts" to a
// stored document therefore needs the mirror updated, and this is the test
// that catches it if it ever is not.
func TestUnmarshalKeepsShifts(t *testing.T) {
	const doc = `{
		"id": "char-5",
		"level": 1,
		"traits": {"name": "wren"},
		"skills": {"offense.Strength": "novice"},
		"shifts": [{"skill_key": "offense.Strength", "shift": -1}]
	}`

	var c Character
	if err := json.Unmarshal([]byte(doc), &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(c.Shifts) != 1 {
		t.Fatalf("shifts decoded as %d cards, want 1: the characterJSON mirror "+
			"must list every live field or JSON drops it on read", len(c.Shifts))
	}
	if c.Shifts[0].SkillKey != "offense.Strength" || c.Shifts[0].Shift != -1 {
		t.Errorf("shift card = %+v, want offense.Strength at -1", c.Shifts[0])
	}

	// And the same through the app's own write path.
	in := c
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out Character
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(out.Shifts) != 1 || out.Shifts[0].Shift != -1 {
		t.Errorf("round trip dropped or changed the shift cards: %+v", out.Shifts)
	}
}

// TestUnmarshalEmptyMapsAreInitialised guards the handlers, which write straight
// into these maps without a nil check.
func TestUnmarshalEmptyMapsAreInitialised(t *testing.T) {
	var c Character
	if err := json.Unmarshal([]byte(`{"id":"char-4","level":1}`), &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if c.Traits == nil || c.Skills == nil {
		t.Errorf("Traits/Skills must be non-nil after decode, got %v / %v", c.Traits, c.Skills)
	}
}
