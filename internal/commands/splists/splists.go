package splists

import (
	"encoding/json"
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/output"
)

// SPListsCmd groups all SharePoint Lists subcommands.
type SPListsCmd struct {
	Sites      SPLSitesCmd      `cmd:"" help:"Search for SharePoint sites"`
	Site       SPLSiteCmd       `cmd:"" help:"Get a site by hostname and path"`
	Subsites   SPLSubsitesCmd   `cmd:"" help:"List child sites"`
	Lists      SPLListsCmd      `cmd:"" help:"List SharePoint lists in a site"`
	Items      SPLItemsCmd      `cmd:"" help:"List items in a list"`
	Columns    SPLColumnsCmd    `cmd:"" help:"List columns in a list"`
	Create     SPLCreateCmd     `cmd:"" help:"Create a new list"`
	AddColumn  SPLAddColumnCmd  `cmd:"" name:"add-column" help:"Add a column to a list"`
	AddItem    SPLAddItemCmd    `cmd:"" name:"add-item" help:"Add an item to a list"`
	Update     SPLUpdateItemCmd `cmd:"" name:"update-item" help:"Update a list item"`
	EditCol    SPLEditColCmd    `cmd:"" name:"edit-column" help:"Edit a list column"`
	DeleteItem SPLDeleteItemCmd `cmd:"" name:"delete-item" help:"Delete a list item"`
	DeleteCol  SPLDeleteColCmd  `cmd:"" name:"delete-column" help:"Delete a list column"`
}

func spListsEndpoint() string {
	return config.Endpoint("sp-lists")
}

// --- Sites ---

// SPLSitesCmd searches for SharePoint sites by name.
type SPLSitesCmd struct {
	Query string `help:"Search query for site name" optional:""`
}

