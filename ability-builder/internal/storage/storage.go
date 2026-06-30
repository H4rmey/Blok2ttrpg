package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/harmey/blok2ttrpg/ability-builder/internal/models"
)

// Store provides JSON file-based persistence for characters and their abilities.
type Store struct {
	mu       sync.RWMutex
	filepath string
	data     StoreData
}

// StoreData is the top-level structure persisted to disk.
type StoreData struct {
	Characters []models.Character `json:"characters"`
}

// New creates a new Store, loading existing data from the file if it exists.
func New(filepath string) (*Store, error) {
	s := &Store{
		filepath: filepath,
		data:     StoreData{Characters: []models.Character{}},
	}

	// Try to load existing data
	if _, err := os.Stat(filepath); err == nil {
		if err := s.load(); err != nil {
			return nil, fmt.Errorf("failed to load store: %w", err)
		}
	}

	return s, nil
}

// load reads the JSON file into memory.
func (s *Store) load() error {
	data, err := os.ReadFile(s.filepath)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, &s.data)
}

// save writes the in-memory data to the JSON file.
func (s *Store) save() error {
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filepath, data, 0644)
}

// ListCharacters returns all characters.
func (s *Store) ListCharacters() []models.Character {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]models.Character, len(s.data.Characters))
	copy(result, s.data.Characters)
	return result
}

// GetCharacter returns a character by ID.
func (s *Store) GetCharacter(id string) (*models.Character, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.data.Characters {
		if s.data.Characters[i].ID == id {
			c := s.data.Characters[i]
			return &c, nil
		}
	}
	return nil, fmt.Errorf("character not found: %s", id)
}

// SaveCharacter creates or updates a character.
func (s *Store) SaveCharacter(c models.Character) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.data.Characters {
		if s.data.Characters[i].ID == c.ID {
			s.data.Characters[i] = c
			return s.save()
		}
	}

	s.data.Characters = append(s.data.Characters, c)
	return s.save()
}

// DeleteCharacter removes a character by ID.
func (s *Store) DeleteCharacter(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.data.Characters {
		if s.data.Characters[i].ID == id {
			s.data.Characters = append(s.data.Characters[:i], s.data.Characters[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("character not found: %s", id)
}

// AddAbility adds an ability to a character.
func (s *Store) AddAbility(characterID string, ability models.Ability) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.data.Characters {
		if s.data.Characters[i].ID == characterID {
			s.data.Characters[i].Abilities = append(s.data.Characters[i].Abilities, ability)
			return s.save()
		}
	}
	return fmt.Errorf("character not found: %s", characterID)
}

// GetAbility returns a specific ability from a character.
func (s *Store) GetAbility(characterID, abilityID string) (*models.Ability, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i := range s.data.Characters {
		if s.data.Characters[i].ID == characterID {
			for j := range s.data.Characters[i].Abilities {
				if s.data.Characters[i].Abilities[j].ID == abilityID {
					a := s.data.Characters[i].Abilities[j]
					return &a, nil
				}
			}
			return nil, fmt.Errorf("ability not found: %s", abilityID)
		}
	}
	return nil, fmt.Errorf("character not found: %s", characterID)
}

// DeleteAbility removes an ability from a character.
func (s *Store) DeleteAbility(characterID, abilityID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.data.Characters {
		if s.data.Characters[i].ID == characterID {
			for j := range s.data.Characters[i].Abilities {
				if s.data.Characters[i].Abilities[j].ID == abilityID {
					s.data.Characters[i].Abilities = append(
						s.data.Characters[i].Abilities[:j],
						s.data.Characters[i].Abilities[j+1:]...,
					)
					return s.save()
				}
			}
			return fmt.Errorf("ability not found: %s", abilityID)
		}
	}
	return fmt.Errorf("character not found: %s", characterID)
}
