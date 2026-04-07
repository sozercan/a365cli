package onedriveremote

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/output"
)

// odrEndpoint returns the agent365 endpoint for the OneDrive Remote MCP server.
func odrEndpoint() string {
	return config.Endpoint("onedrive-remote")
}

// OneDriveRemoteCmd groups all personal OneDrive subcommands.
type OneDriveRemoteCmd struct {
	Info   ODRInfoCmd   `cmd:"" help:"Get OneDrive info"`
	Ls     ODRLsCmd     `cmd:"" help:"List files and folders"`
	Search ODRSearchCmd `cmd:"" help:"Search for files"`
	Get    ODRGetCmd    `cmd:"" help:"Get file/folder metadata"`
	Cat    ODRCatCmd    `cmd:"" help:"Read a text file"`
	Mkdir  ODRMkdirCmd  `cmd:"" help:"Create a folder"`
	Write  ODRWriteCmd  `cmd:"" help:"Create a text file"`
	Rename ODRRenameCmd `cmd:"" help:"Rename a file or folder"`
	Mv     ODRMvCmd     `cmd:"" help:"Move a file"`
	Rm     ODRRmCmd     `cmd:"" help:"Delete a file or folder"`
	Share  ODRShareCmd  `cmd:"" help:"Share a file or folder"`
	Label  ODRLabelCmd  `cmd:"" help:"Set sensitivity label"`
}

// --- getOnedrive ---

// ODRInfoCmd gets OneDrive drive info.
type ODRInfoCmd struct{}

