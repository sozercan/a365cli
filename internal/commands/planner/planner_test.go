package planner

import (
	"encoding/json"
	"testing"

	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/testutil"
)

func TestPlansListCmd_Run(t *testing.T) {
	ctx, buf := testutil.SetupTestServer(t, map[string]string{
		"QueryPlans": `{"plans":[{"id":"plan-001","title":"Q1 Sprint","createdDateTime":"2025-01-01T00:00:00Z"},{"id":"plan-002","title":"Bug Triage","createdDateTime":"2025-01-10T00:00:00Z"}]}`,
	})

	cmd := &PlansListCmd{Max: 50}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	plans, ok := result["plans"]
	if !ok {
		t.Fatalf("expected 'plans' key in output, got: %s", buf.String())
	}
	arr, ok := plans.([]any)
	if !ok {
		t.Fatalf("expected 'plans' to be an array, got: %T", plans)
	}
	if len(arr) != 2 {
		t.Fatalf("expected 2 plans, got %d", len(arr))
	}
}

func TestTasksCreateCmd_DryRun(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "CreateTask",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"planId": map[string]any{"type": "string"},
					"title":  map[string]any{"type": "string"},
				},
				"required": []any{"planId", "title"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &TasksCreateCmd{PlanID: "plan-001", Title: "Fix login bug"}
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
	if result["action"] != "planner.create-task" {
		t.Errorf("expected action=planner.create-task, got %v", result["action"])
	}
	val, ok := result["validation"].(map[string]any)
	if !ok {
		t.Fatal("expected validation object in dry-run output")
	}
	if val["valid"] != true {
		t.Errorf("expected valid=true, got %v; errors: %v", val["valid"], val["errors"])
	}
}
