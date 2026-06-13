package admin

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
)

// AdminCmd groups M365 admin subcommands.
type AdminCmd struct {
	SearchUsers  AdminSearchUsersCmd  `cmd:"" name:"search-users" help:"Search for users in the tenant"`
	ListLicenses AdminListLicensesCmd `cmd:"" name:"list-licenses" help:"List available licenses"`
	SetLicense   AdminSetLicenseCmd   `cmd:"" name:"set-license" help:"Add or remove licenses for a user"`
}

func adminEndpoint() string {
	return config.Endpoint("admin")
}

// AdminSearchUsersCmd searches for users in the tenant.
type AdminSearchUsersCmd struct {
	Query string `arg:"" help:"Search term (name, email, etc.)"`
}

func (c *AdminSearchUsersCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(adminEndpoint(), "mcp_Admin365_SearchUserTools", "search users", map[string]any{
		"searchTerm":       c.Query,
		"ConsistencyLevel": "eventual",
	})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// AdminListLicensesCmd lists available licenses.
type AdminListLicensesCmd struct{}

func (c *AdminListLicensesCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(adminEndpoint(), "mcp_Admin365_ListLicenseTools", "list licenses", map[string]any{})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// AdminSetLicenseCmd adds or removes licenses for a user.
type AdminSetLicenseCmd struct {
	UserID         string   `arg:"" help:"User ID (GUID)"`
	AddLicenses    []string `help:"License SKU IDs to add" name:"add" optional:""`
	RemoveLicenses []string `help:"License SKU IDs to remove" name:"remove" optional:""`
}

func (c *AdminSetLicenseCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(adminEndpoint(), "mcp_Admin365_LicenseMgmtTools",
			fmt.Sprintf("update licenses for user %s", c.UserID),
			map[string]any{
				"action":         "admin.set-license",
				"userId":         c.UserID,
				"addLicenses":    c.AddLicenses,
				"removeLicenses": c.RemoveLicenses,
			},
		)
	}

	addList := make([]map[string]any, 0, len(c.AddLicenses))
	for _, sku := range c.AddLicenses {
		addList = append(addList, map[string]any{"skuId": sku})
	}

	data, err := ctx.CallToolData(adminEndpoint(), "mcp_Admin365_LicenseMgmtTools", "set license", map[string]any{
		"userId":         c.UserID,
		"addLicenses":    addList,
		"removeLicenses": c.RemoveLicenses,
	})
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Licenses updated", data)
}
