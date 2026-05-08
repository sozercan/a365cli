package calendar

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	cmdexec "github.com/sozercan/a365cli/internal/commands/exec"
	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/output"
)

// CalendarCmd groups all Calendar subcommands.
type CalendarCmd struct {
	List      CalListCmd      `cmd:"" help:"List upcoming events"`
	View      CalViewCmd      `cmd:"" help:"List events in a date range"`
	Create    CalCreateCmd    `cmd:"" help:"Create a calendar event"`
	Update    CalUpdateCmd    `cmd:"" help:"Update a calendar event"`
	Delete    CalDeleteCmd    `cmd:"" help:"Delete a calendar event"`
	Accept    CalAcceptCmd    `cmd:"" help:"Accept a meeting invite"`
	Tentative CalTentativeCmd `cmd:"" help:"Tentatively accept a meeting invite"`
	Decline   CalDeclineCmd   `cmd:"" help:"Decline a meeting invite"`
	Cancel    CalCancelCmd    `cmd:"" help:"Cancel a meeting you organized"`
	Forward   CalForwardCmd   `cmd:"" help:"Forward a meeting invite"`
	FreeBusy  CalFreeBusyCmd  `cmd:"" name:"free-busy" help:"Find available meeting times"`
	TimeZone  CalTimeZoneCmd  `cmd:"" name:"timezone" help:"Get user date/time zone settings"`
	Rooms     CalRoomsCmd     `cmd:"" help:"List available rooms"`
}

func calEndpoint() string {
	return config.Endpoint("calendar")
}

// CalListCmd lists upcoming events.
type CalListCmd struct {
	Max int `help:"Maximum number of events" default:"20"`
}

