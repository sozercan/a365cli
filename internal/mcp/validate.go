package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"golang.org/x/text/message"
)

// ValidationResult holds the outcome of validating tool arguments against a schema.
type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors"`
}

// SchemaRegistry compiles and caches JSON Schemas for MCP tools.
type SchemaRegistry struct {
	schemas map[string]*jsonschema.Schema
}

// NewSchemaRegistry builds a registry from the tools returned by ListTools.
// Each tool's InputSchema is compiled into a reusable *jsonschema.Schema.
func NewSchemaRegistry(tools []ToolInfo) (*SchemaRegistry, error) {
	reg := &SchemaRegistry{schemas: make(map[string]*jsonschema.Schema, len(tools))}
	for _, t := range tools {
		if t.InputSchema == nil {
			continue
		}

		schemaBytes, err := json.Marshal(t.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("marshaling schema for tool %q: %w", t.Name, err)
		}

		inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaBytes))
		if err != nil {
			return nil, fmt.Errorf("unmarshaling schema for tool %q: %w", t.Name, err)
		}

		c := jsonschema.NewCompiler()
		resourceURL := "schema://tools/" + t.Name + ".json"
		if err := c.AddResource(resourceURL, inst); err != nil {
			return nil, fmt.Errorf("adding schema resource for tool %q: %w", t.Name, err)
		}

		sch, err := c.Compile(resourceURL)
		if err != nil {
			return nil, fmt.Errorf("compiling schema for tool %q: %w", t.Name, err)
		}

		reg.schemas[t.Name] = sch
	}
	return reg, nil
}

// Validate checks args against the compiled schema for the named tool.
// Returns valid=true if no schema is registered for the tool (permissive).
// Args are normalized via JSON round-trip to convert Go types (e.g., []string)
// to JSON-compatible types (e.g., []any) before validation.
func (r *SchemaRegistry) Validate(toolName string, args map[string]any) ValidationResult {
	sch, ok := r.schemas[toolName]
	if !ok {
		return ValidationResult{Valid: true}
	}

	// Normalize args via JSON round-trip so typed slices ([]string, []int)
	// become []any, matching what JSON Schema validation expects.
	normalized, err := normalizeArgs(args)
	if err != nil {
		return ValidationResult{Valid: false, Errors: []string{fmt.Sprintf("normalizing args: %v", err)}}
	}

	err = sch.Validate(normalized)
	if err == nil {
		return ValidationResult{Valid: true}
	}

	ve, ok := err.(*jsonschema.ValidationError)
	if !ok {
		return ValidationResult{Valid: false, Errors: []string{err.Error()}}
	}

	return ValidationResult{
		Valid:  false,
		Errors: flattenValidationErrors(ve),
	}
}

// normalizeArgs converts Go-typed maps (with []string, []int, etc.) to
// JSON-compatible types via marshal/unmarshal round-trip.
func normalizeArgs(args map[string]any) (any, error) {
	data, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	return jsonschema.UnmarshalJSON(bytes.NewReader(data))
}

// flattenValidationErrors walks the tree of validation errors and returns
// human-readable strings.
func flattenValidationErrors(ve *jsonschema.ValidationError) []string {
	var errs []string
	printer := message.NewPrinter(message.MatchLanguage("en"))

	// Leaf errors have no causes — they are the actual failures.
	if len(ve.Causes) == 0 {
		path := "/" + strings.Join(ve.InstanceLocation, "/")
		if path == "/" {
			path = "(root)"
		}
		msg := ve.ErrorKind.LocalizedString(printer)
		errs = append(errs, fmt.Sprintf("%s: %s", path, msg))
		return errs
	}

	for _, cause := range ve.Causes {
		errs = append(errs, flattenValidationErrors(cause)...)
	}
	return errs
}
