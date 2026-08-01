package commands

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/sozercan/a365cli/internal/output"
)

func TestAuthStatusCmd_ReportsUnverifiedCachedAccount(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := home + "/.a365"
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("create auth dir: %v", err)
	}
	record := `{
  "authority": "https://login.microsoftonline.com/organizations/",
  "clientId": "00000000-0000-0000-0000-000000000000",
  "homeAccountId": "11111111-1111-1111-1111-111111111111.22222222-2222-2222-2222-222222222222",
  "tenantId": "22222222-2222-2222-2222-222222222222",
  "username": "alice@contoso.com",
  "version": "1.0"
}`
	if err := os.WriteFile(dir+"/auth-record.json", []byte(record), 0o600); err != nil {
		t.Fatalf("write auth record: %v", err)
	}

	var buf bytes.Buffer
	ctx := &Context{Output: &output.Formatter{Writer: &buf}}
	if err := (&AuthStatusCmd{}).Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if got := buf.String(); !strings.Contains(got, "Cached account: alice@contoso.com (token not verified)") {
		t.Fatalf("status output = %q", got)
	}
}

func TestAuthStatusCmd_RejectsCorruptRecord(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := home + "/.a365"
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("create auth dir: %v", err)
	}
	if err := os.WriteFile(dir+"/auth-record.json", []byte("not-json"), 0o600); err != nil {
		t.Fatalf("write auth record: %v", err)
	}

	ctx := &Context{Output: &output.Formatter{Writer: &bytes.Buffer{}}}
	err := (&AuthStatusCmd{}).Run(ctx)
	if err == nil {
		t.Fatal("expected corrupt auth record error")
	}
	if !strings.Contains(err.Error(), "a365 auth logout") {
		t.Fatalf("error = %q", err)
	}
}

func TestAuthTokenCmd_JSONOutputIsValidJSON(t *testing.T) {
	token := testJWT(t, map[string]any{
		"appid": "00000000-0000-0000-0000-000000000000",
		"tid":   "11111111-1111-1111-1111-111111111111",
		"upn":   "alice@contoso.com",
		"name":  "Alice",
		"aud":   "api://example",
		"scp":   "Mail.Read",
		"exp":   float64(4102444800),
	})

	var buf bytes.Buffer
	ctx := &Context{
		Ctx: context.Background(),
		TokenProvider: func(context.Context) (string, error) {
			return token, nil
		},
		Output: &output.Formatter{Format: output.FormatJSON, Writer: &buf},
	}

	if err := (&AuthTokenCmd{}).Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var claims map[string]any
	if err := json.Unmarshal(buf.Bytes(), &claims); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if claims["upn"] != "alice@contoso.com" {
		t.Fatalf("expected upn claim in JSON output, got %v", claims["upn"])
	}
}

func TestEnsureAuth_NoInputWithoutCachedAuth(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	ctx := &Context{
		Ctx:     context.Background(),
		NoInput: true,
	}

	err := ctx.EnsureAuth()
	if err == nil {
		t.Fatal("expected error when auth is required in non-interactive mode")
	}
	if !strings.Contains(err.Error(), "a365 auth login") {
		t.Fatalf("expected explicit login guidance, got %v", err)
	}
}

func TestEnsureAuth_WithoutCachedAuthRequiresExplicitLogin(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	ctx := &Context{Ctx: context.Background()}
	err := ctx.EnsureAuth()
	if err == nil {
		t.Fatal("expected error when no cached authentication exists")
	}
	if !strings.Contains(err.Error(), "a365 auth login") {
		t.Fatalf("expected explicit login guidance, got %v", err)
	}
}

func testJWT(t *testing.T, claims map[string]any) string {
	t.Helper()

	encode := func(v any) string {
		data, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("marshal JWT part: %v", err)
		}
		return base64.RawURLEncoding.EncodeToString(data)
	}

	return strings.Join([]string{
		encode(map[string]any{"alg": "none", "typ": "JWT"}),
		encode(claims),
		"signature",
	}, ".")
}
