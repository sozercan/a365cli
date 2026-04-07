package knowledge

import (
	"encoding/json"
	"testing"

	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/testutil"
)

func TestKnowledgeConfigureCmd_DryRun(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "configure_federated_knowledge",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"action":      map[string]any{"type": "string"},
					"consumerId":  map[string]any{"type": "string"},
					"sourceType":  map[string]any{"type": "string"},
					"displayName": map[string]any{"type": "string"},
				},
				"required": []any{"consumerId", "sourceType", "displayName"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &KnowledgeConfigureCmd{
		ConsumerID:  "consumer-001",
		SourceType:  "sharepoint",
		DisplayName: "Engineering Docs",
		Description: "Engineering documentation source",
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

func TestKnowledgeIngestCmd_DryRun(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "ingest_federated_knowledge",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"action":                map[string]any{"type": "string"},
					"consumerId":            map[string]any{"type": "string"},
					"searchConfigurationId": map[string]any{"type": "string"},
				},
				"required": []any{"consumerId", "searchConfigurationId"},
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
