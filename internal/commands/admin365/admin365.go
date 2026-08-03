package admin365

import (
	"fmt"
	"strings"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
)

// admin365Endpoint returns the agent365 endpoint for the Admin365 MCP server.
func admin365Endpoint() string {
	return config.Endpoint("admin365")
}

// Admin365Cmd groups all Admin365 subcommands.
type Admin365Cmd struct {
	BulkAdd       Admin365BulkAddCmd       `cmd:"" name:"bulk-add" help:"Bulk add users to tenant"`
	AgentAccess   Admin365AgentAccessCmd   `cmd:"" name:"agent-access" help:"Get agent access settings"`
	AgentSharing  Admin365AgentSharingCmd  `cmd:"" name:"agent-sharing" help:"Get agent sharing settings"`
	MsApps        Admin365MsAppsCmd        `cmd:"" name:"ms-apps" help:"Get Microsoft apps install settings"`
	ThirdParty    Admin365ThirdPartyCmd    `cmd:"" name:"third-party" help:"Get third-party apps settings"`
	LobApps       Admin365LobAppsCmd       `cmd:"" name:"lob-apps" help:"Get LOB apps settings"`
	SetAccess     Admin365SetAccessCmd     `cmd:"" name:"set-access" help:"Update agent access settings"`
	SetSharing    Admin365SetSharingCmd    `cmd:"" name:"set-sharing" help:"Update agent sharing settings"`
	SetMsApps     Admin365SetMsAppsCmd     `cmd:"" name:"set-ms-apps" help:"Update Microsoft apps settings"`
	SetThirdParty Admin365SetThirdPartyCmd `cmd:"" name:"set-third-party" help:"Update third-party apps settings"`
	SetLobApps    Admin365SetLobAppsCmd    `cmd:"" name:"set-lob-apps" help:"Update LOB apps settings"`
	Readiness     Admin365ReadinessCmd     `cmd:"" name:"copilot-readiness" help:"Check Copilot readiness"`
	CopilotStatus Admin365CopilotStatusCmd `cmd:"" name:"copilot-status" help:"Get Copilot admin settings"`
	SetCopilot    Admin365SetCopilotCmd    `cmd:"" name:"set-copilot" help:"Enable/disable Copilot for admins"`
}

// --- BulkAddUsers ---

// Admin365BulkAddCmd bulk-adds users to the tenant.
type Admin365BulkAddCmd struct {
	FileContent string `arg:"" help:"CSV or JSON content with users to add"`
}

func (c *Admin365BulkAddCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"fileContent": c.FileContent,
	}
	lineCount := 0
	if c.FileContent != "" {
		lineCount = strings.Count(c.FileContent, "\n") + 1
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(admin365Endpoint(), "BulkAddUsers",
			"bulk add users to tenant",
			map[string]any{
				"action":       "admin365.bulk-add",
				"contentBytes": len(c.FileContent),
				"lineCount":    lineCount,
			},
			args,
		)
	}

	data, err := ctx.CallToolData(admin365Endpoint(), "BulkAddUsers", "bulk add users", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Users added", data)
}

// --- GetWhoCanAccessAgentsSettings ---

// Admin365AgentAccessCmd gets agent access settings.
type Admin365AgentAccessCmd struct{}

