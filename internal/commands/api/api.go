package api

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/output"
)

// APICmd groups API explorer subcommands (dev/debug tool).
type APICmd struct {
	Servers  APIServersCmd  `cmd:"" help:"List all MCP servers and their tool counts"`
	Discover APIDiscoverCmd `cmd:"" help:"Discover available MCP servers from agent365 gateway"`
	Tools    APIToolsCmd    `cmd:"" help:"List tools and schemas for a service"`
	Call     APICallCmd     `cmd:"" help:"Call an MCP tool directly with raw JSON arguments"`
}

// APIServersCmd lists all known MCP servers and probes for tool counts.
type APIServersCmd struct {
	Probe bool `help:"Connect to each server and count tools (slow)" default:"false"`
}

func (c *APIServersCmd) Run(ctx *commands.Context) error {
	servers := config.Servers
	names := make([]string, 0, len(servers))
	for name := range servers {
		names = append(names, name)
	}
	sort.Strings(names)

	if !c.Probe {
		rows := make([]map[string]any, 0, len(names))
		for _, name := range names {
			rows = append(rows, map[string]any{
				"service": name,
				"server":  servers[name],
			})
		}
		return ctx.Output.PrintList("servers", output.APIServerColumns, rows)
	}

	// Probe each server for tool counts
	rows := make([]map[string]any, 0, len(names))
	for _, name := range names {
		endpoint := config.Endpoint(name)
		client := ctx.NewMCPClient(endpoint)

		toolCount := 0
		status := "ok"
		if err := client.Initialize(ctx.Ctx); err != nil {
			status = "error: " + err.Error()
		} else {
			resp, err := client.ListTools(ctx.Ctx)
			if err != nil {
				status = "error: " + err.Error()
			} else if resp.Result != nil {
				toolCount = len(resp.Result.Tools)
			}
		}

		rows = append(rows, map[string]any{
			"service": name,
			"server":  servers[name],
			"tools":   toolCount,
			"status":  status,
		})
	}
	return ctx.Output.PrintList("servers", output.APIServerProbeColumns, rows)
}

// APIToolsCmd lists tools for a specific service.
type APIToolsCmd struct {
	Service string `arg:"" help:"Service name (e.g. teams, mail, calendar)"`
}

func (c *APIToolsCmd) Run(ctx *commands.Context) error {
	endpoint := config.Endpoint(c.Service)
	if endpoint == "" {
		return fmt.Errorf("unknown service %q — use 'a365 api servers' to see available services", c.Service)
	}

	client := ctx.NewMCPClient(endpoint)
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.ListTools(ctx.Ctx)
	if err != nil {
		return fmt.Errorf("list tools: %w", err)
	}

	if resp.Result == nil || len(resp.Result.Tools) == 0 {
		fmt.Fprintf(ctx.Output.Writer, "No tools found for %s\n", c.Service)
		return nil
	}

	rows := make([]map[string]any, 0, len(resp.Result.Tools))
	for _, tool := range resp.Result.Tools {
		row := map[string]any{
			"name":        tool.Name,
			"description": tool.Description,
		}
		if tool.InputSchema != nil {
			schema, ok := tool.InputSchema.(map[string]any)
			if ok {
				if req, exists := schema["required"]; exists {
					reqJSON, _ := json.Marshal(req)
					row["required"] = string(reqJSON)
				} else {
					row["required"] = "[]"
				}
			}
		}
		rows = append(rows, row)
	}

	return ctx.Output.PrintList("tools", output.APIToolColumns, rows)
}

// APICallCmd calls an MCP tool directly with raw JSON arguments.
type APICallCmd struct {
	Service string `arg:"" help:"Service name (e.g. teams, mail)"`
	Tool    string `arg:"" help:"MCP tool name (e.g. ListTeams, GetMessage)"`
	Args    string `arg:"" help:"JSON arguments" optional:"" default:"{}"`
}

func (c *APICallCmd) Run(ctx *commands.Context) error {
	endpoint := config.Endpoint(c.Service)
	if endpoint == "" {
		return fmt.Errorf("unknown service %q", c.Service)
	}

	var args map[string]any
	if err := json.Unmarshal([]byte(c.Args), &args); err != nil {
		return fmt.Errorf("invalid JSON arguments: %w", err)
	}

	client := ctx.NewMCPClient(endpoint)
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, c.Tool, args)
	if err != nil {
		return fmt.Errorf("call %s: %w", c.Tool, err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}
