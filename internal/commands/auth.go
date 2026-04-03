package commands

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sozercan/a365cli/internal/auth"
	"github.com/sozercan/a365cli/internal/mcp"
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
	fmt.Println("Opening browser for authentication...")

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
		fmt.Printf("Authenticated as %s\n", username)
	} else {
		fmt.Println("Authenticated successfully")
	}
	fmt.Printf("Token expires: %s\n", token.ExpiresOn.Local().Format("2006-01-02 15:04:05"))

	return nil
}

// AuthStatusCmd shows current auth status.
type AuthStatusCmd struct{}

func (c *AuthStatusCmd) Run(ctx *Context) error {
	if !auth.HasCachedAuth() {
		fmt.Println("Not authenticated. Run 'a365 auth login' to sign in.")
		return nil
	}

	username := auth.GetCachedUsername()
	if username != "" {
		fmt.Printf("Authenticated as %s\n", username)
	} else {
		fmt.Println("Authenticated (cached credentials found)")
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
	fmt.Println("Logged out successfully. Cached credentials removed.")
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

	// Print key claims in human-readable format
	fmt.Printf("App ID:    %v\n", claims["appid"])
	fmt.Printf("Tenant:    %v\n", claims["tid"])
	fmt.Printf("User:      %v\n", claims["upn"])
	fmt.Printf("Name:      %v\n", claims["name"])
	fmt.Printf("Audience:  %v\n", claims["aud"])

	// Scopes
	if scp, ok := claims["scp"]; ok {
		fmt.Printf("Scopes:    %v\n", scp)
	} else {
		fmt.Println("Scopes:    (none — check app registration)")
	}

	// Expiry
	if exp, ok := claims["exp"].(float64); ok {
		expTime := time.Unix(int64(exp), 0)
		remaining := time.Until(expTime)
		if remaining > 0 {
			fmt.Printf("Expires:   %s (in %s)\n", expTime.Local().Format("2006-01-02 15:04:05"), remaining.Round(time.Second))
		} else {
			fmt.Printf("Expires:   %s (EXPIRED %s ago)\n", expTime.Local().Format("2006-01-02 15:04:05"), (-remaining).Round(time.Second))
		}
	}

	// Full claims in JSON if --json output
	if ctx.Output.Format == 1 { // FormatJSON
		return ctx.Output.PrintItem(claims)
	}

	return nil
}
