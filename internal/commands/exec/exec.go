// Package exec centralizes MCP command execution for CLI command handlers.
package exec

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/output"
)

// Executor runs MCP tool calls on behalf of command handlers.
type Executor struct {
	ctx *commands.Context
}

// New returns an Executor bound to the command context.
func New(ctx *commands.Context) *Executor {
	return &Executor{ctx: ctx}
}

// ToolCall describes a read-only MCP tool call.
type ToolCall struct {
	// Service is resolved through config.Endpoint when Endpoint is empty.
	Service string
	// Endpoint overrides Service endpoint resolution when set.
	Endpoint string
	Tool     string
	Args     map[string]any
	// ErrPrefix is used when wrapping CallTool errors.
	ErrPrefix string
	Output    OutputSpec
}

// OperationPlan describes a write MCP operation, including safety metadata.
type OperationPlan struct {
	// Service is resolved through config.Endpoint when Endpoint is empty.
	Service string
	// Endpoint overrides Service endpoint resolution when set.
	Endpoint string
	Tool     string
	Args     map[string]any

	// Display is safe metadata shown in dry-run output. Args are used for
	// schema validation and execution, so dry-run and execution cannot drift.
	Display map[string]any
	Action  string

	Destructive bool
	ConfirmText string

	// ErrPrefix is used when wrapping CallTool errors.
	ErrPrefix string
	Output    OutputSpec
}

// OutputKind selects how domain data should be rendered.
type OutputKind int

const (
	// OutputItem renders the full response object.
	OutputItem OutputKind = iota
	// OutputList renders an extracted collection as a table/array.
	OutputList
	// OutputMutation renders a mutation status message and response payload.
	OutputMutation
)

// OutputSpec describes how extracted MCP content should be rendered.
type OutputSpec struct {
	Kind OutputKind

	Entity         string
	Columns        []output.Column
	CollectionKeys []string
	Max            int
	FallbackToItem bool

	Message string
}

// Item renders extracted content as an item.
func Item() OutputSpec {
	return OutputSpec{Kind: OutputItem}
}

// List renders extracted content as a list. The first matching collection key
// is used; when no key is supplied, entity is used as the collection key.
func List(entity string, columns []output.Column, keys ...string) OutputSpec {
	if len(keys) == 0 {
		keys = []string{entity}
	}
	return OutputSpec{
		Kind:           OutputList,
		Entity:         entity,
		Columns:        columns,
		CollectionKeys: append([]string(nil), keys...),
		FallbackToItem: true,
	}
}

// Mutation renders extracted content as a mutation result.
func Mutation(message string) OutputSpec {
	return OutputSpec{Kind: OutputMutation, Message: message}
}

// WithMax returns a copy of the output spec with a row limit.
func (s OutputSpec) WithMax(max int) OutputSpec {
	s.Max = max
	return s
}

// WithFallbackToItem returns a copy of the output spec with list fallback
// behavior configured.
func (s OutputSpec) WithFallbackToItem(enabled bool) OutputSpec {
	s.FallbackToItem = enabled
	return s
}

// Query executes a read-only MCP tool call and renders its response.
func (e *Executor) Query(call ToolCall) error {
	if err := validateTool(call.Tool); err != nil {
		return err
	}
	endpoint, err := resolveEndpoint(call.Service, call.Endpoint)
	if err != nil {
		return err
	}
	data, err := e.callTool(endpoint, call.Tool, call.Args, call.ErrPrefix)
	if err != nil {
		return err
	}
	return e.Render(data, call.Output)
}

// Mutate executes a write MCP operation. Dry-run validation, destructive
// confirmation, tool execution, and output rendering all use the same Args map.
func (e *Executor) Mutate(plan OperationPlan) error {
	if err := validateTool(plan.Tool); err != nil {
		return err
	}
	endpoint, err := resolveEndpoint(plan.Service, plan.Endpoint)
	if err != nil {
		return err
	}

	action := plan.Action
	if action == "" {
		action = plan.Tool
	}
	display := plan.Display
	if display == nil {
		display = map[string]any{"action": action}
	}
	args := plan.Args
	if args == nil {
		args = map[string]any{}
	}

	if e.ctx.DryRun {
		return e.ctx.ValidateDryRun(endpoint, plan.Tool, action, display, args)
	}

	if plan.Destructive {
		confirmText := plan.ConfirmText
		if confirmText == "" {
			confirmText = action
		}
		if err := e.ctx.Confirm(confirmText); err != nil {
			return err
		}
	}

	data, err := e.callTool(endpoint, plan.Tool, args, plan.ErrPrefix)
	if err != nil {
		return err
	}
	return e.Render(data, plan.Output)
}

// Render renders already-extracted MCP content according to an output spec.
func (e *Executor) Render(data map[string]any, spec OutputSpec) error {
	switch spec.Kind {
	case OutputList:
		for _, key := range spec.CollectionKeys {
			rows := output.ToRows(data, key)
			if rows == nil {
				continue
			}
			if spec.Max > 0 && len(rows) > spec.Max {
				rows = rows[:spec.Max]
			}
			return e.ctx.Output.PrintList(spec.Entity, spec.Columns, rows)
		}
		if spec.FallbackToItem {
			return e.ctx.Output.PrintItem(data)
		}
		return fmt.Errorf("response did not contain any of the expected collection keys: %v", spec.CollectionKeys)
	case OutputMutation:
		return e.ctx.Output.PrintMutation(spec.Message, data)
	case OutputItem:
		fallthrough
	default:
		return e.ctx.Output.PrintItem(data)
	}
}

func (e *Executor) callTool(endpoint, tool string, args map[string]any, errPrefix string) (map[string]any, error) {
	client := e.ctx.NewMCPClient(endpoint)
	if err := client.Initialize(e.ctx.Ctx); err != nil {
		return nil, fmt.Errorf("initialize: %w", err)
	}

	if args == nil {
		args = map[string]any{}
	}
	resp, err := client.CallTool(e.ctx.Ctx, tool, args)
	if err != nil {
		if errPrefix == "" {
			errPrefix = fmt.Sprintf("call %s", tool)
		}
		return nil, fmt.Errorf("%s: %w", errPrefix, err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func resolveEndpoint(service, endpoint string) (string, error) {
	if endpoint != "" {
		return endpoint, nil
	}
	if service == "" {
		return "", fmt.Errorf("service or endpoint is required")
	}
	endpoint = config.Endpoint(service)
	if endpoint == "" {
		return "", fmt.Errorf("unknown service %q", service)
	}
	return endpoint, nil
}

func validateTool(tool string) error {
	if tool == "" {
		return fmt.Errorf("tool is required")
	}
	return nil
}
