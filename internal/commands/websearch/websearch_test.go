package websearch

import (
	"encoding/json"
	"testing"

	"github.com/sozercan/a365cli/internal/testutil"
)

func TestWebSearchCmd_Run(t *testing.T) {
	ctx, buf := testutil.SetupTestServer(t, map[string]string{
		"SearchWeb": `{"results":[{"title":"Contoso Home","url":"https://www.contoso.com","snippet":"Welcome to Contoso."},{"title":"Contoso Blog","url":"https://blog.contoso.com","snippet":"Latest news from Contoso."}]}`,
	})

	cmd := &WebSearchSearchCmd{Query: "contoso", URLs: []string{"https://www.contoso.com"}}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	results, ok := result["results"]
	if !ok {
		t.Fatalf("expected 'results' key in output, got: %s", buf.String())
	}
	arr, ok := results.([]any)
	if !ok {
		t.Fatalf("expected 'results' to be an array, got: %T", results)
	}
	if len(arr) != 2 {
		t.Fatalf("expected 2 results, got %d", len(arr))
	}
}
