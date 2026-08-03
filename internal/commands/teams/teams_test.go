package teams

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
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

type recordedToolCall struct {
	Name      string
	Arguments map[string]any
}

// setupTestServerWithSchemas creates a mock MCP server that also responds to
// tools/list requests with the provided tool schemas.
func setupTestServerWithSchemas(t *testing.T, toolResponses map[string]string, toolSchemas []mcp.ToolInfo) (*commands.Context, *bytes.Buffer, func()) {
	t.Helper()
	return setupTestServerWithRecorder(t, toolResponses, toolSchemas, nil)
}

func setupTestServerWithRecorder(t *testing.T, toolResponses map[string]string, toolSchemas []mcp.ToolInfo, calls chan<- recordedToolCall) (*commands.Context, *bytes.Buffer, func()) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		var req struct {
			ID     int             `json:"id"`
			Method string          `json:"method"`
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
			if calls != nil {
				calls <- recordedToolCall{Name: params.Name, Arguments: params.Arguments}
			}

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

	t.Setenv("A365_ENDPOINT", server.URL+"/")

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
		"SendMessageToChat": `{"id":"msg1","chatId":"chat1","content":"hello"}`,
	})
	defer cleanup()

	payload := "confidential-chat-payload"
	cmd := &ChatsSendCmd{ChatID: "chat1", Message: payload}
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
			Name: "SendMessageToChat",
			InputSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
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

	payload := "confidential-chat-payload"
	cmd := &ChatsSendCmd{ChatID: "chat1", Message: payload}
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
	if result["contentLength"] != float64(len(payload)) {
		t.Errorf("expected contentLength=%d, got %v", len(payload), result["contentLength"])
	}
	if strings.Contains(buf.String(), payload) {
		t.Fatal("dry-run output exposed the full chat message")
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
			Name: "SendMessageToChannel",
			InputSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
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

	payload := "confidential-channel-payload"
	cmd := &ChannelsPostCmd{TeamID: "t1", ChannelID: "ch1", Message: payload}
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
	if result["contentLength"] != float64(len(payload)) {
		t.Errorf("expected contentLength=%d, got %v", len(payload), result["contentLength"])
	}
	if strings.Contains(buf.String(), payload) {
		t.Fatal("dry-run output exposed the full channel message")
	}
}

func TestChatsDeleteCmd_DryRun(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "DeleteChat",
			InputSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
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

func TestChatsCreateCmd_DryRunRedactsTopicAndMembers(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "CreateChat",
			InputSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"chatType":     map[string]any{"type": "string"},
					"members_upns": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"topic":        map[string]any{"type": "string"},
				},
				"required": []any{"chatType", "members_upns"},
			},
		},
	}
	ctx, buf, cleanup := setupTestServerWithSchemas(t, nil, schemas)
	defer cleanup()
	ctx.DryRun = true

	topic := "confidential-chat-topic"
	members := []string{"alice@contoso.com", "bob@contoso.com"}
	cmd := &ChatsCreateCmd{Type: "group", Members: members, Topic: topic}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	val, ok := result["validation"].(map[string]any)
	if !ok || val["valid"] != true {
		t.Fatalf("expected valid dry-run validation, got %v", result["validation"])
	}
	if result["topicLength"] != float64(len(topic)) || result["memberCount"] != float64(len(members)) {
		t.Errorf("unexpected safe metadata: %#v", result)
	}
	for _, payload := range append([]string{topic}, members...) {
		if strings.Contains(buf.String(), payload) {
			t.Fatalf("dry-run output exposed chat payload %q", payload)
		}
	}
}

func TestChatsUpdateCmd_DryRunRedactsTopic(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "UpdateChat",
			InputSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"chatId": map[string]any{"type": "string"},
					"topic":  map[string]any{"type": "string"},
				},
				"required": []any{"chatId"},
			},
		},
	}
	ctx, buf, cleanup := setupTestServerWithSchemas(t, nil, schemas)
	defer cleanup()
	ctx.DryRun = true

	topic := "confidential-updated-topic"
	cmd := &ChatsUpdateCmd{ChatID: "19:example@thread.v2", Topic: topic}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if result["topicLength"] != float64(len(topic)) {
		t.Errorf("expected topicLength=%d, got %v", len(topic), result["topicLength"])
	}
	if strings.Contains(buf.String(), topic) {
		t.Fatal("dry-run output exposed the full chat topic")
	}
}

func TestChannelsCreatePrivateCmd_UsesLiveSchemaToolAndArgs(t *testing.T) {
	calls := make(chan recordedToolCall, 1)
	ctx, _, cleanup := setupTestServerWithRecorder(t, map[string]string{
		"CreateChannel": `{"id":"channel-001"}`,
	}, nil, calls)
	defer cleanup()

	cmd := &ChannelsCreatePrivateCmd{
		TeamID:      "00000000-0000-0000-0000-000000000001",
		DisplayName: "Private Project",
		Description: "Restricted collaboration space",
	}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	call := <-calls
	if call.Name != "CreateChannel" {
		t.Fatalf("expected CreateChannel, got %s", call.Name)
	}
	if call.Arguments["membershipType"] != "private" {
		t.Errorf("expected membershipType=private, got %v", call.Arguments["membershipType"])
	}
	if call.Arguments["teamId"] != cmd.TeamID || call.Arguments["displayName"] != cmd.DisplayName {
		t.Errorf("unexpected CreateChannel arguments: %#v", call.Arguments)
	}
}

func TestChatsAddMemberCmd_PreservesRoleArguments(t *testing.T) {
	calls := make(chan recordedToolCall, 1)
	ctx, _, cleanup := setupTestServerWithRecorder(t, map[string]string{
		"AddChatMember": `{"id":"membership-001"}`,
	}, nil, calls)
	defer cleanup()

	cmd := &ChatsAddMemberCmd{
		ChatID: "19:example@thread.v2",
		UPN:    "alice@contoso.com",
		Roles:  []string{"owner"},
	}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	call := <-calls
	if call.Name != "AddChatMember" {
		t.Fatalf("expected AddChatMember, got %s", call.Name)
	}
	if call.Arguments["chatId"] != cmd.ChatID {
		t.Errorf("chatId = %v, want %s", call.Arguments["chatId"], cmd.ChatID)
	}
	roles, ok := call.Arguments["roles"].([]any)
	if !ok || len(roles) != 1 || roles[0] != "owner" {
		t.Fatalf("roles = %#v, want [owner]", call.Arguments["roles"])
	}
	wantBinding := "https://graph.microsoft.com/v1.0/users('alice@contoso.com')"
	if call.Arguments["userodata_bind"] != wantBinding {
		t.Errorf("userodata_bind = %v, want %s", call.Arguments["userodata_bind"], wantBinding)
	}
	if call.Arguments["odata_type"] != "#microsoft.graph.aadUserConversationMember" {
		t.Errorf("odata_type = %v", call.Arguments["odata_type"])
	}
}

// keys returns the map keys as a sorted string slice for debugging.
func keys(m map[string]any) []string {
	var ks []string
	for k := range m {
		ks = append(ks, k)
	}
	slices.Sort(ks)
	return ks
}
