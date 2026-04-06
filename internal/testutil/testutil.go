package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/output"
)

// SetupTestServer creates a mock MCP server and returns a wired commands.Context,
// the output buffer, and the server URL. It uses t.Setenv so cleanup is automatic.
// toolResponses maps MCP tool names to the JSON strings the mock should return
// in Content[0].Text.
func SetupTestServer(t *testing.T, toolResponses map[string]string) (*commands.Context, *bytes.Buffer) {
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
				MustJSON(map[string]any{
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
				MustJSON(map[string]any{
					"jsonrpc": "2.0",
					"id":      req.ID,
					"result": map[string]any{
						"content": []map[string]any{
							{"type": "text", "text": respText},
						},
					},
				}))
		default:
			fmt.Fprintf(w, "event: message\ndata: %s\n\n",
				MustJSON(map[string]any{
					"jsonrpc": "2.0",
					"id":      req.ID,
					"error":   map[string]any{"code": -32601, "message": "unknown method"},
				}))
		}
	}))

	t.Cleanup(func() { server.Close() })
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

	return ctx, &buf
}

// MustJSON marshals v to a JSON string, panicking on error.
func MustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}
