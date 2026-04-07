//go:build cgo

package auth

import (
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity/cache"
)

// applyTokenCache enables persistent token cache backed by the OS credential store
// (macOS Keychain, Windows Credential Manager, Linux secret service).
func applyTokenCache(opts *azidentity.InteractiveBrowserCredentialOptions) {
	c, err := cache.New(&cache.Options{Name: "a365"})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: persistent token cache unavailable: %v\n", err)
		return
	}
	opts.Cache = c
}
