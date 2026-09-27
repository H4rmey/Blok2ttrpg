// This file is the guard that keeps the documentation from falling behind the
// code. It renders the real ruleset against the real docs tree and fails when
// anything went missing.
//
// The problem it exists to catch: text/template renders a range over an empty
// or absent slice as nothing at all, so when the leveling config moved from a
// hand-written levels list to a computed formula, the ability point table kept
// its header and lost every row, and nothing noticed. These tests turn that
// class of silent loss into a test failure.
package docs

import (
	"os"
	"strings"
	"testing"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/premade"
)

// These paths are relative to the project root, which is where file_order
// entries are authored from and where the application runs.
const (
	configDir  = "config/Blok2Simplified"
	libraryDir = "library"
	docsDir    = "docs"
)

// TestMain runs the package's tests from the project root. The docs pipeline
// resolves file_order entries, the content library and the docs tree relative to
// the working directory the application is started in, so the tests must use the
// same vantage point rather than the package directory Go defaults to.
func TestMain(m *testing.M) {
	if err := os.Chdir("../.."); err != nil {
		panic("changing to project root: " + err.Error())
	}
	os.Exit(m.Run())
}

// loadRuleset loads the shipped ruleset, failing the test when it cannot.
func loadRuleset(t *testing.T) *config.Loaded {
	t.Helper()
	loaded, err := config.Load(configDir)
	if err != nil {
		t.Fatalf("loading config from %s: %v", configDir, err)
	}
	return loaded
}

// TestRenderedDocsAreComplete renders the whole rulebook and asserts the lint
// finds nothing. A failure here means a chapter lost data: most likely a table
// whose rows came from config that has since changed shape.
func TestRenderedDocsAreComplete(t *testing.T) {
	loaded := loadRuleset(t)
	md, err := RenderMarkdown(loaded, premade.New(libraryDir))
	if err != nil {
		t.Fatalf("rendering docs: %v", err)
	}
	if findings := Lint(md); len(findings) > 0 {
		t.Errorf("rendered documentation has %d problem(s):\n%s", len(findings), FindingsText(findings))
	}
}

// TestSchemaCoverage asserts that every yaml key of every documented config type
// has a description. This is what stops a new config key from being added
// without also being documented: the key appears via reflection immediately, so
// the only way to make this pass is to describe it.
func TestSchemaCoverage(t *testing.T) {
	if gaps := LintSchemaCoverage(); len(gaps) > 0 {
		t.Errorf("configuration reference is missing %d description(s):\n  %s",
			len(gaps), strings.Join(gaps, "\n  "))
	}
}

// TestDocsOrderIsComplete asserts that file_order and the docs tree agree: no
// chapter is written but left out of the rulebook, and no listed file is
// missing from disk.
func TestDocsOrderIsComplete(t *testing.T) {
	loaded := loadRuleset(t)
	if err := loaded.Config.ValidateDocsOrder(docsDir); err != nil {
		t.Errorf("file_order does not match the docs tree: %v", err)
	}
}

// TestLevelingTableHasEveryLevel is a direct regression test for the bug that
// started this work: the ability point table rendering with no rows. It asserts
// a row exists for every level up to the configured cap, for both point pools.
func TestLevelingTableHasEveryLevel(t *testing.T) {
	loaded := loadRuleset(t)
	cfg := loaded.Config
	for _, pool := range []string{"skill", "ability"} {
		table := levelingTable(cfg, pool)
		rows := countTableRows(table)
		if want := cfg.MaxLevel(); rows != want {
			t.Errorf("%s leveling table has %d rows, want %d (one per level):\n%s",
				pool, rows, want, table)
		}
	}
}

// TestLevelingTableMatchesBudgets asserts the documented totals are the numbers
// the application actually serves, by checking the rendered table contains the
// budget the config computes for a sample of levels.
func TestLevelingTableMatchesBudgets(t *testing.T) {
	loaded := loadRuleset(t)
	cfg := loaded.Config
	table := levelingTable(cfg, "ability")
	for level := 1; level <= cfg.MaxLevel(); level++ {
		want := cfg.AbilityPointBudget(level)
		if !strings.Contains(table, "| **"+itoa(level)+"** |") {
			t.Errorf("ability leveling table has no row for level %d", level)
			continue
		}
		if !rowHasTotal(table, level, want) {
			t.Errorf("ability leveling table row for level %d does not show the served budget %d:\n%s",
				level, want, table)
		}
	}
}

// TestPackagesTableListsLibrary asserts the content library reaches the docs, so
// the classes, races, backgrounds and items cannot silently vanish again.
func TestPackagesTableListsLibrary(t *testing.T) {
	lib := premade.New(libraryDir)
	pkgs, err := lib.ListPackages()
	if err != nil {
		t.Fatalf("listing packages: %v", err)
	}
	if len(pkgs) == 0 {
		t.Skip("no packages in the library to document")
	}
	table := packagesTable(lib)
	for _, p := range pkgs {
		if !strings.Contains(table, p.Name) {
			t.Errorf("package %q (%s) is missing from the generated package tables", p.Name, p.Category)
		}
	}
}

// countTableRows counts the body rows of a markdown table: pipe-delimited lines
// that are neither the header nor the separator.
func countTableRows(table string) int {
	lines := strings.Split(table, "\n")
	count := 0
	seenSeparator := false
	for _, l := range lines {
		if isTableSeparator(l) {
			seenSeparator = true
			continue
		}
		if seenSeparator && isTableRow(l) {
			count++
		}
	}
	return count
}

// rowHasTotal reports whether the table's row for the given level ends with the
// expected total in its last cell.
func rowHasTotal(table string, level, total int) bool {
	prefix := "| **" + itoa(level) + "** |"
	for _, l := range strings.Split(table, "\n") {
		l = strings.TrimSpace(l)
		if !strings.HasPrefix(l, prefix) {
			continue
		}
		cells := strings.Split(strings.Trim(l, "|"), "|")
		if len(cells) == 0 {
			return false
		}
		return strings.TrimSpace(cells[len(cells)-1]) == itoa(total)
	}
	return false
}

// itoa is a tiny local integer formatter, kept to avoid pulling strconv into a
// file that otherwise only does string comparison.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		return "-" + string(digits)
	}
	return string(digits)
}
