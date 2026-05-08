package exec_test

import (
	"encoding/json"
	"strings"
	"testing"

	cmdexec "github.com/sozercan/a365cli/internal/commands/exec"
	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/testutil"
)

func TestQueryRendersFirstMatchingListKeyAndAppliesMax(t *testing.T) {
	ctx, buf, recorder := testutil.SetupTestServerWithRecorder(t, map[string]string{
		"ListThings": `{"value":[{"id":"one"},{"id":"two"}]}`,
	})

	err := cmdexec.New(ctx).Query(cmdexec.ToolCall{
		Service:   "mail",
		Tool:      "ListThings",
		Args:      map[string]any{"filter": "recent"},
		ErrPrefix: "list things",
		Output:    cmdexec.List("things", nil, "things", "value").WithMax(1),
	})
	if err != nil {
		t.Fatalf("Query() error: %v", err)
	}

	calls := recorder.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected one tools/call, got %d", len(calls))
	}
	if calls[0].Name != "ListThings" {
		t.Fatalf("expected ListThings call, got %q", calls[0].Name)
	}
	if calls[0].Arguments["filter"] != "recent" {
		t.Fatalf("expected recorded filter arg, got %v", calls[0].Arguments)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	things, ok := result["things"].([]any)
	if !ok {
		t.Fatalf("expected things array, got %T in %v", result["things"], result)
	}
	if len(things) != 1 {
		t.Fatalf("expected max-limited list length 1, got %d", len(things))
	}
}

func TestMutateDryRunValidatesActualArgsAndDoesNotCallTool(t *testing.T) {
	schemas := []mcp.ToolInfo{{
		Name: "UpdateThing",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id":    map[string]any{"type": "string"},
				"state": map[string]any{"type": "string"},
			},
			"required": []any{"id", "state"},
		},
	}}
	ctx, buf, recorder := testutil.SetupTestServerWithSchemasAndRecorder(t, nil, schemas)
	ctx.DryRun = true

	err := cmdexec.New(ctx).Mutate(cmdexec.OperationPlan{
		Service: "mail",
		Tool:    "UpdateThing",
		Args: map[string]any{
			"id":    "thing-001",
			"state": "closed",
		},
		Action:  "update thing thing-001",
		Display: map[string]any{"action": "things.update", "id": "thing-001"},
		Output:  cmdexec.Mutation("Thing updated"),
	})
	if err != nil {
		t.Fatalf("Mutate() error: %v", err)
	}
	if len(recorder.Calls()) != 0 {
		t.Fatalf("dry-run should not call tools/call, got %#v", recorder.Calls())
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	validation, ok := result["validation"].(map[string]any)
	if !ok {
		t.Fatalf("expected validation object, got %v", result["validation"])
	}
	if validation["valid"] != true {
		t.Fatalf("expected valid dry-run validation, got %v", validation)
	}
}

func TestMutateDestructiveNoInputDoesNotCallTool(t *testing.T) {
	ctx, _, recorder := testutil.SetupTestServerWithRecorder(t, map[string]string{
		"DeleteThing": `{"deleted":true}`,
	})
	ctx.NoInput = true

	err := cmdexec.New(ctx).Mutate(cmdexec.OperationPlan{
		Service:     "mail",
		Tool:        "DeleteThing",
		Args:        map[string]any{"id": "thing-001"},
		Action:      "delete thing thing-001",
		Display:     map[string]any{"action": "things.delete", "id": "thing-001"},
		Destructive: true,
		ConfirmText: "delete thing thing-001",
		Output:      cmdexec.Mutation("Thing deleted"),
	})
	if err == nil {
		t.Fatal("expected destructive non-interactive confirmation error")
	}
	if !strings.Contains(err.Error(), "without --force") {
		t.Fatalf("expected --force guidance, got %v", err)
	}
	if len(recorder.Calls()) != 0 {
		t.Fatalf("confirmation failure should not call tools/call, got %#v", recorder.Calls())
	}
}
