package knowledge

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/testutil"
)

func TestKnowledgeConfigureCmd_DryRunValidatesActualArgs(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "configure_federated_knowledge",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"consumerId":      map[string]any{"type": "string"},
					"knowledgeConfig": map[string]any{"type": "string"},
					"sourceType":      map[string]any{"type": "string"},
					"displayName":     map[string]any{"type": "string"},
					"description":     map[string]any{"type": "string"},
					"hints":           map[string]any{"type": "string"},
				},
				"required":             []any{"consumerId", "knowledgeConfig", "sourceType", "displayName", "description"},
				"additionalProperties": false,
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &KnowledgeConfigureCmd{
		ConsumerID:  "consumer-001",
		SourceType:  "sharepoint",
		DisplayName: "Internal Engineering Docs",
		Description: "Confidential engineering documentation source",
	}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	result := decodeDryRunResult(t, buf.Bytes())
	assertDryRunValid(t, result)
	if result["display_name_len"] != float64(len(cmd.DisplayName)) {
		t.Errorf("expected display_name_len=%d, got %v", len(cmd.DisplayName), result["display_name_len"])
	}
	if result["description_len"] != float64(len(cmd.Description)) {
		t.Errorf("expected description_len=%d, got %v", len(cmd.Description), result["description_len"])
	}
	if strings.Contains(buf.String(), cmd.DisplayName) || strings.Contains(buf.String(), cmd.Description) {
		t.Fatalf("dry-run output leaked knowledge source text: %s", buf.String())
	}
}

func TestKnowledgeIngestCmd_DryRunValidatesActualArgs(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "ingest_federated_knowledge",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"consumerId":            map[string]any{"type": "string"},
					"searchConfigurationId": map[string]any{"type": "string"},
				},
				"required":             []any{"consumerId", "searchConfigurationId"},
				"additionalProperties": false,
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &KnowledgeIngestCmd{
		ConsumerID: "consumer-001",
		ConfigID:   "config-001",
	}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	assertDryRunValid(t, decodeDryRunResult(t, buf.Bytes()))
}

func TestKnowledgeDeleteCmd_DryRunValidatesActualArgs(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "delete_federated_knowledge",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"consumerId":            map[string]any{"type": "string"},
					"searchConfigurationId": map[string]any{"type": "string"},
				},
				"required":             []any{"searchConfigurationId", "consumerId"},
				"additionalProperties": false,
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &KnowledgeDeleteCmd{
		ConsumerID: "consumer-001",
		ConfigID:   "config-001",
	}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	assertDryRunValid(t, decodeDryRunResult(t, buf.Bytes()))
}

func decodeDryRunResult(t *testing.T, data []byte) map[string]any {
	t.Helper()

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, data)
	}
	return result
}

func assertDryRunValid(t *testing.T, result map[string]any) {
	t.Helper()

	if result["dry_run"] != true {
		t.Errorf("expected dry_run=true, got %v", result["dry_run"])
	}
	validation, ok := result["validation"].(map[string]any)
	if !ok {
		t.Fatal("expected validation object")
	}
	if validation["valid"] != true {
		t.Fatalf("expected valid=true, got %v; errors: %v", validation["valid"], validation["errors"])
	}
}
