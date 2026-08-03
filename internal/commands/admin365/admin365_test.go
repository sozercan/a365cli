package admin365

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/testutil"
)

type admin365DryRunCommand interface {
	Run(*commands.Context) error
}

func assertAdmin365DryRunValid(t *testing.T, output []byte) map[string]any {
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

func TestAdmin365BulkAddCmd_DryRunUsesActualArgsAndRedactsContent(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "BulkAddUsers",
			InputSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"fileContent": map[string]any{"type": "string"},
				},
				"required": []any{"fileContent"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	fileContent := "name,email\nAlice,alice@contoso.com"
	cmd := &Admin365BulkAddCmd{FileContent: fileContent}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	result := assertAdmin365DryRunValid(t, buf.Bytes())
	if strings.Contains(buf.String(), fileContent) {
		t.Fatal("dry-run output leaked bulk-add file content")
	}
	if _, ok := result["fileContent"]; ok {
		t.Fatal("dry-run display data should not include fileContent")
	}
	if result["contentBytes"] != float64(len(fileContent)) {
		t.Errorf("expected contentBytes=%d, got %v", len(fileContent), result["contentBytes"])
	}
	if result["lineCount"] != float64(2) {
		t.Errorf("expected lineCount=2, got %v", result["lineCount"])
	}
}

func TestAdmin365SettingsCommands_DryRunUsesActualArgs(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		argName  string
		command  admin365DryRunCommand
	}{
		{
			name:     "set access",
			toolName: "UpdateWhoCanAccessAgentsSettings",
			argName:  "accessLevel",
			command:  &Admin365SetAccessCmd{AccessLevel: "everyone"},
		},
		{
			name:     "set sharing",
			toolName: "UpdateWhoCanShareAgentsOrgWideSettings",
			argName:  "accessLevel",
			command:  &Admin365SetSharingCmd{AccessLevel: "organization"},
		},
		{
			name:     "set Microsoft apps",
			toolName: "UpdateCanInstallMicrosoftAppsAndAgentsSettings",
			argName:  "allowed",
			command:  &Admin365SetMsAppsCmd{Allowed: "true"},
		},
		{
			name:     "set third-party apps",
			toolName: "UpdateCanInstallThirdPartyAppsAndAgentsSettings",
			argName:  "allowed",
			command:  &Admin365SetThirdPartyCmd{Allowed: "false"},
		},
		{
			name:     "set LOB apps",
			toolName: "UpdateCanInstallLOBAppsAndAgentsSettings",
			argName:  "allowed",
			command:  &Admin365SetLobAppsCmd{Allowed: "true"},
		},
		{
			name:     "set Copilot",
			toolName: "UpdateCopilotAdminSettings",
			argName:  "isEnabled",
			command:  &Admin365SetCopilotCmd{IsEnabled: "true"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemas := []mcp.ToolInfo{
				{
					Name: tt.toolName,
					InputSchema: map[string]any{
						"type":                 "object",
						"additionalProperties": false,
						"properties": map[string]any{
							tt.argName: map[string]any{"type": "string"},
						},
						"required": []any{tt.argName},
					},
				},
			}
			ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
			ctx.DryRun = true

			if err := tt.command.Run(ctx); err != nil {
				t.Fatalf("Run() error: %v", err)
			}
			assertAdmin365DryRunValid(t, buf.Bytes())
		})
	}
}
