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

func TestNormalizeAvailableAgents_DedupesAndAnnotatesSelectors(t *testing.T) {
	agents := normalizeAvailableAgents(map[string]any{
		"availableAgents": []any{
			map[string]any{"titleId": "title-alpha", "name": "Alpha Agent", "selector": "shared", "description": "alpha"},
			map[string]any{"titleId": "title-alpha", "name": "Alpha Agent Duplicate", "selector": "shared", "description": "duplicate"},
			map[string]any{"titleId": "title-beta", "titleName": "Beta Agent", "agentId": "shared"},
			map[string]any{"titleId": "title-gamma", "name": "Gamma Agent", "selector": "gamma"},
			map[string]any{"titleId": "title-missing", "name": "Missing Selector"},
			map[string]any{"name": "No Title", "selector": "delta"},
		},
	})

	if len(agents) != 5 {
		t.Fatalf("expected 5 normalized agents after dedupe, got %d", len(agents))
	}

	wants := []struct {
		name      string
		selector  string
		titleID   string
		shared    bool
		shareSize int
	}{
		{name: "Alpha Agent", selector: "shared", titleID: "title-alpha", shared: true, shareSize: 2},
		{name: "Beta Agent", selector: "shared", titleID: "title-beta", shared: true, shareSize: 2},
		{name: "Gamma Agent", selector: "gamma", titleID: "title-gamma", shared: false, shareSize: 1},
		{name: "Missing Selector", selector: "", titleID: "title-missing", shared: false, shareSize: 0},
		{name: "No Title", selector: "delta", titleID: "", shared: false, shareSize: 1},
	}

	for i, want := range wants {
		got := agents[i]
		if got.Name != want.name {
			t.Fatalf("agent[%d] name = %q, want %q", i, got.Name, want.name)
		}
		if got.Selector != want.selector {
			t.Fatalf("agent[%d] selector = %q, want %q", i, got.Selector, want.selector)
		}
		if got.TitleID != want.titleID {
			t.Fatalf("agent[%d] titleID = %q, want %q", i, got.TitleID, want.titleID)
		}
		if got.SharedSelector != want.shared {
			t.Fatalf("agent[%d] sharedSelector = %v, want %v", i, got.SharedSelector, want.shared)
		}
		if got.SharedSelectorCount != want.shareSize {
			t.Fatalf("agent[%d] sharedSelectorCount = %d, want %d", i, got.SharedSelectorCount, want.shareSize)
		}
	}

	if agents[0].Description != "alpha" {
		t.Fatalf("expected duplicate titleId to keep the first row, got description %q", agents[0].Description)
	}
	if agents[1].TitleName != "Beta Agent" {
		t.Fatalf("expected titleName fallback to be preserved, got %q", agents[1].TitleName)
	}
}

func TestResolveAgent_SuccessCases(t *testing.T) {
	agents := []agentInfo{
		{Name: "Budget Bot", Selector: "budget", TitleID: "title-budget"},
		{Name: "Case Match", Selector: "case-sel", TitleID: "title-case"},
		{Name: "Project Pilot", Selector: "project-pilot", TitleID: "title-project"},
		{Name: "Title Prefix", Selector: "selector-123", TitleID: "tp-001"},
		{Name: "Shared One", Selector: "shared", TitleID: "title-shared-1", SharedSelector: true, SharedSelectorCount: 2},
		{Name: "Shared Two", Selector: "shared", TitleID: "title-shared-2", SharedSelector: true, SharedSelectorCount: 2},
	}

	tests := []struct {
		name         string
		query        string
		wantSelector string
		wantTitleID  string
	}{
		{name: "exact name", query: "Budget Bot", wantSelector: "budget", wantTitleID: "title-budget"},
		{name: "exact case-insensitive name", query: "case match", wantSelector: "case-sel", wantTitleID: "title-case"},
		{name: "exact selector", query: "budget", wantSelector: "budget", wantTitleID: "title-budget"},
		{name: "exact title id", query: "tp-001", wantSelector: "selector-123", wantTitleID: "tp-001"},
		{name: "selector prefix", query: "proj", wantSelector: "project-pilot", wantTitleID: "title-project"},
		{name: "title id prefix", query: "tp-", wantSelector: "selector-123", wantTitleID: "tp-001"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveAgent(agents, tc.query)
			if err != nil {
				t.Fatalf("resolveAgent(%q) error: %v", tc.query, err)
			}
			if got.Selector != tc.wantSelector {
				t.Fatalf("resolveAgent(%q) selector = %q, want %q", tc.query, got.Selector, tc.wantSelector)
			}
			if got.TitleID != tc.wantTitleID {
				t.Fatalf("resolveAgent(%q) titleID = %q, want %q", tc.query, got.TitleID, tc.wantTitleID)
			}
		})
	}
}

