package excel

import (
	"encoding/json"
	"testing"

	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/testutil"
)

func TestExcelCreateCmd_DryRun(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "CreateWorkbook",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"action":          map[string]any{"type": "string"},
					"desiredFileName": map[string]any{"type": "string"},
				},
				"required": []any{"desiredFileName"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &ExcelCreateCmd{FileName: "budget.xlsx"}
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
}

func TestExcelCommentCmd_DryRun(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "CreateComment",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"action":      map[string]any{"type": "string"},
					"driveId":     map[string]any{"type": "string"},
					"documentId":  map[string]any{"type": "string"},
					"cellAddress": map[string]any{"type": "string"},
				},
				"required": []any{"driveId", "documentId", "cellAddress"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &ExcelCommentCmd{
		DriveID:     "drive-001",
		DocumentID:  "doc-001",
		CellAddress: "A1",
		Text:        "Looks good",
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
}
