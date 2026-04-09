package auth

import (
	"errors"
	"os"
	"path/filepath"
)

const persistentTokenCacheName = "a365"

func removeFileIfExists(path string) error {
	err := os.Remove(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func identityServiceArtifactPaths(cacheRoot, name string) []string {
	base := filepath.Join(cacheRoot, ".IdentityService", name)
	return []string{base, base + ".lockfile"}
}
