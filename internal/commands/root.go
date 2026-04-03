package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sozercan/a365cli/internal/auth"
	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/output"
)

// Context is the shared context passed to all command handlers.
type Context struct {
	Ctx           context.Context
	TokenProvider mcp.TokenProvider
	Output        *output.Formatter
	Verbose       bool
	ClientID      string
	TenantID      string
	UserUPN       string // Signed-in user's UPN (email), from cached auth record

	// Safety flags
	Force   bool // Skip confirmation prompts
	NoInput bool // Never prompt; fail instead (CI mode)
	DryRun  bool // Preview write operations without executing
}

// NewMCPClient creates an MCP client for the given endpoint, wired with auth and verbose settings.
func (c *Context) NewMCPClient(endpoint string) *mcp.Client {
	client := mcp.NewClient(endpoint, c.TokenProvider)
	if c.Verbose {
		client.SetVerbose(func(format string, args ...any) {
			fmt.Fprintf(os.Stderr, format+"\n", args...)
		})
	}
	return client
}

// EnsureAuth checks that the user is authenticated and sets up the token provider.
func (c *Context) EnsureAuth() error {
	cred, err := auth.NewCredential(c.ClientID, c.TenantID)
	if err != nil {
		return fmt.Errorf("authentication required — run 'a365 auth login' first: %w", err)
	}
	c.TokenProvider = cred.TokenProvider()
	c.UserUPN = auth.GetCachedUsername()
	return nil
}

// Confirm prompts the user for confirmation before a destructive action.
// Returns nil if confirmed, error if cancelled.
// Respects --force (skip prompt) and --no-input (hard error).
func (c *Context) Confirm(action string) error {
	if c.Force {
		return nil
	}

	if c.NoInput || !isTerminal() {
		return fmt.Errorf("refusing to %s without --force (non-interactive)", action)
	}

	fmt.Fprintf(os.Stderr, "Proceed to %s? [y/N]: ", action)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("cancelled")
	}

	line = strings.TrimSpace(strings.ToLower(line))
	if line == "y" || line == "yes" {
		return nil
	}
	return fmt.Errorf("cancelled")
}

// isTerminal checks if stdin is a terminal (not a pipe).
func isTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
