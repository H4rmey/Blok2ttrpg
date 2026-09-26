// Command gendocs writes the rendered ruleset documentation to generated_docs.md.
// It is a small maintenance utility so the checked-in markdown export stays in
// sync with the config and the docs renderer.
//
// After rendering, the output is linted (see internal/docs/lint.go) and the
// command fails when the lint finds a problem such as a table that lost all of
// its rows. Pass -allow-warnings to write the file anyway and report the
// findings as warnings, which is useful while a fix is in progress.
package main

import (
	"flag"
	"log"
	"os"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/docs"
	"github.com/harmey/blok2ttrpg-v5/internal/premade"
)

func main() {
	allowWarnings := flag.Bool("allow-warnings", false,
		"write the output and report lint findings as warnings instead of failing")
	configDir := flag.String("config", "config/Blok2Simplified", "config directory to render from")
	libraryDir := flag.String("library", "library", "content library directory")
	out := flag.String("out", "generated_docs.md", "output markdown file")
	flag.Parse()

	loaded, err := config.Load(*configDir)
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}
	md, err := docs.RenderMarkdown(loaded, premade.New(*libraryDir))
	if err != nil {
		log.Fatalf("rendering docs: %v", err)
	}

	findings := docs.Lint(md)
	if len(findings) > 0 && !*allowWarnings {
		log.Fatalf("documentation lint found %d problem(s):\n%s\n\n"+
			"These indicate the generated docs lost data (for example a template range "+
			"that produced no rows). Fix the template or the config, or re-run with "+
			"-allow-warnings to write the file anyway.",
			len(findings), docs.FindingsText(findings))
	}
	if len(findings) > 0 {
		log.Printf("warning: documentation lint found %d problem(s):\n%s",
			len(findings), docs.FindingsText(findings))
	}

	if err := os.WriteFile(*out, []byte(md), 0o644); err != nil {
		log.Fatalf("writing %s: %v", *out, err)
	}
	log.Printf("wrote %s (%d bytes)", *out, len(md))
}
