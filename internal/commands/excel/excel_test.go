package excel

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/testutil"
)

func assertExcelDryRunValid(t *testing.T, output []byte) map[string]any {
	t.Helper()

	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, output)
	}
	if result["dry_run"] != true {
		t.Error("expected dry_run=true")
	}
	validation, ok := result["validation"].(map[string]any)
	if !ok {
		t.Fatal("expected validation object")
	}
	if validation["valid"] != true {
		t.Errorf("expected valid=true, got %v; errors: %v", validation["valid"], validation["errors"])
	}
	return result
}

func TestExcelCreateCmd_DryRunUsesActualArgs(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "CreateWorkbook",
			InputSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"fileName":   map[string]any{"type": "string"},
					"csvContent": map[string]any{"type": "string"},
					"shareWith":  map[string]any{"type": "string"},
				},
				"required": []any{"fileName", "csvContent"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	csvContent := "category,amount\nconfidential,100"
	cmd := &ExcelCreateCmd{FileName: "budget.xlsx", CSVContent: csvContent}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	result := assertExcelDryRunValid(t, buf.Bytes())
	if result["action"] != "excel.create" {
		t.Errorf("expected action=excel.create, got %v", result["action"])
	}
	if strings.Contains(buf.String(), csvContent) {
		t.Fatal("dry-run output leaked CSV content")
	}
	if _, ok := result["csvContent"]; ok {
		t.Fatal("dry-run display data should not include csvContent")
	}
	if result["csvBytes"] != float64(len(csvContent)) {
		t.Errorf("expected csvBytes=%d, got %v", len(csvContent), result["csvBytes"])
	}
	if result["rowCount"] != float64(2) {
		t.Errorf("expected rowCount=2, got %v", result["rowCount"])
	}
}

func TestExcelCommentCmd_DryRunUsesActualArgs(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "CreateComment",
			InputSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"driveId":     map[string]any{"type": "string"},
					"documentId":  map[string]any{"type": "string"},
					"cellAddress": map[string]any{"type": "string"},
					"content":     map[string]any{"type": "string"},
				},
				"required": []any{"cellAddress", "driveId", "documentId", "content"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	comment := "Confidential review note\nSecond line"
	cmd := &ExcelCommentCmd{
		DriveID:     "drive-001",
		DocumentID:  "doc-001",
		CellAddress: "Sheet1!A1",
		Text:        comment,
	}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	result := assertExcelDryRunValid(t, buf.Bytes())
	if strings.Contains(buf.String(), comment) {
		t.Fatal("dry-run output leaked Excel comment text")
	}
	if _, ok := result["content"]; ok {
		t.Fatal("dry-run display data should not include comment content")
	}
	if result["contentBytes"] != float64(len(comment)) {
		t.Errorf("expected contentBytes=%d, got %v", len(comment), result["contentBytes"])
	}
	if result["lineCount"] != float64(2) {
		t.Errorf("expected lineCount=2, got %v", result["lineCount"])
	}
}

func TestExcelReplyCmd_DryRunUsesActualArgs(t *testing.T) {
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

	reply := "Confidential reply\nSecond line"
	cmd := &ExcelReplyCmd{
		CommentID:  "comment-001",
		DriveID:    "drive-001",
		DocumentID: "doc-001",
		Text:       reply,
	}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	result := assertExcelDryRunValid(t, buf.Bytes())
	if strings.Contains(buf.String(), reply) {
		t.Fatal("dry-run output leaked Excel reply text")
	}
	if _, ok := result["newComment"]; ok {
		t.Fatal("dry-run display data should not include reply content")
	}
	if result["contentBytes"] != float64(len(reply)) {
		t.Errorf("expected contentBytes=%d, got %v", len(reply), result["contentBytes"])
	}
	if result["lineCount"] != float64(2) {
		t.Errorf("expected lineCount=2, got %v", result["lineCount"])
	}
}
