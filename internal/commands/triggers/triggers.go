package triggers

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/output"
)

// TriggersCmd groups all Triggers subcommands.
type TriggersCmd struct {
	Events   TriggersEventsCmd   `cmd:"" help:"List supported event types"`
	Schema   TriggersSchemaCmd   `cmd:"" help:"Get schema for an event type"`
	Validate TriggersValidateCmd `cmd:"" help:"Validate a trigger request"`
	Create   TriggersCreateCmd   `cmd:"" help:"Create a trigger definition"`
	List     TriggersListCmd     `cmd:"" help:"List trigger definitions"`
	Get      TriggersGetCmd      `cmd:"" help:"Get a trigger definition"`
	Update   TriggersUpdateCmd   `cmd:"" help:"Update a trigger definition"`
	Delete   TriggersDeleteCmd   `cmd:"" help:"Delete a trigger definition"`
	Evaluate TriggersEvaluateCmd `cmd:"" help:"Evaluate event against triggers"`
}

func triggersEndpoint() string {
	return config.Endpoint("tasks")
}

// --- list_event_types ---

// TriggersEventsCmd lists supported event types.
type TriggersEventsCmd struct{}

func (c *TriggersEventsCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(triggersEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "list_event_types", map[string]any{})
	if err != nil {
		return fmt.Errorf("list event types: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- get_event_type_schema ---

// TriggersSchemaCmd gets the schema for an event type.
type TriggersSchemaCmd struct {
	EventType string `arg:"" help:"Event type to get schema for"`
}

func (c *TriggersSchemaCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(triggersEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "get_event_type_schema", map[string]any{
		"eventType": c.EventType,
	})
	if err != nil {
		return fmt.Errorf("get event type schema: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- validate_trigger ---

// TriggersValidateCmd validates a trigger request.
type TriggersValidateCmd struct {
	UserRequest string `arg:"" help:"User request to validate"`
}

func (c *TriggersValidateCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(triggersEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "validate_trigger", map[string]any{
		"userRequest": c.UserRequest,
	})
	if err != nil {
		return fmt.Errorf("validate trigger: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- create_trigger_definition ---

// TriggersCreateCmd creates a trigger definition.
type TriggersCreateCmd struct {
	ValidationToken string `arg:"" help:"Validation token from validate step"`
	Name            string `arg:"" help:"Trigger name"`
	EventType       string `arg:"" help:"Event type"`
	Logic           string `arg:"" help:"Trigger logic"`
	Conditions      string `arg:"" help:"Trigger conditions (JSON)"`
	Instructions    string `arg:"" help:"Trigger instructions"`
}

func (c *TriggersCreateCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.Output.PrintDryRun(fmt.Sprintf("create trigger %q", c.Name),
			map[string]any{
				"action":    "triggers.create",
				"name":      c.Name,
				"eventType": c.EventType,
			})
	}

	client := ctx.NewMCPClient(triggersEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "create_trigger_definition", map[string]any{
		"validationToken": c.ValidationToken,
		"name":            c.Name,
		"eventType":       c.EventType,
		"logic":           c.Logic,
		"conditions":      c.Conditions,
		"instructions":    c.Instructions,
	})
	if err != nil {
		return fmt.Errorf("create trigger: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Trigger created", data)
}

// --- list_trigger_definitions ---

// TriggersListCmd lists trigger definitions.
type TriggersListCmd struct{}

func (c *TriggersListCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(triggersEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "list_trigger_definitions", map[string]any{})
	if err != nil {
		return fmt.Errorf("list triggers: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- get_trigger_definition ---

// TriggersGetCmd gets a trigger definition by ID.
type TriggersGetCmd struct {
	ID string `arg:"" help:"Trigger definition ID"`
}

func (c *TriggersGetCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(triggersEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "get_trigger_definition", map[string]any{
		"id": c.ID,
	})
	if err != nil {
		return fmt.Errorf("get trigger: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- update_trigger_definition ---

// TriggersUpdateCmd updates a trigger definition.
type TriggersUpdateCmd struct {
	ValidationToken string `arg:"" help:"Validation token"`
	ID              string `arg:"" help:"Trigger definition ID"`
}

func (c *TriggersUpdateCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.Output.PrintDryRun(fmt.Sprintf("update trigger %s", c.ID),
			map[string]any{
				"action": "triggers.update",
				"id":     c.ID,
			})
	}

	client := ctx.NewMCPClient(triggersEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "update_trigger_definition", map[string]any{
		"validationToken": c.ValidationToken,
		"id":              c.ID,
	})
	if err != nil {
		return fmt.Errorf("update trigger: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Trigger updated", data)
}

// --- delete_trigger_definition ---

// TriggersDeleteCmd deletes a trigger definition.
type TriggersDeleteCmd struct {
	ID string `arg:"" help:"Trigger definition ID"`
}

func (c *TriggersDeleteCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.Output.PrintDryRun(fmt.Sprintf("delete trigger %s", c.ID),
			map[string]any{
				"action": "triggers.delete",
				"id":     c.ID,
			})
	}

	if err := ctx.Confirm(fmt.Sprintf("delete trigger %s", c.ID)); err != nil {
		return err
	}

	client := ctx.NewMCPClient(triggersEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "delete_trigger_definition", map[string]any{
		"id": c.ID,
	})
	if err != nil {
		return fmt.Errorf("delete trigger: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Trigger deleted", data)
}

// --- evaluate_event_triggers ---

// TriggersEvaluateCmd evaluates an event against triggers.
type TriggersEvaluateCmd struct {
	EventType     string `arg:"" help:"Event type"`
	EventDataJSON string `arg:"" help:"Event data as JSON string"`
}

func (c *TriggersEvaluateCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(triggersEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "evaluate_event_triggers", map[string]any{
		"eventType":     c.EventType,
		"eventDataJson": c.EventDataJSON,
	})
	if err != nil {
		return fmt.Errorf("evaluate triggers: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}
