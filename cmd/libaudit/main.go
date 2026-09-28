// Command libaudit reports the completeness and duplication of the content
// library against the loaded ruleset. It is the repeatable form of a manual
// content review: instead of reading every perk file to notice that only one
// reaction trigger is ever used, run this.
//
// Everything it measures comes from the config and the library files - there
// are no hardcoded type names, trigger ids, tag names or thresholds. Adding a
// perk type or a reaction trigger to the config therefore widens the coverage
// report automatically, and the "unused" sections are the honest answer to
// "what does the ruleset allow that no content demonstrates?".
//
// It always exits 0: everything here is an authoring signal, not a build
// failure. Nothing it reports is unambiguously wrong - a shared enactment shape
// may be healthy flavour variety, an unimported perk may be intentionally
// library-only - so the tool describes the library and leaves the judgement to
// the author.
package main

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/engine"
	"github.com/harmey/blok2ttrpg-v5/internal/model"
	"github.com/harmey/blok2ttrpg-v5/internal/premade"
)

const (
	configDir  = "config/Blok2Simplified"
	libraryDir = "library"

	// rareTagThreshold is the count at or below which a tag is listed
	// separately rather than inline. Single-use tags are deliberately NOT
	// presented as errors: damage types and other genuinely specific labels
	// legitimately appear once, and flagging each as a suspected typo produced
	// enough noise to bury the one case that would matter. They are grouped at
	// the end instead, so a misspelling sits beside its correct twin where the
	// eye can catch it.
	rareTagThreshold = 1
)

func main() {
	cfg, err := config.Load(configDir)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	lib := premade.New(libraryDir)

	perks, err := lib.ListPerks()
	if err != nil {
		log.Fatalf("list perks: %v", err)
	}
	pkgs, err := lib.ListPackages()
	if err != nil {
		log.Fatalf("list packages: %v", err)
	}

	fmt.Printf("Library audit: %d perks, %d packages, %d passive entries\n",
		len(perks), len(pkgs), len(cfg.Passives.Entries))

	reportPerkTypes(cfg, perks)
	reportTriggers(cfg, perks)
	reportEnactments(cfg, perks)
	reportTags(perks)
	reportDuplicates(perks)
	reportOrphans(perks, pkgs)
	reportCostOutliers(cfg, perks)
}

// reportPerkTypes counts library perks per configured perk type. A type with a
// zero count is a part of the ruleset no content demonstrates, which is the
// single most useful number here.
func reportPerkTypes(cfg *config.Loaded, perks []model.Perk) {
	section("Perk type coverage")
	counts := map[string]int{}
	for _, p := range perks {
		counts[p.Type]++
	}
	for _, id := range cfg.PerkTypes.Order {
		n := counts[id]
		note := ""
		// A perk type that allows no enactments cannot be expressed as a library
		// file at all: its content is a catalogue in the config instead (the
		// passive list). A zero count for such a type is structurally correct
		// rather than a content gap, and this is detected from the config's own
		// allowed_enactments rather than by naming the type here.
		if n == 0 {
			if t, ok := cfg.PerkType(id); ok && t.AllowedEnactments != nil && len(t.AllowedEnactments) == 0 {
				note = "  (config catalogue, not library files - expected)"
			}
		}
		fmt.Printf("  %-16s %3d %s%s\n", id, n, bar(n, len(perks)), note)
	}
	// A stored type the config no longer defines. Not fatal, but the perk will
	// price as if it had no type at all.
	for id, n := range counts {
		if _, ok := cfg.PerkType(id); !ok {
			fmt.Printf("  %-16s %3d  UNKNOWN TYPE (not in config)\n", id, n)
		}
	}
}

// reportTriggers shows which configured reaction triggers the library actually
// uses. The trigger list is the richest part of the reaction config, so an
// unused trigger is content debt rather than a config problem.
func reportTriggers(cfg *config.Loaded, perks []model.Perk) {
	section("Reaction trigger coverage")
	opts := cfg.ResolveOptions(config.Field{OptionsSource: "reaction_triggers"})
	if len(opts) == 0 {
		fmt.Println("  (no reaction_triggers option source in config)")
		return
	}
	used := map[string][]string{}
	for _, p := range perks {
		t, _ := p.Fields["trigger"].(string)
		if t != "" {
			used[t] = append(used[t], p.Name)
		}
	}
	for _, o := range opts {
		names := used[o.Value]
		if len(names) == 0 {
			fmt.Printf("  %-20s   -  (unused)\n", o.Value)
			continue
		}
		sort.Strings(names)
		fmt.Printf("  %-20s %3d  %s\n", o.Value, len(names), strings.Join(names, ", "))
	}
	// A trigger stored on a perk that the config does not offer: the builder
	// could not reproduce this perk, so it is worth knowing about.
	for t, names := range used {
		if !hasOption(opts, t) {
			sort.Strings(names)
			fmt.Printf("  %-20s %3d  UNKNOWN TRIGGER: %s\n", t, len(names), strings.Join(names, ", "))
		}
	}
}

// reportEnactments counts how often each configured enactment type appears, and
// separately how often it appears as a perk's first (primary) enactment. A type
// that only ever shows up in a trailing slot is never demonstrated as the point
// of a perk.
func reportEnactments(cfg *config.Loaded, perks []model.Perk) {
	section("Enactment coverage (total / as first enactment)")
	total := map[string]int{}
	first := map[string]int{}
	for _, p := range perks {
		for i, e := range p.Enactments {
			total[e.Type]++
			if i == 0 {
				first[e.Type]++
			}
		}
	}
	for _, id := range cfg.Enactments.Order {
		note := ""
		if total[id] == 0 {
			note = " (unused)"
		} else if first[id] == 0 {
			note = " (never primary)"
		}
		fmt.Printf("  %-16s %3d / %3d%s\n", id, total[id], first[id], note)
	}
}

