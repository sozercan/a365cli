package copilot

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/output"
)

// CopilotCmd groups all Copilot subcommands.
type CopilotCmd struct {
	Chat CopilotChatCmd `cmd:"" help:"Ask Copilot about your M365 content"`
}

func copilotEndpoint() string {
	return config.Endpoint("copilot")
}

// CopilotChatCmd searches internal M365 content using natural language.
type CopilotChatCmd struct {
	Message        string `arg:"" help:"Natural language question about your M365 content"`
	ConversationID string `help:"Conversation ID for follow-up queries" name:"conversation-id" optional:""`
}

func (c *CopilotChatCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(copilotEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	args := map[string]any{
		"message": c.Message,
	}
	if c.ConversationID != "" {
		args["conversationId"] = c.ConversationID
	}

	resp, err := client.CallTool(ctx.Ctx, "copilot_chat", args)
	if err != nil {
		return fmt.Errorf("copilot chat: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}
