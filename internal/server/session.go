package server

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/blok2ttrpg/charsheet/internal/models"
)

const sessionCookieName = "blok2_session"
const sessionTTL = 24 * time.Hour

// sessionEntry wraps a character with a last-accessed timestamp.
type sessionEntry struct {
	char       *models.Character
	lastAccess time.Time
}

// SessionStore manages in-memory character sessions.
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*sessionEntry
}

// NewSessionStore creates a new session store and starts cleanup.
func NewSessionStore() *SessionStore {
	ss := &SessionStore{
		sessions: make(map[string]*sessionEntry),
	}
	go ss.cleanupLoop()
	return ss
}

// GetOrCreate retrieves the character for the request's session,
// creating a new session if one doesn't exist.
func (ss *SessionStore) GetOrCreate(w http.ResponseWriter, r *http.Request) (string, *models.Character) {
	// Try to get existing session
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil && cookie.Value != "" {
		ss.mu.Lock()
		entry, exists := ss.sessions[cookie.Value]
		if exists {
			entry.lastAccess = time.Now()
			ss.mu.Unlock()
			return cookie.Value, entry.char
		}
		ss.mu.Unlock()
	}

	// Create new session
	id := generateSessionID()
	char := models.NewCharacter()

	ss.mu.Lock()
	ss.sessions[id] = &sessionEntry{char: char, lastAccess: time.Now()}
	ss.mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    id,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return id, char
}

// Get retrieves the character for the given session ID.
func (ss *SessionStore) Get(r *http.Request) *models.Character {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return nil
	}
	ss.mu.Lock()
	defer ss.mu.Unlock()
	entry, exists := ss.sessions[cookie.Value]
	if !exists {
		return nil
	}
	entry.lastAccess = time.Now()
	return entry.char
}

// Set replaces the character for the given session.
func (ss *SessionStore) Set(sessionID string, char *models.Character) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.sessions[sessionID] = &sessionEntry{char: char, lastAccess: time.Now()}
}

func (ss *SessionStore) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		ss.mu.Lock()
		now := time.Now()
		for id, entry := range ss.sessions {
			if now.Sub(entry.lastAccess) > sessionTTL {
				delete(ss.sessions, id)
			}
		}
		ss.mu.Unlock()
	}
}

func generateSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
