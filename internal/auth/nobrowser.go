package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

const capturedURLFile = "auth-url"

// browserLaunchers are the commands the underlying browser package resolves
// from PATH: xdg-open, x-www-browser, and www-browser on Linux; open on macOS.
var browserLaunchers = []string{"xdg-open", "x-www-browser", "www-browser", "open"}

// AuthenticateWithoutBrowser runs the interactive browser (auth code + PKCE)
// flow without launching a local browser. The sign-in URL is passed to
// showURL so the user can open it in a browser on another machine. The
// redirect back to http://localhost:<port> must reach this host, for example
// through SSH port forwarding, so pair it with CredentialOptions.RedirectPort.
func (c *Credential) AuthenticateWithoutBrowser(ctx context.Context, showURL func(string)) (azcore.AccessToken, error) {
	dir, restore, err := installBrowserShim()
	if err != nil {
		return azcore.AccessToken{}, fmt.Errorf("prepare sign-in URL capture: %w", err)
	}
	defer restore()

	watchCtx, stopWatching := context.WithCancel(ctx)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if u, err := waitForCapturedURL(watchCtx, filepath.Join(dir, capturedURLFile)); err == nil {
			showURL(u)
		}
	}()

	token, err := c.Authenticate(ctx)
	stopWatching()
	wg.Wait()
	return token, err
}

// installBrowserShim prepends a temp directory to PATH containing stand-in
// browser launchers that record the URL they are asked to open instead of
// opening it. The returned function restores PATH and removes the directory.
func installBrowserShim() (string, func(), error) {
	if runtime.GOOS == "windows" {
		return "", nil, errors.New("not supported on Windows")
	}

	dir, err := os.MkdirTemp("", "a365-login-")
	if err != nil {
		return "", nil, err
	}
	cleanupDir := func() { _ = os.RemoveAll(dir) }
	if strings.ContainsRune(dir, '\'') {
		cleanupDir()
		return "", nil, fmt.Errorf("unsupported temp dir path %q", dir)
	}

	target := filepath.Join(dir, capturedURLFile)
	// Write to a temp file then rename so readers never see a partial URL.
	// printf is a shell builtin; "command -p" resolves mv from the standard
	// utility path so the shim works even when the caller's PATH is empty.
	script := fmt.Sprintf("#!/bin/sh\nprintf '%%s' \"$1\" > '%[1]s.tmp' && command -p mv '%[1]s.tmp' '%[1]s'\n", target)
	for _, name := range browserLaunchers {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0o700); err != nil { //nolint:gosec // executable shim in a private temp dir
			cleanupDir()
			return "", nil, err
		}
	}

	oldPath, hadPath := os.LookupEnv("PATH")
	newPath := dir
	if oldPath != "" {
		newPath += string(os.PathListSeparator) + oldPath
	}
	if err := os.Setenv("PATH", newPath); err != nil {
		cleanupDir()
		return "", nil, err
	}

	restore := func() {
		if hadPath {
			_ = os.Setenv("PATH", oldPath)
		} else {
			_ = os.Unsetenv("PATH")
		}
		cleanupDir()
	}
	return dir, restore, nil
}

func waitForCapturedURL(ctx context.Context, path string) (string, error) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		if data, err := os.ReadFile(path); err == nil {
			if u := strings.TrimSpace(string(data)); u != "" {
				return u, nil
			}
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
		}
	}
}