func (c *ODRInfoCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(odrEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "getOnedrive", map[string]any{})
	if err != nil {
		return fmt.Errorf("get OneDrive info: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- getFolderChildrenInMyOnedrive ---

// ODRLsCmd lists files and folders in the root of OneDrive.
type ODRLsCmd struct{}

func (c *ODRLsCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(odrEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "getFolderChildrenInMyOnedrive", map[string]any{})
	if err != nil {
		return fmt.Errorf("list folder: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- findFileOrFolderInMyDrive ---

// ODRSearchCmd searches for files or folders in OneDrive.
type ODRSearchCmd struct {
	Query string `arg:"" help:"Search query"`
}

func (c *ODRSearchCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(odrEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "findFileOrFolderInMyDrive", map[string]any{
		"searchQuery": c.Query,
	})
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- getFileOrFolderMetadataInMyOnedrive ---

// ODRGetCmd gets file or folder metadata by ID or URL.
type ODRGetCmd struct {
	FileOrFolderID string `help:"File or folder ID" optional:""`
	URL            string `help:"File or folder URL (alternative to ID)" optional:""`
}

func (c *ODRGetCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(odrEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	var toolName string
	args := map[string]any{}

	if c.URL != "" {
		toolName = "getFileOrFolderMetadataByUrl"
		args["fileOrFolderUrl"] = c.URL
	} else {
		toolName = "getFileOrFolderMetadataInMyOnedrive"
		if c.FileOrFolderID != "" {
			args["fileOrFolderId"] = c.FileOrFolderID
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

// --- readSmallTextFileFromMyOnedrive ---

// ODRCatCmd reads a small text file from OneDrive.
type ODRCatCmd struct {
	FileID string `arg:"" help:"File ID"`
}

func (c *ODRCatCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(odrEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "readSmallTextFileFromMyOnedrive", map[string]any{
		"fileId": c.FileID,
	})
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// --- createFolderInMyOnedrive ---

// ODRMkdirCmd creates a folder in OneDrive.
type ODRMkdirCmd struct {
	FolderName string `arg:"" help:"Name for the new folder"`
}

func (c *ODRMkdirCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(odrEndpoint(), "createFolderInMyOnedrive",
			fmt.Sprintf("create folder %q", c.FolderName),
			map[string]any{
				"action":     "onedrive-remote.mkdir",
				"folderName": c.FolderName,
			},
		)
	}

	client := ctx.NewMCPClient(odrEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "createFolderInMyOnedrive", map[string]any{
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

// --- createSmallTextFileInMyOnedrive ---

// ODRWriteCmd creates a text file in OneDrive.
type ODRWriteCmd struct {
	Filename    string `arg:"" help:"File name to create"`
	ContentText string `arg:"" help:"Text content for the file"`
}

func (c *ODRWriteCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(odrEndpoint(), "createSmallTextFileInMyOnedrive",
			fmt.Sprintf("create file %q", c.Filename),
			map[string]any{
				"action":      "onedrive-remote.write",
				"filename":    c.Filename,
				"contentText": c.ContentText,
			},
		)
	}

	client := ctx.NewMCPClient(odrEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "createSmallTextFileInMyOnedrive", map[string]any{
		"filename":    c.Filename,
		"contentText": c.ContentText,
	})
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("File created", data)
}

// --- renameFileOrFolderInMyOnedrive ---

// ODRRenameCmd renames a file or folder in OneDrive.
type ODRRenameCmd struct {
	FileOrFolderID      string `arg:"" help:"File or folder ID"`
	NewFileOrFolderName string `arg:"" help:"New name"`
	Etag                string `arg:"" help:"ETag for concurrency control"`
}

func (c *ODRRenameCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(odrEndpoint(), "renameFileOrFolderInMyOnedrive",
			fmt.Sprintf("rename %s to %q", c.FileOrFolderID, c.NewFileOrFolderName),
			map[string]any{
				"action":              "onedrive-remote.rename",
				"fileOrFolderId":      c.FileOrFolderID,
				"newFileOrFolderName": c.NewFileOrFolderName,
				"etag":               c.Etag,
			},
		)
	}

	client := ctx.NewMCPClient(odrEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "renameFileOrFolderInMyOnedrive", map[string]any{
		"fileOrFolderId":      c.FileOrFolderID,
		"newFileOrFolderName": c.NewFileOrFolderName,
		"etag":                c.Etag,
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

// --- moveSmallFileInMyOnedrive ---

// ODRMvCmd moves a file in OneDrive.
type ODRMvCmd struct {
	FileID            string `arg:"" help:"File ID"`
	NewParentFolderID string `arg:"" help:"Destination folder ID"`
	Etag              string `arg:"" help:"ETag for concurrency control"`
}

func (c *ODRMvCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(odrEndpoint(), "moveSmallFileInMyOnedrive",
			fmt.Sprintf("move file %s to folder %s", c.FileID, c.NewParentFolderID),
			map[string]any{
				"action":            "onedrive-remote.mv",
				"fileId":            c.FileID,
				"newParentFolderId": c.NewParentFolderID,
				"etag":             c.Etag,
			},
		)
	}

	client := ctx.NewMCPClient(odrEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "moveSmallFileInMyOnedrive", map[string]any{
		"fileId":            c.FileID,
		"newParentFolderId": c.NewParentFolderID,
		"etag":              c.Etag,
	})
	if err != nil {
		return fmt.Errorf("move file: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("File moved", data)
}

// --- deleteFileOrFolderInMyOnedrive ---

// ODRRmCmd deletes a file or folder from OneDrive.
type ODRRmCmd struct {
	FileOrFolderID string `arg:"" help:"File or folder ID"`
	Etag           string `arg:"" help:"ETag for concurrency control"`
}

func (c *ODRRmCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(odrEndpoint(), "deleteFileOrFolderInMyOnedrive",
			fmt.Sprintf("delete %s", c.FileOrFolderID),
			map[string]any{
				"action":         "onedrive-remote.rm",
				"fileOrFolderId": c.FileOrFolderID,
				"etag":           c.Etag,
			},
		)
	}

	if err := ctx.Confirm(fmt.Sprintf("delete %s", c.FileOrFolderID)); err != nil {
		return err
	}

	client := ctx.NewMCPClient(odrEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "deleteFileOrFolderInMyOnedrive", map[string]any{
		"fileOrFolderId": c.FileOrFolderID,
		"etag":           c.Etag,
	})
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Deleted", data)
}

// --- shareFileOrFolderInMyOnedrive ---

// ODRShareCmd shares a file or folder in OneDrive.
type ODRShareCmd struct {
	FileOrFolderID  string   `arg:"" help:"File or folder ID"`
	RecipientEmails []string `arg:"" help:"Recipient email addresses"`
	Roles           string   `help:"Sharing roles" required:""`
}

func (c *ODRShareCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(odrEndpoint(), "shareFileOrFolderInMyOnedrive",
			fmt.Sprintf("share %s with %v", c.FileOrFolderID, c.RecipientEmails),
			map[string]any{
				"action":          "onedrive-remote.share",
				"fileOrFolderId":  c.FileOrFolderID,
				"recipientEmails": c.RecipientEmails,
				"roles":           c.Roles,
			},
		)
	}

	client := ctx.NewMCPClient(odrEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "shareFileOrFolderInMyOnedrive", map[string]any{
		"fileOrFolderId":  c.FileOrFolderID,
		"recipientEmails": c.RecipientEmails,
		"roles":           c.Roles,
	})
	if err != nil {
		return fmt.Errorf("share: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Shared", data)
}

// --- setSensitivityLabelOnFileInMyOnedrive ---

// ODRLabelCmd sets a sensitivity label on a file in OneDrive.
type ODRLabelCmd struct {
	FileID             string `arg:"" help:"File ID"`
	SensitivityLabelID string `arg:"" help:"Sensitivity label ID to apply"`
}

func (c *ODRLabelCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(odrEndpoint(), "setSensitivityLabelOnFileInMyOnedrive",
			fmt.Sprintf("set sensitivity label %s on file %s", c.SensitivityLabelID, c.FileID),
			map[string]any{
				"action":             "onedrive-remote.label",
				"fileId":             c.FileID,
				"sensitivityLabelId": c.SensitivityLabelID,
			},
		)
	}

	client := ctx.NewMCPClient(odrEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "setSensitivityLabelOnFileInMyOnedrive", map[string]any{
		"fileId":             c.FileID,
		"sensitivityLabelId": c.SensitivityLabelID,
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
