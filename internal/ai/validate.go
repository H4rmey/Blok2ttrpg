// This file validates a generated character against the loaded ruleset. It
// checks the two build pools the engine owns (skill points and perk points) and
// the referential integrity the cost engine assumes, so an AI reply that names
// an unknown id is reported instead of being silently priced as zero.
package ai

import (
	"fmt"
	"sort"
	"strings"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/engine"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// Pool is one build budget and how much of it the character spent.
type Pool struct {
	Name   string `json:"name"`
	Used   int    `json:"used"`
	Budget int    `json:"budget"`
}

// Report is the outcome of validating one generated character.
type Report struct {
	Valid  bool     `json:"valid"`
	Level  int      `json:"level"`
	Pools  []Pool   `json:"pools"`
	Issues []string `json:"issues"`
}

// Validate checks a generated character. The model chooses the level, so the
// character's own level (clamped to the ruleset cap) drives the budgets.
func Validate(cfg *config.Config, c model.Character) Report {
	c.Level = cfg.ClampLevel(c.Level)
	var issues []string

	// Referential integrity. The cost engine silently prices an unknown perk
	// type or enactment as zero, so these are checked up front; otherwise an
	// invalid reply would look like underspending rather than an error.
	issues = append(issues, validateSkills(cfg, c)...)
	issues = append(issues, validatePerks(cfg, c)...)

	// Normalize before costing so the price matches what would be stored.
	for i := range c.Perks {
		c.Perks[i] = engine.NormalizePerk(cfg, c.Perks[i])
	}

	skillUsed := engine.SkillPointsUsed(cfg, c)
	perkUsed := 0
	for _, p := range c.Perks {
		perkUsed += engine.PerkCost(cfg, p).Build
	}

	pools := []Pool{
		{Name: "skill", Used: skillUsed, Budget: cfg.SkillPointBudget(c.Level)},
		{Name: "perk", Used: perkUsed, Budget: cfg.PerkPointBudget(c.Level)},
	}
	for _, p := range pools {
		if issue := poolIssue(p); issue != "" {
			issues = append(issues, issue)
		}
	}

	return Report{Valid: len(issues) == 0, Level: c.Level, Pools: pools, Issues: issues}
}

// poolIssue describes a spending violation in one pool, or "" when the pool is
// within its rules. Both pools must be spent exactly: the 1-point dials in the
// reference always make that reachable, so any leftover or overage is an error.
func poolIssue(p Pool) string {
	delta := p.Budget - p.Used
	switch {
	case delta > 0:
		return fmt.Sprintf("%d unspent %s points (%d used of %d).", delta, p.Name, p.Used, p.Budget)
	case delta < 0:
		return fmt.Sprintf("%d over budget on %s points (%d used of %d).", -delta, p.Name, p.Used, p.Budget)
	}
	return ""
}

// ValidatePerk checks one generated perk for the referential integrity the
// cost engine assumes and the non-passive-has-enactments rule. It returns the
// issues found, empty when the perk is valid.
func ValidatePerk(cfg *config.Config, p model.Perk) []string {
	return validatePerks(cfg, model.Character{Perks: []model.Perk{p}})
}

// validateSkills reports unknown skill keys or proficiency ids.
func validateSkills(cfg *config.Config, c model.Character) []string {
	known := map[string]bool{}
	for _, g := range cfg.Skills.List() {
		for _, s := range g.Skills {
			known[model.SkillKey(g.ID, s)] = true
		}
	}
	var issues []string
	keys := make([]string, 0, len(c.Skills))
	for k := range c.Skills {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if !known[k] {
			issues = append(issues, fmt.Sprintf("unknown skill key %q.", k))
			continue
		}
		if id := c.Skills[k]; id != "" {
			if _, ok := cfg.Proficiency(id); !ok {
				issues = append(issues, fmt.Sprintf("skill %q has unknown proficiency %q.", k, id))
			}
		}
	}
	return issues
}

// validatePerks reports unknown perk types, enactment types and interaction
// ids. A passive perk (one carrying a resolvable passive_id) has none of those,
// so it is skipped.
func validatePerks(cfg *config.Config, c model.Character) []string {
	var issues []string
	for i, p := range c.Perks {
		name := p.Name
		if name == "" {
			name = fmt.Sprintf("#%d", i+1)
		}
		if pid := fieldString(p.Fields, "passive_id"); pid != "" {
			if _, ok := cfg.PassiveByID(pid); !ok {
				issues = append(issues, fmt.Sprintf("perk %q names unknown passive %q.", name, pid))
			}
			continue
		}
		if strings.EqualFold(p.Type, "passive") {
			issues = append(issues, fmt.Sprintf("passive perk %q is missing fields.passive_id; set it to one of the passive ids in the reference.", name))
			continue
		}
		if p.Type == "" {
			issues = append(issues, fmt.Sprintf("perk %q has no type.", name))
			continue
		}
		if _, ok := cfg.PerkType(p.Type); !ok {
			issues = append(issues, fmt.Sprintf("perk %q has unknown type %q.", name, p.Type))
		}
		if len(p.Enactments) == 0 {
			issues = append(issues, fmt.Sprintf("perk %q has no enactments; every non-passive perk needs at least one enactment, or use the passive shape (type: passive with fields.passive_id).", name))
		}
		for _, en := range p.Enactments {
			if en.Type == "" {
				continue
			}
			ecomp, eok := cfg.Enactment(en.Type)
			if !eok {
				issues = append(issues, fmt.Sprintf("perk %q has unknown enactment %q.", name, en.Type))
				continue
			}
			if en.Interaction != "" {
				if _, ok := cfg.Interaction(en.Interaction); !ok {
					issues = append(issues, fmt.Sprintf("perk %q has unknown interaction %q.", name, en.Interaction))
				}
			}
			issues = append(issues, validateResourceEnactment(cfg, name, en, ecomp)...)
		}
	}
	issues = append(issues, validateResourcePairs(c)...)
	return issues
}

// validateResourceEnactment checks one resource enactment: a generate entry
// only generates, a consume entry only consumes and names a real effect, and
// neither writes the app-built nested key by hand.
func validateResourceEnactment(cfg *config.Config, perkName string, en model.Enactment, ecomp config.Component) []string {
	if en.Type != "resource" {
		return nil
	}
	var issues []string
	mode := resourceMode(en.Fields)
	switch mode {
	case "generate":
		if effect := fieldString(en.Fields, "consumed_effect"); effect != "" && !isDefaultConsumedEffect(cfg, ecomp, effect) {
			issues = append(issues, fmt.Sprintf("perk %q generates resources but also sets consumed_effect; generate and consume must be separate perks sharing one resource_name.", perkName))
		}
		if name := strings.TrimSpace(fieldString(en.Fields, "resource_name")); name == "" {
			issues = append(issues, fmt.Sprintf("perk %q generates resources without a resource_name; name the resource the spender will consume.", perkName))
		}
	case "consume":
		if name := strings.TrimSpace(fieldString(en.Fields, "resource_name")); name == "" {
			issues = append(issues, fmt.Sprintf("perk %q consumes resources without a resource_name; use the generator perk's resource_name.", perkName))
		}
		if effect := fieldString(en.Fields, "consumed_effect"); effect == "" || effect == "generate" {
			issues = append(issues, fmt.Sprintf("perk %q consumes resources without a real consumed_effect (damage, healing, motion, condition, shift, adjustment, phase); the nested fields are configured in the app afterwards.", perkName))
		}
	}
	if _, nested := en.Fields["consumed_effect_ib"]; nested {
		issues = append(issues, fmt.Sprintf("perk %q writes consumed_effect_ib by hand; omit it (the app builds it) and set only resource_mode, resource_name and consumed_effect.", perkName))
	}
	return issues
}

// validateResourcePairs checks that every consumed resource_name has a
// generator perk and every generator's resource is consumed somewhere.
func validateResourcePairs(c model.Character) []string {
	generated, consumed := map[string]string{}, map[string]string{}
	for _, p := range c.Perks {
		name := p.Name
		if name == "" {
			name = "an unnamed perk"
		}
		for _, en := range p.Enactments {
			if en.Type != "resource" {
				continue
			}
			res := strings.ToLower(strings.TrimSpace(fieldString(en.Fields, "resource_name")))
			if res == "" {
				continue
			}
			switch resourceMode(en.Fields) {
			case "generate":
				generated[res] = name
			case "consume":
				consumed[res] = name
			}
		}
	}
	var issues []string
	for res, perk := range consumed {
		if _, ok := generated[res]; !ok {
			issues = append(issues, fmt.Sprintf("perk %q consumes resource %q but no perk generates it; add a separate generator perk with resource_mode: generate and the same resource_name.", perk, res))
		}
	}
	for res, perk := range generated {
		if _, ok := consumed[res]; !ok {
			issues = append(issues, fmt.Sprintf("perk %q generates resource %q but no perk consumes it; add a separate spender perk with resource_mode: consume and the same resource_name.", perk, res))
		}
	}
	return issues
}

// resourceMode reads a resource enactment's mode, defaulting to the config
// default (generate) when unset.
func resourceMode(fields map[string]any) string {
	if mode := strings.ToLower(strings.TrimSpace(fieldString(fields, "resource_mode"))); mode != "" {
		return mode
	}
	return "generate"
}

// isDefaultConsumedEffect reports whether the value equals the resource
// component's own default for consumed_effect, which is present even on a pure
// generator that did not mean to consume anything.
func isDefaultConsumedEffect(cfg *config.Config, ecomp config.Component, value string) bool {
	for _, f := range ecomp.Fields {
		if f.Key != "consumed_effect" {
			continue
		}
		return strings.EqualFold(fmt.Sprintf("%v", f.Default), strings.TrimSpace(value))
	}
	return false
}

// Correction builds the user-role retry message appended after an invalid
// reply. It states the exact pool errors and asks for the full corrected YAML.
func Correction(r Report) string {
	var b strings.Builder
	b.WriteString("Your character failed validation. ")
	b.WriteString(strings.Join(r.Issues, " "))
	b.WriteString(" Spend ALL skill and perk points exactly; to adjust the perk total, tune only the 1-point dials from the reference, without adding or removing perks.")
	if hasResourceSplitIssue(r) {
		b.WriteString(" A resource generator and its spender must be two separate perks sharing one resource_name; never combine generate and consume in one perk.")
	}
	b.WriteString(" Use lowercase proficiency ids (novice, proficient, expert, ...), not tier names, and set fields.passive_id for any passive perk.")
	b.WriteString(" Return the full corrected character YAML only, with no markdown fences or commentary.")
	return b.String()
}

// hasResourceSplitIssue reports whether any issue touches resource usage, so
// the retry can restate the generator/spender split exactly when it is broken.
func hasResourceSplitIssue(r Report) bool {
	for _, issue := range r.Issues {
		if strings.Contains(issue, "resource") {
			return true
		}
	}
	return false
}

// fieldString reads a string field from a loosely-typed values map.
func fieldString(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	if s, ok := values[key].(string); ok {
		return s
	}
	return ""
}
