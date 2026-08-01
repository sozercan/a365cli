package auth

import (
	"context"
	"errors"
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

// CredentialOptions controls interactive fallback behavior.
type CredentialOptions struct {
	DisableAutomaticAuthentication           bool
	AllowAutomaticAuthenticationWithoutCache bool
}

// NewCredential creates a new InteractiveBrowserCredential with PKCE.
// clientID and tenantID are optional; if empty, env vars / defaults are used.
func NewCredential(clientID, tenantID string) (*Credential, error) {
	return NewCredentialWithOptions(clientID, tenantID, CredentialOptions{
		DisableAutomaticAuthentication: true,
	})
}

// NewCredentialWithOptions creates a credential with explicit fallback policy.
func NewCredentialWithOptions(clientID, tenantID string, credentialOpts CredentialOptions) (*Credential, error) {
	// Load the cached account selection used for silent token acquisition.
	record, err := LoadAuthRecord()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("load auth record: %w", err)
	}

	opts := credentialOptions(clientID, tenantID, record, credentialOpts.DisableAutomaticAuthentication)

	// Enable persistent token cache so refresh tokens survive between CLI
	// invocations when the platform credential store is available.
	cacheApplied, err := applyTokenCache(opts)
	if err != nil {
		return nil, fmt.Errorf("initialize persistent token cache: %w", err)
	}
	applyAutomaticAuthenticationFallback(opts, cacheApplied, credentialOpts.AllowAutomaticAuthenticationWithoutCache)

	cred, err := azidentity.NewInteractiveBrowserCredential(opts)
	if err != nil {
		return nil, fmt.Errorf("create credential: %w", err)
	}

	return &Credential{
		cred:  cred,
		scope: config.DefaultScope,
	}, nil
}

func applyAutomaticAuthenticationFallback(opts *azidentity.InteractiveBrowserCredentialOptions, cacheApplied, allowWithoutCache bool) {
	if !cacheApplied && allowWithoutCache {
		opts.DisableAutomaticAuthentication = false
	}
}

func credentialOptions(clientID, tenantID string, record *azidentity.AuthenticationRecord, disableAutomaticAuthentication bool) *azidentity.InteractiveBrowserCredentialOptions {
	opts := &azidentity.InteractiveBrowserCredentialOptions{
		DisableAutomaticAuthentication: disableAutomaticAuthentication,
	}

	if clientID != "" {
		opts.ClientID = clientID
	}
	if tenantID != "" {
		opts.TenantID = tenantID
	}
	if record != nil {
		opts.AuthenticationRecord = *record
	}

	return opts
}

// TokenProvider returns a function that provides bearer tokens for MCP requests.
func (c *Credential) TokenProvider() func(ctx context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		token, err := c.cred.GetToken(ctx, policy.TokenRequestOptions{
			Scopes: []string{c.scope},
		})
		if err != nil {
			var required *azidentity.AuthenticationRequiredError
			if errors.As(err, &required) {
				return "", fmt.Errorf("authentication required — run 'a365 auth login' first: %w", err)
			}
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
