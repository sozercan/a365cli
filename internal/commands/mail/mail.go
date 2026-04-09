package mail

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/output"
)

// MailCmd groups all Mail subcommands.
type MailCmd struct {
	Search         MailSearchCmd         `cmd:"" help:"Search emails (OData query)"`
	SearchNL       MailSearchNLCmd       `cmd:"" name:"search-nl" help:"Search emails (natural language)"`
	Get            MailGetCmd            `cmd:"" help:"Get an email by ID"`
	Send           MailSendCmd           `cmd:"" help:"Send an email"`
	Reply          MailReplyCmd          `cmd:"" help:"Reply to an email"`
	ReplyAll       MailReplyAllCmd       `cmd:"" name:"reply-all" help:"Reply all to an email"`
	Forward        MailForwardCmd        `cmd:"" help:"Forward an email"`
	Delete         MailDeleteCmd         `cmd:"" help:"Delete an email"`
	Flag           MailFlagCmd           `cmd:"" help:"Flag/unflag an email"`
	Draft          MailDraftCmd          `cmd:"" help:"Create a draft email"`
	SendDraft      MailSendDraftCmd      `cmd:"" name:"send-draft" help:"Send a draft email"`
	Attachments    MailAttachmentsCmd    `cmd:"" help:"List attachments on an email"`
	Update         MailUpdateCmd         `cmd:"" help:"Update an email"`
	Download       MailDownloadCmd       `cmd:"" help:"Download an attachment"`
	Upload         MailUploadCmd         `cmd:"" name:"upload-attachment" help:"Upload an attachment to a message"`
	DeleteAttach   MailDeleteAttachCmd   `cmd:"" name:"delete-attachment" help:"Delete an attachment"`
	UpdateDraft    MailUpdateDraftCmd    `cmd:"" name:"update-draft" help:"Update a draft email"`
	DraftAttach    MailDraftAttachCmd    `cmd:"" name:"draft-attachments" help:"Add attachments to a draft"`
	ReplyThread    MailReplyThreadCmd    `cmd:"" name:"reply-thread" help:"Reply with full conversation thread"`
	ReplyAllThread MailReplyAllThreadCmd `cmd:"" name:"reply-all-thread" help:"Reply-all with full conversation thread"`
	ForwardThread  MailForwardThreadCmd  `cmd:"" name:"forward-thread" help:"Forward with full conversation thread"`
}

func mailEndpoint() string {
	return config.Endpoint("mail")
}

// MailSearchCmd searches emails with OData query parameters.
type MailSearchCmd struct {
	Query string `arg:"" help:"Search term or OData query parameters (e.g. 'budget' or '?$search=\\\"budget\\\"' or '?$filter=isRead eq false')"`
	Max   int    `help:"Maximum number of results" default:"20"`
}

