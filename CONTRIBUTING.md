# Contributing to a365

## Prerequisites

- Go 1.26+
- An Entra ID app registration with agent365 scopes (or access to one)
- macOS, Linux, or Windows

## Project Structure

See [docs/architecture.md](docs/architecture.md) for the full file layout and component descriptions.

## Building

```bash
A365_CODESIGN_IDENTITY="<certificate SHA-1 or exact identity name>" make build
make build-adhoc  # Disposable native development build; ad-hoc signed on macOS
make build-static # Portable pure-Go build without persistent OS token caching
make test         # Run all tests with verbose output
make test-short   # Run tests without verbose output
make lint         # Format + vet
make release-check # Statically validate .goreleaser.yml (requires GoReleaser)
make clean        # Remove build artifacts
```

The test targets set `A365_DISABLE_PERSISTENT_TOKEN_CACHE=1` so unit tests can
compile native cache support without reading or writing a developer's real OS
credential store.

Source builds use the `a365-dev` token-cache namespace by default, while
published releases use `a365`. This prevents local and release signatures from
repeatedly taking ownership of the same Keychain item. Override with
`A365_TOKEN_CACHE_NAME` only for controlled testing.

### Native and portable builds

| Build | CGO | Token-cache behavior | Intended use |
|-------|-----|----------------------|--------------|
| `make build` / `make build-cgo` | Enabled | macOS Keychain or the supported Linux OS credential store | Canonical native development/install build |
| `make build-adhoc` | Enabled | Same native cache code, but executable authorization may change after replacement | Disposable macOS development only |
| `make build-static` | Disabled | No persistent OS-backed cache between CLI processes | Portable and cross-compiled artifacts |

The static build still authenticates, but a later interactive CLI process may
open the browser again because the non-CGO cache implementation does not persist
refresh tokens. `--no-input` disables that fallback.

For portable Linux and Windows cross-builds, use the static target explicitly:

```bash
make build-static GOOS=linux GOARCH=amd64 BINARY=/tmp/a365-linux-amd64
make build-static GOOS=windows GOARCH=amd64 BINARY=/tmp/a365-windows-amd64.exe
```

Do not use raw `go build` or `go install` to replace the canonical macOS binary.
Those commands bypass the repository signing policy. Alternating between
unsigned, ad-hoc-signed, and certificate-signed executables that share the same
token-cache item can cause repeated Keychain authorization dialogs.

### macOS signing policy

`make build` and `make install` fail closed on macOS unless
`A365_CODESIGN_IDENTITY` explicitly names a valid local certificate-backed
codesigning identity. The Makefile does not auto-select the first certificate
and does not silently fall back to ad-hoc signing. Canonical binaries use the
stable identifier `com.github.sozercan.a365` and their resulting signature and
identifier are verified before the target succeeds.

Use `make build-adhoc` or `make install-adhoc` only when ad-hoc signing is an
intentional, temporary choice. Do not put that build at the same canonical PATH
location as a certificate-signed binary if you want stable Keychain approval.
Adopting the fixed identifier or a new certificate can require one deliberate
approval; subsequent canonical builds should retain both values.

### Release matrix and signing

GoReleaser uses two disjoint build definitions:

- macOS `amd64` and `arm64`: `CGO_ENABLED=1`, built on the macOS release runner,
  signed with hardened runtime and a secure timestamp, and checked for both the
  native CGO setting and a certificate-backed signature before packaging;
- Linux `amd64`/`arm64` and Windows `amd64`: `CGO_ENABLED=0`, preserving portable
  cross-builds from the macOS runner.

The generated Homebrew formula therefore installs the native Keychain-capable
artifact on macOS and the portable artifact on Linux.

The release workflow requires these GitHub Actions secrets:

- `MACOS_CERTIFICATE`: base64-encoded PKCS#12 archive containing exactly one
  valid **Developer ID Application** identity;
- `MACOS_CERTIFICATE_PWD`: password for that archive;
- `MACOS_KEYCHAIN_PWD`: password for the temporary CI keychain.

Missing, malformed, ambiguous, expired, or non-Developer-ID signing material
fails the release. Release builds never fall back to Apple Development or ad-hoc
signing. The workflow also runs `goreleaser check` before importing signing
material, and removes the temporary keychain after the release step.

## Adding a New Service

Adding a new M365 service is straightforward. Here's the pattern:

### 1. Add the server mapping to `internal/config/config.go`

```go
var Servers = map[string]string{
    // ...existing...
    "myservice": "mcp_MyServiceServer",
}
```

### 2. Create the command file

```bash
mkdir -p internal/commands/myservice
```

Create `internal/commands/myservice/myservice.go`:

```go
package myservice

import (
    "fmt"
    "github.com/sozercan/a365cli/internal/commands"
    "github.com/sozercan/a365cli/internal/config"
    "github.com/sozercan/a365cli/internal/output"
)

type MyServiceCmd struct {
    List MyServiceListCmd `cmd:"" help:"List items"`
    Get  MyServiceGetCmd  `cmd:"" help:"Get an item"`
}

func endpoint() string {
    return config.Endpoint("myservice")
}

type MyServiceListCmd struct {
    Max int `help:"Maximum results" default:"50"`
}

func (c *MyServiceListCmd) Run(ctx *commands.Context) error {
    data, err := ctx.CallToolData(endpoint(), "ListItems", "list items", map[string]any{})
    if err != nil {
        return err
    }
    return ctx.Output.PrintListFromData("items", output.MyColumns, data, c.Max, "items", "value")
}
```

### 3. Register in `main.go`

```go
import "github.com/sozercan/a365cli/internal/commands/myservice"

type CLI struct {
    // ...existing...
    MyService myservice.MyServiceCmd `cmd:"" help:"My Service"`
}
```

### 4. Add column definitions (optional)

If the service returns list data, add columns to `internal/output/columns.go`:

```go
var MyColumns = []Column{
    {Header: "NAME", Width: 30, Extract: func(row map[string]any) string {
        return getString(row, "name")
    }},
    {Header: "ID", Width: 36, Extract: func(row map[string]any) string {
        return getString(row, "id")
    }},
}
```

### 5. Add documentation

Create `docs/myservice.md` with commands table and usage examples.

That's it — no framework changes, no new dependencies.

## Testing

```bash
make test       # Run all tests
go test ./internal/mcp/... -v    # MCP client tests only
go test ./internal/output/... -v # Output formatting tests only
```

The test suite uses `httptest.NewServer` for MCP client tests, `commands.Context.CallToolData` tests for command execution, and `bytes.Buffer` injection for output formatter tests. Use `testutil.SetupTestServerWithSchemas` for dry-run tests that verify schema validation. No real network calls in tests.

## Discovering MCP Tools

See [docs/api-explorer.md](docs/api-explorer.md) for the full guide. Quick reference:

```bash
a365 api servers --probe    # List servers with tool counts
a365 api tools teams        # List tools + required params
a365 api call me GetMyDetails '{}'  # Raw call
```

## Architecture Decisions

See [docs/architecture.md](docs/architecture.md) for the full architecture guide including request lifecycle, MCP protocol, auth flow, output pipeline, and design rationale.
