package analyzer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFindingsMapsMetadataAndRelativePath(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "findings.tsv")
	content := strings.Join([]string{
		"SAST-JAVA-001\tsrc/Vulnerable.java\t12\tRuntime.getRuntime().exec(command)",
		"SAST-JAVA-002\tsrc/Vulnerable.java\t20\tMessageDigest.getInstance(\\\"MD5\\\")",
		"UNKNOWN\tsrc/Other.java\t1\tignored",
	}, "\n")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	findings, err := parseFindings(file, root)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
	if findings[0].FilePath != "src/Vulnerable.java" || findings[0].StartLine != 12 {
		t.Fatalf("unexpected location: %#v", findings[0])
	}
	if findings[1].CWE != "CWE-327" {
		t.Fatalf("unexpected metadata: %#v", findings[1])
	}
}

func TestExtractZipRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	archive := filepath.Join(root, "source.zip")
	writeZip(t, archive, map[string]string{"../Vulnerable.java": "class X {}"})

	_, err := ExtractZip(archive, filepath.Join(root, "out"), ArchiveLimits{})
	if err == nil || !strings.Contains(err.Error(), "不安全路径") {
		t.Fatalf("expected traversal rejection, got %v", err)
	}
}

func TestExtractZipRequiresJava(t *testing.T) {
	root := t.TempDir()
	archive := filepath.Join(root, "source.zip")
	writeZip(t, archive, map[string]string{"README.md": "no java"})

	hasJava, err := ExtractZip(archive, filepath.Join(root, "out"), ArchiveLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if hasJava {
		t.Fatal("did not expect Java source")
	}
}
