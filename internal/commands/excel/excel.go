package excel

import (
	"fmt"
	"strings"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
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

func contentLineCount(content string) int {
	if content == "" {
		return 0
	}
	return strings.Count(content, "\n") + 1
}

// ExcelCreateCmd creates a new Excel workbook.
type ExcelCreateCmd struct {
	FileName   string `arg:"" help:"Desired file name for the new workbook"`
	CSVContent string `help:"CSV content to populate the workbook" name:"csv-content" optional:"" default:""`
}

func (c *ExcelCreateCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"fileName":   c.FileName,
		"csvContent": c.CSVContent,
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(excelEndpoint(), "CreateWorkbook",
			fmt.Sprintf("create Excel workbook %q", c.FileName),
			map[string]any{
				"action":   "excel.create",
				"fileName": c.FileName,
				"csvBytes": len(c.CSVContent),
				"rowCount": contentLineCount(c.CSVContent),
			},
			args,
		)
	}

	data, err := ctx.CallToolData(excelEndpoint(), "CreateWorkbook", "create workbook", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Workbook created", data)
}

// ExcelGetCmd gets Excel workbook content.
type ExcelGetCmd struct {
	URL string `arg:"" help:"SharePoint sharing URL for the workbook"`
}

func (c *ExcelGetCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(excelEndpoint(), "GetDocumentContent", "get workbook content", map[string]any{
		"url": c.URL,
	})
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
	args := map[string]any{
		"driveId":     c.DriveID,
		"documentId":  c.DocumentID,
		"cellAddress": c.CellAddress,
		"content":     c.Text,
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(excelEndpoint(), "CreateComment",
			fmt.Sprintf("add comment to workbook %s at cell %s", c.DocumentID, c.CellAddress),
			map[string]any{
				"action":       "excel.comment",
				"driveId":      c.DriveID,
				"documentId":   c.DocumentID,
				"cellAddress":  c.CellAddress,
				"contentBytes": len(c.Text),
				"lineCount":    contentLineCount(c.Text),
			},
			args,
		)
	}

	data, err := ctx.CallToolData(excelEndpoint(), "CreateComment", "add comment", args)
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
	args := map[string]any{
		"commentId":  c.CommentID,
		"driveId":    c.DriveID,
		"documentId": c.DocumentID,
		"newComment": c.Text,
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(excelEndpoint(), "ReplyToComment",
			fmt.Sprintf("reply to comment %s on workbook %s", c.CommentID, c.DocumentID),
			map[string]any{
				"action":       "excel.reply",
				"commentId":    c.CommentID,
				"driveId":      c.DriveID,
				"documentId":   c.DocumentID,
				"contentBytes": len(c.Text),
				"lineCount":    contentLineCount(c.Text),
			},
			args,
		)
	}

	data, err := ctx.CallToolData(excelEndpoint(), "ReplyToComment", "reply to comment", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Reply added", data)
}
