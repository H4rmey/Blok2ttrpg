package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/harmey/blok2ttrpg/ability-builder/internal/handlers"
	"github.com/harmey/blok2ttrpg/ability-builder/internal/session"
	"github.com/harmey/blok2ttrpg/ability-builder/internal/storage"
)

func main() {
	// Configuration
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	dataFile := "data/characters.json"
	templateDir := "templates"

	// Initialize storage
	store, err := storage.New(dataFile)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize session manager
	sessions := session.NewManager()

	// Initialize app with templates
	app, err := handlers.NewApp(store, sessions, templateDir)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	mux := http.NewServeMux()

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Character routes
	mux.HandleFunc("/", app.IndexHandler)
	mux.HandleFunc("/characters/new", app.NewCharacterHandler)
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

	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Blok2 TTRPG Ability Builder starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
