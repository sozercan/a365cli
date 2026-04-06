package onedriveremote

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/sozercan/a365cli/internal/testutil"
)

func TestODRInfoCmd_Run(t *testing.T) {
	ctx, buf := testutil.SetupTestServer(t, map[string]string{
		"getOnedrive": `{"id":"drive-001","driveType":"personal","owner":{"user":{"displayName":"Alice","email":"alice@contoso.com"}},"quota":{"total":1099511627776,"used":536870912}}`,
	})

	cmd := &ODRInfoCmd{}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if result["id"] != "drive-001" {
		t.Errorf("expected id=drive-001, got %v", result["id"])
	}
	if result["driveType"] != "personal" {
		t.Errorf("expected driveType=personal, got %v", result["driveType"])
	}
}

func TestODRLsCmd_Run(t *testing.T) {
	ctx, buf := testutil.SetupTestServer(t, map[string]string{
		"getFolderChildrenInMyOnedrive": `{"value":[{"id":"item-001","name":"Documents","folder":{"childCount":5}},{"id":"item-002","name":"report.docx","size":12345}]}`,
	})

	cmd := &ODRLsCmd{}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	value, ok := result["value"]
	if !ok {
		t.Fatalf("expected 'value' key in output, got: %s", buf.String())
	}
	arr, ok := value.([]any)
	if !ok {
		t.Fatalf("expected 'value' to be an array, got: %T", value)
	}
	if len(arr) != 2 {
		t.Fatalf("expected 2 items, got %d", len(arr))
	}
}

func TestODRMkdirCmd_DryRun(t *testing.T) {
	ctx, buf := testutil.SetupTestServer(t, nil)
	ctx.DryRun = true

	cmd := &ODRMkdirCmd{FolderName: "NewFolder"}
	if err := cmd.Run(ctx); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if result["dry_run"] != true {
		t.Errorf("expected dry_run=true, got %v", result["dry_run"])
	}
	if result["action"] != "onedrive-remote.mkdir" {
		t.Errorf("expected action=onedrive-remote.mkdir, got %v", result["action"])
	}
}

func TestODRRmCmd_NoInput(t *testing.T) {
	ctx, _ := testutil.SetupTestServer(t, nil)
	ctx.NoInput = true

	cmd := &ODRRmCmd{FileOrFolderID: "item-001", Etag: "etag-001"}
	err := cmd.Run(ctx)
	if err == nil {
		t.Fatal("expected error when NoInput=true and Force=false")
	}
	if !strings.Contains(err.Error(), "without --force") {
		t.Errorf("expected error about --force, got: %v", err)
	}
}
