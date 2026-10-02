package docs

import (
	"strings"
	"testing"
)

// TestMarkdownPageEmitsHeadingIDs is the regression test for the bug this file's
// converter exists to fix: goldmark was running without AutoHeadingID, so no
// heading carried an id and every in-document link in the rulebook silently did
// nothing.
func TestMarkdownPageEmitsHeadingIDs(t *testing.T) {
	html, err := MarkdownPage([]byte("## Preparing an Action\n\nBody.\n"))
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if !strings.Contains(html, `id="preparing-an-action"`) {
		t.Fatalf("heading has no id, so anchors linking to it are dead:\n%s", html)
	}
}

// TestOutlineHTMLSkipsNonNavigationLevels checks the outline reports the three
// heading levels the documentation uses (#, ## and ###) and nothing else. Level
// 4 and below are too fine-grained and would make the sidebar unusable.
func TestOutlineHTMLSkipsNonNavigationLevels(t *testing.T) {
	html, err := MarkdownPage([]byte("# Part\n\n## Chapter\n\n### Section\n\n#### Detail\n"))
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	got := OutlineHTML(html)
	if len(got) != 3 {
		t.Fatalf("expected 3 navigable headings, got %d: %+v", len(got), got)
	}
	if got[0].Level != 1 || got[0].Text != "Part" {
		t.Errorf("first heading wrong: %+v", got[0])
	}
	if got[1].Level != 2 || got[1].Text != "Chapter" {
		t.Errorf("second heading wrong: %+v", got[1])
	}
	if got[2].Level != 3 || got[2].Text != "Section" {
		t.Errorf("third heading wrong: %+v", got[2])
	}
}

// TestOutlineStripsInlineMarkup guards the sidebar text: a heading containing
// emphasis or code must still yield plain text, not HTML tags.
func TestOutlineStripsInlineMarkup(t *testing.T) {
	html, err := MarkdownPage([]byte("## The *Energy* `cost`\n\nBody.\n"))
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	got := OutlineHTML(html)
	if len(got) != 1 {
		t.Fatalf("expected 1 heading, got %d", len(got))
	}
	if strings.Contains(got[0].Text, "<") {
		t.Errorf("heading text carries markup: %q", got[0].Text)
	}
	if got[0].Text != "The Energy cost" {
		t.Errorf("heading text = %q, want %q", got[0].Text, "The Energy cost")
	}
}

// TestTreeNestsThreeLevels covers the shape the sidebar iterates: parts hold
// chapters, chapters hold sections, so the three markdown heading levels can be
// rendered at three visually distinct depths.
func TestTreeNestsThreeLevels(t *testing.T) {
	in := []Heading{
		{Level: 1, Text: "Rules", ID: "rules"},
		{Level: 2, Text: "Combat", ID: "combat"},
		{Level: 3, Text: "Movement", ID: "movement"},
		{Level: 3, Text: "Energy", ID: "energy"},
		{Level: 2, Text: "Perks", ID: "perks"},
		{Level: 1, Text: "Appendix", ID: "appendix"},
	}
	got := Tree(in)
	if len(got) != 2 {
		t.Fatalf("expected 2 parts, got %d: %+v", len(got), got)
	}
	part := got[0]
	if part.Text != "Rules" || len(part.Sections) != 2 {
		t.Fatalf("part did not collect its chapters: %+v", part)
	}
	if part.Sections[0].Text != "Combat" || len(part.Sections[0].Sections) != 2 {
		t.Errorf("combat chapter did not collect its sections: %+v", part.Sections[0])
	}
	if part.Sections[1].Text != "Perks" || len(part.Sections[1].Sections) != 0 {
		t.Errorf("perks chapter wrong: %+v", part.Sections[1])
	}
	if got[1].Text != "Appendix" || len(got[1].Sections) != 0 {
		t.Errorf("last part wrong: %+v", got[1])
	}
}

// TestTreePromotesOrphans guards the promotion rule: a heading appearing before
// any shallower heading must still be reachable from the sidebar rather than
// being dropped because of how a document happens to start.
func TestTreePromotesOrphans(t *testing.T) {
	in := []Heading{
		{Level: 3, Text: "Orphan", ID: "orphan"},
		{Level: 2, Text: "Chapter", ID: "chapter"},
		{Level: 3, Text: "Section", ID: "section"},
	}
	got := Tree(in)
	if len(got) != 1 {
		t.Fatalf("expected 1 top-level entry, got %d: %+v", len(got), got)
	}
	if got[0].Text != "Orphan" {
		t.Fatalf("orphan section was not promoted: %+v", got[0])
	}
	if len(got[0].Sections) != 1 || got[0].Sections[0].Text != "Chapter" {
		t.Fatalf("chapter did not attach to the promoted entry: %+v", got[0])
	}
	if len(got[0].Sections[0].Sections) != 1 || got[0].Sections[0].Sections[0].Text != "Section" {
		t.Errorf("section did not attach to its chapter: %+v", got[0].Sections[0])
	}
}
