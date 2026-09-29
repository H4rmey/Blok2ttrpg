// This file checks that file_order and the docs tree agree with each other.
//
// file_order decides which markdown files are concatenated into the generated
// rulebook. Nothing used to verify the reverse direction, so a chapter could be
// written and simply never appear in the output: six files (the skill tree,
// magic system and world chapters) had drifted out of the rulebook that way,
// while docs/TODO.md had drifted in and was being published as chapter one.
//
// ValidateDocsOrder closes both gaps. It is a config-level check rather than a
// docs-level one because file_order is config, and reporting it at load time
// means the application refuses to start with an inconsistent rulebook instead
// of quietly serving an incomplete one.
package config

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// docsOrderExempt names markdown files under the docs tree that are deliberately
// not part of the generated rulebook. Keys are paths relative to the docs
// directory, using forward slashes.
//
// Paths are matched rather than base names so that exempting one chapter cannot
// silently exempt an unrelated file that happens to share its file name.
var docsOrderExempt = map[string]bool{
	// Maintainer notes, not a chapter.
	"TODO.md": true,

	// Module drafts that are still in progress. They are intentionally kept
	// out of the generated rulebook until their content is finished; remove
	// them from this map and add them to file_order to publish them.
	"modules/magic-system/curses.md":       true,
	"modules/magic-system/imbuing.md":      true,
	"modules/magic-system/magic-system.md": true,
	"modules/skill-trees.md":               true,
	"modules/world/culture.md":             true,
	"modules/world/world.md":               true,
}

// ValidateDocsOrder reports whether file_order and the markdown files under
// docsDir are consistent: every entry in file_order must exist on disk, must not
// be listed twice, and every non-exempt .md file under docsDir must be listed.
//
// It is a no-op when file_order is empty or docsDir does not exist, so rulesets
// that do not ship documentation are unaffected.
func (c *Config) ValidateDocsOrder(docsDir string) error {
	if len(c.FileOrder) == 0 {
		return nil
	}
	if _, err := os.Stat(docsDir); err != nil {
		// No docs tree to compare against (for example a ruleset loaded from
		// somewhere other than the project root). Nothing to verify.
		return nil
	}

	listed, err := listedDocs(c.FileOrder, docsDir)
	if err != nil {
		return err
	}
	onDisk, err := docsOnDisk(docsDir)
	if err != nil {
		return err
	}

	var missing []string
	for _, rel := range onDisk {
		if !listed[rel] {
			missing = append(missing, rel)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("file_order is missing %d documentation file(s), so they would never appear "+
			"in the generated rulebook: %s. Add them to file_order, or add them to docsOrderExempt "+
			"in internal/config/docsorder.go if they are maintainer notes",
			len(missing), strings.Join(missing, ", "))
	}
	return nil
}

// listedDocs normalises the file_order entries to paths relative to docsDir and
// verifies each one exists and appears only once.
func listedDocs(order []string, docsDir string) (map[string]bool, error) {
	listed := map[string]bool{}
	for _, entry := range order {
		clean := filepath.Clean(filepath.FromSlash(entry))
		if _, err := os.Stat(clean); err != nil {
			return nil, fmt.Errorf("file_order references %q which does not exist on disk", entry)
		}
		rel, err := filepath.Rel(docsDir, clean)
		if err != nil {
			// Entries outside the docs tree are allowed; they simply cannot be
			// cross-checked against it.
			continue
		}
		rel = filepath.ToSlash(rel)
		if strings.HasPrefix(rel, "../") {
			continue
		}
		if listed[rel] {
			return nil, fmt.Errorf("file_order lists %q more than once", entry)
		}
		listed[rel] = true
	}
	return listed, nil
}

// docsOnDisk returns every non-exempt markdown file under docsDir, as a path
// relative to docsDir using forward slashes.
func docsOnDisk(docsDir string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(docsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(path)) != ".md" {
			return nil
		}
		rel, err := filepath.Rel(docsDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if docsOrderExempt[rel] {
			return nil
		}
		out = append(out, rel)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walking %s: %w", docsDir, err)
	}
	return out, nil
}
