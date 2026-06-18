package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/blok2ttrpg/charsheet/internal/models"
)

func (s *Server) handleAbilityDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	char := s.sessions.Get(r)
	if char == nil {
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}

	indexStr := r.URL.Query().Get("index")
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	_, err = char.RemoveAbility(index)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Re-render the ability list
	vm := BuildAbilityListVM(char)
	s.renderTemplate(w, "tab_abilities", vm)
}

func (s *Server) handleAbilityExport(w http.ResponseWriter, r *http.Request) {
	char := s.sessions.Get(r)
	if char == nil {
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}

	indexStr := r.URL.Query().Get("index")
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 || index >= len(char.Abilities) {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	ability := &char.Abilities[index]
	data, err := models.MarshalAbility(ability)
	if err != nil {
		http.Error(w, "Failed to export ability", http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("%s.yaml", ability.Name)
	if ability.Name == "" {
		filename = "ability.yaml"
	}

	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.Write(data)
}

func (s *Server) handleAbilityImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	char := s.sessions.Get(r)
	if char == nil {
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}

	// Parse multipart form (file upload)
	if err := r.ParseMultipartForm(1 << 20); err != nil { // 1MB max
		http.Error(w, "File too large or invalid", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("ability_file")
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

	ability, err := models.UnmarshalAbility(buf[:n])
	if err != nil {
		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showError": "Invalid ability YAML: %s"}`, err.Error()))
		vm := BuildAbilityListVM(char)
		s.renderTemplate(w, "tab_abilities", vm)
		return
	}

	// Check budget
	if err := char.AddAbility(*ability); err != nil {
		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showError": "%s"}`, err.Error()))
		vm := BuildAbilityListVM(char)
		s.renderTemplate(w, "tab_abilities", vm)
		return
	}

	vm := BuildAbilityListVM(char)
	s.renderTemplate(w, "tab_abilities", vm)
}

// handleAbilityNew creates a new empty ability with just a name and type, then opens the wizard.
// For now in Task 7, it just adds a minimal ability to the list.
func (s *Server) handleAbilityNew(w http.ResponseWriter, r *http.Request) {
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

	name := r.FormValue("name")
	abilityType := r.FormValue("type")

	if name == "" {
		w.Header().Set("HX-Trigger", `{"showError": "Ability name is required"}`)
		vm := BuildAbilityListVM(char)
		s.renderTemplate(w, "tab_abilities", vm)
		return
	}

	newAbility := models.Ability{
		Name:       name,
		Type:       models.AbilityType(abilityType),
		EnergyCost: 3, // Default base energy cost
		ActionCost: 2, // Default for Execution
	}

	// Set defaults based on type
	switch models.AbilityType(abilityType) {
	case models.AbilityTypeReaction:
		newAbility.ActionCost = 0
		newAbility.Range = 1
		newAbility.Uses = 1
	case models.AbilityTypePhase:
		newAbility.ActionCost = 0
		newAbility.PhaseDuration = 2
	case models.AbilityTypeMinion:
		newAbility.ActionCost = 0
		newAbility.MinionStats = &models.MinionStats{
			Health:   10,
			Attack:   "2d6",
			Defense:  "1d6",
			Speed:    "5m",
			Lifetime: 3,
		}
	}

	if err := char.AddAbility(newAbility); err != nil {
		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showError": "%s"}`, err.Error()))
		vm := BuildAbilityListVM(char)
		s.renderTemplate(w, "tab_abilities", vm)
		return
	}

	vm := BuildAbilityListVM(char)
	s.renderTemplate(w, "tab_abilities", vm)
}
