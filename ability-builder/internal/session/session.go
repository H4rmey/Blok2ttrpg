package session

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"

	"github.com/harmey/blok2ttrpg/ability-builder/internal/models"
)

const cookieName = "blok2_session"

// BuilderState holds the work-in-progress ability being constructed.
type BuilderState struct {
	CharacterID string
	Ability     models.Ability
}

// Manager manages in-memory sessions for the ability builder.
type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*BuilderState
}

// NewManager creates a new session manager.
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*BuilderState),
	}
}

// generateID creates a random session ID.
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GetOrCreate returns the session for the request, creating one if needed.
func (m *Manager) GetOrCreate(w http.ResponseWriter, r *http.Request) (string, *BuilderState) {
	cookie, err := r.Cookie(cookieName)
	if err == nil {
		m.mu.RLock()
		state, ok := m.sessions[cookie.Value]
		m.mu.RUnlock()
		if ok {
			return cookie.Value, state
		}
	}

	// Create new session
	id := generateID()
	state := &BuilderState{}

	m.mu.Lock()
	m.sessions[id] = state
	m.mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    id,
		Path:     "/",
		HttpOnly: true,
	})

	return id, state
}

// Get returns the session for the request, or nil if none exists.
func (m *Manager) Get(r *http.Request) *BuilderState {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[cookie.Value]
}

// Update replaces the session state.
func (m *Manager) Update(sessionID string, state *BuilderState) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[sessionID] = state
}

// Clear removes the session.
func (m *Manager) Clear(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
}
