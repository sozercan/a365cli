package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/output"
)

// SetupTestServer creates a mock MCP server and returns a wired commands.Context,
// the output buffer, and the server URL. It uses t.Setenv so cleanup is automatic.
// toolResponses maps MCP tool names to the JSON strings the mock should return
// in Content[0].Text.
func SetupTestServer(t *testing.T, toolResponses map[string]string) (*commands.Context, *bytes.Buffer) {
	return SetupTestServerWithSchemas(t, toolResponses, nil)
}

// SetupTestServerWithSchemas creates a mock MCP server that also responds to
// tools/list requests with the provided tool schemas. If toolSchemas is nil,
// tools/list returns an empty list.
func SetupTestServerWithSchemas(t *testing.T, toolResponses map[string]string, toolSchemas []mcp.ToolInfo) (*commands.Context, *bytes.Buffer) {
	t.Helper()
	ctx, buf, _ := setupTestServer(t, toolResponses, toolSchemas, nil)
	return ctx, buf
}

// RecordedCall captures a tools/call request sent to the mock MCP server.
type RecordedCall struct {
	Name      string
	Arguments map[string]any
}

// MCPRecorder records tools/call requests for command tests.
type MCPRecorder struct {
	mu    sync.Mutex
	calls []RecordedCall
}

// SetupTestServerWithRecorder creates a mock MCP server and records tools/call requests.
func SetupTestServerWithRecorder(t *testing.T, toolResponses map[string]string) (*commands.Context, *bytes.Buffer, *MCPRecorder) {
	t.Helper()
	return SetupTestServerWithSchemasAndRecorder(t, toolResponses, nil)
}

// SetupTestServerWithSchemasAndRecorder creates a mock MCP server with schemas
// and records tools/call requests.
func SetupTestServerWithSchemasAndRecorder(t *testing.T, toolResponses map[string]string, toolSchemas []mcp.ToolInfo) (*commands.Context, *bytes.Buffer, *MCPRecorder) {
	t.Helper()
	recorder := &MCPRecorder{}
	return setupTestServer(t, toolResponses, toolSchemas, recorder)
}

// Calls returns a snapshot of all recorded tools/call requests.
func (r *MCPRecorder) Calls() []RecordedCall {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	calls := make([]RecordedCall, len(r.calls))
	for i, call := range r.calls {
		calls[i] = RecordedCall{Name: call.Name, Arguments: cloneMap(call.Arguments)}
	}
	return calls
}

// LastCall returns the most recent recorded tools/call request.
func (r *MCPRecorder) LastCall() (RecordedCall, bool) {
	calls := r.Calls()
	if len(calls) == 0 {
		return RecordedCall{}, false
	}
	return calls[len(calls)-1], true
}

// Count returns how many tools/call requests matched a tool name.
func (r *MCPRecorder) Count(name string) int {
	count := 0
	for _, call := range r.Calls() {
		if call.Name == name {
			count++
		}
	}
	return count
}

func (r *MCPRecorder) record(name string, arguments map[string]any) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, RecordedCall{Name: name, Arguments: cloneMap(arguments)})
}

func setupTestServer(t *testing.T, toolResponses map[string]string, toolSchemas []mcp.ToolInfo, recorder *MCPRecorder) (*commands.Context, *bytes.Buffer, *MCPRecorder) {
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
			recorder.record(params.Name, params.Arguments)

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
		case "tools/list":
			tools := toolSchemas
			if tools == nil {
				tools = []mcp.ToolInfo{}
			}
			fmt.Fprintf(w, "event: message\ndata: %s\n\n",
				MustJSON(map[string]any{
					"jsonrpc": "2.0",
					"id":      req.ID,
					"result": map[string]any{
						"tools": tools,
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
		UserUPN: "user@contoso.com",
	}

	return ctx, &buf, recorder
}

func cloneMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	b, err := json.Marshal(in)
	if err != nil {
		out := make(map[string]any, len(in))
		for k, v := range in {
			out[k] = v
		}
		return out
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		out = make(map[string]any, len(in))
		for k, v := range in {
			out[k] = v
		}
	}
	return out
}

// MustJSON marshals v to a JSON string, panicking on error.
func MustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}
