package triggers

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/testutil"
)

func TestTriggersEventsCmd_Run(t *testing.T) {
	ctx, buf := testutil.SetupTestServer(t, map[string]string{
		"list_event_types": `{"eventTypes":["mail.received","calendar.created","teams.message"]}`,
	})

	cmd := &TriggersEventsCmd{}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	eventTypes, ok := result["eventTypes"]
	if !ok {
		t.Fatalf("expected 'eventTypes' key in output, got: %s", buf.String())
	}
	arr, ok := eventTypes.([]any)
	if !ok {
		t.Fatalf("expected 'eventTypes' to be an array, got: %T", eventTypes)
	}
	if len(arr) != 3 {
		t.Fatalf("expected 3 event types, got %d", len(arr))
	}
}

func TestTriggersDeleteCmd_NoInput(t *testing.T) {
	ctx, _ := testutil.SetupTestServer(t, nil)
	ctx.NoInput = true

	cmd := &TriggersDeleteCmd{ID: "trigger-001"}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when NoInput=true and Force=false")
	}
	if !strings.Contains(err.Error(), "without --force") {
		t.Errorf("expected error about --force, got: %v", err)
	}
}

func TestTriggersCreateCmd_DryRun(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "create_trigger_definition",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"validationToken": map[string]any{"type": "string"},
					"name":            map[string]any{"type": "string"},
					"eventType":       map[string]any{"type": "string"},
				},
				"required": []any{"validationToken", "name", "eventType"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &TriggersCreateCmd{
		ValidationToken: "tok-001",
		Name:            "New Mail Alert",
		EventType:       "mail.received",
		Logic:           "always",
		Conditions:      `{"from":"alice@contoso.com"}`,
		Instructions:    "Notify on new mail",
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
	if result["action"] != "triggers.create" {
		t.Errorf("expected action=triggers.create, got %v", result["action"])
	}
	val, ok := result["validation"].(map[string]any)
	if !ok {
		t.Fatal("expected validation object in dry-run output")
	}
	if val["valid"] != true {
		t.Errorf("expected valid=true, got %v; errors: %v", val["valid"], val["errors"])
	}
}
