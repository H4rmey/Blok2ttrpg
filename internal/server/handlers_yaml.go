package server

import (
	"fmt"
	"net/http"

	"github.com/blok2ttrpg/charsheet/internal/models"
)

func (s *Server) handleCharacterExport(w http.ResponseWriter, r *http.Request) {
	char := s.sessions.Get(r)
	if char == nil {
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}

	data, err := models.MarshalCharacter(char)
	if err != nil {
		http.Error(w, "Failed to export character", http.StatusInternalServerError)
		return
	}

	filename := "character.yaml"
	if char.Attributes.Name != "" {
		filename = fmt.Sprintf("%s.yaml", char.Attributes.Name)
	}

	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.Write(data)
}

func (s *Server) handleCharacterImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session ID
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}
	sessionID := cookie.Value

	// Parse multipart form (file upload)
	if err := r.ParseMultipartForm(1 << 20); err != nil { // 1MB max
		http.Error(w, "File too large or invalid", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("character_file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file content
	buf := make([]byte, 1<<20)
	n, _ := file.Read(buf)
	if n == 0 {
		http.Error(w, "Empty file", http.StatusBadRequest)
		return
	}

	char, err := models.UnmarshalCharacter(buf[:n])
	if err != nil {
		// Return the page with an error — redirect to index with error
		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showError": "Invalid character YAML: %s"}`, err.Error()))
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Replace the session's character
	s.sessions.Set(sessionID, char)

	// Redirect to reload the full page with new character data
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleCharacterNew(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}

	// Reset to new character
	s.sessions.Set(cookie.Value, models.NewCharacter())

	// Redirect to reload the full page
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}
