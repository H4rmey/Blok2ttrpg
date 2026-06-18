package server

import "net/http"

func (s *Server) handleTabAttributes(w http.ResponseWriter, r *http.Request) {
	char := s.sessions.Get(r)
	if char == nil {
		_, char = s.sessions.GetOrCreate(w, r)
	}
	s.renderTemplate(w, "tab_attributes", char)
}

func (s *Server) handleTabGeneralTraits(w http.ResponseWriter, r *http.Request) {
	char := s.sessions.Get(r)
	if char == nil {
		_, char = s.sessions.GetOrCreate(w, r)
	}
	vm := BuildGeneralTraitsVM(char)
	s.renderTemplate(w, "tab_general_traits", vm)
}

func (s *Server) handleTabCombativeTraits(w http.ResponseWriter, r *http.Request) {
	char := s.sessions.Get(r)
	if char == nil {
		_, char = s.sessions.GetOrCreate(w, r)
	}
	vm := BuildCombativeTraitsVM(char)
	s.renderTemplate(w, "tab_combative_traits", vm)
}

func (s *Server) handleTabAbilities(w http.ResponseWriter, r *http.Request) {
	char := s.sessions.Get(r)
	if char == nil {
		_, char = s.sessions.GetOrCreate(w, r)
	}
	vm := BuildAbilityListVM(char)
	s.renderTemplate(w, "tab_abilities", vm)
}
