package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestParseSSE_ToolCall(t *testing.T) {
	sseData := `event: message
data: {"jsonrpc":"2.0","id":2,"result":{"content":[{"type":"text","text":"{\"chats\":[{\"id\":\"1\"}]}"}]}}

`
	resp, err := parseSSE(strings.NewReader(sseData))
	if err != nil {
		t.Fatalf("parseSSE failed: %v", err)
	}
	if resp == nil {
		t.Fatal("parseSSE returned nil")
	}
	if resp.ID != 2 {
		t.Errorf("expected id=2, got %d", resp.ID)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	if resp.Result == nil {
		t.Fatal("result is nil")
	}
	if len(resp.Result.Content) != 1 {
		t.Fatalf("expected 1 content block, got %d", len(resp.Result.Content))
	}
	if resp.Result.Content[0].Type != "text" {
		t.Errorf("expected type=text, got %s", resp.Result.Content[0].Type)
	}
}

func TestParseSSE_ToolsList(t *testing.T) {
	sseData := `event: message
data: {"jsonrpc":"2.0","id":1,"result":{"tools":[{"name":"ListChats","description":"List recent chats"},{"name":"GetChat","description":"Get chat"}]}}

`
	resp, err := parseSSE(strings.NewReader(sseData))
	if err != nil {
		t.Fatalf("parseSSE failed: %v", err)
	}
	if resp.Result == nil || len(resp.Result.Tools) != 2 {
		t.Fatalf("expected 2 tools, got %v", resp.Result)
	}
	if resp.Result.Tools[0].Name != "ListChats" {
		t.Errorf("expected ListChats, got %s", resp.Result.Tools[0].Name)
	}
}

func TestParseSSE_Error(t *testing.T) {
	sseData := `event: message
data: {"jsonrpc":"2.0","id":1,"error":{"code":-32600,"message":"Invalid Request"}}

`
	resp, err := parseSSE(strings.NewReader(sseData))
	if err != nil {
		t.Fatalf("parseSSE failed: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("expected error, got nil")
	}
	if resp.Error.Code != -32600 {
		t.Errorf("expected code -32600, got %d", resp.Error.Code)
	}
	if resp.Error.Message != "Invalid Request" {
		t.Errorf("expected 'Invalid Request', got %s", resp.Error.Message)
	}
}

func TestParseSSE_NoTrailingNewline(t *testing.T) {
	// Some servers may close the stream without a trailing blank line
	sseData := `data: {"jsonrpc":"2.0","id":1,"result":{"content":[{"type":"text","text":"hello"}]}}`
	resp, err := parseSSE(strings.NewReader(sseData))
	if err != nil {
		t.Fatalf("parseSSE failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.Result.Content[0].Text != "hello" {
		t.Errorf("expected 'hello', got %s", resp.Result.Content[0].Text)
	}
}

func TestParseSSE_DataWithoutSpace(t *testing.T) {
	// SSE spec: "data:" without trailing space is valid, value starts at next char
	sseData := "data:{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{\"content\":[{\"type\":\"text\",\"text\":\"no-space\"}]}}\n\n"
	resp, err := parseSSE(strings.NewReader(sseData))
	if err != nil {
		t.Fatalf("parseSSE failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.Result.Content[0].Text != "no-space" {
		t.Errorf("expected 'no-space', got %s", resp.Result.Content[0].Text)
	}
}

func TestParseSSE_EmptyStream(t *testing.T) {
	_, err := parseSSE(strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error for empty stream")
	}
}

func TestRPCError_Error(t *testing.T) {
	e := &RPCError{Code: -32600, Message: "test error"}
	if e.Error() != "test error" {
		t.Errorf("expected 'test error', got '%s'", e.Error())
	}
}

func TestClient_Initialize(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Bearer test-token, got %s", r.Header.Get("Authorization"))
		}

		body, _ := io.ReadAll(r.Body)
		var req JSONRPCRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal request: %v", err)
		}
		if req.Method != "initialize" {
			t.Errorf("expected method 'initialize', got %s", req.Method)
		}

		// Return SSE response with session ID
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Mcp-Session-Id", "test-session-123")
		fmt.Fprintf(w, "event: message\ndata: {\"jsonrpc\":\"2.0\",\"id\":%d,\"result\":{\"protocolVersion\":\"2024-11-05\",\"serverInfo\":{\"name\":\"test\",\"version\":\"1.0\"}}}\n\n", req.ID)
	}))
	defer server.Close()

	client := NewClient(server.URL+"/", func(ctx context.Context) (string, error) {
		return "test-token", nil
	})

	err := client.Initialize(context.Background())
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if client.sessionID != "test-session-123" {
		t.Errorf("expected session ID 'test-session-123', got '%s'", client.sessionID)
	}
}

