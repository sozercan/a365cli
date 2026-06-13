package calendar

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
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
	data, err := ctx.CallToolData(calEndpoint(), "ListEvents", "list events", map[string]any{})
	if err != nil {
		return err
	}
	return ctx.Output.PrintListFromData("events", output.CalendarColumns, data, c.Max, "events", "value")
}

// CalViewCmd lists events in a date range.
type CalViewCmd struct {
	Max int `help:"Maximum number of events" default:"50"`
}

func (c *CalViewCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(calEndpoint(), "ListCalendarView", "list calendar view", map[string]any{})
	if err != nil {
		return err
	}
	return ctx.Output.PrintListFromData("events", output.CalendarColumns, data, c.Max, "events", "value")
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

	data, err := ctx.CallToolData(calEndpoint(), "CreateEvent", "create event", args)
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
	if ctx.DryRun {
		return ctx.ValidateDryRun(calEndpoint(), "UpdateEvent",
			fmt.Sprintf("update event %s", c.ID),
			map[string]any{"action": "calendar.update", "eventId": c.ID},
		)
	}

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

	data, err := ctx.CallToolData(calEndpoint(), "UpdateEvent", "update event", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Event updated", data)
}

// CalDeleteCmd deletes a calendar event.
type CalDeleteCmd struct {
	ID string `arg:"" help:"Event ID"`
}

func (c *CalDeleteCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(calEndpoint(), "DeleteEventById", fmt.Sprintf("delete event %s", c.ID),
			map[string]any{"action": "calendar.delete", "eventId": c.ID})
	}
	if err := ctx.Confirm(fmt.Sprintf("delete event %s", c.ID)); err != nil {
		return err
	}

	data, err := ctx.CallToolData(calEndpoint(), "DeleteEventById", "delete event", map[string]any{"eventId": c.ID})
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
	if ctx.DryRun {
		return ctx.ValidateDryRun(calEndpoint(), "AcceptEvent", fmt.Sprintf("accept event %s", c.ID),
			map[string]any{"action": "calendar.accept", "eventId": c.ID})
	}

	data, err := ctx.CallToolData(calEndpoint(), "AcceptEvent", "accept", map[string]any{"eventId": c.ID})
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
	if ctx.DryRun {
		return ctx.ValidateDryRun(calEndpoint(), "TentativelyAcceptEvent", fmt.Sprintf("tentatively accept event %s", c.ID),
			map[string]any{"action": "calendar.tentative", "eventId": c.ID})
	}

	data, err := ctx.CallToolData(calEndpoint(), "TentativelyAcceptEvent", "tentative accept", map[string]any{"eventId": c.ID})
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
	if ctx.DryRun {
		return ctx.ValidateDryRun(calEndpoint(), "DeclineEvent", fmt.Sprintf("decline event %s", c.ID),
			map[string]any{"action": "calendar.decline", "eventId": c.ID})
	}

	data, err := ctx.CallToolData(calEndpoint(), "DeclineEvent", "decline", map[string]any{"eventId": c.ID})
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
	if ctx.DryRun {
		return ctx.ValidateDryRun(calEndpoint(), "CancelEvent", fmt.Sprintf("cancel event %s", c.ID),
			map[string]any{"action": "calendar.cancel", "eventId": c.ID})
	}
	if err := ctx.Confirm(fmt.Sprintf("cancel event %s", c.ID)); err != nil {
		return err
	}

	data, err := ctx.CallToolData(calEndpoint(), "CancelEvent", "cancel", map[string]any{"eventId": c.ID})
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

	args := map[string]any{"eventId": c.ID, "recipientEmails": c.Recipients}
	if c.Comment != "" {
		args["comment"] = c.Comment
	}

	data, err := ctx.CallToolData(calEndpoint(), "ForwardEvent", "forward", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Event forwarded", data)
}

// CalFreeBusyCmd finds available meeting times.
type CalFreeBusyCmd struct{}

func (c *CalFreeBusyCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(calEndpoint(), "FindMeetingTimes", "find meeting times", map[string]any{})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// CalTimeZoneCmd gets user date/time settings.
type CalTimeZoneCmd struct{}

func (c *CalTimeZoneCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(calEndpoint(), "GetUserDateAndTimeZoneSettings", "get timezone", map[string]any{})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// CalRoomsCmd lists available rooms.
type CalRoomsCmd struct{}

func (c *CalRoomsCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(calEndpoint(), "GetRooms", "get rooms", map[string]any{})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}
