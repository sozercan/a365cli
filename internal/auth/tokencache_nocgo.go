//go:build !cgo

package auth

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

// applyTokenCache is a no-op when CGO is disabled. The OS credential store
// requires CGO, so tokens won't persist across CLI invocations. The auth
// record still enables session re-use but may prompt for browser login.
func applyTokenCache(_ *azidentity.InteractiveBrowserCredentialOptions) {}

func clearTokenCache() error { return nil }
