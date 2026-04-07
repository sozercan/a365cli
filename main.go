package main

import (
	"context"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/commands/admin"
	"github.com/sozercan/a365cli/internal/commands/admin365"
	"github.com/sozercan/a365cli/internal/commands/api"
	"github.com/sozercan/a365cli/internal/commands/calendar"
	"github.com/sozercan/a365cli/internal/commands/configcmd"
	"github.com/sozercan/a365cli/internal/commands/copilot"
	"github.com/sozercan/a365cli/internal/commands/dasearch"
	"github.com/sozercan/a365cli/internal/commands/excel"
	"github.com/sozercan/a365cli/internal/commands/knowledge"
	"github.com/sozercan/a365cli/internal/commands/mail"
	"github.com/sozercan/a365cli/internal/commands/me"
	"github.com/sozercan/a365cli/internal/commands/nlweb"
	"github.com/sozercan/a365cli/internal/commands/onedriveremote"
	"github.com/sozercan/a365cli/internal/commands/planner"
	"github.com/sozercan/a365cli/internal/commands/sharepoint"
	"github.com/sozercan/a365cli/internal/commands/splists"
	"github.com/sozercan/a365cli/internal/commands/teams"
	"github.com/sozercan/a365cli/internal/commands/triggers"
	"github.com/sozercan/a365cli/internal/commands/websearch"
	"github.com/sozercan/a365cli/internal/commands/word"
	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/output"
	"github.com/sozercan/a365cli/internal/version"
)

// CLI is the top-level command structure.
type CLI struct {
	// Services
	Teams      teams.TeamsCmd            `cmd:"" help:"Microsoft Teams"`
	Mail       mail.MailCmd              `cmd:"" help:"Microsoft Mail" aliases:"email"`
	Calendar   calendar.CalendarCmd      `cmd:"" help:"Microsoft Calendar" aliases:"cal"`
	Planner    planner.PlannerCmd        `cmd:"" help:"Microsoft Planner"`
	SharePoint sharepoint.SharePointCmd  `cmd:"" help:"SharePoint files and sites" name:"sharepoint" aliases:"sp"`
	SPLists    splists.SPListsCmd        `cmd:"" help:"SharePoint Lists" name:"sp-lists"`
	Me         me.MeCmd                  `cmd:"" help:"User profiles and org info"`
	Copilot    copilot.CopilotCmd        `cmd:"" help:"Microsoft 365 Copilot"`
	Word       word.WordCmd              `cmd:"" help:"Microsoft Word documents"`
	Excel      excel.ExcelCmd            `cmd:"" help:"Microsoft Excel workbooks"`
	Admin      admin.AdminCmd            `cmd:"" help:"M365 tenant administration"`
	Admin365   admin365.Admin365Cmd      `cmd:"" help:"Admin365 agent and Copilot settings" name:"admin365"`
	Knowledge  knowledge.KnowledgeCmd    `cmd:"" help:"Federated knowledge sources"`
	NLWeb      nlweb.NLWebCmd            `cmd:"" help:"NLWeb natural language search" name:"nlweb"`
	OneDriveRemote onedriveremote.OneDriveRemoteCmd `cmd:"" help:"Personal OneDrive files" name:"onedrive-remote" aliases:"odr"`
	WebSearch  websearch.WebSearchCmd    `cmd:"" help:"Web search" name:"websearch"`
	DASearch   dasearch.DASearchCmd      `cmd:"" help:"Declarative Agent search" name:"dasearch"`
	Triggers   triggers.TriggersCmd      `cmd:"" help:"Event triggers and automation"`

	// Auth & utility
	Auth       commands.AuthCmd       `cmd:"" help:"Authentication"`
	Config     configcmd.ConfigCmd    `cmd:"" help:"Manage CLI configuration"`
	Completion commands.CompletionCmd `cmd:"" help:"Generate shell completion script"`

	// Dev tools (hidden from main help)
	API api.APICmd `cmd:"" help:"API explorer for MCP servers (dev/debug)" hidden:""`

	// Output format
	Output string `help:"Output format: table, json, or tsv" enum:"table,json,tsv,," default:"" env:"A365_OUTPUT" short:"o"`
	JSON   bool   `help:"Shorthand for --output=json" hidden:"" xor:"output"`
	Plain  bool   `help:"Shorthand for --output=tsv" hidden:"" xor:"output"`

	// Safety flags
	Force   bool `help:"Skip confirmation prompts"`
	NoInput bool `help:"Never prompt; fail instead (CI mode)" name:"no-input"`
	DryRun  bool `help:"Preview write operations without executing" name:"dry-run"`

	// Connection
	Verbose  bool   `help:"Show MCP request/response for debugging" short:"v"`
	ClientID string `help:"Entra app client ID" env:"A365_CLIENT_ID"`
	TenantID string `help:"Entra tenant ID" env:"A365_TENANT_ID"`

	// Version flag
	Version kong.VersionFlag `help:"Show version" short:"V"`
}

func main() {
	cli := CLI{}

	kongCtx := kong.Parse(&cli,
		kong.Name("a365"),
		kong.Description("CLI for Microsoft 365 via agent365 MCP servers"),
		kong.Vars{"version": version.Version + " (" + version.Commit + ")"},
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Bind(&commands.Context{}),
	)

	// Load config file defaults (best-effort).
	// CLI flags and env vars always take precedence.
	fileCfg := config.LoadFileConfig()

	if cli.ClientID == "" && fileCfg.ClientID != "" {
		cli.ClientID = fileCfg.ClientID
	}
	if cli.ClientID == "" {
		cli.ClientID = config.DefaultClientID
	}
	if cli.TenantID == "" && fileCfg.TenantID != "" {
		cli.TenantID = fileCfg.TenantID
	}
	if cli.Output == "" && !cli.JSON && !cli.Plain && fileCfg.Output != "" {
		switch fileCfg.Output {
		case "json":
			cli.Output = "json"
		case "plain", "tsv":
			cli.Output = "tsv"
		case "table":
			cli.Output = "table"
		}
	}
	if os.Getenv("A365_ENDPOINT") == "" && fileCfg.Endpoint != "" {
		os.Setenv("A365_ENDPOINT", fileCfg.Endpoint)
	}

	// Build shared context
	ctx := &commands.Context{
		Ctx:      context.Background(),
		Output:   output.NewFormatter(resolveOutputFormat(cli.Output, cli.JSON, cli.Plain)),
		Verbose:  cli.Verbose,
		ClientID: cli.ClientID,
		TenantID: cli.TenantID,
		Force:    cli.Force,
		NoInput:  cli.NoInput,
		DryRun:   cli.DryRun,
	}

	// For non-auth, non-completion, non-config commands, ensure authentication
	cmd := kongCtx.Command()
	if cmd != "auth login" &&
		cmd != "auth status" &&
		cmd != "auth logout" &&
		cmd != "auth token" &&
		cmd != "completion <shell>" &&
		!strings.HasPrefix(cmd, "config ") &&
		cmd != "api servers" {
		if err := ctx.EnsureAuth(); err != nil {
			output.PrintError("%v", err)
			os.Exit(1)
		}
	}

	// Run the selected command
	err := kongCtx.Run(ctx)
	if err != nil {
		output.PrintError("%v", err)
		os.Exit(1)
	}
}

// resolveOutputFormat merges the new --output flag with the legacy --json/--plain booleans.
func resolveOutputFormat(outputFlag string, jsonFlag, plainFlag bool) string {
	if outputFlag != "" {
		return outputFlag
	}
	if jsonFlag {
		return "json"
	}
	if plainFlag {
		return "tsv"
	}
	return "table"
}
