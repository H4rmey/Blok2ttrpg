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
	configPath := flag.String("config", "config/ability-builder", "path to ruleset config directory or file")

	templateDir := flag.String("templates", "templates", "path to HTML templates")
	flag.Parse()

	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	if envCfg := os.Getenv("CONFIG"); envCfg != "" {
		*configPath = envCfg
	}

	loaded, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	dataFile := filepath.Join("data", loaded.ProfileID, "characters.json")
	st, err := store.New(dataFile)
	if err != nil {
		log.Fatalf("opening store: %v", err)
	}

	app, err := web.NewApp(loaded, st, *templateDir)
	if err != nil {
		log.Fatalf("initializing app: %v", err)
	}

	log.Printf("Loaded profile %q from %s", loaded.ProfileID, loaded.Dir)
	log.Printf("Characters stored in %s", dataFile)
	addr := fmt.Sprintf(":%s", port)
	log.Printf("%s starting on http://localhost%s", loaded.Title, addr)
	if err := http.ListenAndServe(addr, app.Router()); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
