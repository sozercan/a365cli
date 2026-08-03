package teams

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/output"
)

// ChannelsCmd groups channel subcommands.
type ChannelsCmd struct {
	List          ChannelsListCmd          `cmd:"" help:"List channels in a team"`
	Get           ChannelsGetCmd           `cmd:"" help:"Get a channel by ID"`
	Create        ChannelsCreateCmd        `cmd:"" help:"Create a standard channel"`
	CreatePrivate ChannelsCreatePrivateCmd `cmd:"" name:"create-private" help:"Create a private channel"`
	Update        ChannelsUpdateCmd        `cmd:"" help:"Update a channel"`
	Messages      ChannelsMessagesCmd      `cmd:"" help:"List messages in a channel"`
	Post          ChannelsPostCmd          `cmd:"" help:"Post a message to a channel"`
	Reply         ChannelsReplyCmd         `cmd:"" help:"Reply to a channel message"`
	Members       ChannelsListMembersCmd   `cmd:"" help:"List members of a channel"`
	AddMember     ChannelsAddMemberCmd     `cmd:"" name:"add-member" help:"Add a member to a channel"`
	UpdateMember  ChannelsUpdateMemberCmd  `cmd:"" name:"update-member" help:"Update a channel member's role"`
}

// ChannelsListCmd lists channels in a team.
type ChannelsListCmd struct {
	TeamID string `arg:"" help:"Team ID"`
	Max    int    `help:"Maximum number of results" default:"100"`
}

func (c *ChannelsListCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(teamsEndpoint(), "ListChannels", "list channels", map[string]any{
		"teamId": c.TeamID,
	})
	if err != nil {
		return err
	}
	return ctx.Output.PrintListFromData("channels", output.ChannelsColumns, data, c.Max)
}

// ChannelsGetCmd gets a specific channel.
type ChannelsGetCmd struct {
	TeamID    string `arg:"" help:"Team ID"`
	ChannelID string `arg:"" help:"Channel ID"`
}

