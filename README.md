# a365

A standalone, agent-friendly CLI for Microsoft 365 via [agent365](https://agent365.svc.cloud.microsoft) MCP servers.

## Features

- **18 M365 services** — Teams, Mail, Calendar, Planner, SharePoint, OneDrive, Word, Excel, Copilot, Admin, Triggers, WebSearch, and more
- **170+ MCP tools** — full API coverage with dynamic server discovery
- **Agent-friendly** — structured `--output=json` for LLM tool use, `--no-input` for non-interactive execution, `--dry-run` for safe exploration
- **Three output modes** — human tables (default), JSON for scripting, TSV for piping
- **Interactive browser auth** with PKCE — silent re-auth on subsequent runs
- **Resilient** — automatic retries with backoff on 502/503/429, MCP session caching
- **Configurable** — `~/.a365/config.json` for persistent defaults, env vars, CLI flags
- **Shell completion** — bash, zsh, fish
- **API explorer** — discover and call any MCP tool directly with `a365 api`

## Quick Start

```bash
# Install
brew tap sozercan/repo && brew install a365

# Authenticate (opens browser once, tokens are cached)
a365 auth login

# Use it
a365 teams list                    # List your Teams
a365 mail search '?$top=5'        # Recent emails
a365 cal list                     # Upcoming meetings
a365 copilot chat "Summarize my week"   # Ask Copilot
a365 copilot chat                       # Interactive Copilot prompt
a365 me whoami                    # Your profile
a365 odr ls                      # OneDrive files
a365 websearch search "MCP protocol" https://modelcontextprotocol.io  # Web search
```

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap sozercan/repo
brew install a365
```

### GitHub Releases

Pre-built binaries for Linux, macOS, and Windows are available on the [Releases](https://github.com/sozercan/a365cli/releases) page.

### go install

```bash
go install github.com/sozercan/a365cli@latest
```

## Authentication

a365 uses Entra ID interactive browser authentication with PKCE. On first run it opens your browser; after that, tokens refresh silently.

A built-in client ID is provided by default. If your tenant requires a custom app registration, set your own via `--client-id` or `A365_CLIENT_ID`.

```bash
# Login (uses the built-in client ID by default)
a365 auth login

# Check status
a365 auth status

# View token details (scopes, expiry)
a365 auth token

# Logout
a365 auth logout
```

| Variable | Flag | Description |
|----------|------|-------------|
| `A365_CLIENT_ID` | `--client-id` | Entra app client ID (default: `aebc6443-996d-45c2-90f0-388ff96faa56`) |
| `A365_TENANT_ID` | `--tenant-id` | Entra tenant ID (optional, defaults to `organizations`) |
| `A365_ENDPOINT` | — | Override the agent365 base URL |
| `A365_MCP_RESPONSE_HEADER_TIMEOUT` | — | Override the MCP HTTP response-header timeout (for example `180s`, `5m`) |
| `A365_COPILOT_RESPONSE_HEADER_TIMEOUT` | — | Override the Copilot MCP response-header timeout (default: `5m`) |

## Configuration

Persist defaults in `~/.a365/config.json`:

```bash
a365 config set client-id your-client-id-here  # override the default client ID
a365 config set output json        # or table, tsv
a365 config show                   # view all settings
a365 config path                   # ~/.a365/config.json
```

CLI flags and env vars always take precedence over config file values.

## Services

| Service | Alias | Cmds | Documentation |
|---------|-------|------|---------------|
| [Teams](docs/teams.md) | — | 28 | Channels, chats, messages, members, search |
| [Mail](docs/mail.md) | `email` | 21 | Search, send, reply, forward, drafts, attachments, threading |
| [Calendar](docs/calendar.md) | `cal` | 13 | Events, RSVP, scheduling, rooms |
| [Planner](docs/planner.md) | — | 12 | Plans, tasks, goals |
| [SharePoint](docs/sharepoint.md) | `sp` | 16 | Files, folders, sites, sharing |
| [SharePoint Lists](docs/sp-lists.md) | — | 13 | Lists, items, columns |
| [OneDrive](docs/onedrive-remote.md) | `odr` | 12 | Personal OneDrive file management |
| [Me](docs/me.md) | — | 5 | User profiles, org chart |
| [Copilot](docs/copilot.md) | — | 1 | Natural language M365 search with interactive chat |
| [Word](docs/word.md) | — | 4 | Documents, comments |
| [Excel](docs/excel.md) | — | 4 | Workbooks, comments |
| [Admin](docs/admin.md) | — | 3 | Users, licenses |
| [Admin365](docs/admin365.md) | — | 14 | Agent policies, Copilot settings |
| [Triggers](docs/triggers.md) | — | 9 | Event-driven automation |
| [WebSearch](docs/websearch.md) | — | 1 | Web search |
| [DASearch](docs/dasearch.md) | — | 1 | Discover Copilot agents |
| [Knowledge](docs/knowledge.md) | — | 5 | Federated knowledge |
| [NLWeb](docs/nlweb.md) | — | 3 | Natural language search |

Plus: `config` for settings, hidden `api` for MCP exploration.

## Output Formats

```bash
# Human table (default)
$ a365 teams channels list 00000000-0000-0000-0000-000000000000
DISPLAY NAME         ID                                     TYPE      CREATED
General              19:a1b2c3d4...@thread.tacv2           Standard  Jan 15
Engineering          19:e5f6a7b8...@thread.tacv2           Standard  Feb 20

# JSON (for scripting and agents)
$ a365 teams list -o json
{
  "teams": [
    {"id": "...", "displayName": "Project Alpha", ...}
  ]
}

# TSV (for piping)
$ a365 mail search '?$top=3' -o tsv | cut -f3
SUBJECT
Meeting tomorrow
Q4 Budget Review
```

`--json` and `--plain` still work as shorthand.

## Safety

All write operations support `--dry-run` with schema validation — arguments are validated against the server's published JSON Schema without executing the tool. Destructive operations prompt for confirmation (skip with `--force`, fail with `--no-input`).

```bash
# Preview without executing (validates args against server schema)
$ a365 teams chats send "19:abc@thread.v2" "Hello" --dry-run
Dry run: would send message to chat 19:abc@thread.v2

✓ Arguments valid against server schema

# JSON dry-run (for agents/CI)
$ a365 teams chats send "19:abc@thread.v2" "Hello" --dry-run -o json
{"action":"chats.send","chatId":"19:abc@thread.v2","content":"Hello","dry_run":true,"validation":{"valid":true,"errors":null}}
```

## Global Flags

| Flag | Env Var | Description |
|------|---------|-------------|
| `-o`, `--output` | `A365_OUTPUT` | Output format: `table`, `json`, or `tsv` |
| `--force` | | Skip confirmation prompts |
| `--no-input` | | Never prompt; fail instead (CI/agent mode) |
| `--dry-run` | | Preview write operations with schema validation |
| `-v`, `--verbose` | | Show MCP request/response for debugging |
| `--client-id` | `A365_CLIENT_ID` | Entra app client ID (has default) |
| `--tenant-id` | `A365_TENANT_ID` | Entra tenant ID |
| `-V`, `--version` | | Show version |

## API Explorer

Discover and call any MCP tool directly. See [docs/api-explorer.md](docs/api-explorer.md) for the full guide.

```bash
a365 api servers --probe           # List all servers with tool counts
a365 api discover                  # Live server catalog from gateway
a365 api tools teams               # List tools + required params
a365 api call me GetMyDetails '{}'  # Raw MCP tool call
```

## Shell Completion

```bash
# Bash
a365 completion bash > /etc/bash_completion.d/a365

# Zsh
a365 completion zsh > "${fpath[1]}/_a365"

# Fish
a365 completion fish > ~/.config/fish/completions/a365.fish
```

## Architecture

See [docs/architecture.md](docs/architecture.md) for the full architecture guide covering the request lifecycle, MCP protocol details, authentication flow, output pipeline, and design decisions.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for architecture, adding new services, and development workflow.

## License

[MIT](LICENSE)
