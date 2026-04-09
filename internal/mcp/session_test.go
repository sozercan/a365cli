package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// withTempHome overrides HOME (and the session cache path) for the duration of
// the test, then restores it on cleanup.
func withTempHome(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })
	return tmpDir
}

func TestSaveAndLoadSession(t *testing.T) {
	withTempHome(t)

	endpoint := "https://example.com/mcp/server1/"

	// No session should exist yet.
	if _, ok := LoadSession(endpoint); ok {
		t.Fatal("expected no cached session initially")
	}

	// Save a session.
	SaveSession(endpoint, "sess-abc-123")

	// Load it back.
	sid, ok := LoadSession(endpoint)
	if !ok {
		t.Fatal("expected cached session after save")
	}
	if sid != "sess-abc-123" {
		t.Errorf("expected session ID 'sess-abc-123', got '%s'", sid)
	}
}

func TestLoadSession_Expired(t *testing.T) {
	home := withTempHome(t)

	endpoint := "https://example.com/mcp/server1/"

	// Write an expired session directly.
	sessions := map[string]sessionEntry{
		endpoint: {
			SessionID: "old-session",
			ExpiresAt: time.Now().Add(-time.Hour),
		},
	}
	dir := filepath.Join(home, sessionDir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(sessions)
	if err := os.WriteFile(filepath.Join(dir, sessionFile), data, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, ok := LoadSession(endpoint); ok {
		t.Fatal("expected expired session to not be returned")
	}
}

func TestLoadSession_NoCacheFile(t *testing.T) {
	withTempHome(t)

	// No file on disk — should return false, not error.
	if _, ok := LoadSession("https://example.com/mcp/server1/"); ok {
		t.Fatal("expected no session when cache file doesn't exist")
	}
}

func TestSaveSession_MultipleEndpoints(t *testing.T) {
	withTempHome(t)

	SaveSession("https://example.com/mcp/server1/", "sess-1")
	SaveSession("https://example.com/mcp/server2/", "sess-2")

	sid1, ok1 := LoadSession("https://example.com/mcp/server1/")
	sid2, ok2 := LoadSession("https://example.com/mcp/server2/")

	if !ok1 || sid1 != "sess-1" {
		t.Errorf("expected sess-1, got %q (ok=%v)", sid1, ok1)
	}
	if !ok2 || sid2 != "sess-2" {
		t.Errorf("expected sess-2, got %q (ok=%v)", sid2, ok2)
	}
}

func TestClearSessions(t *testing.T) {
	home := withTempHome(t)

	SaveSession("https://example.com/mcp/server1/", "sess-1")

	// Verify file exists.
	path := filepath.Join(home, sessionDir, sessionFile)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected session file to exist: %v", err)
	}

	ClearSessions()

	// File should be gone.
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected session file to be removed, got err: %v", err)
	}

	// Loading should gracefully return false.
	if _, ok := LoadSession("https://example.com/mcp/server1/"); ok {
		t.Fatal("expected no session after clear")
	}
}

func TestClearSessions_NoFile(t *testing.T) {
	withTempHome(t)

	// Should not panic when no file exists.
	ClearSessions()
}

func TestClearSession_OnlyRemovesTargetEndpoint(t *testing.T) {
	withTempHome(t)

	endpoint1 := "https://example.com/mcp/server1/"
	endpoint2 := "https://example.com/mcp/server2/"

	SaveSession(endpoint1, "sess-1")
	SaveSession(endpoint2, "sess-2")
	SaveTools(endpoint1, []ToolInfo{{Name: "Tool1"}})
	SaveTools(endpoint2, []ToolInfo{{Name: "Tool2"}})

	ClearSession(endpoint1)

	if _, ok := LoadSession(endpoint1); ok {
		t.Fatal("expected endpoint1 session to be cleared")
	}
	if _, ok := LoadSession(endpoint2); !ok {
		t.Fatal("expected endpoint2 session to remain")
	}
	if tools := LoadTools(endpoint1); tools != nil {
		t.Fatal("expected endpoint1 tools to be cleared")
	}
	if tools := LoadTools(endpoint2); len(tools) != 1 || tools[0].Name != "Tool2" {
		t.Fatalf("expected endpoint2 tools to remain, got %+v", tools)
	}
}

