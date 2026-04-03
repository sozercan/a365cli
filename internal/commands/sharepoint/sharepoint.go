package sharepoint

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/output"
)

// spEndpoint returns the agent365 endpoint for the SharePoint MCP server.
func spEndpoint() string {
	ep := config.Endpoint("sharepoint")
	if ep == "" {
		panic("sharepoint server not configured")
	}
	return ep
}

// SharePointCmd groups all SharePoint subcommands.
type SharePointCmd struct {
	Find   SPFindSiteCmd `cmd:"" help:"Find SharePoint sites"`
	Libs   SPListLibsCmd `cmd:"" help:"List document libraries in a site"`
	Ls     SPListCmd     `cmd:"" help:"List folder contents"`
	Get    SPGetCmd      `cmd:"" help:"Get file/folder metadata"`
	Cat    SPCatCmd      `cmd:"" help:"Read a text file"`
	Search SPSearchCmd   `cmd:"" help:"Search for files or folders"`
	Mkdir  SPMkdirCmd    `cmd:"" help:"Create a folder"`
	Write  SPWriteCmd    `cmd:"" help:"Create a text file"`
	Upload SPUploadCmd   `cmd:"" help:"Upload a file from URL"`
	Rm     SPDeleteCmd   `cmd:"" help:"Delete a file or folder"`
	Mv     SPMoveCmd     `cmd:"" help:"Move a file or folder"`
	Cp     SPCopyCmd     `cmd:"" help:"Copy a file or folder"`
	Rename SPRenameCmd   `cmd:"" help:"Rename a file or folder"`
	Share  SPShareCmd    `cmd:"" help:"Create a sharing link"`
	Label  SPLabelCmd    `cmd:"" help:"Set sensitivity label on a file"`
	Status SPStatusCmd   `cmd:"" help:"Check async operation status"`
}

// --- findSite ---

// SPFindSiteCmd finds SharePoint sites.
type SPFindSiteCmd struct {
	Query   string `help:"Search query to find sites" optional:""`
	SiteUrl string `help:"Site URL to look up" optional:""`
}

