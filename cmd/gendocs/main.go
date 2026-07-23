// Command gendocs writes the rendered ruleset documentation to generated_docs.md.
// It is a small maintenance utility so the checked-in markdown export stays in
// sync with the config and the docs renderer.
package main

import (
	"log"
	"os"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/docs"
)

func main() {
	loaded, err := config.Load("config/ability-builder")
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}
	md, err := docs.RenderMarkdown(loaded)
	if err != nil {
		log.Fatalf("rendering docs: %v", err)
	}
	if err := os.WriteFile("generated_docs.md", []byte(md), 0o644); err != nil {
		log.Fatalf("writing generated_docs.md: %v", err)
	}
	log.Printf("wrote generated_docs.md (%d bytes)", len(md))
}
