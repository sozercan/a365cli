package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	sessionDir  = ".a365"
	sessionFile = "sessions.json"
	sessionTTL  = 30 * time.Minute
)

// sessionEntry is a single cached MCP session.
type sessionEntry struct {
	SessionID string    `json:"sessionID"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// sessionCachePath returns ~/.a365/sessions.json.
func sessionCachePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, sessionDir, sessionFile), nil
}

// loadAllSessions reads the entire session map from disk.
func loadAllSessions() (map[string]sessionEntry, error) {
	path, err := sessionCachePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var sessions map[string]sessionEntry
	if err := json.Unmarshal(data, &sessions); err != nil {
		return nil, err
	}
	return sessions, nil
}

// saveAllSessions writes the session map to disk.
func saveAllSessions(sessions map[string]sessionEntry) error {
	path, err := sessionCachePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

// LoadSession returns a cached session ID for the given endpoint if it exists
// and has not expired. Best-effort: errors are silently ignored.
func LoadSession(endpoint string) (string, bool) {
	sessions, err := loadAllSessions()
	if err != nil {
		return "", false
	}

	entry, ok := sessions[endpoint]
	if !ok {
		return "", false
	}

	if time.Now().After(entry.ExpiresAt) {
		return "", false
	}

	return entry.SessionID, true
}

// SaveSession caches a session ID for the given endpoint with a 30-minute TTL.
// Best-effort: errors are silently ignored.
func SaveSession(endpoint, sessionID string) {
	sessions, err := loadAllSessions()
	if err != nil {
		sessions = make(map[string]sessionEntry)
	}

	sessions[endpoint] = sessionEntry{
		SessionID: sessionID,
		ExpiresAt: time.Now().Add(sessionTTL),
	}

	_ = saveAllSessions(sessions)
}

// ClearSessions removes the session cache file entirely.
// Best-effort: errors are silently ignored.
func ClearSessions() {
	path, err := sessionCachePath()
	if err != nil {
		return
	}
	_ = os.Remove(path)
}
