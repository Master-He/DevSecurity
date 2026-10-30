package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveDataPathUsesAbsolutePath(t *testing.T) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.CreateTemp(workingDirectory, "vulnerabilities-*.json")
	if err != nil {
		t.Fatal(err)
	}
	name := file.Name()
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(name)

	resolved, err := resolveDataPath(filepath.Base(name))
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(resolved) {
		t.Fatalf("expected absolute path, got %s", resolved)
	}
	if resolved != name {
		t.Fatalf("expected %s, got %s", name, resolved)
	}
}
