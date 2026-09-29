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
	if !strings.Contains(body, `class="content docs changelog"`) {
		t.Errorf("changelog page is missing its styling hook:\n%s", body)
	}
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
