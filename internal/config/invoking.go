// Schema and accessors for the two play-time subsystems: invoking (the invoke
// point economy, including reactions) and negotiation (the structured social
// encounter).
//
// Neither subsystem is simulated by the application. Both are sets of rules
// parameters that the table applies at play, and the reason they live in the
// config rather than in prose is so the generated documentation reads its
// numbers from the same place a future engine would. Nothing here is allowed to
// hardcode a value that a ruleset author should be able to change.
package config

// InvokeTable describes how the invoke point pool grows with level.
//
// It is deliberately not a LevelTable. The two build pools grow every level, so
// their formula is start + per_level*(level-1); the invoke pool is small enough
// that a point every level would swamp it, so it grants PerStep points every
// LevelsPerStep levels instead. Levels overrides the formula per level when a
// ruleset wants an irregular curve.
type InvokeTable struct {
	// Start is the pool size at level 1.
	Start int `yaml:"start,omitempty" json:"start,omitempty"`
	// PerStep is how many points each step grants.
	PerStep int `yaml:"per_step,omitempty" json:"per_step,omitempty"`
	// LevelsPerStep is how many levels apart the steps are. A value of 2 grants
	// PerStep points at levels 3, 5, 7, 9, and so on.
	LevelsPerStep int `yaml:"levels_per_step,omitempty" json:"levels_per_step,omitempty"`
	// Levels optionally overrides the formula on a per-level basis.
	Levels []LevelEntry `yaml:"levels,omitempty" json:"levels,omitempty"`
}

// Invoking is the invoke point economy.
type Invoking struct {
	// Refresh names the moment the pool returns to its maximum, as a phrase
	// rendered into the rulebook (e.g. "the start of every session").
	Refresh string `yaml:"refresh,omitempty" json:"refresh,omitempty"`

	// AllowOverMaximum permits points earned in play to exceed the per-level
	// maximum. It is a pointer so an unset value defaults to false, which keeps
	// the maximum meaningful.
	AllowOverMaximum *bool `yaml:"allow_over_maximum,omitempty" json:"allow_over_maximum,omitempty"`

	// Spends are the ways a point is spent; Gains are the ways one is earned.
	Spends []InvokeSpend `yaml:"spends,omitempty" json:"spends,omitempty"`
	Gains  []InvokeGain  `yaml:"gains,omitempty" json:"gains,omitempty"`

	// CombatGains are the extra in-combat earning triggers.
	CombatGains CombatGains `yaml:"combat_gains,omitempty" json:"combat_gains,omitempty"`

	// Reactions holds the out-of-turn action limits.
	Reactions Reactions `yaml:"reactions,omitempty" json:"reactions,omitempty"`
}

// AllowsOverMaximum reports whether the invoke pool may exceed its maximum.
// Defaults to false when unset.
func (i Invoking) AllowsOverMaximum() bool {
	return i.AllowOverMaximum != nil && *i.AllowOverMaximum
}