func TestClient_CallTool(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req JSONRPCRequest
		json.Unmarshal(body, &req) //nolint:errcheck

		if r.Header.Get("Mcp-Session-Id") != "session-abc" {
			t.Errorf("expected session ID 'session-abc', got '%s'", r.Header.Get("Mcp-Session-Id"))
		}

		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "event: message\ndata: {\"jsonrpc\":\"2.0\",\"id\":%d,\"result\":{\"content\":[{\"type\":\"text\",\"text\":\"{\\\"chats\\\":[]}\"}]}}\n\n", req.ID)
	}))
	defer server.Close()

	client := NewClient(server.URL+"/", func(ctx context.Context) (string, error) {
		return "test-token", nil
	})
	client.sessionID = "session-abc"

	resp, err := client.CallTool(context.Background(), "ListChats", nil)
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}

	if resp.Result == nil || len(resp.Result.Content) != 1 {
		t.Fatalf("expected 1 content block, got %v", resp.Result)
	}
}

func TestClient_ListTools(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req JSONRPCRequest
		json.Unmarshal(body, &req) //nolint:errcheck

		if req.Method != "tools/list" {
			t.Errorf("expected method 'tools/list', got %s", req.Method)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "data: {\"jsonrpc\":\"2.0\",\"id\":%d,\"result\":{\"tools\":[{\"name\":\"ListChats\",\"description\":\"List chats\"}]}}\n\n", req.ID)
	}))
	defer server.Close()

	client := NewClient(server.URL+"/", func(ctx context.Context) (string, error) {
		return "token", nil
	})

	resp, err := client.ListTools(context.Background())
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}
	if len(resp.Result.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(resp.Result.Tools))
	}
}

func TestClient_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized")) //nolint:errcheck
	}))
	defer server.Close()

	client := NewClient(server.URL+"/", func(ctx context.Context) (string, error) {
		return "bad-token", nil
	})

	_, err := client.CallTool(context.Background(), "ListChats", nil)
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected error to contain '401', got: %s", err.Error())
	}
}

func TestClient_JSONResponse(t *testing.T) {
	// Some servers may return plain JSON instead of SSE
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req JSONRPCRequest
		json.Unmarshal(body, &req) //nolint:errcheck

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%d,"result":{"content":[{"type":"text","text":"plain json"}]}}`, req.ID)
	}))
	defer server.Close()

	client := NewClient(server.URL+"/", func(ctx context.Context) (string, error) {
		return "token", nil
	})

	resp, err := client.CallTool(context.Background(), "Test", nil)
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}
	if resp.Result.Content[0].Text != "plain json" {
		t.Errorf("expected 'plain json', got '%s'", resp.Result.Content[0].Text)
	}
}

func TestClient_TokenProviderError(t *testing.T) {
	client := NewClient("http://localhost/", func(ctx context.Context) (string, error) {
		return "", fmt.Errorf("token expired")
	})

	_, err := client.CallTool(context.Background(), "ListChats", nil)
	if err == nil {
		t.Fatal("expected error from token provider")
	}
	if !strings.Contains(err.Error(), "token expired") {
		t.Errorf("expected 'token expired' in error, got: %s", err.Error())
	}
}

func TestClient_Verbose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req JSONRPCRequest
		json.Unmarshal(body, &req) //nolint:errcheck
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%d,"result":{"content":[]}}`, req.ID)
	}))
	defer server.Close()

	var logs []string
	client := NewClient(server.URL+"/", func(ctx context.Context) (string, error) {
		return "token", nil
	})
	client.SetVerbose(func(format string, args ...any) {
		logs = append(logs, fmt.Sprintf(format, args...))
	})

	_, err := client.CallTool(context.Background(), "Test", nil)
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}

	if len(logs) < 2 {
		t.Errorf("expected at least 2 verbose log entries, got %d", len(logs))
	}
}

