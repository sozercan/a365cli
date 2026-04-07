package teams

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/output"
)

// setupTestServer creates a mock MCP server and returns a wired commands.Context,
// the output buffer, and a cleanup function. toolResponses maps MCP tool names
// to the JSON strings the mock should return in Content[0].Text.
func setupTestServer(t *testing.T, toolResponses map[string]string) (*commands.Context, *bytes.Buffer, func()) {
	return setupTestServerWithSchemas(t, toolResponses, nil)
}

// setupTestServerWithSchemas creates a mock MCP server that also responds to
// tools/list requests with the provided tool schemas.
func setupTestServerWithSchemas(t *testing.T, toolResponses map[string]string, toolSchemas []mcp.ToolInfo) (*commands.Context, *bytes.Buffer, func()) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		var req struct {
			ID     int    `json:"id"`
			Method string `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Mcp-Session-Id", "test-session-id")

		switch req.Method {
		case "initialize":
			fmt.Fprintf(w, "event: message\ndata: %s\n\n",
				mustJSON(map[string]any{
					"jsonrpc": "2.0",
					"id":      req.ID,
					"result": map[string]any{
						"protocolVersion": "2024-11-05",
						"serverInfo":      map[string]any{"name": "test", "version": "1.0"},
					},
				}))
		case "tools/call":
			var params struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments"`
			}
			json.Unmarshal(req.Params, &params) //nolint:errcheck

			respText, ok := toolResponses[params.Name]
			if !ok {
				respText = `{"message":"ok"}`
			}
			fmt.Fprintf(w, "event: message\ndata: %s\n\n",
				mustJSON(map[string]any{
					"jsonrpc": "2.0",
					"id":      req.ID,
					"result": map[string]any{
						"content": []map[string]any{
							{"type": "text", "text": respText},
						},
					},
				}))
		case "tools/list":
			tools := toolSchemas
			if tools == nil {
				tools = []mcp.ToolInfo{}
			}
			fmt.Fprintf(w, "event: message\ndata: %s\n\n",
				mustJSON(map[string]any{
					"jsonrpc": "2.0",
					"id":      req.ID,
					"result": map[string]any{
						"tools": tools,
					},
				}))
		default:
			fmt.Fprintf(w, "event: message\ndata: %s\n\n",
				mustJSON(map[string]any{
					"jsonrpc": "2.0",
					"id":      req.ID,
					"error":   map[string]any{"code": -32601, "message": "unknown method"},
				}))
		}
	}))

	origEndpoint := os.Getenv("A365_ENDPOINT")
	os.Setenv("A365_ENDPOINT", server.URL+"/")

	var buf bytes.Buffer
	ctx := &commands.Context{
		Ctx: context.Background(),
		TokenProvider: func(ctx context.Context) (string, error) {
			return "test-token", nil
		},
		Output:  &output.Formatter{Format: output.FormatJSON, Writer: &buf},
		UserUPN: "test@example.com",
	}

	cleanup := func() {
		server.Close()
		if origEndpoint == "" {
			os.Unsetenv("A365_ENDPOINT")
		} else {
			os.Setenv("A365_ENDPOINT", origEndpoint)
		}
	}

	return ctx, &buf, cleanup
}

// mustJSON marshals v to a JSON string, panicking on error.
func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// --- Tests ---

func TestTeamsListCmd_Run(t *testing.T) {
	ctx, buf, cleanup := setupTestServer(t, map[string]string{
		"ListTeams": `{"teams":[{"id":"t1","displayName":"Team A"},{"id":"t2","displayName":"Team B"}]}`,
	})
	defer cleanup()

	cmd := &TeamsListCmd{Max: 100}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	teams, ok := result["teams"]
	if !ok {
		t.Fatalf("expected 'teams' key in output, got: %s", buf.String())
	}
	arr, ok := teams.([]any)
	if !ok {
		t.Fatalf("expected 'teams' to be an array, got: %T", teams)
	}
	if len(arr) != 2 {
		t.Fatalf("expected 2 teams, got %d", len(arr))
	}
}

