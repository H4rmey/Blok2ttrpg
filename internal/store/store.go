package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/harmey/blok2ttrpg-v5/internal/model"
)

// Store persists characters to a single JSON file. It is safe for concurrent
// use and writes atomically via a temp file rename.
type Store struct {
	path string
	mu   sync.RWMutex
	data map[string]model.Character
}

// New opens (or creates) a store backed by the given JSON file.
func New(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("creating data dir: %w", err)
	}
	s := &Store{path: path, data: map[string]model.Character{}}
	if b, err := os.ReadFile(path); err == nil && len(b) > 0 {
		if err := json.Unmarshal(b, &s.data); err != nil {
			return nil, fmt.Errorf("parsing store %q: %w", path, err)
		}
	}
	return s, nil
}

// List returns all characters sorted by id.
func (s *Store) List() []model.Character {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.Character, 0, len(s.data))
	for _, c := range s.data {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Get returns a character by id.
func (s *Store) Get(id string) (model.Character, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.data[id]
	return c, ok
}

// Save inserts or updates a character and persists to disk.
func (s *Store) Save(c model.Character) error {
	s.mu.Lock()
	s.data[c.ID] = c
	s.mu.Unlock()
	return s.flush()
}

// Delete removes a character and persists to disk.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	delete(s.data, id)
	s.mu.Unlock()
	return s.flush()
}

func (s *Store) flush() error {
	s.mu.RLock()
	b, err := json.MarshalIndent(s.data, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return fmt.Errorf("encoding store: %w", err)
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return fmt.Errorf("writing store: %w", err)
	}
	return os.Rename(tmp, s.path)
}
