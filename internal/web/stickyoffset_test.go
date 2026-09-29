package web

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// The character bar and the builder header both stick to the top of the
// viewport, and so does the site header. They were all pinned at top: 0, so the
// bar scrolled underneath the header and its stat cards were clipped.
//
// The fix offsets every other sticky element by --topbar-h. That is a layout
// invariant no Go test can observe through rendered HTML, because it lives
// entirely in the stylesheet, so it is asserted against the stylesheet itself:
// cheap to run, and it fails loudly if someone reintroduces top: 0 on a sticky
// block that sits below the header.

// stickyBelowHeader names the rules that must clear the sticky site header.
// Adding another sticky element means adding it here too.
var stickyBelowHeader = []string{
	".stats-bar",
	".builder-head",
	".doc-sidebar",
}

// readAppCSS loads the stylesheet from the repository root, the way the server
// serves it.
func readAppCSS(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir("../.."); err != nil {
		t.Fatalf("chdir to repo root: %v", err)
	}
	t.Cleanup(func() { os.Chdir(wd) })

	b, err := os.ReadFile("static/css/app.css")
	if err != nil {
		t.Fatalf("reading stylesheet: %v", err)
	}
	return string(b)
}

// ruleBody returns the declarations of the first top-level rule for selector.
func ruleBody(css, selector string) (string, bool) {
	// Anchored to the start of a line so ".stats-bar" does not match the
	// descendant rules like ".stats-bar .cost-card".
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(selector) + ` \{([^}]*)\}`)
	m := re.FindStringSubmatch(css)
	if m == nil {
		return "", false
	}
	return m[1], true
}

// TestTopbarHeightIsDefined checks the offset variable exists. Every assertion
// below depends on it, so a missing definition should report itself rather than
// surface as several confusing failures.
func TestTopbarHeightIsDefined(t *testing.T) {
	css := readAppCSS(t)
	if !strings.Contains(css, "--topbar-h:") {
		t.Fatal("--topbar-h is not defined; sticky elements have nothing to offset against")
	}
}

// TestStickyElementsClearTheHeader is the regression guard: anything that sticks
// below the site header must offset itself by --topbar-h and must not sit at
// top: 0, which would put it underneath the header.
func TestStickyElementsClearTheHeader(t *testing.T) {
	css := readAppCSS(t)

	for _, selector := range stickyBelowHeader {
		t.Run(selector, func(t *testing.T) {
			body, ok := ruleBody(css, selector)
			if !ok {
				t.Fatalf("no top-level rule found for %s", selector)
			}
			if !strings.Contains(body, "position: sticky") {
				t.Fatalf("%s is no longer sticky; remove it from stickyBelowHeader "+
					"or restore the offset", selector)
			}
			if !strings.Contains(body, "top: var(--topbar-h)") {
				t.Errorf("%s does not offset itself by --topbar-h, so it will scroll "+
					"underneath the sticky site header", selector)
			}
			if strings.Contains(body, "top: 0") {
				t.Errorf("%s is pinned at top: 0, which puts it behind the sticky "+
					"site header", selector)
			}
		})
	}
}

// TestTopbarOutranksOtherStickyElements checks the stacking order. Clearing the
// header by offset is not enough on its own: while scrolling, the bar passes
// through the header's box, and whichever element has the higher z-index is
// drawn on top. The header has to win.
func TestTopbarOutranksOtherStickyElements(t *testing.T) {
	css := readAppCSS(t)

	topbar, ok := ruleBody(css, ".topbar")
	if !ok {
		t.Fatal("no top-level rule found for .topbar")
	}
	if !strings.Contains(topbar, "z-index: var(--z-topbar)") {
		t.Error(".topbar does not use --z-topbar, so its stacking order is no longer " +
			"guaranteed to beat the other sticky elements")
	}

	for _, selector := range []string{".stats-bar", ".builder-head"} {
		body, ok := ruleBody(css, selector)
		if !ok {
			t.Fatalf("no top-level rule found for %s", selector)
		}
		if !strings.Contains(body, "z-index: var(--z-sticky)") {
			t.Errorf("%s does not use --z-sticky, so it may paint over the site header",
				selector)
		}
	}
}

// TestNarrowScreensRaiseTheOffset checks the responsive case. Below the
// breakpoint the character name moves out of the absolutely-centred position
// onto its own row, making the header taller, so a single fixed offset would
// let the bar slide under it again on a phone.
func TestNarrowScreensRaiseTheOffset(t *testing.T) {
	css := readAppCSS(t)

	i := strings.Index(css, "@media (max-width: 720px)")
	if i < 0 {
		t.Fatal("the narrow-screen breakpoint is gone; the offset override needs revisiting")
	}
	if !strings.Contains(css[i:], "--topbar-h:") {
		t.Error("the narrow-screen breakpoint does not raise --topbar-h, but it moves the " +
			"character name onto its own row, which makes the header taller")
	}
}
