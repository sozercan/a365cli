# Copilot

Ask natural language questions about all your Microsoft 365 content, including documents, emails, chats, and files. Copilot uses M365 intelligence to find and summarize information across your tenant.

You can also inspect available Copilot agents and target a specific agent when chatting. Web search is enabled by default for `copilot chat`; use `--no-web-search` to send `enableWebSearch=false` to Copilot.

## Commands

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `copilot chat` | Ask Copilot about your M365 content, or start an interactive prompt | `[message]`, `--conversation-id`, `--agent`, `--no-web-search` |
| `copilot agents` | List available Copilot agents and their chat selectors | _(none)_ |

## Arguments

- **`[message]`** (optional) -- Natural language question about your M365 content. If omitted and stdin is interactive, `a365` starts a Copilot prompt.
- **`--conversation-id`** (optional) -- Conversation ID for follow-up queries. Use `--output json` when you need to inspect returned conversation IDs.
- **`--agent`** (optional) -- Copilot agent name, selector, or title ID. Resolve available selectors with `a365 copilot agents`.
- **`--no-web-search`** (optional) -- Request Copilot with web search disabled. By default, `copilot chat` sends `enableWebSearch=true`; with `--no-web-search`, `a365` sends `enableWebSearch=false`.

## Behavior note

`--no-web-search` controls the request sent by `a365`, not a strict guarantee about Copilot's final grounding behavior. In live testing, Copilot could still return public/external grounding or citations for some prompts even when `enableWebSearch=false`.

## Examples

```sh
# Ask a simple question
a365 copilot chat "What were the key decisions from last week's project standup?"

# List available Copilot agents and selectors
a365 copilot agents

# Target a specific Copilot agent
a365 copilot chat --agent "Researcher" "Summarize the latest customer escalations"

# Summarize recent emails from a colleague
a365 copilot chat "Summarize recent emails from Alice about the Q3 budget"

# Search across documents
a365 copilot chat "Find the latest sales forecast spreadsheet"

# Ask Copilot with web search disabled
a365 copilot chat --no-web-search "Summarize recent emails from Alice about the Q3 budget"

# Start an interactive Copilot session
a365 copilot chat

# Follow up on a previous conversation
a365 copilot chat "Can you give more detail on the second point?" \
  --conversation-id 19:abc123def456

# Output as JSON
a365 copilot chat "Who shared files with me this week?" --output json
```
