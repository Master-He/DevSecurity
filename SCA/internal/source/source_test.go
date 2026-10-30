package source

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestPrepareZipExtractsRegularFiles(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "source.zip")
	writeZip(t, archive, map[string]string{
		"project/pom.xml":       "<project/>",
		"project/src/Main.java": "class Main {}",
	})
	prepared, err := Prepare(archive, Limits{
		MaxArchiveBytes:     10 << 20,
		MaxUncompressedSize: 10 << 20,
		MaxFiles:            10,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer prepared.Close()
	content, err := os.ReadFile(filepath.Join(prepared.Root, "project", "pom.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "<project/>" {
		t.Fatalf("unexpected extracted content: %q", content)
	}
}

func TestPrepareRejectsPathTraversal(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "source.zip")
	writeZip(t, archive, map[string]string{"../escaped.xml": "bad"})
	if _, err := Prepare(archive, Limits{}); err == nil {
		t.Fatal("expected path traversal to be rejected")
	}
}

func TestPrepareRejectsSymbolicLink(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "source.zip")
	file, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	header := &zip.FileHeader{Name: "pom.xml"}
	header.SetMode(os.ModeSymlink | 0o777)
	entry, err := writer.CreateHeader(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("../../outside")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := Prepare(archive, Limits{}); err == nil {
		t.Fatal("expected symbolic link to be rejected")
	}
}

func writeZip(t *testing.T, path string, entries map[string]string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	for name, content := range entries {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}
