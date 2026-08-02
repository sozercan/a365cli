//go:build (darwin || linux) && cgo

package auth

import (
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity/cache"
)

// applyTokenCache enables persistent token cache backed by the OS credential store.
func applyTokenCache(opts *azidentity.InteractiveBrowserCredentialOptions) (bool, error) {
	if persistentTokenCacheDisabled() {
		return false, nil
	}
	c, err := cache.New(&cache.Options{Name: persistentTokenCacheName})
	if err != nil {
		return false, err
	}
	opts.Cache = c
	return true, nil
}

func clearTokenCache() error {
	if persistentTokenCacheDisabled() {
		return nil
	}
	var errs []error
	for _, name := range tokenCacheNamesForCleanup() {
		if err := clearPlatformTokenCache(name); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
