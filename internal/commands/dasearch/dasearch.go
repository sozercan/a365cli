package dasearch

import (
	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
)

// DASearchCmd groups Declarative Agent Search subcommands.
type DASearchCmd struct {
	Agents DASearchAgentsCmd `cmd:"" help:"List available M365 Copilot agents (raw DASearch output)"`
}

func dasearchEndpoint() string {
	return config.Endpoint("dasearch")
}

// DASearchAgentsCmd lists available M365 Copilot agents.
type DASearchAgentsCmd struct{}

func (c *DASearchAgentsCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(dasearchEndpoint(), "M365_Copilot_Get_Available_Agents", "list agents", map[string]any{})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}
