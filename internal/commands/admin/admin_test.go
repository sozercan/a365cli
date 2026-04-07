package admin

import (
	"encoding/json"
	"testing"

	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/testutil"
)

func TestAdminSetLicenseCmd_DryRun(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "mcp_Admin365_LicenseMgmtTools",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"action":         map[string]any{"type": "string"},
					"userId":         map[string]any{"type": "string"},
					"addLicenses":    map[string]any{"type": "array"},
					"removeLicenses": map[string]any{"type": "array"},
				},
				"required": []any{"userId"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &AdminSetLicenseCmd{
		UserID:         "00000000-0000-0000-0000-000000000000",
		AddLicenses:    []string{"sku-001"},
		RemoveLicenses: []string{"sku-002"},
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