func (c *Admin365AgentAccessCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(admin365Endpoint(), "GetWhoCanAccessAgentsSettings", "get agent access settings", map[string]any{})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- GetWhoCanShareAgentsOrgWideSettings ---

// Admin365AgentSharingCmd gets agent sharing settings.
type Admin365AgentSharingCmd struct{}

func (c *Admin365AgentSharingCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(admin365Endpoint(), "GetWhoCanShareAgentsOrgWideSettings", "get agent sharing settings", map[string]any{})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- GetCanInstallMicrosoftAppsAndAgentsSettings ---

// Admin365MsAppsCmd gets Microsoft apps install settings.
type Admin365MsAppsCmd struct{}

func (c *Admin365MsAppsCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(admin365Endpoint(), "GetCanInstallMicrosoftAppsAndAgentsSettings", "get Microsoft apps settings", map[string]any{})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- GetCanInstallThirdPartyAppsAndAgentsSettings ---

// Admin365ThirdPartyCmd gets third-party apps settings.
type Admin365ThirdPartyCmd struct{}

func (c *Admin365ThirdPartyCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(admin365Endpoint(), "GetCanInstallThirdPartyAppsAndAgentsSettings", "get third-party apps settings", map[string]any{})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- GetCanInstallLOBAppsAndAgentsSettings ---

// Admin365LobAppsCmd gets LOB apps settings.
type Admin365LobAppsCmd struct{}

func (c *Admin365LobAppsCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(admin365Endpoint(), "GetCanInstallLOBAppsAndAgentsSettings", "get LOB apps settings", map[string]any{})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- UpdateWhoCanAccessAgentsSettings ---

// Admin365SetAccessCmd updates agent access settings.
type Admin365SetAccessCmd struct {
	AccessLevel string `arg:"" help:"Access level to set"`
}

func (c *Admin365SetAccessCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"accessLevel": c.AccessLevel,
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(admin365Endpoint(), "UpdateWhoCanAccessAgentsSettings",
			fmt.Sprintf("update agent access to %q", c.AccessLevel),
			map[string]any{
				"action":      "admin365.set-access",
				"accessLevel": c.AccessLevel,
			},
			args,
		)
	}

	data, err := ctx.CallToolData(admin365Endpoint(), "UpdateWhoCanAccessAgentsSettings", "update agent access", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Agent access updated", data)
}

// --- UpdateWhoCanShareAgentsOrgWideSettings ---

// Admin365SetSharingCmd updates agent sharing settings.
type Admin365SetSharingCmd struct {
	AccessLevel string `arg:"" help:"Access level to set"`
}

func (c *Admin365SetSharingCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"accessLevel": c.AccessLevel,
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(admin365Endpoint(), "UpdateWhoCanShareAgentsOrgWideSettings",
			fmt.Sprintf("update agent sharing to %q", c.AccessLevel),
			map[string]any{
				"action":      "admin365.set-sharing",
				"accessLevel": c.AccessLevel,
			},
			args,
		)
	}

	data, err := ctx.CallToolData(admin365Endpoint(), "UpdateWhoCanShareAgentsOrgWideSettings", "update agent sharing", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Agent sharing updated", data)
}

// --- UpdateCanInstallMicrosoftAppsAndAgentsSettings ---

// Admin365SetMsAppsCmd updates Microsoft apps install settings.
type Admin365SetMsAppsCmd struct {
	Allowed string `arg:"" help:"Whether to allow (true/false)"`
}

func (c *Admin365SetMsAppsCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"allowed": c.Allowed,
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(admin365Endpoint(), "UpdateCanInstallMicrosoftAppsAndAgentsSettings",
			fmt.Sprintf("update Microsoft apps install to %s", c.Allowed),
			map[string]any{
				"action":  "admin365.set-ms-apps",
				"allowed": c.Allowed,
			},
			args,
		)
	}

	data, err := ctx.CallToolData(admin365Endpoint(), "UpdateCanInstallMicrosoftAppsAndAgentsSettings", "update Microsoft apps settings", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Microsoft apps settings updated", data)
}

// --- UpdateCanInstallThirdPartyAppsAndAgentsSettings ---

// Admin365SetThirdPartyCmd updates third-party apps settings.
type Admin365SetThirdPartyCmd struct {
	Allowed string `arg:"" help:"Whether to allow (true/false)"`
}

func (c *Admin365SetThirdPartyCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"allowed": c.Allowed,
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(admin365Endpoint(), "UpdateCanInstallThirdPartyAppsAndAgentsSettings",
			fmt.Sprintf("update third-party apps install to %s", c.Allowed),
			map[string]any{
				"action":  "admin365.set-third-party",
				"allowed": c.Allowed,
			},
			args,
		)
	}

	data, err := ctx.CallToolData(admin365Endpoint(), "UpdateCanInstallThirdPartyAppsAndAgentsSettings", "update third-party apps settings", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Third-party apps settings updated", data)
}

// --- UpdateCanInstallLOBAppsAndAgentsSettings ---

// Admin365SetLobAppsCmd updates LOB apps settings.
type Admin365SetLobAppsCmd struct {
	Allowed string `arg:"" help:"Whether to allow (true/false)"`
}

func (c *Admin365SetLobAppsCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"allowed": c.Allowed,
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(admin365Endpoint(), "UpdateCanInstallLOBAppsAndAgentsSettings",
			fmt.Sprintf("update LOB apps install to %s", c.Allowed),
			map[string]any{
				"action":  "admin365.set-lob-apps",
				"allowed": c.Allowed,
			},
			args,
		)
	}

	data, err := ctx.CallToolData(admin365Endpoint(), "UpdateCanInstallLOBAppsAndAgentsSettings", "update LOB apps settings", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("LOB apps settings updated", data)
}

// --- GetCopilotReadiness ---

// Admin365ReadinessCmd checks Copilot readiness.
type Admin365ReadinessCmd struct{}

func (c *Admin365ReadinessCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(admin365Endpoint(), "GetCopilotReadiness", "get Copilot readiness", map[string]any{})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- GetCopilotAdminSettings ---

// Admin365CopilotStatusCmd gets Copilot admin settings.
type Admin365CopilotStatusCmd struct{}

func (c *Admin365CopilotStatusCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(admin365Endpoint(), "GetCopilotAdminSettings", "get Copilot admin settings", map[string]any{})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- UpdateCopilotAdminSettings ---

// Admin365SetCopilotCmd enables or disables Copilot for admins.
type Admin365SetCopilotCmd struct {
	IsEnabled string `arg:"" help:"Whether Copilot is enabled (true/false)"`
}

func (c *Admin365SetCopilotCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"isEnabled": c.IsEnabled,
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(admin365Endpoint(), "UpdateCopilotAdminSettings",
			fmt.Sprintf("update Copilot admin setting to isEnabled=%s", c.IsEnabled),
			map[string]any{
				"action":    "admin365.set-copilot",
				"isEnabled": c.IsEnabled,
			},
			args,
		)
	}

	data, err := ctx.CallToolData(admin365Endpoint(), "UpdateCopilotAdminSettings", "update Copilot admin settings", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Copilot admin settings updated", data)
}
