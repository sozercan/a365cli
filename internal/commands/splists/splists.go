package splists

import (
	"encoding/json"
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
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
	args := map[string]any{}
	if c.Query != "" {
		args["searchQuery"] = c.Query
	}
	data, err := ctx.CallToolData(spListsEndpoint(), "searchSitesByName", "search sites", args)
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
	data, err := ctx.CallToolData(spListsEndpoint(), "getSiteByPath", "get site", map[string]any{
		"hostname":           c.Hostname,
		"serverRelativePath": c.ServerRelativePath,
	})
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
	data, err := ctx.CallToolData(spListsEndpoint(), "listSubsites", "list subsites", map[string]any{
		"siteId": c.SiteID,
	})
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
	data, err := ctx.CallToolData(spListsEndpoint(), "listLists", "list lists", map[string]any{
		"siteId": c.SiteID,
	})
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
	data, err := ctx.CallToolData(spListsEndpoint(), "listListItems", "list items", map[string]any{
		"siteId": c.SiteID,
		"listId": c.ListID,
	})
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
	data, err := ctx.CallToolData(spListsEndpoint(), "listListColumns", "list columns", map[string]any{
		"siteId": c.SiteID,
		"listId": c.ListID,
	})
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
	Template    string `help:"List template (for example genericList, documentLibrary, or issueTracking)" default:"genericList"`
}

