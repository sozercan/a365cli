package calendar

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/testutil"
)

type calendarDryRunCommand interface {
	Run(*commands.Context) error
}

func assertCalendarDryRunValid(t *testing.T, output []byte) map[string]any {
	t.Helper()

	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, output)
	}
	if result["dry_run"] != true {
		t.Errorf("expected dry_run=true, got %v", result["dry_run"])
	}
	validation, ok := result["validation"].(map[string]any)
	if !ok {
		t.Fatal("expected validation object in dry-run output")
	}
	if validation["valid"] != true {
		t.Errorf("expected valid=true, got %v; errors: %v", validation["valid"], validation["errors"])
	}
	return result
}

func TestCalListCmd_Run(t *testing.T) {
	ctx, buf := testutil.SetupTestServer(t, map[string]string{
		"ListEvents": "Events retrieved.\n{\"value\":[{\"id\":\"evt-001\",\"subject\":\"Standup\",\"start\":{\"dateTime\":\"2026-01-15T09:00:00\"},\"organizer\":{\"emailAddress\":{\"name\":\"Bob\",\"address\":\"bob@contoso.com\"}}}]}",
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

func TestCalCreateCmd_DryRunUsesLiveArgs(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "CreateEvent",
			InputSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"subject":         map[string]any{"type": "string"},
					"startDateTime":   map[string]any{"type": "string"},
					"endDateTime":     map[string]any{"type": "string"},
					"attendeeEmails":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"bodyContent":     map[string]any{"type": "string"},
					"isOnlineMeeting": map[string]any{"type": "boolean"},
				},
				"required": []any{"subject", "attendeeEmails", "startDateTime", "endDateTime"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	subject := "Confidential Team Meeting"
	body := "Private weekly planning notes"
	attendee := "alice@contoso.com"
	cmd := &CalCreateCmd{
		Subject:   subject,
		Start:     "2026-08-10T10:00:00",
		End:       "2026-08-10T11:00:00",
		Attendees: []string{attendee},
		Body:      body,
		IsOnline:  true,
	}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	result := assertCalendarDryRunValid(t, buf.Bytes())
	if result["action"] != "calendar.create" {
		t.Errorf("expected action=calendar.create, got %v", result["action"])
	}
	for _, sensitive := range []string{subject, body, attendee} {
		if strings.Contains(buf.String(), sensitive) {
			t.Fatalf("dry-run output leaked calendar create value %q", sensitive)
		}
	}
	if result["subjectBytes"] != float64(len(subject)) {
		t.Errorf("expected subjectBytes=%d, got %v", len(subject), result["subjectBytes"])
	}
	if result["bodyBytes"] != float64(len(body)) {
		t.Errorf("expected bodyBytes=%d, got %v", len(body), result["bodyBytes"])
	}
	if result["attendeeCount"] != float64(1) {
		t.Errorf("expected attendeeCount=1, got %v", result["attendeeCount"])
	}
}

func TestCalUpdateCmd_DryRunValidatesActualArgs(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "UpdateEvent",
			InputSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"eventId":       map[string]any{"type": "string"},
					"subject":       map[string]any{"type": "string"},
					"startDateTime": map[string]any{"type": "string"},
					"endDateTime":   map[string]any{"type": "string"},
					"body":          map[string]any{"type": "string"},
				},
				"required": []any{"eventId"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &CalUpdateCmd{
		ID:      "evt-001",
		Subject: "Updated subject",
		Start:   "2026-08-11T09:00:00",
		End:     "2026-08-11T09:30:00",
		Body:    "Updated body",
	}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	result := assertCalendarDryRunValid(t, buf.Bytes())
	if result["action"] != "calendar.update" {
		t.Errorf("expected action=calendar.update, got %v", result["action"])
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

func TestCalEventIDMutations_DryRunUseActualArgs(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		command  calendarDryRunCommand
	}{
		{name: "delete", toolName: "DeleteEventById", command: &CalDeleteCmd{ID: "evt-001"}},
		{name: "accept", toolName: "AcceptEvent", command: &CalAcceptCmd{ID: "evt-001"}},
		{name: "tentative", toolName: "TentativelyAcceptEvent", command: &CalTentativeCmd{ID: "evt-001"}},
		{name: "decline", toolName: "DeclineEvent", command: &CalDeclineCmd{ID: "evt-001"}},
		{name: "cancel", toolName: "CancelEvent", command: &CalCancelCmd{ID: "evt-001"}},
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
							"eventId": map[string]any{"type": "string"},
						},
						"required": []any{"eventId"},
					},
				},
			}
			ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
			ctx.DryRun = true

			if err := tt.command.Run(ctx); err != nil {
				t.Fatalf("Run() error: %v", err)
			}
			assertCalendarDryRunValid(t, buf.Bytes())
		})
	}
}

func TestCalForwardCmd_DryRunUsesActualArgs(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "ForwardEvent",
			InputSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"eventId":         map[string]any{"type": "string"},
					"recipientEmails": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"comment":         map[string]any{"type": "string"},
				},
				"required": []any{"eventId", "recipientEmails"},
			},
		},
	}
	ctx, buf := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	recipient := "alice@contoso.com"
	comment := "Private forwarding note"
	cmd := &CalForwardCmd{
		ID:         "evt-001",
		Recipients: []string{recipient},
		Comment:    comment,
	}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	result := assertCalendarDryRunValid(t, buf.Bytes())
	for _, sensitive := range []string{recipient, comment} {
		if strings.Contains(buf.String(), sensitive) {
			t.Fatalf("dry-run output leaked calendar forward value %q", sensitive)
		}
	}
	if result["recipientCount"] != float64(1) {
		t.Errorf("expected recipientCount=1, got %v", result["recipientCount"])
	}
	if result["commentBytes"] != float64(len(comment)) {
		t.Errorf("expected commentBytes=%d, got %v", len(comment), result["commentBytes"])
	}
}

func TestCalCreateCmd_DryRunUsesEmptyAttendeeArray(t *testing.T) {
	schemas := []mcp.ToolInfo{
		{
			Name: "CreateEvent",
			InputSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"subject":       map[string]any{"type": "string"},
					"startDateTime": map[string]any{"type": "string"},
					"endDateTime":   map[string]any{"type": "string"},
					"attendeeEmails": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "string"},
					},
				},
				"required": []any{"subject", "attendeeEmails", "startDateTime", "endDateTime"},
			},
		},
	}
	ctx, _ := testutil.SetupTestServerWithSchemas(t, nil, schemas)
	ctx.DryRun = true

	cmd := &CalCreateCmd{
		Subject: "Focus time",
		Start:   "2026-08-04T10:00:00",
		End:     "2026-08-04T10:30:00",
	}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}
}