func (c *CalListCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(calEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "ListEvents", map[string]any{})
	if err != nil {
		return fmt.Errorf("list events: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	rows := output.ToRows(data, "events")
	if rows == nil {
		rows = output.ToRows(data, "value")
	}
	if rows == nil {
		return ctx.Output.PrintItem(data)
	}
	if c.Max > 0 && len(rows) > c.Max {
		rows = rows[:c.Max]
	}
	return ctx.Output.PrintList("events", output.CalendarColumns, rows)
}

// CalViewCmd lists events in a date range.
type CalViewCmd struct {
	Max int `help:"Maximum number of events" default:"50"`
}

func (c *CalViewCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(calEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "ListCalendarView", map[string]any{})
	if err != nil {
		return fmt.Errorf("list calendar view: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	rows := output.ToRows(data, "events")
	if rows == nil {
		rows = output.ToRows(data, "value")
	}
	if rows == nil {
		return ctx.Output.PrintItem(data)
	}
	if c.Max > 0 && len(rows) > c.Max {
		rows = rows[:c.Max]
	}
	return ctx.Output.PrintList("events", output.CalendarColumns, rows)
}

// CalCreateCmd creates a calendar event.
type CalCreateCmd struct {
	Subject   string   `arg:"" help:"Event subject"`
	Start     string   `help:"Start time (ISO 8601, e.g. 2025-01-15T10:00:00)" required:""`
	End       string   `help:"End time (ISO 8601)" required:""`
	Attendees []string `help:"Attendee email addresses" name:"attendee" optional:""`
	Body      string   `help:"Event body/description" optional:""`
	IsOnline  bool     `help:"Add Teams meeting link" name:"teams" default:"false"`
}

func (c *CalCreateCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		mcpArgs := map[string]any{
			"subject":        c.Subject,
			"startDateTime":  c.Start,
			"endDateTime":    c.End,
			"attendeeEmails": c.Attendees,
		}
		if c.Body != "" {
			mcpArgs["body"] = c.Body
		}
		if c.IsOnline {
			mcpArgs["isOnlineMeeting"] = true
		}
		return ctx.ValidateDryRun(calEndpoint(), "CreateEvent",
			fmt.Sprintf("create event %q from %s to %s", c.Subject, c.Start, c.End),
			map[string]any{
				"action": "calendar.create", "subject": c.Subject,
				"start": c.Start, "end": c.End, "attendees": c.Attendees,
			},
			mcpArgs,
		)
	}

	client := ctx.NewMCPClient(calEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	args := map[string]any{
		"subject":        c.Subject,
		"startDateTime":  c.Start,
		"endDateTime":    c.End,
		"attendeeEmails": c.Attendees,
	}
	if c.Body != "" {
		args["body"] = c.Body
	}
	if c.IsOnline {
		args["isOnlineMeeting"] = true
	}

	resp, err := client.CallTool(ctx.Ctx, "CreateEvent", args)
	if err != nil {
		return fmt.Errorf("create event: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Event created", data)
}

// CalUpdateCmd updates a calendar event.
type CalUpdateCmd struct {
	ID      string `arg:"" help:"Event ID"`
	Subject string `help:"New subject" optional:""`
	Start   string `help:"New start time" optional:""`
	End     string `help:"New end time" optional:""`
	Body    string `help:"New body" optional:""`
}

func (c *CalUpdateCmd) Run(ctx *commands.Context) error {
	args := map[string]any{"eventId": c.ID}
	if c.Subject != "" {
		args["subject"] = c.Subject
	}
	if c.Start != "" {
		args["startDateTime"] = c.Start
	}
	if c.End != "" {
		args["endDateTime"] = c.End
	}
	if c.Body != "" {
		args["body"] = c.Body
	}

	return cmdexec.New(ctx).Mutate(cmdexec.OperationPlan{
		Service:   "calendar",
		Tool:      "UpdateEvent",
		Args:      args,
		Action:    fmt.Sprintf("update event %s", c.ID),
		Display:   map[string]any{"action": "calendar.update", "eventId": c.ID},
		ErrPrefix: "update event",
		Output:    cmdexec.Mutation("Event updated"),
	})
}

// CalDeleteCmd deletes a calendar event.
type CalDeleteCmd struct {
	ID string `arg:"" help:"Event ID"`
}

func (c *CalDeleteCmd) Run(ctx *commands.Context) error {
	args := map[string]any{"eventId": c.ID}

	if ctx.DryRun {
		return ctx.ValidateDryRun(calEndpoint(), "DeleteEventById", fmt.Sprintf("delete event %s", c.ID),
			map[string]any{"action": "calendar.delete", "eventId": c.ID},
			args,
		)
	}
	if err := ctx.Confirm(fmt.Sprintf("delete event %s", c.ID)); err != nil {
		return err
	}

	client := ctx.NewMCPClient(calEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "DeleteEventById", args)
	if err != nil {
		return fmt.Errorf("delete event: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Event deleted", data)
}

// CalAcceptCmd accepts a meeting invite.
type CalAcceptCmd struct {
	ID string `arg:"" help:"Event ID"`
}

func (c *CalAcceptCmd) Run(ctx *commands.Context) error {
	args := map[string]any{"eventId": c.ID}

	if ctx.DryRun {
		return ctx.ValidateDryRun(calEndpoint(), "AcceptEvent", fmt.Sprintf("accept event %s", c.ID),
			map[string]any{"action": "calendar.accept", "eventId": c.ID},
			args,
		)
	}

	client := ctx.NewMCPClient(calEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "AcceptEvent", args)
	if err != nil {
		return fmt.Errorf("accept: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Event accepted", data)
}

// CalTentativeCmd tentatively accepts a meeting.
type CalTentativeCmd struct {
	ID string `arg:"" help:"Event ID"`
}

func (c *CalTentativeCmd) Run(ctx *commands.Context) error {
	args := map[string]any{"eventId": c.ID}

	if ctx.DryRun {
		return ctx.ValidateDryRun(calEndpoint(), "TentativelyAcceptEvent", fmt.Sprintf("tentatively accept event %s", c.ID),
			map[string]any{"action": "calendar.tentative", "eventId": c.ID},
			args,
		)
	}

	client := ctx.NewMCPClient(calEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "TentativelyAcceptEvent", args)
	if err != nil {
		return fmt.Errorf("tentative accept: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Event tentatively accepted", data)
}

// CalDeclineCmd declines a meeting.
type CalDeclineCmd struct {
	ID string `arg:"" help:"Event ID"`
}

func (c *CalDeclineCmd) Run(ctx *commands.Context) error {
	args := map[string]any{"eventId": c.ID}

	if ctx.DryRun {
		return ctx.ValidateDryRun(calEndpoint(), "DeclineEvent", fmt.Sprintf("decline event %s", c.ID),
			map[string]any{"action": "calendar.decline", "eventId": c.ID},
			args,
		)
	}

	client := ctx.NewMCPClient(calEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "DeclineEvent", args)
	if err != nil {
		return fmt.Errorf("decline: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Event declined", data)
}

// CalCancelCmd cancels a meeting you organized.
type CalCancelCmd struct {
	ID string `arg:"" help:"Event ID"`
}

func (c *CalCancelCmd) Run(ctx *commands.Context) error {
	args := map[string]any{"eventId": c.ID}

	if ctx.DryRun {
		return ctx.ValidateDryRun(calEndpoint(), "CancelEvent", fmt.Sprintf("cancel event %s", c.ID),
			map[string]any{"action": "calendar.cancel", "eventId": c.ID},
			args,
		)
	}
	if err := ctx.Confirm(fmt.Sprintf("cancel event %s", c.ID)); err != nil {
		return err
	}

	client := ctx.NewMCPClient(calEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "CancelEvent", args)
	if err != nil {
		return fmt.Errorf("cancel: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Event cancelled", data)
}

// CalForwardCmd forwards a meeting invite.
type CalForwardCmd struct {
	ID         string   `arg:"" help:"Event ID"`
	Recipients []string `arg:"" help:"Recipient email addresses"`
	Comment    string   `help:"Comment" optional:""`
}

func (c *CalForwardCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		mcpArgs := map[string]any{"eventId": c.ID, "recipientEmails": c.Recipients}
		if c.Comment != "" {
			mcpArgs["comment"] = c.Comment
		}
		return ctx.ValidateDryRun(calEndpoint(), "ForwardEvent", fmt.Sprintf("forward event %s to %v", c.ID, c.Recipients),
			map[string]any{"action": "calendar.forward", "eventId": c.ID, "to": c.Recipients},
			mcpArgs,
		)
	}

	client := ctx.NewMCPClient(calEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	args := map[string]any{"eventId": c.ID, "recipientEmails": c.Recipients}
	if c.Comment != "" {
		args["comment"] = c.Comment
	}

	resp, err := client.CallTool(ctx.Ctx, "ForwardEvent", args)
	if err != nil {
		return fmt.Errorf("forward: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Event forwarded", data)
}

// CalFreeBusyCmd finds available meeting times.
type CalFreeBusyCmd struct{}

func (c *CalFreeBusyCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(calEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "FindMeetingTimes", map[string]any{})
	if err != nil {
		return fmt.Errorf("find meeting times: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// CalTimeZoneCmd gets user date/time settings.
type CalTimeZoneCmd struct{}

func (c *CalTimeZoneCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(calEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "GetUserDateAndTimeZoneSettings", map[string]any{})
	if err != nil {
		return fmt.Errorf("get timezone: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// CalRoomsCmd lists available rooms.
type CalRoomsCmd struct{}

func (c *CalRoomsCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(calEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "GetRooms", map[string]any{})
	if err != nil {
		return fmt.Errorf("get rooms: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}
