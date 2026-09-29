// Config-derived reference tables for the invoking and negotiation chapters.
//
// Every number these render is read from the loaded ruleset rather than written
// into the markdown, for the same reason the leveling tables are: a rules
// parameter that appears in two places will eventually disagree with itself.
// Each helper returns a "_No ... configured._" placeholder when its section is
// absent, which the docs lint reports, so a chapter cannot silently lose its
// numbers.
package docs

import (
	"fmt"
	"strings"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
)

// invokePointsTable renders the invoke point pool for every level as
// Level / Points Gained / Total, using the same accessor the application uses to
// hand a character its pool.
func invokePointsTable(cfg *config.Config) string {
	if cfg == nil {
		return "_No leveling configured._"
	}
	max := cfg.MaxLevel()
	if max < 1 {
		return "_No leveling configured._"
	}
	var b strings.Builder
	b.WriteString("| Level | Points Gained | Invoke Points |\n")
	b.WriteString("| --- | --- | --- |\n")
	prev := 0
	for level := 1; level <= max; level++ {
		total := cfg.InvokePointBudget(level)
		gained := total - prev
		if level == 1 {
			fmt.Fprintf(&b, "| **%d** | %d (starting) | %d |\n", level, total, total)
		} else if gained == 0 {
			// Most levels grant nothing, because the pool steps rather than
			// rising every level. Saying so explicitly is clearer than a zero.
			fmt.Fprintf(&b, "| **%d** | - | %d |\n", level, total)
		} else {
			fmt.Fprintf(&b, "| **%d** | %s | %d |\n", level, signed(gained), total)
		}
		prev = total
	}
	return strings.TrimSpace(b.String())
}

// invokeSpendsTable renders what an invoke point buys.
func invokeSpendsTable(cfg *config.Config) string {
	if cfg == nil || len(cfg.Invoking.Spends) == 0 {
		return "_No invoke spends configured._"
	}
	var b strings.Builder
	b.WriteString("| Spend | Cost | Effect |\n")
	b.WriteString("| --- | --- | --- |\n")
	for _, s := range cfg.Invoking.Spends {
		fmt.Fprintf(&b, "| **%s** | %s | %s |\n",
			orDash(s.Name), invokeCostWords(s.InvokeCost, s.EnergyCost), orDash(oneLine(s.Description)))
	}
	return strings.TrimSpace(b.String())
}

// invokeGainsTable renders how an invoke point is earned outside combat.
func invokeGainsTable(cfg *config.Config) string {
	if cfg == nil || len(cfg.Invoking.Gains) == 0 {
		return "_No invoke gains configured._"
	}
	var b strings.Builder
	b.WriteString("| How | Points | Detail |\n")
	b.WriteString("| --- | --- | --- |\n")
	for _, g := range cfg.Invoking.Gains {
		fmt.Fprintf(&b, "| **%s** | %s | %s |\n",
			orDash(g.Name), signed(g.Points), orDash(oneLine(g.Description)))
	}
	return strings.TrimSpace(b.String())
}

// combatGainsTable renders the in-combat earning triggers plus the per-combat
// cap, which is the limit that keeps them from becoming a faucet.
func combatGainsTable(cfg *config.Config) string {
	if cfg == nil || len(cfg.Invoking.CombatGains.Triggers) == 0 {
		return "_No combat invoke gains configured._"
	}
	cg := cfg.Invoking.CombatGains
	var b strings.Builder
	b.WriteString("| Trigger | Points | Limit | Detail |\n")
	b.WriteString("| --- | --- | --- | --- |\n")
	for _, t := range cg.Triggers {
		fmt.Fprintf(&b, "| **%s** | %s | %s | %s |\n",
			orDash(t.Name), signed(t.Points), triggerLimit(t), orDash(oneLine(t.Description)))
	}
	if cg.MaxPerCombat > 0 {
		fmt.Fprintf(&b, "\nNo matter how many of these trigger, a character may earn at most **%d** invoke points from a single combat.", cg.MaxPerCombat)
	}
	return strings.TrimSpace(b.String())
}

// triggerLimit describes how often a combat trigger may fire.
func triggerLimit(t config.CombatGainTrigger) string {
	switch {
	case t.EveryRounds > 0:
		return fmt.Sprintf("Start of every %s round", ordinal(t.EveryRounds))
	case t.OncePerCombat:
		return "Once per combat"
	default:
		return "Each time it happens"
	}
}