func TestClient_RetryOn502(t *testing.T) {
	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := callCount.Add(1)
		if n == 1 {
			// First call: return 502
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("bad gateway")) //nolint:errcheck
			return
		}
		// Second call: return 200
		body, _ := io.ReadAll(r.Body)
		var req JSONRPCRequest
		json.Unmarshal(body, &req) //nolint:errcheck
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%d,"result":{"content":[{"type":"text","text":"ok"}]}}`, req.ID)
	}))
	defer server.Close()

	var logs []string
	client := NewClient(server.URL+"/", func(ctx context.Context) (string, error) {
		return "token", nil
	})
	client.retryBaseDelay = time.Millisecond // fast retries for tests
	client.SetVerbose(func(format string, args ...any) {
		logs = append(logs, fmt.Sprintf(format, args...))
	})

	resp, err := client.CallTool(context.Background(), "Test", nil)
	if err != nil {
		t.Fatalf("expected success after retry, got error: %v", err)
	}
	if resp.Result == nil || len(resp.Result.Content) != 1 || resp.Result.Content[0].Text != "ok" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if int(callCount.Load()) != 2 {
		t.Errorf("expected 2 HTTP calls, got %d", callCount.Load())
	}

	// Verify retry was logged
	found := false
	for _, log := range logs {
		if strings.Contains(log, "retrying request (attempt 1/2)") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected retry log message, got logs: %v", logs)
	}
}

func TestClient_RetryExhausted(t *testing.T) {
	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("bad gateway")) //nolint:errcheck
	}))
	defer server.Close()

	client := NewClient(server.URL+"/", func(ctx context.Context) (string, error) {
		return "token", nil
	})
	client.retryBaseDelay = time.Millisecond // fast retries for tests

	_, err := client.CallTool(context.Background(), "Test", nil)
	if err == nil {
		t.Fatal("expected error after exhausted retries")
	}
	if !strings.Contains(err.Error(), "502") {
		t.Errorf("expected error to contain '502', got: %s", err.Error())
	}
	// 1 initial + 2 retries = 3 total calls
	if int(callCount.Load()) != 3 {
		t.Errorf("expected 3 HTTP calls (1 + 2 retries), got %d", callCount.Load())
	}
}

func TestClient_NoRetryOn400(t *testing.T) {
	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request")) //nolint:errcheck
	}))
	defer server.Close()

	client := NewClient(server.URL+"/", func(ctx context.Context) (string, error) {
		return "token", nil
	})
	client.retryBaseDelay = time.Millisecond // fast retries for tests

	_, err := client.CallTool(context.Background(), "Test", nil)
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("expected error to contain '400', got: %s", err.Error())
	}
	// Should NOT retry — only 1 call
	if int(callCount.Load()) != 1 {
		t.Errorf("expected exactly 1 HTTP call (no retry for 400), got %d", callCount.Load())
	}
}

func TestClient_Initialize_UsesCachedSession(t *testing.T) {
	withTempHome(t)

	var initCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req JSONRPCRequest
		json.Unmarshal(body, &req) //nolint:errcheck

		if req.Method == "initialize" {
			initCalls.Add(1)
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Mcp-Session-Id", "fresh-session")
			fmt.Fprintf(w, "event: message\ndata: {\"jsonrpc\":\"2.0\",\"id\":%d,\"result\":{\"protocolVersion\":\"2024-11-05\"}}\n\n", req.ID)
			return
		}
	}))
	defer server.Close()

	endpoint := server.URL + "/"
	tokenProvider := func(ctx context.Context) (string, error) { return "token", nil }

	// First Initialize — should call the server.
	c1 := NewClient(endpoint, tokenProvider)
	if err := c1.Initialize(context.Background()); err != nil {
		t.Fatalf("first Initialize failed: %v", err)
	}
	if int(initCalls.Load()) != 1 {
		t.Fatalf("expected 1 init call, got %d", initCalls.Load())
	}
	if c1.sessionID != "fresh-session" {
		t.Fatalf("expected sessionID 'fresh-session', got '%s'", c1.sessionID)
	}

	// Second Initialize (new client, same endpoint) — should use cache.
	c2 := NewClient(endpoint, tokenProvider)
	if err := c2.Initialize(context.Background()); err != nil {
		t.Fatalf("second Initialize failed: %v", err)
	}
	if int(initCalls.Load()) != 1 {
		t.Errorf("expected init call count still 1 (cached), got %d", initCalls.Load())
	}
	if c2.sessionID != "fresh-session" {
		t.Errorf("expected cached sessionID 'fresh-session', got '%s'", c2.sessionID)
	}
}

func TestClient_CallTool_RetriesOnSessionError(t *testing.T) {
	withTempHome(t)

	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req JSONRPCRequest
		json.Unmarshal(body, &req) //nolint:errcheck

		n := callCount.Add(1)

		if req.Method == "initialize" {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Mcp-Session-Id", "new-session")
			fmt.Fprintf(w, "event: message\ndata: {\"jsonrpc\":\"2.0\",\"id\":%d,\"result\":{\"protocolVersion\":\"2024-11-05\"}}\n\n", req.ID)
			return
		}

		// First tools/call: 401 (stale session)
		if n == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("session expired")) //nolint:errcheck
			return
		}

		// After re-init, tools/call succeeds
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%d,"result":{"content":[{"type":"text","text":"ok"}]}}`, req.ID)
	}))
	defer server.Close()

	client := NewClient(server.URL+"/", func(ctx context.Context) (string, error) {
		return "token", nil
	})
	client.retryBaseDelay = time.Millisecond
	client.sessionID = "stale-session"

	resp, err := client.CallTool(context.Background(), "Test", nil)
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}
	if resp.Result == nil || len(resp.Result.Content) != 1 || resp.Result.Content[0].Text != "ok" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if client.sessionID != "new-session" {
		t.Errorf("expected sessionID 'new-session' after re-init, got '%s'", client.sessionID)
	}
}

