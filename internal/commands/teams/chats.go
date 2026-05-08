package teams

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/output"
)

// ChatsCmd groups chat subcommands.
type ChatsCmd struct {
	List          ChatsListCmd          `cmd:"" help:"List recent chats"`
	Get           ChatsGetCmd           `cmd:"" help:"Get a chat by ID"`
	Create        ChatsCreateCmd        `cmd:"" help:"Create or get a chat"`
	Delete        ChatsDeleteCmd        `cmd:"" help:"Delete a chat"`
	Update        ChatsUpdateCmd        `cmd:"" help:"Update a group chat topic"`
	Messages      ChatsMessagesCmd      `cmd:"" help:"List messages in a chat"`
	Send          ChatsSendCmd          `cmd:"" help:"Send a message to a chat"`
	SendSelf      ChatsSendSelfCmd      `cmd:"" name:"send-self" help:"Send a message/note to yourself"`
	GetMessage    ChatsGetMessageCmd    `cmd:"" name:"get-message" help:"Get a specific message from a chat"`
	UpdateMessage ChatsUpdateMessageCmd `cmd:"" name:"update-message" help:"Update a chat message"`
	DeleteMessage ChatsDeleteMessageCmd `cmd:"" name:"delete-message" help:"Delete a chat message"`
	Members       ChatsListMembersCmd   `cmd:"" help:"List members of a chat"`
	AddMember     ChatsAddMemberCmd     `cmd:"" name:"add-member" help:"Add a member to a chat"`
}

// ChatsListCmd lists recent chats.
type ChatsListCmd struct {
	Topic string   `help:"Filter chats by topic (case-insensitive, partial match)" optional:""`
	UPNs  []string `name:"upn" help:"Filter by member UPN emails" optional:""`
	Max   int      `help:"Maximum number of results" default:"50"`
}

