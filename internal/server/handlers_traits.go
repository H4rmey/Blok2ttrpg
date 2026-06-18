package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/blok2ttrpg/charsheet/internal/models"
)

func (s *Server) handleGeneralTraitUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	char := s.sessions.Get(r)
	if char == nil {
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	traitName := r.FormValue("trait")
	profStr := r.FormValue("proficiency")

	profVal, err := strconv.Atoi(profStr)
	if err != nil || profVal < 0 || profVal > int(models.Master) {
		http.Error(w, "Invalid proficiency value", http.StatusBadRequest)
		return
	}

	newProf := models.Proficiency(profVal)
	if err := char.SetGeneralTrait(traitName, newProf); err != nil {
		// Return the current state with an error indicator
		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showError": "%s"}`, err.Error()))
		vm := BuildGeneralTraitsVM(char)
		s.renderTemplate(w, "tab_general_traits", vm)
		return
	}

	// Return updated tab content
	vm := BuildGeneralTraitsVM(char)
	s.renderTemplate(w, "tab_general_traits", vm)
}

func (s *Server) handleCombativeTraitUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	char := s.sessions.Get(r)
	if char == nil {
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	section := r.FormValue("section")
	traitName := r.FormValue("trait")
	profStr := r.FormValue("proficiency")

	profVal, err := strconv.Atoi(profStr)
	if err != nil || profVal < 0 || profVal > int(models.Master) {
		http.Error(w, "Invalid proficiency value", http.StatusBadRequest)
		return
	}

	newProf := models.Proficiency(profVal)
	if err := char.SetCombativeTrait(section, traitName, newProf); err != nil {
		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showError": "%s"}`, err.Error()))
		vm := BuildCombativeTraitsVM(char)
		s.renderTemplate(w, "tab_combative_traits", vm)
		return
	}

	vm := BuildCombativeTraitsVM(char)
	s.renderTemplate(w, "tab_combative_traits", vm)
}
