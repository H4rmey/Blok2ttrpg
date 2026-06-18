package server

import (
	"fmt"
	"net/http"
)

func (s *Server) handleLevelUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	char := s.sessions.Get(r)
	if char == nil {
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}

	if err := char.LevelUp(); err != nil {
		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showError": "%s"}`, err.Error()))
	}

	// Return the updated header partial
	s.renderTemplate(w, "level_controls", char)
}

func (s *Server) handleLevelDown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	char := s.sessions.Get(r)
	if char == nil {
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}

	if err := char.LevelDown(); err != nil {
		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showError": "%s"}`, err.Error()))
	}

	// Return the updated header partial
	s.renderTemplate(w, "level_controls", char)
}
