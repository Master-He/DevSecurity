package analyzer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParsePOMResolvesPropertyAndUnknownVersion(t *testing.T) {
	root := t.TempDir()
	pom := `<?xml version="1.0" encoding="UTF-8"?>
<project>
  <modelVersion>4.0.0</modelVersion>
  <groupId>com.example</groupId>
  <artifactId>demo</artifactId>
  <version>1.0.0</version>
  <properties><log4j.version>2.14.1</log4j.version></properties>
  <dependencies>
    <dependency>
      <groupId>org.apache.logging.log4j</groupId>
      <artifactId>log4j-core</artifactId>
      <version>${log4j.version}</version>
    </dependency>
    <dependency>
      <groupId>com.example</groupId>
      <artifactId>internal-sdk</artifactId>
    </dependency>
  </dependencies>
</project>`
	if err := os.WriteFile(filepath.Join(root, "pom.xml"), []byte(pom), 0o600); err != nil {
		t.Fatal(err)
	}
	components, err := ParsePoms(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(components) != 2 {
		t.Fatalf("expected 2 components, got %d", len(components))
	}
	byArtifact := map[string]string{}
	for _, component := range components {
		byArtifact[component.ArtifactID] = component.Version
	}
	if byArtifact["log4j-core"] != "2.14.1" {
		t.Fatalf("property was not resolved: %s", byArtifact["log4j-core"])
	}
	if byArtifact["internal-sdk"] != "unknown" {
		t.Fatalf("missing version must be unknown, got %s", byArtifact["internal-sdk"])
	}
}

func TestParseDependencyTreePreservesPath(t *testing.T) {
	output := `[INFO] com.example:demo:jar:1.0.0
[INFO] +- org.example:direct:jar:1.0.0:compile
[INFO] |  \- org.example:transitive:jar:2.0.0:runtime
[INFO] \- org.example:last:jar:3.0.0:test`
	components := ParseDependencyTree(output)
	if len(components) != 3 {
		t.Fatalf("expected 3 components, got %d", len(components))
	}
	for _, component := range components {
		if component.ArtifactID == "transitive" {
			if component.Direct {
				t.Fatal("transitive component marked direct")
			}
			if !strings.Contains(component.DependencyPath, "direct") {
				t.Fatalf("dependency path missing parent: %s", component.DependencyPath)
			}
			return
		}
	}
	t.Fatal("transitive component not found")
}

func TestPOMXMLWithoutNamespaceIsAccepted(t *testing.T) {
	root := t.TempDir()
	pom := `<project><dependencies><dependency><groupId>g</groupId><artifactId>a</artifactId><version>1</version></dependency></dependencies></project>`
	if err := os.WriteFile(filepath.Join(root, "pom.xml"), []byte(pom), 0o600); err != nil {
		t.Fatal(err)
	}
	components, err := ParsePoms(root)
	if err != nil || len(components) != 1 {
		t.Fatalf("expected one component, got %d, err=%v", len(components), err)
	}
}
