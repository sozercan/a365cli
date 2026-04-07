package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/sozercan/a365cli/internal/mcp"
)

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		format   string
		expected Format
	}{
		{"", FormatHuman},
		{"table", FormatHuman},
		{"json", FormatJSON},
		{"tsv", FormatPlain},
		{"plain", FormatPlain},
	}

	for _, tt := range tests {
		f := NewFormatter(tt.format)
		if f.Format != tt.expected {
			t.Errorf("NewFormatter(%q) = %d, expected %d", tt.format, f.Format, tt.expected)
		}
	}
}

func TestPrintList_Human(t *testing.T) {
	var buf bytes.Buffer
	f := &Formatter{Format: FormatHuman, Writer: &buf}

	columns := []Column{
		{Header: "NAME", Extract: func(r map[string]any) string { return getString(r, "name") }},
		{Header: "ID", Extract: func(r map[string]any) string { return getString(r, "id") }},
	}
	rows := []map[string]any{
		{"name": "Team A", "id": "1"},
		{"name": "Team B", "id": "2"},
	}

	err := f.PrintList("teams", columns, rows)
	if err != nil {
		t.Fatalf("PrintList failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "NAME") {
		t.Error("expected header row")
	}
	if !strings.Contains(out, "Team A") {
		t.Error("expected data row")
	}
}

func TestPrintList_JSON(t *testing.T) {
	var buf bytes.Buffer
	f := &Formatter{Format: FormatJSON, Writer: &buf}

	columns := []Column{
		{Header: "NAME", Extract: func(r map[string]any) string { return getString(r, "name") }},
	}
	rows := []map[string]any{
		{"name": "Team A", "id": "1"},
	}

	err := f.PrintList("teams", columns, rows)
	if err != nil {
		t.Fatalf("PrintList failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if _, ok := parsed["teams"]; !ok {
		t.Error("expected 'teams' key in JSON envelope")
	}
}

func TestPrintList_Plain(t *testing.T) {
	var buf bytes.Buffer
	f := &Formatter{Format: FormatPlain, Writer: &buf}

	columns := []Column{
		{Header: "NAME", Extract: func(r map[string]any) string { return getString(r, "name") }},
		{Header: "ID", Extract: func(r map[string]any) string { return getString(r, "id") }},
	}
	rows := []map[string]any{
		{"name": "Team A", "id": "1"},
	}

	err := f.PrintList("teams", columns, rows)
	if err != nil {
		t.Fatalf("PrintList failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if lines[0] != "NAME\tID" {
		t.Errorf("expected TSV header 'NAME\\tID', got %q", lines[0])
	}
	if lines[1] != "Team A\t1" {
		t.Errorf("expected TSV row 'Team A\\t1', got %q", lines[1])
	}
}

func TestPrintItem_JSON(t *testing.T) {
	var buf bytes.Buffer
	f := &Formatter{Format: FormatJSON, Writer: &buf}

	err := f.PrintItem(map[string]any{"id": "1", "name": "Test"})
	if err != nil {
		t.Fatalf("PrintItem failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if parsed["name"] != "Test" {
		t.Errorf("expected 'Test', got %v", parsed["name"])
	}
}

func TestPrintItem_Human(t *testing.T) {
	var buf bytes.Buffer
	f := &Formatter{Format: FormatHuman, Writer: &buf}

	err := f.PrintItem(map[string]any{"id": "1", "name": "Test"})
	if err != nil {
		t.Fatalf("PrintItem failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "id:") || !strings.Contains(out, "name:") {
		t.Error("expected key-value format")
	}
}

func TestPrintMutation_Human(t *testing.T) {
	var buf bytes.Buffer
	f := &Formatter{Format: FormatHuman, Writer: &buf}

	err := f.PrintMutation("Message sent", map[string]any{"id": "123"})
	if err != nil {
		t.Fatalf("PrintMutation failed: %v", err)
	}

	if strings.TrimSpace(buf.String()) != "Message sent" {
		t.Errorf("expected 'Message sent', got %q", buf.String())
	}
}

func TestPrintMutation_JSON(t *testing.T) {
	var buf bytes.Buffer
	f := &Formatter{Format: FormatJSON, Writer: &buf}

	err := f.PrintMutation("Message sent", map[string]any{"id": "123"})
	if err != nil {
		t.Fatalf("PrintMutation failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if parsed["id"] != "123" {
		t.Errorf("expected id='123', got %v", parsed["id"])
	}
}

func TestPrintDryRun_Human(t *testing.T) {
	var buf bytes.Buffer
	f := &Formatter{Format: FormatHuman, Writer: &buf}

	err := f.PrintDryRun("send message to chat abc", map[string]any{"action": "chats.send"})
	if err != nil {
		t.Fatalf("PrintDryRun failed: %v", err)
	}

	out := strings.TrimSpace(buf.String())
	if out != "Dry run: would send message to chat abc" {
		t.Errorf("expected dry run message, got %q", out)
	}
}

func TestPrintDryRun_JSON(t *testing.T) {
	var buf bytes.Buffer
	f := &Formatter{Format: FormatJSON, Writer: &buf}

	err := f.PrintDryRun("send message", map[string]any{"action": "chats.send"})
	if err != nil {
		t.Fatalf("PrintDryRun failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if parsed["dry_run"] != true {
		t.Error("expected dry_run=true in JSON")
	}
	if parsed["action"] != "chats.send" {
		t.Error("expected action='chats.send'")
	}
}

func TestPrintDryRunValidated_Human_Valid(t *testing.T) {
	var buf bytes.Buffer
	f := &Formatter{Format: FormatHuman, Writer: &buf}

	v := &mcp.ValidationResult{Valid: true}
	err := f.PrintDryRunValidated("send message to chat abc", map[string]any{"action": "chats.send"}, v)
	if err != nil {
		t.Fatalf("PrintDryRunValidated failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Dry run: would send message to chat abc") {
		t.Error("expected dry run message")
	}
	if !strings.Contains(out, "Arguments valid") {
		t.Error("expected valid message")
	}
}

func TestPrintDryRunValidated_Human_Invalid(t *testing.T) {
	var buf bytes.Buffer
	f := &Formatter{Format: FormatHuman, Writer: &buf}

	v := &mcp.ValidationResult{Valid: false, Errors: []string{"/chatId: expected string, got number"}}
	err := f.PrintDryRunValidated("send message", map[string]any{"action": "chats.send"}, v)
	if err == nil {
		t.Fatal("expected error for invalid validation")
	}

	out := buf.String()
	if !strings.Contains(out, "Validation errors") {
		t.Error("expected validation errors header")
	}
	if !strings.Contains(out, "chatId") {
		t.Error("expected chatId in error output")
	}
}

func TestPrintDryRunValidated_Human_Nil(t *testing.T) {
	var buf bytes.Buffer
	f := &Formatter{Format: FormatHuman, Writer: &buf}

	err := f.PrintDryRunValidated("send message", map[string]any{"action": "chats.send"}, nil)
	if err != nil {
		t.Fatalf("PrintDryRunValidated with nil should not error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Validation skipped") {
		t.Error("expected validation skipped warning")
	}
}

func TestPrintDryRunValidated_JSON_Valid(t *testing.T) {
	var buf bytes.Buffer
	f := &Formatter{Format: FormatJSON, Writer: &buf}

	v := &mcp.ValidationResult{Valid: true, Errors: []string{}}
	err := f.PrintDryRunValidated("send message", map[string]any{"action": "chats.send"}, v)
	if err != nil {
		t.Fatalf("PrintDryRunValidated failed: %v", err)
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

func TestPrintDryRunValidated_JSON_Invalid(t *testing.T) {
	var buf bytes.Buffer
	f := &Formatter{Format: FormatJSON, Writer: &buf}

	v := &mcp.ValidationResult{Valid: false, Errors: []string{"missing required: content"}}
	err := f.PrintDryRunValidated("send message", map[string]any{"action": "chats.send"}, v)
	if err == nil {
		t.Fatal("expected error for invalid validation")
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
	errs, ok := val["errors"].([]any)
	if !ok || len(errs) != 1 {
		t.Errorf("expected 1 error, got %v", val["errors"])
	}
}

func TestPrintDryRunValidated_JSON_Nil(t *testing.T) {
	var buf bytes.Buffer
	f := &Formatter{Format: FormatJSON, Writer: &buf}

	err := f.PrintDryRunValidated("send message", map[string]any{"action": "chats.send"}, nil)
	if err != nil {
		t.Fatalf("PrintDryRunValidated with nil should not error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if parsed["validation"] != nil {
		t.Errorf("expected validation=null, got %v", parsed["validation"])
	}
}

func TestPrintRaw_Error(t *testing.T) {
	var buf bytes.Buffer
	f := &Formatter{Format: FormatHuman, Writer: &buf}

	resp := &mcp.JSONRPCResponse{
		Error: &mcp.RPCError{Code: -32600, Message: "Invalid Request"},
	}

	err := f.PrintRaw(resp)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Invalid Request") {
		t.Errorf("expected error message, got: %s", err.Error())
	}
}

func TestPrintRaw_NilResult(t *testing.T) {
	var buf bytes.Buffer
	f := &Formatter{Format: FormatHuman, Writer: &buf}

	resp := &mcp.JSONRPCResponse{JSONRPC: "2.0", ID: 1}

	err := f.PrintRaw(resp)
	if err != nil {
		t.Fatalf("PrintRaw failed: %v", err)
	}
	if !strings.Contains(buf.String(), "(no result)") {
		t.Errorf("expected '(no result)', got %q", buf.String())
	}
}

func TestPrintRaw_WithContent(t *testing.T) {
	var buf bytes.Buffer
	f := &Formatter{Format: FormatHuman, Writer: &buf}

	resp := &mcp.JSONRPCResponse{
		Result: &mcp.Result{
			Content: []mcp.Content{
				{Type: "text", Text: `{"displayName":"Test Team","id":"abc-123"}`},
			},
		},
	}

	err := f.PrintRaw(resp)
	if err != nil {
		t.Fatalf("PrintRaw failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "displayName:") {
		t.Error("expected key-value output with displayName")
	}
	if !strings.Contains(out, "Test Team") {
		t.Error("expected 'Test Team' in output")
	}
}

func TestPrintRaw_WithContentJSON(t *testing.T) {
	var buf bytes.Buffer
	f := &Formatter{Format: FormatJSON, Writer: &buf}

	resp := &mcp.JSONRPCResponse{
		Result: &mcp.Result{
			Content: []mcp.Content{
				{Type: "text", Text: `{"displayName":"Test Team","id":"abc-123"}`},
			},
		},
	}

	err := f.PrintRaw(resp)
	if err != nil {
		t.Fatalf("PrintRaw failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if parsed["displayName"] != "Test Team" {
		t.Errorf("expected 'Test Team', got %v", parsed["displayName"])
	}
}

func TestPrintList_EmptyRows(t *testing.T) {
	tests := []struct {
		name   string
		format Format
	}{
		{"Human", FormatHuman},
		{"JSON", FormatJSON},
		{"Plain", FormatPlain},
	}

	columns := []Column{
		{Header: "NAME", Extract: func(r map[string]any) string { return getString(r, "name") }},
		{Header: "ID", Extract: func(r map[string]any) string { return getString(r, "id") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := &Formatter{Format: tt.format, Writer: &buf}

			err := f.PrintList("teams", columns, []map[string]any{})
			if err != nil {
				t.Fatalf("PrintList failed: %v", err)
			}

			out := buf.String()
			if out == "" {
				t.Error("expected some output even with empty rows")
			}

			switch tt.format {
			case FormatHuman:
				if !strings.Contains(out, "NAME") {
					t.Error("expected header in human output")
				}
			case FormatJSON:
				var parsed map[string]any
				if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
					t.Fatalf("output is not valid JSON: %v", err)
				}
				teams, ok := parsed["teams"]
				if !ok {
					t.Error("expected 'teams' key")
				}
				arr, ok := teams.([]any)
				if !ok || len(arr) != 0 {
					t.Errorf("expected empty array, got %v", teams)
				}
			case FormatPlain:
				if !strings.Contains(out, "NAME\tID") {
					t.Error("expected TSV header")
				}
				lines := strings.Split(strings.TrimSpace(out), "\n")
				if len(lines) != 1 {
					t.Errorf("expected only header line, got %d lines", len(lines))
				}
			}
		})
	}
}
