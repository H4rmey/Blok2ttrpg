// These tests prove the documentation lint actually fires on the failure modes
// it claims to catch, and stays quiet on legitimate markdown. Without them the
// lint is only asserted against the current docs, which pass; a lint that
// silently never triggers would give exactly the same green result as a working
// one.
package docs

import "testing"

// TestLintCatchesEmptyTable is the regression case that motivated the lint: a
// table whose rows came from a template range that produced nothing.
func TestLintCatchesEmptyTable(t *testing.T) {
	md := "## Leveling Table\n\n| Level | Points |\n| --- | --- |\n\n## Next Section\n\nBody.\n"
	findings := Lint(md)
	if len(findings) == 0 {
		t.Fatal("expected a finding for a table with no body rows, got none")
	}
}

// TestLintAcceptsPopulatedTable guards against the lint flagging healthy tables.
func TestLintAcceptsPopulatedTable(t *testing.T) {
	md := "## Leveling Table\n\n| Level | Points |\n| --- | --- |\n| 1 | 9 |\n| 2 | 12 |\n"
	if findings := Lint(md); len(findings) > 0 {
		t.Errorf("expected no findings for a populated table, got:\n%s", FindingsText(findings))
	}
}

// TestLintCatchesTemplateLeftovers covers an action that never executed.
func TestLintCatchesTemplateLeftovers(t *testing.T) {
	md := "## Section\n\nThe budget is {{ .Leveling.Missing }} points.\n"
	if findings := Lint(md); len(findings) == 0 {
		t.Fatal("expected a finding for unexecuted template syntax, got none")
	}
}

// TestLintCatchesPlaceholder covers a helper reaching its no-data fallback.
func TestLintCatchesPlaceholder(t *testing.T) {
	md := "## Conditions\n\n_No conditions configured._\n"
	if findings := Lint(md); len(findings) == 0 {
		t.Fatal("expected a finding for an empty-data placeholder, got none")
	}
}

// TestLintCatchesDanglingHeading covers a heading whose section was never
// filled, which is what a heading followed by a shallower heading means.
func TestLintCatchesDanglingHeading(t *testing.T) {
	md := "# Chapter\n\nIntro.\n\n## Section\n\nBody.\n\n### Perk List\n\n# Next Chapter\n\nBody.\n"
	findings := Lint(md)
	if len(findings) == 0 {
		t.Fatal("expected a finding for a dangling heading, got none")
	}
	found := false
	for _, f := range findings {
		if f.Message == "heading has no content: ### Perk List" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected the dangling '### Perk List' heading to be reported, got:\n%s",
			FindingsText(findings))
	}
}

// TestLintAcceptsNestedAndSiblingHeadings confirms the two legitimate shapes are
// not reported: a title followed by a subheading, and this project's house style
// of a slug title followed by a same-level section heading.
func TestLintAcceptsNestedAndSiblingHeadings(t *testing.T) {
	cases := map[string]string{
		"nested":  "# Chapter\n\n## Section\n\nBody.\n",
		"sibling": "# items\n## Items\n\nBody.\n",
	}
	for name, md := range cases {
		if findings := Lint(md); len(findings) > 0 {
			t.Errorf("%s headings should not be reported, got:\n%s", name, FindingsText(findings))
		}
	}
}
