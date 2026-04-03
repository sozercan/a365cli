package teams

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/output"
)

// TeamsCmd groups all Teams subcommands.
type TeamsCmd struct {
	List     TeamsListCmd     `cmd:"" help:"List joined teams"`
	Get      TeamsGetCmd      `cmd:"" help:"Get a team by ID"`
	Channels ChannelsCmd      `cmd:"" help:"Team channels"`
	Chats    ChatsCmd         `cmd:"" help:"Team chats"`
	Search   SearchCmd        `cmd:"" help:"Search Teams messages (KQL)"`
	SearchNL SearchNLCmd      `cmd:"" name:"search-nl" help:"Search Teams messages (natural language)"`
}

// teamsEndpoint returns the agent365 endpoint for the Teams MCP server.
func teamsEndpoint() string {
	ep := config.Endpoint("teams")
	if ep == "" {
		panic("teams server not configured")
	}
	return ep
}

// TeamsListCmd lists joined teams.
type TeamsListCmd struct {
	UserID string `help:"User ID (GUID) to list teams for. Defaults to signed-in user." optional:""`
	Max    int    `help:"Maximum number of results" default:"100"`
}

func (c *TeamsListCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(teamsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	args := map[string]any{}
	if c.UserID != "" {
		args["userId"] = c.UserID
	} else if ctx.UserUPN != "" {
		args["userId"] = ctx.UserUPN
	}

	resp, err := client.CallTool(ctx.Ctx, "ListTeams", args)
	if err != nil {
		return fmt.Errorf("list teams: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	rows := output.ToRows(data, "teams")
	if c.Max > 0 && len(rows) > c.Max {
		rows = rows[:c.Max]
	}
	return ctx.Output.PrintList("teams", output.TeamsColumns, rows)
}

// TeamsGetCmd gets a team by ID.
type TeamsGetCmd struct {
	ID string `arg:"" help:"Team ID"`
}

func (c *TeamsGetCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(teamsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "GetTeam", map[string]any{
		"teamId": c.ID,
	})
	if err != nil {
		return fmt.Errorf("get team: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}
