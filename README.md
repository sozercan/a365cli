# a365

A standalone, agent-friendly CLI for Microsoft 365 via [agent365](https://agent365.svc.cloud.microsoft) MCP servers.

## Features

- **18 M365 services** — Teams, Mail, Calendar, Planner, SharePoint, OneDrive, Word, Excel, Copilot, Admin, Triggers, WebSearch, and more
- **170+ MCP tools** — full API coverage with dynamic server discovery
- **Agent-friendly** — structured `--output=json` for LLM tool use, `--no-input` for non-interactive execution, `--dry-run` for safe exploration
- **Three output modes** — human tables (default), JSON for scripting, TSV for piping
- **Interactive browser auth** with PKCE — persistent silent re-auth on supported native-cache builds
- **Resilient** — automatic retries with backoff on 502/503/429, MCP session caching
- **Configurable** — `~/.a365/config.json` for persistent defaults, env vars, CLI flags
- **Shell completion** — bash, zsh, fish
- **API explorer** — discover and call any MCP tool directly with `a365 api`

## Quick Start

```bash
# Install
brew tap sozercan/repo && brew install a365

# Authenticate (native-cache builds persist tokens in the OS credential store)
a365 auth login

# Use it
a365 teams list                               # List your Teams
a365 mail search '?$top=5'                    # Recent emails
a365 cal list                                 # Upcoming meetings
a365 copilot chat "Summarize my week"         # Ask Copilot
a365 copilot agents                           # List available Copilot agents
a365 copilot chat                             # Interactive Copilot prompt
a365 me whoami                                # Your profile
a365 odr ls                                   # OneDrive files
a365 websearch search "MCP protocol" \
  https://modelcontextprotocol.io             # Web search
```

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap sozercan/repo
brew install a365
```

The macOS Homebrew artifact is a CGO-enabled native build with
certificate-backed code signing. The release workflow accepts exactly one
**Developer ID Application** or **Apple Development** identity and reuses the
fixed `com.github.sozercan.a365` identifier so Keychain authorization remains
stable across versions signed by the same identity. Apple Development signing
does not provide Developer ID distribution trust or notarization. Homebrew on
Linux installs the portable CGO-disabled artifact, which does not persist the
OS-backed token cache between CLI processes.

### GitHub Releases

Pre-built binaries are available on the [Releases](https://github.com/sozercan/a365cli/releases) page:

| Platform | Release build | Persistent OS token cache |
|----------|---------------|---------------------------|
| macOS | Native, CGO enabled, certificate-backed | Yes, through Keychain |
| Linux | Portable, CGO disabled | No |
| Windows | Portable, CGO disabled | No |

Static/portable builds still authenticate, but a later interactive CLI process
may open the browser again because the non-CGO implementation does not persist
the refresh-token cache. `--no-input` always disables that fallback.

### go install

```bash
go install github.com/sozercan/a365cli@latest
```

`go install` uses the local Go toolchain's CGO setting and does not apply the
project's macOS certificate-signing policy. On macOS, do not use a raw
`go install` or `go build` result as the canonical `a365` on your PATH: replacing
a Keychain-enabled executable with differently signed builds can cause repeated
Keychain authorization prompts. Use the Makefile workflow below for source
builds.

### Build from source

On macOS, the canonical native build automatically selects the first valid
**Apple Development** codesigning identity, falling back to **Developer ID
Application** when needed:

```bash
make build
```

Override the detected identity with a certificate SHA-1 or exact identity name:

```bash
A365_CODESIGN_IDENTITY="<certificate SHA-1 or exact identity name>" make build
```

If no valid certificate-backed identity exists, the build fails with guidance
instead of silently producing an unstable canonical binary.

Use `make build-adhoc` only for a disposable development binary. It is
intentionally ad-hoc signed and may trigger a new Keychain authorization prompt
when it replaces another build. The first migration to the fixed canonical
identifier and certificate may also require one deliberate Keychain approval;
keep the identity and install path stable afterward. Use `make build-static` for
a portable pure-Go binary without persistent OS token caching. See
[CONTRIBUTING.md](CONTRIBUTING.md) for the complete build and release matrix.

## Authentication

a365 uses Entra ID interactive browser authentication with PKCE. On a supported native-cache build, the refresh-token cache is persisted in the OS credential store so later processes can refresh silently; ordinary commands fail with explicit login guidance instead of unexpectedly opening a browser when that cache is unavailable. Portable CGO-disabled builds have no durable token cache, so interactive commands may open the browser again to remain usable. `--no-input` is a hard no-browser mode on every build.

A built-in client ID is provided by default. If your tenant requires a custom app registration, set your own via `--client-id` or `A365_CLIENT_ID`.

```bash
# Login (uses the built-in client ID by default)
a365 auth login

# Check for cached account metadata (does not validate a live token)
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
a365 config set client-id your-client-id-here  # Override the default client ID
a365 config set output json                    # Set default output: table, json, or tsv
a365 config show                               # Show all settings
a365 config path                               # Show the config file path
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
| [Copilot](docs/copilot.md) | — | 2 | Natural language M365 search with agent-aware chat |
| [Word](docs/word.md) | — | 4 | Documents, comments |
| [Excel](docs/excel.md) | — | 4 | Workbooks, comments |
| [Admin](docs/admin.md) | — | 3 | Users, licenses |
| [Admin365](docs/admin365.md) | — | 14 | Agent policies, Copilot settings |
| [Triggers](docs/triggers.md) | — | 9 | Event-driven automation |
| [WebSearch](docs/websearch.md) | — | 1 | Web search |
| [DASearch](docs/dasearch.md) | — | 1 | Low-level Copilot agent discovery (raw DASearch output) |
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
a365 api servers --probe              # List all servers with tool counts
a365 api discover                     # Live server catalog from gateway
a365 api tools teams                  # List tools and required params
a365 api call me GetMyDetails '{}'    # Raw MCP tool call
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
