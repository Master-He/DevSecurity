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

func TestResolveAuthTokenDefaultsOnlyForLoopback(t *testing.T) {
	t.Setenv("SCA_SERVICE_TOKEN", "")
	token, err := resolveAuthToken("127.0.0.1:8086")
	if err != nil {
		t.Fatal(err)
	}
	if token != "123" {
		t.Fatalf("expected local default token, got %s", token)
	}
	if _, err := resolveAuthToken("0.0.0.0:8086"); err == nil {
		t.Fatal("expected external listener to require an explicit token")
	}
}

func TestDefaultWorkRootUsesUserHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	expected := filepath.Join(home, ".dev-security", "sca-work")
	if actual := defaultWorkRoot(); actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}
}
