package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/sozercan/a365cli/internal/config"
)

// interactiveCredential is the subset of azidentity user credentials used here.
// Both InteractiveBrowserCredential and DeviceCodeCredential satisfy it.
type interactiveCredential interface {
	GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error)
	Authenticate(ctx context.Context, opts *policy.TokenRequestOptions) (azidentity.AuthenticationRecord, error)
}

// Credential wraps the Azure Identity credential for agent365 auth.
type Credential struct {
	cred       interactiveCredential
	scope      string
	persistent bool
}

// ErrHeadlessWithoutPersistentCache reports that a login can't be reused by
// later processes because there is no durable token cache, and no local
// browser is available to sign in again.
var ErrHeadlessWithoutPersistentCache = errors.New("this a365 build has no persistent token cache, so a login on a machine without a browser can't be reused by later commands — install a CGO-enabled build with 'make install' (see README › Authentication)")

// PersistentCache reports whether tokens acquired by this credential are
// stored in a durable cache that later CLI processes can use silently.
func (c *Credential) PersistentCache() bool {
	return c.persistent
}

// CredentialOptions controls interactive fallback behavior.
type CredentialOptions struct {
	DisableAutomaticAuthentication           bool
	AllowAutomaticAuthenticationWithoutCache bool

	// DeviceCode selects the device code flow instead of opening a browser.
	// Useful on headless machines (SSH sessions, containers, remote VMs).
	DeviceCode bool
	// DeviceCodePrompt receives the device code sign-in instructions.
	// Defaults to os.Stderr.
	DeviceCodePrompt io.Writer
	// RedirectPort fixes the local port of the browser flow's
	// http://localhost redirect listener (0 picks a random free port).
	RedirectPort int
}

// BrowserAvailable reports whether a browser launcher is likely available.
// On Linux this mirrors the launchers the underlying MSAL library tries.
func BrowserAvailable() bool {
	if runtime.GOOS != "linux" {
		return true
	}
	for _, name := range []string{"xdg-open", "x-www-browser", "www-browser"} {
		if _, err := exec.LookPath(name); err == nil {
			return true
		}
	}
	return false
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
	if credentialOpts.RedirectPort > 0 {
		opts.RedirectURL = fmt.Sprintf("http://localhost:%d", credentialOpts.RedirectPort)
	}

	// Enable persistent token cache so refresh tokens survive between CLI
	// invocations when the platform credential store is available.
	cacheApplied, err := applyTokenCache(&opts.Cache)
	persistent := cacheApplied && err == nil
	if policyErr := configureTokenCacheFallback(opts, cacheApplied, err, credentialOpts.AllowAutomaticAuthenticationWithoutCache); policyErr != nil {
		return nil, fmt.Errorf("initialize persistent token cache: %w", policyErr)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: persistent token cache unavailable; using interactive in-memory authentication: %v\n", err)
	}

	var cred interactiveCredential
	if credentialOpts.DeviceCode {
		cred, err = azidentity.NewDeviceCodeCredential(deviceCodeOptions(opts, credentialOpts.DeviceCodePrompt))
	} else {
		cred, err = azidentity.NewInteractiveBrowserCredential(opts)
	}
	if err != nil {
		return nil, fmt.Errorf("create credential: %w", err)
	}

	return &Credential{
		cred:       cred,
		scope:      config.DefaultScope,
		persistent: persistent,
	}, nil
}

func configureTokenCacheFallback(opts *azidentity.InteractiveBrowserCredentialOptions, cacheApplied bool, cacheErr error, allowWithoutCache bool) error {
	if cacheErr != nil && !allowWithoutCache {
		return cacheErr
	}
	if (!cacheApplied || cacheErr != nil) && allowWithoutCache {
		opts.DisableAutomaticAuthentication = false
	}
	return nil
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

// deviceCodeOptions mirrors the browser credential options for the device code
// flow so both share the same client, tenant, account record, and token cache.
func deviceCodeOptions(src *azidentity.InteractiveBrowserCredentialOptions, prompt io.Writer) *azidentity.DeviceCodeCredentialOptions {
	if prompt == nil {
		prompt = os.Stderr
	}
	return &azidentity.DeviceCodeCredentialOptions{
		ClientOptions:                  src.ClientOptions,
		AuthenticationRecord:           src.AuthenticationRecord,
		Cache:                          src.Cache,
		ClientID:                       src.ClientID,
		DisableAutomaticAuthentication: src.DisableAutomaticAuthentication,
		TenantID:                       src.TenantID,
		UserPrompt: func(_ context.Context, msg azidentity.DeviceCodeMessage) error {
			_, err := fmt.Fprintln(prompt, msg.Message)
			return err
		},
	}
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

	// A successful login must be durable enough for the next command to select
	// the same account. Report persistence failures instead of claiming success.
	if err := cacheAuthenticationRecord(&record); err != nil {
		return azcore.AccessToken{}, err
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

func cacheAuthenticationRecord(record *azidentity.AuthenticationRecord) error {
	if err := SaveAuthRecord(record); err != nil {
		return fmt.Errorf("cache auth record: %w", err)
	}
	return nil
}
