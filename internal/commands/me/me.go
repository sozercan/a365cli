package me

import (
	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/output"
)

// MeCmd groups user/profile subcommands.
type MeCmd struct {
	Whoami  MeWhoamiCmd  `cmd:"" help:"Get your own profile details"`
	Get     MeGetCmd     `cmd:"" help:"Get a user's details by UPN or ID"`
	Search  MeSearchCmd  `cmd:"" help:"Search for multiple users"`
	Manager MeManagerCmd `cmd:"" help:"Get a user's manager"`
	Reports MeReportsCmd `cmd:"" help:"Get a user's direct reports"`
}

func meEndpoint() string {
	return config.Endpoint("me")
}

// MeWhoamiCmd gets your own profile.
type MeWhoamiCmd struct{}

func (c *MeWhoamiCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(meEndpoint(), "GetMyDetails", "get my details", map[string]any{})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// MeGetCmd gets a user's details.
type MeGetCmd struct {
	User string `arg:"" help:"User UPN (email) or ID"`
}

func (c *MeGetCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(meEndpoint(), "GetUserDetails", "get user", map[string]any{
		"userIdentifier": c.User,
	})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// MeSearchCmd searches for multiple users.
type MeSearchCmd struct {
	Query []string `arg:"" help:"Search values (names, emails, or IDs)"`
}

func (c *MeSearchCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(meEndpoint(), "GetMultipleUsersDetails", "search users", map[string]any{
		"searchValues": c.Query,
	})
	if err != nil {
		return err
	}

	rows := output.ToRows(data, "users")
	if rows == nil {
		rows = output.ToRows(data, "value")
	}
	if rows == nil {
		return ctx.Output.PrintItem(data)
	}
	return ctx.Output.PrintList("users", output.UserColumns, rows)
}

// MeManagerCmd gets a user's manager.
type MeManagerCmd struct {
	UserID string `arg:"" help:"User ID (GUID)"`
}

func (c *MeManagerCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(meEndpoint(), "GetManagerDetails", "get manager", map[string]any{
		"userId": c.UserID,
	})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// MeReportsCmd gets a user's direct reports.
type MeReportsCmd struct {
	UserID string `arg:"" help:"User ID (GUID)"`
}

func (c *MeReportsCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(meEndpoint(), "GetDirectReportsDetails", "get reports", map[string]any{
		"userId": c.UserID,
	})
	if err != nil {
		return err
	}

	rows := output.ToRows(data, "directReports")
	if rows == nil {
		rows = output.ToRows(data, "value")
	}
	if rows == nil {
		return ctx.Output.PrintItem(data)
	}
	return ctx.Output.PrintList("directReports", output.UserColumns, rows)
}
