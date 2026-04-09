package commands

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/sozercan/a365cli/internal/output"
)

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
	if !strings.Contains(err.Error(), "non-interactive") {
		t.Fatalf("expected non-interactive error, got %v", err)
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
