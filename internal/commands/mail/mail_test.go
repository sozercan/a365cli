package mail

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/sozercan/a365cli/internal/commands"
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

func TestMailSendCmd_DryRunValidatesActualArgsAndRedactsPayload(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "SendEmailWithAttachments",
			InputSchema: mailObjectSchema(map[string]string{
				"to":      "array",
				"cc":      "array",
				"bcc":     "array",
				"subject": "string",
				"body":    "string",
			}),
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &MailSendCmd{
		To:      []string{"alice@contoso.com"},
		CC:      []string{"bob@contoso.com"},
		BCC:     []string{"carol@contoso.com"},
		Subject: "Confidential launch plan",
		Body:    "Sensitive message body that must not appear in dry-run output",
	}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	result := decodeMailDryRunResult(t, buf.Bytes())
	assertMailDryRunValid(t, result)
	if result["action"] != "mail.send" {
		t.Errorf("expected action=mail.send, got %v", result["action"])
	}
	if result["to_count"] != float64(len(cmd.To)) || result["cc_count"] != float64(len(cmd.CC)) || result["bcc_count"] != float64(len(cmd.BCC)) {
		t.Errorf("unexpected recipient counts: %v", result)
	}
	forbidden := append([]string{}, cmd.To...)
	forbidden = append(forbidden, cmd.CC...)
	forbidden = append(forbidden, cmd.BCC...)
	forbidden = append(forbidden, cmd.Subject, cmd.Body)
	assertOutputOmits(t, buf.String(), forbidden...)
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

func TestMailReplyCmd_DryRunUsesActualArgsAndRedactsComment(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "ReplyToMessage",
			InputSchema: mailObjectSchema(map[string]string{
				"id":              "string",
				"comment":         "string",
				"sendImmediately": "boolean",
			}, "id"),
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &MailReplyCmd{ID: "msg-001", Comment: "Private reply text", Send: false}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	result := decodeMailDryRunResult(t, buf.Bytes())
	assertMailDryRunValid(t, result)
	if result["send_immediately"] != false {
		t.Errorf("expected send_immediately=false, got %v", result["send_immediately"])
	}
	assertOutputOmits(t, buf.String(), cmd.Comment)
}

func TestMailUpdateCmd_DryRunValidatesActualArgs(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "UpdateMessage",
			InputSchema: mailObjectSchema(map[string]string{
				"id":          "string",
				"subject":     "string",
				"body":        "string",
				"importance":  "string",
				"categories":  "array",
				"contentType": "string",
				"sensitivity": "string",
			}, "id"),
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &MailUpdateCmd{
		ID:         "msg-001",
		Subject:    "Updated confidential subject",
		Body:       "Updated private body",
		Importance: "High",
		Categories: []string{"Internal", "Review"},
	}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	result := decodeMailDryRunResult(t, buf.Bytes())
	assertMailDryRunValid(t, result)
	if result["categories_count"] != float64(len(cmd.Categories)) {
		t.Errorf("expected categories_count=%d, got %v", len(cmd.Categories), result["categories_count"])
	}
	assertOutputOmits(t, buf.String(), cmd.Subject, cmd.Body)
}

func TestMailUploadCmd_LargeDryRunUsesLargeToolSchemaAndRedactsContent(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "UploadAttachment",
			InputSchema: mailObjectSchema(map[string]string{
				"smallOnly": "boolean",
			}, "smallOnly"),
		},
		{
			Name: "UploadLargeAttachment",
			InputSchema: mailObjectSchema(map[string]string{
				"messageId":     "string",
				"fileName":      "string",
				"contentBase64": "string",
				"contentType":   "string",
			}, "messageId", "fileName", "contentBase64"),
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &MailUploadCmd{
		MessageID:     "msg-001",
		FileName:      "confidential-report.txt",
		ContentBase64: "c2Vuc2l0aXZlLWZpbGUtY29udGVudA==",
		Large:         true,
	}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	assertMailDryRunValid(t, decodeMailDryRunResult(t, buf.Bytes()))
	assertOutputOmits(t, buf.String(), cmd.FileName, cmd.ContentBase64)
}

func TestMailDraftAttachmentsCmd_DryRunRedactsURIs(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "AddDraftAttachments",
			InputSchema: mailObjectSchema(map[string]string{
				"messageId":      "string",
				"attachmentUris": "array",
			}, "messageId", "attachmentUris"),
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &MailDraftAttachCmd{
		MessageID: "msg-001",
		AttachmentUris: []string{
			"https://contoso.sharepoint.com/sites/example/private-file.docx",
			"https://contoso.sharepoint.com/sites/example/private-file-2.pdf",
		},
	}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	result := decodeMailDryRunResult(t, buf.Bytes())
	assertMailDryRunValid(t, result)
	if result["attachment_count"] != float64(len(cmd.AttachmentUris)) {
		t.Errorf("expected attachment_count=%d, got %v", len(cmd.AttachmentUris), result["attachment_count"])
	}
	assertOutputOmits(t, buf.String(), cmd.AttachmentUris...)
}

func TestMailOtherMutationDryRunsValidateActualArgs(t *testing.T) {
	tests := []struct {
		name      string
		toolName  string
		schema    map[string]any
		run       func(*commands.Context) error
		forbidden []string
	}{
		{
			name:     "reply all",
			toolName: "ReplyAllToMessage",
			schema: mailObjectSchema(map[string]string{
				"id":              "string",
				"comment":         "string",
				"sendImmediately": "boolean",
			}, "id"),
			run: func(ctx *commands.Context) error {
				return (&MailReplyAllCmd{ID: "msg-001", Comment: "Private reply-all text", Send: true}).Run(ctx)
			},
			forbidden: []string{"Private reply-all text"},
		},
		{
			name:     "forward",
			toolName: "ForwardMessage",
			schema: mailObjectSchema(map[string]string{
				"messageId":    "string",
				"additionalTo": "array",
				"introComment": "string",
			}, "messageId"),
			run: func(ctx *commands.Context) error {
				return (&MailForwardCmd{ID: "msg-001", To: []string{"alice@contoso.com"}, Comment: "Private forward note"}).Run(ctx)
			},
			forbidden: []string{"alice@contoso.com", "Private forward note"},
		},
		{
			name:     "delete",
			toolName: "DeleteMessage",
			schema: mailObjectSchema(map[string]string{
				"id": "string",
			}, "id"),
			run: func(ctx *commands.Context) error {
				return (&MailDeleteCmd{ID: "msg-001"}).Run(ctx)
			},
		},
		{
			name:     "flag",
			toolName: "FlagEmail",
			schema: mailObjectSchema(map[string]string{
				"messageId":  "string",
				"flagStatus": "string",
			}, "messageId", "flagStatus"),
			run: func(ctx *commands.Context) error {
				return (&MailFlagCmd{ID: "msg-001", Status: "Flagged"}).Run(ctx)
			},
		},
		{
			name:     "create draft",
			toolName: "CreateDraftMessage",
			schema: mailObjectSchema(map[string]string{
				"to":      "array",
				"cc":      "array",
				"subject": "string",
				"body":    "string",
			}),
			run: func(ctx *commands.Context) error {
				return (&MailDraftCmd{
					To:      []string{"alice@contoso.com"},
					CC:      []string{"bob@contoso.com"},
					Subject: "Private draft subject",
					Body:    "Private draft body",
				}).Run(ctx)
			},
			forbidden: []string{"alice@contoso.com", "bob@contoso.com", "Private draft subject", "Private draft body"},
		},
		{
			name:     "send draft",
			toolName: "SendDraftMessage",
			schema: mailObjectSchema(map[string]string{
				"id": "string",
			}, "id"),
			run: func(ctx *commands.Context) error {
				return (&MailSendDraftCmd{ID: "msg-001"}).Run(ctx)
			},
		},
		{
			name:     "delete attachment",
			toolName: "DeleteAttachment",
			schema: mailObjectSchema(map[string]string{
				"messageId":    "string",
				"attachmentId": "string",
			}, "messageId", "attachmentId"),
			run: func(ctx *commands.Context) error {
				return (&MailDeleteAttachCmd{MessageID: "msg-001", AttachmentID: "attachment-001"}).Run(ctx)
			},
		},
		{
			name:     "update draft",
			toolName: "UpdateDraft",
			schema: mailObjectSchema(map[string]string{
				"messageId": "string",
				"to":        "array",
				"cc":        "array",
				"bcc":       "array",
				"subject":   "string",
				"body":      "string",
			}, "messageId"),
			run: func(ctx *commands.Context) error {
				return (&MailUpdateDraftCmd{
					MessageID: "msg-001",
					To:        []string{"alice@contoso.com"},
					CC:        []string{"bob@contoso.com"},
					BCC:       []string{"carol@contoso.com"},
					Subject:   "Private updated draft subject",
					Body:      "Private updated draft body",
				}).Run(ctx)
			},
			forbidden: []string{
				"alice@contoso.com",
				"bob@contoso.com",
				"carol@contoso.com",
				"Private updated draft subject",
				"Private updated draft body",
			},
		},
		{
			name:     "reply with full thread",
			toolName: "ReplyWithFullThread",
			schema: mailObjectSchema(map[string]string{
				"messageId": "string",
			}, "messageId"),
			run: func(ctx *commands.Context) error {
				return (&MailReplyThreadCmd{MessageID: "msg-001"}).Run(ctx)
			},
		},
		{
			name:     "reply all with full thread",
			toolName: "ReplyAllWithFullThread",
			schema: mailObjectSchema(map[string]string{
				"messageId": "string",
			}, "messageId"),
			run: func(ctx *commands.Context) error {
				return (&MailReplyAllThreadCmd{MessageID: "msg-001"}).Run(ctx)
			},
		},
		{
			name:     "forward with full thread",
			toolName: "ForwardMessageWithFullThread",
			schema: mailObjectSchema(map[string]string{
				"messageId": "string",
			}, "messageId"),
			run: func(ctx *commands.Context) error {
				return (&MailForwardThreadCmd{MessageID: "msg-001"}).Run(ctx)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, []mcp.ToolInfo{{
				Name:        tt.toolName,
				InputSchema: tt.schema,
			}})
			ctx.DryRun = true

			if err := tt.run(ctx); err != nil {
				t.Fatalf("Run() error: %v", err)
			}

			assertMailDryRunValid(t, decodeMailDryRunResult(t, buf.Bytes()))
			assertOutputOmits(t, buf.String(), tt.forbidden...)
		})
	}
}

