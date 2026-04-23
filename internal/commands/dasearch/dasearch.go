package dasearch

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/output"
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
	client := ctx.NewMCPClient(dasearchEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "M365_Copilot_Get_Available_Agents", map[string]any{})
	if err != nil {
		return fmt.Errorf("list agents: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}
