package teams

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/output"
)

// SearchCmd searches Teams messages.
type SearchCmd struct {
	Query string `arg:"" help:"Search query (KQL syntax: keywords, from:user, sent>=date)"`
	Size  int    `help:"Number of results (max 25)" default:"25" name:"max"`
}

func (c *SearchCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(teamsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	args := map[string]any{
		"queryString": c.Query,
	}
	if c.Size > 0 && c.Size != 25 {
		args["size"] = c.Size
	}

	resp, err := client.CallTool(ctx.Ctx, "SearchTeamMessagesQueryParameters", args)
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}

	// Search results are nested: results[0].hitsContainers[0].hits[]
	// Try to unwrap the nested structure
	rows := extractSearchHits(data)
	if rows == nil {
		// Fallback: print as item
		return ctx.Output.PrintItem(data)
	}

	return ctx.Output.PrintList("results", output.SearchColumns, rows)
}

// extractSearchHits unwraps the Graph Search API nested response structure.
// Shape: {results: [{hitsContainers: [{hits: [{resource: {...}, summary: ...}]}]}]}
func extractSearchHits(data map[string]any) []map[string]any {
	// Try flat structures first
	if rows := output.ToRows(data, "hits"); rows != nil {
		return rows
	}

	// Unwrap nested Graph Search response
	results := output.ToRows(data, "results")
	if results == nil {
		return nil
	}

	for _, result := range results {
		containers := output.ToRows(result, "hitsContainers")
		for _, container := range containers {
			hits := output.ToRows(container, "hits")
			if len(hits) > 0 {
				// Flatten: pull resource fields up and keep summary
				rows := make([]map[string]any, 0, len(hits))
				for _, hit := range hits {
					row := map[string]any{}
					// Copy summary
					if s, ok := hit["summary"]; ok {
						row["summary"] = s
					}
					// Merge resource fields into row
					if res, ok := hit["resource"].(map[string]any); ok {
						for k, v := range res {
							row[k] = v
						}
					}
					rows = append(rows, row)
				}
				return rows
			}
		}
	}

	return nil
}

// SearchNLCmd searches Teams messages using natural language.
type SearchNLCmd struct {
	Query          string `arg:"" help:"Natural language query (e.g. 'my chats with John about budget')"`
	ConversationID string `help:"Existing conversation ID for follow-up queries" optional:""`
}

func (c *SearchNLCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(teamsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	args := map[string]any{
		"message": c.Query,
	}
	if c.ConversationID != "" {
		args["conversationId"] = c.ConversationID
	}

	resp, err := client.CallTool(ctx.Ctx, "SearchTeamsMessages", args)
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}

	// Natural language search returns a Copilot reply + chatIds
	return ctx.Output.PrintItem(data)
}