func (c *SPLCreateCmd) Run(ctx *commands.Context) error {
	template := c.Template
	if template == "" {
		template = "genericList"
	}
	args := map[string]any{
		"siteId":      c.SiteID,
		"displayName": c.DisplayName,
		"list": map[string]any{
			"template": template,
		},
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(spListsEndpoint(), "createList",
			fmt.Sprintf("create list %q in site %s", c.DisplayName, c.SiteID),
			map[string]any{
				"action":      "sp-lists.create-list",
				"siteId":      c.SiteID,
				"displayName": c.DisplayName,
				"template":    template,
			},
			args,
		)
	}

	data, err := ctx.CallToolData(spListsEndpoint(), "createList", "create list", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("List created", data)
}

// --- AddColumn ---

// SPLAddColumnCmd adds a column to a SharePoint list.
type SPLAddColumnCmd struct {
	SiteID         string `arg:"" help:"Site ID"`
	ListID         string `arg:"" help:"List ID"`
	Name           string `arg:"" help:"API/static column name (no spaces)"`
	ColumnType     string `arg:"" help:"Column type" enum:"boolean,choice,dateTime,hyperlinkOrPicture,lookup,number,personOrGroup,text"`
	ColumnSettings string `help:"Column type settings as a JSON object (choice columns require a choices array)" optional:""`
}

func (c *SPLAddColumnCmd) Run(ctx *commands.Context) error {
	settings := map[string]any{}
	if c.ColumnSettings != "" {
		if err := json.Unmarshal([]byte(c.ColumnSettings), &settings); err != nil {
			return fmt.Errorf("invalid column settings JSON: %w", err)
		}
		if settings == nil {
			return fmt.Errorf("column settings must be a JSON object")
		}
	}

	args := map[string]any{
		"siteId":     c.SiteID,
		"listId":     c.ListID,
		"name":       c.Name,
		c.ColumnType: settings,
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(spListsEndpoint(), "createListColumn",
			fmt.Sprintf("add column %q (type %s) to list %s in site %s", c.Name, c.ColumnType, c.ListID, c.SiteID),
			map[string]any{
				"action":      "sp-lists.add-column",
				"siteId":      c.SiteID,
				"listId":      c.ListID,
				"name":        c.Name,
				"columnType":  c.ColumnType,
				"settingsSet": c.ColumnSettings != "",
			},
			args,
		)
	}

	data, err := ctx.CallToolData(spListsEndpoint(), "createListColumn", "add column", args)
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
	if len(fields) == 0 {
		return fmt.Errorf("fields must contain at least one property")
	}

	args := map[string]any{
		"siteId": c.SiteID,
		"listId": c.ListID,
		"fields": fields,
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(spListsEndpoint(), "createListItem",
			fmt.Sprintf("add item to list %s in site %s", c.ListID, c.SiteID),
			map[string]any{
				"action":     "sp-lists.add-item",
				"siteId":     c.SiteID,
				"listId":     c.ListID,
				"fieldCount": len(fields),
			},
			args,
		)
	}

	data, err := ctx.CallToolData(spListsEndpoint(), "createListItem", "add item", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Item added", data)
}

// --- UpdateItem ---

// SPLUpdateItemCmd updates a list item's fields.
type SPLUpdateItemCmd struct {
	SiteID  string `arg:"" help:"Site ID"`
	ListID  string `arg:"" help:"List ID"`
	ItemID  string `arg:"" help:"Item ID"`
	Fields  string `help:"Updated fields as a JSON object" required:""`
	IfMatch string `help:"Optional ETag for concurrency control (use * to force the update)" optional:""`
}

func (c *SPLUpdateItemCmd) Run(ctx *commands.Context) error {
	if c.Fields == "" {
		return fmt.Errorf("--fields is required")
	}

	var fields map[string]any
	if err := json.Unmarshal([]byte(c.Fields), &fields); err != nil {
		return fmt.Errorf("invalid fields JSON: %w", err)
	}
	if len(fields) == 0 {
		return fmt.Errorf("fields must contain at least one property")
	}

	args := map[string]any{
		"siteId": c.SiteID,
		"listId": c.ListID,
		"itemId": c.ItemID,
		"fields": fields,
	}
	if c.IfMatch != "" {
		args["ifMatch"] = c.IfMatch
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(spListsEndpoint(), "updateListItem",
			fmt.Sprintf("update item %s in list %s (site %s)", c.ItemID, c.ListID, c.SiteID),
			map[string]any{
				"action":     "sp-lists.update-item",
				"siteId":     c.SiteID,
				"listId":     c.ListID,
				"itemId":     c.ItemID,
				"fieldCount": len(fields),
			},
			args,
		)
	}

	data, err := ctx.CallToolData(spListsEndpoint(), "updateListItem", "update item", args)
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
	Name     string `help:"New user-facing display name" optional:""`
}

func (c *SPLEditColCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"siteId":   c.SiteID,
		"listId":   c.ListID,
		"columnId": c.ColumnID,
	}
	if c.Name != "" {
		args["displayName"] = c.Name
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(spListsEndpoint(), "editListColumn",
			fmt.Sprintf("edit column %s in list %s (site %s)", c.ColumnID, c.ListID, c.SiteID),
			map[string]any{
				"action":         "sp-lists.edit-column",
				"siteId":         c.SiteID,
				"listId":         c.ListID,
				"columnId":       c.ColumnID,
				"newDisplayName": c.Name,
			},
			args,
		)
	}

	data, err := ctx.CallToolData(spListsEndpoint(), "editListColumn", "edit column", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Column updated", data)
}

// --- DeleteItem ---

// SPLDeleteItemCmd deletes a list item. This is a destructive operation.
type SPLDeleteItemCmd struct {
	SiteID  string `arg:"" help:"Site ID"`
	ListID  string `arg:"" help:"List ID"`
	ItemID  string `arg:"" help:"Item ID"`
	IfMatch string `help:"Optional ETag for concurrency control" optional:""`
}

func (c *SPLDeleteItemCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"siteId": c.SiteID,
		"listId": c.ListID,
		"itemId": c.ItemID,
	}
	if c.IfMatch != "" {
		args["ifMatch"] = c.IfMatch
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(spListsEndpoint(), "deleteListItem",
			fmt.Sprintf("delete item %s from list %s (site %s)", c.ItemID, c.ListID, c.SiteID),
			map[string]any{
				"action": "sp-lists.delete-item",
				"siteId": c.SiteID,
				"listId": c.ListID,
				"itemId": c.ItemID,
			},
			args,
		)
	}
	if err := ctx.Confirm(fmt.Sprintf("delete item %s from list %s", c.ItemID, c.ListID)); err != nil {
		return err
	}

	data, err := ctx.CallToolData(spListsEndpoint(), "deleteListItem", "delete item", args)
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
	args := map[string]any{
		"siteId":   c.SiteID,
		"listId":   c.ListID,
		"columnId": c.ColumnID,
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(spListsEndpoint(), "deleteListColumn",
			fmt.Sprintf("delete column %s from list %s (site %s)", c.ColumnID, c.ListID, c.SiteID),
			map[string]any{
				"action":   "sp-lists.delete-column",
				"siteId":   c.SiteID,
				"listId":   c.ListID,
				"columnId": c.ColumnID,
			},
			args,
		)
	}
	if err := ctx.Confirm(fmt.Sprintf("delete column %s from list %s", c.ColumnID, c.ListID)); err != nil {
		return err
	}

	data, err := ctx.CallToolData(spListsEndpoint(), "deleteListColumn", "delete column", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Column deleted", data)
}
