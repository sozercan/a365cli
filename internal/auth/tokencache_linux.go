//go:build linux && cgo

package auth

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AzureAD/microsoft-authentication-extensions-for-go/cache/accessor"
)

func clearPlatformTokenCache(name string) error {
	storage, err := accessor.New(name)
	if err != nil {
		return err
	}
	if err := storage.Delete(context.Background()); err != nil {
		return err
	}

	dir := os.Getenv("XDG_CACHE_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get home dir: %w", err)
		}
		dir = filepath.Join(home, ".cache")
	}

	for _, path := range identityServiceArtifactPaths(dir, name) {
		if err := removeFileIfExists(path); err != nil {
			return err
		}
	}
	return nil
}
