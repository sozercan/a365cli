package commands

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/sozercan/a365cli/internal/auth"
	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/output"
)

// AuthCmd groups authentication subcommands.
type AuthCmd struct {
	Login  AuthLoginCmd  `cmd:"" help:"Log in via browser or device code (interactive)"`
	Status AuthStatusCmd `cmd:"" help:"Show current authentication status"`
	Token  AuthTokenCmd  `cmd:"" help:"Show current token details (scopes, expiry)"`
	Logout AuthLogoutCmd `cmd:"" help:"Log out and clear cached credentials"`
}

// AuthLoginCmd performs interactive browser or device code login.
type AuthLoginCmd struct {
	NoBrowser  bool `help:"Don't launch a browser; print the sign-in URL to open on another machine. The browser there must reach this host's --port on localhost (e.g. ssh -L). Used automatically when no browser launcher is found." xor:"flow"`
	Port       int  `help:"Local redirect port used with --no-browser." default:"8400"`
	DeviceCode bool `help:"Sign in with a device code instead of a browser redirect. Some tenants block this flow with Conditional Access." xor:"flow"`
}

func (c *AuthLoginCmd) Run(ctx *Context) error {
	if ctx.NoInput || !isTerminal() {
		return fmt.Errorf("login requires interactive input")
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("invalid --port %d: must be between 1 and 65535", c.Port)
	}

	noBrowser := c.NoBrowser
	if !noBrowser && !c.DeviceCode && !auth.BrowserAvailable() {
		fmt.Fprintln(ctx.Output.Writer, "No browser launcher found; printing the sign-in URL instead (--no-browser).")
		noBrowser = true
	}
	if !noBrowser && !c.DeviceCode {
		fmt.Fprintln(ctx.Output.Writer, "Opening browser for authentication...")
	}

	credOpts := auth.CredentialOptions{
		DisableAutomaticAuthentication:           true,
		AllowAutomaticAuthenticationWithoutCache: true,
		DeviceCode:                               c.DeviceCode,
		DeviceCodePrompt:                         ctx.Output.Writer,
	}
	if noBrowser {
		credOpts.RedirectPort = c.Port
	}
	cred, err := auth.NewCredentialWithOptions(ctx.ClientID, ctx.TenantID, credOpts)
	if err != nil {
		return fmt.Errorf("create credential: %w", err)
	}
	// A headless login is only useful if later commands can reuse its tokens.
	if (noBrowser || c.DeviceCode) && !cred.PersistentCache() {
		return auth.ErrHeadlessWithoutPersistentCache
	}

	// Cancel cleanly on Ctrl+C so temporary login state is cleaned up.
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var token azcore.AccessToken
	if noBrowser {
		token, err = cred.AuthenticateWithoutBrowser(sigCtx, func(u string) {
			printNoBrowserInstructions(ctx.Output.Writer, u, c.Port)
		})
	} else {
		token, err = cred.Authenticate(sigCtx)
	}
	if err != nil {
		if !noBrowser && !c.DeviceCode {
			return fmt.Errorf("login failed (on a headless machine, try 'a365 auth login --no-browser'): %w", err)
		}
		return fmt.Errorf("login failed: %w", err)
	}

	username := auth.GetCachedUsername()
	if username != "" {
		fmt.Fprintf(ctx.Output.Writer, "Authenticated as %s\n", username)
	} else {
		fmt.Fprintln(ctx.Output.Writer, "Authenticated successfully")
	}
	fmt.Fprintf(ctx.Output.Writer, "Token expires: %s\n", token.ExpiresOn.Local().Format("2006-01-02 15:04:05"))

	return nil
}

// AuthStatusCmd shows current auth status.
type AuthStatusCmd struct{}

func (c *AuthStatusCmd) Run(ctx *Context) error {
	record, err := auth.LoadAuthRecord()
	if errors.Is(err, os.ErrNotExist) {
		fmt.Fprintln(ctx.Output.Writer, "Not authenticated. Run 'a365 auth login' to sign in.")
		return nil
	}
	if err != nil {
		return fmt.Errorf("cached authentication record is invalid — run 'a365 auth logout' then 'a365 auth login': %w", err)
	}

	username := record.Username
	if username != "" {
		fmt.Fprintf(ctx.Output.Writer, "Cached account: %s (token not verified)\n", username)
	} else {
		fmt.Fprintln(ctx.Output.Writer, "Cached account record found (token not verified)")
	}

	return nil
}

// AuthLogoutCmd clears cached credentials.
type AuthLogoutCmd struct{}

func (c *AuthLogoutCmd) Run(ctx *Context) error {
	if err := auth.RemoveAuthRecord(); err != nil {
		return fmt.Errorf("logout: %w", err)
	}
	mcp.ClearSessions()
	fmt.Fprintln(ctx.Output.Writer, "Logged out successfully. Cached credentials removed.")
	return nil
}

// AuthTokenCmd shows decoded token details.
type AuthTokenCmd struct{}

func (c *AuthTokenCmd) Run(ctx *Context) error {
	if err := ctx.EnsureAuth(); err != nil {
		return err
	}

	token, err := ctx.TokenProvider(context.Background())
	if err != nil {
		return fmt.Errorf("get token: %w", err)
	}

	// Decode JWT payload (part 2 of 3)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid token format (expected JWT)")
	}

	// Add padding if needed
	payload := parts[1]
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return fmt.Errorf("decode token: %w", err)
	}

	var claims map[string]any
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return fmt.Errorf("parse token claims: %w", err)
	}

	if ctx.Output.Format == output.FormatJSON {
		return ctx.Output.PrintItem(claims)
	}

	printTokenClaims(ctx.Output.Writer, claims)
	return nil
}

func printNoBrowserInstructions(w io.Writer, signInURL string, port int) {
	fmt.Fprintf(w, `
After sign-in, your browser is redirected to http://localhost:%[1]d.
If your browser runs on a different machine, forward that port to this host first, e.g.:

    ssh -L %[1]d:localhost:%[1]d <this-host>

(VS Code Remote and similar tools may forward it automatically.)

Open this URL in your browser:

    %[2]s

Waiting for sign-in... (Ctrl+C to cancel)
`, port, signInURL)
}

func printTokenClaims(w io.Writer, claims map[string]any) {
	fmt.Fprintf(w, "App ID:    %v\n", claims["appid"])
	fmt.Fprintf(w, "Tenant:    %v\n", claims["tid"])
	fmt.Fprintf(w, "User:      %v\n", claims["upn"])
	fmt.Fprintf(w, "Name:      %v\n", claims["name"])
	fmt.Fprintf(w, "Audience:  %v\n", claims["aud"])

	if scp, ok := claims["scp"]; ok {
		fmt.Fprintf(w, "Scopes:    %v\n", scp)
	} else {
		fmt.Fprintln(w, "Scopes:    (none - check app registration)")
	}

	if exp, ok := claims["exp"].(float64); ok {
		expTime := time.Unix(int64(exp), 0)
		remaining := time.Until(expTime)
		if remaining > 0 {
			fmt.Fprintf(w, "Expires:   %s (in %s)\n", expTime.Local().Format("2006-01-02 15:04:05"), remaining.Round(time.Second))
		} else {
			fmt.Fprintf(w, "Expires:   %s (EXPIRED %s ago)\n", expTime.Local().Format("2006-01-02 15:04:05"), (-remaining).Round(time.Second))
		}
	}
}