func TestTeamsGetCmd_Run(t *testing.T) {
	ctx, buf, cleanup := setupTestServer(t, map[string]string{
		"GetTeam": `{"id":"t1","displayName":"Team A","description":"Test team"}`,
	})
	defer cleanup()

	cmd := &TeamsGetCmd{ID: "t1"}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if result["id"] != "t1" {
		t.Errorf("expected id=t1, got %v", result["id"])
	}
	if result["displayName"] != "Team A" {
		t.Errorf("expected displayName=Team A, got %v", result["displayName"])
	}
}

func TestChatsSendCmd_Run(t *testing.T) {
	ctx, buf, cleanup := setupTestServer(t, map[string]string{
		"PostMessage": `{"id":"msg1","chatId":"chat1","content":"hello"}`,
	})
	defer cleanup()

	cmd := &ChatsSendCmd{ChatID: "chat1", Message: "hello"}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if result["id"] != "msg1" {
		t.Errorf("expected id=msg1, got %v", result["id"])
	}
}

func TestChatsSendCmd_DryRun(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "PostMessage",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"chatId":  map[string]any{"type": "string"},
					"content": map[string]any{"type": "string"},
				},
				"required": []any{"chatId", "content"},
			},
		},
	}
	ctx, buf, cleanup := setupTestServerWithSchemas(t, nil, schemas)
	defer cleanup()
	ctx.DryRun = true

	cmd := &ChatsSendCmd{ChatID: "chat1", Message: "hello"}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	dryRun, ok := result["dry_run"]
	if !ok {
		t.Fatal("expected 'dry_run' key in output")
	}
	if dryRun != true {
		t.Errorf("expected dry_run=true, got %v", dryRun)
	}
	if result["action"] != "chats.send" {
		t.Errorf("expected action=chats.send, got %v", result["action"])
	}
	val, ok := result["validation"].(map[string]any)
	if !ok {
		t.Fatal("expected validation object in dry-run output")
	}
	if val["valid"] != true {
		t.Errorf("expected valid=true, got %v; errors: %v", val["valid"], val["errors"])
	}
}

func TestChatsDeleteCmd_Force(t *testing.T) {
	ctx, buf, cleanup := setupTestServer(t, map[string]string{
		"DeleteChat": `{"message":"Chat deleted"}`,
	})
	defer cleanup()
	ctx.Force = true

	cmd := &ChatsDeleteCmd{ChatID: "chat1"}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if result["message"] != "Chat deleted" {
		t.Errorf("expected message='Chat deleted', got %v", result["message"])
	}
}

func TestChatsDeleteCmd_NoInput(t *testing.T) {
	ctx, _, cleanup := setupTestServer(t, nil)
	defer cleanup()
	ctx.NoInput = true

	cmd := &ChatsDeleteCmd{ChatID: "chat1"}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when NoInput=true and Force=false")
	}
	if !strings.Contains(err.Error(), "without --force") {
		t.Errorf("expected error about --force, got: %v", err)
	}
}

func TestTeamsListCmd_Max(t *testing.T) {
	// Return 5 teams, but set Max=2 to verify truncation.
	teamsJSON := `{"teams":[` +
		`{"id":"t1","displayName":"Team 1"},` +
		`{"id":"t2","displayName":"Team 2"},` +
		`{"id":"t3","displayName":"Team 3"},` +
		`{"id":"t4","displayName":"Team 4"},` +
		`{"id":"t5","displayName":"Team 5"}` +
		`]}`

	ctx, buf, cleanup := setupTestServer(t, map[string]string{
		"ListTeams": teamsJSON,
	})
	defer cleanup()

	cmd := &TeamsListCmd{Max: 2}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	teams := result["teams"].([]any)
	if len(teams) != 2 {
		t.Fatalf("expected 2 teams after truncation, got %d", len(teams))
	}
	// Verify we kept the first 2
	first := teams[0].(map[string]any)
	if first["id"] != "t1" {
		t.Errorf("expected first team id=t1, got %v", first["id"])
	}
}

