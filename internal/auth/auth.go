package auth

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/sozercan/a365cli/internal/config"
)

// Credential wraps the Azure Identity credential for agent365 auth.
type Credential struct {
	cred  *azidentity.InteractiveBrowserCredential
	scope string
}

// NewCredential creates a new InteractiveBrowserCredential with PKCE.
// clientID and tenantID are optional; if empty, env vars / defaults are used.
func NewCredential(clientID, tenantID string) (*Credential, error) {
	opts := &azidentity.InteractiveBrowserCredentialOptions{}

	if clientID != "" {
		opts.ClientID = clientID
	}

	if tenantID != "" {
		opts.TenantID = tenantID
	}

	// Load cached auth record for silent re-auth.
	record, err := LoadAuthRecord()
	if err == nil && record != nil {
		opts.AuthenticationRecord = *record
	}

	// Enable persistent token cache so refresh tokens survive between CLI
	// invocations when the platform credential store is available.
	applyTokenCache(opts)

	cred, err := azidentity.NewInteractiveBrowserCredential(opts)
	if err != nil {
		return nil, fmt.Errorf("create credential: %w", err)
	}

	return &Credential{
		cred:  cred,
		scope: config.DefaultScope,
	}, nil
}

// TokenProvider returns a function that provides bearer tokens for MCP requests.
func (c *Credential) TokenProvider() func(ctx context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		token, err := c.cred.GetToken(ctx, policy.TokenRequestOptions{
			Scopes: []string{c.scope},
		})
		if err != nil {
			return "", fmt.Errorf("get token: %w", err)
		}
		return token.Token, nil
	}
}

// Authenticate performs the interactive login and caches the auth record.
func (c *Credential) Authenticate(ctx context.Context) (azcore.AccessToken, error) {
	opts := &policy.TokenRequestOptions{
		Scopes: []string{c.scope},
	}

	record, err := c.cred.Authenticate(ctx, opts)
	if err != nil {
		return azcore.AccessToken{}, fmt.Errorf("authenticate: %w", err)
	}

	// Cache the auth record for silent re-auth on next run
	if saveErr := SaveAuthRecord(&record); saveErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not cache auth record: %v\n", saveErr)
	}

	// Now get the actual access token
	token, err := c.cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{c.scope},
	})
	if err != nil {
		return azcore.AccessToken{}, fmt.Errorf("get token after auth: %w", err)
	}

	return token, nil
}