// outOfTurnRules renders the cost, frequency and timing of acting out of turn.
// Both routes are covered, because the reader needs to compare them: an invoke
// action bought in play, and a Reaction perk built in advance. They are priced
// differently but share the per-round limit.
func outOfTurnRules(cfg *config.Config) string {
	if cfg == nil {
		return "_No out-of-turn rules configured._"
	}
	a := cfg.Invoking.InvokeActions
	lim := cfg.Invoking.ReactionLimit
	if a == (config.InvokeActions{}) && !reactionLimitDeclared(lim) {
		return "_No out-of-turn rules configured._"
	}
	var b strings.Builder
	b.WriteString("| Rule | Value |\n")
	b.WriteString("| --- | --- |\n")
	fmt.Fprintf(&b, "| %s (improvised) | %s |\n",
		a.DisplayName(), invokeCostWords(a.InvokeCost, a.EnergyCost))
	fmt.Fprintf(&b, "| Reaction perk (built) | %s, plus the perk's own energy cost |\n",
		invokeCostWords(lim.PerkInvokeCost, 0))
	if lim.MaxPerRound > 0 {
		fmt.Fprintf(&b, "| Frequency | %d per round%s |\n", lim.MaxPerRound, sharedPhrase(lim, a))
	}
	if a.Timing != "" {
		fmt.Fprintf(&b, "| Timing | %s |\n", timingPhrase(a.Timing))
	}
	if a.Description != "" {
		fmt.Fprintf(&b, "| What it is | %s |\n", oneLine(a.Description))
	}
	return strings.TrimSpace(b.String())
}

// reactionLimitDeclared reports whether a reaction_limit block carries any
// value. It mirrors the loader's own check and exists because ReactionLimit
// holds a pointer field, so it cannot be compared against its zero value.
func reactionLimitDeclared(r config.ReactionLimit) bool {
	return r.MaxPerRound != 0 ||
		r.Shared != nil ||
		r.PerkInvokeCost != 0
}

// sharedPhrase spells out whether the per-round limit is one budget covering
// both routes or a separate allowance for each.
func sharedPhrase(r config.ReactionLimit, a config.InvokeActions) string {
	if r.IsShared() {
		return fmt.Sprintf(", counting %ss and Reaction perks together. Owning several Reaction perks does not let you act out of turn more than once in a round",
			strings.ToLower(a.DisplayName()))
	}
	return fmt.Sprintf(", counted separately for %ss and Reaction perks",
		strings.ToLower(a.DisplayName()))
}

// energyRules renders the Energy economy: what a perk costs to use, what a rest
// gives back, and what happens when a character cannot pay.
//
// The overdraft rule used to live as prose in the perk builder introduction,
// where a player looking up "I am out of Energy" would never find it, and its
// HP rate was written into the markdown. Reading it from the config means the
// rate can be retuned without touching a sentence.
func energyRules(cfg *config.Config) string {
	if cfg == nil {
		return "_No energy rules configured._"
	}
	var b strings.Builder
	b.WriteString("| Rule | Value |\n")
	b.WriteString("| --- | --- |\n")
	fmt.Fprintf(&b, "| Cost to use a perk | Its own Energy cost, which is 1 per enactment before any options. |\n")
	if cfg.Combat.EnergyRecoveryPerRest > 0 {
		fmt.Fprintf(&b, "| Regained on a rest | %d Energy. Energy never returns on its own. |\n",
			cfg.Combat.EnergyRecoveryPerRest)
	}

	o := cfg.Combat.EnergyOverdraft
	if !config.EnergyOverdraftConfigured(o) {
		fmt.Fprintf(&b, "| Running out | Not configured: a perk you cannot pay for simply cannot be used. |\n")
		return strings.TrimSpace(b.String())
	}

	fmt.Fprintf(&b, "| Running out | You may still use the perk. Pay for the Energy you are missing out of your own HP, or cut the perk short. |\n")
	fmt.Fprintf(&b, "| Overdraft rate | %s |\n", overdraftRateWords(o))
	if o.AllowsPartialExecution() {
		fmt.Fprintf(&b, "| Partial execution | Permitted. Instead of paying HP, drop the enactments you cannot afford: the fireball still burns, it just no longer explodes. |\n")
	} else {
		fmt.Fprintf(&b, "| Partial execution | Not permitted. The only way to use a perk you cannot pay for is to pay the HP. |\n")
	}
	if o.ConditionOnOverdraft != "" {
		fmt.Fprintf(&b, "| Also applies | %s |\n", overdraftConditionWords(cfg, o.ConditionOnOverdraft))
	}
	return strings.TrimSpace(b.String())
}