func TestChannelsListCmd_Run(t *testing.T) {
	ctx, buf, cleanup := setupTestServer(t, map[string]string{
		"ListChannels": `{"channels":[{"id":"ch1","displayName":"General","membershipType":"standard"}]}`,
	})
	defer cleanup()

	cmd := &ChannelsListCmd{TeamID: "t1", Max: 100}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	channels, ok := result["channels"]
	if !ok {
		t.Fatalf("expected 'channels' key in output, got: %s", buf.String())
	}
	arr := channels.([]any)
	if len(arr) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(arr))
	}
	ch := arr[0].(map[string]any)
	if ch["displayName"] != "General" {
		t.Errorf("expected displayName=General, got %v", ch["displayName"])
	}
}

func TestSearchCmd_Run(t *testing.T) {
	ctx, buf, cleanup := setupTestServer(t, map[string]string{
		"SearchTeamMessagesQueryParameters": `{"hits":[{"summary":"budget meeting","createdDateTime":"2024-01-15T10:00:00Z"}]}`,
	})
	defer cleanup()

	cmd := &SearchCmd{Query: "budget", Size: 25}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	hits, ok := result["results"]
	if !ok {
		t.Fatalf("expected 'results' key in output, got keys: %v", keys(result))
	}
	arr := hits.([]any)
	if len(arr) != 1 {
		t.Fatalf("expected 1 result, got %d", len(arr))
	}
}

func TestChatsListCmd_Run(t *testing.T) {
	ctx, buf, cleanup := setupTestServer(t, map[string]string{
		"ListChats": `{"chats":[{"id":"c1","chatType":"oneOnOne","topic":""},{"id":"c2","chatType":"group","topic":"Project X"}]}`,
	})
	defer cleanup()

	cmd := &ChatsListCmd{Max: 50}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	chats, ok := result["chats"]
	if !ok {
		t.Fatalf("expected 'chats' key in output, got: %s", buf.String())
	}
	arr := chats.([]any)
	if len(arr) != 2 {
		t.Fatalf("expected 2 chats, got %d", len(arr))
	}
}

func TestChannelsPostCmd_DryRun(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "PostChannelMessage",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"teamId":    map[string]any{"type": "string"},
					"channelId": map[string]any{"type": "string"},
					"content":   map[string]any{"type": "string"},
				},
				"required": []any{"teamId", "channelId", "content"},
			},
		},
	}
	ctx, buf, cleanup := setupTestServerWithSchemas(t, nil, schemas)
	defer cleanup()
	ctx.DryRun = true

	cmd := &ChannelsPostCmd{TeamID: "t1", ChannelID: "ch1", Message: "hello"}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if result["dry_run"] != true {
		t.Errorf("expected dry_run=true, got %v", result["dry_run"])
	}
	if result["action"] != "channels.post" {
		t.Errorf("expected action=channels.post, got %v", result["action"])
	}
	val, ok := result["validation"].(map[string]any)
	if !ok {
		t.Fatal("expected validation object in dry-run output")
	}
	if val["valid"] != true {
		t.Errorf("expected valid=true, got %v; errors: %v", val["valid"], val["errors"])
	}
}

func TestChatsDeleteCmd_DryRun(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "DeleteChat",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"chatId": map[string]any{"type": "string"},
				},
				"required": []any{"chatId"},
			},
		},
	}
	ctx, buf, cleanup := setupTestServerWithSchemas(t, nil, schemas)
	defer cleanup()
	ctx.DryRun = true

	cmd := &ChatsDeleteCmd{ChatID: "chat1"}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if result["dry_run"] != true {
		t.Errorf("expected dry_run=true, got %v", result["dry_run"])
	}
	if result["action"] != "chats.delete" {
		t.Errorf("expected action=chats.delete, got %v", result["action"])
	}
	val, ok := result["validation"].(map[string]any)
	if !ok {
		t.Fatal("expected validation object in dry-run output")
	}
	if val["valid"] != true {
		t.Errorf("expected valid=true, got %v; errors: %v", val["valid"], val["errors"])
	}
}

// keys returns the map keys as a sorted string slice for debugging.
func keys(m map[string]any) []string {
	var ks []string
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}
