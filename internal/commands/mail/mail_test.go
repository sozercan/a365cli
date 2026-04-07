package mail

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/testutil"
)

func TestMailSearchCmd_Run(t *testing.T) {
	ctx, buf := testutil.SetupTestServer(t, map[string]string{
		"SearchMessagesQueryParameters": `{"value":[{"id":"msg-001","subject":"Budget Review","receivedDateTime":"2025-01-15T10:00:00Z","from":{"emailAddress":{"name":"Alice","address":"alice@contoso.com"}},"isRead":true}]}`,
	})

	cmd := &MailSearchCmd{Query: `budget`, Max: 20}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	messages, ok := result["messages"]
	if !ok {
		t.Fatalf("expected 'messages' key in output, got: %s", buf.String())
	}
	arr, ok := messages.([]any)
	if !ok {
		t.Fatalf("expected 'messages' to be an array, got: %T", messages)
	}
	if len(arr) != 1 {
		t.Fatalf("expected 1 message, got %d", len(arr))
	}
}

func TestMailGetCmd_Run(t *testing.T) {
	ctx, buf := testutil.SetupTestServer(t, map[string]string{
		"GetMessage": `{"id":"msg-001","subject":"Budget Review","from":{"emailAddress":{"name":"Alice","address":"alice@contoso.com"}},"body":{"content":"Hello"}}`,
	})

	cmd := &MailGetCmd{ID: "msg-001"}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if result["id"] != "msg-001" {
		t.Errorf("expected id=msg-001, got %v", result["id"])
	}
	if result["subject"] != "Budget Review" {
		t.Errorf("expected subject=Budget Review, got %v", result["subject"])
	}
}

func TestMailSendCmd_DryRun(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "SendEmailWithAttachments",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"to":      map[string]any{"type": "array"},
					"subject": map[string]any{"type": "string"},
					"body":    map[string]any{"type": "string"},
				},
				"required": []any{"to", "subject", "body"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &MailSendCmd{
		To:      []string{"bob@contoso.com"},
		Subject: "Test Subject",
		Body:    "Hello Bob",
	}
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
	if result["action"] != "mail.send" {
		t.Errorf("expected action=mail.send, got %v", result["action"])
	}
	val, ok := result["validation"].(map[string]any)
	if !ok {
		t.Fatal("expected validation object in dry-run output")
	}
	if val["valid"] != true {
		t.Errorf("expected valid=true, got %v; errors: %v", val["valid"], val["errors"])
	}
}

func TestMailDeleteCmd_NoInput(t *testing.T) {
	ctx, _ := testutil.SetupTestServer(t, nil)
	ctx.NoInput = true

	cmd := &MailDeleteCmd{ID: "msg-001"}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when NoInput=true and Force=false")
	}
	if !strings.Contains(err.Error(), "without --force") {
		t.Errorf("expected error about --force, got: %v", err)
	}
}