func mailObjectSchema(properties map[string]string, required ...string) map[string]any {
	props := make(map[string]any, len(properties))
	for name, propertyType := range properties {
		property := map[string]any{"type": propertyType}
		if propertyType == "array" {
			property["items"] = map[string]any{"type": "string"}
		}
		props[name] = property
	}

	requiredValues := make([]any, len(required))
	for i, name := range required {
		requiredValues[i] = name
	}

	return map[string]any{
		"type":                 "object",
		"properties":           props,
		"required":             requiredValues,
		"additionalProperties": false,
	}
}

func decodeMailDryRunResult(t *testing.T, data []byte) map[string]any {
	t.Helper()

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, data)
	}
	return result
}

func assertMailDryRunValid(t *testing.T, result map[string]any) {
	t.Helper()

	if result["dry_run"] != true {
		t.Errorf("expected dry_run=true, got %v", result["dry_run"])
	}
	validation, ok := result["validation"].(map[string]any)
	if !ok {
		t.Fatal("expected validation object in dry-run output")
	}
	if validation["valid"] != true {
		t.Fatalf("expected valid=true, got %v; errors: %v", validation["valid"], validation["errors"])
	}
}

func assertOutputOmits(t *testing.T, output string, forbidden ...string) {
	t.Helper()

	for _, value := range forbidden {
		if value != "" && strings.Contains(output, value) {
			t.Errorf("dry-run output leaked %q: %s", value, output)
		}
	}
}
