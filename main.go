package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/harmey/blok2ttrpg/ability-builder/internal/config"
	"github.com/harmey/blok2ttrpg/ability-builder/internal/handlers"
	"github.com/harmey/blok2ttrpg/ability-builder/internal/session"
	"github.com/harmey/blok2ttrpg/ability-builder/internal/storage"
)

func main() {
	configPath := resolveConfigPath()

	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	templateDir := "templates"

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load ability builder config: %v", err)
	}

	dataFile, err := ensureProfileStore(cfg.ProfileID)
	if err != nil {
		log.Fatalf("Failed to prepare profile storage: %v", err)
	}

	store, err := storage.New(dataFile)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	sessions := session.NewManagerForProfile(cfg.ProfileID)

	log.Printf("Using config %s with profile %s", configPath, cfg.ProfileID)
	log.Printf("Using character storage %s", dataFile)

	// Initialize app with templates
	app, err := handlers.NewApp(store, sessions, templateDir, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	mux := http.NewServeMux()

	// Static files. We disable heuristic caching so edits to builder.js / css
	// are picked up immediately during development (browser must revalidate).
	mux.Handle("/static/", noCache(http.StripPrefix("/static/", http.FileServer(http.Dir("static")))))

	// Character routes
	mux.HandleFunc("/", app.IndexHandler)
	mux.HandleFunc("/characters/new", app.NewCharacterHandler)
	mux.HandleFunc("/characters/import", app.ImportCharacterYAMLHandler)
	mux.HandleFunc("/characters", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/characters" {
			app.CreateCharacterHandler(w, r)
			return
		}
		http.NotFound(w, r)
	})

	// Character CRUD - pattern: /characters/{id}[/...]
	mux.HandleFunc("/characters/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/characters/")
		parts := strings.Split(path, "/")

		if len(parts) == 0 || parts[0] == "" {
			http.NotFound(w, r)
			return
		}

		charID := parts[0]
		if charID == "new" {
			app.NewCharacterHandler(w, r)
			return
		}
		if charID == "import" {
			app.ImportCharacterYAMLHandler(w, r)
			return
		}

		// /characters/{id}
		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				app.ViewCharacterHandler(w, r)
			case http.MethodPost:
				app.UpdateCharacterHandler(w, r)
			case http.MethodDelete:
				app.DeleteCharacterHandler(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if len(parts) == 2 && parts[1] == "export" {
			app.ExportCharacterYAMLHandler(w, r)
			return
		}

		if len(parts) == 2 && parts[1] == "pdf" {
			app.PdfCharacterHandler(w, r)
			return
		}

		// /characters/{id}/abilities[/...]
		if parts[1] == "abilities" {
			if len(parts) == 2 {
				// /characters/{id}/abilities
				app.AbilityListHandler(w, r)
				return
			}

			if parts[2] == "new" {
				// /characters/{id}/abilities/new
				app.BuilderHandler(w, r)
				return
			}

			abilityID := parts[2]
			_ = abilityID

			if len(parts) == 3 {
				// /characters/{id}/abilities/{aid}
				switch r.Method {
				case http.MethodGet:
					app.AbilityDetailHandler(w, r)
				case http.MethodDelete:
					app.DeleteAbilityHandler(w, r)
				default:
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
				return
			}

			if len(parts) == 4 && parts[3] == "export" {
				// /characters/{id}/abilities/{aid}/export
				app.ExportAbilityYAMLHandler(w, r)
				return
			}

			if len(parts) == 4 && parts[3] == "edit" {
				// /characters/{id}/abilities/{aid}/edit
				app.EditAbilityHandler(w, r)
				return
			}
		}

		http.NotFound(w, r)
	})

	// Ability builder partial routes
	mux.HandleFunc("/partials/ability-type-config", app.AbilityTypeConfigHandler)
	mux.HandleFunc("/partials/enactment", app.AddEnactmentHandler)
	mux.HandleFunc("/partials/enactment-config", app.EnactmentConfigHandler)
	mux.HandleFunc("/partials/interaction-config", app.InteractionConfigHandler)

	// Ability save/review/reset
	mux.HandleFunc("/ability/save", app.SaveAbilityHandler)
	mux.HandleFunc("/ability/review", app.ReviewHandler)
	mux.HandleFunc("/ability/reset", app.ResetHandler)

	// Documentation download
	mux.HandleFunc("/docs/download", app.DownloadDocsHandler)
	mux.HandleFunc("/docs/pdf", app.PdfDocsHandler)

	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Blok2 TTRPG Ability Builder starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func resolveConfigPath() string {
	configFlag := flag.String("config", "", "path to ability builder YAML config")
	flag.Parse()
	if *configFlag != "" {
		return *configFlag
	}
	if configPath := os.Getenv("ABILITY_BUILDER_CONFIG"); configPath != "" {
		return configPath
	}
	return config.DefaultPath()
}

func ensureProfileStore(profileID string) (string, error) {
	dataFile := filepath.Join("data", profileID, "characters.json")
	if err := os.MkdirAll(filepath.Dir(dataFile), 0755); err != nil {
		return "", err
	}
	if profileID == "ability-builder" {
		legacyFile := filepath.Join("data", "characters.json")
		if !fileExists(dataFile) && fileExists(legacyFile) {
			data, err := os.ReadFile(legacyFile)
			if err != nil {
				return "", err
			}
			if err := os.WriteFile(dataFile, data, 0644); err != nil {
				return "", err
			}
		}
	}
	return dataFile, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// noCache wraps a handler so responses are never served from a stale
// heuristic cache: the browser must always revalidate with the origin.
func noCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, must-revalidate")
		h.ServeHTTP(w, r)
	})
}
