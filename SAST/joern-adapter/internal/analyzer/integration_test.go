package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestJoernIntegration(t *testing.T) {
	joernHome := os.Getenv("TEST_JOERN_HOME")
	if joernHome == "" {
		t.Skip("set TEST_JOERN_HOME to run the Joern integration test")
	}
	root := t.TempDir()
	sourceRoot := filepath.Join(root, "source")
	workRoot := filepath.Join(root, "work")
	if err := os.MkdirAll(sourceRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	source := `import java.io.*;
import java.security.*;
public class Vulnerable {
  public void run(String command) throws Exception { Runtime.getRuntime().exec(command); }
  public byte[] hash(byte[] value) throws Exception { return MessageDigest.getInstance("MD5").digest(value); }
  public Object read(InputStream input) throws Exception { return new ObjectInputStream(input).readObject(); }
}`
	if err := os.WriteFile(filepath.Join(sourceRoot, "Vulnerable.java"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := (Scanner{JoernHome: joernHome, Timeout: 10 * time.Minute}).
		Analyze(context.Background(), sourceRoot, workRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(result.Findings))
	}
	for _, finding := range result.Findings {
		if finding.FilePath != "Vulnerable.java" || finding.StartLine <= 0 {
			t.Fatalf("invalid finding location: %#v", finding)
		}
	}
}
