package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveFileIfExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache")

	if err := removeFileIfExists(path); err != nil {
		t.Fatalf("removeFileIfExists on missing file returned error: %v", err)
	}

	if err := os.WriteFile(path, []byte("content"), 0o600); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	if err := removeFileIfExists(path); err != nil {
		t.Fatalf("removeFileIfExists returned error: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected file to be removed, got err=%v", err)
	}
}

func TestIdentityServiceArtifactPaths(t *testing.T) {
	paths := identityServiceArtifactPaths("/tmp/cache-root", "a365")
	if len(paths) != 2 {
		t.Fatalf("expected 2 artifact paths, got %d", len(paths))
	}
	if got, want := paths[0], filepath.Join("/tmp/cache-root", ".IdentityService", "a365"); got != want {
		t.Fatalf("artifact path[0] = %q, want %q", got, want)
	}
	if got, want := paths[1], filepath.Join("/tmp/cache-root", ".IdentityService", "a365")+".lockfile"; got != want {
		t.Fatalf("artifact path[1] = %q, want %q", got, want)
	}
}
