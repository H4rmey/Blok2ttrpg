// Package docs renders the ruleset documentation. Unlike v4, it does not name
// any specific ability type or enactment: it simply passes the whole config to
// each markdown template and lets the template iterate. Docs therefore stay in
// sync with the config automatically.
package docs

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

// RenderMarkdown builds the full markdown documentation by rendering each file
// listed in cfg.Docs.Order (relative to dir) against the config as template
// data, then concatenating the results.
func RenderMarkdown(loaded *config.Loaded) (string, error) {
	if loaded == nil || loaded.Config == nil {
		return "", fmt.Errorf("config is nil")
	}
	order := loaded.FileOrder
	if len(order) == 0 {
		return "", fmt.Errorf("no file_order configured")
	}

	var sections []string
	for _, rel := range order {
		path := filepath.Join(loaded.Dir, filepath.FromSlash(rel))
		raw, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("reading doc %q: %w", rel, err)
		}
		tmpl, err := template.New(filepath.Base(path)).Parse(string(raw))
		if err != nil {
			return "", fmt.Errorf("parsing doc %q: %w", rel, err)
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, loaded.Config); err != nil {
			return "", fmt.Errorf("executing doc %q: %w", rel, err)
		}
		sections = append(sections, strings.TrimSpace(buf.String()))
	}
	return strings.Join(sections, "\n\n"), nil
}

// RenderHTML converts the markdown documentation to an HTML fragment.
func RenderHTML(loaded *config.Loaded) (string, error) {
	md, err := RenderMarkdown(loaded)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	gm := goldmark.New(goldmark.WithExtensions(extension.Table))
	if err := gm.Convert([]byte(md), &buf); err != nil {
		return "", fmt.Errorf("converting markdown: %w", err)
	}
	return buf.String(), nil
}
