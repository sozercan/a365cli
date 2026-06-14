package nlweb

import (
	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
)

// NLWebCmd groups NLWeb search subcommands.
type NLWebCmd struct {
	Ask   NLWebAskCmd   `cmd:"" help:"Ask a natural language question"`
	Who   NLWebWhoCmd   `cmd:"" help:"Find people related to a query"`
	Sites NLWebSitesCmd `cmd:"" help:"List available NLWeb sites"`
}

func nlwebEndpoint() string {
	return config.Endpoint("nlweb")
}

// NLWebAskCmd asks a natural language question.
type NLWebAskCmd struct {
	Query string `arg:"" help:"Natural language question"`
}

func (c *NLWebAskCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(nlwebEndpoint(), "ask", "ask", map[string]any{
		"query": c.Query,
	})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// NLWebWhoCmd finds people related to a query.
type NLWebWhoCmd struct {
	Query string `arg:"" help:"People search query"`
}

func (c *NLWebWhoCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(nlwebEndpoint(), "who", "who", map[string]any{
		"query": c.Query,
	})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// NLWebSitesCmd lists available NLWeb sites.
type NLWebSitesCmd struct{}

func (c *NLWebSitesCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(nlwebEndpoint(), "list_sites", "list sites", map[string]any{})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}