func (c *SPLSitesCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(spListsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	args := map[string]any{}
	if c.Query != "" {
		args["searchQuery"] = c.Query
	}
	resp, err := client.CallTool(ctx.Ctx, "searchSitesByName", args)
	if err != nil {
		return fmt.Errorf("search sites: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- Site ---

// SPLSiteCmd resolves a SharePoint site by hostname and server-relative path.
type SPLSiteCmd struct {
	Hostname           string `arg:"" help:"Site hostname (e.g. contoso.sharepoint.com)"`
	ServerRelativePath string `arg:"" help:"Server-relative path (e.g. /sites/engineering)"`
}

func (c *SPLSiteCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(spListsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "getSiteByPath", map[string]any{
		"hostname":           c.Hostname,
		"serverRelativePath": c.ServerRelativePath,
	})
	if err != nil {
		return fmt.Errorf("get site: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- Subsites ---

// SPLSubsitesCmd lists child sites of a SharePoint site.
type SPLSubsitesCmd struct {
	SiteID string `arg:"" help:"Site ID"`
}

func (c *SPLSubsitesCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(spListsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "listSubsites", map[string]any{
		"siteId": c.SiteID,
	})
	if err != nil {
		return fmt.Errorf("list subsites: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- Lists ---

// SPLListsCmd lists SharePoint lists in a site.
type SPLListsCmd struct {
	SiteID string `arg:"" help:"Site ID"`
}

func (c *SPLListsCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(spListsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "listLists", map[string]any{
		"siteId": c.SiteID,
	})
	if err != nil {
		return fmt.Errorf("list lists: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- Items ---

// SPLItemsCmd lists items in a SharePoint list.
type SPLItemsCmd struct {
	SiteID string `arg:"" help:"Site ID"`
	ListID string `arg:"" help:"List ID"`
}

func (c *SPLItemsCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(spListsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "listListItems", map[string]any{
		"siteId": c.SiteID,
		"listId": c.ListID,
	})
	if err != nil {
		return fmt.Errorf("list items: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- Columns ---

// SPLColumnsCmd lists columns in a SharePoint list.
type SPLColumnsCmd struct {
	SiteID string `arg:"" help:"Site ID"`
	ListID string `arg:"" help:"List ID"`
}

func (c *SPLColumnsCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(spListsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "listListColumns", map[string]any{
		"siteId": c.SiteID,
		"listId": c.ListID,
	})
	if err != nil {
		return fmt.Errorf("list columns: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- Create ---

// SPLCreateCmd creates a new SharePoint list.
type SPLCreateCmd struct {
	SiteID      string `arg:"" help:"Site ID"`
	DisplayName string `arg:"" help:"Display name for the new list"`
}

func (c *SPLCreateCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(spListsEndpoint(), "createList",
			fmt.Sprintf("create list %q in site %s", c.DisplayName, c.SiteID),
			map[string]any{"action": "sp-lists.create-list", "siteId": c.SiteID, "displayName": c.DisplayName},
		)
	}
	client := ctx.NewMCPClient(spListsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "createList", map[string]any{
		"siteId":      c.SiteID,
		"displayName": c.DisplayName,
	})
	if err != nil {
		return fmt.Errorf("create list: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("List created", data)
}

// --- AddColumn ---

// SPLAddColumnCmd adds a column to a SharePoint list.
type SPLAddColumnCmd struct {
	SiteID     string `arg:"" help:"Site ID"`
	ListID     string `arg:"" help:"List ID"`
	Name       string `arg:"" help:"Column name"`
	ColumnType string `arg:"" help:"Column type (e.g. text, number, boolean, dateTime, choice, etc.)"`
}

func (c *SPLAddColumnCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(spListsEndpoint(), "createListColumn",
			fmt.Sprintf("add column %q (type %s) to list %s in site %s", c.Name, c.ColumnType, c.ListID, c.SiteID),
			map[string]any{"action": "sp-lists.add-column", "siteId": c.SiteID, "listId": c.ListID, "name": c.Name, "columnType": c.ColumnType},
		)
	}
	client := ctx.NewMCPClient(spListsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "createListColumn", map[string]any{
		"siteId":     c.SiteID,
		"listId":     c.ListID,
		"name":       c.Name,
		"columnType": c.ColumnType,
	})
	if err != nil {
		return fmt.Errorf("add column: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Column added", data)
}

// --- AddItem ---

// SPLAddItemCmd adds an item to a SharePoint list.
type SPLAddItemCmd struct {
	SiteID string `arg:"" help:"Site ID"`
	ListID string `arg:"" help:"List ID"`
	Fields string `arg:"" help:"Item fields as JSON object (e.g. '{\"Title\":\"My Item\"}')"`
}

func (c *SPLAddItemCmd) Run(ctx *commands.Context) error {
	var fields map[string]any
	if err := json.Unmarshal([]byte(c.Fields), &fields); err != nil {
		return fmt.Errorf("invalid fields JSON: %w", err)
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(spListsEndpoint(), "createListItem",
			fmt.Sprintf("add item to list %s in site %s", c.ListID, c.SiteID),
			map[string]any{"action": "sp-lists.add-item", "siteId": c.SiteID, "listId": c.ListID, "fields": fields},
		)
	}
	client := ctx.NewMCPClient(spListsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "createListItem", map[string]any{
		"siteId": c.SiteID,
		"listId": c.ListID,
		"fields": fields,
	})
	if err != nil {
		return fmt.Errorf("add item: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Item added", data)
}

// --- UpdateItem ---

// SPLUpdateItemCmd updates a list item's fields.
type SPLUpdateItemCmd struct {
	SiteID string `arg:"" help:"Site ID"`
	ListID string `arg:"" help:"List ID"`
	ItemID string `arg:"" help:"Item ID"`
	Fields string `help:"Updated fields as JSON object" optional:""`
}

func (c *SPLUpdateItemCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(spListsEndpoint(), "updateListItem",
			fmt.Sprintf("update item %s in list %s (site %s)", c.ItemID, c.ListID, c.SiteID),
			map[string]any{"action": "sp-lists.update-item", "siteId": c.SiteID, "listId": c.ListID, "itemId": c.ItemID},
		)
	}
	client := ctx.NewMCPClient(spListsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	args := map[string]any{
		"siteId": c.SiteID,
		"listId": c.ListID,
		"itemId": c.ItemID,
	}
	if c.Fields != "" {
		var fields map[string]any
		if err := json.Unmarshal([]byte(c.Fields), &fields); err != nil {
			return fmt.Errorf("invalid fields JSON: %w", err)
		}
		args["fields"] = fields
	}
	resp, err := client.CallTool(ctx.Ctx, "updateListItem", args)
	if err != nil {
		return fmt.Errorf("update item: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Item updated", data)
}

// --- EditCol ---

// SPLEditColCmd edits (updates) a list column.
type SPLEditColCmd struct {
	SiteID   string `arg:"" help:"Site ID"`
	ListID   string `arg:"" help:"List ID"`
	ColumnID string `arg:"" help:"Column ID"`
	Name     string `help:"New column name" optional:""`
}

func (c *SPLEditColCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(spListsEndpoint(), "editListColumn",
			fmt.Sprintf("edit column %s in list %s (site %s)", c.ColumnID, c.ListID, c.SiteID),
			map[string]any{"action": "sp-lists.edit-column", "siteId": c.SiteID, "listId": c.ListID, "columnId": c.ColumnID},
		)
	}
	client := ctx.NewMCPClient(spListsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	args := map[string]any{
		"siteId":   c.SiteID,
		"listId":   c.ListID,
		"columnId": c.ColumnID,
	}
	if c.Name != "" {
		args["name"] = c.Name
	}
	resp, err := client.CallTool(ctx.Ctx, "editListColumn", args)
	if err != nil {
		return fmt.Errorf("edit column: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Column updated", data)
}

// --- DeleteItem ---

// SPLDeleteItemCmd deletes a list item. This is a destructive operation.
type SPLDeleteItemCmd struct {
	SiteID string `arg:"" help:"Site ID"`
	ListID string `arg:"" help:"List ID"`
	ItemID string `arg:"" help:"Item ID"`
}

func (c *SPLDeleteItemCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(spListsEndpoint(), "deleteListItem",
			fmt.Sprintf("delete item %s from list %s (site %s)", c.ItemID, c.ListID, c.SiteID),
			map[string]any{"action": "sp-lists.delete-item", "siteId": c.SiteID, "listId": c.ListID, "itemId": c.ItemID},
		)
	}
	if err := ctx.Confirm(fmt.Sprintf("delete item %s from list %s", c.ItemID, c.ListID)); err != nil {
		return err
	}
	client := ctx.NewMCPClient(spListsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "deleteListItem", map[string]any{
		"siteId": c.SiteID,
		"listId": c.ListID,
		"itemId": c.ItemID,
	})
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Item deleted", data)
}

// --- DeleteCol ---

// SPLDeleteColCmd deletes a list column. This is a destructive operation.
type SPLDeleteColCmd struct {
	SiteID   string `arg:"" help:"Site ID"`
	ListID   string `arg:"" help:"List ID"`
	ColumnID string `arg:"" help:"Column ID"`
}

func (c *SPLDeleteColCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(spListsEndpoint(), "deleteListColumn",
			fmt.Sprintf("delete column %s from list %s (site %s)", c.ColumnID, c.ListID, c.SiteID),
			map[string]any{"action": "sp-lists.delete-column", "siteId": c.SiteID, "listId": c.ListID, "columnId": c.ColumnID},
		)
	}
	if err := ctx.Confirm(fmt.Sprintf("delete column %s from list %s", c.ColumnID, c.ListID)); err != nil {
		return err
	}
	client := ctx.NewMCPClient(spListsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "deleteListColumn", map[string]any{
		"siteId":   c.SiteID,
		"listId":   c.ListID,
		"columnId": c.ColumnID,
	})
	if err != nil {
		return fmt.Errorf("delete column: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Column deleted", data)
}
