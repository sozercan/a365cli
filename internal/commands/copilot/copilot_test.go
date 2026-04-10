package copilot

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/output"
	"github.com/sozercan/a365cli/internal/testutil"
)

func TestCopilotChatCmd_Run(t *testing.T) {
	ctx, buf := testutil.SetupTestServer(t, map[string]string{
		copilotChatTool: `{"message":"Quarterly summary","conversationId":"conv-123"}`,
	})

	cmd := &CopilotChatCmd{Message: "Summarize my week"}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if result["message"] != "Quarterly summary" {
		t.Fatalf("expected message to round-trip, got %v", result["message"])
	}
	if result["conversationId"] != "conv-123" {
		t.Fatalf("expected conversationId to round-trip, got %v", result["conversationId"])
	}
}

func TestPrintCopilotResponse_Human(t *testing.T) {
	var buf bytes.Buffer
	ctx := &commands.Context{
		Ctx:    context.Background(),
		Output: &output.Formatter{Format: output.FormatHuman, Writer: &buf},
	}

	err := printCopilotResponse(ctx, map[string]any{
		"message":        "Here is the answer",
		"conversationId": "conv-123",
		"references":     []any{"doc-1"},
	})
	if err != nil {
		t.Fatalf("printCopilotResponse() error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Here is the answer") {
		t.Fatalf("expected human message output, got %q", out)
	}
	if strings.Contains(out, "conversationId") {
		t.Fatalf("expected conversationId to be hidden from human output body, got %q", out)
	}
	if !strings.Contains(out, "Copilot: Here is the answer") {
		t.Fatalf("expected Copilot-prefixed output, got %q", out)
	}
	if !strings.Contains(out, "references:") || !strings.Contains(out, "doc-1") {
		t.Fatalf("expected extra metadata to be preserved, got %q", out)
	}
}

func TestPrintCopilotResponse_PlainUsesChatStyle(t *testing.T) {
	var buf bytes.Buffer
	ctx := &commands.Context{
		Ctx:    context.Background(),
		Output: &output.Formatter{Format: output.FormatPlain, Writer: &buf},
	}

	err := printCopilotResponse(ctx, map[string]any{
		"message":        "Here is the answer",
		"conversationId": "conv-123",
	})
	if err != nil {
		t.Fatalf("printCopilotResponse() error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Copilot: Here is the answer") {
		t.Fatalf("expected Copilot-prefixed plain output, got %q", out)
	}
	if !strings.Contains(out, "conversationId:") || !strings.Contains(out, "conv-123") {
		t.Fatalf("expected conversationId to remain visible in plain output, got %q", out)
	}
}

func TestPrintCopilotResponse_ConversationPayload(t *testing.T) {
	var buf bytes.Buffer
	ctx := &commands.Context{
		Ctx:    context.Background(),
		Output: &output.Formatter{Format: output.FormatHuman, Writer: &buf},
	}

	err := printCopilotResponse(ctx, map[string]any{
		"@odata.context": "https://graph.microsoft.com/beta/$metadata#microsoft.graph.copilotConversation",
		"id":             "conv-123",
		"displayName":    "hello",
		"messages": []any{
			map[string]any{"text": "hello"},
			map[string]any{"text": "Hello, Sertac"},
		},
		"state":     "active",
		"turnCount": float64(1),
	})
	if err != nil {
		t.Fatalf("printCopilotResponse() error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Copilot: Hello, Sertac") {
		t.Fatalf("expected latest conversation turn to be rendered, got %q", out)
	}
	if strings.Contains(out, "turnCount") || strings.Contains(out, "@odata.context") || strings.Contains(out, "conv-123") {
		t.Fatalf("expected conversation metadata to be hidden, got %q", out)
	}
}

func TestNormalizeCopilotResponse(t *testing.T) {
	data := normalizeCopilotResponse(map[string]any{
		"id": "conv-123",
		"messages": []any{
			map[string]any{"text": "hello"},
			map[string]any{"text": "Hello, Sertac"},
		},
	}, "conv-123")

	if data["message"] != "Hello, Sertac" {
		t.Fatalf("expected normalized message, got %v", data["message"])
	}
	if data["conversationId"] != "conv-123" {
		t.Fatalf("expected normalized conversationId, got %v", data["conversationId"])
	}
	if _, ok := data["messages"]; ok {
		t.Fatalf("expected normalized payload to hide raw messages, got %v", data["messages"])
	}
}

func TestRunInteractiveLoop_ReusesConversationID(t *testing.T) {
	var calls []map[string]any

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
		w.Header().Set("Mcp-Session-Id", "test-session-id")

		switch req.Method {
		case "initialize":
			io.WriteString(w, "event: message\ndata: "+testutil.MustJSON(map[string]any{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"result": map[string]any{
					"protocolVersion": "2024-11-05",
					"serverInfo":      map[string]any{"name": "test", "version": "1.0"},
				},
			})+"\n\n")
		case "tools/call":
			var params struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments"`
			}
			json.Unmarshal(req.Params, &params) //nolint:errcheck
			calls = append(calls, params.Arguments)

			respText := `{"id":"conv-123","messages":[{"text":"first"},{"text":"first answer"}],"turnCount":1,"state":"active"}`
			if len(calls) == 2 {
				respText = `{"id":"conv-123","messages":[{"text":"first"},{"text":"first answer"},{"text":"second"},{"text":"second answer"}],"turnCount":2,"state":"active"}`
			}

			io.WriteString(w, "event: message\ndata: "+testutil.MustJSON(map[string]any{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"result": map[string]any{
					"content": []map[string]any{{"type": "text", "text": respText}},
				},
			})+"\n\n")
		default:
			http.Error(w, "unknown method", http.StatusBadRequest)
		}
	}))
	t.Cleanup(func() { server.Close() })
	t.Setenv("A365_ENDPOINT", server.URL+"/")

	var out bytes.Buffer
	var prompt bytes.Buffer
	ctx := &commands.Context{
		Ctx: context.Background(),
		TokenProvider: func(context.Context) (string, error) {
			return "test-token", nil
		},
		Output: &output.Formatter{Format: output.FormatHuman, Writer: &out},
	}

	err := runInteractiveLoop(ctx, strings.NewReader("first\nsecond\nquit\n"), &prompt, "")
	if err != nil {
		t.Fatalf("runInteractiveLoop() error: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 Copilot calls, got %d", len(calls))
	}
	if _, ok := calls[0]["conversationId"]; ok {
		t.Fatalf("expected first call to start without a conversation ID, got %v", calls[0]["conversationId"])
	}
	if calls[1]["conversationId"] != "conv-123" {
		t.Fatalf("expected second call to reuse conversationId, got %v", calls[1]["conversationId"])
	}

	rendered := out.String()
	if !strings.Contains(rendered, "first answer") || !strings.Contains(rendered, "second answer") {
		t.Fatalf("expected both answers in output, got %q", rendered)
	}
	if strings.Contains(prompt.String(), "Conversation ID:") {
		t.Fatalf("expected interactive prompt output to avoid internal conversation IDs, got %q", prompt.String())
	}
}

