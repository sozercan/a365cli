# Copilot

Ask natural language questions about all your Microsoft 365 content, including documents, emails, chats, and files. Copilot uses M365 intelligence to find and summarize information across your tenant.

## Commands

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `copilot chat` | Ask Copilot about your M365 content, or start an interactive prompt | `[message]`, `--conversation-id` |

## Arguments

- **`[message]`** (optional) -- Natural language question about your M365 content. If omitted and stdin is interactive, `a365` starts a Copilot prompt.
- **`--conversation-id`** (optional) -- Conversation ID for follow-up queries. Use `--output json` when you need to inspect returned conversation IDs.

## Examples

```sh
# Ask a simple question
a365 copilot chat "What were the key decisions from last week's project standup?"

# Summarize recent emails from a colleague
a365 copilot chat "Summarize recent emails from Alice about the Q3 budget"

# Search across documents
a365 copilot chat "Find the latest sales forecast spreadsheet"

# Start an interactive Copilot session
a365 copilot chat

# Follow up on a previous conversation
a365 copilot chat "Can you give more detail on the second point?" \
  --conversation-id 19:abc123def456

# Output as JSON
a365 copilot chat "Who shared files with me this week?" --output json
```


## Timeout tuning

Copilot requests can take longer than typical MCP tool calls. By default, `a365` waits up to `5m` for Copilot response headers before failing. Override this with `A365_COPILOT_RESPONSE_HEADER_TIMEOUT`, or use `A365_MCP_RESPONSE_HEADER_TIMEOUT` to change the timeout for every MCP service.

```sh
A365_COPILOT_RESPONSE_HEADER_TIMEOUT=10m a365 copilot chat "Summarize recent project updates"
```