// overdraftRateWords states the HP-per-Energy rate, including the escalation
// when the ruleset makes repeated overdrafts progressively worse.
func overdraftRateWords(o config.EnergyOverdraft) string {
	base := fmt.Sprintf("%d HP per point of missing Energy", o.HPPerEnergy)
	if o.EscalationPerUse == 0 {
		return base + "."
	}
	return fmt.Sprintf("%s the first time in a scene, rising by %d per point on each overdraft after that (%d, then %d, then %d).",
		base, o.EscalationPerUse,
		o.OverdraftRateForUse(1), o.OverdraftRateForUse(2), o.OverdraftRateForUse(3))
}

// overdraftConditionWords names the condition applied on overdraw and quotes
// what it does, so the reader does not have to look it up.
func overdraftConditionWords(cfg *config.Config, id string) string {
	c, ok := cfg.ConditionByID(id)
	if !ok {
		return fmt.Sprintf("The **%s** condition.", id)
	}
	if c.Description == "" {
		return fmt.Sprintf("The **%s** condition.", c.Name)
	}
	return fmt.Sprintf("The **%s** condition: %s", c.Name, oneLine(c.Description))
}

// reactionTriggersTable renders the triggers a prebuilt Reaction may be built
// around, with the extra build cost each one carries. The list comes from the
// reaction_triggers option source, so adding a trigger in the config adds a row
// here without a code change.
func reactionTriggersTable(cfg *config.Config) string {
	if cfg == nil {
		return "_No reaction triggers configured._"
	}
	opts := cfg.OptionsFor("reaction_triggers")
	if len(opts) == 0 {
		return "_No reaction triggers configured._"
	}
	var b strings.Builder
	b.WriteString("| Trigger | Build Cost | Notes |\n")
	b.WriteString("| --- | --- | --- |\n")
	for _, o := range opts {
		label := o.Label
		if label == "" {
			label = o.Value
		}
		cost := 0
		if o.Cost != nil {
			cost = o.Cost.BuildCost
		}
		fmt.Fprintf(&b, "| **%s** | %s | %s |\n",
			label, buildCostWords(cost), orDash(oneLine(o.Information)))
	}
	return strings.TrimSpace(b.String())
}

// buildCostWords renders a build-point surcharge, naming zero explicitly so a
// free trigger does not read as a missing value.
func buildCostWords(build int) string {
	if build == 0 {
		return "Included"
	}
	return fmt.Sprintf("%s point%s", signed(build), plural(build))
}

// timingPhrase turns a timing id into a reader-facing sentence.
func timingPhrase(timing string) string {
	switch timing {
	case "between_actions":
		return "Between actions. A reaction resolves before or after a whole action, never in the middle of one, so it cannot undo a roll that has already been made."
	case "anytime":
		return "At any moment, including in the middle of another action."
	default:
		return timing
	}
}

// invokeRefreshPhrase states when the pool refills and whether banked points may
// exceed the maximum.
func invokeRefreshPhrase(cfg *config.Config) string {
	if cfg == nil {
		return "-"
	}
	inv := cfg.Invoking
	refresh := inv.Refresh
	if strings.TrimSpace(refresh) == "" {
		refresh = "the start of every session"
	}
	s := fmt.Sprintf("Every character's invoke points return to their maximum at %s.", refresh)
	if inv.AllowsOverMaximum() {
		return s + " Points earned in play may be banked above that maximum."
	}
	return s + " Points earned in play stop at that maximum: a full pool must be spent before more can be banked."
}

// motivationTable renders the motivation ladder, highest rung first so it reads
// as best-to-worst the way a player thinks about it.
func motivationTable(cfg *config.Config) string {
	if cfg == nil || len(cfg.Negotiation.Motivation.Rungs) == 0 {
		return "_No motivation ladder configured._"
	}
	rungs := cfg.Negotiation.Motivation.Rungs
	var b strings.Builder
	b.WriteString("| Motivation | Name | Outcome |\n")
	b.WriteString("| --- | --- | --- |\n")
	for i := len(rungs) - 1; i >= 0; i-- {
		r := rungs[i]
		fmt.Fprintf(&b, "| **%d** | %s | %s |\n", r.Value, orDash(r.Name), orDash(oneLine(r.Outcome)))
	}
	return strings.TrimSpace(b.String())
}

