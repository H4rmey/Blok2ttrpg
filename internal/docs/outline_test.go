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

// TestOutlineHTMLSkipsNonNavigationLevels checks the outline reports chapters and
// sections and nothing else. Level 1 is the document title and level 4 is too
// fine-grained, so including either would make the sidebar unusable.
func TestOutlineHTMLSkipsNonNavigationLevels(t *testing.T) {
	html, err := MarkdownPage([]byte("# Title\n\n## Chapter\n\n### Section\n\n#### Detail\n"))
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	got := OutlineHTML(html)
	if len(got) != 2 {
		t.Fatalf("expected 2 navigable headings, got %d: %+v", len(got), got)
	}
	if got[0].Level != 2 || got[0].Text != "Chapter" {
		t.Errorf("first heading wrong: %+v", got[0])
	}
	if got[1].Level != 3 || got[1].Text != "Section" {
		t.Errorf("second heading wrong: %+v", got[1])
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

// TestTreeNestsSectionsUnderChapters covers the shape the sidebar iterates,
// including the promotion rule: a section appearing before any chapter becomes a
// chapter of its own rather than being dropped, so it stays reachable.
func TestTreeNestsSectionsUnderChapters(t *testing.T) {
	in := []Heading{
		{Level: 3, Text: "Orphan", ID: "orphan"},
		{Level: 2, Text: "Combat", ID: "combat"},
		{Level: 3, Text: "Movement", ID: "movement"},
		{Level: 3, Text: "Energy", ID: "energy"},
		{Level: 2, Text: "Perks", ID: "perks"},
	}
	got := Tree(in)
	if len(got) != 3 {
		t.Fatalf("expected 3 top-level entries, got %d: %+v", len(got), got)
	}
	if got[0].Text != "Orphan" || len(got[0].Sections) != 0 {
		t.Errorf("orphan section was not promoted: %+v", got[0])
	}
	if got[1].Text != "Combat" || len(got[1].Sections) != 2 {
		t.Errorf("combat chapter did not collect its sections: %+v", got[1])
	}
	if got[2].Text != "Perks" || len(got[2].Sections) != 0 {
		t.Errorf("last chapter wrong: %+v", got[2])
	}
}
