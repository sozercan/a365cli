package knowledge

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
)

// KnowledgeCmd groups federated knowledge subcommands.
type KnowledgeCmd struct {
	Query     KnowledgeQueryCmd     `cmd:"" help:"Query federated knowledge"`
	List      KnowledgeListCmd      `cmd:"" help:"List federated knowledge configurations"`
	Configure KnowledgeConfigureCmd `cmd:"" help:"Configure a federated knowledge source"`
	Ingest    KnowledgeIngestCmd    `cmd:"" help:"Trigger ingestion for a knowledge source"`
	Delete    KnowledgeDeleteCmd    `cmd:"" help:"Delete a federated knowledge configuration"`
}

func knowledgeEndpoint() string {
	return config.Endpoint("knowledge")
}

// KnowledgeQueryCmd queries federated knowledge.
type KnowledgeQueryCmd struct {
	ConsumerID string `arg:"" help:"Consumer ID"`
	Query      string `arg:"" help:"Search query"`
}

func (c *KnowledgeQueryCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(knowledgeEndpoint(), "query_federated_knowledge", "query", map[string]any{
		"consumerId": c.ConsumerID,
		"query":      c.Query,
	})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// KnowledgeListCmd lists federated knowledge configs.
type KnowledgeListCmd struct {
	ConsumerID string `arg:"" help:"Consumer ID"`
}

func (c *KnowledgeListCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(knowledgeEndpoint(), "retrieve_federated_knowledge", "list", map[string]any{
		"consumerId": c.ConsumerID,
	})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

// KnowledgeConfigureCmd configures a federated knowledge source.
type KnowledgeConfigureCmd struct {
	ConsumerID  string `arg:"" help:"Consumer ID"`
	SourceType  string `arg:"" help:"Source type"`
	DisplayName string `arg:"" help:"Display name"`
	Description string `arg:"" help:"Description"`
}

func (c *KnowledgeConfigureCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"consumerId":      c.ConsumerID,
		"knowledgeConfig": "{}",
		"sourceType":      c.SourceType,
		"displayName":     c.DisplayName,
		"description":     c.Description,
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(knowledgeEndpoint(), "configure_federated_knowledge",
			"configure knowledge source",
			map[string]any{
				"action":           "knowledge.configure",
				"consumerId":       c.ConsumerID,
				"sourceType":       c.SourceType,
				"display_name_len": len(c.DisplayName),
				"description_len":  len(c.Description),
			},
			args,
		)
	}

	data, err := ctx.CallToolData(knowledgeEndpoint(), "configure_federated_knowledge", "configure", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Knowledge source configured", data)
}

// KnowledgeIngestCmd triggers ingestion.
type KnowledgeIngestCmd struct {
	ConsumerID string `arg:"" help:"Consumer ID"`
	ConfigID   string `arg:"" help:"Search configuration ID"`
}

func (c *KnowledgeIngestCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"consumerId":            c.ConsumerID,
		"searchConfigurationId": c.ConfigID,
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(knowledgeEndpoint(), "ingest_federated_knowledge",
			fmt.Sprintf("ingest knowledge config %s", c.ConfigID),
			map[string]any{
				"action":                "knowledge.ingest",
				"consumerId":            c.ConsumerID,
				"searchConfigurationId": c.ConfigID,
			},
			args,
		)
	}

	data, err := ctx.CallToolData(knowledgeEndpoint(), "ingest_federated_knowledge", "ingest", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Ingestion triggered", data)
}

// KnowledgeDeleteCmd deletes a knowledge config.
type KnowledgeDeleteCmd struct {
	ConsumerID string `arg:"" help:"Consumer ID"`
	ConfigID   string `arg:"" help:"Search configuration ID"`
}

func (c *KnowledgeDeleteCmd) Run(ctx *commands.Context) error {
	args := map[string]any{
		"consumerId":            c.ConsumerID,
		"searchConfigurationId": c.ConfigID,
	}

	if ctx.DryRun {
		return ctx.ValidateDryRun(knowledgeEndpoint(), "delete_federated_knowledge",
			fmt.Sprintf("delete knowledge config %s", c.ConfigID),
			map[string]any{
				"action":                "knowledge.delete",
				"consumerId":            c.ConsumerID,
				"searchConfigurationId": c.ConfigID,
			},
			args,
		)
	}

	if err := ctx.Confirm(fmt.Sprintf("delete knowledge config %s", c.ConfigID)); err != nil {
		return err
	}

	data, err := ctx.CallToolData(knowledgeEndpoint(), "delete_federated_knowledge", "delete", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Knowledge config deleted", data)
}
