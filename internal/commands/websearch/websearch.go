package websearch

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/output"
)

// WebSearchCmd groups Web Search subcommands.
type WebSearchCmd struct {
	Search WebSearchSearchCmd `cmd:"" help:"Search the web"`
}

func websearchEndpoint() string {
	return config.Endpoint("websearch")
}

// WebSearchSearchCmd searches the web.
type WebSearchSearchCmd struct {
	Query string   `arg:"" help:"Search query"`
	URLs  []string `help:"URLs to search" name:"urls" optional:""`
}

func (c *WebSearchSearchCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(websearchEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	args := map[string]any{
		"query": c.Query,
		"urls":  c.URLs,
	}

	resp, err := client.CallTool(ctx.Ctx, "SearchWeb", args)
	if err != nil {
		return fmt.Errorf("search web: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}
