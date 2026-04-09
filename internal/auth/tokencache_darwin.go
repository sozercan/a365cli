//go:build darwin && cgo

package auth

import (
	"context"
	"fmt"
	"os"

	"github.com/AzureAD/microsoft-authentication-extensions-for-go/cache/accessor"
)

func clearPlatformTokenCache(name string) error {
	storage, err := accessor.New(name, accessor.WithAccount("MSALCache"))
	if err != nil {
		return err
	}
	if err := storage.Delete(context.Background()); err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}
	for _, path := range identityServiceArtifactPaths(home, name) {
		if err := removeFileIfExists(path); err != nil {
			return err
		}
	}
	return nil
}