func TestIsSessionError(t *testing.T) {
	tests := []struct {
		name   string
		resp   *JSONRPCResponse
		err    error
		expect bool
	}{
		{
			name:   "HTTP 401 error",
			err:    fmt.Errorf("HTTP 401: unauthorized"),
			expect: true,
		},
		{
			name:   "HTTP 403 error",
			err:    fmt.Errorf("HTTP 403: forbidden"),
			expect: true,
		},
		{
			name:   "session keyword in error",
			err:    fmt.Errorf("invalid session ID"),
			expect: true,
		},
		{
			name: "RPC error with session message",
			resp: &JSONRPCResponse{
				Error: &RPCError{Code: -32000, Message: "Session not found"},
			},
			expect: true,
		},
		{
			name: "RPC error with invalid message (not session-related)",
			resp: &JSONRPCResponse{
				Error: &RPCError{Code: -32600, Message: "Invalid request"},
			},
			expect: false,
		},
		{
			name: "RPC error with invalid session message",
			resp: &JSONRPCResponse{
				Error: &RPCError{Code: -32000, Message: "Invalid session"},
			},
			expect: true,
		},
		{
			name:   "regular error",
			err:    fmt.Errorf("connection timeout"),
			expect: false,
		},
		{
			name:   "nil everything",
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSessionError(tt.resp, tt.err)
			if got != tt.expect {
				t.Errorf("isSessionError() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestClient_ListToolsCached(t *testing.T) {
	withTempHome(t)

	var callCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req JSONRPCRequest
		json.Unmarshal(body, &req) //nolint:errcheck

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Mcp-Session-Id", "test-session")

		switch req.Method {
		case "initialize":
			fmt.Fprintf(w, "event: message\ndata: %s\n\n",
				fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"result":{"protocolVersion":"2024-11-05","serverInfo":{"name":"test","version":"1.0"}}}`, req.ID))
		case "tools/list":
			atomic.AddInt32(&callCount, 1)
			fmt.Fprintf(w, "event: message\ndata: %s\n\n",
				fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"result":{"tools":[{"name":"ListChats","description":"List chats","inputSchema":{"type":"object","properties":{}}}]}}`, req.ID))
		}
	}))
	defer server.Close()

	client := NewClient(server.URL+"/", func(ctx context.Context) (string, error) {
		return "token", nil
	})

	if err := client.Initialize(context.Background()); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// First call should hit the server.
	tools, err := client.ListToolsCached(context.Background())
	if err != nil {
		t.Fatalf("ListToolsCached (1st): %v", err)
	}
	if len(tools) != 1 || tools[0].Name != "ListChats" {
		t.Fatalf("expected 1 tool 'ListChats', got %v", tools)
	}

	if c := atomic.LoadInt32(&callCount); c != 1 {
		t.Fatalf("expected 1 tools/list call, got %d", c)
	}

	// Second call should use the cache (no additional server call).
	tools2, err := client.ListToolsCached(context.Background())
	if err != nil {
		t.Fatalf("ListToolsCached (2nd): %v", err)
	}
	if len(tools2) != 1 {
		t.Fatalf("expected 1 cached tool, got %d", len(tools2))
	}

	if c := atomic.LoadInt32(&callCount); c != 1 {
		t.Errorf("expected tools/list called only once (cached), got %d", c)
	}
}
