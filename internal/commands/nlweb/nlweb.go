package nlweb

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/output"
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
	client := ctx.NewMCPClient(nlwebEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "ask", map[string]any{
		"query": c.Query,
	})
	if err != nil {
		return fmt.Errorf("ask: %w", err)
	}

	data, err := output.ExtractContent(resp)
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
	client := ctx.NewMCPClient(nlwebEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "who", map[string]any{
		"query": c.Query,
	})
	if err != nil {
		return fmt.Errorf("who: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// NLWebSitesCmd lists available NLWeb sites.
type NLWebSitesCmd struct{}

func (c *NLWebSitesCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(nlwebEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "list_sites", map[string]any{})
	if err != nil {
		return fmt.Errorf("list sites: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}
