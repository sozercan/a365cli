package auth

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestApplyAutomaticAuthenticationFallback(t *testing.T) {
	tests := []struct {
		name              string
		cacheApplied      bool
		allowWithoutCache bool
		wantDisabled      bool
	}{
		{name: "persistent cache keeps fallback disabled", cacheApplied: true, allowWithoutCache: true, wantDisabled: true},
		{name: "portable interactive enables fallback", allowWithoutCache: true, wantDisabled: false},
		{name: "no input remains disabled", wantDisabled: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := credentialOptions("", "", nil, true)
			applyAutomaticAuthenticationFallback(opts, tt.cacheApplied, tt.allowWithoutCache)
			if opts.DisableAutomaticAuthentication != tt.wantDisabled {
				t.Fatalf("DisableAutomaticAuthentication = %v, want %v", opts.DisableAutomaticAuthentication, tt.wantDisabled)
			}
		})
	}
}

func TestNewCredentialRejectsCorruptAuthRecord(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

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
