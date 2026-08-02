package triggers

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
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
	data, err := ctx.CallToolData(triggersEndpoint(), "list_event_types", "list event types", map[string]any{})
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
	data, err := ctx.CallToolData(triggersEndpoint(), "get_event_type_schema", "get event type schema", map[string]any{
		"eventType": c.EventType,
	})
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
	data, err := ctx.CallToolData(triggersEndpoint(), "validate_trigger", "validate trigger", map[string]any{
		"userRequest": c.UserRequest,
	})
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
		mcpArgs := map[string]any{
			"validationToken": c.ValidationToken,
			"name":            c.Name,
			"eventType":       c.EventType,
			"logic":           c.Logic,
			"conditions":      c.Conditions,
			"instructions":    c.Instructions,
		}
		return ctx.ValidateDryRun(triggersEndpoint(), "create_trigger_definition", fmt.Sprintf("create trigger %q", c.Name),
			map[string]any{
				"action":    "triggers.create",
				"name":      c.Name,
				"eventType": c.EventType,
			},
			mcpArgs,
		)
	}

	data, err := ctx.CallToolData(triggersEndpoint(), "create_trigger_definition", "create trigger", map[string]any{
		"validationToken": c.ValidationToken,
		"name":            c.Name,
		"eventType":       c.EventType,
		"logic":           c.Logic,
		"conditions":      c.Conditions,
		"instructions":    c.Instructions,
	})
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Trigger created", data)
}

// --- list_trigger_definitions ---

// TriggersListCmd lists trigger definitions.
type TriggersListCmd struct{}

func (c *TriggersListCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(triggersEndpoint(), "list_trigger_definitions", "list triggers", map[string]any{})
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
	data, err := ctx.CallToolData(triggersEndpoint(), "get_trigger_definition", "get trigger", map[string]any{
		"id": c.ID,
	})
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
		mcpArgs := map[string]any{
			"validationToken": c.ValidationToken,
			"id":              c.ID,
		}
		return ctx.ValidateDryRun(triggersEndpoint(), "update_trigger_definition", fmt.Sprintf("update trigger %s", c.ID),
			map[string]any{
				"action": "triggers.update",
				"id":     c.ID,
			},
			mcpArgs,
		)
	}

	data, err := ctx.CallToolData(triggersEndpoint(), "update_trigger_definition", "update trigger", map[string]any{
		"validationToken": c.ValidationToken,
		"id":              c.ID,
	})
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
		return ctx.ValidateDryRun(triggersEndpoint(), "delete_trigger_definition", fmt.Sprintf("delete trigger %s", c.ID),
			map[string]any{
				"action": "triggers.delete",
				"id":     c.ID,
			})
	}

	if err := ctx.Confirm(fmt.Sprintf("delete trigger %s", c.ID)); err != nil {
		return err
	}

	data, err := ctx.CallToolData(triggersEndpoint(), "delete_trigger_definition", "delete trigger", map[string]any{
		"id": c.ID,
	})
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
	data, err := ctx.CallToolData(triggersEndpoint(), "evaluate_event_triggers", "evaluate triggers", map[string]any{
		"eventType":     c.EventType,
		"eventDataJson": c.EventDataJSON,
	})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}