func TestResolveAgent_ErrorCases(t *testing.T) {
	agents := []agentInfo{
		{Name: "Budget Bot", Selector: "budget", TitleID: "title-budget"},
		{Name: "Missing Selector", TitleID: "title-missing"},
		{Name: "Shared One", Selector: "shared", TitleID: "title-shared-1", SharedSelector: true, SharedSelectorCount: 2},
		{Name: "Shared Two", Selector: "shared", TitleID: "title-shared-2", SharedSelector: true, SharedSelectorCount: 2},
		{Name: "Duplicate Name", Selector: "dup-a", TitleID: "title-dup-a"},
		{Name: "Duplicate Name", Selector: "dup-b", TitleID: "title-dup-b"},
	}

	tests := []struct {
		name     string
		query    string
		contains []string
	}{
		{
			name:     "ambiguous exact name",
			query:    "Duplicate Name",
			contains: []string{"ambiguous", "Duplicate Name [dup-a]", "Duplicate Name [dup-b]"},
		},
		{
			name:     "shared selector rejected",
			query:    "shared",
			contains: []string{"shared selector \"shared\"", "Shared One, Shared Two"},
		},
		{
			name:     "missing selector rejected",
			query:    "Missing Selector",
			contains: []string{"does not expose a usable chat selector"},
		},
		{
			name:     "unknown agent suggests matches",
			query:    "udget",
			contains: []string{"unknown Copilot agent \"udget\"", "Suggestions: Budget Bot [budget]"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := resolveAgent(agents, tc.query)
			if err == nil {
				t.Fatalf("resolveAgent(%q) expected error", tc.query)
			}
			for _, want := range tc.contains {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("resolveAgent(%q) error = %q, want substring %q", tc.query, err.Error(), want)
				}
			}
		})
	}
}

func TestCopilotAgentsCmd_Run_JSON(t *testing.T) {
	ctx, buf := testutil.SetupTestServer(t, map[string]string{
		copilotAvailableAgentsTool: testutil.MustJSON(map[string]any{
			"availableAgents": []map[string]any{
				{"titleId": "title-alpha", "name": "Alpha Agent", "selector": "shared"},
				{"titleId": "title-beta", "titleName": "Beta Agent", "agentId": "shared"},
				{"titleId": "title-gamma", "name": "Gamma Agent", "selector": "gamma", "developerName": "Contoso", "type": "bot"},
				{"titleId": "title-missing", "name": "Missing Selector"},
			},
		}),
	})

	cmd := &CopilotAgentsCmd{}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	rows := mustAgentRows(t, buf)
	if len(rows) != 4 {
		t.Fatalf("expected 4 agents in output, got %d", len(rows))
	}

	shared := mustFindAgentRow(t, rows, "title-alpha")
	if shared["status"] != "shared" {
		t.Fatalf("expected shared agent status=shared, got %v", shared["status"])
	}
	if shared["targetable"] != false {
		t.Fatalf("expected shared agent targetable=false, got %v", shared["targetable"])
	}
	if shared["sharedSelector"] != true {
		t.Fatalf("expected sharedSelector=true, got %v", shared["sharedSelector"])
	}
	if shared["sharedSelectorCount"] != float64(2) {
		t.Fatalf("expected sharedSelectorCount=2, got %v", shared["sharedSelectorCount"])
	}
	if shared["agentId"] != "shared" {
		t.Fatalf("expected agentId to mirror selector, got %v", shared["agentId"])
	}

	unique := mustFindAgentRow(t, rows, "title-gamma")
	if unique["status"] != "ok" {
		t.Fatalf("expected unique agent status=ok, got %v", unique["status"])
	}
	if unique["targetable"] != true {
		t.Fatalf("expected unique agent targetable=true, got %v", unique["targetable"])
	}
	if unique["sharedSelector"] != false {
		t.Fatalf("expected unique agent sharedSelector=false, got %v", unique["sharedSelector"])
	}
	if unique["sharedSelectorCount"] != float64(1) {
		t.Fatalf("expected unique agent sharedSelectorCount=1, got %v", unique["sharedSelectorCount"])
	}
	if unique["developerName"] != "Contoso" {
		t.Fatalf("expected developerName to round-trip, got %v", unique["developerName"])
	}
	if unique["type"] != "bot" {
		t.Fatalf("expected type to round-trip, got %v", unique["type"])
	}

	missing := mustFindAgentRow(t, rows, "title-missing")
	if missing["status"] != "missing" {
		t.Fatalf("expected missing-selector agent status=missing, got %v", missing["status"])
	}
	if missing["targetable"] != false {
		t.Fatalf("expected missing-selector agent targetable=false, got %v", missing["targetable"])
	}
	if missing["selector"] != "" {
		t.Fatalf("expected missing-selector agent selector to be empty, got %v", missing["selector"])
	}
}

