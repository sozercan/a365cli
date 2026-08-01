//go:build !cgo

package auth

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

// applyTokenCache is a no-op when CGO is disabled. The OS credential store
// requires CGO, so tokens won't persist across CLI invocations. The auth
// record still selects the account, but a later process may require another
// explicit login.
func applyTokenCache(_ *azidentity.InteractiveBrowserCredentialOptions) (bool, error) {
	// Keep the link-time cache namespace present in all build variants.
	_ = persistentTokenCacheName
	return false, nil
}

func clearTokenCache() error { return nil }
