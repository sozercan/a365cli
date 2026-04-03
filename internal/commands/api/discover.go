package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/output"
	"github.com/sozercan/a365cli/internal/version"
)

// APIDiscoverCmd discovers available MCP servers from the agent365 gateway.
type APIDiscoverCmd struct{}

func (c *APIDiscoverCmd) Run(ctx *commands.Context) error {
	// The discover endpoint is at the agents root, not under /servers/
	baseURL := config.BaseURL()
	// Strip trailing "servers/" to get the agents root
	agentsRoot := strings.TrimSuffix(baseURL, "servers/")
	discoverURL := agentsRoot + "discoverToolServers"

	// Get bearer token
	token, err := ctx.TokenProvider(ctx.Ctx)
	if err != nil {
		return fmt.Errorf("get token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx.Ctx, http.MethodGet, discoverURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("a365/%s (Go)", version.Version))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("discover: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Try to parse as array of servers
	var servers []map[string]any
	if err := json.Unmarshal(body, &servers); err != nil {
		// Try as object with a key
		var wrapper map[string]any
		if err2 := json.Unmarshal(body, &wrapper); err2 != nil {
			return fmt.Errorf("parse response: %w", err)
		}
		// Look for array in common keys
		for _, key := range []string{"mcpServers", "servers", "value"} {
			if arr, ok := wrapper[key]; ok {
				if typed, ok := arr.([]any); ok {
					for _, item := range typed {
						if m, ok := item.(map[string]any); ok {
							servers = append(servers, m)
						}
					}
					break
				}
			}
		}
		if servers == nil {
			return ctx.Output.PrintItem(wrapper)
		}
	}

	rows := make([]map[string]any, 0, len(servers))
	for _, s := range servers {
		row := map[string]any{
			"name":     getString(s, "mcpServerName", "name", "mcpServerUniqueName"),
			"url":      getString(s, "url", "endpoint"),
			"scope":    getString(s, "scope"),
			"audience": getString(s, "audience"),
		}
		rows = append(rows, row)
	}

	return ctx.Output.PrintList("servers", output.APIDiscoverColumns, rows)
}

// getString returns the first non-empty value from the map for the given keys.
func getString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}