func TestCopilotChatCmd_Run_ResolvesAgentAndPassesAgentID(t *testing.T) {
	var agentCalls int
	var chatCalls int
	var chatArgs map[string]any

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

			var respText string
			switch params.Name {
			case copilotAvailableAgentsTool:
				agentCalls++
				respText = testutil.MustJSON(map[string]any{
					"availableAgents": []map[string]any{
						{"titleId": "title-budget", "name": "Budget Bot", "selector": "budget"},
					},
				})
			case copilotChatTool:
				chatCalls++
				chatArgs = params.Arguments
				respText = `{"message":"Quarterly summary","conversationId":"conv-123"}`
			default:
				http.Error(w, "unknown tool", http.StatusBadRequest)
				return
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

	var buf bytes.Buffer
	ctx := &commands.Context{
		Ctx: context.Background(),
		TokenProvider: func(context.Context) (string, error) {
			return "test-token", nil
		},
		Output: &output.Formatter{Format: output.FormatJSON, Writer: &buf},
	}

	cmd := &CopilotChatCmd{Message: "Summarize my week", Agent: "Budget Bot"}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if agentCalls != 1 {
		t.Fatalf("expected 1 agent lookup call, got %d", agentCalls)
	}
	if chatCalls != 1 {
		t.Fatalf("expected 1 Copilot chat call, got %d", chatCalls)
	}
	if chatArgs["agentId"] != "budget" {
		t.Fatalf("expected chat call to include agentId=budget, got %v", chatArgs["agentId"])
	}
	if enabled, ok := chatArgs["enableWebSearch"].(bool); !ok || !enabled {
		t.Fatalf("expected chat call to enable web search by default, got %v", chatArgs["enableWebSearch"])
	}
	if chatArgs["message"] != "Summarize my week" {
		t.Fatalf("expected chat call to include message, got %v", chatArgs["message"])
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

func mustAgentRows(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}

	rawRows, ok := result["agents"].([]any)
	if !ok {
		t.Fatalf("expected top-level 'agents' array, got %T", result["agents"])
	}

	rows := make([]map[string]any, 0, len(rawRows))
	for i, raw := range rawRows {
		row, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("agents[%d] has type %T, want object", i, raw)
		}
		rows = append(rows, row)
	}
	return rows
}

func mustFindAgentRow(t *testing.T, rows []map[string]any, titleID string) map[string]any {
	t.Helper()
	for _, row := range rows {
		if row["titleId"] == titleID {
			return row
		}
	}
	t.Fatalf("no agent row found for titleId %q", titleID)
	return nil
}