func (c *SPFindSiteCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(spEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	args := map[string]any{}
	if c.Query != "" {
		args["searchQuery"] = c.Query
	}
	if c.SiteUrl != "" {
		args["siteUrl"] = c.SiteUrl
	}

	resp, err := client.CallTool(ctx.Ctx, "findSite", args)
	if err != nil {
		return fmt.Errorf("find site: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- listDocumentLibrariesInSite / getDefaultDocumentLibraryInSite ---

// SPListLibsCmd lists document libraries in a site, or gets the default library.
type SPListLibsCmd struct {
	SiteID   string `help:"Site ID" optional:""`
	SitePath string `help:"Site path (e.g. /sites/teamsite)" optional:""`
	Default  bool   `help:"Get the default document library only" optional:""`
}

func (c *SPListLibsCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(spEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	args := map[string]any{}
	if c.SiteID != "" {
		args["siteId"] = c.SiteID
	}
	if c.SitePath != "" {
		args["sitePath"] = c.SitePath
	}

	var toolName string
	if c.Default {
		toolName = "getDefaultDocumentLibraryInSite"
	} else {
		toolName = "listDocumentLibrariesInSite"
	}

	resp, err := client.CallTool(ctx.Ctx, toolName, args)
	if err != nil {
		return fmt.Errorf("list libraries: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- getFolderChildren ---

// SPListCmd lists folder contents.
type SPListCmd struct {
	DriveID    string `arg:"" help:"Drive ID"`
	FolderPath string `help:"Folder path (e.g. /Documents/Reports)" optional:""`
	FolderID   string `help:"Folder ID (alternative to path)" optional:""`
}

func (c *SPListCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(spEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	args := map[string]any{
		"driveId": c.DriveID,
	}
	if c.FolderPath != "" {
		args["folderPath"] = c.FolderPath
	}
	if c.FolderID != "" {
		args["folderId"] = c.FolderID
	}

	resp, err := client.CallTool(ctx.Ctx, "getFolderChildren", args)
	if err != nil {
		return fmt.Errorf("list folder: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- getFileOrFolderMetadata / getFileOrFolderMetadataByUrl ---

// SPGetCmd gets file or folder metadata.
type SPGetCmd struct {
	DriveID  string `help:"Drive ID (required unless --url is used)" optional:""`
	ItemPath string `help:"Item path" optional:""`
	ItemID   string `help:"Item ID (alternative to path)" optional:""`
	URL      string `help:"SharePoint URL to look up metadata by URL" optional:""`
}

func (c *SPGetCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(spEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	var toolName string
	args := map[string]any{}

	if c.URL != "" {
		toolName = "getFileOrFolderMetadataByUrl"
		args["sharePointUrl"] = c.URL
	} else {
		toolName = "getFileOrFolderMetadata"
		if c.DriveID != "" {
			args["driveId"] = c.DriveID
		}
		if c.ItemPath != "" {
			args["itemPath"] = c.ItemPath
		}
		if c.ItemID != "" {
			args["itemId"] = c.ItemID
		}
	}

	resp, err := client.CallTool(ctx.Ctx, toolName, args)
	if err != nil {
		return fmt.Errorf("get metadata: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- readSmallTextFile ---

// SPCatCmd reads a small text file from SharePoint.
type SPCatCmd struct {
	DriveID  string `arg:"" help:"Drive ID"`
	FilePath string `help:"File path" optional:""`
	FileID   string `help:"File ID (alternative to path)" optional:""`
	Binary   bool   `help:"Read as binary (base64 output)" optional:""`
}

func (c *SPCatCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(spEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	args := map[string]any{
		"driveId": c.DriveID,
	}
	if c.FilePath != "" {
		args["filePath"] = c.FilePath
	}
	if c.FileID != "" {
		args["fileId"] = c.FileID
	}

	var toolName string
	if c.Binary {
		toolName = "readSmallBinaryFile"
	} else {
		toolName = "readSmallTextFile"
	}

	resp, err := client.CallTool(ctx.Ctx, toolName, args)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- findFileOrFolder ---

// SPSearchCmd searches for files or folders.
type SPSearchCmd struct {
	Query   string `arg:"" help:"Search query"`
	DriveID string `help:"Limit search to a specific drive" optional:""`
	SiteID  string `help:"Limit search to a specific site" optional:""`
}

func (c *SPSearchCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(spEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	args := map[string]any{
		"searchQuery": c.Query,
	}
	if c.DriveID != "" {
		args["driveId"] = c.DriveID
	}
	if c.SiteID != "" {
		args["siteId"] = c.SiteID
	}

	resp, err := client.CallTool(ctx.Ctx, "findFileOrFolder", args)
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- createFolder ---

// SPMkdirCmd creates a folder in SharePoint.
type SPMkdirCmd struct {
	DriveID    string `arg:"" help:"Drive ID"`
	ParentPath string `arg:"" help:"Parent folder path"`
	FolderName string `arg:"" help:"Name for the new folder"`
}

func (c *SPMkdirCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.Output.PrintDryRun(
			fmt.Sprintf("create folder %q in %s", c.FolderName, c.ParentPath),
			map[string]any{
				"action":     "sharepoint.mkdir",
				"driveId":    c.DriveID,
				"parentPath": c.ParentPath,
				"folderName": c.FolderName,
			},
		)
	}

	client := ctx.NewMCPClient(spEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "createFolder", map[string]any{
		"driveId":    c.DriveID,
		"parentPath": c.ParentPath,
		"folderName": c.FolderName,
	})
	if err != nil {
		return fmt.Errorf("create folder: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Folder created", data)
}

// --- createSmallTextFile / createSmallBinaryFile ---

// SPWriteCmd creates a text or binary file in SharePoint.
type SPWriteCmd struct {
	DriveID       string `arg:"" help:"Drive ID"`
	FolderPath    string `arg:"" help:"Destination folder path"`
	FileName      string `arg:"" help:"File name to create"`
	Content       string `help:"Text content for the file" optional:""`
	ContentBase64 string `help:"Base64-encoded content (for binary files)" optional:""`
}

func (c *SPWriteCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.Output.PrintDryRun(
			fmt.Sprintf("create file %q in %s", c.FileName, c.FolderPath),
			map[string]any{
				"action":     "sharepoint.write",
				"driveId":    c.DriveID,
				"folderPath": c.FolderPath,
				"fileName":   c.FileName,
			},
		)
	}

	client := ctx.NewMCPClient(spEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	var toolName string
	args := map[string]any{
		"driveId":    c.DriveID,
		"folderPath": c.FolderPath,
		"fileName":   c.FileName,
	}

	if c.ContentBase64 != "" {
		toolName = "createSmallBinaryFile"
		args["contentBase64"] = c.ContentBase64
	} else {
		toolName = "createSmallTextFile"
		args["content"] = c.Content
	}

	resp, err := client.CallTool(ctx.Ctx, toolName, args)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("File created", data)
}

// --- uploadFileFromUrl ---

// SPUploadCmd uploads a file from a URL to SharePoint.
type SPUploadCmd struct {
	SourceURL             string `arg:"" help:"Source URL to download from"`
	DestinationDriveID    string `arg:"" help:"Destination drive ID"`
	DestinationFolderPath string `arg:"" help:"Destination folder path"`
	FileName              string `arg:"" help:"File name for the uploaded file"`
}

func (c *SPUploadCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.Output.PrintDryRun(
			fmt.Sprintf("upload %q from %s to %s", c.FileName, c.SourceURL, c.DestinationFolderPath),
			map[string]any{
				"action":                "sharepoint.upload",
				"sourceUrl":             c.SourceURL,
				"destinationDriveId":    c.DestinationDriveID,
				"destinationFolderPath": c.DestinationFolderPath,
				"fileName":              c.FileName,
			},
		)
	}

	client := ctx.NewMCPClient(spEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "uploadFileFromUrl", map[string]any{
		"sourceUrl":             c.SourceURL,
		"destinationDriveId":    c.DestinationDriveID,
		"destinationFolderPath": c.DestinationFolderPath,
		"fileName":              c.FileName,
	})
	if err != nil {
		return fmt.Errorf("upload file: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("File uploaded", data)
}

// --- deleteFileOrFolder ---

// SPDeleteCmd deletes a file or folder from SharePoint.
type SPDeleteCmd struct {
	DriveID  string `arg:"" help:"Drive ID"`
	ItemPath string `help:"Item path to delete" optional:""`
	ItemID   string `help:"Item ID to delete (alternative to path)" optional:""`
}

func (c *SPDeleteCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		target := c.ItemPath
		if target == "" {
			target = c.ItemID
		}
		return ctx.Output.PrintDryRun(
			fmt.Sprintf("delete %s from drive %s", target, c.DriveID),
			map[string]any{
				"action":   "sharepoint.rm",
				"driveId":  c.DriveID,
				"itemPath": c.ItemPath,
				"itemId":   c.ItemID,
			},
		)
	}

	target := c.ItemPath
	if target == "" {
		target = c.ItemID
	}
	if err := ctx.Confirm(fmt.Sprintf("delete %s from drive %s", target, c.DriveID)); err != nil {
		return err
	}

	client := ctx.NewMCPClient(spEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	args := map[string]any{
		"driveId": c.DriveID,
	}
	if c.ItemPath != "" {
		args["itemPath"] = c.ItemPath
	}
	if c.ItemID != "" {
		args["itemId"] = c.ItemID
	}

	resp, err := client.CallTool(ctx.Ctx, "deleteFileOrFolder", args)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Deleted", data)
}

// --- moveFileOrFolder ---

// SPMoveCmd moves a file or folder within SharePoint.
type SPMoveCmd struct {
	SourceDriveID         string `arg:"" help:"Source drive ID"`
	SourceItemPath        string `arg:"" help:"Source item path"`
	DestinationDriveID    string `arg:"" help:"Destination drive ID"`
	DestinationFolderPath string `arg:"" help:"Destination folder path"`
}

func (c *SPMoveCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.Output.PrintDryRun(
			fmt.Sprintf("move %s to %s", c.SourceItemPath, c.DestinationFolderPath),
			map[string]any{
				"action":                "sharepoint.mv",
				"sourceDriveId":         c.SourceDriveID,
				"sourceItemPath":        c.SourceItemPath,
				"destinationDriveId":    c.DestinationDriveID,
				"destinationFolderPath": c.DestinationFolderPath,
			},
		)
	}

	client := ctx.NewMCPClient(spEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "moveFileOrFolder", map[string]any{
		"sourceDriveId":         c.SourceDriveID,
		"sourceItemPath":        c.SourceItemPath,
		"destinationDriveId":    c.DestinationDriveID,
		"destinationFolderPath": c.DestinationFolderPath,
	})
	if err != nil {
		return fmt.Errorf("move: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Moved", data)
}

// --- copyFileOrFolder ---

// SPCopyCmd copies a file or folder within SharePoint.
type SPCopyCmd struct {
	SourceDriveID         string `arg:"" help:"Source drive ID"`
	SourceItemPath        string `arg:"" help:"Source item path"`
	DestinationDriveID    string `arg:"" help:"Destination drive ID"`
	DestinationFolderPath string `arg:"" help:"Destination folder path"`
}

func (c *SPCopyCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.Output.PrintDryRun(
			fmt.Sprintf("copy %s to %s", c.SourceItemPath, c.DestinationFolderPath),
			map[string]any{
				"action":                "sharepoint.cp",
				"sourceDriveId":         c.SourceDriveID,
				"sourceItemPath":        c.SourceItemPath,
				"destinationDriveId":    c.DestinationDriveID,
				"destinationFolderPath": c.DestinationFolderPath,
			},
		)
	}

	client := ctx.NewMCPClient(spEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "copyFileOrFolder", map[string]any{
		"sourceDriveId":         c.SourceDriveID,
		"sourceItemPath":        c.SourceItemPath,
		"destinationDriveId":    c.DestinationDriveID,
		"destinationFolderPath": c.DestinationFolderPath,
	})
	if err != nil {
		return fmt.Errorf("copy: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Copied", data)
}

// --- renameFileOrFolder ---

// SPRenameCmd renames a file or folder.
type SPRenameCmd struct {
	DriveID  string `arg:"" help:"Drive ID"`
	ItemPath string `arg:"" help:"Item path to rename"`
	NewName  string `arg:"" help:"New name for the item"`
}

func (c *SPRenameCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.Output.PrintDryRun(
			fmt.Sprintf("rename %s to %q", c.ItemPath, c.NewName),
			map[string]any{
				"action":   "sharepoint.rename",
				"driveId":  c.DriveID,
				"itemPath": c.ItemPath,
				"newName":  c.NewName,
			},
		)
	}

	client := ctx.NewMCPClient(spEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "renameFileOrFolder", map[string]any{
		"driveId":  c.DriveID,
		"itemPath": c.ItemPath,
		"newName":  c.NewName,
	})
	if err != nil {
		return fmt.Errorf("rename: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Renamed", data)
}

// --- shareFileOrFolder ---

// SPShareCmd creates a sharing link for a file or folder.
type SPShareCmd struct {
	DriveID  string `arg:"" help:"Drive ID"`
	ItemPath string `arg:"" help:"Item path to share"`
	Type     string `arg:"" help:"Link type: view or edit" enum:"view,edit"`
	Scope    string `arg:"" help:"Link scope: anonymous, organization, or users" enum:"anonymous,organization,users"`
}

func (c *SPShareCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.Output.PrintDryRun(
			fmt.Sprintf("create %s/%s sharing link for %s", c.Type, c.Scope, c.ItemPath),
			map[string]any{
				"action":   "sharepoint.share",
				"driveId":  c.DriveID,
				"itemPath": c.ItemPath,
				"type":     c.Type,
				"scope":    c.Scope,
			},
		)
	}

	client := ctx.NewMCPClient(spEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "shareFileOrFolder", map[string]any{
		"driveId":  c.DriveID,
		"itemPath": c.ItemPath,
		"type":     c.Type,
		"scope":    c.Scope,
	})
	if err != nil {
		return fmt.Errorf("share: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Sharing link created", data)
}

// --- setSensitivityLabelOnFile ---

// SPLabelCmd sets a sensitivity label on a file.
type SPLabelCmd struct {
	DriveID string `arg:"" help:"Drive ID"`
	ItemID  string `arg:"" help:"Item ID of the file"`
	LabelID string `arg:"" help:"Sensitivity label ID to apply"`
}

func (c *SPLabelCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.Output.PrintDryRun(
			fmt.Sprintf("set sensitivity label %s on item %s", c.LabelID, c.ItemID),
			map[string]any{
				"action":  "sharepoint.label",
				"driveId": c.DriveID,
				"itemId":  c.ItemID,
				"labelId": c.LabelID,
			},
		)
	}

	client := ctx.NewMCPClient(spEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "setSensitivityLabelOnFile", map[string]any{
		"driveId": c.DriveID,
		"itemId":  c.ItemID,
		"labelId": c.LabelID,
	})
	if err != nil {
		return fmt.Errorf("set label: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Sensitivity label set", data)
}

// --- checkOperationStatus ---

// SPStatusCmd checks the status of an async operation.
type SPStatusCmd struct {
	OperationURL string `arg:"" help:"Operation URL returned by an async operation"`
}

func (c *SPStatusCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(spEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "checkOperationStatus", map[string]any{
		"operationUrl": c.OperationURL,
	})
	if err != nil {
		return fmt.Errorf("check status: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}