// reportTags counts the free-form tags in use. This is the safety net for the
// deliberate lack of a tag vocabulary: a misspelling cannot be caught by
// validation, but it shows up here as a tag used once.
func reportTags(perks []model.Perk) {
	section("Tag usage (free-form, no config vocabulary)")
	counts := map[string]int{}
	untagged := 0
	for _, p := range perks {
		if len(p.Tags) == 0 {
			untagged++
			continue
		}
		for _, t := range p.Tags {
			counts[strings.TrimSpace(t)]++
		}
	}
	tags := make([]string, 0, len(counts))
	for t := range counts {
		tags = append(tags, t)
	}
	// Most-used first: the shape of the library is easier to read that way, and
	// the suspicious rare tags collect at the bottom.
	sort.Slice(tags, func(i, j int) bool {
		if counts[tags[i]] != counts[tags[j]] {
			return counts[tags[i]] > counts[tags[j]]
		}
		return tags[i] < tags[j]
	})
	var rare []string
	for _, t := range tags {
		if counts[t] <= rareTagThreshold {
			rare = append(rare, t)
			continue
		}
		fmt.Printf("  %-16s %3d\n", t, counts[t])
	}
	fmt.Printf("  %-16s %3d\n", "(untagged)", untagged)
	if len(rare) > 0 {
		sort.Strings(rare)
		fmt.Printf("\n  Used %d time(s). Expected for specific labels such as damage\n", rareTagThreshold)
		fmt.Printf("  types; scan for near-duplicates of the tags above:\n    %s\n",
			strings.Join(rare, ", "))
	}
}

// reportDuplicates groups perks by a structural signature: their type plus the
// ordered list of enactment type/interaction pairs. Perks sharing a signature
// are built the same way and differ only in their field values, which is where
// re-skins hide.
//
// A shared signature is not automatically a fault: a damage+condition melee
// attack is a legitimate shape that many perks should have. It only means these
// are the perks to compare when looking for redundancy.
func reportDuplicates(perks []model.Perk) {
	section("Shared structural signatures (candidates for duplication)")
	groups := map[string][]string{}
	for _, p := range perks {
		parts := []string{p.Type}
		for _, e := range p.Enactments {
			parts = append(parts, e.Type+"/"+e.Interaction)
		}
		sig := strings.Join(parts, " + ")
		groups[sig] = append(groups[sig], p.Name)
	}
	sigs := make([]string, 0, len(groups))
	for s, names := range groups {
		if len(names) > 1 {
			sigs = append(sigs, s)
		}
	}
	if len(sigs) == 0 {
		fmt.Println("  none: every perk has a distinct structure.")
		return
	}
	sort.Slice(sigs, func(i, j int) bool {
		if len(groups[sigs[i]]) != len(groups[sigs[j]]) {
			return len(groups[sigs[i]]) > len(groups[sigs[j]])
		}
		return sigs[i] < sigs[j]
	})
	for _, s := range sigs {
		names := groups[s]
		sort.Strings(names)
		fmt.Printf("  %d x  %s\n", len(names), s)
		fmt.Printf("        %s\n", strings.Join(names, ", "))
	}
}

// reportOrphans lists perks no package imports. These are browsable-only
// content: reachable from the perk library but never handed to a character by
// picking a class, race, background or item.
func reportOrphans(perks []model.Perk, pkgs []premade.Package) {
	section("Perks imported by no package")
	imported := map[string]bool{}
	for _, pkg := range pkgs {
		for _, p := range pkg.Perks {
			imported[p.Name] = true
		}
	}
	var orphans []string
	for _, p := range perks {
		if !imported[p.Name] {
			orphans = append(orphans, p.Name)
		}
	}
	sort.Strings(orphans)
	fmt.Printf("  %d of %d perks are library-only\n", len(orphans), len(perks))
	for _, n := range orphans {
		fmt.Printf("    %s\n", n)
	}
}

// reportCostOutliers lists perks costing more than a level 1 character's whole
// perk budget. Such a perk is fine to exist but cannot appear in a starting
// package, which is exactly the trap the elf package hit with Gale Step.
func reportCostOutliers(cfg *config.Loaded, perks []model.Perk) {
	budget := cfg.PerkPointBudget(1)
	section(fmt.Sprintf("Perks costing more than the level 1 budget (%d pt)", budget))
	type row struct {
		name string
		cost int
	}
	var rows []row
	for _, p := range perks {
		c := engine.PerkCost(cfg.Config, engine.NormalizePerk(cfg.Config, p))
		if c.Build > budget {
			rows = append(rows, row{p.Name, c.Build})
		}
	}
	if len(rows) == 0 {
		fmt.Println("  none: every perk fits a starting character's budget.")
		return
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].cost > rows[j].cost })
	for _, r := range rows {
		fmt.Printf("  %-24s %3d pt\n", r.name, r.cost)
	}
}

func section(title string) {
	fmt.Printf("\n%s\n%s\n", title, strings.Repeat("-", len(title)))
}

// bar renders a proportional bar so a lopsided distribution is visible without
// comparing numbers. Width is fixed; an empty count renders as nothing.
func bar(n, total int) string {
	if total == 0 || n == 0 {
		return ""
	}
	const width = 30
	w := n * width / total
	if w == 0 {
		w = 1
	}
	return strings.Repeat("#", w)
}

func hasOption(opts []config.Option, value string) bool {
	for _, o := range opts {
		if o.Value == value {
			return true
		}
	}
	return false
}
