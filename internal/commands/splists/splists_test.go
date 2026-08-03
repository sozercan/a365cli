package splists

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/testutil"
)

func TestSPLCreateCmd_DryRun(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "createList",
			InputSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"siteId":      map[string]any{"type": "string"},
					"displayName": map[string]any{"type": "string"},
					"list": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"template": map[string]any{"type": "string"},
						},
						"required": []any{"template"},
					},
				},
				"required": []any{"siteId", "displayName", "list"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &SPLCreateCmd{
		SiteID:      "site-001",
		DisplayName: "Project Tasks",
		Template:    "genericList",
	}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	assertSPLDryRunValid(t, decodeSPLDryRunResult(t, buf.String()))
}

func TestSPLAddColumnCmd_DryRun(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "createListColumn",
			InputSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"siteId": map[string]any{"type": "string"},
					"listId": map[string]any{"type": "string"},
					"name":   map[string]any{"type": "string"},
					"choice": map[string]any{
						"type":                 "object",
						"additionalProperties": false,
						"properties": map[string]any{
							"choices": map[string]any{
								"type":  "array",
								"items": map[string]any{"type": "string"},
							},
						},
						"required": []any{"choices"},
					},
				},
				"required": []any{"siteId", "listId", "name"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &SPLAddColumnCmd{
		SiteID:         "site-001",
		ListID:         "list-001",
		Name:           "Priority",
		ColumnType:     "choice",
		ColumnSettings: `{"choices":["High","Low"]}`,
	}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	assertSPLDryRunValid(t, decodeSPLDryRunResult(t, buf.String()))
}

func TestSPLUpdateItemCmd_DryRunValidatesActualFields(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "updateListItem",
			InputSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"siteId": map[string]any{"type": "string"},
					"listId": map[string]any{"type": "string"},
					"itemId": map[string]any{"type": "string"},
					"fields": map[string]any{
						"type":          "object",
						"minProperties": 1,
					},
					"ifMatch": map[string]any{"type": "string"},
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
		ItemID: "42",
		Fields: `{"Title":"Updated item"}`,
	}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if strings.Contains(buf.String(), "Updated item") || strings.Contains(buf.String(), `"fields"`) {
		t.Fatalf("dry-run output leaked list item field values: %s", buf.String())
	}
	assertSPLDryRunValid(t, decodeSPLDryRunResult(t, buf.String()))
}

func decodeSPLDryRunResult(t *testing.T, output string) map[string]any {
	t.Helper()

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, output)
	}
	if result["dry_run"] != true {
		t.Errorf("expected dry_run=true, got %v", result["dry_run"])
	}
	return result
}

func assertSPLDryRunValid(t *testing.T, result map[string]any) {
	t.Helper()

	validation, ok := result["validation"].(map[string]any)
	if !ok {
		t.Fatal("expected validation object")
	}
	if validation["valid"] != true {
		t.Errorf("expected valid=true, got %v; errors: %v", validation["valid"], validation["errors"])
	}
}
