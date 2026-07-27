// Command blok2ttrpg-v5 is a config-driven TTRPG character and ability builder.
// The entire ruleset lives in a YAML config directory; the Go code only knows
// how to render generic fields, compute advisory costs, and persist characters.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/harmey/blok2ttrpg-v5/internal/config"
	"github.com/harmey/blok2ttrpg-v5/internal/store"
	"github.com/harmey/blok2ttrpg-v5/internal/web"
)

func main() {
	// --system selects which ruleset under config/ to load and which library
	// under library/ to serve. It is the primary switch for choosing a game
	// system on startup. --config still allows pointing at an explicit config
	// path (directory or file) and takes precedence when set to a non-default
	// value.
	system := flag.String("system", "ability-builder", "game system to load (subdirectory under config/ and library/)")
	configPath := flag.String("config", "", "explicit path to ruleset config directory or file (overrides --system)")

	templateDir := flag.String("templates", "templates", "path to HTML templates")
	libraryDir := flag.String("library", "", "explicit path to the built-in content library (overrides the per-system default)")
	flag.Parse()

	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	// Environment overrides. SYSTEM mirrors --system; CONFIG mirrors --config.
	if envSys := os.Getenv("SYSTEM"); envSys != "" {
		*system = envSys
	}
	if envCfg := os.Getenv("CONFIG"); envCfg != "" {
		*configPath = envCfg
	}
	if envLib := os.Getenv("LIBRARY"); envLib != "" {
		*libraryDir = envLib
	}

	// Resolve the config path: an explicit --config/CONFIG wins; otherwise the
	// path is derived from the selected system.
	resolvedConfig := *configPath
	if resolvedConfig == "" {
		resolvedConfig = filepath.Join("config", *system)
	}

	loaded, err := config.Load(resolvedConfig)
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	// Resolve the library root. An explicit --library/LIBRARY wins. Otherwise
	// prefer a per-system library (library/<profile_id>), falling back to the
	// flat "library" directory for backward compatibility.
	resolvedLibrary := *libraryDir
	if resolvedLibrary == "" {
		perSystem := filepath.Join("library", loaded.ProfileID)
		if info, statErr := os.Stat(perSystem); statErr == nil && info.IsDir() {
			resolvedLibrary = perSystem
		} else {
			resolvedLibrary = "library"
		}
	}

	dataFile := filepath.Join("data", loaded.ProfileID, "characters.json")
	st, err := store.New(dataFile)
	if err != nil {
		log.Fatalf("opening store: %v", err)
	}

	app, err := web.NewApp(loaded, st, *templateDir, resolvedLibrary)
	if err != nil {
		log.Fatalf("initializing app: %v", err)
	}

	log.Printf("Loaded profile %q from %s", loaded.ProfileID, loaded.Dir)
	log.Printf("Serving library from %s", resolvedLibrary)
	log.Printf("Characters stored in %s", dataFile)
	addr := fmt.Sprintf(":%s", port)
	log.Printf("%s starting on http://localhost%s", loaded.Title, addr)
	if err := http.ListenAndServe(addr, app.Router()); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
