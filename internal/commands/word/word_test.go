package word

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/testutil"
)

func TestWordCreateCmd_DryRun(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "CreateDocument",
			InputSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"fileName":      map[string]any{"type": "string"},
					"contentInHtml": map[string]any{"type": "string"},
					"shareWith":     map[string]any{"type": "string"},
				},
				"required": []any{"fileName"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	payload := "<p>confidential-document-payload</p>"
	cmd := &WordCreateCmd{FileName: "report.docx", ContentInHTML: payload}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if result["dry_run"] != true {
		t.Error("expected dry_run=true")
	}
	val, ok := result["validation"].(map[string]any)
	if !ok {
		t.Fatal("expected validation object")
	}
	if val["valid"] != true {
		t.Errorf("expected valid=true, got %v; errors: %v", val["valid"], val["errors"])
	}
	if result["contentLength"] != float64(len(payload)) {
		t.Errorf("expected contentLength=%d, got %v", len(payload), result["contentLength"])
	}
	if content, ok := result["contentInHtml"]; ok {
		t.Fatalf("dry-run output exposed contentInHtml=%v", content)
	}
}

func TestWordCommentCmd_DryRun(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "AddComment",
			InputSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"driveId":    map[string]any{"type": "string"},
					"documentId": map[string]any{"type": "string"},
					"newComment": map[string]any{"type": "string"},
				},
				"required": []any{"driveId", "documentId", "newComment"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	payload := "confidential-comment-payload"
	cmd := &WordCommentCmd{
		DriveID:    "drive-001",
		DocumentID: "doc-001",
		Text:       payload,
	}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if result["dry_run"] != true {
		t.Error("expected dry_run=true")
	}
	val, ok := result["validation"].(map[string]any)
	if !ok {
		t.Fatal("expected validation object")
	}
	if val["valid"] != true {
		t.Errorf("expected valid=true, got %v; errors: %v", val["valid"], val["errors"])
	}
	if result["commentLength"] != float64(len(payload)) {
		t.Errorf("expected commentLength=%d, got %v", len(payload), result["commentLength"])
	}
	if strings.Contains(buf.String(), payload) {
		t.Fatal("dry-run output exposed the full comment")
	}
}

func TestWordReplyCmd_DryRun(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "ReplyToComment",
			InputSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"commentId":  map[string]any{"type": "string"},
					"driveId":    map[string]any{"type": "string"},
					"documentId": map[string]any{"type": "string"},
					"newComment": map[string]any{"type": "string"},
				},
				"required": []any{"commentId", "driveId", "documentId", "newComment"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	payload := "confidential-reply-payload"
	cmd := &WordReplyCmd{
		CommentID:  "comment-001",
		DriveID:    "drive-001",
		DocumentID: "doc-001",
		Text:       payload,
	}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	val, ok := result["validation"].(map[string]any)
	if !ok {
		t.Fatal("expected validation object")
	}
	if val["valid"] != true {
		t.Errorf("expected valid=true, got %v; errors: %v", val["valid"], val["errors"])
	}
	if result["commentLength"] != float64(len(payload)) {
		t.Errorf("expected commentLength=%d, got %v", len(payload), result["commentLength"])
	}
	if strings.Contains(buf.String(), payload) {
		t.Fatal("dry-run output exposed the full reply")
	}
}