func (c *ChannelsGetCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(teamsEndpoint(), "GetChannel", "get channel", map[string]any{
		"teamId":    c.TeamID,
		"channelId": c.ChannelID,
	})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// ChannelsMessagesCmd lists messages in a channel.
type ChannelsMessagesCmd struct {
	TeamID    string `arg:"" help:"Team ID"`
	ChannelID string `arg:"" help:"Channel ID"`
	Max       int    `help:"Maximum number of messages" default:"20"`
}

func (c *ChannelsMessagesCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"teamId":    c.TeamID,
		"channelId": c.ChannelID,
	}
	if c.Max > 0 && c.Max != 20 {
		args["top"] = c.Max
	}

	data, err := ctx.CallToolData(teamsEndpoint(), "ListChannelMessages", "list channel messages", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintListFromData("messages", output.MessagesColumns, data, c.Max)
}

// ChannelsPostCmd posts a message to a channel.
type ChannelsPostCmd struct {
	TeamID    string `arg:"" help:"Team ID"`
	ChannelID string `arg:"" help:"Channel ID"`
	Message   string `arg:"" help:"Message text"`
}

func (c *ChannelsPostCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"teamId":    c.TeamID,
		"channelId": c.ChannelID,
		"content":   c.Message,
	}
	if ctx.DryRun {
		return ctx.ValidateDryRun(teamsEndpoint(), "SendMessageToChannel",
			fmt.Sprintf("post message to channel %s in team %s", c.ChannelID, c.TeamID),
			map[string]any{
				"action":        "channels.post",
				"teamId":        c.TeamID,
				"channelId":     c.ChannelID,
				"contentLength": len(c.Message),
			},
			args,
		)
	}

	data, err := ctx.CallToolData(teamsEndpoint(), "SendMessageToChannel", "post channel message", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Message posted to channel", data)
}

// ChannelsReplyCmd replies to a channel message.
type ChannelsReplyCmd struct {
	TeamID    string `arg:"" help:"Team ID"`
	ChannelID string `arg:"" help:"Channel ID"`
	MessageID string `arg:"" help:"Message ID to reply to"`
	Message   string `arg:"" help:"Reply text"`
}

func (c *ChannelsReplyCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"teamId":    c.TeamID,
		"channelId": c.ChannelID,
		"messageId": c.MessageID,
		"content":   c.Message,
	}
	if ctx.DryRun {
		return ctx.ValidateDryRun(teamsEndpoint(), "ReplyToChannelMessage",
			fmt.Sprintf("reply to message %s in channel %s", c.MessageID, c.ChannelID),
			map[string]any{
				"action":        "channels.reply",
				"teamId":        c.TeamID,
				"channelId":     c.ChannelID,
				"messageId":     c.MessageID,
				"contentLength": len(c.Message),
			},
			args,
		)
	}

	data, err := ctx.CallToolData(teamsEndpoint(), "ReplyToChannelMessage", "reply to channel message", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Reply posted", data)
}

// --- Channel management commands ---

// ChannelsCreateCmd creates a standard channel.
type ChannelsCreateCmd struct {
	TeamID      string `arg:"" help:"Team ID"`
	DisplayName string `arg:"" help:"Channel name"`
	Description string `help:"Channel description" optional:""`
}

func (c *ChannelsCreateCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"teamId":      c.TeamID,
		"displayName": c.DisplayName,
	}
	if c.Description != "" {
		args["description"] = c.Description
	}
	if ctx.DryRun {
		return ctx.ValidateDryRun(teamsEndpoint(), "CreateChannel",
			fmt.Sprintf("create channel %q in team %s", c.DisplayName, c.TeamID),
			map[string]any{
				"action":            "channels.create",
				"teamId":            c.TeamID,
				"displayName":       c.DisplayName,
				"descriptionLength": len(c.Description),
			},
			args,
		)
	}

	data, err := ctx.CallToolData(teamsEndpoint(), "CreateChannel", "create channel", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Channel created", data)
}

// ChannelsCreatePrivateCmd creates a private channel.
type ChannelsCreatePrivateCmd struct {
	TeamID      string `arg:"" help:"Team ID"`
	DisplayName string `arg:"" help:"Channel name"`
	Description string `help:"Channel description" optional:""`
}

func (c *ChannelsCreatePrivateCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"teamId":         c.TeamID,
		"displayName":    c.DisplayName,
		"membershipType": "private",
	}
	if c.Description != "" {
		args["description"] = c.Description
	}
	if ctx.DryRun {
		return ctx.ValidateDryRun(teamsEndpoint(), "CreateChannel",
			fmt.Sprintf("create private channel %q in team %s", c.DisplayName, c.TeamID),
			map[string]any{
				"action":            "channels.create-private",
				"teamId":            c.TeamID,
				"displayName":       c.DisplayName,
				"membershipType":    "private",
				"descriptionLength": len(c.Description),
			},
			args,
		)
	}

	data, err := ctx.CallToolData(teamsEndpoint(), "CreateChannel", "create private channel", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Private channel created", data)
}

// ChannelsUpdateCmd updates a channel's name or description.
type ChannelsUpdateCmd struct {
	TeamID      string `arg:"" help:"Team ID"`
	ChannelID   string `arg:"" help:"Channel ID"`
	DisplayName string `help:"New channel name" optional:""`
	Description string `help:"New channel description" optional:""`
}

func (c *ChannelsUpdateCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"teamId":    c.TeamID,
		"channelId": c.ChannelID,
	}
	if c.DisplayName != "" {
		args["displayName"] = c.DisplayName
	}
	if c.Description != "" {
		args["description"] = c.Description
	}
	if ctx.DryRun {
		return ctx.ValidateDryRun(teamsEndpoint(), "UpdateChannel",
			fmt.Sprintf("update channel %s in team %s", c.ChannelID, c.TeamID),
			map[string]any{
				"action":            "channels.update",
				"teamId":            c.TeamID,
				"channelId":         c.ChannelID,
				"displayName":       c.DisplayName,
				"descriptionLength": len(c.Description),
			},
			args,
		)
	}

	data, err := ctx.CallToolData(teamsEndpoint(), "UpdateChannel", "update channel", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Channel updated", data)
}

// --- Channel member commands ---

// ChannelsListMembersCmd lists members of a channel.
type ChannelsListMembersCmd struct {
	TeamID    string `arg:"" help:"Team ID"`
	ChannelID string `arg:"" help:"Channel ID"`
	Max       int    `help:"Maximum number of members" default:"100"`
}

func (c *ChannelsListMembersCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"teamId":    c.TeamID,
		"channelId": c.ChannelID,
	}
	if c.Max > 0 && c.Max != 100 {
		args["top"] = c.Max
	}

	data, err := ctx.CallToolData(teamsEndpoint(), "ListChannelMembers", "list channel members", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintListFromData("members", output.MembersColumns, data, c.Max, "members", "value")
}

// ChannelsAddMemberCmd adds a member to a channel.
type ChannelsAddMemberCmd struct {
	TeamID    string `arg:"" help:"Team ID"`
	ChannelID string `arg:"" help:"Channel ID"`
	UserID    string `arg:"" help:"User ID (GUID) to add"`
}

func (c *ChannelsAddMemberCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"teamId":    c.TeamID,
		"channelId": c.ChannelID,
		"userId":    c.UserID,
	}
	if ctx.DryRun {
		return ctx.ValidateDryRun(teamsEndpoint(), "AddChannelMember",
			fmt.Sprintf("add member to channel %s", c.ChannelID),
			map[string]any{
				"action":                 "channels.add-member",
				"teamId":                 c.TeamID,
				"channelId":              c.ChannelID,
				"memberIdentifierLength": len(c.UserID),
			},
			args,
		)
	}

	data, err := ctx.CallToolData(teamsEndpoint(), "AddChannelMember", "add channel member", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Member added to channel", data)
}

// ChannelsUpdateMemberCmd updates a channel member's role.
type ChannelsUpdateMemberCmd struct {
	TeamID       string `arg:"" help:"Team ID"`
	ChannelID    string `arg:"" help:"Channel ID"`
	MembershipID string `arg:"" help:"Membership ID (from members list)"`
	Role         string `arg:"" help:"New role: owner or member" enum:"owner,member"`
}

func (c *ChannelsUpdateMemberCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"teamId":       c.TeamID,
		"channelId":    c.ChannelID,
		"membershipId": c.MembershipID,
		"role":         c.Role,
	}
	if ctx.DryRun {
		return ctx.ValidateDryRun(teamsEndpoint(), "UpdateChannelMember",
			fmt.Sprintf("update member %s role to %s in channel %s", c.MembershipID, c.Role, c.ChannelID),
			map[string]any{
				"action":       "channels.update-member",
				"teamId":       c.TeamID,
				"channelId":    c.ChannelID,
				"membershipId": c.MembershipID,
				"role":         c.Role,
			},
			args,
		)
	}

	data, err := ctx.CallToolData(teamsEndpoint(), "UpdateChannelMember", "update channel member", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Channel member role updated", data)
}
