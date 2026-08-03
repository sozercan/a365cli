package word

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
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
	FileName      string `arg:"" help:"Desired file name for the new document"`
	ContentInHTML string `help:"HTML or plain text content for the document body" name:"content" optional:"" default:""`
}

func (c *WordCreateCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"fileName":      c.FileName,
		"contentInHtml": c.ContentInHTML,
	}
	if ctx.DryRun {
		return ctx.ValidateDryRun(wordEndpoint(), "CreateDocument",
			fmt.Sprintf("create Word document %q", c.FileName),
			map[string]any{
				"action":        "word.create",
				"fileName":      c.FileName,
				"contentLength": len(c.ContentInHTML),
			},
			args,
		)
	}

	data, err := ctx.CallToolData(wordEndpoint(), "CreateDocument", "create document", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Document created", data)
}

// WordGetCmd gets Word document content.
type WordGetCmd struct {
	URL string `arg:"" help:"SharePoint sharing URL for the document"`
}

func (c *WordGetCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(wordEndpoint(), "GetDocumentContent", "get document content", map[string]any{
		"url": c.URL,
	})
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
	args := map[string]any{
		"driveId":    c.DriveID,
		"documentId": c.DocumentID,
		"newComment": c.Text,
	}
	if ctx.DryRun {
		return ctx.ValidateDryRun(wordEndpoint(), "AddComment",
			fmt.Sprintf("add comment to document %s", c.DocumentID),
			map[string]any{
				"action":        "word.comment",
				"driveId":       c.DriveID,
				"documentId":    c.DocumentID,
				"commentLength": len(c.Text),
			},
			args,
		)
	}

	data, err := ctx.CallToolData(wordEndpoint(), "AddComment", "add comment", args)
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
	args := map[string]any{
		"commentId":  c.CommentID,
		"driveId":    c.DriveID,
		"documentId": c.DocumentID,
		"newComment": c.Text,
	}
	if ctx.DryRun {
		return ctx.ValidateDryRun(wordEndpoint(), "ReplyToComment",
			fmt.Sprintf("reply to comment %s on document %s", c.CommentID, c.DocumentID),
			map[string]any{
				"action":        "word.reply",
				"commentId":     c.CommentID,
				"driveId":       c.DriveID,
				"documentId":    c.DocumentID,
				"commentLength": len(c.Text),
			},
			args,
		)
	}

	data, err := ctx.CallToolData(wordEndpoint(), "ReplyToComment", "reply to comment", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Reply added", data)
}
