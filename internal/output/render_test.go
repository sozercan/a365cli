package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderTable(t *testing.T) {
	columns := []Column{
		{Header: "NAME", Width: 20, Extract: func(r map[string]any) string { return getString(r, "name") }},
		{Header: "ID", Width: 0, Extract: func(r map[string]any) string { return getString(r, "id") }},
	}
	rows := []map[string]any{
		{"name": "Alice", "id": "1"},
		{"name": "Bob", "id": "2"},
	}

	var buf bytes.Buffer
	RenderTable(&buf, columns, rows)
	out := buf.String()

	if !strings.Contains(out, "NAME") || !strings.Contains(out, "ID") {
		t.Error("expected header row")
	}
	if !strings.Contains(out, "Alice") || !strings.Contains(out, "Bob") {
		t.Error("expected data rows")
	}
	// Should have alignment (tabwriter adds spaces)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines (header + 2 rows), got %d", len(lines))
	}
}

func TestRenderTable_UsesConfiguredColumnWidthForAlignment(t *testing.T) {
	columns := []Column{
		{Header: "NAME", Width: 10, Extract: func(r map[string]any) string { return getString(r, "name") }},
		{Header: "ID", Width: 0, Extract: func(r map[string]any) string { return getString(r, "id") }},
	}
	rows := []map[string]any{
		{"name": "Alice", "id": "1"},
		{"name": "Bob", "id": "2"},
	}

	var buf bytes.Buffer
	RenderTable(&buf, columns, rows)
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	headerID := strings.Index(lines[0], "ID")
	row1ID := strings.Index(lines[1], "1")
	row2ID := strings.Index(lines[2], "2")

	if headerID != 12 {
		t.Fatalf("expected ID header to start at column 12, got %d in %q", headerID, lines[0])
	}
	if row1ID != headerID || row2ID != headerID {
		t.Fatalf(
			"expected ID column to align at index %d, got row1=%d row2=%d\nheader=%q\nrow1=%q\nrow2=%q",
			headerID,
			row1ID,
			row2ID,
			lines[0],
			lines[1],
			lines[2],
		)
	}
}

func TestRenderTSV(t *testing.T) {
	columns := []Column{
		{Header: "NAME", Extract: func(r map[string]any) string { return getString(r, "name") }},
		{Header: "ID", Extract: func(r map[string]any) string { return getString(r, "id") }},
	}
	rows := []map[string]any{
		{"name": "Alice", "id": "1"},
	}

	var buf bytes.Buffer
	RenderTSV(&buf, columns, rows)
	out := buf.String()

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	// Header should be tab-separated
	if lines[0] != "NAME\tID" {
		t.Errorf("expected 'NAME\\tID', got %q", lines[0])
	}
	if lines[1] != "Alice\t1" {
		t.Errorf("expected 'Alice\\t1', got %q", lines[1])
	}
}

func TestRenderKeyValue(t *testing.T) {
	item := map[string]any{
		"name":  "Test Team",
		"id":    "abc-123",
		"count": 42.0,
	}

	var buf bytes.Buffer
	RenderKeyValue(&buf, item)
	out := buf.String()

	if !strings.Contains(out, "name:") || !strings.Contains(out, "Test Team") {
		t.Error("expected name key-value")
	}
	if !strings.Contains(out, "id:") || !strings.Contains(out, "abc-123") {
		t.Error("expected id key-value")
	}

	// Keys should be sorted
	nameIdx := strings.Index(out, "count:")
	idIdx := strings.Index(out, "id:")
	if nameIdx > idIdx {
		t.Error("expected keys to be sorted alphabetically")
	}
}

func TestRenderTable_EmptyRows(t *testing.T) {
	columns := []Column{
		{Header: "NAME", Extract: func(r map[string]any) string { return getString(r, "name") }},
	}

	var buf bytes.Buffer
	RenderTable(&buf, columns, nil)
	out := buf.String()

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 {
		t.Errorf("expected header only for empty rows, got %d lines", len(lines))
	}
}

func TestRenderKeyValue_NestedMap(t *testing.T) {
	item := map[string]any{
		"from": map[string]any{"displayName": "Alice", "id": "1"},
	}

	var buf bytes.Buffer
	RenderKeyValue(&buf, item)
	out := buf.String()

	if !strings.Contains(out, "from:") {
		t.Error("expected 'from:' key")
	}
	if !strings.Contains(out, "displayName=Alice") {
		t.Error("expected nested map to be flattened")
	}
}

func TestRenderTSV_EmptyRows(t *testing.T) {
	columns := []Column{
		{Header: "NAME", Extract: func(r map[string]any) string { return getString(r, "name") }},
		{Header: "ID", Extract: func(r map[string]any) string { return getString(r, "id") }},
	}

	var buf bytes.Buffer
	RenderTSV(&buf, columns, nil)
	out := buf.String()

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 {
		t.Errorf("expected only header line for empty rows, got %d lines", len(lines))
	}
	if lines[0] != "NAME\tID" {
		t.Errorf("expected 'NAME\\tID', got %q", lines[0])
	}
}

func TestRenderKeyValue_NilValue(t *testing.T) {
	item := map[string]any{
		"name":   "Test",
		"detail": nil,
	}

	var buf bytes.Buffer
	RenderKeyValue(&buf, item)
	out := buf.String()

	if !strings.Contains(out, "name:") {
		t.Error("expected 'name:' key")
	}
	if !strings.Contains(out, "Test") {
		t.Error("expected 'Test' value")
	}
	// nil values should render as empty string
	if !strings.Contains(out, "detail:") {
		t.Error("expected 'detail:' key for nil value")
	}
}

func TestRenderKeyValue_ArrayValue(t *testing.T) {
	item := map[string]any{
		"tags": []any{"urgent", "bug", "frontend"},
	}

	var buf bytes.Buffer
	RenderKeyValue(&buf, item)
	out := buf.String()

	if !strings.Contains(out, "tags:") {
		t.Error("expected 'tags:' key")
	}
	if !strings.Contains(out, "urgent") {
		t.Error("expected 'urgent' in array output")
	}
	if !strings.Contains(out, "bug") {
		t.Error("expected 'bug' in array output")
	}
	if !strings.Contains(out, "frontend") {
		t.Error("expected 'frontend' in array output")
	}
}

func TestRenderTable_Truncation(t *testing.T) {
	columns := []Column{
		{Header: "NAME", Width: 10, Extract: func(r map[string]any) string { return getString(r, "name") }},
		{Header: "ID", Width: 5, Extract: func(r map[string]any) string { return getString(r, "id") }},
	}
	rows := []map[string]any{
		{"name": "This is a very long name that should be truncated", "id": "12345678"},
	}

	var buf bytes.Buffer
	RenderTable(&buf, columns, rows)
	out := buf.String()

	// The long name should be truncated to 10 chars (7 + "...")
	if strings.Contains(out, "This is a very long") {
		t.Error("expected long name to be truncated")
	}
	if !strings.Contains(out, "This is...") {
		t.Errorf("expected truncated name 'This is...' with ellipsis, got %q", out)
	}
	// ID column Width=5 should truncate "12345678" to "12..."
	if strings.Contains(out, "12345678") {
		t.Error("expected long ID to be truncated")
	}
	if !strings.Contains(out, "12...") {
		t.Errorf("expected truncated ID '12...' with ellipsis, got %q", out)
	}
}