func (c *MailSearchCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(mailEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	// Auto-wrap bare search terms into OData $search query parameters.
	// If the user provides a raw OData string (starting with ? or $), pass it through.
	query := c.Query
	if query != "" && query[0] != '?' && query[0] != '$' {
		query = `?$search="` + query + `"`
	}

	resp, err := client.CallTool(ctx.Ctx, "SearchMessagesQueryParameters", map[string]any{
		"queryParameters": query,
	})
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	rows := output.ToRows(data, "messages")
	if rows == nil {
		rows = output.ToRows(data, "value")
	}
	if rows == nil {
		return ctx.Output.PrintItem(data)
	}
	if c.Max > 0 && len(rows) > c.Max {
		rows = rows[:c.Max]
	}
	return ctx.Output.PrintList("messages", output.MailColumns, rows)
}

// MailSearchNLCmd searches emails with natural language.
type MailSearchNLCmd struct {
	Query string `arg:"" help:"Natural language query (e.g. 'emails from John about budget')"`
}

func (c *MailSearchNLCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(mailEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "SearchMessages", map[string]any{
		"message": c.Query,
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

// MailGetCmd gets an email by ID.
type MailGetCmd struct {
	ID string `arg:"" help:"Message ID"`
}

func (c *MailGetCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(mailEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "GetMessage", map[string]any{
		"id": c.ID,
	})
	if err != nil {
		return fmt.Errorf("get message: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// MailSendCmd sends an email.
type MailSendCmd struct {
	To      []string `arg:"" help:"Recipient email addresses"`
	Subject string   `help:"Email subject" required:""`
	Body    string   `help:"Email body" required:""`
	CC      []string `help:"CC recipients" optional:""`
	BCC     []string `help:"BCC recipients" optional:""`
}

func (c *MailSendCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		mcpArgs := map[string]any{
			"to":      c.To,
			"subject": c.Subject,
			"body":    c.Body,
		}
		if len(c.CC) > 0 {
			mcpArgs["cc"] = c.CC
		}
		if len(c.BCC) > 0 {
			mcpArgs["bcc"] = c.BCC
		}
		return ctx.ValidateDryRun(mailEndpoint(), "SendEmailWithAttachments",
			fmt.Sprintf("send email to %v with subject %q", c.To, c.Subject),
			map[string]any{"action": "mail.send", "to": c.To, "subject": c.Subject, "body_len": len(c.Body)},
			mcpArgs,
		)
	}

	client := ctx.NewMCPClient(mailEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	args := map[string]any{
		"to":      c.To,
		"subject": c.Subject,
		"body":    c.Body,
	}
	if len(c.CC) > 0 {
		args["cc"] = c.CC
	}
	if len(c.BCC) > 0 {
		args["bcc"] = c.BCC
	}

	resp, err := client.CallTool(ctx.Ctx, "SendEmailWithAttachments", args)
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Email sent", data)
}

// MailReplyCmd replies to an email.
type MailReplyCmd struct {
	ID      string `arg:"" help:"Message ID to reply to"`
	Comment string `arg:"" help:"Reply text"`
	Send    bool   `help:"Send immediately (otherwise creates draft)" default:"true"`
}

func (c *MailReplyCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(mailEndpoint(), "ReplyToMessage",
			fmt.Sprintf("reply to message %s", c.ID),
			map[string]any{"action": "mail.reply", "id": c.ID, "comment_len": len(c.Comment)},
			map[string]any{"id": c.ID, "comment": c.Comment, "sendImmediately": c.Send},
		)
	}

	client := ctx.NewMCPClient(mailEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "ReplyToMessage", map[string]any{
		"id":              c.ID,
		"comment":         c.Comment,
		"sendImmediately": c.Send,
	})
	if err != nil {
		return fmt.Errorf("reply: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	msg := "Reply sent"
	if !c.Send {
		msg = "Reply draft created"
	}
	return ctx.Output.PrintMutation(msg, data)
}

// MailReplyAllCmd replies all to an email.
type MailReplyAllCmd struct {
	ID      string `arg:"" help:"Message ID to reply-all to"`
	Comment string `arg:"" help:"Reply text"`
	Send    bool   `help:"Send immediately (otherwise creates draft)" default:"true"`
}

func (c *MailReplyAllCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(mailEndpoint(), "ReplyAllToMessage",
			fmt.Sprintf("reply-all to message %s", c.ID),
			map[string]any{"action": "mail.reply-all", "id": c.ID},
			map[string]any{"id": c.ID, "comment": c.Comment, "sendImmediately": c.Send},
		)
	}

	client := ctx.NewMCPClient(mailEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "ReplyAllToMessage", map[string]any{
		"id":              c.ID,
		"comment":         c.Comment,
		"sendImmediately": c.Send,
	})
	if err != nil {
		return fmt.Errorf("reply-all: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	msg := "Reply-all sent"
	if !c.Send {
		msg = "Reply-all draft created"
	}
	return ctx.Output.PrintMutation(msg, data)
}

// MailForwardCmd forwards an email.
type MailForwardCmd struct {
	ID      string   `arg:"" help:"Message ID to forward"`
	To      []string `arg:"" help:"Forward recipients"`
	Comment string   `help:"Introductory comment" optional:""`
}

func (c *MailForwardCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		mcpArgs := map[string]any{
			"messageId":    c.ID,
			"additionalTo": c.To,
		}
		if c.Comment != "" {
			mcpArgs["introComment"] = c.Comment
		}
		return ctx.ValidateDryRun(mailEndpoint(), "ForwardMessage",
			fmt.Sprintf("forward message %s to %v", c.ID, c.To),
			map[string]any{"action": "mail.forward", "id": c.ID, "to": c.To},
			mcpArgs,
		)
	}

	client := ctx.NewMCPClient(mailEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	args := map[string]any{
		"messageId":    c.ID,
		"additionalTo": c.To,
	}
	if c.Comment != "" {
		args["introComment"] = c.Comment
	}

	resp, err := client.CallTool(ctx.Ctx, "ForwardMessage", args)
	if err != nil {
		return fmt.Errorf("forward: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Email forwarded", data)
}

// MailDeleteCmd deletes an email.
type MailDeleteCmd struct {
	ID string `arg:"" help:"Message ID"`
}

func (c *MailDeleteCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(mailEndpoint(), "DeleteMessage",
			fmt.Sprintf("delete message %s", c.ID),
			map[string]any{"action": "mail.delete", "id": c.ID},
		)
	}
	if err := ctx.Confirm(fmt.Sprintf("delete email %s", c.ID)); err != nil {
		return err
	}

	client := ctx.NewMCPClient(mailEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "DeleteMessage", map[string]any{"id": c.ID})
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Email deleted", data)
}

// MailFlagCmd flags or unflags an email.
type MailFlagCmd struct {
	ID     string `arg:"" help:"Message ID"`
	Status string `arg:"" help:"Flag status: Flagged, Complete, or NotFlagged" enum:"Flagged,Complete,NotFlagged"`
}

func (c *MailFlagCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(mailEndpoint(), "FlagEmail",
			fmt.Sprintf("flag email %s as %s", c.ID, c.Status),
			map[string]any{"action": "mail.flag", "messageId": c.ID, "flagStatus": c.Status},
		)
	}

	client := ctx.NewMCPClient(mailEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "FlagEmail", map[string]any{
		"messageId":  c.ID,
		"flagStatus": c.Status,
	})
	if err != nil {
		return fmt.Errorf("flag: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation(fmt.Sprintf("Email flagged as %s", c.Status), data)
}

// MailDraftCmd creates a draft email.
type MailDraftCmd struct {
	Subject string   `help:"Email subject" optional:""`
	Body    string   `help:"Email body" optional:""`
	To      []string `help:"To recipients" optional:""`
	CC      []string `help:"CC recipients" optional:""`
}

func (c *MailDraftCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(mailEndpoint(), "CreateDraftMessage", "create draft email",
			map[string]any{"action": "mail.draft", "subject": c.Subject, "to": c.To},
		)
	}

	client := ctx.NewMCPClient(mailEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	args := map[string]any{}
	if c.Subject != "" {
		args["subject"] = c.Subject
	}
	if c.Body != "" {
		args["body"] = c.Body
	}
	if len(c.To) > 0 {
		args["to"] = c.To
	}
	if len(c.CC) > 0 {
		args["cc"] = c.CC
	}

	resp, err := client.CallTool(ctx.Ctx, "CreateDraftMessage", args)
	if err != nil {
		return fmt.Errorf("create draft: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Draft created", data)
}

// MailSendDraftCmd sends a draft email.
type MailSendDraftCmd struct {
	ID string `arg:"" help:"Draft message ID"`
}

func (c *MailSendDraftCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(mailEndpoint(), "SendDraftMessage",
			fmt.Sprintf("send draft %s", c.ID),
			map[string]any{"action": "mail.send-draft", "id": c.ID},
		)
	}

	client := ctx.NewMCPClient(mailEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "SendDraftMessage", map[string]any{"id": c.ID})
	if err != nil {
		return fmt.Errorf("send draft: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Draft sent", data)
}

// MailAttachmentsCmd lists attachments on an email.
type MailAttachmentsCmd struct {
	ID string `arg:"" help:"Message ID"`
}

func (c *MailAttachmentsCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(mailEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "GetAttachments", map[string]any{
		"messageId": c.ID,
	})
	if err != nil {
		return fmt.Errorf("get attachments: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// MailUpdateCmd updates an existing email.
type MailUpdateCmd struct {
	ID         string   `arg:"" help:"Message ID"`
	Subject    string   `help:"New subject" optional:""`
	Body       string   `help:"New body" optional:""`
	Importance string   `help:"Importance: Low, Normal, or High" optional:""`
	Categories []string `help:"Categories to set" optional:""`
}

func (c *MailUpdateCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(mailEndpoint(), "UpdateMessage",
			fmt.Sprintf("update message %s", c.ID),
			map[string]any{"action": "mail.update", "id": c.ID},
		)
	}

	client := ctx.NewMCPClient(mailEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	args := map[string]any{
		"id": c.ID,
	}
	if c.Subject != "" {
		args["subject"] = c.Subject
	}
	if c.Body != "" {
		args["body"] = c.Body
	}
	if c.Importance != "" {
		args["importance"] = c.Importance
	}
	if len(c.Categories) > 0 {
		args["categories"] = c.Categories
	}

	resp, err := client.CallTool(ctx.Ctx, "UpdateMessage", args)
	if err != nil {
		return fmt.Errorf("update message: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Email updated", data)
}

// MailDownloadCmd downloads an attachment from a message.
type MailDownloadCmd struct {
	MessageID    string `arg:"" help:"Message ID"`
	AttachmentID string `arg:"" help:"Attachment ID"`
}

func (c *MailDownloadCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(mailEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "DownloadAttachment", map[string]any{
		"messageId":    c.MessageID,
		"attachmentId": c.AttachmentID,
	})
	if err != nil {
		return fmt.Errorf("download attachment: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// MailUploadCmd uploads an attachment to a message.
type MailUploadCmd struct {
	MessageID     string `arg:"" help:"Message ID"`
	FileName      string `arg:"" help:"File name for the attachment"`
	ContentBase64 string `arg:"" help:"Base64-encoded file content"`
	Large         bool   `help:"Use large attachment upload (for files >3MB)" default:"false"`
}

func (c *MailUploadCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		toolName := "UploadAttachment"
		if c.Large {
			toolName = "UploadLargeAttachment"
		}
		return ctx.ValidateDryRun(mailEndpoint(), toolName,
			fmt.Sprintf("upload attachment %q to message %s", c.FileName, c.MessageID),
			map[string]any{"action": "mail.upload-attachment", "messageId": c.MessageID, "fileName": c.FileName},
			map[string]any{
				"messageId":     c.MessageID,
				"fileName":      c.FileName,
				"contentBase64": c.ContentBase64,
			},
		)
	}

	client := ctx.NewMCPClient(mailEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	toolName := "UploadAttachment"
	if c.Large {
		toolName = "UploadLargeAttachment"
	}

	resp, err := client.CallTool(ctx.Ctx, toolName, map[string]any{
		"messageId":     c.MessageID,
		"fileName":      c.FileName,
		"contentBase64": c.ContentBase64,
	})
	if err != nil {
		return fmt.Errorf("upload attachment: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Attachment uploaded", data)
}

// MailDeleteAttachCmd deletes an attachment from a message.
type MailDeleteAttachCmd struct {
	MessageID    string `arg:"" help:"Message ID"`
	AttachmentID string `arg:"" help:"Attachment ID"`
}

func (c *MailDeleteAttachCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(mailEndpoint(), "DeleteAttachment",
			fmt.Sprintf("delete attachment %s from message %s", c.AttachmentID, c.MessageID),
			map[string]any{"action": "mail.delete-attachment", "messageId": c.MessageID, "attachmentId": c.AttachmentID},
		)
	}
	if err := ctx.Confirm(fmt.Sprintf("delete attachment %s from message %s", c.AttachmentID, c.MessageID)); err != nil {
		return err
	}

	client := ctx.NewMCPClient(mailEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "DeleteAttachment", map[string]any{
		"messageId":    c.MessageID,
		"attachmentId": c.AttachmentID,
	})
	if err != nil {
		return fmt.Errorf("delete attachment: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Attachment deleted", data)
}

// MailUpdateDraftCmd updates a draft email.
type MailUpdateDraftCmd struct {
	MessageID string   `arg:"" help:"Draft message ID"`
	To        []string `help:"To recipients" optional:""`
	CC        []string `help:"CC recipients" optional:""`
	BCC       []string `help:"BCC recipients" optional:""`
	Subject   string   `help:"New subject" optional:""`
	Body      string   `help:"New body" optional:""`
}

func (c *MailUpdateDraftCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(mailEndpoint(), "UpdateDraft",
			fmt.Sprintf("update draft %s", c.MessageID),
			map[string]any{"action": "mail.update-draft", "messageId": c.MessageID},
		)
	}

	client := ctx.NewMCPClient(mailEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	args := map[string]any{
		"messageId": c.MessageID,
	}
	if len(c.To) > 0 {
		args["to"] = c.To
	}
	if len(c.CC) > 0 {
		args["cc"] = c.CC
	}
	if len(c.BCC) > 0 {
		args["bcc"] = c.BCC
	}
	if c.Subject != "" {
		args["subject"] = c.Subject
	}
	if c.Body != "" {
		args["body"] = c.Body
	}

	resp, err := client.CallTool(ctx.Ctx, "UpdateDraft", args)
	if err != nil {
		return fmt.Errorf("update draft: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Draft updated", data)
}

// MailDraftAttachCmd adds attachments to a draft.
type MailDraftAttachCmd struct {
	MessageID      string   `arg:"" help:"Draft message ID"`
	AttachmentUris []string `arg:"" help:"URIs of attachments to add"`
}

func (c *MailDraftAttachCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(mailEndpoint(), "AddDraftAttachments",
			fmt.Sprintf("add attachments to draft %s", c.MessageID),
			map[string]any{"action": "mail.draft-attachments", "messageId": c.MessageID, "attachmentUris": c.AttachmentUris},
		)
	}

	client := ctx.NewMCPClient(mailEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "AddDraftAttachments", map[string]any{
		"messageId":      c.MessageID,
		"attachmentUris": c.AttachmentUris,
	})
	if err != nil {
		return fmt.Errorf("add draft attachments: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Attachments added to draft", data)
}

// MailReplyThreadCmd replies with full conversation thread.
type MailReplyThreadCmd struct {
	MessageID string `arg:"" help:"Message ID to reply to"`
}

func (c *MailReplyThreadCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(mailEndpoint(), "ReplyWithFullThread",
			fmt.Sprintf("reply with full thread to message %s", c.MessageID),
			map[string]any{"action": "mail.reply-thread", "messageId": c.MessageID},
		)
	}

	client := ctx.NewMCPClient(mailEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "ReplyWithFullThread", map[string]any{
		"messageId": c.MessageID,
	})
	if err != nil {
		return fmt.Errorf("reply with thread: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Reply with thread sent", data)
}

// MailReplyAllThreadCmd replies-all with full conversation thread.
type MailReplyAllThreadCmd struct {
	MessageID string `arg:"" help:"Message ID to reply-all to"`
}

func (c *MailReplyAllThreadCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(mailEndpoint(), "ReplyAllWithFullThread",
			fmt.Sprintf("reply-all with full thread to message %s", c.MessageID),
			map[string]any{"action": "mail.reply-all-thread", "messageId": c.MessageID},
		)
	}

	client := ctx.NewMCPClient(mailEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "ReplyAllWithFullThread", map[string]any{
		"messageId": c.MessageID,
	})
	if err != nil {
		return fmt.Errorf("reply-all with thread: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Reply-all with thread sent", data)
}

// MailForwardThreadCmd forwards a message with full conversation thread.
type MailForwardThreadCmd struct {
	MessageID string `arg:"" help:"Message ID to forward"`
}

func (c *MailForwardThreadCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(mailEndpoint(), "ForwardMessageWithFullThread",
			fmt.Sprintf("forward with full thread message %s", c.MessageID),
			map[string]any{"action": "mail.forward-thread", "messageId": c.MessageID},
		)
	}

	client := ctx.NewMCPClient(mailEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "ForwardMessageWithFullThread", map[string]any{
		"messageId": c.MessageID,
	})
	if err != nil {
		return fmt.Errorf("forward with thread: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Forward with thread sent", data)
}
