package main

import (
	"fmt"
	"os"

	"github.com/harmey/blok2ttrpg/ability-builder/internal/config"
	"github.com/harmey/blok2ttrpg/ability-builder/internal/docs"
)

func main() {
	cfgPath := config.DefaultPath()
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	outPath := "generated_docs.md"
	if len(os.Args) > 2 {
		outPath = os.Args[2]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	rendered, err := docs.RenderFullDocumentation(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to render docs: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outPath, []byte(rendered), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated documentation: %s\n", outPath)
}
