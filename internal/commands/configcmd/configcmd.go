package configcmd

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/output"
)

// ConfigCmd groups config subcommands.
type ConfigCmd struct {
	Show ConfigShowCmd `cmd:"" help:"Show current configuration"`
	Set  ConfigSetCmd  `cmd:"" help:"Set a configuration value"`
	Path ConfigPathCmd `cmd:"" help:"Show config file path"`
}

// ConfigShowCmd prints the current config file contents.
type ConfigShowCmd struct{}

func (c *ConfigShowCmd) Run(ctx *commands.Context) error {
	cfg := config.LoadFileConfig()
	if ctx.Output.Format == output.FormatJSON {
		return ctx.Output.PrintItem(map[string]any{
			"client-id": cfg.ClientID,
			"tenant-id": cfg.TenantID,
			"output":    cfg.Output,
			"endpoint":  cfg.Endpoint,
		})
	}
	fmt.Fprintf(ctx.Output.Writer, "client-id: %s\n", valueOrUnset(cfg.ClientID))
	fmt.Fprintf(ctx.Output.Writer, "tenant-id: %s\n", valueOrUnset(cfg.TenantID))
	fmt.Fprintf(ctx.Output.Writer, "output:    %s\n", valueOrUnset(cfg.Output))
	fmt.Fprintf(ctx.Output.Writer, "endpoint:  %s\n", valueOrUnset(cfg.Endpoint))
	return nil
}

// ConfigSetCmd sets a key in the config file.
type ConfigSetCmd struct {
	Key   string `arg:"" help:"Config key (client-id, tenant-id, output, endpoint)"`
	Value string `arg:"" help:"Config value (use empty string to clear)"`
}

func (c *ConfigSetCmd) Run(ctx *commands.Context) error {
	cfg := config.LoadFileConfig()

	switch c.Key {
	case "client-id":
		cfg.ClientID = c.Value
	case "tenant-id":
		cfg.TenantID = c.Value
	case "output":
		switch c.Value {
		case "", "table", "json", "tsv":
			// valid
		case "plain":
			c.Value = "tsv" // normalize
		case "human":
			c.Value = "table" // normalize
		default:
			return fmt.Errorf("invalid output format %q (use table, json, or tsv)", c.Value)
		}
		if c.Value == "table" {
			cfg.Output = "" // default — don't persist
		} else {
			cfg.Output = c.Value
		}
	case "endpoint":
		if err := config.ValidateEndpointURL(c.Value); err != nil {
			return fmt.Errorf("invalid endpoint %q: %w", c.Value, err)
		}
		cfg.Endpoint = c.Value
	default:
		return fmt.Errorf("unknown config key %q (valid: client-id, tenant-id, output, endpoint)", c.Key)
	}

	if err := config.SaveFileConfig(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	if ctx.Output.Format == output.FormatJSON {
		return ctx.Output.PrintItem(map[string]any{"key": c.Key, "value": c.Value})
	}
	fmt.Fprintf(ctx.Output.Writer, "Set %s = %s\n", c.Key, c.Value)
	return nil
}

// ConfigPathCmd prints the config file path.
type ConfigPathCmd struct{}

func (c *ConfigPathCmd) Run(ctx *commands.Context) error {
	p := config.ConfigPath()
	if p == "" {
		return fmt.Errorf("could not determine config path")
	}
	if ctx.Output.Format == output.FormatJSON {
		return ctx.Output.PrintItem(map[string]any{"path": p})
	}
	fmt.Fprintln(ctx.Output.Writer, p)
	return nil
}

func valueOrUnset(v string) string {
	if v == "" {
		return "(not set)"
	}
	return v
}
