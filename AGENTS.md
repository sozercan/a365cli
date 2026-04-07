# AGENTS.md

## Critical

- **NEVER leak PII or confidential information** — no real email addresses, user IDs, team IDs, client IDs, tokens, keys, or tenant IDs in source code, docs, examples, or commit messages. Use fake/placeholder values (e.g. `your-client-id-here`, `alice@contoso.com`, `00000000-0000-0000-0000-000000000000`).
- **Always validate with the API explorer before implementing** — run `a365 api tools <service>` to check tool names, required params, and schemas before writing command handlers. Do not guess tool names or argument shapes.

## Build and Test

```bash
make build          # build
make test           # go test ./... -v
go test ./... -cover  # with coverage
go vet ./...        # lint
```

Binary is always called `a365`. Module is `github.com/sozercan/a365cli`.

## Adding a Command

Every command follows this exact pattern:

```go
func (c *MyCmd) Run(ctx *commands.Context) error {
    // 1. Dry-run guard with schema validation (write ops only)
    if ctx.DryRun { return ctx.ValidateDryRun(endpoint(), "ToolName", "do X", displayData, mcpArgs) }
    // 2. Confirm guard (destructive ops only)
    if err := ctx.Confirm("delete X"); err != nil { return err }
    // 3. Create client + initialize
    client := ctx.NewMCPClient(endpoint())
    if err := client.Initialize(ctx.Ctx); err != nil { ... }
    // 4. Call MCP tool
    resp, err := client.CallTool(ctx.Ctx, "ToolName", map[string]any{...})
    // 5. Extract + print
    data, err := output.ExtractContent(resp)
    return ctx.Output.PrintList("items", output.SomeColumns, rows)  // or PrintItem/PrintMutation
}
```

Adding a new service = new directory in `internal/commands/`, register in `main.go`.

## Key Conventions

- Use `config.Endpoint("service")` for server URLs, never hardcode
- Use `output.ExtractContent()` then `ToRows()` for list data
- Write ops: always add `--dry-run` guard with `ctx.ValidateDryRun(endpoint, toolName, action, displayData, mcpArgs)`
- When display keys differ from MCP arg keys, pass mcpArgs as the 5th parameter for correct validation
- Destructive ops: add `ctx.Confirm()` after dry-run check
- Kong struct tags: `arg:""` for positional, `help:""` for description, `optional:""` for optional
- Test with `httptest.NewServer` mocks, override endpoint via `A365_ENDPOINT` env var
- Use `testutil.SetupTestServerWithSchemas` for dry-run tests that verify validation
- Keep column `Width` values consistent: names=30-40, IDs=36, dates=10, types=10

## Workflow

- Use plan mode for architectural changes or new service additions
- Prefer `make build` over `go build`
- Run `go test ./... -count=1` after any code change
- Run `go vet ./...` before committing
- When adding a new M365 service, follow the pattern in CONTRIBUTING.md "Adding a New Service"
- For live testing, set `A365_CLIENT_ID` to your Entra app client ID
- Never post messages or modify data during testing without explicit user permission — use `--dry-run`