func TestRunInteractiveLoop_ReturnsErrorOnEOFFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		var req struct {
			ID int `json:"id"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Mcp-Session-Id", "test-session-id")

		if bytes.Contains(body, []byte(`"method":"initialize"`)) {
			io.WriteString(w, "event: message\ndata: "+testutil.MustJSON(map[string]any{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"result": map[string]any{
					"protocolVersion": "2024-11-05",
					"serverInfo":      map[string]any{"name": "test", "version": "1.0"},
				},
			})+"\n\n")
			return
		}

		io.WriteString(w, "event: message\ndata: "+testutil.MustJSON(map[string]any{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"error":   map[string]any{"code": -32000, "message": "upstream unavailable"},
		})+"\n\n")
	}))
	t.Cleanup(func() { server.Close() })
	t.Setenv("A365_ENDPOINT", server.URL+"/")

	var out bytes.Buffer
	var prompt bytes.Buffer
	ctx := &commands.Context{
		Ctx: context.Background(),
		TokenProvider: func(context.Context) (string, error) {
			return "test-token", nil
		},
		Output: &output.Formatter{Format: output.FormatHuman, Writer: &out},
	}

	err := runInteractiveLoop(ctx, strings.NewReader("first"), &prompt, "")
	if err == nil {
		t.Fatal("expected EOF request failure to return an error")
	}
	if !strings.Contains(err.Error(), "MCP error -32000: upstream unavailable") {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(prompt.String(), "Error: MCP error -32000: upstream unavailable") {
		t.Fatalf("expected prompt to surface the error, got %q", prompt.String())
	}
}
