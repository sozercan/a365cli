package calendar

import (
	"encoding/json"
	"testing"

	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/testutil"
)

func TestCalListCmd_Run(t *testing.T) {
	ctx, buf := testutil.SetupTestServer(t, map[string]string{
		"ListEvents": "Events retrieved.\n{\"value\":[{\"id\":\"evt-001\",\"subject\":\"Standup\",\"start\":{\"dateTime\":\"2025-01-15T09:00:00\"},\"organizer\":{\"emailAddress\":{\"name\":\"Bob\",\"address\":\"bob@contoso.com\"}}}]}",
	})

	cmd := &CalListCmd{Max: 20}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	events, ok := result["events"]
	if !ok {
		t.Fatalf("expected 'events' key in output, got: %s", buf.String())
	}
	arr, ok := events.([]any)
	if !ok {
		t.Fatalf("expected 'events' to be an array, got: %T", events)
	}
	if len(arr) != 1 {
		t.Fatalf("expected 1 event, got %d", len(arr))
	}
}

func TestCalCreateCmd_DryRun(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "CreateEvent",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"subject":        map[string]any{"type": "string"},
					"startDateTime":  map[string]any{"type": "string"},
					"endDateTime":    map[string]any{"type": "string"},
					"attendeeEmails": map[string]any{"type": "array"},
				},
				"required": []any{"subject", "startDateTime", "endDateTime"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &CalCreateCmd{
		Subject:   "Team Meeting",
		Start:     "2025-01-20T10:00:00",
		End:       "2025-01-20T11:00:00",
		Attendees: []string{"alice@contoso.com"},
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
	if result["action"] != "calendar.create" {
		t.Errorf("expected action=calendar.create, got %v", result["action"])
	}
	val, ok := result["validation"].(map[string]any)
	if !ok {
		t.Fatal("expected validation object in dry-run output")
	}
	if val["valid"] != true {
		t.Errorf("expected valid=true, got %v; errors: %v", val["valid"], val["errors"])
	}
}

func TestCalDeleteCmd_Force(t *testing.T) {
	ctx, buf := testutil.SetupTestServer(t, map[string]string{
		"DeleteEventById": `{"message":"Event deleted"}`,
	})
	ctx.Force = true

	cmd := &CalDeleteCmd{ID: "evt-001"}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if result["message"] != "Event deleted" {
		t.Errorf("expected message='Event deleted', got %v", result["message"])
	}
}

func TestCalAcceptCmd_DryRun(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "AcceptEvent",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"eventId": map[string]any{"type": "string"},
				},
				"required": []any{"eventId"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &CalAcceptCmd{ID: "evt-001"}
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
	if result["action"] != "calendar.accept" {
		t.Errorf("expected action=calendar.accept, got %v", result["action"])
	}
	val, ok := result["validation"].(map[string]any)
	if !ok {
		t.Fatal("expected validation object in dry-run output")
	}
	if val["valid"] != true {
		t.Errorf("expected valid=true, got %v; errors: %v", val["valid"], val["errors"])
	}
}

func TestCalUpdateCmd_DryRunValidatesActualArgs(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "UpdateEvent",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"eventId": map[string]any{"type": "string"},
					"subject": map[string]any{"type": "string"},
				},
				"required": []any{"eventId", "subject"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &CalUpdateCmd{ID: "evt-001", Subject: "Project Sync"}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	val, ok := result["validation"].(map[string]any)
	if !ok || val["valid"] != true {
		t.Fatalf("expected valid dry-run output, got %v", result["validation"])
	}
}
