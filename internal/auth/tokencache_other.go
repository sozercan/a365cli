//go:build cgo && !darwin && !linux

package auth

import "github.com/Azure/azure-sdk-for-go/sdk/azidentity"

// applyTokenCache is a no-op on platforms where the OS-backed cache
// integration is not implemented yet.
func applyTokenCache(_ *azidentity.InteractiveBrowserCredentialOptions) (bool, error) {
	// Keep the link-time cache namespace present in all build variants.
	_ = persistentTokenCacheName
	return false, nil
}

func clearTokenCache() error { return nil }
