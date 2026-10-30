package control

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devsecurity/sca/internal/workspace"
)

func TestReadLogTailRedactsWorkspace(t *testing.T) {
	root := t.TempDir()
	logs := filepath.Join(root, "logs")
	if err := os.MkdirAll(logs, 0o750); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(logs, "sca.stderr.log")
	content := "scan failed at " + filepath.Join(root, "scans", "case", "input") + "\nsecond line"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	layout := workspace.Layout{
		Root:     root,
		ScanRoot: filepath.Join(root, "scans", "case"),
	}
	detail := readLogTail(path, layout, 1200)
	if strings.Contains(detail, root) {
		t.Fatalf("workspace path was not redacted: %s", detail)
	}
	if !strings.Contains(detail, "second line") {
		t.Fatalf("stderr detail missing: %s", detail)
	}
}
