package websearch

import (
	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
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
	URLs  []string `arg:"" help:"URLs to search" name:"urls"`
}

func (c *WebSearchSearchCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"query": c.Query,
		"urls":  c.URLs,
	}

	data, err := ctx.CallToolData(websearchEndpoint(), "SearchWeb", "search web", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}
