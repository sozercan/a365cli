package output

import (
	"testing"

	"github.com/sozercan/a365cli/internal/mcp"
)

func TestExtractContent_JSONObject(t *testing.T) {
	resp := &mcp.JSONRPCResponse{
		Result: &mcp.Result{
			Content: []mcp.Content{
				{Type: "text", Text: `{"teams":[{"id":"1","displayName":"Test"}]}`},
			},
		},
	}

	data, err := ExtractContent(resp)
	if err != nil {
		t.Fatalf("ExtractContent failed: %v", err)
	}
	if _, ok := data["teams"]; !ok {
		t.Error("expected 'teams' key in data")
	}
}

func TestExtractContent_JSONArray(t *testing.T) {
	resp := &mcp.JSONRPCResponse{
		Result: &mcp.Result{
			Content: []mcp.Content{
				{Type: "text", Text: `[{"id":"1"},{"id":"2"}]`},
			},
		},
	}

	data, err := ExtractContent(resp)
	if err != nil {
		t.Fatalf("ExtractContent failed: %v", err)
	}
	items, ok := data["items"]
	if !ok {
		t.Fatal("expected 'items' key for array response")
	}
	arr, ok := items.([]any)
	if !ok || len(arr) != 2 {
		t.Errorf("expected 2 items, got %v", items)
	}
}

func TestExtractContent_Error(t *testing.T) {
	resp := &mcp.JSONRPCResponse{
		Error: &mcp.RPCError{Code: -32600, Message: "Invalid"},
	}

	_, err := ExtractContent(resp)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExtractContent_Nil(t *testing.T) {
	_, err := ExtractContent(nil)
	if err == nil {
		t.Fatal("expected error for nil response")
	}
}

func TestExtractContent_EmptyResult(t *testing.T) {
	resp := &mcp.JSONRPCResponse{Result: &mcp.Result{}}
	_, err := ExtractContent(resp)
	if err == nil {
		t.Fatal("expected error for empty result")
	}
}

func TestExtractContent_PlainText(t *testing.T) {
	resp := &mcp.JSONRPCResponse{
		Result: &mcp.Result{
			Content: []mcp.Content{
				{Type: "text", Text: "Message sent successfully."},
			},
		},
	}

	data, err := ExtractContent(resp)
	if err != nil {
		t.Fatalf("ExtractContent failed: %v", err)
	}
	msg, ok := data["message"]
	if !ok || msg != "Message sent successfully." {
		t.Errorf("expected message text, got %v", data)
	}
}

func TestExtractContent_MultipleContent(t *testing.T) {
	resp := &mcp.JSONRPCResponse{
		Result: &mcp.Result{
			Content: []mcp.Content{
				{Type: "text", Text: `{"chats":[{"id":"1"}]}`},
				{Type: "text", Text: "CorrelationId: abc-123"},
			},
		},
	}

	data, err := ExtractContent(resp)
	if err != nil {
		t.Fatalf("ExtractContent failed: %v", err)
	}
	if _, ok := data["chats"]; !ok {
		t.Error("expected 'chats' key — should pick first JSON block")
	}
}

func TestToRows(t *testing.T) {
	data := map[string]any{
		"teams": []any{
			map[string]any{"id": "1", "displayName": "Team A"},
			map[string]any{"id": "2", "displayName": "Team B"},
		},
	}

	rows := ToRows(data, "teams")
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0]["displayName"] != "Team A" {
		t.Errorf("expected 'Team A', got %v", rows[0]["displayName"])
	}
}

func TestToRows_MissingKey(t *testing.T) {
	data := map[string]any{"teams": []any{}}
	rows := ToRows(data, "channels")
	if rows != nil {
		t.Error("expected nil for missing key")
	}
}

func TestExtractContent_EmbeddedJSON(t *testing.T) {
	resp := &mcp.JSONRPCResponse{
		Result: &mcp.Result{
			Content: []mcp.Content{
				{Type: "text", Text: "Events retrieved successfully.\n{\"value\":[{\"id\":\"1\",\"subject\":\"Standup\"}]}"},
			},
		},
	}

	data, err := ExtractContent(resp)
	if err != nil {
		t.Fatalf("ExtractContent failed: %v", err)
	}
	val, ok := data["value"]
	if !ok {
		t.Fatal("expected 'value' key after embedded JSON extraction")
	}
	arr, ok := val.([]any)
	if !ok || len(arr) != 1 {
		t.Fatalf("expected 1 item in value array, got %v", val)
	}
	item, ok := arr[0].(map[string]any)
	if !ok {
		t.Fatal("expected map in array")
	}
	if item["subject"] != "Standup" {
		t.Errorf("expected subject 'Standup', got %v", item["subject"])
	}
}

func TestExtractContent_EmbeddedArray(t *testing.T) {
	resp := &mcp.JSONRPCResponse{
		Result: &mcp.Result{
			Content: []mcp.Content{
				{Type: "text", Text: "Done.\n[{\"id\":\"1\"},{\"id\":\"2\"}]"},
			},
		},
	}

	data, err := ExtractContent(resp)
	if err != nil {
		t.Fatalf("ExtractContent failed: %v", err)
	}
	items, ok := data["items"]
	if !ok {
		t.Fatal("expected 'items' key for embedded array")
	}
	arr, ok := items.([]any)
	if !ok || len(arr) != 2 {
		t.Fatalf("expected 2 items, got %v", items)
	}
}

func TestUnwrapRawResponse(t *testing.T) {
	resp := &mcp.JSONRPCResponse{
		Result: &mcp.Result{
			Content: []mcp.Content{
				{Type: "text", Text: `{"rawResponse":"{\"value\":[{\"id\":\"1\"}]}","message":"ok"}`},
			},
		},
	}

	data, err := ExtractContent(resp)
	if err != nil {
		t.Fatalf("ExtractContent failed: %v", err)
	}
	// Should have unwrapped to the inner value
	val, ok := data["value"]
	if !ok {
		t.Fatal("expected 'value' key from unwrapped rawResponse")
	}
	arr, ok := val.([]any)
	if !ok || len(arr) != 1 {
		t.Fatalf("expected 1 item in value, got %v", val)
	}
}

func TestUnwrapRawResponse_NoRaw(t *testing.T) {
	data := map[string]any{
		"teams": []any{
			map[string]any{"id": "1"},
		},
	}
	result := unwrapRawResponse(data)
	if _, ok := result["teams"]; !ok {
		t.Error("expected data to pass through unchanged when no rawResponse")
	}
}

func TestUnwrapRawResponse_InvalidJSON(t *testing.T) {
	data := map[string]any{
		"rawResponse": "this is not json {{{",
		"message":     "ok",
	}
	result := unwrapRawResponse(data)
	// Should return original data since rawResponse is not valid JSON
	if result["rawResponse"] != "this is not json {{{" {
		t.Error("expected original data to be returned for invalid rawResponse JSON")
	}
	if result["message"] != "ok" {
		t.Error("expected original message to be preserved")
	}
}

func TestUnwrapRawResponse_NonStringRaw(t *testing.T) {
	data := map[string]any{
		"rawResponse": 42,
		"message":     "ok",
	}
	result := unwrapRawResponse(data)
	// Should return original data since rawResponse is not a string
	if result["rawResponse"] != 42 {
		t.Error("expected original data when rawResponse is not a string")
	}
}
