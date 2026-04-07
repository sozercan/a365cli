package mcp

import (
	"strings"
	"testing"
)

func TestSchemaRegistry_ValidArgs(t *testing.T) {
	tools := []ToolInfo{
		{
			Name: "SendMessage",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"chatId":  map[string]any{"type": "string"},
					"content": map[string]any{"type": "string"},
				},
				"required": []any{"chatId", "content"},
			},
		},
	}

	reg, err := NewSchemaRegistry(tools)
	if err != nil {
		t.Fatalf("NewSchemaRegistry: %v", err)
	}

	result := reg.Validate("SendMessage", map[string]any{
		"chatId":  "19:abc@thread.v2",
		"content": "Hello",
	})

	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got %v", result.Errors)
	}
}

func TestSchemaRegistry_MissingRequired(t *testing.T) {
	tools := []ToolInfo{
		{
			Name: "SendMessage",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"chatId":  map[string]any{"type": "string"},
					"content": map[string]any{"type": "string"},
				},
				"required": []any{"chatId", "content"},
			},
		},
	}

	reg, err := NewSchemaRegistry(tools)
	if err != nil {
		t.Fatalf("NewSchemaRegistry: %v", err)
	}

	result := reg.Validate("SendMessage", map[string]any{
		"chatId": "19:abc@thread.v2",
	})

	if result.Valid {
		t.Error("expected invalid, got valid")
	}

	found := false
	for _, e := range result.Errors {
		if strings.Contains(e, "content") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error mentioning 'content', got %v", result.Errors)
	}
}

func TestSchemaRegistry_WrongType(t *testing.T) {
	tools := []ToolInfo{
		{
			Name: "SendMessage",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"chatId":  map[string]any{"type": "string"},
					"content": map[string]any{"type": "string"},
				},
				"required": []any{"chatId", "content"},
			},
		},
	}

	reg, err := NewSchemaRegistry(tools)
	if err != nil {
		t.Fatalf("NewSchemaRegistry: %v", err)
	}

	result := reg.Validate("SendMessage", map[string]any{
		"chatId":  123,
		"content": "Hello",
	})

	if result.Valid {
		t.Error("expected invalid, got valid")
	}

	found := false
	for _, e := range result.Errors {
		if strings.Contains(e, "chatId") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error mentioning 'chatId', got %v", result.Errors)
	}
}

func TestSchemaRegistry_UnknownToolPermissive(t *testing.T) {
	reg, err := NewSchemaRegistry([]ToolInfo{})
	if err != nil {
		t.Fatalf("NewSchemaRegistry: %v", err)
	}

	result := reg.Validate("NonExistent", map[string]any{"anything": "goes"})
	if !result.Valid {
		t.Errorf("expected valid for unknown tool, got errors: %v", result.Errors)
	}
}

func TestSchemaRegistry_NilInputSchema(t *testing.T) {
	tools := []ToolInfo{
		{Name: "SimpleTool", InputSchema: nil},
	}

	reg, err := NewSchemaRegistry(tools)
	if err != nil {
		t.Fatalf("NewSchemaRegistry: %v", err)
	}

	result := reg.Validate("SimpleTool", map[string]any{"any": "args"})
	if !result.Valid {
		t.Errorf("expected valid for nil schema, got errors: %v", result.Errors)
	}
}

func TestSchemaRegistry_EmptySchema(t *testing.T) {
	tools := []ToolInfo{
		{
			Name: "FlexibleTool",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
	}

	reg, err := NewSchemaRegistry(tools)
	if err != nil {
		t.Fatalf("NewSchemaRegistry: %v", err)
	}

	result := reg.Validate("FlexibleTool", map[string]any{"any": "args"})
	if !result.Valid {
		t.Errorf("expected valid for empty schema, got errors: %v", result.Errors)
	}
}

func TestSchemaRegistry_MultipleTools(t *testing.T) {
	tools := []ToolInfo{
		{
			Name: "ToolA",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
				},
				"required": []any{"name"},
			},
		},
		{
			Name: "ToolB",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"count": map[string]any{"type": "integer"},
				},
				"required": []any{"count"},
			},
		},
	}

	reg, err := NewSchemaRegistry(tools)
	if err != nil {
		t.Fatalf("NewSchemaRegistry: %v", err)
	}

	// ToolA with valid args
	r := reg.Validate("ToolA", map[string]any{"name": "test"})
	if !r.Valid {
		t.Errorf("ToolA valid args: expected valid, got errors: %v", r.Errors)
	}

	// ToolA with missing name
	r = reg.Validate("ToolA", map[string]any{})
	if r.Valid {
		t.Error("ToolA missing name: expected invalid")
	}

	// ToolB with valid args
	r = reg.Validate("ToolB", map[string]any{"count": 5})
	if !r.Valid {
		t.Errorf("ToolB valid args: expected valid, got errors: %v", r.Errors)
	}

	// ToolB with wrong type
	r = reg.Validate("ToolB", map[string]any{"count": "not-a-number"})
	if r.Valid {
		t.Error("ToolB wrong type: expected invalid")
	}
}

func TestSchemaRegistry_NestedObject(t *testing.T) {
	tools := []ToolInfo{
		{
			Name: "CreateEvent",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"subject": map[string]any{"type": "string"},
					"location": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"displayName": map[string]any{"type": "string"},
						},
						"required": []any{"displayName"},
					},
				},
				"required": []any{"subject"},
			},
		},
	}

	reg, err := NewSchemaRegistry(tools)
	if err != nil {
		t.Fatalf("NewSchemaRegistry: %v", err)
	}

	// Valid with nested object
	r := reg.Validate("CreateEvent", map[string]any{
		"subject":  "Meeting",
		"location": map[string]any{"displayName": "Room 1"},
	})
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}

	// Invalid nested — missing required displayName
	r = reg.Validate("CreateEvent", map[string]any{
		"subject":  "Meeting",
		"location": map[string]any{},
	})
	if r.Valid {
		t.Error("expected invalid for missing nested required field")
	}
}

func TestSchemaRegistry_EnumConstraint(t *testing.T) {
	tools := []ToolInfo{
		{
			Name: "SetPriority",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"priority": map[string]any{
						"type": "string",
						"enum": []any{"low", "medium", "high"},
					},
				},
				"required": []any{"priority"},
			},
		},
	}

	reg, err := NewSchemaRegistry(tools)
	if err != nil {
		t.Fatalf("NewSchemaRegistry: %v", err)
	}

	r := reg.Validate("SetPriority", map[string]any{"priority": "high"})
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}

	r = reg.Validate("SetPriority", map[string]any{"priority": "critical"})
	if r.Valid {
		t.Error("expected invalid for non-enum value")
	}
}
