package cli

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"devsecurity/sca/internal/model"
)

func TestAnalyzeDirectoryWritesResult(t *testing.T) {
	t.Setenv("SCA_USE_MAVEN", "false")
	root := t.TempDir()
	pom := `<project>
  <groupId>com.example</groupId>
  <artifactId>demo</artifactId>
  <version>1.0.0</version>
  <dependencies>
    <dependency>
      <groupId>org.apache.logging.log4j</groupId>
      <artifactId>log4j-core</artifactId>
      <version>2.14.1</version>
    </dependency>
  </dependencies>
</project>`
	if err := os.WriteFile(filepath.Join(root, "pom.xml"), []byte(pom), 0o600); err != nil {
		t.Fatal(err)
	}
	data := filepath.Join(t.TempDir(), "vulnerabilities.json")
	records := `[{
      "id":"CVE-TEST-1",
      "title":"test",
      "severity":"HIGH",
      "package":"org.apache.logging.log4j:log4j-core",
      "introduced":"2.0.0",
      "fixed":"2.15.0"
    }]`
	if err := os.WriteFile(data, []byte(records), 0o600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "result.json")
	var stdout, stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{
		"analyze", "--input", root, "--output", output, "--data", data,
	}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit=%d stderr=%s", exitCode, stderr.String())
	}
	var result model.Result
	content, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(content, &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Components) != 1 || len(result.Findings) != 1 {
		t.Fatalf("unexpected result: components=%d findings=%d",
			len(result.Components), len(result.Findings))
	}
}

func TestAnalyzeRequiresInputAndOutput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if exitCode := Run(context.Background(), []string{"analyze", "--input", "."}, &stdout, &stderr); exitCode != 2 {
		t.Fatalf("expected usage exit 2, got %d", exitCode)
	}
}

func TestAnalyzeZipWritesResult(t *testing.T) {
	t.Setenv("SCA_USE_MAVEN", "false")
	archivePath := filepath.Join(t.TempDir(), "source.zip")
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	entry, err := writer.Create("project/pom.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte(`<project>
  <dependencies>
    <dependency><groupId>g</groupId><artifactId>a</artifactId><version>1.0</version></dependency>
  </dependencies>
</project>`)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	data := filepath.Join(t.TempDir(), "vulnerabilities.json")
	if err := os.WriteFile(data, []byte("[]"), 0o600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "result.json")
	var stdout, stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{
		"analyze", "--input", archivePath, "--output", output, "--data", data,
	}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit=%d stderr=%s", exitCode, stderr.String())
	}
	if _, err := os.Stat(output); err != nil {
		t.Fatal(err)
	}
}

func TestVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if exitCode := Run(context.Background(), []string{"version"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("version exit=%d stderr=%s", exitCode, stderr.String())
	}
	if stdout.Len() == 0 {
		t.Fatal("version output is empty")
	}
}
