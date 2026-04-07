package sharepoint

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/testutil"
)

func TestSPFindCmd_Run(t *testing.T) {
	ctx, buf := testutil.SetupTestServer(t, map[string]string{
		"findSite": `{"id":"site-001","displayName":"Team Site","webUrl":"https://contoso.sharepoint.com/sites/teamsite"}`,
	})

	cmd := &SPFindSiteCmd{Query: "teamsite"}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if result["id"] != "site-001" {
		t.Errorf("expected id=site-001, got %v", result["id"])
	}
	if result["displayName"] != "Team Site" {
		t.Errorf("expected displayName=Team Site, got %v", result["displayName"])
	}
}

func TestSPMkdirCmd_DryRun(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "createFolder",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"driveId":    map[string]any{"type": "string"},
					"parentPath": map[string]any{"type": "string"},
					"folderName": map[string]any{"type": "string"},
				},
				"required": []any{"driveId", "folderName"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &SPMkdirCmd{
		DriveID:    "drive-001",
		ParentPath: "/Documents",
		FolderName: "NewFolder",
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
	if result["action"] != "sharepoint.mkdir" {
		t.Errorf("expected action=sharepoint.mkdir, got %v", result["action"])
	}
	val, ok := result["validation"].(map[string]any)
	if !ok {
		t.Fatal("expected validation object in dry-run output")
	}
	if val["valid"] != true {
		t.Errorf("expected valid=true, got %v; errors: %v", val["valid"], val["errors"])
	}
}

func TestSPRmCmd_NoInput(t *testing.T) {
	ctx, _ := testutil.SetupTestServer(t, nil)
	ctx.NoInput = true

	cmd := &SPDeleteCmd{DriveID: "drive-001", ItemPath: "/Documents/old.txt"}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when NoInput=true and Force=false")
	}
	if !strings.Contains(err.Error(), "without --force") {
		t.Errorf("expected error about --force, got: %v", err)
	}
}
