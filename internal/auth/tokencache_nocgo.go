//go:build !cgo

package auth

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

// applyTokenCache is a no-op when CGO is disabled. The OS credential store
// (macOS Keychain, etc.) requires CGO, so tokens won't persist across CLI
// invocations but will still work within each session.
func applyTokenCache(_ *azidentity.InteractiveBrowserCredentialOptions) {}
