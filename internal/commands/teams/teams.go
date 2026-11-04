package teams

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/output"
)

// TeamsCmd groups all Teams subcommands.
type TeamsCmd struct {
	List     TeamsListCmd `cmd:"" help:"List joined teams"`
	Get      TeamsGetCmd  `cmd:"" help:"Get a team by ID"`
	Channels ChannelsCmd  `cmd:"" help:"Team channels"`
	Chats    ChatsCmd     `cmd:"" help:"Team chats"`
	Search   SearchCmd    `cmd:"" help:"Search Teams messages (KQL)"`
	SearchNL SearchNLCmd  `cmd:"" name:"search-nl" help:"Search Teams messages (natural language)"`
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
	// UserID is kept (hidden) only to reject it with a clear message: the
	// ListTeams tool takes no arguments and always lists the signed-in user's
	// teams, so the server fails on any userId argument.
	UserID string `help:"Unsupported: teams can only be listed for the signed-in user." hidden:"" optional:""`
	Max    int    `help:"Maximum number of results" default:"100"`
}

func (c *TeamsListCmd) Run(ctx *commands.Context) error {
	if c.UserID != "" {
		return fmt.Errorf("--user-id is no longer supported: 'teams list' only lists teams for the signed-in user")
	}

	// ListTeams takes no arguments.
	data, err := ctx.CallToolData(teamsEndpoint(), "ListTeams", "list teams", map[string]any{})
	if err != nil {
		return err
	}
	return ctx.Output.PrintListFromData("teams", output.TeamsColumns, data, c.Max)
}

// TeamsGetCmd gets a team by ID.
type TeamsGetCmd struct {
	ID string `arg:"" help:"Team ID"`
}

func (c *TeamsGetCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(teamsEndpoint(), "GetTeam", "get team", map[string]any{
		"teamId": c.ID,
	})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}
