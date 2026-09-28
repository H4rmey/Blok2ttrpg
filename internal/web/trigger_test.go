package web

import (
	"strings"
	"testing"

	"github.com/harmey/blok2ttrpg-v5/internal/engine"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// reactionPerk builds a minimal reaction carrying a trigger and watch radius.
// The enactment is incidental; these tests are about the perk-level trigger.
func reactionPerk(trigger string, rangeVal any) model.Perk {
	f := map[string]any{"trigger": trigger}
	if rangeVal != nil {
		f["trigger_range"] = rangeVal
	}
	return model.Perk{
		Name:       "Test Reaction",
		Type:       "reaction",
		Fields:     f,
		Enactments: []model.Enactment{{Type: "negation", Interaction: "self"}},
	}
}

// TestReactionInstructionStatesTrigger pins that a reaction's generated rules
// text says when it fires. A reaction is not used on your turn, so without this
// the instructions describe the effect but never the circumstance - the one
// thing a player has to know to use it at all.
func TestReactionInstructionStatesTrigger(t *testing.T) {
	cfg := loadCfg(t)
	ins := engine.PerkInstructions(cfg, reactionPerk("attack_misses", 0))
	if len(ins) == 0 {
		t.Fatal("no instructions generated")
	}
	if ins[0].Trigger == "" {
		t.Fatal("first instruction has no trigger line")
	}
	// The text is the configured label, not the raw id, so a player reads a
	// sentence rather than a key.
	if strings.Contains(ins[0].Trigger, "attack_misses") {
		t.Errorf("trigger shows the raw id instead of its label: %q", ins[0].Trigger)
	}
	if !strings.Contains(ins[0].Trigger, "misses") {
		t.Errorf("trigger does not describe the circumstance: %q", ins[0].Trigger)
	}
}

// TestTriggerRangeZeroReadsAsSelf pins the distinction that separates a
// self-defence reaction from one that guards a neighbour. Range 0 is not "no
// range" - it restricts the trigger to the engager, and the text has to say so
// or Quick Dodge and Bodyguard read identically.
func TestTriggerRangeZeroReadsAsSelf(t *testing.T) {
	cfg := loadCfg(t)
	self := engine.PerkInstructions(cfg, reactionPerk("attacked", 0))[0].Trigger
	ally := engine.PerkInstructions(cfg, reactionPerk("attacked", 1))[0].Trigger

	if self == ally {
		t.Fatalf("range 0 and range 1 produced identical text: %q", self)
	}
	// At range 0 the trigger is about the engager, so the text must not still be
	// talking about "someone" - that is the phrasing bug this replaced.
	if strings.Contains(strings.ToLower(self), "someone") {
		t.Errorf("range 0 still reads as happening to someone else: %q", self)
	}
	if !strings.Contains(strings.ToLower(self), "you") {
		t.Errorf("range 0 does not name you as the subject: %q", self)
	}
	// At a real radius the figure has to appear, and the placeholder must be
	// gone: leaving both produces "someone within range ... within 1m".
	if !strings.Contains(ally, "1m") {
		t.Errorf("range 1 does not state the radius: %q", ally)
	}
	if strings.Contains(ally, "within range") {
		t.Errorf("range 1 left the unsubstituted placeholder in: %q", ally)
	}
}

// TestNonReactionHasNoTriggerLine pins that the line is absent rather than empty
// or misleading on every other perk type. The generator is asked for a trigger
// unconditionally, so this guards against a stray line on an execution.
func TestNonReactionHasNoTriggerLine(t *testing.T) {
	cfg := loadCfg(t)
	execution := model.Perk{
		Name:       "Plain Execution",
		Type:       "execution",
		Fields:     map[string]any{"comment": "no trigger here"},
		Enactments: []model.Enactment{{Type: "damage", Interaction: "direct"}},
	}
	for _, in := range engine.PerkInstructions(cfg, execution) {
		if in.Trigger != "" {
			t.Errorf("execution carries a trigger line: %q", in.Trigger)
		}
	}
}

// TestTriggerStatedOnceNotPerEnactment pins that the trigger appears on the
// first enactment only. One trigger fires the whole perk, so repeating it on
// every enactment would imply each one waits for its own.
func TestTriggerStatedOnceNotPerEnactment(t *testing.T) {
	cfg := loadCfg(t)
	multi := reactionPerk("attack_hits", 0)
	multi.Enactments = append(multi.Enactments,
		model.Enactment{Type: "damage", Interaction: "direct", NewTarget: true})

	ins := engine.PerkInstructions(cfg, multi)
	if len(ins) < 2 {
		t.Fatalf("expected 2 instructions, got %d", len(ins))
	}
	if ins[0].Trigger == "" {
		t.Error("first enactment is missing the trigger")
	}
	for i, in := range ins[1:] {
		if in.Trigger != "" {
			t.Errorf("enactment %d repeats the trigger: %q", i+2, in.Trigger)
		}
	}
}

// TestUnknownTriggerStillRenders pins the fallback. A stored trigger the config
// no longer offers must degrade to its raw id rather than silently dropping the
// line, so an outdated perk stays legible instead of looking like an execution.
func TestUnknownTriggerStillRenders(t *testing.T) {
	cfg := loadCfg(t)
	ins := engine.PerkInstructions(cfg, reactionPerk("no_such_trigger", nil))
	if ins[0].Trigger == "" {
		t.Fatal("unknown trigger produced no line at all")
	}
	if !strings.Contains(ins[0].Trigger, "no_such_trigger") {
		t.Errorf("unknown trigger did not fall back to its id: %q", ins[0].Trigger)
	}
}

// TestEveryLibraryReactionStatesItsTrigger is the content guard: every reaction
// that ships in the library must produce a trigger line. This is what would have
// caught the original problem, where perks that were conceptually reactions were
// typed as executions and so never told the player when they fired.
func TestEveryLibraryReactionStatesItsTrigger(t *testing.T) {
	app, _ := testAppWithPerk(t)
	perks, err := app.Library.ListPerks()
	if err != nil {
		t.Fatalf("list perks: %v", err)
	}
	seen := 0
	for _, p := range perks {
		if p.Type != "reaction" {
			continue
		}
		seen++
		norm := engine.NormalizePerk(app.Cfg.Config, p)
		ins := engine.PerkInstructions(app.Cfg.Config, norm)
		if len(ins) == 0 {
			t.Errorf("%s: no instructions generated", p.Name)
			continue
		}
		if ins[0].Trigger == "" {
			t.Errorf("%s is a reaction but its instructions never say when it fires", p.Name)
		}
	}
	if seen == 0 {
		t.Error("no reactions in the library: the reaction perk type is undemonstrated")
	}
}