func (c *ChatsListCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(teamsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	args := map[string]any{}
	if len(c.UPNs) > 0 {
		args["userUpns"] = c.UPNs
	} else if ctx.UserUPN != "" {
		args["userUpns"] = []string{ctx.UserUPN}
	} else {
		args["userUpns"] = []string{}
	}
	if c.Topic != "" {
		args["topic"] = c.Topic
	}
	if c.Max > 0 && c.Max != 50 {
		args["top"] = c.Max
	}

	resp, err := client.CallTool(ctx.Ctx, "ListChats", args)
	if err != nil {
		return fmt.Errorf("list chats: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	rows := output.ToRows(data, "chats")
	if c.Max > 0 && len(rows) > c.Max {
		rows = rows[:c.Max]
	}
	return ctx.Output.PrintList("chats", output.ChatsColumns, rows)
}

// ChatsGetCmd gets a specific chat.
type ChatsGetCmd struct {
	ChatID string `arg:"" help:"Chat ID"`
}

func (c *ChatsGetCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(teamsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "GetChat", map[string]any{
		"chatId": c.ChatID,
	})
	if err != nil {
		return fmt.Errorf("get chat: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// ChatsMessagesCmd lists messages in a chat.
type ChatsMessagesCmd struct {
	ChatID string `arg:"" help:"Chat ID"`
	Max    int    `help:"Maximum number of messages" default:"20"`
}

func (c *ChatsMessagesCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(teamsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "ListChatMessages", map[string]any{
		"chatId": c.ChatID,
	})
	if err != nil {
		return fmt.Errorf("list chat messages: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	rows := output.ToRows(data, "messages")
	if c.Max > 0 && len(rows) > c.Max {
		rows = rows[:c.Max]
	}
	return ctx.Output.PrintList("messages", output.MessagesColumns, rows)
}

// ChatsSendCmd sends a message to a chat.
type ChatsSendCmd struct {
	ChatID  string `arg:"" help:"Chat ID"`
	Message string `arg:"" help:"Message text"`
}

func (c *ChatsSendCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"chatId":  c.ChatID,
		"content": c.Message,
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(teamsEndpoint(), "PostMessage",
			fmt.Sprintf("send message to chat %s", c.ChatID),
			map[string]any{
				"action":  "chats.send",
				"chatId":  c.ChatID,
				"content": c.Message,
			},
			args,
		)
	}

	client := ctx.NewMCPClient(teamsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "PostMessage", args)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Message sent", data)
}

// ChatsSendSelfCmd sends a message/note to yourself in Teams.
type ChatsSendSelfCmd struct {
	Message string `arg:"" help:"Message text"`
}

func (c *ChatsSendSelfCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"content": c.Message,
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(teamsEndpoint(), "SendMessageToSelf", "send message to self", map[string]any{
			"action":  "chats.send-self",
			"content": c.Message,
		}, args)
	}

	client := ctx.NewMCPClient(teamsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "SendMessageToSelf", args)
	if err != nil {
		return fmt.Errorf("send to self: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Message sent to self", data)
}

// ChatsGetMessageCmd gets a specific message from a chat.
type ChatsGetMessageCmd struct {
	ChatID    string `arg:"" help:"Chat ID"`
	MessageID string `arg:"" help:"Message ID"`
}

func (c *ChatsGetMessageCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(teamsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "GetChatMessage", map[string]any{
		"chatId":    c.ChatID,
		"messageId": c.MessageID,
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

// --- Chat management commands ---

// ChatsCreateCmd creates or retrieves a chat.
type ChatsCreateCmd struct {
	Type    string   `arg:"" help:"Chat type: oneOnOne or group" enum:"oneOnOne,group"`
	Members []string `arg:"" help:"Member UPN emails (you are auto-added)"`
	Topic   string   `help:"Chat topic (group chats only)" optional:""`
}

func (c *ChatsCreateCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"chatType":     c.Type,
		"members_upns": c.Members,
	}
	if c.Topic != "" {
		args["topic"] = c.Topic
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(teamsEndpoint(), "CreateChat",
			fmt.Sprintf("create %s chat with %v", c.Type, c.Members),
			map[string]any{
				"action":       "chats.create",
				"chatType":     c.Type,
				"members_upns": c.Members,
				"topic":        c.Topic,
			},
			args,
		)
	}

	client := ctx.NewMCPClient(teamsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "CreateChat", args)
	if err != nil {
		return fmt.Errorf("create chat: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Chat created", data)
}

// ChatsDeleteCmd deletes a chat.
type ChatsDeleteCmd struct {
	ChatID string `arg:"" help:"Chat ID"`
}

func (c *ChatsDeleteCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"chatId": c.ChatID,
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(teamsEndpoint(), "DeleteChat",
			fmt.Sprintf("delete chat %s", c.ChatID),
			map[string]any{"action": "chats.delete", "chatId": c.ChatID},
			args,
		)
	}

	if err := ctx.Confirm(fmt.Sprintf("delete chat %s", c.ChatID)); err != nil {
		return err
	}

	client := ctx.NewMCPClient(teamsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "DeleteChat", args)
	if err != nil {
		return fmt.Errorf("delete chat: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Chat deleted", data)
}

// ChatsUpdateCmd updates a group chat's topic.
type ChatsUpdateCmd struct {
	ChatID string `arg:"" help:"Chat ID"`
	Topic  string `arg:"" help:"New topic/display name"`
}

func (c *ChatsUpdateCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"chatId": c.ChatID,
		"topic":  c.Topic,
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(teamsEndpoint(), "UpdateChat",
			fmt.Sprintf("update chat %s topic to %q", c.ChatID, c.Topic),
			map[string]any{"action": "chats.update", "chatId": c.ChatID, "topic": c.Topic},
			args,
		)
	}

	client := ctx.NewMCPClient(teamsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "UpdateChat", args)
	if err != nil {
		return fmt.Errorf("update chat: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Chat updated", data)
}

// --- Message mutation commands ---

// ChatsUpdateMessageCmd updates a chat message.
type ChatsUpdateMessageCmd struct {
	ChatID    string `arg:"" help:"Chat ID"`
	MessageID string `arg:"" help:"Message ID"`
	Content   string `arg:"" help:"New message content"`
}

func (c *ChatsUpdateMessageCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"chatId":    c.ChatID,
		"messageId": c.MessageID,
		"content":   c.Content,
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(teamsEndpoint(), "UpdateChatMessage",
			fmt.Sprintf("update message %s in chat %s", c.MessageID, c.ChatID),
			map[string]any{
				"action":    "chats.update-message",
				"chatId":    c.ChatID,
				"messageId": c.MessageID,
				"content":   c.Content,
			},
			args,
		)
	}

	client := ctx.NewMCPClient(teamsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "UpdateChatMessage", args)
	if err != nil {
		return fmt.Errorf("update message: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Message updated", data)
}

// ChatsDeleteMessageCmd deletes a chat message.
type ChatsDeleteMessageCmd struct {
	ChatID    string `arg:"" help:"Chat ID"`
	MessageID string `arg:"" help:"Message ID"`
}

func (c *ChatsDeleteMessageCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"chatId":    c.ChatID,
		"messageId": c.MessageID,
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(teamsEndpoint(), "DeleteChatMessage",
			fmt.Sprintf("delete message %s in chat %s", c.MessageID, c.ChatID),
			map[string]any{"action": "chats.delete-message", "chatId": c.ChatID, "messageId": c.MessageID},
			args,
		)
	}

	if err := ctx.Confirm(fmt.Sprintf("delete message %s in chat %s", c.MessageID, c.ChatID)); err != nil {
		return err
	}

	client := ctx.NewMCPClient(teamsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "DeleteChatMessage", args)
	if err != nil {
		return fmt.Errorf("delete message: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Message deleted", data)
}

// --- Chat member commands ---

// ChatsListMembersCmd lists members of a chat.
type ChatsListMembersCmd struct {
	ChatID string `arg:"" help:"Chat ID"`
}

func (c *ChatsListMembersCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(teamsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "ListChatMembers", map[string]any{
		"chatId": c.ChatID,
	})
	if err != nil {
		return fmt.Errorf("list chat members: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	rows := output.ToRows(data, "members")
	if rows == nil {
		rows = output.ToRows(data, "value")
	}
	return ctx.Output.PrintList("members", output.MembersColumns, rows)
}

// ChatsAddMemberCmd adds a member to a chat.
type ChatsAddMemberCmd struct {
	ChatID string   `arg:"" help:"Chat ID"`
	UPN    string   `arg:"" help:"User UPN (email) to add"`
	Roles  []string `help:"Roles for the member (owner or guest)" default:"owner"`
}

func (c *ChatsAddMemberCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"chatId":         c.ChatID,
		"roles":          c.Roles,
		"userodata_bind": fmt.Sprintf("https://graph.microsoft.com/v1.0/users('%s')", c.UPN),
		"odata_type":     "#microsoft.graph.aadUserConversationMember",
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(teamsEndpoint(), "AddChatMember",
			fmt.Sprintf("add member %s to chat %s", c.UPN, c.ChatID),
			map[string]any{"action": "chats.add-member", "chatId": c.ChatID, "upn": c.UPN, "roles": c.Roles},
			args,
		)
	}

	client := ctx.NewMCPClient(teamsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, "AddChatMember", args)
	if err != nil {
		return fmt.Errorf("add chat member: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation(fmt.Sprintf("Member %s added to chat", c.UPN), data)
}
