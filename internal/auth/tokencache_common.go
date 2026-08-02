package auth

import (
	"errors"
	"os"
	"path/filepath"
)

// persistentTokenCacheName may be overridden at link time so development and
// release binaries don't compete for the same Keychain ACL.
var persistentTokenCacheName = "a365-dev"

const disablePersistentTokenCacheEnv = "A365_DISABLE_PERSISTENT_TOKEN_CACHE"

func persistentTokenCacheDisabled() bool {
	return os.Getenv(disablePersistentTokenCacheEnv) == "1"
}

func tokenCacheNamesForCleanup() []string {
	names := []string{persistentTokenCacheName}
	for _, name := range []string{"a365", "a365-dev"} {
		seen := false
		for _, existing := range names {
			if existing == name {
				seen = true
				break
			}
		}
		if !seen {
			names = append(names, name)
		}
	}
	return names
}

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
