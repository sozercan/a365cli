package mcp

import "strings"

// toolErrorPrefix is how agent365 MCP servers report a failed tool execution
// in plain-text content, e.g.
// "Error: Error executing tool: Unknown argument: 'x' for tool 'Y'.".
const toolErrorPrefix = "error executing tool"

// ToolError reports a tools/call that reached the server but failed while
// executing the tool. The server signals it either with MCP's isError flag or
// with an "Error executing tool: ..." text payload instead of a JSON-RPC error.
type ToolError struct {
	Tool    string
	Message string
}

func (e *ToolError) Error() string {
	return e.Message
}

// ToolCallError returns a *ToolError when resp describes a failed tool call,
// or nil when the call succeeded (or resp carries no result to inspect).
// JSON-RPC level errors are left to the caller's normal error handling.
func ToolCallError(tool string, resp *JSONRPCResponse) *ToolError {
	if resp == nil || resp.Result == nil {
		return nil
	}

	var texts []string
	for _, c := range resp.Result.Content {
		if c.Type == "text" && strings.TrimSpace(c.Text) != "" {
			texts = append(texts, c.Text)
		}
	}
	message := SanitizeToolErrorMessage(strings.Join(texts, "\n"))

	if !resp.Result.IsError && !IsToolErrorMessage(message) {
		return nil
	}
	if message == "" {
		message = "tool " + tool + " reported an error"
	}
	return &ToolError{Tool: tool, Message: message}
}

// IsToolErrorMessage reports whether a (sanitized) message is an
// "Error executing tool: ..." failure payload.
func IsToolErrorMessage(message string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(message)), toolErrorPrefix)
}

// SanitizeToolErrorMessage normalizes a tool error payload: it strips the
// redundant leading "Error:" prefix, normalizes line endings, and drops blank
// and duplicate lines (servers often repeat the CorrelationId line).
func SanitizeToolErrorMessage(message string) string {
	message = strings.TrimSpace(message)
	message = strings.ReplaceAll(message, "\r\n", "\n")
	message = strings.ReplaceAll(message, "\r", "\n")

	seen := map[string]struct{}{}
	lines := strings.Split(message, "\n")
	cleaned := make([]string, 0, len(lines))

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if i == 0 {
			line = strings.TrimSpace(strings.TrimPrefix(line, "Error:"))
		}
		if line == "" {
			continue
		}
		if _, ok := seen[line]; ok {
			continue
		}
		seen[line] = struct{}{}
		cleaned = append(cleaned, line)
	}

	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}
