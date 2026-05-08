package splists

import (
	"encoding/json"
	"testing"

	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/testutil"
)

func TestSPLCreateCmd_DryRun(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "createList",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"action":      map[string]any{"type": "string"},
					"siteId":      map[string]any{"type": "string"},
					"displayName": map[string]any{"type": "string"},
				},
				"required": []any{"siteId", "displayName"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &SPLCreateCmd{
		SiteID:      "site-001",
		DisplayName: "Project Tasks",
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

func TestSPLAddColumnCmd_DryRun(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "createListColumn",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"action":     map[string]any{"type": "string"},
					"siteId":     map[string]any{"type": "string"},
					"listId":     map[string]any{"type": "string"},
					"name":       map[string]any{"type": "string"},
					"columnType": map[string]any{"type": "string"},
				},
				"required": []any{"siteId", "listId", "name", "columnType"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &SPLAddColumnCmd{
		SiteID:     "site-001",
		ListID:     "list-001",
		Name:       "Priority",
		ColumnType: "choice",
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

func TestSPLUpdateItemCmd_DryRunValidatesActualArgs(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "updateListItem",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"siteId": map[string]any{"type": "string"},
					"listId": map[string]any{"type": "string"},
					"itemId": map[string]any{"type": "string"},
					"fields": map[string]any{"type": "object"},
				},
				"required": []any{"siteId", "listId", "itemId", "fields"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &SPLUpdateItemCmd{
		SiteID: "site-001",
		ListID: "list-001",
		ItemID: "item-001",
		Fields: `{"Priority":"High"}`,
	}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	val, ok := result["validation"].(map[string]any)
	if !ok || val["valid"] != true {
		t.Fatalf("expected valid dry-run output, got %v", result["validation"])
	}
}
