package mcp

import (
	"strings"
	"testing"
)

func textResult(isError bool, texts ...string) *JSONRPCResponse {
	content := make([]Content, 0, len(texts))
	for _, text := range texts {
		content = append(content, Content{Type: "text", Text: text})
	}
	return &JSONRPCResponse{Result: &Result{Content: content, IsError: isError}}
}

func TestToolCallError(t *testing.T) {
	tests := []struct {
		name    string
		resp    *JSONRPCResponse
		wantErr string // empty means no error expected
	}{
		{name: "nil response", resp: nil},
		{name: "nil result", resp: &JSONRPCResponse{}},
		{name: "json success", resp: textResult(false, `{"teams":[]}`)},
		{name: "plain text success", resp: textResult(false, "Message sent successfully.")},
		{
			name:    "error executing tool text",
			resp:    textResult(false, "Error: Error executing tool: Unknown argument: 'userId' for tool 'ListTeams'.\r\nCorrelationId: 00000000-0000-0000-0000-000000000000\nCorrelationId: 00000000-0000-0000-0000-000000000000"),
			wantErr: "Error executing tool: Unknown argument: 'userId' for tool 'ListTeams'.\nCorrelationId: 00000000-0000-0000-0000-000000000000",
		},
		{
			name:    "error executing tool without prefix",
			resp:    textResult(false, "Error executing tool: upstream unavailable."),
			wantErr: "Error executing tool: upstream unavailable.",
		},
		{
			name:    "isError flag",
			resp:    textResult(true, "Mailbox not found."),
			wantErr: "Mailbox not found.",
		},
		{
			name:    "isError flag without content",
			resp:    textResult(true),
			wantErr: "tool ListTeams reported an error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToolCallError("ListTeams", tt.resp)
			if tt.wantErr == "" {
				if got != nil {
					t.Fatalf("expected no tool error, got %q", got.Error())
				}
				return
			}
			if got == nil {
				t.Fatalf("expected tool error %q, got nil", tt.wantErr)
			}
			if got.Error() != tt.wantErr {
				t.Fatalf("error = %q, want %q", got.Error(), tt.wantErr)
			}
			if got.Tool != "ListTeams" {
				t.Errorf("Tool = %q, want ListTeams", got.Tool)
			}
		})
	}
}

func TestSanitizeToolErrorMessage_KeepsNonLeadingErrorPrefix(t *testing.T) {
	got := SanitizeToolErrorMessage("Error: first\nError: second")
	if !strings.HasPrefix(got, "first\nError: second") {
		t.Fatalf("unexpected sanitized message %q", got)
	}
}
