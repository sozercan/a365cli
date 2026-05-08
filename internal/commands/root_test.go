package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/output"
)

func TestConfirm_Force(t *testing.T) {
	ctx := &Context{
		Ctx:   context.Background(),
		Force: true,
	}
	if err := ctx.Confirm("delete something"); err != nil {
		t.Fatalf("Confirm() with Force=true should return nil, got: %v", err)
	}
}

func TestConfirm_NoInput(t *testing.T) {
	ctx := &Context{
		Ctx:     context.Background(),
		NoInput: true,
	}
	err := ctx.Confirm("delete something")
	if err == nil {
		t.Fatal("Confirm() with NoInput=true should return error")
	}
	if got := err.Error(); got != "refusing to delete something without --force (non-interactive)" {
		t.Errorf("unexpected error message: %q", got)
	}
}

func TestNewMCPClient(t *testing.T) {
	tp := func(ctx context.Context) (string, error) {
		return "test-token", nil
	}
	ctx := &Context{
		Ctx:           context.Background(),
		TokenProvider: mcp.TokenProvider(tp),
		Output:        &output.Formatter{Format: output.FormatJSON},
	}
	client := ctx.NewMCPClient("https://example.com/mcp/")
	if client == nil {
		t.Fatal("NewMCPClient() returned nil")
	}
}

func TestNewMCPClient_Verbose(t *testing.T) {
	tp := func(ctx context.Context) (string, error) {
		return "test-token", nil
	}
	ctx := &Context{
		Ctx:           context.Background(),
		TokenProvider: mcp.TokenProvider(tp),
		Output:        &output.Formatter{Format: output.FormatJSON},
		Verbose:       true,
	}
	client := ctx.NewMCPClient("https://example.com/mcp/")
	if client == nil {
		t.Fatal("NewMCPClient() returned nil")
	}
	// The client should have a verbose logger set.
	// We can't directly inspect the private field, but we verify it
	// doesn't panic or error when created with Verbose=true.
}

// setupMockMCPServer creates an httptest server that handles initialize, tools/list,
// and tools/call for ValidateDryRun tests.
func setupMockMCPServer(t *testing.T, toolSchemas []mcp.ToolInfo) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req struct {
			ID     int    `json:"id"`
			Method string `json:"method"`
		}
		json.Unmarshal(body, &req) //nolint:errcheck

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Mcp-Session-Id", "test-session")

		switch req.Method {
		case "initialize":
			fmt.Fprintf(w, "event: message\ndata: %s\n\n",
				mustJSON(map[string]any{
					"jsonrpc": "2.0", "id": req.ID,
					"result": map[string]any{
						"protocolVersion": "2024-11-05",
						"serverInfo":      map[string]any{"name": "test", "version": "1.0"},
					},
				}))
		case "tools/list":
			tools := toolSchemas
			if tools == nil {
				tools = []mcp.ToolInfo{}
			}
			fmt.Fprintf(w, "event: message\ndata: %s\n\n",
				mustJSON(map[string]any{
					"jsonrpc": "2.0", "id": req.ID,
					"result": map[string]any{"tools": tools},
				}))
		default:
			fmt.Fprintf(w, "event: message\ndata: %s\n\n",
				mustJSON(map[string]any{
					"jsonrpc": "2.0", "id": req.ID,
					"error": map[string]any{"code": -32601, "message": "unknown method"},
				}))
		}
	}))
	t.Cleanup(func() { server.Close() })
	return server
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func TestValidateDryRun_ValidArgs(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "SendMessage",
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
	server := setupMockMCPServer(t, schemas)

	var buf bytes.Buffer
	ctx := &Context{
		Ctx:           context.Background(),
		TokenProvider: func(ctx context.Context) (string, error) { return "test-token", nil },
		Output:        &output.Formatter{Format: output.FormatJSON, Writer: &buf},
		DryRun:        true,
	}

	err := ctx.ValidateDryRun(server.URL+"/", "SendMessage", "send message",
		map[string]any{"chatId": "abc", "content": "hello"})
	if err != nil {
		t.Fatalf("ValidateDryRun with valid args should not error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if parsed["dry_run"] != true {
		t.Error("expected dry_run=true")
	}
	val, ok := parsed["validation"].(map[string]any)
	if !ok {
		t.Fatal("expected validation object")
	}
	if val["valid"] != true {
		t.Error("expected valid=true")
	}
}

func TestValidateDryRun_UsesExplicitMCPArgsForValidation(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "SendMessage",
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
	server := setupMockMCPServer(t, schemas)

	var buf bytes.Buffer
	ctx := &Context{
		Ctx:           context.Background(),
		TokenProvider: func(ctx context.Context) (string, error) { return "test-token", nil },
		Output:        &output.Formatter{Format: output.FormatJSON, Writer: &buf},
		DryRun:        true,
	}

	err := ctx.ValidateDryRun(server.URL+"/", "SendMessage", "send message",
		map[string]any{"action": "send-message", "message": "hello"},
		map[string]any{"chatId": "abc", "content": "hello"})
	if err != nil {
		t.Fatalf("ValidateDryRun with explicit valid MCP args should not error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if parsed["dry_run"] != true {
		t.Error("expected dry_run=true")
	}
	if parsed["action"] != "send-message" {
		t.Errorf("expected display action to be preserved, got %v", parsed["action"])
	}
	if _, ok := parsed["chatId"]; ok {
		t.Error("expected raw MCP-only chatId to stay out of dry-run display data")
	}
	val, ok := parsed["validation"].(map[string]any)
	if !ok {
		t.Fatal("expected validation object")
	}
	if val["valid"] != true {
		t.Error("expected valid=true")
	}
}

func TestValidateDryRun_InvalidArgs(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "SendMessage",
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
	server := setupMockMCPServer(t, schemas)

	var buf bytes.Buffer
	ctx := &Context{
		Ctx:           context.Background(),
		TokenProvider: func(ctx context.Context) (string, error) { return "test-token", nil },
		Output:        &output.Formatter{Format: output.FormatJSON, Writer: &buf},
		DryRun:        true,
	}

	err := ctx.ValidateDryRun(server.URL+"/", "SendMessage", "send message",
		map[string]any{"chatId": "abc"}) // missing required "content"
	if err == nil {
		t.Fatal("ValidateDryRun with invalid args should return error")
	}

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	val, ok := parsed["validation"].(map[string]any)
	if !ok {
		t.Fatal("expected validation object")
	}
	if val["valid"] != false {
		t.Error("expected valid=false")
	}
}

func TestValidateDryRun_ServerUnreachable(t *testing.T) {
	var buf bytes.Buffer
	ctx := &Context{
		Ctx:           context.Background(),
		TokenProvider: func(ctx context.Context) (string, error) { return "test-token", nil },
		Output:        &output.Formatter{Format: output.FormatHuman, Writer: &buf},
		DryRun:        true,
	}

	// Use a URL that won't connect.
	err := ctx.ValidateDryRun("http://127.0.0.1:1/", "SendMessage", "send message",
		map[string]any{"chatId": "abc"})
	if err != nil {
		t.Fatalf("ValidateDryRun should degrade gracefully, got error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Dry run: would send message") {
		t.Error("expected dry run message")
	}
	if !strings.Contains(out, "Validation skipped") {
		t.Error("expected validation skipped warning")
	}
}
