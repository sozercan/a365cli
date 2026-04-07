package word

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/output"
)

// WordCmd groups all Word subcommands.
type WordCmd struct {
	Create  WordCreateCmd  `cmd:"" help:"Create a new Word document"`
	Get     WordGetCmd     `cmd:"" help:"Get Word document content"`
	Comment WordCommentCmd `cmd:"" help:"Add a comment to a document"`
	Reply   WordReplyCmd   `cmd:"" help:"Reply to a document comment"`
}

func wordEndpoint() string {
	return config.Endpoint("word")
}

// WordCreateCmd creates a new Word document.
type WordCreateCmd struct {
	FileName string `arg:"" help:"Desired file name for the new document"`
}

func (c *WordCreateCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(wordEndpoint(), "CreateDocument",
			fmt.Sprintf("create Word document %q", c.FileName),
			map[string]any{"action": "word.create", "desiredFileName": c.FileName},
		)
	}

	client := ctx.NewMCPClient(wordEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "CreateDocument", map[string]any{
		"desiredFileName": c.FileName,
	})
	if err != nil {
		return fmt.Errorf("create document: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Document created", data)
}

// WordGetCmd gets Word document content.
type WordGetCmd struct {
	DriveID    string `arg:"" help:"Drive ID"`
	DocumentID string `arg:"" help:"Document ID"`
}

func (c *WordGetCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(wordEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "GetDocumentContent", map[string]any{
		"driveId":    c.DriveID,
		"documentId": c.DocumentID,
	})
	if err != nil {
		return fmt.Errorf("get document content: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// WordCommentCmd adds a comment to a Word document.
type WordCommentCmd struct {
	DriveID    string `arg:"" help:"Drive ID"`
	DocumentID string `arg:"" help:"Document ID"`
	Text       string `arg:"" help:"Comment text"`
}

func (c *WordCommentCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(wordEndpoint(), "AddComment",
			fmt.Sprintf("add comment to document %s", c.DocumentID),
			map[string]any{"action": "word.comment", "driveId": c.DriveID, "documentId": c.DocumentID},
		)
	}

	client := ctx.NewMCPClient(wordEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "AddComment", map[string]any{
		"driveId":    c.DriveID,
		"documentId": c.DocumentID,
		"text":       c.Text,
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

// WordReplyCmd replies to a comment on a Word document.
type WordReplyCmd struct {
	CommentID  string `arg:"" help:"Comment ID to reply to"`
	DriveID    string `arg:"" help:"Drive ID"`
	DocumentID string `arg:"" help:"Document ID"`
	Text       string `arg:"" help:"Reply text"`
}

func (c *WordReplyCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(wordEndpoint(), "ReplyToComment",
			fmt.Sprintf("reply to comment %s on document %s", c.CommentID, c.DocumentID),
			map[string]any{"action": "word.reply", "commentId": c.CommentID, "driveId": c.DriveID, "documentId": c.DocumentID},
		)
	}

	client := ctx.NewMCPClient(wordEndpoint())
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
