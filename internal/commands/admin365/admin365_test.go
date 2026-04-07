package admin365

import (
	"encoding/json"
	"testing"

	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/testutil"
)

func TestAdmin365BulkAddCmd_DryRun(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "BulkAddUsers",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"action":      map[string]any{"type": "string"},
					"fileContent": map[string]any{"type": "string"},
				},
				"required": []any{"fileContent"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &Admin365BulkAddCmd{FileContent: "name,email\nAlice,alice@contoso.com"}
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

func TestAdmin365SetAccessCmd_DryRun(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "UpdateWhoCanAccessAgentsSettings",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"action":      map[string]any{"type": "string"},
					"accessLevel": map[string]any{"type": "string"},
				},
				"required": []any{"accessLevel"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &Admin365SetAccessCmd{AccessLevel: "everyone"}
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
