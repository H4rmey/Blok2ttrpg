// This file provides the shared markdown-to-HTML converter used by every page
// that renders documentation-style markdown, together with the heading outline
// those pages build their sidebar from.
//
// Two problems are solved here.
//
// First, heading anchors. The docs are written as many markdown files that link
// into each other by anchor ("see the [Invoking](#invoking) chapter"). Goldmark
// does not emit heading ids unless AutoHeadingID is enabled, and it was not, so
// every one of those links silently did nothing. Enabling it here fixes them all
// at once and is the reason the converter is centralised rather than constructed
// at each call site.
//
// Second, the outline. A sidebar has to list exactly the anchors that exist in
// the document, so the outline is extracted from the *rendered HTML* rather than
// re-derived from the markdown. That way the ids in the sidebar are by
// construction the ids goldmark actually produced - including its disambiguation
// of headings that share a title - instead of a second slug implementation that
// could drift from it.
package docs

import (
	"bytes"
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

// Heading is one entry in a document's outline.
type Heading struct {
	// Level is the markdown heading level: 2 for a chapter, 3 for a section.
	Level int
	// Text is the heading's visible text, with any inline markup removed.
	Text string
	// ID is the anchor to link to, as emitted by goldmark.
	ID string
}

// Children is the nested outline used by the sidebar: a chapter and the sections
// under it. Deeper headings are deliberately not nested further, because a
// three-level tree stops being scannable.
type Children struct {
	Heading
	Sections []Heading
}

// MarkdownPage converts arbitrary documentation-style markdown to an HTML
// fragment with heading ids. It is the entry point for pages whose markdown does
// not come from the ruleset templates, such as the changelog, so that every
// rendered page in the app shares one converter and one anchor scheme.
func MarkdownPage(md []byte) (string, error) {
	return markdownToHTML(md)
}

// markdownToHTML converts markdown to an HTML fragment with heading ids.
//
// Every documentation page must go through this, so the anchors on the docs page
// and the anchors on the changelog page are generated the same way.
func markdownToHTML(md []byte) (string, error) {
	gm := goldmark.New(
		goldmark.WithExtensions(extension.Table),
		// Without this no heading carries an id and every in-document link in
		// the rulebook is dead.
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	)
	var buf bytes.Buffer
	if err := gm.Convert(md, &buf); err != nil {
		return "", fmt.Errorf("converting markdown: %w", err)
	}
	return buf.String(), nil
}

// headingRe matches a rendered heading and captures its level, its id and its
// inner HTML. Anchors are read back out of the rendered output on purpose; see
// the file comment.
var headingRe = regexp.MustCompile(`(?is)<h([1-6])[^>]*\sid="([^"]*)"[^>]*>(.*?)</h[1-6]>`)

// tagRe strips inline markup from a heading's inner HTML, so a heading
// containing emphasis or code still yields plain sidebar text.
var tagRe = regexp.MustCompile(`(?s)<[^>]*>`)

// OutlineHTML returns the headings of a rendered HTML fragment, in document
// order. Only levels 2 and 3 are returned: level 1 is the document title and
// levels 4 and below are too fine-grained for navigation.
func OutlineHTML(htmlFragment string) []Heading {
	var out []Heading
	for _, m := range headingRe.FindAllStringSubmatch(htmlFragment, -1) {
		level, err := strconv.Atoi(m[1])
		if err != nil || level < 2 || level > 3 {
			continue
		}
		text := strings.TrimSpace(html.UnescapeString(tagRe.ReplaceAllString(m[3], "")))
		if text == "" || m[2] == "" {
			continue
		}
		out = append(out, Heading{Level: level, Text: text, ID: m[2]})
	}
	return out
}

// Tree groups a flat outline into chapters with their sections, which is the
// shape the sidebar template iterates.
//
// A level-3 heading that appears before any level-2 heading is promoted to a
// chapter of its own rather than dropped, so no section can become unreachable
// from the sidebar because of how a document happens to start.
func Tree(headings []Heading) []Children {
	var out []Children
	for _, h := range headings {
		if h.Level == 2 || len(out) == 0 {
			out = append(out, Children{Heading: h})
			continue
		}
		out[len(out)-1].Sections = append(out[len(out)-1].Sections, h)
	}
	return out
}

// RenderHTMLOutline renders the documentation to HTML and returns it together
// with its outline, so a handler does not have to convert or re-parse twice.
func RenderHTMLOutline(loaded *config.Loaded, lib PackageLister) (string, []Children, error) {
	md, err := RenderMarkdown(loaded, lib)
	if err != nil {
		return "", nil, err
	}
	h, err := markdownToHTML([]byte(md))
	if err != nil {
		return "", nil, err
	}
	return h, Tree(OutlineHTML(h)), nil
}
