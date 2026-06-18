package server

import (
	"net/http"
	"strconv"

	"github.com/blok2ttrpg/charsheet/internal/models"
)

func (s *Server) handleAttributeUpdate(w http.ResponseWriter, r *http.Request) {
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

	// Update whichever field was submitted
	field := r.FormValue("field")
	value := r.FormValue("value")

	switch field {
	case "name":
		char.Attributes.Name = value
	case "age":
		char.Attributes.Age = value
	case "size":
		char.Attributes.Size = value
	case "alignment":
		char.Attributes.Alignment = value
	case "backstory":
		char.Attributes.Backstory = value
	case "personality":
		char.Attributes.Personality = value
	case "traits":
		char.Attributes.Traits = value
	case "appearance":
		char.Attributes.Appearance = value
	case "hobbies":
		char.Attributes.Hobbies = value
	case "occupation":
		char.Attributes.Occupation = value
	case "inventory":
		char.Attributes.Inventory = value
	case "quirks":
		char.Attributes.Quirks = value
	default:
		http.Error(w, "Unknown field", http.StatusBadRequest)
		return
	}

	// Return empty 200 — the field is already updated in the input
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleTempAttributeAdd(w http.ResponseWriter, r *http.Request) {
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
	description := r.FormValue("description")
	duration := r.FormValue("duration")

	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	char.TemporaryAttributes = append(char.TemporaryAttributes, models.TemporaryAttribute{
		Name:        name,
		Description: description,
		Duration:    duration,
	})

	// Re-render the temp attributes section
	s.renderTemplate(w, "temp_attributes_list", char)
}

func (s *Server) handleTempAttributeRemove(w http.ResponseWriter, r *http.Request) {
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
	if err != nil || index < 0 || index >= len(char.TemporaryAttributes) {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	char.TemporaryAttributes = append(
		char.TemporaryAttributes[:index],
		char.TemporaryAttributes[index+1:]...,
	)

	// Re-render the temp attributes section
	s.renderTemplate(w, "temp_attributes_list", char)
}

func (s *Server) handleCustomFieldAdd(w http.ResponseWriter, r *http.Request) {
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

	label := r.FormValue("label")
	if label == "" {
		http.Error(w, "Label is required", http.StatusBadRequest)
		return
	}

	char.Attributes.Custom = append(char.Attributes.Custom, models.CustomField{
		Label: label,
		Value: "",
	})

	s.renderTemplate(w, "custom_fields_list", char)
}

func (s *Server) handleCustomFieldUpdate(w http.ResponseWriter, r *http.Request) {
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

	indexStr := r.FormValue("index")
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 || index >= len(char.Attributes.Custom) {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	char.Attributes.Custom[index].Value = r.FormValue("value")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleCustomFieldRemove(w http.ResponseWriter, r *http.Request) {
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
	if err != nil || index < 0 || index >= len(char.Attributes.Custom) {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	char.Attributes.Custom = append(
		char.Attributes.Custom[:index],
		char.Attributes.Custom[index+1:]...,
	)

	s.renderTemplate(w, "custom_fields_list", char)
}
