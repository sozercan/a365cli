package excel

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/output"
)

// ExcelCmd groups all Excel subcommands.
type ExcelCmd struct {
	Create  ExcelCreateCmd  `cmd:"" help:"Create a new Excel workbook"`
	Get     ExcelGetCmd     `cmd:"" help:"Get Excel workbook content"`
	Comment ExcelCommentCmd `cmd:"" help:"Add a comment to a workbook"`
	Reply   ExcelReplyCmd   `cmd:"" help:"Reply to a workbook comment"`
}

func excelEndpoint() string {
	return config.Endpoint("excel")
}

// ExcelCreateCmd creates a new Excel workbook.
type ExcelCreateCmd struct {
	FileName string `arg:"" help:"Desired file name for the new workbook"`
}

func (c *ExcelCreateCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(excelEndpoint(), "CreateWorkbook",
			fmt.Sprintf("create Excel workbook %q", c.FileName),
			map[string]any{"action": "excel.create", "desiredFileName": c.FileName},
		)
	}

	client := ctx.NewMCPClient(excelEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "CreateWorkbook", map[string]any{
		"desiredFileName": c.FileName,
	})
	if err != nil {
		return fmt.Errorf("create workbook: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Workbook created", data)
}

// ExcelGetCmd gets Excel workbook content.
type ExcelGetCmd struct {
	DriveID    string `arg:"" help:"Drive ID"`
	DocumentID string `arg:"" help:"Document ID"`
}

func (c *ExcelGetCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(excelEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "GetDocumentContent", map[string]any{
		"driveId":    c.DriveID,
		"documentId": c.DocumentID,
	})
	if err != nil {
		return fmt.Errorf("get workbook content: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// ExcelCommentCmd adds a comment to an Excel workbook.
type ExcelCommentCmd struct {
	DriveID     string `arg:"" help:"Drive ID"`
	DocumentID  string `arg:"" help:"Document ID"`
	CellAddress string `arg:"" help:"Cell address (e.g. A1, B2)"`
	Text        string `arg:"" help:"Comment text"`
}

func (c *ExcelCommentCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(excelEndpoint(), "CreateComment",
			fmt.Sprintf("add comment to workbook %s at cell %s", c.DocumentID, c.CellAddress),
			map[string]any{"action": "excel.comment", "driveId": c.DriveID, "documentId": c.DocumentID, "cellAddress": c.CellAddress},
		)
	}

	client := ctx.NewMCPClient(excelEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "CreateComment", map[string]any{
		"driveId":     c.DriveID,
		"documentId":  c.DocumentID,
		"cellAddress": c.CellAddress,
		"text":        c.Text,
	})
	if err != nil {
		return fmt.Errorf("add comment: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Comment added", data)
}

// ExcelReplyCmd replies to a comment on an Excel workbook.
type ExcelReplyCmd struct {
	CommentID  string `arg:"" help:"Comment ID to reply to"`
	DriveID    string `arg:"" help:"Drive ID"`
	DocumentID string `arg:"" help:"Document ID"`
	Text       string `arg:"" help:"Reply text"`
}

func (c *ExcelReplyCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(excelEndpoint(), "ReplyToComment",
			fmt.Sprintf("reply to comment %s on workbook %s", c.CommentID, c.DocumentID),
			map[string]any{"action": "excel.reply", "commentId": c.CommentID, "driveId": c.DriveID, "documentId": c.DocumentID},
		)
	}

	client := ctx.NewMCPClient(excelEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "ReplyToComment", map[string]any{
		"commentId":  c.CommentID,
		"driveId":    c.DriveID,
		"documentId": c.DocumentID,
		"text":       c.Text,
	})
	if err != nil {
		return fmt.Errorf("reply to comment: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Reply added", data)
}
