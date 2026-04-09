package commands

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/sozercan/a365cli/internal/auth"
	"github.com/sozercan/a365cli/internal/mcp"
	"github.com/sozercan/a365cli/internal/output"
)

// AuthCmd groups authentication subcommands.
type AuthCmd struct {
	Login  AuthLoginCmd  `cmd:"" help:"Log in via browser (interactive)"`
	Status AuthStatusCmd `cmd:"" help:"Show current authentication status"`
	Token  AuthTokenCmd  `cmd:"" help:"Show current token details (scopes, expiry)"`
	Logout AuthLogoutCmd `cmd:"" help:"Log out and clear cached credentials"`
}

// AuthLoginCmd performs interactive browser login.
type AuthLoginCmd struct{}

func (c *AuthLoginCmd) Run(ctx *Context) error {
	if ctx.NoInput || !isTerminal() {
		return fmt.Errorf("login requires interactive input")
	}
	fmt.Fprintln(ctx.Output.Writer, "Opening browser for authentication...")

	cred, err := auth.NewCredential(ctx.ClientID, ctx.TenantID)
	if err != nil {
		return fmt.Errorf("create credential: %w", err)
	}

	token, err := cred.Authenticate(context.Background())
	if err != nil {
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
	if !auth.HasCachedAuth() {
		fmt.Fprintln(ctx.Output.Writer, "Not authenticated. Run 'a365 auth login' to sign in.")
		return nil
	}

	username := auth.GetCachedUsername()
	if username != "" {
		fmt.Fprintf(ctx.Output.Writer, "Authenticated as %s\n", username)
	} else {
		fmt.Fprintln(ctx.Output.Writer, "Authenticated (cached credentials found)")
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
