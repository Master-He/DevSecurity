package workspace

import (
	"os"
	"testing"
)

const scanID = "123e4567-e89b-42d3-a456-426614174000"

func TestValidateDirectoryInput(t *testing.T) {
	root := t.TempDir()
	layout, err := New(root, scanID)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(layout.InputRoot, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(layout.Ready, []byte("ready"), 0o600); err != nil {
		t.Fatal(err)
	}
	input, err := layout.ValidateInput("directory")
	if err != nil {
		t.Fatal(err)
	}
	if input != layout.InputRoot {
		t.Fatalf("unexpected input path: %s", input)
	}
}

func TestValidateRequiresReadyMarker(t *testing.T) {
	root := t.TempDir()
	layout, err := New(root, scanID)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(layout.InputRoot, 0o750); err != nil {
		t.Fatal(err)
	}
	if _, err := layout.ValidateInput("directory"); err == nil {
		t.Fatal("expected missing READY to fail")
	}
}

func TestRejectsPathLikeScanID(t *testing.T) {
	if _, err := New(t.TempDir(), "../../outside"); err == nil {
		t.Fatal("expected invalid scan ID to fail")
	}
}

func TestRejectsInputSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	layout, err := New(root, scanID)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(layout.ScanRoot, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(layout.Ready, []byte("ready"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, layout.InputRoot); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}
	if _, err := layout.ValidateInput("directory"); err == nil {
		t.Fatal("expected symlink escape to fail")
	}
}