// InvokeSpend is one way to spend invoke points.
type InvokeSpend struct {
	ID          string `yaml:"id" json:"id"`
	Name        string `yaml:"name" json:"name"`
	InvokeCost  int    `yaml:"invoke_cost,omitempty" json:"invoke_cost,omitempty"`
	EnergyCost  int    `yaml:"energy_cost,omitempty" json:"energy_cost,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// InvokeGain is one way to earn invoke points.
type InvokeGain struct {
	ID          string `yaml:"id" json:"id"`
	Name        string `yaml:"name" json:"name"`
	Points      int    `yaml:"points,omitempty" json:"points,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// CombatGains is the set of in-combat earning triggers plus the per-combat cap
// that keeps them from turning a long fight into an endless supply.
type CombatGains struct {
	MaxPerCombat int                 `yaml:"max_per_combat,omitempty" json:"max_per_combat,omitempty"`
	Triggers     []CombatGainTrigger `yaml:"triggers,omitempty" json:"triggers,omitempty"`
}

// CombatGainTrigger is one in-combat way to earn a point.
type CombatGainTrigger struct {
	ID     string `yaml:"id" json:"id"`
	Name   string `yaml:"name" json:"name"`
	Points int    `yaml:"points,omitempty" json:"points,omitempty"`
	// OncePerCombat limits the trigger to a single use per fight.
	OncePerCombat bool `yaml:"once_per_combat,omitempty" json:"once_per_combat,omitempty"`
	// EveryRounds, when non-zero, makes the trigger fire at the start of every
	// Nth round rather than on an event.
	EveryRounds int    `yaml:"every_rounds,omitempty" json:"every_rounds,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// Reactions configures acting out of turn: what it costs and how often it may
// be done. There are two routes to it and both are priced here.
//
// A freeform reaction is improvised at the table and pays InvokeCost plus
// EnergyCost. A prebuilt reaction is the Reaction perk type: it was bought with
// build points, so it pays PrebuiltInvokeCost (normally zero) and only its own
// perk energy cost. MaxPerRound bounds them together when SharedPerRound is set.
type Reactions struct {
	InvokeCost int `yaml:"invoke_cost,omitempty" json:"invoke_cost,omitempty"`
	EnergyCost int `yaml:"energy_cost,omitempty" json:"energy_cost,omitempty"`

	// PrebuiltInvokeCost is what a Reaction perk costs in invoke points to
	// fire. It is normally zero: the invoke point is considered pre-paid by the
	// build point spent on the perk.
	PrebuiltInvokeCost int `yaml:"prebuilt_invoke_cost,omitempty" json:"prebuilt_invoke_cost,omitempty"`

	MaxPerRound int `yaml:"max_per_round,omitempty" json:"max_per_round,omitempty"`

	// SharedPerRound reports whether MaxPerRound is a single budget covering
	// freeform and prebuilt reactions together. It is a pointer so an unset
	// value defaults to true, which is the limit that keeps a character with
	// several reaction perks from taking one of each in the same round.
	SharedPerRound *bool `yaml:"shared_per_round,omitempty" json:"shared_per_round,omitempty"`

	Timing      string `yaml:"timing,omitempty" json:"timing,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// SharesPerRound reports whether the per-round reaction limit is shared between
// freeform and prebuilt reactions. Defaults to true when unset.
func (r Reactions) SharesPerRound() bool {
	return r.SharedPerRound == nil || *r.SharedPerRound
}

// Negotiation is the structured social encounter.
type Negotiation struct {
	Motivation     Motivation     `yaml:"motivation,omitempty" json:"motivation,omitempty"`
	Patience       Patience       `yaml:"patience,omitempty" json:"patience,omitempty"`
	Argument       Argument       `yaml:"argument,omitempty" json:"argument,omitempty"`
	TraitAlignment TraitAlignment `yaml:"trait_alignment,omitempty" json:"trait_alignment,omitempty"`
}

// Motivation is the ladder of outcomes a negotiation can land on.
type Motivation struct {
	Min   int              `yaml:"min,omitempty" json:"min,omitempty"`
	Max   int              `yaml:"max,omitempty" json:"max,omitempty"`
	Rungs []MotivationRung `yaml:"rungs,omitempty" json:"rungs,omitempty"`
}

// MotivationRung is one step of the motivation ladder. Outcome describes what
// happens when a negotiation ends here; it is guidance for the GM rather than a
// result to be applied mechanically.
type MotivationRung struct {
	ID      string `yaml:"id" json:"id"`
	Value   int    `yaml:"value" json:"value"`
	Name    string `yaml:"name" json:"name"`
	Outcome string `yaml:"outcome,omitempty" json:"outcome,omitempty"`
}

// Patience is the countdown that bounds a negotiation. The starting value is set
// per NPC by the GM, so only the range and the per-argument cost live here.
type Patience struct {
	Min int `yaml:"min,omitempty" json:"min,omitempty"`
	Max int `yaml:"max,omitempty" json:"max,omitempty"`
	// LossPerArgument is the patience spent by each argument, win or lose.
	LossPerArgument int `yaml:"loss_per_argument,omitempty" json:"loss_per_argument,omitempty"`
}

// Argument is how one attempt to persuade resolves.
type Argument struct {
	// OnSuccess/OnFailure are the motivation deltas applied by a successful or
	// failed skill check.
	OnSuccess int `yaml:"on_success,omitempty" json:"on_success,omitempty"`
	OnFailure int `yaml:"on_failure,omitempty" json:"on_failure,omitempty"`
	// AllowRepeats permits the same argument or trait to be used more than once
	// in a negotiation. It is a pointer so an unset value defaults to false.
	AllowRepeats *bool `yaml:"allow_repeats,omitempty" json:"allow_repeats,omitempty"`
}

// AllowsRepeats reports whether an argument may be reused within a single
// negotiation. Defaults to false when unset.
func (a Argument) AllowsRepeats() bool {
	return a.AllowRepeats != nil && *a.AllowRepeats
}

// TraitAlignment describes how an NPC's own traits clamp the motivation ladder.
// The two values are clamp names ("cannot_increase" / "cannot_decrease") rather
// than booleans so a ruleset can describe a different relationship without a
// code change.
type TraitAlignment struct {
	Against string `yaml:"against,omitempty" json:"against,omitempty"`
	With    string `yaml:"with,omitempty" json:"with,omitempty"`
}

// invokingConfigured reports whether a section file actually declared an
// invoking block. Every sub-part is checked because a ruleset may legitimately
// configure only some of them (for example reactions but no combat gains).
func invokingConfigured(i Invoking) bool {
	return i.Refresh != "" ||
		i.AllowOverMaximum != nil ||
		len(i.Spends) > 0 ||
		len(i.Gains) > 0 ||
		i.CombatGains.MaxPerCombat != 0 ||
		len(i.CombatGains.Triggers) > 0 ||
		reactionsConfigured(i.Reactions)
}

// reactionsConfigured reports whether a reactions block was declared. It cannot
// be a struct comparison against the zero value because Reactions now holds a
// pointer field (SharedPerRound), which makes the type non-comparable.
func reactionsConfigured(r Reactions) bool {
	return r.InvokeCost != 0 ||
		r.EnergyCost != 0 ||
		r.PrebuiltInvokeCost != 0 ||
		r.MaxPerRound != 0 ||
		r.SharedPerRound != nil ||
		r.Timing != "" ||
		r.Description != ""
}

// negotiationConfigured reports whether a section file actually declared a
// negotiation block.
func negotiationConfigured(n Negotiation) bool {
	return len(n.Motivation.Rungs) > 0 ||
		n.Motivation.Min != 0 || n.Motivation.Max != 0 ||
		n.Patience != (Patience{}) ||
		n.Argument.OnSuccess != 0 || n.Argument.OnFailure != 0 ||
		n.Argument.AllowRepeats != nil ||
		n.TraitAlignment != (TraitAlignment{})
}

// InvokePointBudget returns the invoke point pool for a given character level,

// clamped to the configured maximum level first so an out-of-range level never
// grants more than the cap.
func (c *Config) InvokePointBudget(level int) int {
	return invokeBudgetForLevel(c.Leveling.InvokePoints, c.ClampLevel(level))
}

// invokeBudgetForLevel resolves the invoke pool for a level. An explicit row for
// the requested level always wins, so a ruleset can override the curve. Failing
// that the pool is Start plus PerStep for every completed step of
// LevelsPerStep levels above level 1, which is the formula documented in
// leveling.yaml.
func invokeBudgetForLevel(t InvokeTable, level int) int {
	if level < 1 {
		level = 1
	}
	for _, e := range t.Levels {
		if e.Level == level {
			return e.Total
		}
	}
	total := t.Start
	if t.PerStep != 0 && t.LevelsPerStep > 0 {
		total += t.PerStep * ((level - 1) / t.LevelsPerStep)
	}
	return total
}

// MotivationRungByValue returns the ladder rung with the given value.
func (c *Config) MotivationRungByValue(value int) (MotivationRung, bool) {
	for _, r := range c.Negotiation.Motivation.Rungs {
		if r.Value == value {
			return r, true
		}
	}
	return MotivationRung{}, false
}
