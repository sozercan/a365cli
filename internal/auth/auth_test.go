package auth

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

func TestCredentialOptionsDisableAutomaticAuthentication(t *testing.T) {
	opts := credentialOptions("00000000-0000-0000-0000-000000000000", "contoso.onmicrosoft.com", nil, true)
	if !opts.DisableAutomaticAuthentication {
		t.Fatal("ordinary credential must not automatically open a browser")
	}

	opts = credentialOptions("00000000-0000-0000-0000-000000000000", "contoso.onmicrosoft.com", nil, false)
	if opts.DisableAutomaticAuthentication {
		t.Fatal("portable interactive credential must allow browser authentication")
	}
}

func TestConfigureTokenCacheFallback(t *testing.T) {
	tests := []struct {
		name              string
		cacheApplied      bool
		cacheErr          error
		allowWithoutCache bool
		wantDisabled      bool
		wantErr           bool
	}{
		{name: "persistent cache keeps fallback disabled", cacheApplied: true, allowWithoutCache: true, wantDisabled: true},
		{name: "portable interactive enables fallback", allowWithoutCache: true, wantDisabled: false},
		{name: "cache failure degrades for interactive use", cacheErr: errors.New("cache unavailable"), allowWithoutCache: true, wantDisabled: false},
		{name: "cache failure remains fatal without fallback", cacheErr: errors.New("cache unavailable"), wantDisabled: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := credentialOptions("", "", nil, true)
			err := configureTokenCacheFallback(opts, tt.cacheApplied, tt.cacheErr, tt.allowWithoutCache)
			if (err != nil) != tt.wantErr {
				t.Fatalf("configureTokenCacheFallback() error = %v, wantErr %v", err, tt.wantErr)
			}
			if opts.DisableAutomaticAuthentication != tt.wantDisabled {
				t.Fatalf("DisableAutomaticAuthentication = %v, want %v", opts.DisableAutomaticAuthentication, tt.wantDisabled)
			}
		})
	}
}

func TestNewCredentialRejectsCorruptAuthRecord(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dir := filepath.Join(home, ".a365")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("create auth dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "auth-record.json"), []byte("not-json"), 0o600); err != nil {
		t.Fatalf("write auth record: %v", err)
	}

	_, err := NewCredential("00000000-0000-0000-0000-000000000000", "")
	if err == nil {
		t.Fatal("expected corrupt auth record error")
	}
	if !strings.Contains(err.Error(), "load auth record") {
		t.Fatalf("error = %q, want load auth record context", err)
	}
}

func TestCacheAuthenticationRecordReturnsPersistenceError(t *testing.T) {
	homeFile := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(homeFile, []byte("file"), 0o600); err != nil {
		t.Fatalf("write home sentinel: %v", err)
	}
	t.Setenv("HOME", homeFile)
	t.Setenv("USERPROFILE", homeFile)

	err := cacheAuthenticationRecord(&azidentity.AuthenticationRecord{})
	if err == nil {
		t.Fatal("expected auth-record persistence error")
	}
	if !strings.Contains(err.Error(), "cache auth record") {
		t.Fatalf("error = %q", err)
	}
}

func TestDeviceCodeOptionsMirrorBrowserOptions(t *testing.T) {
	record := &azidentity.AuthenticationRecord{Username: "alice@contoso.com", TenantID: "00000000-0000-0000-0000-000000000000"}
	src := credentialOptions("00000000-0000-0000-0000-000000000001", "contoso.onmicrosoft.com", record, true)

	var buf bytes.Buffer
	opts := deviceCodeOptions(src, &buf)

	if opts.ClientID != src.ClientID || opts.TenantID != src.TenantID {
		t.Fatalf("client/tenant not mirrored: got %q/%q", opts.ClientID, opts.TenantID)
	}
	if opts.AuthenticationRecord != src.AuthenticationRecord {
		t.Fatalf("authentication record not mirrored")
	}
	if !opts.DisableAutomaticAuthentication {
		t.Fatalf("DisableAutomaticAuthentication not mirrored")
	}
	if opts.UserPrompt == nil {
		t.Fatal("UserPrompt not set")
	}
	if err := opts.UserPrompt(context.Background(), azidentity.DeviceCodeMessage{Message: "enter code ABC123"}); err != nil {
		t.Fatalf("UserPrompt() error = %v", err)
	}
	if got := buf.String(); got != "enter code ABC123\n" {
		t.Fatalf("UserPrompt wrote %q", got)
	}
}

func TestBrowserAvailableWithoutLaunchers(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("launcher detection only applies on linux")
	}
	t.Setenv("PATH", t.TempDir())
	if BrowserAvailable() {
		t.Fatal("BrowserAvailable() = true with no launchers on PATH")
	}
}
