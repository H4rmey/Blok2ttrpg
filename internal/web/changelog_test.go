package web

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// TestChangelogRenders checks that the changelog page serves the repository's
// CHANGELOG.md as HTML. The handler resolves the file relative to the working
// directory, so the test changes into the repository root the way the running
// server does.
func TestChangelogRenders(t *testing.T) {
	app, _ := testAppWithPerk(t)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir("../.."); err != nil {
		t.Fatalf("chdir to repo root: %v", err)
	}
	t.Cleanup(func() { os.Chdir(wd) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/changelog", nil)
	app.handleChangelog(rec, req)

	if rec.Code != 200 {
		t.Fatalf("changelog returned %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	// The markdown must actually have been converted, not echoed: a heading
	// proves goldmark ran over the real file.
	if !strings.Contains(body, "<h1") {
		t.Errorf("changelog markdown was not rendered to HTML:\n%s", body)
	}
	if strings.Contains(body, "No changelog is available") {
		t.Errorf("CHANGELOG.md was not found from the repository root")
	}
	// The section styling hangs off this class; without it the page renders as
	// one undifferentiated wall of text.
	if !strings.Contains(body, "changelog") {
		t.Errorf("changelog page is missing its styling hook:\n%s", body)
	}
	assertDocReader(t, body, "changelog")
}

// assertDocReader checks the wiki-style reading affordances a long generated
// document depends on: the contents tree, the scoped search box, and heading ids
// for the tree to link to. A document this long is unusable without them, so
// they are asserted rather than left to manual inspection.
func assertDocReader(t *testing.T, body, page string) {
	t.Helper()
	if !strings.Contains(body, `id="doc-sidebar"`) {
		t.Errorf("%s page has no contents sidebar:\n%s", page, body)
	}
	if !strings.Contains(body, `id="doc-search-input"`) {
		t.Errorf("%s page has no search box", page)
	}
	if !strings.Contains(body, `class="doc-chapter-link"`) {
		t.Errorf("%s page sidebar has no chapter entries, so the outline came back empty", page)
	}
	// The sidebar links by anchor, so the body must actually carry ids.
	if !strings.Contains(body, `<h2 id="`) {
		t.Errorf("%s page headings have no ids, so every sidebar link is dead", page)
	}
	if !strings.Contains(body, "data-doc-page") {
		t.Errorf("%s page is not marked for the reader script", page)
	}
}

// TestDocsPageHasReader covers the same affordances on the rulebook, which is the
// longer of the two documents and the one that most needs them.
func TestDocsPageHasReader(t *testing.T) {
	app, _ := testAppWithPerk(t)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	// The rulebook's file_order paths resolve relative to the project root.
	if err := os.Chdir("../.."); err != nil {
		t.Fatalf("chdir to repo root: %v", err)
	}
	t.Cleanup(func() { os.Chdir(wd) })

	rec := httptest.NewRecorder()
	app.handleDocs(rec, httptest.NewRequest("GET", "/docs", nil))
	if rec.Code != 200 {
		t.Fatalf("docs returned %d, want 200", rec.Code)
	}
	assertDocReader(t, rec.Body.String(), "docs")
}

// TestChangelogNavLink guards the entry point: a changelog nobody can reach is
// no better than no changelog.
func TestChangelogNavLink(t *testing.T) {
	app, c := testAppWithPerk(t)

	rec := httptest.NewRecorder()
	app.renderPerkList(rec, c)
	body := rec.Body.String()

	if !strings.Contains(body, `href="/changelog"`) {
		t.Errorf("top bar does not link to the changelog:\n%s", body)
	}
	// The mobile drawer needs both halves to work: the button and the nav it
	// controls.
	if !strings.Contains(body, `id="nav-toggle"`) || !strings.Contains(body, `id="nav-links"`) {
		t.Errorf("responsive nav markup missing from the top bar:\n%s", body)
	}
}
