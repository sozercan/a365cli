package auth

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestInstallBrowserShimCapturesURL(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("browser shim is not supported on windows")
	}
	origPath := os.Getenv("PATH")

	dir, restore, err := installBrowserShim()
	if err != nil {
		t.Fatalf("installBrowserShim() error = %v", err)
	}

	launcher := "xdg-open"
	if runtime.GOOS == "darwin" {
		launcher = "open"
	}
	resolved, err := exec.LookPath(launcher)
	if err != nil {
		restore()
		t.Fatalf("LookPath(%q) error = %v", launcher, err)
	}
	if filepath.Dir(resolved) != dir {
		restore()
		t.Fatalf("%s resolved to %q, want shim in %q", launcher, resolved, dir)
	}

	const want = "https://login.microsoftonline.com/organizations/oauth2/v2.0/authorize?client_id=00000000-0000-0000-0000-000000000000&state=a&b=c"
	if out, err := exec.Command(launcher, want).CombinedOutput(); err != nil {
		restore()
		t.Fatalf("running shim: %v: %s", err, out)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	got, err := waitForCapturedURL(ctx, filepath.Join(dir, capturedURLFile))
	if err != nil {
		restore()
		t.Fatalf("waitForCapturedURL() error = %v", err)
	}
	if got != want {
		restore()
		t.Fatalf("captured %q, want %q", got, want)
	}

	restore()
	if os.Getenv("PATH") != origPath {
		t.Fatal("PATH not restored")
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("shim dir not removed: %v", err)
	}
}

func TestWaitForCapturedURLHonorsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := waitForCapturedURL(ctx, filepath.Join(t.TempDir(), capturedURLFile)); err == nil {
		t.Fatal("expected context error")
	}
}
