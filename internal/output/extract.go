package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sozercan/a365cli/internal/mcp"
)

// ExtractContent unwraps the MCP JSON-RPC response to get the domain data.
// It parses Content[].Text JSON strings into a map[string]any.
func ExtractContent(resp *mcp.JSONRPCResponse) (map[string]any, error) {
	if resp == nil {
		return nil, fmt.Errorf("nil response")
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("MCP error %d: %s", resp.Error.Code, resp.Error.Message)
	}
	if resp.Result == nil || len(resp.Result.Content) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	// Find the first text content block that contains JSON
	for _, c := range resp.Result.Content {
		if c.Type != "text" {
			continue
		}
		text := strings.TrimSpace(c.Text)
		if text == "" {
			continue
		}

		// Try to parse as JSON object
		if strings.HasPrefix(text, "{") {
			var data map[string]any
			if err := json.Unmarshal([]byte(text), &data); err != nil {
				continue
			}
			return unwrapRawResponse(data), nil
		}

		// Try to parse as JSON array — wrap in a generic key
		if strings.HasPrefix(text, "[") {
			var arr []any
			if err := json.Unmarshal([]byte(text), &arr); err != nil {
				continue
			}
			return map[string]any{"items": arr}, nil
		}

		// Some servers embed JSON after a status message, e.g.
		// "Events retrieved successfully.\n{\"value\":[...]}"
		if idx := strings.Index(text, "\n{"); idx >= 0 {
			jsonPart := strings.TrimSpace(text[idx+1:])
			var data map[string]any
			if err := json.Unmarshal([]byte(jsonPart), &data); err == nil {
				return unwrapRawResponse(data), nil
			}
		}
		if idx := strings.Index(text, "\n["); idx >= 0 {
			jsonPart := strings.TrimSpace(text[idx+1:])
			var arr []any
			if err := json.Unmarshal([]byte(jsonPart), &arr); err == nil {
				return map[string]any{"items": arr}, nil
			}
		}
	}

	// If no JSON found, return the raw text
	var texts []string
	for _, c := range resp.Result.Content {
		if c.Type == "text" && strings.TrimSpace(c.Text) != "" {
			texts = append(texts, c.Text)
		}
	}
	if len(texts) > 0 {
		return map[string]any{"message": strings.Join(texts, "\n")}, nil
	}

	return nil, fmt.Errorf("no content in response")
}

// unwrapRawResponse handles servers that return {"rawResponse": "<json-string>", "message": "..."}
// by parsing the rawResponse string and merging or replacing the outer data.
func unwrapRawResponse(data map[string]any) map[string]any {
	raw, ok := data["rawResponse"]
	if !ok {
		return data
	}
	rawStr, ok := raw.(string)
	if !ok {
		return data
	}
	var inner map[string]any
	if err := json.Unmarshal([]byte(rawStr), &inner); err != nil {
		return data
	}
	// Return the inner parsed JSON (the actual Graph response)
	return inner
}

// ToRows extracts a named array from domain data as a slice of maps.
// e.g., ToRows(data, "teams") extracts data["teams"] as []map[string]any.
func ToRows(data map[string]any, key string) []map[string]any {
	arr, ok := data[key]
	if !ok {
		return nil
	}

	slice, ok := arr.([]any)
	if !ok {
		return nil
	}

	rows := make([]map[string]any, 0, len(slice))
	for _, item := range slice {
		if m, ok := item.(map[string]any); ok {
			rows = append(rows, m)
		}
	}
	return rows
}