// negotiationRulesTable renders the patience clock and the argument resolution
// rules: the numbers a GM needs at the table.
func negotiationRulesTable(cfg *config.Config) string {
	if cfg == nil {
		return "_No negotiation rules configured._"
	}
	n := cfg.Negotiation
	var b strings.Builder
	b.WriteString("| Rule | Value |\n")
	b.WriteString("| --- | --- |\n")
	fmt.Fprintf(&b, "| Motivation range | %d to %d |\n", n.Motivation.Min, n.Motivation.Max)
	fmt.Fprintf(&b, "| Patience range | %d to %d |\n", n.Patience.Min, n.Patience.Max)
	fmt.Fprintf(&b, "| Patience per argument | %s |\n", patiencePhrase(n.Patience.LossPerArgument))
	fmt.Fprintf(&b, "| Successful argument | %s motivation |\n", signed(n.Argument.OnSuccess))
	fmt.Fprintf(&b, "| Failed argument | %s motivation |\n", signed(n.Argument.OnFailure))
	fmt.Fprintf(&b, "| Repeating an argument | %s |\n", repeatsPhrase(n.Argument.AllowsRepeats()))
	fmt.Fprintf(&b, "| Starting values | Set by the GM per NPC. A guard who owes you a favour and one you just insulted do not open at the same motivation or with the same patience. |\n")
	return strings.TrimSpace(b.String())
}

// patiencePhrase explains the per-argument patience cost.
func patiencePhrase(loss int) string {
	if loss == 0 {
		return "No patience is spent by making an argument."
	}
	return fmt.Sprintf("-%d, whether the argument succeeds or fails. When patience reaches zero the negotiation ends.", loss)
}

// repeatsPhrase explains whether an argument may be reused.
func repeatsPhrase(allow bool) string {
	if allow {
		return "Permitted: the same argument or trait may be raised more than once."
	}
	return "Not permitted. An argument, and the trait behind it, may each be used only once per negotiation."
}

// traitAlignmentTable renders how an NPC's own traits clamp the ladder. This is
// the heart of the negotiation system, so it gets its own table rather than a
// row in the rules above.
func traitAlignmentTable(cfg *config.Config) string {
	if cfg == nil {
		return "_No trait alignment rules configured._"
	}
	ta := cfg.Negotiation.TraitAlignment
	if ta == (config.TraitAlignment{}) {
		return "_No trait alignment rules configured._"
	}
	var b strings.Builder
	b.WriteString("| The argument | Effect on motivation |\n")
	b.WriteString("| --- | --- |\n")
	fmt.Fprintf(&b, "| **Goes against** one of the NPC's traits | %s |\n", clampPhrase(ta.Against))
	fmt.Fprintf(&b, "| **Goes with** one of the NPC's traits | %s |\n", clampPhrase(ta.With))
	fmt.Fprintf(&b, "| Touches none of their traits | Resolves normally: it rises on a success and falls on a failure. |\n")
	return strings.TrimSpace(b.String())
}

// clampPhrase turns a clamp id into a reader-facing rule.
func clampPhrase(clamp string) string {
	switch clamp {
	case "cannot_increase":
		return "Cannot rise. A success holds motivation where it is instead of gaining ground; a failure still loses it."
	case "cannot_decrease":
		return "Cannot fall. A failure holds motivation where it is instead of losing ground; a success still gains it."
	case "":
		return "-"
	default:
		return clamp
	}
}

// invokeCostWords renders an invoke/energy cost pair in plain language.
func invokeCostWords(invoke, energy int) string {
	var parts []string
	if invoke != 0 {
		parts = append(parts, fmt.Sprintf("%d invoke point%s", invoke, plural(invoke)))
	}
	if energy != 0 {
		parts = append(parts, fmt.Sprintf("%d energy", energy))
	}
	if len(parts) == 0 {
		return "Free"
	}
	return strings.Join(parts, " + ")
}

func plural(n int) string {
	if n == 1 || n == -1 {
		return ""
	}
	return "s"
}

// ordinal renders a small number as an ordinal word for the round cadence.
func ordinal(n int) string {
	switch n {
	case 1:
		return "1st"
	case 2:
		return "2nd"
	case 3:
		return "3rd"
	default:
		return fmt.Sprintf("%dth", n)
	}
}