func TestSaveSession_CorruptFile(t *testing.T) {
	home := withTempHome(t)

	// Write garbage to the sessions file.
	dir := filepath.Join(home, sessionDir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, sessionFile), []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}

	// SaveSession should handle corrupt file gracefully (start fresh).
	SaveSession("https://example.com/mcp/server1/", "new-sess")

	sid, ok := LoadSession("https://example.com/mcp/server1/")
	if !ok || sid != "new-sess" {
		t.Errorf("expected 'new-sess', got %q (ok=%v)", sid, ok)
	}
}

func TestSaveAndLoadTools(t *testing.T) {
	withTempHome(t)

	endpoint := "https://example.com/mcp/server1/"

	// No tools should exist yet.
	if tools := LoadTools(endpoint); tools != nil {
		t.Fatal("expected no cached tools initially")
	}

	// Save a session first (tools are stored within a session entry).
	SaveSession(endpoint, "sess-123")

	// Save tools for that endpoint.
	toolList := []ToolInfo{
		{Name: "ListChats", Description: "List chats"},
		{Name: "GetChat", Description: "Get a chat", InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"chatId": map[string]any{"type": "string"},
			},
			"required": []any{"chatId"},
		}},
	}
	SaveTools(endpoint, toolList)

	// Load tools back.
	loaded := LoadTools(endpoint)
	if len(loaded) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(loaded))
	}
	if loaded[0].Name != "ListChats" {
		t.Errorf("expected 'ListChats', got %q", loaded[0].Name)
	}
	if loaded[1].Name != "GetChat" {
		t.Errorf("expected 'GetChat', got %q", loaded[1].Name)
	}
}

func TestLoadTools_ExpiredSession(t *testing.T) {
	home := withTempHome(t)

	endpoint := "https://example.com/mcp/server1/"

	// Write an expired session with tools directly.
	sessions := map[string]sessionEntry{
		endpoint: {
			SessionID: "old-session",
			ExpiresAt: time.Now().Add(-time.Hour),
			Tools:     []ToolInfo{{Name: "StaleTool"}},
		},
	}
	dir := filepath.Join(home, sessionDir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(sessions)
	if err := os.WriteFile(filepath.Join(dir, sessionFile), data, 0o600); err != nil {
		t.Fatal(err)
	}

	// Should not return tools from an expired session.
	if tools := LoadTools(endpoint); tools != nil {
		t.Fatal("expected nil tools from expired session")
	}
}

func TestSaveTools_NoSession(t *testing.T) {
	withTempHome(t)

	// SaveTools without a session should be a no-op (no panic).
	SaveTools("https://example.com/mcp/server1/", []ToolInfo{{Name: "Tool1"}})

	// Should still not have tools since there was no session.
	if tools := LoadTools("https://example.com/mcp/server1/"); tools != nil {
		t.Fatal("expected nil tools when no session exists")
	}
}

func TestLoadTools_ClearedBySessionClear(t *testing.T) {
	withTempHome(t)

	endpoint := "https://example.com/mcp/server1/"
	SaveSession(endpoint, "sess-123")
	SaveTools(endpoint, []ToolInfo{{Name: "Tool1"}})

	// Verify tools exist.
	if tools := LoadTools(endpoint); len(tools) != 1 {
		t.Fatal("expected tools before clear")
	}

	ClearSessions()

	// After clearing sessions, tools should be gone too.
	if tools := LoadTools(endpoint); tools != nil {
		t.Fatal("expected nil tools after ClearSessions")
	}
}
