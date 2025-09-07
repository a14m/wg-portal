package internal

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type Session struct {
	Expires time.Time
}

type SessionManager struct {
	sessions map[string]*Session
	mutex    sync.RWMutex
}

func NewSessionManager() *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*Session),
	}
	// Start cleanup goroutine
	go sm.cleanupExpiredSessions()
	return sm
}

// Using the same logic that powers the pi-hole authentication
// GeneratePasswordHash creates a double SHA256 hash from the password param to
// validate against config.PasswordHash
func GeneratePasswordHash(password string) string {
	first := sha256.Sum256([]byte(password))
	firstHex := hex.EncodeToString(first[:])
	second := sha256.Sum256([]byte(firstHex))
	return hex.EncodeToString(second[:])
}

func ValidatePassword(password, hash string) bool {
	return GeneratePasswordHash(password) == hash
}

func (sm *SessionManager) CreateSession() (string, time.Time, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sessionID, err := generateSecureToken()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate session ID: %w", err)
	}

	expires := time.Now().Add(1 * time.Hour)
	sm.sessions[sessionID] = &Session{
		Expires: expires,
	}

	return sessionID, expires, nil
}

func (sm *SessionManager) ValidateSession(sessionID string) (*Session, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, false
	}

	if time.Now().After(session.Expires) {
		return nil, false
	}

	return session, true
}

func (sm *SessionManager) DeleteSession(sessionID string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	delete(sm.sessions, sessionID)
}

func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// cleanupExpiredSessions periodically removes expired sessions every 1 hour
func (sm *SessionManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		sm.mutex.Lock()
		now := time.Now()
		for sessionID, session := range sm.sessions {
			if now.After(session.Expires) {
				delete(sm.sessions, sessionID)
			}
		}
		sm.mutex.Unlock()
	}
}
